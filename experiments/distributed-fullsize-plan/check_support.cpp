// Bounded source audit only: compare retained support with the production
// Cartesian membership predicate. This does not select a new discovery algorithm.
#include "../../encoding/liftrs-cpp/monomial.hpp"
#include <cstdint>
#include <fstream>
#include <iostream>
#include <stdexcept>
#include <vector>
int main(int argc,char**argv){
 try{
  if(argc!=5)throw std::runtime_error("s t word-bytes support-path");
  unsigned s=std::stoul(argv[1]),t=std::stoul(argv[2]),width=std::stoul(argv[3]);
  if(s<2||s>10||t<1||(width!=4&&width!=8))throw std::runtime_error("bounded shape/width");
  const uint64_t q=uint64_t(1)<<s,n=q*q;
  std::ifstream file(argv[4],std::ios::binary);if(!file)throw std::runtime_error("support input");
  std::vector<unsigned char> support(n);uint64_t count=0,previous=0;
  while(true){unsigned char bytes[8]{};file.read(reinterpret_cast<char*>(bytes),width);if(file.gcount()==0&&file.eof())break;if(file.gcount()!=width)throw std::runtime_error("short support");
   uint64_t v=0;for(unsigned j=0;j<width;j++)v|=uint64_t(bytes[j])<<(8*j);
   if(v>=n-1||(count&&v<=previous))throw std::runtime_error("support order/range");support[v]=1;previous=v;count++;
  }
  uint64_t good=0,different=0,first=n;std::vector<uint64_t> owners(16),predicted(16);
  for(uint64_t v=0;v<n;v++){
   unsigned weights[20]{};for(unsigned j=0;j<s;j++)weights[j]=unsigned((v>>j)&1)+unsigned((v>>(s+j))&1);
   bool member=liftrs::preparation::good_weights(weights,s,t);good+=member;
   if(member!=bool(support[v])){different++;if(first==n)first=v;}
   if(support[v])owners[std::min(v,v*q%(n-1))%16]++;
   if(member)predicted[std::max(v%q,v/q)%16]++;
  }
  auto vector=[](const std::vector<uint64_t>&v){std::cout<<'[';for(size_t i=0;i<v.size();i++){if(i)std::cout<<',';std::cout<<v[i];}std::cout<<']';};
  std::cout<<"{\"s\":"<<s<<",\"t\":"<<t<<",\"n\":"<<n<<",\"support_count\":"<<count<<",\"cartesian_count\":"<<good<<",\"different\":"<<different<<",\"first_difference\":"<<first<<",\"actual_owners16\":";vector(owners);std::cout<<",\"cartesian_owners16\":";vector(predicted);std::cout<<"}\n";
 }catch(const std::exception&e){std::cerr<<e.what()<<'\n';return 1;}
}
