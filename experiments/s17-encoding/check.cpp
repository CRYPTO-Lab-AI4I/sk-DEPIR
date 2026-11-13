//go:build ignore

// Synthetic polynomial schedule edges, not discovered PIR codes. Independent
// carryless evaluation/convolution checks K=1 and expanded2L>N arena shapes.
#include "wide_api.h"
#include <vector>
#include <cstring>
#include <iostream>
#include <algorithm>
#include <stdexcept>
uint32_t base(uint32_t a,uint32_t b){uint32_t x=0;for(unsigned j=0;j<2;j++)if(b&(1u<<j))x^=a<<j;for(int j=2;j>=2;j--)if(x&(1u<<j))x^=7u<<(j-2);return x;}
uint64_t mul(uint64_t x,uint64_t y){const auto a=uint32_t(x&3),b=uint32_t(x>>2),c=uint32_t(y&3),d=uint32_t(y>>2);return uint64_t(base(a,c)^base(2,base(b,d)))|(uint64_t(base(a,d)^base(b,c)^base(b,d))<<2);}
uint64_t power(uint64_t x,uint64_t n){uint64_t y=1;while(n){if(n&1)y=mul(y,x);x=mul(x,x);n>>=1;}return y;}
int main(){try{
 uint64_t omega=4;while(omega<16&&(power(omega,5)==1||power(omega,3)==1))omega++;
 if(omega==16)throw std::runtime_error("oracle primitive search");char error[512];size_t cases=0;
 for(size_t k:{size_t(1),size_t(3),size_t(7)}){
  std::vector<uint32_t> g(16-k),p(k+1),message(k);for(size_t j=0;j<g.size();j++)g[j]=k==1?1:uint32_t((j*3+1)&3);p[0]=1;
  for(size_t i=1;i<=k;i++)for(size_t j=1;j<=i&&j<g.size();j++)p[i]^=base(g[j],p[i-j]);for(size_t j=0;j<k;j++)message[j]=uint32_t((j*3+2)&3);
  std::vector<uint32_t> spectrum32(16);std::vector<uint64_t> spectrum64(16);
  for(size_t x=0;x<16;x++){uint64_t y=0;for(size_t j=g.size();j--;)y=mul(y,x)^g[j];spectrum32[x]=uint32_t(y);spectrum64[x]=y;}
  std::vector<uint32_t> hExpected(k),coefficients(16),want(16);
  for(size_t i=0;i<k;i++)for(size_t j=0;j<=i;j++)hExpected[i]^=base(p[j],message[i-j]);
  for(size_t i=0;i<k;i++)for(size_t j=0;j<g.size();j++)coefficients[i+j]^=base(hExpected[i],g[j]);
  uint64_t point=omega;for(size_t j=0;j<15;j++){want[point]=coefficients[j];want[0]^=coefficients[j];point=mul(point,omega);}
  size_t l=1;while(l<2*k-1)l*=2;const size_t arenaCount=std::max(2*l,size_t(16));
  liftrs_wide_online* encoder=nullptr;if(liftrs_wide_online_create(2,7,2,omega,k,&encoder,error,512))throw std::runtime_error(error);
  for(bool wide:{false,true})for(unsigned workers:{1u,4u}){
   const uint64_t guard=0xfacecafedeadbeefULL;std::vector<uint64_t> arena(arenaCount+2,guard),scratch(18,guard);std::vector<uint32_t> h(k+2,0xfaced00d);uint64_t stats[7];
   auto run=[&]{return liftrs_wide_online_encode(encoder,workers,message.data(),p.data(),wide?nullptr:spectrum32.data(),wide?spectrum64.data():nullptr,arena.data()+1,arenaCount,h.data()+1,k,scratch.data()+1,16,stats,error,512);};
   if(run())throw std::runtime_error(error);std::vector<uint32_t> got(16);std::memcpy(got.data(),scratch.data()+1,64);if(got!=want)throw std::runtime_error("polynomial schedule mismatch");
   if(arena.front()!=guard||arena.back()!=guard||scratch.front()!=guard||scratch.back()!=guard||h.front()!=0xfaced00d||h.back()!=0xfaced00d)throw std::runtime_error("wide encoding canary");
   if(wide){auto value=spectrum64[0];spectrum64[0]=uint64_t(1)<<32;if(!run())throw std::runtime_error("64-bit spectrum truncated");spectrum64[0]=value;if(run())throw std::runtime_error("spectrum recovery");}
   auto saved=message[0];message[0]=4;if(!run())throw std::runtime_error("base range accepted");message[0]=saved;if(run())throw std::runtime_error("input recovery");
   if(!liftrs_wide_online_encode(encoder,workers,message.data(),p.data(),wide?nullptr:spectrum32.data(),wide?spectrum64.data():nullptr,arena.data()+1,arenaCount,h.data()+1,k,arena.data()+1,16,stats,error,512))throw std::runtime_error("workspace alias accepted");
   cases++;
  }
  liftrs_wide_online_destroy(encoder);
 }
 liftrs_wide_online* bad=nullptr;if(!liftrs_wide_online_create(2,7,2,1,3,&bad,error,512)||bad)throw std::runtime_error("nonprimitive generator accepted");
 if(cases!=12)throw std::runtime_error("case count");std::cout<<"{\"cases\":"<<cases<<",\"match\":true}\n";
 }catch(const std::exception& e){std::cerr<<e.what()<<'\n';return 1;}}
