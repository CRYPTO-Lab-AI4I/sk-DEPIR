//go:build ignore

// Exact full-matrix CPU/GPU comparisons and resident polynomial convolution.
#include "gpu_api.h"
#include "../../encoding/liftrs-cpp/c_api.h"
#include "../../encoding/liftrs-cpp/transform.hpp"
#include <fstream>
#include <iostream>
#include <iomanip>
#include <chrono>
#include <array>
#include <memory>
#include <string>

using Clock=std::chrono::steady_clock;
double elapsed(Clock::time_point start){return std::chrono::duration<double>(Clock::now()-start).count();}
struct Reader{
 std::ifstream in;explicit Reader(const char* path):in(path,std::ios::binary){}
 uint32_t word(){uint32_t v;in.read(reinterpret_cast<char*>(&v),4);if(!in)throw std::runtime_error("fixture truncated");return v;}
 std::vector<uint32_t> words(size_t n){if(n>(1u<<22))throw std::runtime_error("fixture cap");std::vector<uint32_t> v(n);in.read(reinterpret_cast<char*>(v.data()),n*4);if(!in)throw std::runtime_error("fixture truncated");return v;}
 std::vector<uint32_t> vector(){return words(word());}
};
void checked(int code,const char* error){if(code)throw std::runtime_error(error);}
struct EngineOwner{void* p=nullptr;~EngineOwner(){gm_destroy(p);}};
struct PlanOwner{liftrs_plan* p=nullptr;~PlanOwner(){liftrs_plan_destroy(p);}};
void print_stats(const double* stats){std::cout<<'[';for(unsigned i=0;i<8;i++){if(i)std::cout<<',';std::cout<<stats[i];}std::cout<<']';}
int main(int argc,char** argv){try{
 if(argc<2)throw std::runtime_error("fixture required");bool small=argc>2&&std::string(argv[2])=="--small";
 std::cout<<std::setprecision(10);Reader r(argv[1]);if(r.word()!=0x314d474c)throw std::runtime_error("fixture magic");auto s=r.word(),poly=r.word(),delta=r.word();auto logs=r.vector(),powers=r.vector();char error[512];EngineOwner gpu;
 auto start=Clock::now();checked(gm_create(s,poly,delta,logs.data(),logs.size(),powers.data(),powers.size(),&gpu.p,error,512),error);
 std::cout<<"{\"kind\":\"setup\",\"seconds\":"<<elapsed(start)<<",\"owned_device_bytes\":"<<gm_owned_bytes(gpu.p)<<"}\n";
 unsigned cases=r.word();for(unsigned ci=0;ci<cases;ci++){
  auto description=r.vector(),row_words=r.vector(),column_words=r.vector();if(description.size()<3)throw std::runtime_error("matrix header");size_t n=description[0];unsigned split=description[1];auto input=r.words(n),want=r.words(n);if(small&&n>256)continue;
  auto plan_start=Clock::now();checked(gm_add_plan(gpu.p,description.data(),description.size(),error,512),error);double gpu_setup=elapsed(plan_start);
  PlanOwner rows,columns;plan_start=Clock::now();checked(liftrs_plan_create(row_words.data(),row_words.size(),&rows.p,error,512),error);checked(liftrs_plan_create(column_words.data(),column_words.size(),&columns.p,error,512),error);unsigned dim=0;for(auto v=n;v>1;v>>=1)dim++;checked(liftrs_plan_prepare_taylor(rows.p,dim-split,error,512),error);
  std::cout<<"{\"kind\":\"plan\",\"n\":"<<n<<",\"gpu_setup_seconds\":"<<gpu_setup<<",\"cpu_setup_seconds\":"<<elapsed(plan_start)<<",\"owned_device_bytes\":"<<gm_owned_bytes(gpu.p)<<"}\n";
  std::vector<uint32_t> scratch(n);
  auto cpu_apply=[&](std::vector<uint32_t>& data,bool inverse,unsigned workers,uint64_t* stats){checked(liftrs_matrix_apply(rows.p,columns.p,data.data(),scratch.data(),n,split,workers,inverse,stats,error,512),error);};
  for(unsigned workers:{4u,16u})for(unsigned repeat=0;repeat<3;repeat++){
   auto data=input;uint64_t stats[3];start=Clock::now();cpu_apply(data,false,workers,stats);double forward=elapsed(start);if(data!=want)throw std::runtime_error("CPU forward mismatch");start=Clock::now();cpu_apply(data,true,workers,stats);double inverse=elapsed(start);if(data!=input)throw std::runtime_error("CPU inverse mismatch");
   std::cout<<"{\"kind\":\"cpu\",\"n\":"<<n<<",\"workers\":"<<workers<<",\"repeat\":"<<repeat<<",\"forward_seconds\":"<<forward<<",\"inverse_seconds\":"<<inverse<<",\"match\":true}\n";
  }
  for(unsigned repeat=0;repeat<3;repeat++){
   auto data=input;double forward[8],inverse[8];checked(gm_apply(gpu.p,data.data(),n,0,forward,error,512),error);if(data!=want)throw std::runtime_error("GPU forward mismatch");checked(gm_apply(gpu.p,data.data(),n,1,inverse,error,512),error);if(data!=input)throw std::runtime_error("GPU inverse mismatch");
   std::cout<<"{\"kind\":\"gpu\",\"n\":"<<n<<",\"repeat\":"<<repeat<<",\"forward\":";print_stats(forward);std::cout<<",\"inverse\":";print_stats(inverse);std::cout<<",\"match\":true}\n";
  }
  if(n>=1u<<18){
   std::vector<uint32_t> a(input.begin(),input.begin()+n/2),b(input.begin()+n/2,input.end()),expected(n-1);
   liftrs::Field field;field.s=s;field.polynomial=poly;field.delta=delta;field.q=uint64_t(1)<<(2*s);
   for(unsigned workers:{4u,16u}){
    start=Clock::now();std::vector<uint32_t> x(n),y(n);std::copy(a.begin(),a.end(),x.begin());std::copy(b.begin(),b.end(),y.begin());uint64_t stats[3];cpu_apply(x,false,workers,stats);cpu_apply(y,false,workers,stats);for(size_t i=0;i<n;i++)x[i]=field.mul(x[i],y[i]);cpu_apply(x,true,workers,stats);double seconds=elapsed(start);
    if(workers==4)std::copy(x.begin(),x.end()-1,expected.begin());else if(!std::equal(expected.begin(),expected.end(),x.begin()))throw std::runtime_error("CPU convolution mismatch");
    std::cout<<"{\"kind\":\"cpu-convolution\",\"n\":"<<n<<",\"workers\":"<<workers<<",\"seconds\":"<<seconds<<",\"match\":true}\n";
   }
   for(unsigned repeat=0;repeat<3;repeat++){
    std::vector<uint32_t> out(n-1);double stats[8];checked(gm_convolution(gpu.p,a.data(),b.data(),out.data(),n,stats,error,512),error);if(out!=expected)throw std::runtime_error("GPU resident convolution mismatch");
    std::cout<<"{\"kind\":\"gpu-convolution\",\"n\":"<<n<<",\"repeat\":"<<repeat<<",\"stats\":";print_stats(stats);std::cout<<",\"match\":true}\n";
   }
   auto zero=a;std::fill(zero.begin(),zero.end(),0);std::vector<uint32_t> out(n-1);double stats[8];checked(gm_convolution(gpu.p,zero.data(),b.data(),out.data(),n,stats,error,512),error);if(std::any_of(out.begin(),out.end(),[](uint32_t x){return x!=0;}))throw std::runtime_error("stale convolution padding");
   std::cout<<"{\"kind\":\"zero-convolution\",\"n\":"<<n<<",\"match\":true}\n";
  }
 }
 // Shapes and malformed descriptions fail before touching caller buffers.
 double stats[8];uint32_t value=7;if(!gm_apply(gpu.p,&value,3,0,stats,error,512))throw std::runtime_error("accepted invalid transform");
 uint32_t bad[3]={1024,1,32};if(!gm_add_plan(gpu.p,bad,3,error,512))throw std::runtime_error("accepted invalid plan");
 std::cout<<"{\"kind\":\"negative-checks\",\"match\":true,\"owned_device_bytes\":"<<gm_owned_bytes(gpu.p)<<"}\n";
}catch(const std::exception& e){std::cerr<<e.what()<<'\n';return 1;}}
