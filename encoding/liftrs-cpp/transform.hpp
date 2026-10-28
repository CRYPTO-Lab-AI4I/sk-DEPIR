#pragma once
// Portable polynomial-basis kernels matching liftrs-go. Tables and transform
// constants are trusted experiment fixtures, not a public cache format. There is
// no cryptographic stage or full PIR encoder here; arithmetic is not constant-time.
#include <cstdint>
#include <vector>
#include <stdexcept>
#include <algorithm>
namespace liftrs {
struct Field {
 uint32_t s=13, polynomial=0x201b, delta=1, q=1u<<26;
 uint32_t base_mul(uint32_t a,uint32_t b) const {
  uint32_t r=0; while(b){if(b&1)r^=a;b>>=1;a<<=1;if(a&(1u<<s))a^=polynomial;}return r;
 }
 uint32_t mul(uint32_t a,uint32_t b) const {
  if(q==(1u<<s))return base_mul(a,b);
  const uint32_t mask=(1u<<s)-1,aa=a&mask,ab=a>>s,ba=b&mask,bb=b>>s;
  const uint32_t ac=base_mul(aa,ba),bd=base_mul(ab,bb);
  return (ac^base_mul(delta,bd))|((base_mul(aa^ab,ba^bb)^ac)<<s);
 }
};
struct Product {
 uint32_t coefficient=0;std::vector<uint32_t> rows;
 uint32_t mul(const Field& f,uint32_t x) const {
  if(coefficient==0)return 0;if(coefficient==1)return x;
  if(rows.empty())return f.mul(coefficient,x);
  uint32_t y=0;for(size_t j=0;j<rows.size();j+=16){y^=rows[j+(x&15)];x>>=4;}return y;
 }
};
struct Term {uint32_t degree;Product product;};
struct Transform {
 Field f;std::vector<std::vector<Term>> terms;std::vector<std::vector<uint32_t>> values;
 std::vector<uint32_t> delta,inverse;std::vector<Product> delta_products,inverse_products;
 uint32_t offset(unsigned k,uint32_t x) const {
  uint32_t out=0;while(x){unsigned j=__builtin_ctz(x);out^=values[k][j];x&=x-1;}return out;
 }
 void recurse(uint32_t* data,size_t length,unsigned level,uint32_t start,bool invert) const {
  if(!level)return;const unsigned k=level-1;const size_t half=length/2;const uint32_t c=offset(k,start);
  const auto& ts=terms[k];
  if(!invert){
   for(size_t i=length;i-- > half;){uint32_t v=data[i];for(size_t j=0;j+1<ts.size();j++)data[i-half+ts[j].degree]^=ts[j].product.mul(f,v);}
   for(size_t j=0;j<half;j++){uint32_t v=data[half+j],left=data[j]^f.mul(c,v);data[j]=left;data[half+j]=left^delta_products[k].mul(f,v);}
  }
  recurse(data,half,k,start,invert);recurse(data+half,half,k,start^uint32_t(half),invert);
  if(invert){
   for(size_t j=0;j<half;j++){uint32_t v=inverse_products[k].mul(f,data[j]^data[half+j]);data[j]^=f.mul(c,v);data[half+j]=v;}
   for(size_t j=0;j<half;j++){uint32_t v=data[half+j];for(size_t n=0;n+1<ts.size();n++)data[j+ts[n].degree]^=ts[n].product.mul(f,v);}
  }
 }
 void apply(std::vector<uint32_t>& data,bool invert=false)const {
  if(data.empty()||(data.size()&(data.size()-1))||data.size()>f.q)throw std::runtime_error("invalid transform length");
  unsigned level=0;for(size_t n=data.size();n>1;n>>=1)level++;recurse(data.data(),data.size(),level,0,invert);
 }
};
#if defined(__GNUC__) || defined(__clang__)
__attribute__((noinline))
#endif
inline void batch(const std::vector<Product>& products,const Field& f,const std::vector<uint32_t>& in,std::vector<uint32_t>& out,bool rotating){
 for(size_t start=0;start<in.size();start+=4096){size_t end=std::min(start+4096,in.size());size_t index=rotating?(start/4096)%products.size():products.size()/2;const auto& p=products[index];for(size_t i=start;i<end;i++)out[i]=p.mul(f,in[i]);}
}
}
