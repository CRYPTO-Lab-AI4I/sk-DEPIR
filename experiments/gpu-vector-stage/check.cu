// The adapter and GPU math are maintained code; CPU recursive transforms and
// wide sparse Horner use independent polynomial arithmetic, not GPU tables.
#include "../../encoding/liftrs-cuda/preparation.cu"
#include "../../encoding/liftrs-cuda/vector_stage.cuh"
#include <filesystem>
#include <fstream>
#include <iostream>
#include <sys/resource.h>
#include <ctime>
#include <thread>
#include <condition_variable>
using Clock = std::chrono::steady_clock;
void require(bool ok, const char *text) { if (!ok) throw std::runtime_error(text); }
struct Reader {
  std::ifstream file;
  explicit Reader(const char *path) : file(path, std::ios::binary) { require(bool(file), "fixture open"); }
  template<class T> T word() {
    T value=0;
    for(unsigned b=0;b<sizeof(T);b++) { int c=file.get(); require(c!=EOF,"fixture truncated"); value|=T(uint8_t(c))<<(8*b); }
    return value;
  }
  std::vector<uint32_t> table(size_t count) {
    require(word<uint32_t>()==count,"fixture table count");
    std::vector<uint32_t> out(count); for(auto &x:out)x=word<uint32_t>(); return out;
  }
};
uint64_t rss() {
  rusage u{}; require(!getrusage(RUSAGE_SELF,&u) && u.ru_maxrss < (1<<20),"1 GiB RSS guard");
  return uint64_t(u.ru_maxrss)*1024;
}
// Job-side observations are explicit heartbeats, never claimed completed work.
struct Watch {
  std::mutex mutex;
  std::condition_variable wake;
  bool stopped=false;
  std::string phase="startup";
  size_t done=0,total=0;
  Clock::time_point began=Clock::now();
  std::thread worker;
  Watch():worker([this]{
    std::unique_lock<std::mutex> lock(mutex);
    while(!wake.wait_for(lock,std::chrono::seconds(30),[&]{return stopped;})) {
      const auto ms=std::chrono::duration_cast<std::chrono::milliseconds>(Clock::now()-began).count();
      liftrs::execution::progress_record(phase.c_str(),"case-heartbeat","heartbeat",done,total,ms,"case steps (not new progress)");
      rusage u{};
      if(getrusage(RUSAGE_SELF,&u)||u.ru_maxrss>=(1<<20)) {
        std::fprintf(stderr,"vector stage RSS guard failed\n");std::fflush(stderr);std::_Exit(98);
      }
      std::fprintf(stderr,"{\"kind\":\"resource\",\"peak_rss_bytes\":%llu}\n",static_cast<unsigned long long>(u.ru_maxrss)*1024);
      std::fflush(stderr);
    }
  }){}
  ~Watch(){ {std::lock_guard<std::mutex> lock(mutex);stopped=true;}wake.notify_one();worker.join(); }
};
Watch *active_watch=nullptr;
void event(const std::string &phase, const char *kind, size_t done, size_t total, Clock::time_point began) {
  if(active_watch) {std::lock_guard<std::mutex> lock(active_watch->mutex);active_watch->phase=phase;active_watch->done=done;active_watch->total=total;active_watch->began=began;}
  char stamp[32];auto now=std::time(nullptr);std::strftime(stamp,sizeof(stamp),"%FT%TZ",std::gmtime(&now));
  std::cout<<"{\"utc\":\""<<stamp<<"\",\"phase\":\""<<phase<<"\",\"event\":\""<<kind
    <<"\",\"completed\":"<<done<<",\"total\":"<<total<<",\"elapsed_seconds\":"
    <<std::chrono::duration<double>(Clock::now()-began).count()<<",\"peak_rss_bytes\":"<<rss()<<"}"<<std::endl;
}
template<class Word>void save(const std::filesystem::path &p,const std::vector<Word>&a) {
  std::ofstream f(p,std::ios::binary);require(bool(f),"output open");
  for(Word value:a)for(unsigned b=0;b<sizeof(Word);b++)f.put(char(value>>(8*b)));
  f.close();require(bool(f),"output close");
}
template<class F>void rejects(F fn) { bool caught=false;try{fn();}catch(const std::exception&){caught=true;}require(caught,"invalid request accepted"); }
struct Field {
  unsigned s;uint32_t polynomial,delta;uint64_t omega;
  std::vector<uint32_t> logs,powers;
  explicit Field(Reader &r) {
    s=r.word<uint32_t>();require(s==4||s==17,"fixture field");
    polynomial=r.word<uint32_t>();delta=r.word<uint32_t>();omega=r.word<uint64_t>();
    logs=r.table(size_t(1)<<s);powers=r.table(2*((size_t(1)<<s)-1));
  }
};
template<class Word>void cases(const Field &f,const std::filesystem::path &out) {
  int id=0;
  for(unsigned dim: f.s==4?std::vector<unsigned>{4,5,8}:std::vector<unsigned>{7,12,34}) {
    const size_t n=size_t(1)<<dim, longest=size_t(1)<<((dim+1)/2),capacity=3*longest;
    const size_t admitted=std::min(n,capacity);
    Resident<Word> engine(f.s,f.polynomial,f.delta,f.omega,f.logs.data(),f.logs.size(),f.powers.data(),f.powers.size(),&id,1,n,128ull<<20);
    rejects([&]{VectorStage<Word> bad(engine,n,longest-1);});
    rejects([&]{VectorStage<Word> bad(engine,n,n+1);});
    const auto setup_begin=Clock::now();
    VectorStage<Word> stage(engine,n,admitted);
    const double setup=std::chrono::duration<double>(Clock::now()-setup_begin).count();
    const auto owned=engine.owned();require(owned<=128ull<<20,"owned cap");
    auto cpu=liftrs::preparation::build_matrix<liftrs::MatrixPlan<Word>>(f.s,f.polynomial,f.delta,dim,dim/2);
    for(bool columns:{false,true}) {
      const size_t length=stage.length(columns),jobs=n/length,count=dim==34?1:std::min<size_t>(3,jobs);
      for(size_t first:std::vector<size_t>{0,jobs-count}) {
        const auto label="s"+std::to_string(f.s)+"-d"+std::to_string(dim)+(columns?"-columns-":"-rows-")+std::to_string(first);
        const auto began=Clock::now(); event(label,"started",0,4,began);
        auto &plan=columns?cpu->columns->transform:cpu->rows->transform;
        std::vector<Word> input(count*length),expected(input.size());
        for(size_t j=0;j<count;j++) {
          std::vector<Word> vector(length);
          for(size_t k=0;k<(dim==34?4:length);k++)
            vector[k]=Word((uint64_t(k+1)*0x9e3779b9ull^((uint64_t(j+1)<<32)|uint64_t(j+1)))&((uint64_t(1)<<(2*f.s))-1));
          for(size_t k=0;k<length;k++)input[k*count+j]=vector[k];
          const Word origin=columns?Word(0):Word((first+j)*length);
          if(dim==34) {
            // At wide coordinates evaluate every point independently with a
            // degree-three Horner polynomial; no full grid or recursive oracle.
            for(size_t k=0;k<length;k++) {
              Word point=origin;
              for(size_t bit=0;bit<plan.basis.size();bit++)if((k>>bit)&1)point^=plan.basis[bit];
              Word v=0;for(size_t degree=4;degree--;)v=plan.f.mul(v,point)^vector[degree];
              expected[k*count+j]=v;
            }
          } else {
            plan.apply(vector.data(),length,false,origin);
            for(size_t k=0;k<length;k++)expected[k*count+j]=vector[k];
          }
          event(label,"oracle",j+1,count,began);
        }
        rejects([&]{stage.apply(columns,false,first,count);}); // no uploaded tile
        stage.upload(input.data(),input.size());
        const auto transform_begin=Clock::now();
        stage.apply(columns,false,first,count);
        const double forward=std::chrono::duration<double>(Clock::now()-transform_begin).count();
        std::vector<Word> actual(input.size());stage.download(actual.data(),actual.size());require(actual==expected,"forward CPU oracle mismatch");
        save(out/(label+"-expected.bin"),expected);save(out/(label+"-forward.bin"),actual);
        event(label,"checked-forward",1,4,began);
        // Invalid origin/extent/upload must preserve the resident forward tile.
        rejects([&]{stage.apply(columns,false,jobs,count);});
        rejects([&]{stage.apply(columns,false,first,0);});
        rejects([&]{stage.apply(columns,false,first,SIZE_MAX);});
        rejects([&]{stage.download(actual.data(),actual.size()-1);});
        auto bad=input;bad[0]=Word(uint64_t(1)<<(2*f.s));
        rejects([&]{stage.upload(bad.data(),bad.size());});
        rejects([&]{stage.upload(input.data(),admitted+1);});
        stage.download(actual.data(),actual.size());require(actual==expected,"rejection mutated tile");
        const auto inverse_begin=Clock::now();stage.apply(columns,true,first,count);
        const double inverse=std::chrono::duration<double>(Clock::now()-inverse_begin).count();
        stage.download(actual.data(),actual.size());require(actual==input,"resident inverse mismatch");
        save(out/(label+"-input.bin"),input);save(out/(label+"-inverse.bin"),actual);
        event(label,"checked-resident-inverse",2,4,began);
        stage.upload(expected.data(),expected.size());stage.apply(columns,true,first,count);stage.download(actual.data(),actual.size());
        require(actual==input,"inverse CPU oracle mismatch");
        stage.upload(input.data(),input.size());stage.apply(columns,false,first,count);stage.download(actual.data(),actual.size());
        require(actual==expected && engine.owned()==owned,"reuse mismatch or allocation growth");
        event(label,"completed",4,4,began);
        std::cout<<"{\"kind\":\"measurement\",\"case\":\""<<label<<"\",\"s\":"<<f.s<<",\"dimension\":"<<dim
          <<",\"columns\":"<<(columns?"true":"false")<<",\"first\":"<<first<<",\"count\":"<<count<<",\"length\":"<<length
          <<",\"word_bytes\":"<<sizeof(Word)<<",\"capacity_words\":"<<admitted<<",\"owned_device_bytes\":"<<owned
          <<",\"setup_seconds\":"<<setup<<",\"forward_kernel_wall_seconds\":"<<forward<<",\"inverse_kernel_wall_seconds\":"<<inverse
          <<",\"peak_rss_bytes\":"<<rss()<<",\"matches\":true}"<<std::endl;
        // Clear loaded state for the next case without allocating another tile.
        // A fresh stage is unnecessary: upload one word makes its extent invalid.
        stage.upload(input.data(),1);
      }
    }
  }
}
int main(int argc,char **argv) {
  try {
    Watch watch;active_watch=&watch;
    require(argc==3,"fields output-dir");std::filesystem::create_directories(argv[2]);std::cout.precision(12);
    Reader r(argv[1]);require(r.word<uint32_t>()==2,"field count");
    Field narrow(r);cases<uint32_t>(narrow,argv[2]);Field wide(r);cases<uint64_t>(wide,argv[2]);
    require(r.file.get()==EOF,"fixture trailing data");
    std::cout<<"{\"kind\":\"completed\",\"cases\":24}"<<std::endl;
    return 0;
  } catch(const std::exception &e) {std::cerr<<"vector stage failed: "<<e.what()<<std::endl;return 1;}
}
