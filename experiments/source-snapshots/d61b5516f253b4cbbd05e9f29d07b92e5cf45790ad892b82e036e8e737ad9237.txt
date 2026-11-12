//go:build ignore

// Standalone sanitizer fixture for the general-basis ABI and parallel matrix
// schedule. Coefficient evaluation by Horner is independent of the decomposition.
#include "c_api.h"
#include "transform.hpp"
#include <iostream>
#include <random>
#include <cstring>
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
  liftrs::Field f; f.s=9; f.polynomial=0x211; f.q=1u<<(2*f.s);
  for (f.delta=1; f.delta<(1u<<f.s); f.delta<<=1) {
    uint32_t trace=f.delta,x=trace;
    for(unsigned j=1;j<f.s;j++){x=f.base_mul(x,x);trace^=x;}
    if(trace==1)break;
  }
  const size_t n=f.q;
  std::vector<uint32_t> low;for(unsigned j=0;j<=f.s;j++)low.push_back(1u<<j);
  std::vector<uint32_t> h{1};
  for(unsigned k=0;k<f.s;k++) {
    auto d=evaluate(f,h,1u<<k);std::vector<uint32_t> next(h.size()+1);
    for(unsigned j=0;j<h.size();j++){next[j+1]^=f.mul(h[j],h[j]);next[j]^=f.mul(d,h[j]);}
    h=std::move(next);
  }
  std::vector<uint32_t> gamma(f.s);
  for(unsigned j=0;j<f.s;j++)gamma[j]=evaluate(f,h,1u<<(f.s+j));
  auto rw=description(f,low),cw=description(f,gamma);
  liftrs_plan *rows=nullptr,*columns=nullptr;char error[512];
  if(liftrs_plan_create(rw.data(),rw.size(),&rows,error,sizeof(error)) ||
     liftrs_plan_create(cw.data(),cw.size(),&columns,error,sizeof(error)) ||
     liftrs_plan_prepare_taylor(rows,f.s,error,sizeof(error)))return 1;
  auto primitive=[&](uint32_t value){size_t order=n-1;for(size_t d=2;d<=order/d;d++){if(order%d)continue;if(power(f,value,(n-1)/d)==1)return false;while(order%d==0)order/=d;}return order==1||power(f,value,(n-1)/order)!=1;};
  uint32_t omega=2;while(!primitive(omega))omega++;
  auto prw=description(f,{1,2}),pcw=description(f,{evaluate(f,{1,1},2)});
  liftrs_plan *prefix_rows=nullptr,*prefix_columns=nullptr;
  if(liftrs_plan_create(prw.data(),prw.size(),&prefix_rows,error,sizeof(error))||liftrs_plan_create(pcw.data(),pcw.size(),&prefix_columns,error,sizeof(error))||liftrs_plan_prepare_taylor(prefix_rows,1,error,sizeof(error)))return 8;
  std::vector<uint32_t> spectrum(n,1),scratch(n),arena(2*n),saved(1),word(n),message{257},p{1,1};
  spectrum.back()=0;uint64_t stages[6];
  if(liftrs_matrix_apply(rows,columns,spectrum.data(),scratch.data(),n,f.s,4,0,stages,error,sizeof(error)))return 2;
  for(unsigned workers:{1u,4u}) {
    if(liftrs_online_encode(rows,columns,rows,columns,f.s,f.s,workers,omega,message.data(),p.data(),spectrum.data(),1,n,n,arena.data(),arena.size(),saved.data(),scratch.data(),word.data(),stages,error,sizeof(error))) {
      std::cerr<<error<<"\n";return 3;
    }
    for(auto x:word)if(x!=message[0])return 4;
    // The same array is scratch during transforms, output only after they join.
    if(liftrs_online_encode(rows,columns,rows,columns,f.s,f.s,workers,omega,message.data(),p.data(),spectrum.data(),1,n,n,arena.data(),arena.size(),saved.data(),scratch.data(),scratch.data(),stages,error,sizeof(error)))return 6;
    for(auto x:scratch)if(x!=message[0])return 7;
    uint64_t compact_stages[8];
    if(liftrs_online_encode_compact(scratch.size(),prefix_rows,prefix_columns,rows,columns,1,f.s,workers,omega,message.data(),p.data(),spectrum.data(),1,n,4,arena.data(),n,nullptr,scratch.data(),arena.data(),compact_stages,error,sizeof(error))){std::cerr<<error<<"\n";return 9;}
    for(size_t i=0;i<n;i++)if(arena[i]!=message[0])return 10;
    // Corrupt an extension coefficient above bit15: narrowing must not hide it.
    spectrum[0]^=1u<<16;
    int invalid=liftrs_online_encode_compact(scratch.size(),prefix_rows,prefix_columns,rows,columns,1,f.s,workers,omega,message.data(),p.data(),spectrum.data(),1,n,4,arena.data(),n,nullptr,scratch.data(),arena.data(),compact_stages,error,sizeof(error));
    spectrum[0]^=1u<<16;
    if(!invalid||!std::strstr(error,"escaped base field"))return 11;
  }
  if(!liftrs_online_encode(rows,columns,rows,columns,f.s,f.s,4,1,message.data(),p.data(),spectrum.data(),1,n,n,arena.data(),arena.size(),saved.data(),scratch.data(),word.data(),stages,error,sizeof(error)))return 5;
  liftrs_plan_destroy(prefix_rows);liftrs_plan_destroy(prefix_columns);
  liftrs_plan_destroy(rows);liftrs_plan_destroy(columns);
  std::cout<<"Complete online ABI, compact bounded columns and parallel scatter passed\n";
}
