#include "fixture.hpp"
#include <chrono>
#include <iostream>
using V=std::vector<uint32_t>;
V karatsuba(const Fixture& f,const uint32_t* a,const uint32_t* b,size_t n){
 V out(2*n);
 if(n<=32){for(size_t i=0;i<n;i++)for(size_t j=0;j<n;j++)out[i+j]^=f.mul(a[i],b[j]);return out;}
 size_t h=n/2;V x(h),y(h);for(size_t i=0;i<h;i++){x[i]=a[i]^a[h+i];y[i]=b[i]^b[h+i];}
 auto z0=karatsuba(f,a,b,h),z2=karatsuba(f,a+h,b+h,h),z1=karatsuba(f,x.data(),y.data(),h);
 for(size_t i=0;i<n;i++){out[i]^=z0[i];out[n+i]^=z2[i];out[h+i]^=z0[i]^z1[i]^z2[i];}return out;
}
int main(int argc,char** argv){try{
 if(argc!=2)throw std::runtime_error("fixture path required");Fixture f(argv[1]);auto count=f.word();
 for(unsigned c=0;c<count;c++){
  auto n=f.word();if(!n||(n&(n-1))||n>32768)throw std::runtime_error("polynomial shape");
  auto a=f.words(n),b=f.words(n),want=f.words(2*n-1);
  for(unsigned repeat=0;repeat<3;repeat++){
   auto start=std::chrono::steady_clock::now();auto out=karatsuba(f,a.data(),b.data(),n);
   double seconds=std::chrono::duration<double>(std::chrono::steady_clock::now()-start).count();
   if(out.back()!=0||!std::equal(want.begin(),want.end(),out.begin()))throw std::runtime_error("native product mismatch");
   std::cout<<"{\"n\":"<<n<<",\"repeat\":"<<repeat<<",\"seconds\":"<<seconds<<",\"match\":true}\n";
  }
 }
}catch(const std::exception& e){std::cerr<<e.what()<<'\n';return 1;}}
