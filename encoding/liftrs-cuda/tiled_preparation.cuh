#pragma once
// Experimental storage adapter, included after preparation.cu by component tools.
// It borrows a fresh resident executor's validated field/metadata/allocations;
// every arithmetic update still calls transform.cuh. No public ABI/default changes.
namespace {
#include "taylor_window.cuh"
struct TiledStats {
  double seconds = 0, validation_seconds = 0, pack_seconds = 0,
         h2d_seconds = 0, kernel_seconds = 0, d2h_seconds = 0,
         taylor_seconds = 0, columns_seconds = 0, rows_seconds = 0;
  uint64_t h2d_bytes = 0, d2h_bytes = 0, host_stage_bytes = 0,
           owned_device_bytes = 0, waves = 0;
};

// Single-caller borrowed executor: do not call/reconfigure it while this adapter
// is live. Caller keeps the full host domain; only bounded tiles live on device.
// Transfer/kernel phases deliberately join instead of promising overlap.
template <class Word> class HostTiled {
  using Clock = std::chrono::steady_clock;
  Resident<Word> &engine_;
  size_t n_, capacity_;
  Matrix<Word> metadata_;
  std::vector<std::vector<Word>> host_;
  std::vector<Matrix<Word> *> device_;
  TiledStats stats_;
  struct Ticket { size_t first, count, upload, download, block = 0, half = 0; };
  static double since(Clock::time_point start) {
    return std::chrono::duration<double>(Clock::now() - start).count();
  }
  void event(const std::string &label, const char *kind, size_t done, size_t total,
             Clock::time_point start) {
    engine_.event(label.c_str(), kind, done, total, start);
  }
  template <class Launch>
  void wave(const std::vector<Ticket> &tickets, Launch launch) {
    auto began = Clock::now();
    for (size_t d = 0; d < tickets.size(); d++) {
      auto &c = *engine_.cards_[d];
      check(cudaSetDevice(c.id));
      check(cudaMemcpy(c.x.data, host_[d].data(), tickets[d].upload * sizeof(Word), cudaMemcpyHostToDevice));
      stats_.h2d_bytes += tickets[d].upload * sizeof(Word);
    }
    stats_.h2d_seconds += since(began);
    began = Clock::now();
    for (size_t d = 0; d < tickets.size(); d++) {
      check(cudaSetDevice(engine_.cards_[d]->id));
      launch(d, tickets[d]);
      check(cudaGetLastError());
    }
    engine_.sync(unsigned(tickets.size()));
    stats_.kernel_seconds += since(began);
    began = Clock::now();
    for (size_t d = 0; d < tickets.size(); d++) {
      auto &c = *engine_.cards_[d];
      check(cudaSetDevice(c.id));
      check(cudaMemcpy(host_[d].data(), c.x.data, tickets[d].download * sizeof(Word), cudaMemcpyDeviceToHost));
      stats_.d2h_bytes += tickets[d].download * sizeof(Word);
    }
    stats_.d2h_seconds += since(began);
    stats_.waves++;
  }
  void taylor_pass(Word *data, unsigned level, bool quotient) {
    const auto &terms = metadata_.taylor[level];
    if (!terms.count) return;
    const size_t half = size_t(1) << (metadata_.split + level);
    size_t minimum = half, maximum = 0;
    for (unsigned j = 0; j < terms.count; j++) {
      minimum = std::min(minimum, size_t(terms.degree[j]));
      maximum = std::max(maximum, size_t(terms.degree[j]));
    }
    if (!minimum || maximum > half / 2)
      throw std::runtime_error("tiled Taylor independence bound");
    const size_t begin = quotient ? 0 : minimum, end = quotient ? maximum : half;
    const size_t total = (n_ / (2 * half)) * (end - begin);
    const size_t span = capacity_ / (terms.count + 1);
    if (!span) throw std::runtime_error("tiled Taylor staging cap");
    const auto began = Clock::now(); auto last = began; size_t completed = 0;
    const std::string label = std::string("host-Taylor-") + (quotient ? "quotient-" : "remainder-") + std::to_string(level);
    event(label, "started", 0, total, began);
    for (size_t block = 0; block < n_; block += 2 * half) {
      for (size_t first = begin; first < end;) {
        std::vector<Ticket> tickets;
        auto packing = Clock::now();
        for (size_t d = 0; d < host_.size() && first < end; d++) {
          const size_t take = std::min(span, end - first);
          auto &stage = host_[d];
          const size_t target = block + (quotient ? half : 0) + first;
          std::copy_n(data + target, take, stage.data());
          for (unsigned t = 0; t < terms.count; t++) {
            const size_t source = block + (quotient ? 2 * half : half) + first - size_t(terms.degree[t]);
            const size_t available = source < block + 2 * half ? std::min(take, block + 2 * half - source) : 0;
            std::fill_n(stage.data() + (t + 1) * take, take, Word(0));
            if (available) std::copy_n(data + source, available, stage.data() + (t + 1) * take);
          }
          tickets.push_back({first, take, (terms.count + 1) * take, take, block, half});
          first += take;
        }
        stats_.pack_seconds += since(packing);
        wave(tickets, [&](size_t d, const Ticket &t) {
          auto &c = *engine_.cards_[d];
          TaylorWindow<Word> view{static_cast<Word *>(c.x.data), device_[d], t.block, t.half, t.first, t.count, level, quotient};
          window_taylor<<<(t.count + 255) / 256, 256>>>(view, (t.block / (2 * t.half)) * t.half + t.first,
              static_cast<uint32_t *>(c.logs.data), static_cast<uint32_t *>(c.powers.data));
        });
        packing = Clock::now();
        for (size_t d = 0; d < tickets.size(); d++) {
          const auto &t = tickets[d];
          std::copy_n(host_[d].data(), t.count, data + t.block + (quotient ? t.half : 0) + t.first);
          completed += t.count;
        }
        stats_.pack_seconds += since(packing);
        if (Clock::now() - last >= std::chrono::seconds(30)) {
          event(label, "progress", completed, total, began); last = Clock::now();
        }
      }
    }
    event(label, "completed", completed, total, began);
  }
  void launch_vectors(size_t d, const Ticket &t, size_t length, bool columns, bool inverse) {
    auto &c = *engine_.cards_[d];
    local_vectors<<<(t.count + 127) / 128, 128>>>(static_cast<Word *>(c.x.data), device_[d], columns, inverse,
        length, t.count, t.first, 0, t.count, static_cast<uint32_t *>(c.logs.data), static_cast<uint32_t *>(c.powers.data));
  }
  void vectors_pass(Word *data, bool columns, bool inverse) {
    const size_t rows = n_ >> metadata_.split, cols = size_t(1) << metadata_.split;
    const size_t length = columns ? rows : cols, jobs = columns ? cols : rows;
    const size_t batch = capacity_ / length;
    const auto began = Clock::now(); auto last = began;
    const std::string label = columns ? "host-columns-vectors" : "host-rows-vectors";
    size_t completed = 0; event(label, "started", 0, jobs, began);
    for (size_t first = 0; first < jobs;) {
      std::vector<Ticket> tickets; auto packing = Clock::now();
      for (size_t d = 0; d < host_.size() && first < jobs; d++) {
        const size_t take = std::min(batch, jobs - first);
        for (size_t k = 0; k < length; k++)
          for (size_t j = 0; j < take; j++)
            host_[d][k * take + j] = columns ? data[k * cols + first + j] : data[(first + j) * cols + k];
        tickets.push_back({first, take, take * length, take * length}); first += take;
      }
      stats_.pack_seconds += since(packing);
      wave(tickets, [&](size_t d, const Ticket &t) { launch_vectors(d, t, length, columns, inverse); });
      packing = Clock::now();
      for (size_t d = 0; d < tickets.size(); d++) {
        const auto &t = tickets[d];
        for (size_t k = 0; k < length; k++)
          for (size_t j = 0; j < t.count; j++) {
            const Word value = host_[d][k * t.count + j];
            if (columns) data[k * cols + t.first + j] = value;
            else data[(t.first + j) * cols + k] = value;
          }
        completed += t.count;
      }
      stats_.pack_seconds += since(packing);
      if (Clock::now() - last >= std::chrono::seconds(30)) {
        event(label, "progress", completed, jobs, began); last = Clock::now();
      }
    }
    event(label, "completed", completed, jobs, began);
  }
public:
  HostTiled(Resident<Word> &engine, size_t n, size_t tile_bytes)
      : engine_(engine), n_(n), capacity_(std::min(n, tile_bytes / sizeof(Word))),
        metadata_{} {
    static_assert(sizeof(size_t) == 8, "64-bit host indexing required");
    engine_.shape(n);
    if (n < 4 || tile_bytes < 8 || tile_bytes > (64ull << 20) || engine_.encoding_k_)
      throw std::runtime_error("tiled transform parameters");
    metadata_ = matrix<Word>(engine.s_, engine.polynomial_, engine.delta_, n);
    const size_t longest = size_t(1) << (metadata_.dimension - metadata_.split);
    size_t sources = 1;
    for (const auto &t : metadata_.taylor) sources = std::max(sources, size_t(t.count) + 1);
    if (capacity_ < longest || capacity_ < sources)
      throw std::runtime_error("tile cannot hold complete vector/source set");
    for (const auto &c : engine_.cards_)
      if (c->x.bytes || c->y.bytes || c->scratch.bytes)
        throw std::runtime_error("host tiled executor requires fresh storage");
    for (auto &c : engine_.cards_) {
      check(cudaSetDevice(c->id));
      device_.push_back(engine_.plan(*c, n));
      engine_.ensure(*c, c->x, capacity_ * sizeof(Word));
      host_.emplace_back(capacity_);
      if (host_.back().capacity() != capacity_) throw std::runtime_error("host staging allocation bound");
    }
  }
  HostTiled(const HostTiled &) = delete;
  HostTiled &operator=(const HostTiled &) = delete;
  // A bounded affine row tile also provides a >32-bit coordinate witness
  // without allocating a full domain. Storage convention matches vectors_pass.
  void rows_at(Word *interleaved, size_t first, size_t count, bool inverse) {
    const size_t length = size_t(1) << metadata_.split, jobs = n_ / length;
    if (!interleaved || !count || first > jobs || count > jobs - first || count > capacity_ / length)
      throw std::runtime_error("tiled row origin/extent");
    const size_t words = count * length;
    const uint64_t limit = uint64_t(1) << (2 * engine_.s_);
    for (size_t j = 0; j < words; j++) if (uint64_t(interleaved[j]) >= limit)
      throw std::runtime_error("tiled row field value");
    std::copy_n(interleaved, words, host_[0].data());
    std::vector<Ticket> tickets{{first,count,words,words}};
    wave(tickets,[&](size_t d,const Ticket &t){launch_vectors(d,t,length,false,inverse);});
    std::copy_n(host_[0].data(), words, interleaved);
  }
  TiledStats apply(Word *data, bool inverse) {
    if (!data) throw std::runtime_error("tiled host buffer");
    stats_ = {};
    stats_.owned_device_bytes = engine_.owned();
    for (const auto &stage : host_) stats_.host_stage_bytes += stage.capacity() * sizeof(Word);
    const auto began = Clock::now();
    engine_.active_domain_ = n_;
    const uint64_t limit = uint64_t(1) << (2 * engine_.s_);
    auto last = began;
    event("host-tiled-field-validation", "started", 0, n_, began);
    for (size_t first = 0; first < n_; first += 1 << 20) {
      const size_t end = std::min(n_, first + (1 << 20));
      for (size_t j = first; j < end; j++)
        if (uint64_t(data[j]) >= limit) {
          event("host-tiled-field-validation", "failed", j, n_, began);
          throw std::runtime_error("tiled input outside field");
        }
      if (Clock::now() - last >= std::chrono::seconds(30)) {
        event("host-tiled-field-validation", "progress", end, n_, began); last = Clock::now();
      }
    }
    event("host-tiled-field-validation", "completed", n_, n_, began);
    stats_.validation_seconds = since(began);
    auto convert = [&] {
      const auto started = Clock::now();
      const unsigned levels = metadata_.dimension - metadata_.split;
      for (unsigned step = 0; step < levels; step++) {
        const unsigned k = inverse ? step : levels - 1 - step;
        taylor_pass(data, k, !inverse); taylor_pass(data, k, inverse);
      }
      stats_.taylor_seconds = since(started);
    };
    auto columns = [&] { auto t = Clock::now(); vectors_pass(data, true, inverse); stats_.columns_seconds = since(t); };
    auto rows = [&] { auto t = Clock::now(); vectors_pass(data, false, inverse); stats_.rows_seconds = since(t); };
    if (!inverse) { convert(); columns(); rows(); }
    else { rows(); columns(); convert(); }
    stats_.seconds = since(began);
    return stats_;
  }
};
} // namespace
