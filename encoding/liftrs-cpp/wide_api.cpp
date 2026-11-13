//go:build liftrs_cpp && cgo

#include "wide_api.h"
#include "matrix.hpp"
#include <memory>
#include <cstdio>
#include <limits>
#include <cstring>

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

struct liftrs_wide_online {
 std::unique_ptr<liftrs_wide_matrix> full,prefix;
 liftrs::Field64 field;
 liftrs::Product64 step;
 uint64_t omega,k,n,l;
 size_t owned_bytes;
};
namespace {
struct wide_range {uintptr_t begin,end;};
wide_range range(const void* p,size_t bytes){auto begin=reinterpret_cast<uintptr_t>(p);if(!p||begin>std::numeric_limits<uintptr_t>::max()-bytes)throw std::runtime_error("invalid wide memory range");return {begin,begin+bytes};}
bool overlaps(wide_range a,wide_range b){return a.begin<b.end&&b.begin<a.end;}
}
extern "C" int liftrs_wide_online_create(unsigned s,uint32_t polynomial,uint32_t delta,uint64_t omega,uint64_t k,liftrs_wide_online** out,char* error,size_t capacity){
 if(out)*out=nullptr;try{
 if(!out)throw std::runtime_error("missing wide encoder handle");
 auto e=std::make_unique<liftrs_wide_online>();field(e->field,s,polynomial,delta);e->n=e->field.q;e->k=k;e->omega=omega;
 if(!k||k>e->n/2||!omega||omega>=e->n||power(e->field,omega,e->n-1)!=1)throw std::runtime_error("invalid wide online shape/generator");
 auto remaining=e->n-1;for(uint64_t d=2;d<=remaining/d;d++)if(remaining%d==0){if(power(e->field,omega,(e->n-1)/d)==1)throw std::runtime_error("nonprimitive wide generator");while(remaining%d==0)remaining/=d;}
 if(remaining>1&&power(e->field,omega,(e->n-1)/remaining)==1)throw std::runtime_error("nonprimitive wide generator");
 e->l=1;unsigned prefix_dim=0;while(e->l<2*k-1){e->l*=2;prefix_dim++;}
 liftrs_wide_matrix* matrix=nullptr;
 if(liftrs_wide_matrix_create(s,polynomial,delta,2*s,s,&matrix,error,capacity))throw std::runtime_error(error);
 e->full.reset(matrix);
 if(e->l>1&&e->l!=e->n){matrix=nullptr;if(liftrs_wide_matrix_create(s,polynomial,delta,prefix_dim,std::max(1u,prefix_dim/2),&matrix,error,capacity))throw std::runtime_error(error);e->prefix.reset(matrix);}
 e->step=product(e->field,omega);e->owned_bytes=sizeof(*e)+e->step.rows.capacity()*8+liftrs_wide_matrix_bytes(e->full.get())+liftrs_wide_matrix_bytes(e->prefix.get());
 *out=e.release();return 0;
 }catch(const std::exception& ex){fail(error,capacity,ex.what());return 1;}}
