//go:build ignore
#include "online.hpp"
#include "file_io.hpp"
#include <cstring>
#include <iostream>

// Directly check arithmetic helpers at high-bit values, offset scatter and
// unaligned/multi-block file I/O independently of Go/cgo ownership.
int main() {
 try {
  liftrs::Field64 field;field.s=17;field.polynomial=0x20009;field.delta=1;field.q=uint64_t(1)<<34;
  const uint64_t coefficient=(uint64_t(1)<<33)|7;
  const auto fixed=liftrs::constant_product(field,coefficient);
  for(uint64_t i=0;i<4096;i++) {
   const auto x=((i<<22)^i)&(field.q-1);
   if(fixed.mul(field,x)!=field.mul(coefficient,x))throw std::runtime_error("high fixed product");
  }
  liftrs::Field f;f.s=2;f.polynomial=7;f.delta=2;f.q=16;
  if(!liftrs::is_primitive(f,uint32_t(4))||liftrs::is_primitive(f,uint32_t(1)))throw std::runtime_error("primitive check");
  auto step=liftrs::constant_product(f,uint32_t(4));
  std::vector<uint32_t> got(16),expected(16),sums(3);std::atomic<bool> bad{false};
  // Global coefficient offset5 is deliberately not chunk-aligned.
  uint32_t point=1;for(unsigned i=0;i<6;i++)point=f.mul(point,4);
  for(unsigned i=0;i<7;i++){expected[point]=(i*3)&3;point=f.mul(point,4);}
  liftrs::execution::scatter(f,uint32_t(4),step,5,7,4,3,
    [&](size_t i){return uint32_t((i*3)&3);},
    [&](uint32_t p,uint32_t v){got[p]=v;},
    [&](size_t j,uint32_t v){sums[j]=v;},bad);
  if(got!=expected||bad.load())throw std::runtime_error("offset scatter");
  auto* file=std::tmpfile();if(!file)throw std::runtime_error("tmpfile");
  const int fd=fileno(file);const size_t count=16389;
  std::vector<uint32_t> words(count),read(count);for(size_t i=0;i<count;i++)words[i]=uint32_t(i*17);
  const auto written=liftrs::io::write_words(fd,3,words.data(),count);
  std::vector<unsigned char> bytes(count*4);
  if(pread(fd,bytes.data(),bytes.size(),3)!=ssize_t(bytes.size()))throw std::runtime_error("raw read");
  for(size_t i=0;i<count;i++)for(unsigned j=0;j<4;j++)if(bytes[4*i+j]!=static_cast<unsigned char>(words[i]>>(8*j)))throw std::runtime_error("little endian words");
  const auto loaded=liftrs::io::read_words(fd,3,read.data(),count,uint64_t(1)<<32,4093);
  if(read!=words||written.bytes!=count*4||loaded.bytes!=count*4||loaded.calls<2)throw std::runtime_error("file chunks");
  bool rejected=false;try{liftrs::io::read_words(fd,3,read.data(),count,2);}catch(const std::exception&){rejected=true;}if(!rejected)throw std::runtime_error("file range");
  std::vector<uint32_t> oversized(count+1);rejected=false;try{liftrs::io::read_words(fd,3,oversized.data(),count+1,uint64_t(1)<<32);}catch(const std::exception&){rejected=true;}if(!rejected)throw std::runtime_error("file truncation");
  rejected=false;try{liftrs::io::write_words(fd,uint64_t(INT64_MAX),words.data(),1);}catch(const std::exception&){rejected=true;}if(!rejected)throw std::runtime_error("file overflow");
  std::fclose(file);
  // Exercise periodic work output without a30s sleep; production interval30s.
  setenv("LIFTRS_CPU_PROGRESS","1",1);
  liftrs::execution::WorkProgress progress("self-check-units",2,0);
  progress.advance();progress.advance();
  unsetenv("LIFTRS_CPU_PROGRESS");
  std::cout<<"{\"helpers\":true,\"high_values\":4096,\"file_words\":"<<count<<"}\n";
 } catch(const std::exception& e){std::cerr<<e.what()<<'\n';return 1;}
}
