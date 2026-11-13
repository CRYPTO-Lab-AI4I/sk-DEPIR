//go:build liftrs_cpp && cgo

#include "wide_api.h"
#include "matrix.hpp"
#include <memory>
#include <cstdio>
#include <limits>

struct liftrs_wide_plan {
 liftrs::Transform64 transform;
 std::vector<std::vector<liftrs::Term64>> taylor_terms;
 size_t owned_bytes=0;
};
struct liftrs_wide_matrix {
 std::unique_ptr<liftrs_wide_plan> rows,columns;
 unsigned dim,split;
};
namespace {
void fail(char* error,size_t capacity,const char* text){if(error&&capacity)std::snprintf(error,capacity,"%s",text);}
void field(liftrs::Field64& f,unsigned s,uint32_t polynomial,uint32_t delta){
 if(sizeof(size_t)<8 || s<2 || s>20 || (polynomial>>s)!=1 || !(polynomial&1) || !delta || delta>=(1u<<s))throw std::runtime_error("invalid wide field");
 f.s=s;f.polynomial=polynomial;f.delta=delta;f.q=uint64_t(1)<<(2*s);
}
uint64_t power(const liftrs::Field64& f,uint64_t x,uint64_t n){uint64_t y=1;while(n){if(n&1)y=f.mul(y,x);n>>=1;if(n)x=f.mul(x,x);}return y;}
liftrs::Product64 product(const liftrs::Field64& f,uint64_t c){
 liftrs::Product64 p;p.coefficient=c;
 if(c>1){p.rows.resize(((2*f.s+3)/4)*16);for(size_t i=0;i<p.rows.size();i++)p.rows[i]=f.mul(c,(uint64_t(i%16)<<(4*(i/16)))&(f.q-1));}
 return p;
}
uint64_t evaluate(const liftrs::Field64& f,const std::vector<uint64_t>& p,uint64_t x){uint64_t out=0,v=x;for(auto c:p){out^=f.mul(c,v);v=f.mul(v,v);}return out;}
void account(liftrs_wide_plan& p){
 const auto& a=p.transform;
 size_t bytes=sizeof(p)+a.basis.capacity()*8+(a.delta.capacity()+a.inverse.capacity())*8;
 bytes+=a.terms.capacity()*sizeof(a.terms[0])+a.values.capacity()*sizeof(a.values[0]);
 bytes+=(a.delta_products.capacity()+a.inverse_products.capacity())*sizeof(liftrs::Product64);
 for(const auto& row:a.values)bytes+=row.capacity()*8;
 for(const auto& ts:a.terms){bytes+=ts.capacity()*sizeof(liftrs::Term64);for(const auto& t:ts)bytes+=t.product.rows.capacity()*8;}
 for(const auto& ps:{&a.delta_products,&a.inverse_products})for(const auto& x:*ps)bytes+=x.rows.capacity()*8;
 bytes+=p.taylor_terms.capacity()*sizeof(p.taylor_terms[0]);for(const auto& ts:p.taylor_terms){bytes+=ts.capacity()*sizeof(liftrs::Term64);for(const auto& t:ts)bytes+=t.product.rows.capacity()*8;}
 p.owned_bytes=bytes;
}
std::unique_ptr<liftrs_wide_plan> build(unsigned s,uint32_t polynomial,uint32_t delta,const uint64_t* basis,size_t levels){
 auto result=std::make_unique<liftrs_wide_plan>();auto& a=result->transform;field(a.f,s,polynomial,delta);
 if(!basis||!levels||levels>2*s)throw std::runtime_error("invalid wide basis length");
 a.basis.assign(basis,basis+levels);a.terms.resize(levels);a.values.resize(levels);a.delta.resize(levels);a.inverse.resize(levels);a.delta_products.resize(levels);a.inverse_products.resize(levels);
 std::vector<uint64_t> p{1};
 for(size_t k=0;k<levels;k++){
  if(!basis[k]||basis[k]>=a.f.q)throw std::runtime_error("wide basis range");
  for(size_t j=0;j<p.size();j++)if(p[j])a.terms[k].push_back({uint64_t(1)<<j,product(a.f,p[j])});
  a.values[k].resize(2*s);for(unsigned j=0;j<2*s;j++)a.values[k][j]=evaluate(a.f,p,uint64_t(1)<<j);
  const auto d=evaluate(a.f,p,basis[k]);if(!d)throw std::runtime_error("dependent wide basis");
  a.delta[k]=d;a.inverse[k]=power(a.f,d,a.f.q-2);
  if(a.f.mul(d,a.inverse[k])!=1)throw std::runtime_error("invalid wide field inverse");
  a.delta_products[k]=product(a.f,d);a.inverse_products[k]=product(a.f,a.inverse[k]);
  std::vector<uint64_t> next(p.size()+1);for(size_t j=0;j<p.size();j++){next[j+1]^=a.f.mul(p[j],p[j]);next[j]^=a.f.mul(d,p[j]);}p=std::move(next);
 }
 account(*result);return result;
}
void inputs(const liftrs::Field64& f,const uint64_t* data,size_t count){for(size_t i=0;i<count;i++)if(data[i]>=f.q)throw std::runtime_error("wide input outside field");}
uint64_t elapsed(std::chrono::steady_clock::time_point start){return std::chrono::duration_cast<std::chrono::nanoseconds>(std::chrono::steady_clock::now()-start).count();}
}
extern "C" int liftrs_wide_plan_create(unsigned s,uint32_t polynomial,uint32_t delta,const uint64_t* basis,size_t levels,liftrs_wide_plan** out,char* error,size_t capacity){
 if(out)*out=nullptr;try{if(!out)throw std::runtime_error("missing wide handle");*out=build(s,polynomial,delta,basis,levels).release();return 0;}catch(const std::exception& e){fail(error,capacity,e.what());return 1;}
}
extern "C" int liftrs_wide_plan_apply(const liftrs_wide_plan* p,uint64_t* data,size_t count,uint64_t start,int inverse,uint64_t* ns,char* error,size_t capacity){try{
 if(!p||!data||!ns||(inverse!=0&&inverse!=1)||!count||(count&(count-1))||count>(uint64_t(1)<<p->transform.terms.size())||start>=p->transform.f.q)throw std::runtime_error("invalid wide transform call");
 auto begin=std::chrono::steady_clock::now();inputs(p->transform.f,data,count);p->transform.apply(data,count,inverse!=0,start);*ns=elapsed(begin);return 0;
 }catch(const std::exception& e){fail(error,capacity,e.what());return 1;}}
