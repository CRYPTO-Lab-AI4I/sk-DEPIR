//go:build liftrs_cpp && cgo

#include "c_api.h"
#include "transform.hpp"
#include "matrix.hpp"
#include <chrono>
#include <time.h>
#include <cstdio>
#include <memory>
#include <thread>
#include <atomic>
#include <cstring>
#include <unistd.h>
#include <cerrno>

struct liftrs_plan {
  liftrs::Transform transform;
  std::vector<std::vector<liftrs::Term>> taylor_terms; // Code-only maps prepared before encoding.
  size_t owned_bytes = 0; // Vector capacity payloads + descriptors, not allocator RSS.
};
namespace {
struct Reader {
  const uint32_t* data;
  size_t count, position = 0;
  uint32_t word() {
    if (position == count) throw std::runtime_error("short native plan");
    return data[position++];
  }
};
void fail(char* error, size_t capacity, const char* message) {
  if (error && capacity) std::snprintf(error, capacity, "%s", message);
}
liftrs::Product product(Reader& r, const liftrs::Field& f, unsigned width,
                         size_t& bytes) {
  liftrs::Product p;
  p.coefficient = r.word();
  const auto count = r.word();
  if (p.coefficient >= f.q || (count && count != ((width+3)/4)*16))
    throw std::runtime_error("invalid native multiplication map");
  p.rows.resize(count);
  for (auto& x : p.rows) {
    x = r.word();
    if (x >= f.q) throw std::runtime_error("native table value outside field");
  }
  bytes += p.rows.capacity()*sizeof(uint32_t);
  return p;
}
}
extern "C" int liftrs_plan_create(const uint32_t* words, size_t count,
    liftrs_plan** out, char* error, size_t capacity) {
  if (out) *out = nullptr;
  try {
    if (!out || !words || count > (1u<<20))
      throw std::runtime_error("invalid native plan size");
    Reader r{words,count};
    const auto version=r.word();
    if(version<1 || version>4) throw std::runtime_error("unsupported native plan version");
    auto plan = std::make_unique<liftrs_plan>();
    auto& a = plan->transform;
    a.f.s = r.word(); a.f.polynomial = r.word(); a.f.delta = r.word();
    a.f.q = r.word();
    if(version>=3) a.f.q |= uint64_t(r.word())<<32;
    const unsigned levels = r.word();
    unsigned width=0; for(auto q=a.f.q;q>1;q>>=1)width++;
    if (a.f.s < 2 || a.f.s > 16 || width>32 ||
        (width != a.f.s && width != 2*a.f.s) || a.f.q != (uint64_t(1)<<width) ||
        !levels || levels>width || ((version&1) && levels!=width) ||
        (a.f.polynomial >> a.f.s) != 1 || !(a.f.polynomial&1) ||
        a.f.delta >= (1u<<a.f.s))
      throw std::runtime_error("invalid native field description");
    a.terms.resize(levels); a.values.resize(levels);
    a.delta.resize(levels); a.inverse.resize(levels);
    a.delta_products.resize(levels); a.inverse_products.resize(levels);
    if(!(version&1)){a.basis.resize(levels);for(auto& x:a.basis){x=r.word();if(!x || x>=a.f.q)throw std::runtime_error("invalid ordered basis");}}
    auto& bytes = plan->owned_bytes;
    bytes = a.basis.capacity()*sizeof(uint32_t) + sizeof(liftrs_plan) + a.terms.capacity()*sizeof(a.terms[0]) +
      a.values.capacity()*sizeof(a.values[0]) +
      (a.delta.capacity()+a.inverse.capacity())*sizeof(uint32_t) +
      (a.delta_products.capacity()+a.inverse_products.capacity())*sizeof(liftrs::Product);
    for (unsigned k=0;k<levels;k++) {
      const unsigned terms = r.word();
      if (!terms || terms > k+2) throw std::runtime_error("invalid native terms");
      a.terms[k].reserve(terms);
      uint32_t previous = 0;
      for (unsigned j=0;j<terms;j++) {
        liftrs::Term t; t.degree = r.word();
        if (t.degree > (1u<<k) || (j && t.degree <= previous))
          throw std::runtime_error("invalid native term degree");
        previous = t.degree;
        t.product = product(r,a.f,width,bytes);
        a.terms[k].push_back(std::move(t));
      }
      const auto& leading = a.terms[k].back();
      if (leading.degree != (1u<<k) || leading.product.coefficient != 1)
        throw std::runtime_error("nonmonic native transform polynomial");
      bytes += a.terms[k].capacity()*sizeof(liftrs::Term);
      a.values[k].resize(width);
      for (auto& x : a.values[k]) {
        x = r.word();
        if (x >= a.f.q) throw std::runtime_error("invalid native offset value");
      }
      bytes += a.values[k].capacity()*sizeof(uint32_t);
      a.delta[k] = r.word(); a.inverse[k] = r.word();
      if (!a.delta[k] || !a.inverse[k] || a.delta[k]>=a.f.q || a.inverse[k]>=a.f.q)
        throw std::runtime_error("invalid native butterfly factor");
      a.delta_products[k] = product(r,a.f,width,bytes);
      a.inverse_products[k] = product(r,a.f,width,bytes);
    }
    if (r.position != count) throw std::runtime_error("trailing native plan data");
    *out = plan.release(); return 0;
  } catch (const std::exception& e) { fail(error,capacity,e.what()); return 1; }
    catch (...) { fail(error,capacity,"native initialization failed"); return 1; }
}
extern "C" int liftrs_plan_apply(const liftrs_plan* plan, uint32_t* data,
    size_t count, int inverse, uint64_t* kernel_ns, char* error, size_t capacity) {
  try {
    if (!plan || !data || !kernel_ns || (inverse!=0 && inverse!=1))
      throw std::runtime_error("invalid native transform call");
    const auto start = std::chrono::steady_clock::now();
    plan->transform.apply(data,count,inverse!=0);
    *kernel_ns = std::chrono::duration_cast<std::chrono::nanoseconds>(
      std::chrono::steady_clock::now()-start).count();
    return 0;
  } catch (const std::exception& e) { fail(error,capacity,e.what()); return 1; }
    catch (...) { fail(error,capacity,"native transform failed"); return 1; }
}
extern "C" size_t liftrs_plan_owned_bytes(const liftrs_plan* plan) {
  return plan ? plan->owned_bytes : 0;
}
extern "C" void liftrs_plan_destroy(liftrs_plan* plan) { delete plan; }

