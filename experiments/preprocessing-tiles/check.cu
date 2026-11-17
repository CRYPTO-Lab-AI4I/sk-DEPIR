// Component driver includes the maintained executor and recurrences, not copied
// GPU math. New storage/split orchestration remains explicit and experimental.
#include "../../encoding/liftrs-cuda/preparation.cu"
#include "../../encoding/liftrs-cuda/tiled_preparation.cuh"
#include <filesystem>
#include <fstream>
#include <iostream>
#include <sstream>
#include <limits>
#include "../../encoding/liftrs-cuda/split_preparation.cuh"
using Timer = std::chrono::steady_clock;
double elapsed(Timer::time_point begin) {
  return std::chrono::duration<double>(Timer::now()-begin).count();
}
void require(bool ok,const char *message) {
  if(!ok)throw std::runtime_error(message);
}
struct Reader {
  std::ifstream file;
  explicit Reader(const char *path):file(path,std::ios::binary){
    require(bool(file),"fixture open");
  }
  template<class T>T scalar(){
    T v=0;
    for(unsigned j=0;j<sizeof(T);j++){
      int c=file.get();
      require(c!=EOF,"fixture truncated");
      v|=T(static_cast<unsigned char>(c))<<(8*j);
    }return v;
  }
  template<class T>std::vector<T> array(size_t n){
    require(n<=1u<<23,"fixture array cap");
    std::vector<T> a(n);
    for(auto &v:a)v=scalar<T>();
    return a;
  }
  void end(){
    require(file.get()==EOF,"fixture trailing data");
  }
};
struct FieldCase {
  unsigned s;
  uint32_t polynomial,delta;
  uint64_t omega;
  size_t n;
  std::vector<uint32_t> logs,powers;
  explicit FieldCase(Reader &r){
    s=r.scalar<uint32_t>();
    require(s==4||s==17,"fixture field");
    polynomial=r.scalar<uint32_t>();
    delta=r.scalar<uint32_t>();
    omega=r.scalar<uint64_t>();
    n=r.scalar<uint64_t>();
    require(n>=4&&n<=1u<<22&&!(n&(n-1))&&n<=(uint64_t(1)<<(2*s)),"fixture shape");
    size_t size=r.scalar<uint64_t>();
    require(size==size_t(1)<<s,"fixture logs");
    logs=r.array<uint32_t>(size);
    size=r.scalar<uint64_t>();
    require(size==2*((size_t(1)<<s)-1),"fixture powers");
    powers=r.array<uint32_t>(size);
  }
  template<class Word>std::unique_ptr<Resident<Word>> engine(size_t count)const{
    int ids[]={
      0,1,2,3
    };
    return std::make_unique<Resident<Word>>(s,polynomial,delta,omega,logs.data(),logs.size(),powers.data(),powers.size(),ids,count,n,128ull<<20);
  }
};
template<class T>std::string save(const std::filesystem::path &directory,const std::string &name,const std::vector<T> &data){
  // Explicit little-endian files let the parent hash actual output, outside clocks.
  std::ofstream out(directory/name,std::ios::binary);
  require(bool(out),"output open");
  const uint16_t one=1;
  if(*reinterpret_cast<const unsigned char *>(&one)==1)out.write(reinterpret_cast<const char *>(data.data()),std::streamsize(data.size()*sizeof(T)));
  else for(T v:data)for(unsigned k=0;k<sizeof(T);k++)out.put(char((v>>(8*k))&255));
  out.close();
  require(bool(out),"output write");
  return name;
}
std::string tiled_json(const TiledStats &s){
  std::ostringstream out;
  out.precision(12);
  out<<"\"seconds\":"<<s.seconds<<",\"validation_seconds\":"<<s.validation_seconds
  <<",\"pack_seconds\":"<<s.pack_seconds<<",\"h2d_seconds\":"<<s.h2d_seconds
  <<",\"kernel_wall_seconds\":"<<s.kernel_seconds<<",\"d2h_seconds\":"<<s.d2h_seconds
  <<",\"taylor_seconds\":"<<s.taylor_seconds<<",\"columns_seconds\":"<<s.columns_seconds
  <<",\"rows_seconds\":"<<s.rows_seconds<<",\"h2d_bytes\":"<<s.h2d_bytes
  <<",\"d2h_bytes\":"<<s.d2h_bytes<<",\"host_stage_bytes\":"<<s.host_stage_bytes
  <<",\"owned_device_bytes\":"<<s.owned_device_bytes<<",\"waves\":"<<s.waves;
  return out.str();
}
template<class Word>void transforms(const FieldCase &f,const std::vector<uint64_t>&input64,const std::vector<uint64_t>&expected64,
size_t index,bool small,const std::filesystem::path &directory){
  const unsigned repetitions=small?1:2;
  std::vector<Word> input(input64.begin(),input64.end()),expected(expected64.begin(),expected64.end());
  for(size_t j=0;j<f.n;j++)require(uint64_t(input[j])==input64[j]&&uint64_t(expected[j])==expected64[j],"fixture width");
  const std::vector<size_t> counts=f.n>=1u<<20?std::vector<size_t>{
    1,4
  }:std::vector<size_t>{
    1,2,4
  };
  const auto widths=f.n>=1u<<20?std::vector<size_t>{
    1u<<20,8u<<20
  }:std::vector<size_t>{
    1536,8192
  };
  for(size_t cards:counts){
    {
      const auto case_begin=Timer::now();
      const auto case_label="transform-"+std::to_string(index)+"-g"+std::to_string(cards)+"-resident";
      liftrs::execution::progress_record(case_label.c_str(),"component-case","started",0,2*(repetitions+1),0,"transforms");
      auto begin=Timer::now();
      auto engine=f.engine<Word>(cards);
      const double setup=elapsed(begin);
      auto data=input;
      begin=Timer::now();
      engine->apply(data.data(),f.n,false);
      double first=elapsed(begin);
      require(data==expected,"resident vs independent CPU");
      engine->apply(data.data(),f.n,true);
      require(data==input,"resident inverse");
      for(unsigned sample=0;sample<repetitions;sample++){
        data=input;
        begin=Timer::now();
        engine->apply(data.data(),f.n,false);
        double forward=elapsed(begin);
        require(data==expected,"resident forward equality");
        const auto prefix="transform-"+std::to_string(index)+"-g"+std::to_string(cards)+"-resident";
        std::string forward_file,inverse_file;
        if(sample+1==repetitions)forward_file=save(directory,prefix+"-forward.bin",data);
        begin=Timer::now();
        engine->apply(data.data(),f.n,true);
        double inverse=elapsed(begin);
        require(data==input,"resident inverse equality");
        if(sample+1==repetitions)inverse_file=save(directory,prefix+"-inverse.bin",data);
        std::cout<<"{\"kind\":\"transform\",\"backend\":\"resident\",\"index\":"<<index<<",\"s\":"<<f.s<<",\"n\":"<<f.n<<",\"gpus\":"<<cards<<",\"sample\":"<<sample
        <<",\"setup_seconds\":"<<setup<<",\"first_forward_seconds\":"<<first<<",\"forward_seconds\":"<<forward<<",\"inverse_seconds\":"<<inverse
        <<",\"owned_device_bytes\":"<<engine->owned()<<",\"h2d_bytes_each\":"<<f.n*sizeof(Word)<<",\"d2h_bytes_each\":"<<f.n*sizeof(Word)
        <<",\"forward_file\":\""<<forward_file<<"\",\"inverse_file\":\""<<inverse_file<<"\",\"matches\":true}"<<std::endl;
        if(sample+1==repetitions)liftrs::execution::progress_record(case_label.c_str(),"component-case","completed",2*(repetitions+1),2*(repetitions+1),uint64_t(elapsed(case_begin)*1000),"transforms");
      }
    }
    for(size_t tile:widths){
      const auto case_begin=Timer::now();
      const auto case_label="transform-"+std::to_string(index)+"-g"+std::to_string(cards)+"-tile"+std::to_string(tile);
      liftrs::execution::progress_record(case_label.c_str(),"component-case","started",0,2*(repetitions+1),0,"transforms");
      auto begin=Timer::now();
      auto engine=f.engine<Word>(cards);
      HostTiled<Word> backed(*engine,f.n,tile);
      double setup=elapsed(begin);
      auto data=input;
      begin=Timer::now();
      backed.apply(data.data(),false);
      double first=elapsed(begin);
      require(data==expected,"host-tiled vs independent CPU");
      backed.apply(data.data(),true);
      require(data==input,"host-tiled inverse");
      auto bad=input;
      bad[0]=Word(uint64_t(1)<<(2*f.s));
      // The narrow s=16 boundary is not represented by this fixture; s4/s17 fit.
      bool rejected=false;
      try{
        backed.apply(bad.data(),false);
      }catch(const std::exception&){
        rejected=true;
      }
      require(rejected&&bad[0]==Word(uint64_t(1)<<(2*f.s)),"tiled input error/recovery");
      for(unsigned sample=0;sample<repetitions;sample++){
        data=input;
        auto forward=backed.apply(data.data(),false);
        require(data==expected,"tiled forward equality after error");
        const auto prefix="transform-"+std::to_string(index)+"-g"+std::to_string(cards)+"-t"+std::to_string(tile);
        std::string forward_file,inverse_file;
        if(sample+1==repetitions)forward_file=save(directory,prefix+"-forward.bin",data);
        auto inverse=backed.apply(data.data(),true);
        require(data==input,"tiled inverse equality");
        if(sample+1==repetitions)inverse_file=save(directory,prefix+"-inverse.bin",data);
        std::cout<<"{\"kind\":\"transform\",\"backend\":\"host-tiled\",\"index\":"<<index<<",\"s\":"<<f.s<<",\"n\":"<<f.n<<",\"gpus\":"<<cards<<",\"tile_bytes\":"<<tile<<",\"sample\":"<<sample
        <<",\"setup_seconds\":"<<setup<<",\"first_forward_seconds\":"<<first<<",\"forward\":{"<<tiled_json(forward)<<"},\"inverse\":{"<<tiled_json(inverse)<<"}"
        <<",\"forward_file\":\""<<forward_file<<"\",\"inverse_file\":\""<<inverse_file<<"\",\"errors_checked\":true,\"matches\":true}"<<std::endl;
        if(sample+1==repetitions)liftrs::execution::progress_record(case_label.c_str(),"component-case","completed",2*(repetitions+1),2*(repetitions+1),uint64_t(elapsed(case_begin)*1000),"transforms");
      }
    }
  }
  auto engine=f.engine<Word>(1);
  bool rejected=false;
  try{
    HostTiled<Word> bad(*engine,f.n,1);
  }catch(const std::exception&){
    rejected=true;
  }require(rejected,"tiny tile admitted");
}
template<class Word>void products(const FieldCase&f,const std::vector<uint32_t>&a,const std::vector<uint32_t>&b,const std::vector<uint32_t>&expected,size_t block,size_t index,bool small,const std::filesystem::path &directory){
  require(a.size()+b.size()-1==f.n,"product extent");
  require((a.size()+block-1)/block==2&&(b.size()+block-1)/block==7,"two-by-seven shape");
  for(size_t cards: f.n>=1u<<20?std::vector<size_t>{1,4}:std::vector<size_t>{1,2,4})for(bool split:{false,true}){
    auto begin=Timer::now();
    auto engine=f.engine<Word>(cards);
    double setup=elapsed(begin);
    double first_call=0;
    {
      ProductStats warm;
      auto actual=product(*engine,a,b,split?block:0,512ull<<20,warm);
      require(actual==expected,"warm product CPU equality");
      first_call=warm.seconds;
    }
    const unsigned repetitions=small?1:2;
    for(unsigned sample=0;sample<repetitions;sample++){
      ProductStats stats;
      auto actual=product(*engine,a,b,split?block:0,512ull<<20,stats);
      require(actual==expected,"full/split vs independent CPU");
      require(stats.calls==(split?14:1),"subproduct count");
      std::string file;
      if(sample+1==repetitions)file=save(directory,"product-"+std::to_string(index)+"-g"+std::to_string(cards)+(split?"-split.bin":"-resident.bin"),actual);
      std::cout<<"{\"kind\":\"product\",\"backend\":\""<<(split?"split":"resident")<<"\",\"index\":"<<index<<",\"s\":"<<f.s<<",\"n\":"<<f.n<<",\"gpus\":"<<cards<<",\"sample\":"<<sample
      <<",\"setup_seconds\":"<<setup<<",\"first_call_seconds\":"<<first_call<<",\"seconds\":"<<stats.seconds<<",\"assembly_seconds\":"<<stats.assembly_seconds<<",\"subproducts\":"<<stats.calls<<",\"largest_domain\":"<<stats.largest_domain<<",\"owned_device_bytes\":"<<engine->owned()
      <<",\"temporary_host_bytes\":"<<stats.temporary_host_bytes<<",\"h2d_coefficient_bytes\":"<<stats.h2d_coeff_bytes<<",\"d2h_coefficient_bytes\":"<<stats.d2h_coeff_bytes<<",\"output_file\":\""<<file<<"\",\"matches\":true}"<<std::endl;
    }
  }
  auto engine=f.engine<Word>(1);
  for(auto invalid:std::vector<std::pair<size_t,uint64_t>>{{1,512ull<<20},{block,0}}){
    bool rejected=false;
    ProductStats stats;
    try{
      product(*engine,a,b,invalid.first,invalid.second,stats);
    }catch(const std::exception&){
      rejected=true;
    }
    require(rejected,"split host/block admission");
  }
}
void wide_origin(const FieldCase &f,const std::filesystem::path &fixture,const std::filesystem::path &directory){
  Reader r(fixture.string().c_str());
  const size_t length=r.scalar<uint64_t>(),first=r.scalar<uint64_t>();
  require(length==(1u<<17)&&first==length-1,"wide origin fixture");
  auto input=r.array<uint64_t>(length),expected=r.array<uint64_t>(length);
  r.end();
  int id=0;
  Resident<uint64_t> engine(17,f.polynomial,f.delta,f.omega,f.logs.data(),f.logs.size(),f.powers.data(),f.powers.size(),&id,1,1ull<<34,128ull<<20);
  auto began=Timer::now();
  progress("wide-affine-row","started",0,2,began);
  HostTiled<uint64_t> backed(engine,1ull<<34,1u<<20);
  auto actual=input;
  backed.rows_at(actual.data(),first,1,false);
  require(actual==expected,"wide affine row scalar oracle");
  const auto forward=save(directory,"origin-forward.bin",actual);
  backed.rows_at(actual.data(),first,1,true);
  require(actual==input,"wide affine row inverse");
  const auto inverse=save(directory,"origin-inverse.bin",actual);
  require(first*length>(1ull<<32),"coordinate witness too small");
  progress("wide-affine-row","completed",2,2,began);
  std::cout<<"{\"kind\":\"origin\",\"index\":0,\"s\":17,\"n\":131072,\"first_coordinate\":"<<first*length<<",\"forward_file\":\""<<forward<<"\",\"inverse_file\":\""<<inverse<<"\",\"matches\":true}"<<std::endl;
}
int main(int argc,char **argv){
  try{
    require(argc==4||argc==5,"transforms products output-dir [--small]");
    bool small=argc==5;
    require(!small||std::string(argv[4])=="--small","unknown flag");
    std::filesystem::path directory=argv[3];
    std::filesystem::create_directories(directory);
    std::cout.precision(12);
    Reader tr(argv[1]);
    require(tr.scalar<uint32_t>()==8,"transform count");
    for(size_t i=0;i<8;i++){
      FieldCase f(tr);
      auto input=tr.array<uint64_t>(f.n),expected=tr.array<uint64_t>(f.n);
      if(small&&f.n>4096)continue;
      if(i==4)wide_origin(f,std::filesystem::path(argv[1]).parent_path()/"origin.bin",directory);
      if(f.s<=16)transforms<uint32_t>(f,input,expected,i,small,directory);
      else transforms<uint64_t>(f,input,expected,i,small,directory);
    }tr.end();
    Reader pr(argv[2]);
    require(pr.scalar<uint32_t>()==3,"product count");
    for(size_t i=0;i<3;i++){
      FieldCase f(pr);
      size_t block=pr.scalar<uint64_t>();
      require(block>=4&&block==f.n/8,"block shape");
      size_t ac=pr.scalar<uint64_t>();
      require(ac>=1&&ac<=f.n,"product input extent");
      auto a=pr.array<uint32_t>(ac);
      size_t bc=pr.scalar<uint64_t>();
      require(bc>=1&&bc<=f.n,"product input extent");
      auto b=pr.array<uint32_t>(bc),expected=pr.array<uint32_t>(f.n);
      if(small&&f.n>4096)continue;
      if(f.s<=16)products<uint32_t>(f,a,b,expected,block,i,small,directory);
      else products<uint64_t>(f,a,b,expected,block,i,small,directory);
    }pr.end();
    std::cout<<"{\"kind\":\"completed\",\"small\":"<<(small?"true":"false")<<"}"<<std::endl;
    return 0;
  }catch(const std::exception&e){
    std::cerr<<"component failure: "<<e.what()<<std::endl;
    return 1;
  }
}
