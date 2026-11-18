// Two-node staged communication component, not a distributed PIR transform.
// Field kernels reuse the maintained CUDA arithmetic; CPU oracle uses the
// independent polynomial-basis implementation from the native reference.
#include <cuda_runtime.h>
#include <mpi.h>
#include "../../encoding/liftrs-cuda/transform.cuh"
#include "../../encoding/liftrs-cpp/arithmetic.hpp"
#include "../../encoding/liftrs-cpp/prepare_plan.hpp"
#include <sys/resource.h>
#include <array>
#include <chrono>
#include <ctime>
#include <cstdlib>
#include <fstream>
#include <iostream>
#include <limits>
#include <string>

using Clock=std::chrono::steady_clock;
int rank_id=0;
Clock::time_point began;
void gpu(cudaError_t status) {if(status!=cudaSuccess)throw std::runtime_error(cudaGetErrorString(status));}
void mpi(int status) {if(status!=MPI_SUCCESS)throw std::runtime_error("MPI operation failed");}
double seconds(Clock::time_point start) {return std::chrono::duration<double>(Clock::now()-start).count();}
uint64_t rss() {
  rusage usage{};
  if(getrusage(RUSAGE_SELF,&usage)||usage.ru_maxrss>2*(1<<20))throw std::runtime_error("rank RSS exceeds2GiB or unavailable");
  return uint64_t(usage.ru_maxrss)*1024;
}
void event(const char *phase,const char *kind,size_t done,size_t total,size_t bytes,const char *unit="passes") {
  auto now=std::time(nullptr);char stamp[32];std::strftime(stamp,sizeof(stamp),"%FT%TZ",std::gmtime(&now));
  std::cout<<"{\"utc\":\""<<stamp<<"\",\"rank\":"<<rank_id<<",\"phase\":\""<<phase
    <<"\",\"event\":\""<<kind<<"\",\"completed\":"<<done<<",\"total\":"<<total
    <<",\"unit\":\""<<unit<<"\",\"bytes\":"<<bytes<<",\"elapsed_seconds\":"<<seconds(began)
    <<",\"peak_rss_bytes\":"<<rss()<<"}"<<std::endl;
}
// CPU references are outside pipeline clocks but still expose advancing work
// if contention makes a fixed-size verification phase unexpectedly long.
template<class Check>void verify_words(const char *phase,size_t words,size_t bytes,Check check) {
  event(phase,"started",0,words,bytes,"words");auto last=Clock::now();
  for(size_t i=0;i<words;i++) {
    if(!check(i))throw std::runtime_error(phase);
    if((i&65535)==65535&&Clock::now()-last>=std::chrono::seconds(30)) {
      event(phase,"progress",i+1,words,bytes,"words");last=Clock::now();
    }
  }
  event(phase,"completed",words,words,bytes,"words");
}
struct Field {
  unsigned s;uint32_t polynomial,delta;std::vector<uint32_t> logs,powers;
  explicit Field(const char *path) {
    std::ifstream input(path,std::ios::binary);if(!input)throw std::runtime_error("field fixture open");
    auto word=[&](){uint32_t value=0;for(unsigned i=0;i<4;i++){int c=input.get();if(c==EOF)throw std::runtime_error("field fixture truncated");value|=uint32_t(uint8_t(c))<<(8*i);}return value;};
    if(word()!=0x534b4754)throw std::runtime_error("field fixture magic");
    s=word();polynomial=word();delta=word();
    if(s!=17||(polynomial>>s)!=1||!delta||delta>=(1u<<s))throw std::runtime_error("field fixture shape");
    auto read=[&](std::vector<uint32_t>&v,size_t count,uint32_t limit){if(word()!=count)throw std::runtime_error("table count");v.resize(count);for(auto &x:v){x=word();if(x>=limit)throw std::runtime_error("table range");}};
    read(logs,1u<<s,(1u<<s)-1);read(powers,2*((1u<<s)-1),1u<<s);
    if(input.get()!=EOF)throw std::runtime_error("field fixture trailing data");
  }
};
struct Device {
  void *data=nullptr;size_t bytes;
  explicit Device(size_t n):bytes(n){gpu(cudaMalloc(&data,n));}
  ~Device(){if(data)cudaFree(data);}
  Device(const Device&)=delete;Device&operator=(const Device&)=delete;
};
struct Pinned {
  uint64_t *data=nullptr;
  explicit Pinned(size_t bytes){gpu(cudaMallocHost(&data,bytes));}
  ~Pinned(){if(data)cudaFreeHost(data);}
  Pinned(const Pinned&)=delete;Pinned&operator=(const Pinned&)=delete;
};
__host__ __device__ uint64_t pattern(int source,size_t index) {
  return (uint64_t(source+1)<<32)|uint64_t(index);
}
__global__ void produce(uint64_t *data,size_t count,int rank,uint64_t factor,
                        unsigned s,uint32_t delta,const uint32_t *logs,const uint32_t *powers) {
  size_t i=size_t(blockIdx.x)*blockDim.x+threadIdx.x;
  if(i<count)data[i]=liftrs_cuda::multiply(pattern(rank,i),factor,s,delta,logs,powers);
}
__global__ void consume(uint64_t *data,size_t count,int rank,uint64_t inverse,
                        unsigned s,uint32_t delta,const uint32_t *logs,const uint32_t *powers,uint32_t *bad) {
  size_t i=size_t(blockIdx.x)*blockDim.x+threadIdx.x;
  if(i<count){
    data[i]=liftrs_cuda::multiply(data[i],inverse,s,delta,logs,powers);
    size_t half=count/2;int source=int(i/half);
    if(data[i]!=pattern(source,size_t(rank)*half+i%half))atomicOr(bad,1u);
  }
}

