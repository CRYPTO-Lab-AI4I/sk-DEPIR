//go:build ignore

// Direct C ABI sanitizer checks; expected evaluations use carryless long division.
#include "wide_api.h"
#include <vector>
#include <iostream>
#include <random>
#include <stdexcept>
uint32_t base(uint32_t a,uint32_t b,unsigned s,uint32_t p){uint64_t v=0;for(unsigned j=0;j<s;j++)if(b&(1u<<j))v^=uint64_t(a)<<j;for(int j=int(2*s-2);j>=int(s);j--)if(v&(uint64_t(1)<<j))v^=uint64_t(p)<<(j-s);return uint32_t(v);}
uint64_t mul(uint64_t x,uint64_t y,unsigned s,uint32_t p,uint32_t delta){const uint64_t mask=(uint64_t(1)<<s)-1;auto a=uint32_t(x&mask),b=uint32_t(x>>s),c=uint32_t(y&mask),d=uint32_t(y>>s);return uint64_t(base(a,c,s,p)^base(delta,base(b,d,s,p),s,p))|uint64_t(base(a,d,s,p)^base(b,c,s,p)^base(b,d,s,p))<<s;}
int main(){try{
 char error[512];std::mt19937_64 random(170056);size_t cases=0;
 for(unsigned s:{17u,20u}){
  const uint32_t p=s==17?0x20009:0x100009;const uint64_t q=uint64_t(1)<<(2*s);uint32_t delta=0;
  for(unsigned bit=0;bit<s;bit++){auto x=uint32_t(1)<<bit,tr=x;for(unsigned j=1;j<s;j++){x=base(x,x,s,p);tr^=x;}if(tr==1){delta=uint32_t(1)<<bit;break;}}
  if(!delta)throw std::runtime_error("delta oracle");
  for(unsigned dim=2;dim<=8;dim++)for(unsigned split=1;split<dim;split++){
   liftrs_wide_matrix* m=nullptr;if(liftrs_wide_matrix_create(s,p,delta,dim,split,&m,error,512))throw std::runtime_error(error);
   if(!liftrs_wide_matrix_bytes(m))throw std::runtime_error("metadata bytes");
   const size_t n=size_t(1)<<dim;std::vector<uint64_t> input(n),want(n),scratch(n);for(auto& v:input)v=random()&(q-1);
   for(size_t i=0;i<n;i++)for(size_t j=n;j--;)want[i]=mul(want[i],i,s,p,delta)^input[j];
   for(unsigned workers:{1u,4u}){
    const uint64_t canary=0xcafebabedeadfaceULL;std::vector<uint64_t> data(n+2,canary);std::copy(input.begin(),input.end(),data.begin()+1);
    uint64_t ns,stages[3];
    if(liftrs_wide_matrix_apply(m,data.data()+1,scratch.data(),n,workers,0,&ns,stages,error,512)||!std::equal(want.begin(),want.end(),data.begin()+1))throw std::runtime_error("wide matrix oracle");
    if(liftrs_wide_matrix_apply(m,data.data()+1,scratch.data(),n,workers,1,&ns,stages,error,512)||!std::equal(input.begin(),input.end(),data.begin()+1)||data.front()!=canary||data.back()!=canary)throw std::runtime_error("wide matrix inverse/canary");
    if(!liftrs_wide_matrix_apply(m,data.data(),data.data()+1,n,workers,0,&ns,stages,error,512))throw std::runtime_error("overlap accepted");
    data[1]=q;if(!liftrs_wide_matrix_apply(m,data.data()+1,scratch.data(),n,workers,0,&ns,stages,error,512))throw std::runtime_error("range accepted");
    cases++;
   }
   liftrs_wide_matrix_destroy(m);
  }
  uint64_t basis[]={1,1};liftrs_wide_plan* bad=nullptr;if(!liftrs_wide_plan_create(s,p,delta,basis,2,&bad,error,512)||bad)throw std::runtime_error("dependent basis accepted");liftrs_wide_plan_destroy(bad);
 }
 if(cases!=112)throw std::runtime_error("case count");std::cout<<"{\"matrix_cases\":"<<cases<<",\"match\":true}\n";
 }catch(const std::exception& e){std::cerr<<e.what()<<'\n';return 1;}}