extern "C" int liftrs_plan_apply_at(const liftrs_plan* plan, uint32_t* data,
    size_t count, uint32_t offset, int inverse, uint64_t* ns, char* error, size_t capacity) {
  try {
    if(!plan || !data || !ns || (inverse!=0 && inverse!=1))throw std::runtime_error("invalid affine call");
    const auto start=std::chrono::steady_clock::now();
    plan->transform.apply(data,count,inverse!=0,offset);
    *ns=std::chrono::duration_cast<std::chrono::nanoseconds>(std::chrono::steady_clock::now()-start).count();return 0;
  }catch(const std::exception& e){fail(error,capacity,e.what());return 1;}
}
namespace {
// Each job owns disjoint output. Exceptions during thread creation join already
// launched workers; mathematical kernels cannot throw once shape is validated.
using liftrs::execution::parallel_jobs;
using liftrs::execution::transpose;
liftrs::Product make_product(const liftrs::Field& f,uint32_t c){
  liftrs::Product p;p.coefficient=c;
  if(c>1){unsigned width=0;for(auto q=f.q;q>1;q>>=1)width++;
    p.rows.resize(((width+3)/4)*16);
    for(size_t j=0;j<p.rows.size();j++)p.rows[j]=f.mul(c,uint32_t(j%16)<<(4*(j/16)));
  }return p;
}
uint32_t frobenius(const liftrs::Field& f,uint32_t c,unsigned k){while(k--)c=f.mul(c,c);return c;}

}
static int matrix_apply(bool bounded,size_t scratch_count,const liftrs_plan* rows,const liftrs_plan* columns,
    uint32_t* data,uint32_t* scratch,size_t count,unsigned split,unsigned workers,int inverse,uint64_t* stage_ns,char* error,size_t capacity){
  try{
    liftrs::execution::matrix(bounded,scratch_count,rows,columns,data,scratch,count,split,workers,inverse,stage_ns);
    return 0;
  }catch(const std::exception& e){fail(error,capacity,e.what());return 1;}
   catch(...){fail(error,capacity,"native matrix execution failed");return 1;}
}

