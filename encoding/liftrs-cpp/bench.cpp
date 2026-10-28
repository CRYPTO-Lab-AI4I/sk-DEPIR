// Standalone batched-kernel / complete-FFT comparison. Shared little-endian
// fixtures come from the actual Go implementation; all file I/O is outside timing.
#include "transform.hpp"
#include <fstream>
#include <iostream>
#include <chrono>
#include <string>
#include <memory>
#include <thread>
using namespace liftrs;
uint32_t read32(std::istream& in){unsigned char b[4];if(!in.read((char*)b,4))throw std::runtime_error("short fixture");return uint32_t(b[0])|(uint32_t(b[1])<<8)|(uint32_t(b[2])<<16)|(uint32_t(b[3])<<24);}
Product read_product(std::istream& in){Product p;p.coefficient=read32(in);uint32_t n=read32(in);if(n>1024)throw std::runtime_error("bad row count");p.rows.resize(n);for(auto& x:p.rows)x=read32(in);return p;}
Transform read_transform(const std::string& path){std::ifstream in(path,std::ios::binary);Transform a;a.f.s=read32(in);a.f.polynomial=read32(in);a.f.delta=read32(in);a.f.q=read32(in);unsigned levels=read32(in);if(levels>26||a.f.s>13)throw std::runtime_error("bad field");a.terms.resize(levels);a.values.resize(levels);a.delta.resize(levels);a.inverse.resize(levels);a.delta_products.resize(levels);a.inverse_products.resize(levels);
 for(unsigned k=0;k<levels;k++){unsigned count=read32(in);if(count>27)throw std::runtime_error("bad terms");for(unsigned j=0;j<count;j++){Term t;t.degree=read32(in);t.product=read_product(in);a.terms[k].push_back(std::move(t));}a.values[k].resize(levels);for(auto& x:a.values[k])x=read32(in);a.delta[k]=read32(in);a.inverse[k]=read32(in);a.delta_products[k]=read_product(in);a.inverse_products[k]=read_product(in);}return a;}
int main(int argc,char** argv){try{
 if(argc!=8)throw std::runtime_error("usage: bench fixtures mode case pattern size output result");
 const std::string dir=argv[1],mode=argv[2],kind=argv[3],pattern=argv[4];size_t size=std::stoull(argv[5]);Field f;std::vector<Product> products;std::vector<uint32_t> input;Transform transform;
 if(mode=="kernel"){
  std::ifstream in(dir+"/"+kind+".bin",std::ios::binary);f.q=read32(in);unsigned count=read32(in),n=read32(in);if(count>1024||size>n)throw std::runtime_error("bad dimensions");for(unsigned i=0;i<count;i++)products.push_back(read_product(in));input.resize(size);for(auto& x:input)x=read32(in);
 }else{transform=read_transform(dir+"/transform.bin");std::ifstream in(dir+"/transform-input.bin",std::ios::binary);input.resize(size);for(auto& x:input)x=read32(in);}
 std::cout<<"READY"<<std::endl;
 std::this_thread::sleep_for(std::chrono::milliseconds(300));
 std::vector<uint32_t> output(size);unsigned passes=mode=="kernel"?(1u<<27)/size:1;std::vector<int64_t> times;
 for(int repeat=0;repeat<3;repeat++){
  std::copy(input.begin(),input.end(),output.begin());auto start=std::chrono::steady_clock::now();
  if(mode=="kernel"){for(unsigned pass=0;pass<passes;pass++)batch(products,f,input,output,pattern=="rotating");}else transform.apply(output,mode=="inverse");
  auto end=std::chrono::steady_clock::now();times.push_back(std::chrono::duration_cast<std::chrono::nanoseconds>(end-start).count());
 }
 // Save canonical uint32 output; x86-64/arm64 little-endian targets are measured.
 {std::ofstream out(argv[6],std::ios::binary);out.write((const char*)output.data(),output.size()*sizeof(uint32_t));if(!out)throw std::runtime_error("output write failed");}
 if(mode=="forward"){transform.apply(output,true);if(output!=input)throw std::runtime_error("FFT inverse mismatch");}
 size_t table_bytes=0;for(const auto& p:products)table_bytes+=4*p.rows.size();
 std::ofstream result(argv[7]);result<<"{\"language\":\"cpp\",\"mode\":\""<<mode<<"\",\"case\":\""<<kind<<"\",\"size\":"<<size<<",\"passes\":"<<passes<<",\"table_bytes\":"<<table_bytes<<",\"elapsed_ns\":[";
 for(size_t i=0;i<times.size();i++){if(i)result<<',';result<<times[i];}result<<"]}\n";
 }catch(const std::exception& e){std::cerr<<e.what()<<'\n';return 1;}}
