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
  std::vector<uint32_t> out{2,f.s,f.polynomial,f.delta,uint32_t(f.q),uint32_t(basis.size())};out.insert(out.end(),basis.begin(),basis.end());
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
// Whole online ABI sanitizer fixture: P=1+X, G=sum X^j, K=1.
// Their product is 1+X^(N-1); every canonical output must equal the message.
int main() {
  liftrs::Field f; f.s=4; f.polynomial=0x13; f.q=256;
  for (f.delta=1; f.delta<16; f.delta<<=1) {
    uint32_t trace=f.delta,x=trace;
    for(unsigned j=1;j<4;j++){x=f.base_mul(x,x);trace^=x;}
    if(trace==1)break;
  }
  const size_t n=256;
  std::vector<uint32_t> low{1,2,4,8,16};
  std::vector<uint32_t> h{1};
  for(unsigned k=0;k<4;k++) {
    auto d=evaluate(f,h,1u<<k);std::vector<uint32_t> next(h.size()+1);
    for(unsigned j=0;j<h.size();j++){next[j+1]^=f.mul(h[j],h[j]);next[j]^=f.mul(d,h[j]);}
    h=std::move(next);
  }
  std::vector<uint32_t> gamma(4);
  for(unsigned j=0;j<4;j++)gamma[j]=evaluate(f,h,1u<<(4+j));
  auto rw=description(f,low),cw=description(f,gamma);
  liftrs_plan *rows=nullptr,*columns=nullptr;char error[512];
  if(liftrs_plan_create(rw.data(),rw.size(),&rows,error,sizeof(error)) ||
     liftrs_plan_create(cw.data(),cw.size(),&columns,error,sizeof(error)) ||
     liftrs_plan_prepare_taylor(rows,4,error,sizeof(error)))return 1;
  uint32_t omega=2;
  while(power(f,omega,255/3)==1 || power(f,omega,255/5)==1 || power(f,omega,255/17)==1)omega++;
  std::vector<uint32_t> spectrum(n,1),scratch(n),arena(2*n),saved(1),word(n),message{7},p{1,1};
  spectrum.back()=0;uint64_t stages[6];
  if(liftrs_matrix_apply(rows,columns,spectrum.data(),scratch.data(),n,4,4,0,stages,error,sizeof(error)))return 2;
  for(unsigned workers:{1u,4u}) {
    if(liftrs_online_encode(rows,columns,rows,columns,4,4,workers,omega,message.data(),p.data(),spectrum.data(),1,n,n,arena.data(),arena.size(),saved.data(),scratch.data(),word.data(),stages,error,sizeof(error))) {
      std::cerr<<error<<"\n";return 3;
    }
    for(auto x:word)if(x!=7)return 4;
    // The same array is scratch during transforms, output only after they join.
    if(liftrs_online_encode(rows,columns,rows,columns,4,4,workers,omega,message.data(),p.data(),spectrum.data(),1,n,n,arena.data(),arena.size(),saved.data(),scratch.data(),scratch.data(),stages,error,sizeof(error)))return 6;
    for(auto x:scratch)if(x!=7)return 7;

  }
  if(!liftrs_online_encode(rows,columns,rows,columns,4,4,4,1,message.data(),p.data(),spectrum.data(),1,n,n,arena.data(),arena.size(),saved.data(),scratch.data(),word.data(),stages,error,sizeof(error)))return 5;
  liftrs_plan_destroy(rows);liftrs_plan_destroy(columns);
  std::cout<<"Complete online ABI, parallel scatter and same-domain fixture passed\n";
}