void run_case(const Field &field,size_t bytes,bool small) {
  const size_t words=bytes/8,chunk=words/2;
  const auto setup=Clock::now();
  Pinned send(bytes),receive(bytes);
  Device data(bytes),logs(field.logs.size()*4),powers(field.powers.size()*4),bad(4);
  const size_t owned=data.bytes+logs.bytes+powers.bytes+bad.bytes;
  if(owned>128u<<20)throw std::runtime_error("owned device cap exceeded");
  gpu(cudaMemcpy(logs.data,field.logs.data(),logs.bytes,cudaMemcpyHostToDevice));
  gpu(cudaMemcpy(powers.data,field.powers.data(),powers.bytes,cudaMemcpyHostToDevice));
  liftrs::FieldT<uint64_t> cpu;liftrs::preparation::set_field(cpu,field.s,field.polynomial,field.delta);
  const uint64_t factor=(uint64_t(1)<<field.s)|3;
  const auto inverse=liftrs::field_power(cpu,factor,cpu.q-2);
  if(cpu.mul(factor,inverse)!=1)throw std::runtime_error("field inverse oracle");
  const double setup_seconds=seconds(setup);
  const unsigned passes=small?2:4;
  auto *device=static_cast<uint64_t*>(data.data);
  auto *log=static_cast<uint32_t*>(logs.data),*power=static_cast<uint32_t*>(powers.data);
  for(unsigned pass=0;pass<passes;pass++) {
    event("staged-exchange","started",pass,passes,bytes);
    // Host-only transport on the same pinned buffers supplies a matched local
    // baseline. Correctness is independently checked before the GPU pipeline.
    for(size_t i=0;i<words;i++)send.data[i]=pattern(rank_id,i);
    mpi(MPI_Barrier(MPI_COMM_WORLD));auto start=Clock::now();
    mpi(MPI_Alltoall(send.data,int(chunk),MPI_UINT64_T,receive.data,int(chunk),MPI_UINT64_T,MPI_COMM_WORLD));
    const double host_seconds=seconds(start);
    verify_words("host-ownership-reference",words,bytes,[&](size_t i){return receive.data[i]==pattern(int(i/chunk),size_t(rank_id)*chunk+i%chunk);});
    mpi(MPI_Barrier(MPI_COMM_WORLD));const auto whole=Clock::now();
    std::array<double,7> times{};times[0]=host_seconds;
    start=Clock::now();produce<<<(words+255)/256,256>>>(device,words,rank_id,factor,field.s,field.delta,log,power);
    gpu(cudaGetLastError());gpu(cudaDeviceSynchronize());times[1]=seconds(start);
    start=Clock::now();gpu(cudaMemcpy(send.data,device,bytes,cudaMemcpyDeviceToHost));times[2]=seconds(start);
    start=Clock::now();mpi(MPI_Alltoall(send.data,int(chunk),MPI_UINT64_T,receive.data,int(chunk),MPI_UINT64_T,MPI_COMM_WORLD));times[3]=seconds(start);
    start=Clock::now();gpu(cudaMemcpy(device,receive.data,bytes,cudaMemcpyHostToDevice));times[4]=seconds(start);
    start=Clock::now();gpu(cudaMemset(bad.data,0,4));
    consume<<<(words+255)/256,256>>>(device,words,rank_id,inverse,field.s,field.delta,log,power,static_cast<uint32_t*>(bad.data));
    gpu(cudaGetLastError());gpu(cudaDeviceSynchronize());uint32_t flag=0;
    gpu(cudaMemcpy(&flag,bad.data,4,cudaMemcpyDeviceToHost));if(flag)throw std::runtime_error("GPU consumer ownership mismatch");
    times[5]=seconds(start);times[6]=seconds(whole);
    std::array<double,7> maxima{};mpi(MPI_Allreduce(times.data(),maxima.data(),7,MPI_DOUBLE,MPI_MAX,MPI_COMM_WORLD));
    // Full independent CPU oracle and full result download stay outside clocks.
    verify_words("CPU-producer-reference",words,bytes,[&](size_t i){return send.data[i]==cpu.mul(pattern(rank_id,i),factor);});
    gpu(cudaMemcpy(receive.data,device,bytes,cudaMemcpyDeviceToHost));
    verify_words("CPU-consumer-reference",words,bytes,[&](size_t i){return receive.data[i]==pattern(int(i/chunk),size_t(rank_id)*chunk+i%chunk);});
    mpi(MPI_Alltoall(receive.data,int(chunk),MPI_UINT64_T,send.data,int(chunk),MPI_UINT64_T,MPI_COMM_WORLD));
    verify_words("reverse-ownership-reference",words,bytes,[&](size_t i){return send.data[i]==pattern(rank_id,i);});
    std::cout<<"{\"kind\":\"measurement\",\"rank\":"<<rank_id<<",\"bytes\":"<<bytes<<",\"sample\":"<<int(pass)-1
      <<",\"host_seconds\":"<<maxima[0]<<",\"producer_seconds\":"<<maxima[1]<<",\"d2h_seconds\":"<<maxima[2]
      <<",\"exchange_seconds\":"<<maxima[3]<<",\"h2d_seconds\":"<<maxima[4]<<",\"consumer_seconds\":"<<maxima[5]
      <<",\"pipeline_seconds\":"<<maxima[6]<<",\"setup_seconds_local\":"<<setup_seconds<<",\"owned_device_bytes\":"<<owned
      <<",\"pinned_host_bytes\":"<<2*bytes<<",\"peak_rss_bytes\":"<<rss()<<",\"matches\":true}"<<std::endl;
    event("staged-exchange","completed",pass+1,passes,bytes);
  }
}
int main(int argc,char **argv) {
  MPI_Init(&argc,&argv);MPI_Comm_rank(MPI_COMM_WORLD,&rank_id);began=Clock::now();
  try {
    if(argc!=2&&argc!=3)throw std::runtime_error("field.bin [--small]");
    const bool small=argc==3;if(small&&std::string(argv[2])!="--small")throw std::runtime_error("unknown option");
    int ranks;mpi(MPI_Comm_size(MPI_COMM_WORLD,&ranks));if(ranks!=2)throw std::runtime_error("two ranks required");
    mpi(MPI_Comm_set_errhandler(MPI_COMM_WORLD,MPI_ERRORS_RETURN));
    std::array<char,MPI_MAX_PROCESSOR_NAME> host{};int length;mpi(MPI_Get_processor_name(host.data(),&length));
    std::array<char,2*MPI_MAX_PROCESSOR_NAME> hosts{};
    mpi(MPI_Allgather(host.data(),MPI_MAX_PROCESSOR_NAME,MPI_CHAR,hosts.data(),MPI_MAX_PROCESSOR_NAME,MPI_CHAR,MPI_COMM_WORLD));
    if(std::string(hosts.data())==std::string(hosts.data()+MPI_MAX_PROCESSOR_NAME))throw std::runtime_error("distinct nodes required");
    const char *slurm_rank=std::getenv("SLURM_PROCID");
    if(!slurm_rank||std::stoi(slurm_rank)!=rank_id)throw std::runtime_error("MPI/Slurm rank mapping mismatch");
    int visible=0;gpu(cudaGetDeviceCount(&visible));
    if(visible!=1)throw std::runtime_error("exactly one Slurm-visible GPU per rank required");
    gpu(cudaSetDevice(0));cudaDeviceProp properties{};gpu(cudaGetDeviceProperties(&properties,0));
    int runtime_version=0,driver_version=0;
    gpu(cudaRuntimeGetVersion(&runtime_version));gpu(cudaDriverGetVersion(&driver_version));
    std::cout<<"{\"kind\":\"host\",\"rank\":"<<rank_id<<",\"host\":\""<<host.data()
      <<"\",\"gpu\":\""<<properties.name<<"\",\"device_capacity_bytes\":"<<properties.totalGlobalMem
      <<",\"cuda_runtime_version\":"<<runtime_version<<",\"cuda_driver_api_version\":"<<driver_version<<"}"<<std::endl;
    Field field(argv[1]);std::cout.precision(12);
    const auto sizes=small?std::vector<size_t>{32u<<10}:std::vector<size_t>{1u<<20,8u<<20,32u<<20};
    for(auto bytes:sizes)run_case(field,bytes,small);
    event("workflow","completed",sizes.size(),sizes.size(),sizes.back());
    mpi(MPI_Finalize());return 0;
  } catch(const std::exception &error) {
    std::cerr<<"utc_epoch="<<std::time(nullptr)<<" phase=workflow event=failed rank="<<rank_id<<" error="<<error.what()<<std::endl;
    MPI_Abort(MPI_COMM_WORLD,1);return 1;
  }
}
