//go:build ignore

// Bounded trusted-fixture consumer, not a reusable code-data loader or ABI.
#include "transform.hpp"
#include <fstream>
#include <iostream>
#include <chrono>
#include <limits>
using namespace liftrs;
uint64_t read(std::istream& in) {
 uint64_t x=0;for(unsigned j=0;j<8;j++){int c=in.get();if(c==EOF)throw std::runtime_error("short fixture");x|=uint64_t(c)<<(8*j);}return x;
}
template<class Word> TransformT<Word> narrow(const Transform64& a) {
 TransformT<Word> b;b.f.s=a.f.s;b.f.polynomial=a.f.polynomial;b.f.delta=a.f.delta;b.f.q=a.f.q;
 b.basis.assign(a.basis.begin(),a.basis.end());b.values.resize(a.values.size());b.terms.resize(a.terms.size());
 b.delta.assign(a.delta.begin(),a.delta.end());b.inverse.assign(a.inverse.begin(),a.inverse.end());
 for(size_t k=0;k<a.terms.size();k++) {
  b.values[k].assign(a.values[k].begin(),a.values[k].end());
  for(const auto& t:a.terms[k]) {TermT<Word> v;v.degree=Word(t.degree);v.product.coefficient=Word(t.product.coefficient);v.product.rows.assign(t.product.rows.begin(),t.product.rows.end());b.terms[k].push_back(v);}
  ProductT<Word> d;d.coefficient=Word(a.delta_products[k].coefficient);d.rows.assign(a.delta_products[k].rows.begin(),a.delta_products[k].rows.end());b.delta_products.push_back(d);
  d.coefficient=Word(a.inverse_products[k].coefficient);d.rows.assign(a.inverse_products[k].rows.begin(),a.inverse_products[k].rows.end());b.inverse_products.push_back(d);
 }
 return b;
}
void tables(Transform64& a) {
 auto prepare=[&](Product64& p){p.rows.resize(((2*a.f.s+3)/4)*16);for(size_t j=0;j<p.rows.size()/16;j++)for(unsigned k=0;k<16;k++)p.rows[16*j+k]=a.f.mul(p.coefficient,(uint64_t(k)<<(4*j))&(a.f.q-1));};
 for(size_t k=0;k<a.terms.size();k++){for(auto& t:a.terms[k])prepare(t.product);prepare(a.delta_products[k]);prepare(a.inverse_products[k]);}
}
int main(int argc,char** argv){try {
 if(argc!=2)throw std::runtime_error("fixture path required");
 static_assert(sizeof(size_t)>=8,"wide native experiments require a64-bit host");
 std::ifstream in(argv[1],std::ios::binary);if(read(in)!=1)throw std::runtime_error("fixture version");
 Term64 wide_degree;wide_degree.degree=uint64_t(1)<<33;if(wide_degree.degree<=std::numeric_limits<uint32_t>::max())throw std::runtime_error("degree truncation");
 const auto count=read(in);if(count!=60)throw std::runtime_error("fixture count");
 for(uint64_t trial=0;trial<count;trial++) {
  Transform64 a;a.f.s=unsigned(read(in));a.f.polynomial=uint32_t(read(in));a.f.delta=uint32_t(read(in));
  const auto levels=read(in),start=read(in);
  if(a.f.s<2||a.f.s>20||levels<1||levels>10||levels>2*a.f.s)throw std::runtime_error("bounded shape");
  a.f.q=uint64_t(1)<<(2*a.f.s);if(start>=a.f.q)throw std::runtime_error("offset range");
  a.basis.resize(levels);for(auto& x:a.basis){x=read(in);if(x>=a.f.q)throw std::runtime_error("basis range");}
  a.terms.resize(levels);a.values.resize(levels);a.delta.resize(levels);a.inverse.resize(levels);a.delta_products.resize(levels);a.inverse_products.resize(levels);
  for(size_t k=0;k<levels;k++) {
   const auto terms=read(in);if(!terms||terms>k+2)throw std::runtime_error("term count");
   for(size_t j=0;j<terms;j++){Term64 t;t.degree=read(in);t.product.coefficient=read(in);if(t.degree>(uint64_t(1)<<k)||t.product.coefficient>=a.f.q)throw std::runtime_error("term range");a.terms[k].push_back(t);}
   if(a.terms[k].back().degree!=(uint64_t(1)<<k)||a.terms[k].back().product.coefficient!=1)throw std::runtime_error("nonmonic fixture");
   a.values[k].resize(2*a.f.s);for(auto& x:a.values[k]){x=read(in);if(x>=a.f.q)throw std::runtime_error("value range");}
   a.delta[k]=read(in);a.inverse[k]=read(in);if(!a.delta[k]||a.f.mul(a.delta[k],a.inverse[k])!=1)throw std::runtime_error("inverse factor");
   a.delta_products[k].coefficient=a.delta[k];a.inverse_products[k].coefficient=a.inverse[k];
  }
  const auto n=read(in);if(n!=(uint64_t(1)<<levels))throw std::runtime_error("length");
  std::vector<uint64_t> input(n),expected(n);for(auto& x:input){x=read(in);if(x>=a.f.q)throw std::runtime_error("input range");}for(auto& x:expected)x=read(in);
  for(bool cached:{false,true}) {
   if(cached)tables(a);
   // Guard both ends of the caller-owned transform range.
   const uint64_t canary=0xfacecafedeadbeefULL;std::vector<uint64_t> word(n+2,canary);std::copy(input.begin(),input.end(),word.begin()+1);
   auto begin=std::chrono::steady_clock::now();a.apply(word.data()+1,n,false,start);auto elapsed=std::chrono::duration_cast<std::chrono::nanoseconds>(std::chrono::steady_clock::now()-begin).count();
   if(!std::equal(expected.begin(),expected.end(),word.begin()+1))throw std::runtime_error("independent forward mismatch");
   if(a.f.s<=16){auto old=narrow<uint32_t>(a);std::vector<uint32_t> small(input.begin(),input.end());old.apply(small.data(),n,false,uint32_t(start));if(!std::equal(small.begin(),small.end(),expected.begin()))throw std::runtime_error("32/64 mismatch");}
   a.apply(word.data()+1,n,true,start);if(word.front()!=canary||word.back()!=canary||!std::equal(input.begin(),input.end(),word.begin()+1))throw std::runtime_error("inverse/canary mismatch");
   bool rejected=false;try{a.apply(word.data()+1,3,false,start);}catch(const std::runtime_error&){rejected=true;}if(!rejected)throw std::runtime_error("invalid length accepted");
   rejected=false;try{a.apply(word.data()+1,n,false,a.f.q);}catch(const std::runtime_error&){rejected=true;}if(!rejected)throw std::runtime_error("invalid offset accepted");
   std::cout<<"{\"s\":"<<a.f.s<<",\"n\":"<<n<<",\"tables\":"<<cached<<",\"forward_ns\":"<<elapsed<<",\"match\":true}\n";
  }
 }
 if(in.get()!=EOF)throw std::runtime_error("trailing fixture data");
 return 0;
 }catch(const std::exception& e){std::cerr<<e.what()<<'\n';return 1;}}