extern "C" int liftrs_wide_online_encode(const liftrs_wide_online* e,unsigned workers,const uint32_t* message,const uint32_t* p,const uint32_t* spectrum32,const uint64_t* spectrum64,uint64_t* arena,size_t arena_count,uint32_t* h,size_t h_count,uint64_t* scratch,size_t scratch_count,uint64_t* stats,char* error,size_t capacity){try{
 if(!e||!stats||!workers||workers>64||!message||!p||!arena||!h||!scratch||arena_count!=std::max(2*e->l,e->n)||h_count!=e->k||scratch_count!=e->n||
   (bool(spectrum32)==bool(spectrum64)))throw std::runtime_error("invalid wide online buffers/width");
 const auto begin=std::chrono::steady_clock::now();std::fill(stats,stats+7,0);
 const wide_range writes[]={range(arena,arena_count*8),range(h,h_count*4),range(scratch,scratch_count*8)};
 const wide_range reads[]={range(message,e->k*4),range(p,(e->k+1)*4),range(spectrum32?static_cast<const void*>(spectrum32):spectrum64,e->n*(spectrum32?4:8))};
 for(size_t i=0;i<3;i++){for(size_t j=i+1;j<3;j++)if(overlaps(writes[i],writes[j]))throw std::runtime_error("wide online workspace overlap");for(auto r:reads)if(overlaps(writes[i],r))throw std::runtime_error("wide online input/workspace overlap");}
 auto measured=[&](unsigned index,auto fn){auto start=std::chrono::steady_clock::now();fn();stats[index]+=elapsed(start);};
 const uint64_t base_q=uint64_t(1)<<e->field.s;
 measured(4,[&]{
  if(p[0]!=1)throw std::runtime_error("wide P normalization");
  for(size_t i=0;i<e->k;i++)if(message[i]>=base_q||p[i]>=base_q)throw std::runtime_error("wide online base value range");
  if(p[e->k]>=base_q)throw std::runtime_error("wide P end range");
  for(size_t i=0;i<e->n;i++)if((spectrum32?uint64_t(spectrum32[i]):spectrum64[i])>=e->n)throw std::runtime_error("wide spectrum range");
  std::fill(arena,arena+2*e->l,0);for(size_t i=0;i<e->k;i++){arena[i]=p[i];arena[e->l+i]=message[i];}
 });
 auto fft=[&](bool prefix,uint64_t* data,bool inverse){
  if(prefix&&e->l==1)return;
  const auto* m=prefix&&e->prefix?e->prefix.get():e->full.get();uint64_t stages[3]={};
  liftrs::execution::matrix(false,prefix?e->l:e->n,m->rows.get(),m->columns.get(),data,scratch,prefix?e->l:e->n,m->split,workers,inverse,stages);
  for(unsigned j=0;j<3;j++)stats[1+j]+=stages[j];
 };
 const size_t chunk=4096;
 fft(true,arena,false);fft(true,arena+e->l,false);
 measured(5,[&]{liftrs::execution::parallel_jobs((e->l+chunk-1)/chunk,workers,[&](size_t job){for(size_t i=job*chunk;i<std::min(e->l,(job+1)*chunk);i++)arena[i]=e->field.mul(arena[i],arena[e->l+i]);});});
 fft(true,arena,true);
 measured(4,[&]{for(size_t i=0;i<e->k;i++){if(arena[i]>=base_q)throw std::runtime_error("wide prefix escaped base field");h[i]=uint32_t(arena[i]);}std::fill(arena,arena+e->n,0);for(size_t i=0;i<e->k;i++)arena[i]=h[i];});
 fft(false,arena,false);
 measured(5,[&]{liftrs::execution::parallel_jobs((e->n+chunk-1)/chunk,workers,[&](size_t job){for(size_t i=job*chunk;i<std::min(e->n,(job+1)*chunk);i++)arena[i]=e->field.mul(arena[i],spectrum32?uint64_t(spectrum32[i]):spectrum64[i]);});});
 fft(false,arena,true);
 measured(6,[&]{
  // Scratch is dead. Character access may alias its uint64 representation.
  // Use relaxed atomic byte stores: two output words can share one original
  // uint64 memory location even though primitive-generator scatter is bijective.
  // No typed uint64 access occurs until a later borrowed call reuses scratch.
  auto* output=reinterpret_cast<unsigned char*>(scratch);
  const size_t jobs=(e->n-1+chunk-1)/chunk;std::vector<uint32_t> parity(jobs);std::atomic<bool> bad{false};
  liftrs::execution::parallel_jobs(jobs,workers,[&](size_t job){auto point=power(e->field,e->omega,job*chunk+1);uint32_t sum=0;
   for(size_t i=job*chunk;i<std::min(e->n-1,(job+1)*chunk);i++){
    if(arena[i]>=base_q)bad.store(true,std::memory_order_relaxed);const auto value=uint32_t(arena[i]);const auto* bytes=reinterpret_cast<const unsigned char*>(&value);for(unsigned b=0;b<4;b++)__atomic_store_n(output+4*size_t(point)+b,bytes[b],__ATOMIC_RELAXED);sum^=value;point=e->step.mul(e->field,point);
   }parity[job]=sum;
  });
  if(bad.load(std::memory_order_relaxed))throw std::runtime_error("wide product escaped base field");
  uint32_t sum=0;for(auto x:parity)sum^=x;std::memcpy(output,&sum,4);
 });
 stats[0]=elapsed(begin);return 0;
 }catch(const std::exception& ex){fail(error,capacity,ex.what());return 1;}}
extern "C" size_t liftrs_wide_online_bytes(const liftrs_wide_online* e){return e?e->owned_bytes:0;}
extern "C" void liftrs_wide_online_destroy(liftrs_wide_online* e){delete e;}
