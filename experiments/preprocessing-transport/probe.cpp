// Optional server experiment only: canonical uint64 patterns above bit32 are
// exchanged in RAM. MPI is not linked into ordinary client/library builds.
#include <mpi.h>
#include <sys/resource.h>
#include <chrono>
#include <ctime>
#include <iostream>
#include <stdexcept>
#include <string>
#include <vector>

using Clock = std::chrono::steady_clock;
int rank_id;
Clock::time_point began;
void check(int code) {
  if (code != MPI_SUCCESS) throw std::runtime_error("MPI operation failed");
}
void event(const char *phase, const char *kind, size_t done, size_t total,
           size_t bytes, double seconds = 0) {
  rusage usage{};
  if (getrusage(RUSAGE_SELF, &usage) || usage.ru_maxrss > (1<<20))
    throw std::runtime_error("per-rank RSS exceeds1GiB or sample unavailable");
  auto now = std::time(nullptr);
  char stamp[32]; std::strftime(stamp,sizeof(stamp),"%FT%TZ",std::gmtime(&now));
  std::cout << "{\"utc\":\"" << stamp << "\",\"rank\":" << rank_id
    << ",\"phase\":\"" << phase << "\",\"event\":\"" << kind
    << "\",\"completed\":" << done << ",\"total\":" << total
    << ",\"unit\":\"passes\",\"buffer_bytes_each\":" << bytes
    << ",\"elapsed_seconds\":" << std::chrono::duration<double>(Clock::now()-began).count()
    << ",\"exchange_seconds\":" << seconds << ",\"peak_rss_KiB\":" << usage.ru_maxrss
    << "}" << std::endl;
}
uint64_t value(int source, size_t index) {
  return (uint64_t(source) << 33) | uint64_t(index);
}
int main(int argc, char **argv) {
  MPI_Init(&argc,&argv);
  MPI_Comm_rank(MPI_COMM_WORLD,&rank_id);
  began = Clock::now();
  try {
    int count; check(MPI_Comm_size(MPI_COMM_WORLD,&count));
    if (count != 2) throw std::runtime_error("exactly two ranks required");
    check(MPI_Comm_set_errhandler(MPI_COMM_WORLD,MPI_ERRORS_RETURN));
    char host[MPI_MAX_PROCESSOR_NAME]{}; int length;
    check(MPI_Get_processor_name(host,&length));
    std::vector<char> hosts(2*MPI_MAX_PROCESSOR_NAME);
    check(MPI_Allgather(host,MPI_MAX_PROCESSOR_NAME,MPI_CHAR,hosts.data(),MPI_MAX_PROCESSOR_NAME,MPI_CHAR,MPI_COMM_WORLD));
    if (std::string(hosts.data()) == std::string(hosts.data()+MPI_MAX_PROCESSOR_NAME))
      throw std::runtime_error("ranks must run on different nodes");
    std::cout << "{\"rank\":" << rank_id << ",\"host\":\"" << host << "\"}" << std::endl;
    for (size_t bytes : {1u<<20,8u<<20,32u<<20}) {
      const size_t words=bytes/8, chunk=words/2;
      std::vector<uint64_t> send(words),receive(words);
      for (size_t i=0;i<words;i++) send[i]=value(rank_id,i);
      for (unsigned pass=0;pass<4;pass++) {
        event("redistribution","started",pass,4,bytes);
        check(MPI_Barrier(MPI_COMM_WORLD));
        const auto start=Clock::now();
        check(MPI_Alltoall(send.data(),int(chunk),MPI_UINT64_T,receive.data(),int(chunk),MPI_UINT64_T,MPI_COMM_WORLD));
        double seconds=std::chrono::duration<double>(Clock::now()-start).count(),worst;
        check(MPI_Allreduce(&seconds,&worst,1,MPI_DOUBLE,MPI_MAX,MPI_COMM_WORLD));
        for (int source=0;source<2;source++)
          for (size_t i=0;i<chunk;i++)
            if (receive[source*chunk+i]!=value(source,rank_id*chunk+i))
              throw std::runtime_error("field word/order mismatch");
        check(MPI_Alltoall(receive.data(),int(chunk),MPI_UINT64_T,send.data(),int(chunk),MPI_UINT64_T,MPI_COMM_WORLD));
        for (size_t i=0;i<words;i++)
          if (send[i]!=value(rank_id,i)) throw std::runtime_error("roundtrip mismatch");
        // Pass0 is warm-up; timing excludes validation and the reverse exchange.
        event(pass ? "timed-redistribution" : "warmup-redistribution","completed",pass+1,4,bytes,worst);
      }
    }
    event("workflow","completed",3,3,32u<<20);
    check(MPI_Finalize());
    return 0;
  } catch(const std::exception &error) {
    std::cerr << "utc_epoch=" << std::time(nullptr) << " phase=workflow event=failed rank="
      << rank_id << " error=" << error.what() << std::endl;
    MPI_Abort(MPI_COMM_WORLD,1);
    return 1;
  }
}
