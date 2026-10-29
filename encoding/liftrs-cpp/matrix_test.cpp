//go:build ignore

// Standalone sanitizer fixture for the general-basis ABI and parallel matrix
// schedule. Coefficient evaluation by Horner is independent of the decomposition.
#include "c_api.h"
#include "transform.hpp"
#include <iostream>
#include <random>
uint32_t power(const liftrs::Field& f,uint32_t x,uint32_t n){uint32_t y=1;while(n){if(n&1)y=f.mul(y,x);x=f.mul(x,x);n>>=1;}return y;}
uint32_t evaluate(const liftrs::Field& f,const std::vector<uint32_t>& p,uint32_t x){uint32_t y=0;for(size_t j=0;j<p.size();j++)y^=f.mul(p[j],power(f,x,1u<<j));return y;}
// Sparse S_k preparation without tables; constants remain canonical field values.
std::vector<uint32_t> description(const liftrs::Field& f,const std::vector<uint32_t>& basis){
  unsigned width=0;for(auto q=f.q;q>1;q>>=1)width++;
  std::vector<uint32_t> out{2,f.s,f.polynomial,f.delta,f.q,uint32_t(basis.size())};out.insert(out.end(),basis.begin(),basis.end());
  std::vector<uint32_t> p{1};
  for(auto x:basis){
    unsigned count=0;for(auto c:p)if(c)count++;out.push_back(count);
    for(unsigned j=0;j<p.size();j++)if(p[j]){out.push_back(1u<<j);out.push_back(p[j]);out.push_back(0);}
    for(unsigned j=0;j<width;j++)out.push_back(evaluate(f,p,1u<<j));
    auto delta=evaluate(f,p,x),inverse=power(f,delta,f.q-2);if(!delta)throw std::runtime_error("dependent test basis");
    out.insert(out.end(),{delta,inverse,delta,0,inverse,0});
    std::vector<uint32_t> next(p.size()+1);for(unsigned j=0;j<p.size();j++){next[j+1]^=f.mul(p[j],p[j]);next[j]^=f.mul(delta,p[j]);}p=std::move(next);
  }return out;
}
int main(){
 liftrs::Field f;f.s=4;f.polynomial=0x13;f.q=256;
 for(f.delta=1;f.delta<16;f.delta<<=1){uint32_t trace=f.delta,x=trace;for(unsigned j=1;j<4;j++){x=f.base_mul(x,x);trace^=x;}if(trace==1)break;}
 std::mt19937 random(1504);char error[512];
 for(unsigned dim=2;dim<=8;dim++)for(unsigned split=1;split<dim;split++){
  std::vector<uint32_t> low(split+1);for(unsigned j=0;j<=split;j++)low[j]=1u<<j;
  std::vector<uint32_t> h{1};for(unsigned k=0;k<split;k++){auto d=evaluate(f,h,low[k]);std::vector<uint32_t> next(h.size()+1);for(unsigned j=0;j<h.size();j++){next[j+1]^=f.mul(h[j],h[j]);next[j]^=f.mul(d,h[j]);}h=std::move(next);}
  std::vector<uint32_t> gamma(dim-split);for(unsigned j=0;j<gamma.size();j++)gamma[j]=evaluate(f,h,1u<<(split+j));
  auto rw=description(f,low),cw=description(f,gamma);liftrs_plan *rows=nullptr,*columns=nullptr;
  if(liftrs_plan_create(rw.data(),rw.size(),&rows,error,sizeof(error)) || liftrs_plan_create(cw.data(),cw.size(),&columns,error,sizeof(error)))return 1;
  if(liftrs_plan_prepare_taylor(rows,dim-split,error,sizeof(error)))return 2;
  const size_t n=size_t(1)<<dim;std::vector<uint32_t> input(n),want(n),scratch(n);
  for(auto& x:input)x=random()%f.q;
  for(size_t point=0;point<n;point++)for(size_t j=n;j--;)want[point]=f.mul(want[point],point)^input[j];
  for(unsigned workers:{1u,4u}){auto data=input;uint64_t stages[3];
   if(liftrs_matrix_apply(rows,columns,data.data(),scratch.data(),n,split,workers,0,stages,error,sizeof(error)) || data!=want)return 3;
   if(liftrs_matrix_apply(rows,columns,data.data(),scratch.data(),n,split,workers,1,stages,error,sizeof(error)) || data!=input)return 4;
   if(!liftrs_matrix_apply(rows,columns,data.data(),scratch.data(),n,split,0,0,stages,error,sizeof(error)))return 5;
  }
  liftrs_plan_destroy(rows);liftrs_plan_destroy(columns);
 }
 std::cout<<"General-basis matrix/Taylor/Horner/inverse checks passed at all small splits\n";
}
