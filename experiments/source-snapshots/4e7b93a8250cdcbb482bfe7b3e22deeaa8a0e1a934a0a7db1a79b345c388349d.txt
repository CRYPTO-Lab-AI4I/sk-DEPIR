//go:build ignore

// Native boundary safety checks use the shared Go-prepared transform plan but
// an independent Horner evaluation for the small borrowed-buffer transform.
#include "c_api.h"
#include "transform.hpp"
#include <fstream>
#include <iostream>
#include <vector>
uint32_t read_word(std::istream& in) {
 unsigned char b[4];if(!in.read(reinterpret_cast<char*>(b),4))throw std::runtime_error("short fixture");
 return uint32_t(b[0])|uint32_t(b[1])<<8|uint32_t(b[2])<<16|uint32_t(b[3])<<24;
}
int main(int argc,char** argv) {
 if(argc!=2)return 1;
 std::ifstream input(argv[1],std::ios::binary);std::vector<uint32_t> words{1};
 while(input.peek()!=EOF)words.push_back(read_word(input));
 char error[512];liftrs_plan* plan=nullptr;
 if(liftrs_plan_create(words.data(),words.size(),&plan,error,sizeof(error))) {std::cerr<<error<<'\n';return 2;}
 if(!liftrs_plan_owned_bytes(plan))return 3;
 std::vector<uint32_t> original(16),data(16),expected(16);liftrs::Field field;
 field.s=words[1];field.polynomial=words[2];field.delta=words[3];field.q=words[4];
 for(size_t i=0;i<data.size();i++)original[i]=data[i]=uint32_t(i*173+7)&(field.q-1);
 for(size_t point=0;point<data.size();point++)for(size_t j=data.size();j--;)expected[point]=field.mul(expected[point],uint32_t(point))^original[j];
 uint64_t ns=0;
 if(liftrs_plan_apply(plan,data.data(),data.size(),0,&ns,error,sizeof(error)) || data!=expected)return 4;
 if(liftrs_plan_apply(plan,data.data(),data.size(),1,&ns,error,sizeof(error)) || data!=original)return 5;
 if(!liftrs_plan_apply(plan,data.data(),3,0,&ns,error,sizeof(error)))return 6;
 if(!liftrs_plan_apply(plan,data.data(),data.size(),2,&ns,error,sizeof(error)))return 7;
 liftrs_plan_destroy(plan);liftrs_plan_destroy(nullptr);plan=nullptr;
 if(!liftrs_plan_apply(nullptr,data.data(),data.size(),0,&ns,error,sizeof(error)))return 8;
 if(!liftrs_plan_create(words.data(),words.size()-1,&plan,error,sizeof(error)))return 9;
 words.push_back(0);if(!liftrs_plan_create(words.data(),words.size(),&plan,error,sizeof(error)))return 10;
 std::cout<<"C ABI shape/error/ownership and borrowed-buffer Horner/round-trip checks passed\n";
}