extern "C" int liftrs_matrix_apply(const liftrs_plan* rows,const liftrs_plan* columns,uint32_t* data,uint32_t* scratch,size_t count,unsigned split,unsigned workers,int inverse,uint64_t* stages,char* error,size_t capacity){
 return matrix_apply(false,count,rows,columns,data,scratch,count,split,workers,inverse,stages,error,capacity);
}

extern "C" int liftrs_plan_prepare_taylor(liftrs_plan* plan,unsigned high_levels,char* error,size_t capacity){
  try{
    if(!plan || !high_levels || !plan->taylor_terms.empty())throw std::runtime_error("invalid/repeated Taylor preparation");
    const auto& h=plan->transform;unsigned split=h.terms.size()-1,width=0;for(auto q=h.f.q;q>1;q>>=1)width++;
    if(split+high_levels>width)throw std::runtime_error("Taylor domain exceeds field");
    std::vector<std::vector<liftrs::Term>> prepared(high_levels);
    size_t bytes=prepared.capacity()*sizeof(prepared[0]);
    for(unsigned k=0;k<high_levels;k++){
      auto& ts=prepared[k];const auto& source=h.terms[split];ts.reserve(source.size()-1);
      for(size_t j=0;j+1<source.size();j++)ts.push_back({source[j].degree<<k,make_product(h.f,frobenius(h.f,source[j].product.coefficient,k))});
      bytes+=ts.capacity()*sizeof(liftrs::Term);for(const auto& t:ts)bytes+=t.product.rows.capacity()*sizeof(uint32_t);
    }
    plan->taylor_terms=std::move(prepared);plan->owned_bytes+=bytes;return 0;
  }catch(const std::exception& e){fail(error,capacity,e.what());return 1;}
}

