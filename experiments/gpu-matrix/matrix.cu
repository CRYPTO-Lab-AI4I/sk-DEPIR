// Experiment-only complete matrix additive transform. Coordinates match the CPU.
#include "gpu_api.h"
#include <cuda_runtime.h>
#include <algorithm>
#include <chrono>
#include <cstdio>
#include <memory>
#include <stdexcept>
#include <unordered_map>
#include <vector>

namespace {
constexpr unsigned levels=16,terms=17;
constexpr size_t max_points=size_t(1)<<22,device_cap=size_t(512)<<20;
struct Tile {unsigned count[levels];uint32_t degree[levels][terms],coef[levels][terms],values[levels][32],basis[levels],delta[levels],inverse[levels];};
struct Taylor {unsigned count;uint32_t degree[levels],coef[levels];};
struct Matrix {Tile rows,columns;Taylor taylor[levels];unsigned split,dimension,s;uint32_t delta;};
void check(cudaError_t e){if(e!=cudaSuccess)throw std::runtime_error(cudaGetErrorString(e));}
struct Device {
 void* p=nullptr;size_t bytes=0;
 Device()=default;Device(const Device&)=delete;Device& operator=(const Device&)=delete;
 ~Device(){if(p)cudaFree(p);}
 void allocate(size_t n){if(p){check(cudaFree(p));p=nullptr;bytes=0;}check(cudaMalloc(&p,n));bytes=n;}
 uint32_t* words(){return static_cast<uint32_t*>(p);}
 void swap(Device& other){std::swap(p,other.p);std::swap(bytes,other.bytes);}
};
__host__ __device__ uint32_t base_mul(uint32_t a,uint32_t b,const uint32_t* log,const uint32_t* powers){return a&&b?powers[log[a]+log[b]]:0;}
__host__ __device__ uint32_t multiply(uint32_t a,uint32_t b,unsigned s,uint32_t delta,const uint32_t* log,const uint32_t* powers){
 if(!a||!b)return 0;if(a==1)return b;if(b==1)return a;
 uint32_t mask=(1u<<s)-1,aa=a&mask,ab=a>>s,ba=b&mask,bb=b>>s,ac=base_mul(aa,ba,log,powers),bd=base_mul(ab,bb,log,powers);
 return (ac^base_mul(delta,bd,log,powers))|((base_mul(aa^ab,ba^bb,log,powers)^ac)<<s);
}
__global__ void transpose(const uint32_t* src,uint32_t* dst,size_t rows,size_t cols){
 size_t i=size_t(blockIdx.x)*blockDim.x+threadIdx.x;if(i>=rows*cols)return;dst[(i%cols)*rows+i/cols]=src[i];
}
__global__ void tiles(uint32_t* data,size_t n,const Matrix* matrix,bool columns,bool inverse,const uint32_t* log,const uint32_t* powers){
 const auto& m=*matrix;const auto& p=columns?m.columns:m.rows;
 unsigned level=columns?m.dimension-m.split:m.split;
 size_t length=size_t(1)<<level,jobs=n/length,tile=size_t(blockIdx.x)*blockDim.x+threadIdx.x;
 if(tile>=jobs)return;
 for(unsigned stage=0;stage<level;stage++){
  unsigned k=inverse?stage:level-1-stage;size_t size=size_t(1)<<(k+1),half=size/2;
  for(size_t group=0;group<length;group+=size){
   uint32_t point=columns?0:uint32_t(tile*length);
   for(unsigned bit=0;bit<level;bit++)if((group>>bit)&1)point^=p.basis[bit];
   uint32_t c=0;for(unsigned bit=0;bit<2*m.s;bit++)if((point>>bit)&1)c^=p.values[k][bit];
   if(!inverse){
    for(size_t i=size;i-- > half;){uint32_t z=data[(group+i)*jobs+tile];
     for(unsigned t=0;t+1<p.count[k];t++)data[(group+i-half+p.degree[k][t])*jobs+tile]^=multiply(p.coef[k][t],z,m.s,m.delta,log,powers);
    }
    for(size_t j=0;j<half;j++){size_t a=(group+j)*jobs+tile,b=(group+half+j)*jobs+tile;uint32_t v=data[b],left=data[a]^multiply(c,v,m.s,m.delta,log,powers);data[a]=left;data[b]=left^multiply(p.delta[k],v,m.s,m.delta,log,powers);}
   }else{
    for(size_t j=0;j<half;j++){size_t a=(group+j)*jobs+tile,b=(group+half+j)*jobs+tile;uint32_t v=multiply(p.inverse[k],data[a]^data[b],m.s,m.delta,log,powers);data[a]^=multiply(c,v,m.s,m.delta,log,powers);data[b]=v;}
    for(size_t j=0;j<half;j++){uint32_t v=data[(group+half+j)*jobs+tile];for(unsigned t=0;t+1<p.count[k];t++)data[(group+j+p.degree[k][t])*jobs+tile]^=multiply(p.coef[k][t],v,m.s,m.delta,log,powers);}
   }
  }
 }
}
// For linearized h, every lower degree d is <= half/2. Quotient correction
// q[j] reads only the untouched upper half of q. A second gather computes the
// remainder. Inverse reverses these passes. Each thread owns one destination;
// no atomics or racy scatter updates are required.
__global__ void taylor_pass(uint32_t* x,size_t n,const Matrix* matrix,unsigned k,bool quotient,const uint32_t* log,const uint32_t* powers){
 size_t index=size_t(blockIdx.x)*blockDim.x+threadIdx.x;if(index>=n/2)return;
 const auto& m=*matrix;const auto& p=m.taylor[k];size_t half=size_t(1)<<(m.split+k),j=index%half,offset=(index/half)*2*half;
 if(quotient){
  if(j>=half/2)return;
  uint32_t value=x[offset+half+j];
  for(unsigned t=0;t<p.count;t++){auto d=p.degree[t];if(d>j)value^=multiply(p.coef[t],x[offset+2*half+j-d],m.s,m.delta,log,powers);}
  x[offset+half+j]=value;
 }else{
  uint32_t value=x[offset+j];
  for(unsigned t=0;t<p.count;t++){auto d=p.degree[t];if(d<=j)value^=multiply(p.coef[t],x[offset+half+j-d],m.s,m.delta,log,powers);}
  x[offset+j]=value;
 }
}
__global__ void pointwise(uint32_t* a,const uint32_t* b,size_t n,unsigned s,uint32_t delta,const uint32_t* log,const uint32_t* powers){size_t i=size_t(blockIdx.x)*blockDim.x+threadIdx.x;if(i<n)a[i]=multiply(a[i],b[i],s,delta,log,powers);}
template<class F>double measured(F fn){
 cudaEvent_t begin,end;check(cudaEventCreate(&begin));check(cudaEventCreate(&end));check(cudaEventRecord(begin));fn();check(cudaGetLastError());check(cudaEventRecord(end));check(cudaEventSynchronize(end));float ms;check(cudaEventElapsedTime(&ms,begin,end));check(cudaEventDestroy(begin));check(cudaEventDestroy(end));return ms/1000.0;
}
using Clock=std::chrono::steady_clock;
double elapsed(Clock::time_point start){return std::chrono::duration<double>(Clock::now()-start).count();}
struct Engine {
 unsigned s,polynomial;uint32_t delta;Device log,powers,a,b,scratch;size_t points=0;
 std::unordered_map<size_t,std::unique_ptr<Device>> plans;
 std::vector<uint32_t> host_log,host_powers;
 uint64_t owned(){uint64_t bytes=log.bytes+powers.bytes+a.bytes+b.bytes+scratch.bytes;for(auto& kv:plans)bytes+=kv.second->bytes;return bytes;}
 void grow(size_t n){if(n>max_points||!n||(n&(n-1)))throw std::runtime_error("GPU shape bound");if(n<=points)return;
  if(owned()+3*n*4>device_cap)throw std::runtime_error("device allocation bound");Device next_a,next_b,next_scratch;next_a.allocate(n*4);next_b.allocate(n*4);next_scratch.allocate(n*4);a.swap(next_a);b.swap(next_b);scratch.swap(next_scratch);points=n;
 }
 void apply(uint32_t* data,size_t n,bool inverse,double* stats){
  auto it=plans.find(n);if(it==plans.end())throw std::runtime_error("unprepared GPU matrix shape");auto* p=static_cast<Matrix*>(it->second->p);
  unsigned dimension=0;for(size_t v=n;v>1;v>>=1)dimension++;unsigned split=dimension/2,v=dimension-split;
  auto convert=[&]{for(unsigned stage=0;stage<v;stage++){unsigned k=inverse?stage:v-1-stage;
   if(!inverse){stats[1]+=measured([&]{taylor_pass<<<(n/2+255)/256,256>>>(data,n,p,k,true,log.words(),powers.words());});stats[1]+=measured([&]{taylor_pass<<<(n/2+255)/256,256>>>(data,n,p,k,false,log.words(),powers.words());});}
   else{stats[1]+=measured([&]{taylor_pass<<<(n/2+255)/256,256>>>(data,n,p,k,false,log.words(),powers.words());});stats[1]+=measured([&]{taylor_pass<<<(n/2+255)/256,256>>>(data,n,p,k,true,log.words(),powers.words());});}
  }};
  auto columns=[&]{size_t jobs=size_t(1)<<split;stats[2]+=measured([&]{tiles<<<(jobs+127)/128,128>>>(data,n,p,true,inverse,log.words(),powers.words());});};
  auto rows=[&]{size_t length=size_t(1)<<split,jobs=n/length;
   stats[4]+=measured([&]{transpose<<<(n+255)/256,256>>>(data,scratch.words(),jobs,length);});
   stats[3]+=measured([&]{tiles<<<(jobs+127)/128,128>>>(scratch.words(),n,p,false,inverse,log.words(),powers.words());});
   stats[4]+=measured([&]{transpose<<<(n+255)/256,256>>>(scratch.words(),data,length,jobs);});
  };
  if(!inverse){convert();columns();rows();}else{rows();columns();convert();}
 }
};
struct Reader{const uint32_t* p;size_t size,offset=0;uint32_t word(){if(offset==size)throw std::runtime_error("matrix description truncated");return p[offset++];}};
void read_tile(Reader& r,Tile& p,unsigned expected,unsigned s){
 auto width=2*s;uint64_t q=uint64_t(1)<<width;unsigned count=r.word();if(count!=expected||count>levels)throw std::runtime_error("matrix basis shape");
 for(unsigned i=0;i<count;i++){p.basis[i]=r.word();if(!p.basis[i]||p.basis[i]>=q)throw std::runtime_error("basis coordinate");}
 for(unsigned k=0;k<count;k++){
  auto ts=r.word();if(!ts||ts>k+2||ts>terms)throw std::runtime_error("term bound");p.count[k]=ts;
  for(unsigned t=0;t<ts;t++){auto d=r.word(),c=r.word();if(!d||(d&(d-1))||d>(1u<<k)||c>=q||(t&&d<=p.degree[k][t-1]))throw std::runtime_error("linearized terms");p.degree[k][t]=d;p.coef[k][t]=c;}
  if(p.degree[k][ts-1]!=(1u<<k)||p.coef[k][ts-1]!=1)throw std::runtime_error("nonmonic matrix polynomial");
  for(unsigned bit=0;bit<width;bit++){p.values[k][bit]=r.word();if(p.values[k][bit]>=q)throw std::runtime_error("offset coordinate");}
  p.delta[k]=r.word();p.inverse[k]=r.word();if(!p.delta[k]||!p.inverse[k]||p.delta[k]>=q||p.inverse[k]>=q)throw std::runtime_error("butterfly coordinate");
 }
}
template<class F>int api(char* error,size_t capacity,F fn){try{fn();return 0;}catch(const std::exception& e){if(error&&capacity)std::snprintf(error,capacity,"%s",e.what());return 1;}}
}
extern "C" int gm_create(unsigned s,unsigned polynomial,unsigned delta,const uint32_t* logs,size_t lc,const uint32_t* powers,size_t pc,void** out,char* error,size_t capacity){return api(error,capacity,[&]{
 if(!out||!logs||!powers||s<2||s>16||lc!=(size_t(1)<<s)||pc!=2*(lc-1)||!delta||delta>=lc||polynomial>>s!=1)throw std::runtime_error("GPU field/table shape");
 auto e=std::make_unique<Engine>();e->s=s;e->polynomial=polynomial;e->delta=delta;check(cudaSetDevice(0));check(cudaFree(nullptr));
 e->host_log.assign(logs,logs+lc);e->host_powers.assign(powers,powers+pc);
 for(size_t i=1;i<lc;i++)if(logs[i]>=lc-1||powers[logs[i]]!=i)throw std::runtime_error("GPU lookup coordinate");
 e->log.allocate(lc*4);e->powers.allocate(pc*4);check(cudaMemcpy(e->log.p,logs,lc*4,cudaMemcpyHostToDevice));check(cudaMemcpy(e->powers.p,powers,pc*4,cudaMemcpyHostToDevice));*out=e.release();
});}
extern "C" int gm_add_plan(void* handle,const uint32_t* words,size_t count,char* error,size_t capacity){return api(error,capacity,[&]{
 if(!handle||!words)throw std::runtime_error("null GPU plan");auto& e=*static_cast<Engine*>(handle);Reader r{words,count};auto n=r.word(),split=r.word(),width=r.word();unsigned dim=0;for(auto v=n;v>1;v>>=1)dim++;
 if(n<4||n>max_points||(n&(n-1))||split!=dim/2||dim>2*e.s||width!=2*e.s||e.plans.count(n))throw std::runtime_error("GPU matrix shape or repeated plan");
 Matrix m{};m.s=e.s;m.delta=e.delta;m.dimension=dim;m.split=split;read_tile(r,m.rows,split+1,e.s);read_tile(r,m.columns,dim-split,e.s);if(r.offset!=count)throw std::runtime_error("matrix description trailing words");
 for(unsigned j=0;j<=split;j++)if(m.rows.basis[j]!=(1u<<j))throw std::runtime_error("row affine schedule requires canonical basis");
 auto multiply_host=[&](uint32_t a,uint32_t b){return multiply(a,b,e.s,e.delta,e.host_log.data(),e.host_powers.data());};
 for(unsigned k=0;k<dim-split;k++){auto& t=m.taylor[k];t.count=m.rows.count[split]-1;for(unsigned j=0;j<t.count;j++){t.degree[j]=m.rows.degree[split][j]<<k;auto c=m.rows.coef[split][j];for(unsigned z=0;z<k;z++)c=multiply_host(c,c);t.coef[j]=c;}}
 e.grow(n);if(e.owned()+sizeof(Matrix)>device_cap)throw std::runtime_error("device plan bound");auto device=std::make_unique<Device>();device->allocate(sizeof(Matrix));check(cudaMemcpy(device->p,&m,sizeof(m),cudaMemcpyHostToDevice));e.plans.emplace(n,std::move(device));
});}
extern "C" int gm_apply(void* handle,uint32_t* data,size_t n,int inverse,double* stats,char* error,size_t capacity){return api(error,capacity,[&]{
 if(!handle||!data||!stats||(inverse!=0&&inverse!=1))throw std::runtime_error("GPU apply arguments");auto& e=*static_cast<Engine*>(handle);if(!e.plans.count(n))throw std::runtime_error("GPU apply shape");std::fill(stats,stats+8,0);auto start=Clock::now(),copy=Clock::now();check(cudaMemcpy(e.a.p,data,n*4,cudaMemcpyHostToDevice));stats[0]=elapsed(copy);e.apply(e.a.words(),n,inverse,stats);copy=Clock::now();check(cudaMemcpy(data,e.a.p,n*4,cudaMemcpyDeviceToHost));stats[5]=elapsed(copy);stats[7]=elapsed(start);
});}
extern "C" int gm_convolution(void* handle,const uint32_t* a,const uint32_t* b,uint32_t* out,size_t n,double* stats,char* error,size_t capacity){return api(error,capacity,[&]{
 if(!handle||!a||!b||!out||!stats)throw std::runtime_error("GPU convolution arguments");auto& e=*static_cast<Engine*>(handle);if(!e.plans.count(n))throw std::runtime_error("GPU convolution shape");std::fill(stats,stats+8,0);auto start=Clock::now(),copy=Clock::now();check(cudaMemset(e.a.p,0,n*4));check(cudaMemset(e.b.p,0,n*4));check(cudaMemcpy(e.a.p,a,n/2*4,cudaMemcpyHostToDevice));check(cudaMemcpy(e.b.p,b,n/2*4,cudaMemcpyHostToDevice));stats[0]=elapsed(copy);e.apply(e.a.words(),n,false,stats);e.apply(e.b.words(),n,false,stats);stats[6]=measured([&]{pointwise<<<(n+255)/256,256>>>(e.a.words(),e.b.words(),n,e.s,e.delta,e.log.words(),e.powers.words());});e.apply(e.a.words(),n,true,stats);copy=Clock::now();check(cudaMemcpy(out,e.a.p,(n-1)*4,cudaMemcpyDeviceToHost));stats[5]=elapsed(copy);stats[7]=elapsed(start);
});}
extern "C" uint64_t gm_owned_bytes(void* handle){return handle?static_cast<Engine*>(handle)->owned():0;}
extern "C" void gm_destroy(void* handle){delete static_cast<Engine*>(handle);}
