// Independent carryless convolution/long division checks for the native port.
#include "transform.hpp"
#include <iostream>
#include <random>
using namespace liftrs;
uint32_t oracle(uint32_t a,uint32_t b,unsigned s,uint32_t p){uint64_t product=0;for(unsigned i=0;i<s;i++)for(unsigned j=0;j<s;j++)product^=uint64_t((a>>i)&(b>>j)&1)<<(i+j);for(int i=int(2*s-2);i>=int(s);i--)if(product&(uint64_t(1)<<i))product^=uint64_t(p)<<(i-s);return uint32_t(product);}
uint32_t extension_oracle(const Field& f,uint32_t x,uint32_t y){uint32_t mask=(1u<<f.s)-1,a=x&mask,b=x>>f.s,c=y&mask,d=y>>f.s;auto mul=[&](uint32_t v,uint32_t w){return oracle(v,w,f.s,f.polynomial);};return (mul(a,c)^mul(f.delta,mul(b,d)))|((mul(a,d)^mul(b,c)^mul(b,d))<<f.s);}
int main(){const uint32_t polys[]={0,0,0x7,0xb,0x13,0x25,0x43,0x83,0x11d,0x211,0x409,0x805,0x1053,0x201b};std::mt19937 random(1313);
 for(unsigned s=2;s<=13;s++){Field f;f.s=s;f.polynomial=polys[s];uint32_t q=1u<<s;
  for(f.delta=1;f.delta<q;f.delta<<=1){uint32_t tr=f.delta,x=tr;for(unsigned i=1;i<s;i++){x=oracle(x,x,s,f.polynomial);tr^=x;}if(tr==1)break;}if(f.delta>=q)return 1;
  if(s<=8)for(uint32_t a=0;a<q;a++)for(uint32_t b=0;b<q;b++)if(f.base_mul(a,b)!=oracle(a,b,s,f.polynomial))return 2;
  for(bool extension:{false,true}){f.q=extension?q*q:q;for(int trial=0;trial<1000;trial++){
   uint32_t a=random()&(f.q-1),b=random()&(f.q-1);auto expected=[&](uint32_t x,uint32_t y){return extension?extension_oracle(f,x,y):oracle(x,y,s,f.polynomial);};if(f.mul(a,b)!=expected(a,b))return 3;
   Product p;p.coefficient=a;unsigned width=extension?2*s:s;p.rows.resize(((width+3)/4)*16);
   for(unsigned j=0;j<p.rows.size()/16;j++)for(unsigned n=0;n<16;n++)p.rows[16*j+n]=expected(a,(n<<(4*j))&(f.q-1));
   if(p.mul(f,b)!=expected(a,b))return 4;
  }}
 }
 std::cout<<"Independent base/extension/table checks passed through s=13\n";
}