// Complete database-dependent cyclic encoding. All large arrays are borrowed
// for this call. arena aliases prefix x/z with the later full convolution;
// Fast execution saves H before clearing y; compact execution keeps its prefix.
// Constants remain immutable.
// Output may reuse scratch exactly: no transform reads scratch after assembly starts.
static int online_encode(bool compact,size_t scratch_count,int code_fd,uint64_t p_offset,uint64_t spectrum_offset,size_t tile_count,
    const liftrs_plan* pr, const liftrs_plan* pc,
    const liftrs_plan* fr, const liftrs_plan* fc,
    unsigned ps, unsigned fs, unsigned workers, uint32_t omega,
    const uint32_t* message, const uint32_t* p, const uint32_t* spectrum,
    size_t k, size_t n, size_t l, uint32_t* arena, size_t arena_count,
    uint32_t* h, uint32_t* scratch, uint32_t* word,
    uint64_t* stats, char* error, size_t capacity, bool packed=false) {
  try {
    if ((packed&&(compact||code_fd<0)) || !pr || !pc || !fr || !fc || !message || (code_fd<0&&(!p || !spectrum)) ||
        !arena || (!compact&&code_fd<0&&!h) || !scratch || !word || !stats ||
        !k || k > n/2 || !l || (l & (l-1)) || l < 2*k-1 || l > n ||
        arena_count < std::max(2*l, n) || ((compact||code_fd>=0)&&l>n/2) || (compact&&word!=arena) || !workers || workers > 64 ||
        fr->transform.f.q != n) {
      throw std::runtime_error("invalid online shape/buffers");
    }
    const auto& f = fr->transform.f;
    auto same = [&](const liftrs_plan* a) {
      return a->transform.f.q == f.q && a->transform.f.s == f.s &&
             a->transform.f.polynomial == f.polynomial &&
             a->transform.f.delta == f.delta;
    };
    if (!same(pr) || !same(pc) || !same(fc)) {
      throw std::runtime_error("online field mismatch");
    }
    auto power = [&](uint32_t x, size_t exponent) {
      uint32_t result = 1;
      while (exponent) {
        if (exponent & 1) result = f.mul(result, x);
        x = f.mul(x, x);
        exponent >>= 1;
      }
      return result;
    };
    // A primitive omega guarantees distinct scatter destinations, including when
    // chunks execute concurrently. Reject invalid generators before writing.
    if (!omega || omega >= n || power(omega, n-1) != 1) {
      throw std::runtime_error("invalid online generator");
    }
    size_t order = n-1;
    for (size_t d = 2; d*d <= order; d++) {
      if (order % d != 0) continue;
      if (power(omega, (n-1)/d) == 1) {
        throw std::runtime_error("nonprimitive online generator");
      }
      while (order % d == 0) order /= d;
    }
    if (order > 1 && power(omega, (n-1)/order) == 1) {
      throw std::runtime_error("nonprimitive online generator");
    }
    for (size_t i = 0; i < k; i++) {
      if (message[i] >= (1u << f.s) || (code_fd<0&&p[i] >= (1u << f.s))) {
        throw std::runtime_error("online input outside base field");
      }
    }
    if(code_fd>=0&&(p_offset>uint64_t(INT64_MAX)-uint64_t(k+1)*4||spectrum_offset>uint64_t(INT64_MAX)-uint64_t(n)*4))throw std::runtime_error("file offset overflow");
    if(code_fd>=0 && (compact||word!=scratch||!tile_count||tile_count>std::min(n,size_t(1)<<26)))throw std::runtime_error("file streaming shape/tile");
    std::fill(stats, stats+(code_fd>=0?11:compact?8:6), 0);
    auto measured = [&](unsigned stage, auto fn) {
      auto start = std::chrono::steady_clock::now();
      fn();
      stats[stage] += std::chrono::duration_cast<std::chrono::nanoseconds>(
          std::chrono::steady_clock::now()-start).count();
    };
    auto read_words=[&](uint32_t* dst,size_t count,uint64_t offset,uint64_t limit){
      static_assert(sizeof(off_t)>=8,"file encoding requires wide offsets");
      if(offset>uint64_t(INT64_MAX)||uint64_t(count)>((uint64_t(INT64_MAX)-offset)/4))throw std::runtime_error("file offset overflow");
      auto begin=std::chrono::steady_clock::now();size_t total=count*4,done=0;auto* bytes=reinterpret_cast<unsigned char*>(dst);
      while(done<total){ssize_t got=::pread(code_fd,bytes+done,std::min(total-done,size_t(1)<<22),off_t(offset+done));if(got<0&&errno==EINTR)continue;if(got<=0)throw std::runtime_error("code file read failed/truncated");done+=size_t(got);}
      for(size_t i=0;i<count;i++){
#if __BYTE_ORDER__ != __ORDER_LITTLE_ENDIAN__
        dst[i]=uint32_t(bytes[4*i])|(uint32_t(bytes[4*i+1])<<8)|(uint32_t(bytes[4*i+2])<<16)|(uint32_t(bytes[4*i+3])<<24);
#endif
        if(uint64_t(dst[i])>=limit)throw std::runtime_error("file code element outside field");
      }
      stats[8]+=std::chrono::duration_cast<std::chrono::nanoseconds>(std::chrono::steady_clock::now()-begin).count();stats[9]+=total;stats[10]++;
    };
    auto fft = [&](bool prefix, uint32_t* data, bool inverse) {
      uint64_t stages[3] = {};
      if (matrix_apply(compact,scratch_count,prefix ? pr : fr, prefix ? pc : fc,
          data, scratch, prefix ? l : n, prefix ? ps : fs, workers,
          inverse, stages, error, capacity)) {
        throw std::runtime_error(error && capacity ? error : "online matrix failed");
      }
      for (unsigned j = 0; j < 3; j++) stats[j] += stages[j];
    };
    auto* x = arena;
    auto* z = arena+l;
    measured(3, [&] {
      std::fill(arena, arena+2*l, 0);
      if(code_fd>=0){
        for(size_t first=0;first<k+1;first+=tile_count){size_t count=std::min(tile_count,k+1-first);read_words(x+first,count,p_offset+first*4,uint64_t(1)<<f.s);}
        if(x[0]!=1)throw std::runtime_error("file P normalization");x[k]=0; // P[K] is verified but excluded by the existing prefix algorithm.
      }else{std::copy(p, p+k, x);}
      std::copy(message, message+k, z);
    });
    fft(true, x, false);
    fft(true, z, false);
    constexpr size_t chunk = 65536;
    auto multiply = [&](uint32_t* a, const uint32_t* b, size_t count) {
      parallel_jobs((count+chunk-1)/chunk, workers, [&](size_t job) {
        for (size_t i = job*chunk; i < std::min(count, (job+1)*chunk); i++) {
          a[i] = f.mul(a[i], b[i]);
        }
      });
    };
    measured(4, [&] { multiply(x, z, l); });
    fft(true, x, true);
    measured(3, [&] {
      if(compact||code_fd>=0){std::fill(arena+k,arena+n,0);}else{
        std::copy(x, x+k, h);std::fill(arena, arena+n, 0);std::copy(h, h+k, arena);
      }
    });
    fft(false, arena, false);
    if(code_fd>=0){
      // Full forward has joined. Scratch is dead until inverse, so the file tile
      // needs no allocation. Each product joins before the next read reuses it.
      for(size_t first=0;first<n;first+=tile_count){size_t count=std::min(tile_count,n-first);read_words(scratch,count,spectrum_offset+first*4,f.q);measured(4,[&]{multiply(arena+first,scratch,count);});}
    }else{measured(4, [&] { multiply(arena, spectrum, n); });}
    fft(false, arena, true);
    auto* bytes=reinterpret_cast<unsigned char*>(arena);
    if(compact)measured(6,[&]{
      // Read each uint32 before writing its earlier byte representation. Forward
      // order cannot overwrite a future source; last coefficient is unused.
      for(size_t i=0;i<n;i++){
        uint32_t value=i<n-1?arena[i]:0;
        if(value>=(1u<<f.s))throw std::runtime_error("product escaped base field before compaction");
        bytes[2*i]=static_cast<unsigned char>(value);bytes[2*i+1]=static_cast<unsigned char>(value>>8);
      }
    });
    measured(5, [&] {
      // One small fixed-constant lookup map replaces scalar multiplication by
      // omega. Each chunk starts at omega^(first+1), so no N-sized map is needed.
      // Distinct field coordinates can share packed bytes. Atomic byte OR
      // joins disjoint bit fields without endian-dependent word stores.
      auto* packed_bytes=reinterpret_cast<unsigned char*>(word);
      auto put_packed=[&](size_t point,uint32_t value){
        size_t bit=point*f.s,first=bit/8;unsigned shift=bit%8;
        uint32_t bits=value<<shift;
        for(unsigned j=0;j<(shift+f.s+7)/8;j++)__atomic_fetch_or(packed_bytes+first+j,static_cast<unsigned char>(bits>>(8*j)),__ATOMIC_RELAXED);
      };
      if(packed)std::memset(packed_bytes,0,(n*f.s+7)/8);
      const auto step = make_product(f, omega);
      const size_t jobs = (n-1+chunk-1)/chunk;
      std::vector<uint32_t> parity(jobs);
      std::atomic<bool> bad{false};
      parallel_jobs(jobs, workers, [&](size_t job) {
        uint32_t point = power(omega, job*chunk+1), sum = 0;
        for (size_t i = job*chunk; i < std::min(n-1, (job+1)*chunk); i++) {
          const uint32_t value = compact ? uint32_t(bytes[2*i])|(uint32_t(bytes[2*i+1])<<8) : arena[i];
          if (value >= (1u << f.s)) bad.store(true, std::memory_order_relaxed);
          if(packed){if(value<(1u<<f.s))put_packed(point,value);}else if(compact){auto* out=bytes+2*n;out[2*size_t(point)]=static_cast<unsigned char>(value);out[2*size_t(point)+1]=static_cast<unsigned char>(value>>8);}else{word[point]=value;}
          sum ^= value;
          point = step.mul(f, point);
        }
        parity[job] = sum;
      });
      if (bad.load()) throw std::runtime_error("product escaped base field");
      uint32_t zero=0;for(auto value:parity)zero^=value;
      if(packed){put_packed(0,zero);}else if(compact){bytes[2*n]=static_cast<unsigned char>(zero);bytes[2*n+1]=static_cast<unsigned char>(zero>>8);}else{word[0]=zero;}
    });
    if(compact){
      measured(6,[&]{std::memcpy(bytes,bytes+2*n,2*n);});
      measured(7,[&]{for(size_t i=n;i-->0;)arena[i]=uint32_t(bytes[2*i])|(uint32_t(bytes[2*i+1])<<8);});
    }
    return 0;
  } catch (const std::exception& e) {
    fail(error, capacity, e.what());
    return 1;
  } catch (...) {
    fail(error, capacity, "native online encoding failed");
    return 1;
  }
}

