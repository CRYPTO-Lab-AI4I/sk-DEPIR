#pragma once
// Experiment-only fixture reader. Never a public or untrusted artifact loader.
#include <fstream>
#include <vector>
#include <cstdint>
#include <stdexcept>
#include "../../encoding/liftrs-cpp/transform.hpp"
struct Fixture {
 std::ifstream in;std::vector<uint32_t> log,powers;liftrs::Transform transform;
 explicit Fixture(const char* path):in(path,std::ios::binary){
  if(word()!=0x3154504c)throw std::runtime_error("fixture magic");
  auto& f=transform.f;f.s=word();f.polynomial=word();f.delta=word();f.q=uint64_t(1)<<32;
  if(f.s!=16||f.polynomial!=0x1002d)throw std::runtime_error("fixture field");
  log=words(word());powers=words(word());
  if(log.size()!=65536||powers.size()!=131070)throw std::runtime_error("table shape");
  auto levels=word();if(levels!=8)throw std::runtime_error("tile levels");
  transform.terms.resize(levels);transform.values.resize(levels);
  transform.delta.resize(levels);transform.inverse.resize(levels);
  transform.delta_products.resize(levels);transform.inverse_products.resize(levels);
  for(unsigned k=0;k<levels;k++){
   auto count=word();if(!count||count>k+2)throw std::runtime_error("term shape");
   for(unsigned j=0;j<count;j++){auto degree=word(),c=word();transform.terms[k].push_back({degree,product(c)});}
   transform.values[k]=words(32);transform.delta[k]=word();transform.inverse[k]=word();
   transform.delta_products[k]=product(transform.delta[k]);transform.inverse_products[k]=product(transform.inverse[k]);
  }
 }
 uint32_t word(){uint32_t x;in.read(reinterpret_cast<char*>(&x),4);if(!in)throw std::runtime_error("fixture truncated");return x;}
 std::vector<uint32_t> words(size_t n){if(n>(1u<<24))throw std::runtime_error("fixture allocation cap");std::vector<uint32_t> v(n);in.read(reinterpret_cast<char*>(v.data()),n*4);if(!in)throw std::runtime_error("fixture truncated");return v;}
 liftrs::Product product(uint32_t c){liftrs::Product p;p.coefficient=c;if(c>1){p.rows.resize(128);for(unsigned j=0;j<128;j++)p.rows[j]=transform.f.mul(c,(j%16)<<(4*(j/16)));}return p;}
 uint32_t mul(uint32_t a,uint32_t b)const{return a&&b?powers[log[a]+log[b]]:0;}
};