extern "C" size_t liftrs_wide_plan_bytes(const liftrs_wide_plan* p){return p?p->owned_bytes:0;}
extern "C" void liftrs_wide_plan_destroy(liftrs_wide_plan* p){delete p;}
extern "C" int liftrs_wide_matrix_create(unsigned s,uint32_t polynomial,uint32_t delta,unsigned dim,unsigned split,liftrs_wide_matrix** out,char* error,size_t capacity){
 if(out)*out=nullptr;try{
 if(!out||!split||split>=dim||s<2||s>20||dim>2*s)throw std::runtime_error("invalid wide matrix dimensions");
 auto m=std::make_unique<liftrs_wide_matrix>();m->dim=dim;m->split=split;
 std::vector<uint64_t> low(split+1);for(unsigned j=0;j<=split;j++)low[j]=uint64_t(1)<<j;
 m->rows=build(s,polynomial,delta,low.data(),low.size());auto& r=m->rows->transform;
 std::vector<uint64_t> gamma(dim-split);for(unsigned j=0;j<dim-split;j++)gamma[j]=r.offset(split,uint64_t(1)<<(split+j));
 m->columns=build(s,polynomial,delta,gamma.data(),gamma.size());m->rows->taylor_terms.resize(dim-split);
 for(unsigned k=0;k<dim-split;k++)for(size_t j=0;j+1<r.terms[split].size();j++){
  const auto& t=r.terms[split][j];auto c=t.product.coefficient;for(unsigned b=0;b<k;b++)c=r.f.mul(c,c);
  m->rows->taylor_terms[k].push_back({t.degree<<k,product(r.f,c)});
 }
 account(*m->rows);*out=m.release();return 0;
 }catch(const std::exception& e){fail(error,capacity,e.what());return 1;}}
extern "C" int liftrs_wide_matrix_apply(const liftrs_wide_matrix* m,uint64_t* data,uint64_t* scratch,size_t count,unsigned workers,int inverse,uint64_t* ns,uint64_t* stages,char* error,size_t capacity){try{
 if(!m||!data||!scratch||!ns||!stages||!workers||workers>64||(inverse!=0&&inverse!=1)||count!=(uint64_t(1)<<m->dim))throw std::runtime_error("invalid wide matrix call");
 const auto bytes=count*sizeof(uint64_t),a=reinterpret_cast<uintptr_t>(data),b=reinterpret_cast<uintptr_t>(scratch);
 if(a>std::numeric_limits<uintptr_t>::max()-bytes||b>std::numeric_limits<uintptr_t>::max()-bytes||(a<b+bytes&&b<a+bytes))throw std::runtime_error("wide matrix buffers overlap");
 auto begin=std::chrono::steady_clock::now();inputs(m->rows->transform.f,data,count);
 liftrs::execution::matrix(false,count,m->rows.get(),m->columns.get(),data,scratch,count,m->split,workers,inverse,stages);*ns=elapsed(begin);return 0;
 }catch(const std::exception& e){fail(error,capacity,e.what());return 1;}}
extern "C" size_t liftrs_wide_matrix_bytes(const liftrs_wide_matrix* m){return m?sizeof(*m)+m->rows->owned_bytes+m->columns->owned_bytes:0;}
extern "C" void liftrs_wide_matrix_destroy(liftrs_wide_matrix* m){delete m;}