extern "C" int liftrs_online_encode(const liftrs_plan* pr,const liftrs_plan* pc,const liftrs_plan* fr,const liftrs_plan* fc,unsigned ps,unsigned fs,unsigned workers,uint32_t omega,const uint32_t* message,const uint32_t* p,const uint32_t* spectrum,size_t k,size_t n,size_t l,uint32_t* arena,size_t arena_count,uint32_t* h,uint32_t* scratch,uint32_t* word,uint64_t* stats,char* error,size_t capacity){return online_encode(false,n,-1,0,0,0,pr,pc,fr,fc,ps,fs,workers,omega,message,p,spectrum,k,n,l,arena,arena_count,h,scratch,word,stats,error,capacity);}
extern "C" int liftrs_online_encode_compact(size_t scratch_count,const liftrs_plan* pr,const liftrs_plan* pc,const liftrs_plan* fr,const liftrs_plan* fc,unsigned ps,unsigned fs,unsigned workers,uint32_t omega,const uint32_t* message,const uint32_t* p,const uint32_t* spectrum,size_t k,size_t n,size_t l,uint32_t* arena,size_t arena_count,uint32_t* h,uint32_t* scratch,uint32_t* word,uint64_t* stats,char* error,size_t capacity){return online_encode(true,scratch_count,-1,0,0,0,pr,pc,fr,fc,ps,fs,workers,omega,message,p,spectrum,k,n,l,arena,arena_count,h,scratch,word,stats,error,capacity);}

// Explicit file-backed fast schedule: stats[11] add read/range-check ns, bytes,
// logical chunks. P reads are also inside buffer time; do not sum that twice.
extern "C" int liftrs_online_encode_file(int fd,uint64_t p_offset,uint64_t spectrum_offset,size_t tile,const liftrs_plan* pr,const liftrs_plan* pc,const liftrs_plan* fr,const liftrs_plan* fc,unsigned ps,unsigned fs,unsigned workers,uint32_t omega,const uint32_t* message,size_t k,size_t n,size_t l,uint32_t* arena,uint32_t* scratch,uint64_t* stats,char* error,size_t capacity){
 if(fd<0){fail(error,capacity,"invalid code descriptor");return 1;}
 return online_encode(false,n,fd,p_offset,spectrum_offset,tile,pr,pc,fr,fc,ps,fs,workers,omega,message,nullptr,nullptr,k,n,l,arena,n,nullptr,scratch,scratch,stats,error,capacity);
}

// Read the same steady_clock used by native stage timers, plus independent
// clocks for measurement diagnostics. These are observation only, not encoding.
extern "C" int liftrs_read_clocks(uint64_t* values, char* error, size_t capacity) {
  if (!values) { fail(error, capacity, "missing clock output"); return 1; }
  timespec monotonic{}, cpu{};
  if (clock_gettime(CLOCK_MONOTONIC, &monotonic) ||
      clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &cpu)) {
    fail(error, capacity, "clock_gettime failed"); return 1;
  }
  values[0] = std::chrono::duration_cast<std::chrono::nanoseconds>(
      std::chrono::steady_clock::now().time_since_epoch()).count();
  values[1] = std::chrono::duration_cast<std::chrono::nanoseconds>(
      std::chrono::system_clock::now().time_since_epoch()).count();
  values[2] = uint64_t(monotonic.tv_sec)*1000000000 + monotonic.tv_nsec;
  values[3] = uint64_t(cpu.tv_sec)*1000000000 + cpu.tv_nsec;
  return 0;
}

extern "C" int liftrs_online_encode_file_packed(int fd,uint64_t p_offset,uint64_t spectrum_offset,size_t tile,const liftrs_plan* pr,const liftrs_plan* pc,const liftrs_plan* fr,const liftrs_plan* fc,unsigned ps,unsigned fs,unsigned workers,uint32_t omega,const uint32_t* message,size_t k,size_t n,size_t l,uint32_t* arena,uint32_t* scratch,uint64_t* stats,char* error,size_t capacity){
 return online_encode(false,n,fd,p_offset,spectrum_offset,tile,pr,pc,fr,fc,ps,fs,workers,omega,message,nullptr,nullptr,k,n,l,arena,n,nullptr,scratch,scratch,stats,error,capacity,true);
}

#include "disk.hpp"
