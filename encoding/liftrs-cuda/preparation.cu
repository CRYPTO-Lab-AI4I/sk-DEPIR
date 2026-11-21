// Same-node resident public-code preprocessing. No CPU transform fallback and
// no host-backed arithmetic arrays. All mathematical recurrences are shared.
#include "../liftrs-cpp/cyclic_schedule.hpp"
#include "../liftrs-cpp/progress.hpp"
#include "code_encoding_kernels.cuh"
#include "preparation_api.h"
#include "preparation_kernels.cuh"
#include "transform.cuh"
#include <array>
#include <cstdlib>
#include <cuda_runtime.h>
#include <map>
#include <mutex>
#include <unordered_map>
#include <variant>

namespace {
using namespace liftrs_cuda::preparation;
void check(cudaError_t status) {
  if (status != cudaSuccess)
    throw std::runtime_error(cudaGetErrorString(status));
}
// CUDA device selection belongs to the calling thread. Restore it so this
// server plugin can coexist with the ordinary single-device client engine.
struct CurrentDevice {
  int previous = 0;
  bool valid = false;
  CurrentDevice() { valid = cudaGetDevice(&previous) == cudaSuccess; }
  ~CurrentDevice() {
    if (valid)
      cudaSetDevice(previous);
  }
};
// Peer state belongs to CUDA primary contexts, not a field-width
// specialization. Share a lease across coexisting engines and undo only links
// this plugin owns.
struct PeerRegistry {
  struct Entry {
    size_t references;
    bool owned;
  };
  std::mutex mutex;
  std::map<std::pair<int, int>, Entry> links;
};
PeerRegistry &peer_registry() {
  static PeerRegistry state;
  return state;
}
class PeerLease {
  std::vector<std::pair<int, int>> links_;

public:
  PeerLease() = default;
  PeerLease(const PeerLease &) = delete;
  PeerLease &operator=(const PeerLease &) = delete;
  void acquire(int from, int to) {
    auto &registry = peer_registry();
    std::lock_guard<std::mutex> lock(registry.mutex);
    const auto key = std::make_pair(from, to);
    links_.push_back(key);
    auto found = registry.links.find(key);
    if (found != registry.links.end()) {
      found->second.references++;
      return;
    }
    try {
      auto entry =
          registry.links.emplace(key, PeerRegistry::Entry{1, false}).first;
      check(cudaSetDevice(from));
      const auto result = cudaDeviceEnablePeerAccess(to, 0);
      if (result == cudaErrorPeerAccessAlreadyEnabled) {
        // External mappings are borrowed, never disabled by our last lease.
        cudaGetLastError();
      } else {
        check(result);
        entry->second.owned = true;
      }
    } catch (...) {
      registry.links.erase(key);
      links_.pop_back();
      throw;
    }
  }
  ~PeerLease() {
    if (links_.empty())
      return;
    CurrentDevice restore;
    auto &registry = peer_registry();
    std::lock_guard<std::mutex> lock(registry.mutex);
    for (auto it = links_.rbegin(); it != links_.rend(); it++) {
      auto entry = registry.links.find(*it);
      if (entry == registry.links.end())
        continue;
      if (--entry->second.references == 0) {
        if (entry->second.owned) {
          cudaSetDevice(it->first);
          cudaDeviceDisablePeerAccess(it->second);
        }
        registry.links.erase(entry);
      }
    }
  }
};
struct Buffer {
  int device = 0;
  void *data = nullptr;
  size_t bytes = 0;
  Buffer() = default;
  Buffer(const Buffer &) = delete;
  Buffer &operator=(const Buffer &) = delete;
  ~Buffer() {
    if (data) {
      cudaSetDevice(device);
      cudaFree(data);
    }
  }
  void clear() {
    if (data) {
      check(cudaSetDevice(device));
      check(cudaFree(data));
    }
    data = nullptr;
    bytes = 0;
  }
  void allocate(int id, size_t size) {
    clear();
    device = id;
    check(cudaSetDevice(id));
    check(cudaMalloc(&data, size));
    bytes = size;
  }
};
struct Card {
  int id = 0;
  Buffer logs, powers, x, y, scratch, flag, base, roots, indices, jumps, weights,
      ternary, good, total, code_p, code_spectrum, parity, transport, windows;
  std::unordered_map<size_t, std::unique_ptr<Buffer>> plans;
  uint64_t owned() const {
    uint64_t bytes = 0;
    for (auto p :
         {&logs, &powers, &x, &y, &scratch, &flag, &base, &roots, &indices, &jumps,
          &weights, &ternary, &good, &total, &code_p, &code_spectrum, &parity, &transport, &windows})
      bytes += p->bytes;
    for (const auto &p : plans)
      bytes += p.second->bytes;
    return bytes;
  }
};
// Shared resident execution/storage for server preprocessing and client encoding.
// Pipeline-specific sequences live outside this executor.
template <class Word> class HostTiled;
template <class Word> class VectorStage;
template <class Word> class RankStorage;
template <class Word> class Resident {
  template <class> friend class HostTiled;
  template <class> friend class VectorStage;
  template <class> friend class RankStorage;
  unsigned s_;
  uint32_t polynomial_, delta_;
  Word omega_;
  size_t maximum_;
  uint64_t cap_;
  bool local_vectors_ = false;
  size_t encoding_k_ = 0, encoding_l_ = 0;
  size_t spectrum_ready_ = 0;
  // Member destruction frees buffers before releasing their peer-access leases.
  PeerLease peers_;
  std::vector<std::unique_ptr<Card>> cards_;
  size_t active_domain_ = 0;
  uint64_t good_count_ = 0;
  unsigned good_active_ = 0;
  using Clock = std::chrono::steady_clock;
  void event(const char *phase, const char *kind, size_t done, size_t total,
             Clock::time_point began) const {
    if (liftrs::execution::progress_enabled() &&
        (total >= 65536 || active_domain_ >= (1 << 20))) {
      const char *unit = "elements";
      if (std::strstr(phase,"-bytes")) unit="bytes";
      if (std::strstr(phase, "encoding-assembly"))
        unit = "chunks";
      if (std::strstr(phase, "vectors") || std::strstr(phase, "columns") ||
          std::strstr(phase, "rows"))
        unit = "vectors";
      if (std::strstr(phase, "Taylor"))
        unit = "coefficients";
      if (std::strstr(phase, "weight-classes"))
        unit = "classes per GPU";
      if (std::strstr(phase, "good-expansion"))
        unit = "monomials";
      if (std::strstr(phase, "good-list"))
        unit = "indices";
      liftrs::execution::progress_record(
          phase, "gpu-preparation", kind, done, total,
          std::chrono::duration_cast<std::chrono::milliseconds>(Clock::now() -
                                                                began)
              .count(),
          unit);
    }
  }
  // Shared bounded complete-vector launches for host-tile and resident rank
  // storage. Callers validate layout/extent/values and select the current GPU.
  void vector_stage(Card &c, Word *values, Matrix<Word> *metadata, size_t domain,
      bool columns, bool inverse, size_t len, size_t first, size_t count) {
    const auto began = std::chrono::steady_clock::now();
    auto last = began;
    const char *phase = columns
        ? (inverse ? "stage-columns-inverse-vectors" : "stage-columns-forward-vectors")
        : (inverse ? "stage-rows-inverse-vectors" : "stage-rows-forward-vectors");
    active_domain_ = domain;
    event(phase, "started", 0, count, began);
    // Bound launches so progress names an increasing vector count even for large
    // admitted tiles; each thread still owns one complete vector.
    try {
      for (size_t offset = 0; offset < count;) {
        const size_t take = std::min<size_t>(4096, count - offset);
        local_vectors<<<(take + 127) / 128, 128>>>(values, metadata, columns,
            inverse, len, count, first, offset, take,
            static_cast<uint32_t *>(c.logs.data), static_cast<uint32_t *>(c.powers.data));
        check(cudaGetLastError());
        check(cudaDeviceSynchronize());
        offset += take;
        if (offset == count || std::chrono::steady_clock::now() - last >= std::chrono::seconds(30)) {
          event(phase, offset == count ? "completed" : "progress", offset, count, began);
          last = std::chrono::steady_clock::now();
        }
      }
    } catch (...) {
      event(phase, "failed", 0, count, began);
      throw;
    }
  }

  // Bounded Taylor arithmetic shared by host-tile and resident-window storage.
  void taylor_stage(Card &, Word *, Matrix<Word> *, const Matrix<Word> &,
      size_t, size_t, size_t, unsigned, bool, size_t, size_t, size_t);
  size_t taylor_shape(const Matrix<Word> &, size_t, size_t, size_t, unsigned,
      bool, size_t, size_t, size_t);
  template <class View> void taylor_launch(Card &, View, size_t);
  void admit(const Card &card, uint64_t bytes) const {
    const uint64_t used = card.owned();
    if (used > cap_ || bytes > cap_ - used)
      throw std::runtime_error("GPU preprocessing per-device allocation cap");
  }
  void ensure(Card &card, Buffer &buffer, size_t bytes) {
    if (buffer.bytes >= bytes)
      return;
    buffer.clear();
    admit(card, bytes);
    buffer.allocate(card.id, bytes);
  }
  void sync(unsigned active) {
    for (unsigned d = 0; d < active; d++) {
      check(cudaSetDevice(cards_[d]->id));
      check(cudaDeviceSynchronize());
    }
  }
  unsigned active(size_t n) const {
    return unsigned(std::min<size_t>(cards_.size(), n >= 4 ? n : 1));
  }
  void shape(size_t n) const {
    if (!n || (n & (n - 1)) || n > maximum_)
      throw std::runtime_error("GPU preprocessing domain limit");
  }
  void grow(size_t n, bool product = false) {
    spectrum_ready_ = 0;
    shape(n);
    const unsigned count = active(n);
    const size_t bytes = n / count * sizeof(Word);
    // Release dead old buffers before growing; the session owns no live input
    // between operations. Do not grow retained and replacement arrays together.
    for (unsigned d = 0; d < count; d++) {
      auto &c = *cards_[d];
      if (c.x.bytes < bytes || c.scratch.bytes < bytes) {
        c.x.clear();
        c.y.clear();
        c.scratch.clear();
        admit(c, uint64_t(bytes) * (product ? 3 : 2));
        c.x.allocate(c.id, bytes);
        c.scratch.allocate(c.id, bytes);
        if (product)
          c.y.allocate(c.id, bytes);
      } else if (product && c.y.bytes < bytes) {
        // Discovery may already own full-sized data/scratch. Reuse those
        // buffers; growing only the second operand keeps the same peak cap.
        c.y.clear();
        admit(c, bytes);
        c.y.allocate(c.id, bytes);
      }
    }
  }
  Shards<Word> shards(size_t n, int which) {
    Shards<Word> view;
    const unsigned count = active(n);
    view.chunk = n / count;
    for (auto size = view.chunk; size > 1; size >>= 1)
      view.shift++;
    for (unsigned d = 0; d < count; d++) {
      auto &c = *cards_[d];
      view.data[d] = static_cast<Word *>((which == 0   ? c.x
                                          : which == 1 ? c.y
                                                       : c.scratch)
                                             .data);
    }
    return view;
  }
  Matrix<Word> *plan(Card &c, size_t n, Matrix<Word> *host = nullptr) {
    auto it = c.plans.find(n);
    if (it == c.plans.end()) {
      const auto metadata = matrix<Word>(s_, polynomial_, delta_, n);
      if (host) *host = metadata;
      admit(c, sizeof(metadata));
      auto value = std::make_unique<Buffer>();
      value->allocate(c.id, sizeof(metadata));
      check(cudaMemcpy(value->data, &metadata, sizeof(metadata),
                       cudaMemcpyHostToDevice));
      c.plans.emplace(n, std::move(value));
      it = c.plans.find(n);
    } else if (host) {
      check(cudaMemcpy(host, it->second->data, sizeof(*host), cudaMemcpyDeviceToHost));
    }
    return static_cast<Matrix<Word> *>(it->second->data);
  }
  template <class Launch>
  void stage(const char *label, size_t work, unsigned count, Launch launch) {
    const auto began = Clock::now();
    event(label, "started", 0, work, began);
    for (unsigned d = 0; d < count; d++) {
      check(cudaSetDevice(cards_[d]->id));
      launch(d);
      check(cudaGetLastError());
    }
    sync(count);
    event(label, "completed", work, work, began);
  }
  void transpose_array(size_t n, Shards<Word> from, Shards<Word> to,
                       size_t rows, size_t columns) {
    const unsigned count = active(n);
    stage("peer-transpose", n, count, [&](unsigned d) {
      if (to.chunk % rows == 0) {
        const size_t tiles = ((rows + 31) / 32) * ((to.chunk / rows + 31) / 32);
        transpose_tiles<<<tiles, dim3(32, 8)>>>(from, to.data[d], d * to.chunk,
                                                to.chunk, rows, columns);
      } else {
        transpose_scalar<<<(to.chunk + 255) / 256, 256>>>(
            from, to.data[d], d * to.chunk, to.chunk, rows, columns);
      }
    });
  }
  // Gather whole vectors, transpose each local slab for coalesced recurrence
  // accesses, then restore canonical slabs. Reuse the existing scratch only.
  void local_transpose(size_t n, Shards<Word> from, Shards<Word> to,
                       size_t rows, size_t columns) {
    stage("local-vector-transpose", n, active(n), [&](unsigned d) {
      Shards<Word> local;
      local.data[0] = from.data[d];
      local.chunk = from.chunk;
      local.shift = from.shift;
      const size_t tiles = ((rows + 31) / 32) * ((columns + 31) / 32);
      transpose_tiles<<<tiles, dim3(32, 8)>>>(local, to.data[d], 0, to.chunk,
                                              rows, columns);
    });
  }
  void vector_pass(size_t n, Shards<Word> values, Shards<Word> scratch,
                   bool columns, bool inverse, unsigned split,
                   bool base = false) {
    unsigned dim = 0;
    for (auto v = n; v > 1; v >>= 1)
      dim++;
    const size_t length = size_t(1) << (columns ? dim - split : split),
                 jobs = n / length;
    const unsigned count = std::min<unsigned>(active(n), unsigned(jobs));
    // Tiny shapes with fewer vectors than cards retain the validated peer
    // schedule; complete vectors cannot be evenly partitioned there.
    const bool local = local_vectors_ && jobs >= active(n);
    const size_t local_jobs = jobs / count;
    if (local) {
      transpose_array(n, values, scratch, length, jobs);
      local_transpose(n, scratch, values, local_jobs, length);
    }
    const char *label = base      ? "base-grid-vectors"
                        : columns ? "extension-columns"
                                  : "extension-rows";
    const auto began = Clock::now();
    auto last = began;
    event(label, "started", 0, jobs, began);
    // Bound vector waves to expose advancing work without per-element counters
    // or host-backed arithmetic. Vectors are independent within this phase.
    const size_t span = (jobs + count - 1) / count;
    const size_t wave = liftrs::execution::progress_enabled()
                            ? 8192
                            : std::max<size_t>(1, span);
    size_t completed = 0;
    for (size_t offset = 0; offset < span; offset += wave) {
      size_t advanced = 0;
      for (unsigned d = 0; d < count; d++) {
        const size_t begin = jobs * d / count, end = jobs * (d + 1) / count;
        const size_t first = std::min(end, begin + offset),
                     take = std::min(wave, end - first);
        if (!take)
          continue;
        auto &c = *cards_[d];
        check(cudaSetDevice(c.id));
        auto *metadata =
            base ? static_cast<Matrix<Word> *>(c.base.data) : plan(c, n);
        if (local)
          local_vectors<<<(take + 127) / 128, 128>>>(
              values.data[d], metadata, columns, inverse, length, local_jobs,
              begin, first - begin, take, static_cast<uint32_t *>(c.logs.data),
              static_cast<uint32_t *>(c.powers.data));
        else
          vectors<<<(take + 127) / 128, 128>>>(
              values, metadata, columns, inverse, length, jobs, first, take,
              static_cast<uint32_t *>(c.logs.data),
              static_cast<uint32_t *>(c.powers.data));
        check(cudaGetLastError());
        advanced += take;
      }
      sync(count);
      completed += advanced;
      if (Clock::now() - last >= std::chrono::seconds(30)) {
        event(label, "progress", completed, jobs, began);
        last = Clock::now();
      }
    }
    if (local) {
      local_transpose(n, values, scratch, length, local_jobs);
      transpose_array(n, scratch, values, jobs, length);
    }
    event(label, "completed", completed, jobs, began);
  }

  void transform(size_t n, Shards<Word> values, bool inverse) {
    transform(n, values, shards(n, 2), inverse, nullptr);
  }
  // Client encoding can transform the old scratch while the dead arena becomes
  // transpose scratch. Server preprocessing retains its original default views.
  void transform(size_t n, Shards<Word> values, Shards<Word> scratch, bool inverse,
                 double *stats) {
    active_domain_ = n;
    if (n == 1)
      return;
    if (n == 2)
      throw std::runtime_error("GPU preparation transform minimum4 points");
    const unsigned count = active(n);
    unsigned dim = 0;
    for (auto v = n; v > 1; v >>= 1)
      dim++;
    const unsigned split = dim / 2, levels = dim - split;
    // Build all plan metadata before dispatch, keeping allocation admission
    // outside partially submitted multi-device stages.
    for (unsigned d = 0; d < count; d++) {
      check(cudaSetDevice(cards_[d]->id));
      plan(*cards_[d], n);
    }
    auto convert = [&] {
      for (unsigned stage_index = 0; stage_index < levels; stage_index++) {
        const unsigned k = inverse ? stage_index : levels - 1 - stage_index;
        for (unsigned pass = 0; pass < 2; pass++) {
          const bool quotient = inverse ? pass != 0 : pass == 0;
          const unsigned participants =
              unsigned(std::min<size_t>(count, n / 2));
          stage(quotient ? "extension-Taylor-quotient"
                         : "extension-Taylor-remainder",
                n / 2, participants, [&](unsigned d) {
                  const size_t first = (n / 2) * d / participants,
                               take = (n / 2) * (d + 1) / participants - first;
                  auto &c = *cards_[d];
                  taylor<<<(take + 255) / 256, 256>>>(
                      values, plan(c, n), k, quotient, first, take,
                      static_cast<uint32_t *>(c.logs.data),
                      static_cast<uint32_t *>(c.powers.data));
                });
        }
      }
    };
    auto rows = [&] {
      const size_t length = size_t(1) << split, jobs = n / length;
      transpose_array(n, values, scratch, jobs, length);
      vector_pass(n, scratch, values, false, inverse, split);
      transpose_array(n, scratch, values, length, jobs);
    };
    auto timed = [&](unsigned slot, auto fn) {
      if (!stats) {
        fn();
        return;
      }
      auto began = Clock::now();
      fn();
      stats[slot] += std::chrono::duration<double>(Clock::now() - began).count();
    };
    if (!inverse) {
      timed(1, convert);
      timed(2, [&] { vector_pass(n, values, scratch, true, false, split); });
      timed(3, rows);
    } else {
      timed(3, rows);
      timed(2, [&] { vector_pass(n, values, scratch, true, true, split); });
      timed(1, convert);
    }
  }
  void upload(const Word *host, size_t n, Shards<Word> target) {
    stage("host-to-device", n, active(n), [&](unsigned d) {
      check(cudaMemcpyAsync(target.data[d], host + d * target.chunk,
                            target.chunk * sizeof(Word),
                            cudaMemcpyHostToDevice));
    });
  }
  void download(Word *host, size_t n, Shards<Word> source) {
    stage("device-to-host", n, active(n), [&](unsigned d) {
      check(cudaMemcpyAsync(host + d * source.chunk, source.data[d],
                            source.chunk * sizeof(Word),
                            cudaMemcpyDeviceToHost));
    });
  }
  void upload_base(const uint32_t *host, size_t count, size_t n,
                   Shards<Word> target) {
    const auto scratch = shards(n, 2);
    // Transpose scratch is dead before the first transform. Use it for widening
    // the whole input slab, with one upload per card instead of tiny staged
    // copies.
    stage("base-upload-padding", n, active(n), [&](unsigned d) {
      const size_t first = d * target.chunk,
                   have = first < count ? std::min(target.chunk, count - first)
                                        : 0;
      if constexpr (sizeof(Word) == 4) {
        if (have)
          check(cudaMemcpyAsync(target.data[d], host + first, have * 4,
                                cudaMemcpyHostToDevice));
        if (have < target.chunk)
          check(cudaMemsetAsync(target.data[d] + have, 0,
                                (target.chunk - have) * sizeof(Word)));
      } else {
        if (have)
          check(cudaMemcpyAsync(scratch.data[d], host + first, have * 4,
                                cudaMemcpyHostToDevice));
        widen<<<(target.chunk + 255) / 256, 256>>>(
            reinterpret_cast<uint32_t *>(scratch.data[d]), target.data[d],
            target.chunk, have);
      }
    });
  }

public:
  // Resident client constants/workspaces; implementation and ownership invariants
  // are in resident_encoding.cuh, separate from preprocessing mathematics.
  void set_code(const uint32_t *, size_t, const Word *, size_t, int);
  void encode_code(const uint32_t *, size_t, uint32_t *, size_t, double *);
  void drain() noexcept {
    for (const auto &c : cards_) {
      cudaSetDevice(c->id);
      cudaDeviceSynchronize();
    }
  }
  Resident(unsigned s, uint32_t polynomial, uint32_t delta, uint64_t omega,
           const uint32_t *logs, size_t lc, const uint32_t *powers, size_t pc,
           const int *ids, size_t count, size_t maximum, uint64_t cap)
      : s_(s), polynomial_(polynomial), delta_(delta), omega_(Word(omega)),
        maximum_(maximum), cap_(cap) {
    const char *layout = std::getenv("LIFTRS_GPU_PREPARATION_VECTOR_LAYOUT");
    if (layout && std::strcmp(layout, "peer") && std::strcmp(layout, "local"))
      throw std::runtime_error(
          "GPU preprocessing vector layout must be peer or local");
    local_vectors_ = layout && !std::strcmp(layout, "local");
    liftrs::FieldT<Word> field;
    liftrs::preparation::set_field(field, s, polynomial, delta);
    if (s > 17 || !logs || !powers || lc != (size_t(1) << s) ||
        pc != 2 * (lc - 1) || !ids || !count || count > max_devices ||
        (count & (count - 1)) || !maximum || (maximum & (maximum - 1)) ||
        maximum > field.q || maximum < count || !cap || omega >= field.q ||
        !liftrs::is_primitive(field, omega_))
      throw std::runtime_error("GPU preprocessing parameters");
    for (size_t i = 1; i < lc; i++)
      if (logs[i] >= lc - 1 || powers[logs[i]] != i)
        throw std::runtime_error("GPU preprocessing lookup representation");
    int available = 0;
    check(cudaGetDeviceCount(&available));
    for (size_t d = 0; d < count; d++) {
      if (ids[d] < 0 || ids[d] >= available)
        throw std::runtime_error("GPU preprocessing device ordinal");
      for (size_t j = 0; j < d; j++)
        if (ids[j] == ids[d])
          throw std::runtime_error("duplicate GPU preprocessing device");
      auto c = std::make_unique<Card>();
      c->id = ids[d];
      cards_.push_back(std::move(c));
    }
    // Initialize every selected primary context before enabling peer links.
    for (const auto &card : cards_) {
      check(cudaSetDevice(card->id));
      check(cudaFree(nullptr));
    }
    const auto metadata =
        matrix<Word>(s, polynomial, delta, size_t(1) << (2 * s), true);
    for (auto &owner : cards_) {
      auto &c = *owner;
      check(cudaSetDevice(c.id));
      for (const auto &other : cards_)
        if (other->id != c.id) {
          int peer = 0;
          check(cudaDeviceCanAccessPeer(&peer, c.id, other->id));
          if (!peer)
            throw std::runtime_error("GPU preprocessing requires peer access " +
                                     std::to_string(c.id) + " -> " +
                                     std::to_string(other->id) +
                                     "; no fallback");
          peers_.acquire(c.id, other->id);
        }
      size_t free = 0, total = 0;
      check(cudaMemGetInfo(&free, &total));
      // Explicit allocation cap excludes runtime/context; preserve a small
      // minimum outside it rather than claiming all device memory as usable
      // payload.
      if (cap > free || free - cap < (64ull << 20))
        throw std::runtime_error("GPU preprocessing free-memory headroom");
      ensure(c, c.logs, lc * 4);
      ensure(c, c.powers, pc * 4);
      ensure(c, c.base, sizeof(metadata));
      ensure(c, c.flag, 4);
      ensure(c, c.jumps, 64 * sizeof(Word));
      check(cudaMemcpy(c.logs.data, logs, lc * 4, cudaMemcpyHostToDevice));
      check(cudaMemcpy(c.powers.data, powers, pc * 4, cudaMemcpyHostToDevice));
      check(cudaMemcpy(c.base.data, &metadata, sizeof(metadata),
                       cudaMemcpyHostToDevice));
      Word jumps[64], step = liftrs::field_power(field, omega_, field.q - 2);
      jumps[0] = 1;
      for (unsigned j = 1; j < 64; j++)
        jumps[j] = field.mul(jumps[j - 1], step);
      check(cudaMemcpy(c.jumps.data, jumps, sizeof(jumps),
                       cudaMemcpyHostToDevice));
    }
  }
  void apply(Word *host, size_t n, bool inverse) {
    if (!host)
      throw std::runtime_error("GPU preprocessing buffer");
    shape(n);
    const uint64_t field = uint64_t(1) << (2 * s_);
    for (size_t j = 0; j < n; j++)
      if (uint64_t(host[j]) >= field)
        throw std::runtime_error("GPU preprocessing field value");
    grow(n);
    auto values = shards(n, 0);
    upload(host, n, values);
    transform(n, values, inverse);
    download(host, n, values);
  }
  void grid(Word *host, size_t n) {
    active_domain_ = n;
    if (!host || n != (size_t(1) << (2 * s_)))
      throw std::runtime_error("GPU base-grid shape");
    const size_t q = size_t(1) << s_;
    for (size_t j = 0; j < n; j++)
      if (uint64_t(host[j]) >= q)
        throw std::runtime_error("GPU base-grid coefficient");
    grow(n);
    auto values = shards(n, 0), scratch = shards(n, 2);
    upload(host, n, values);
    transpose_array(n, values, scratch, q, q);
    vector_pass(n, scratch, values, true, false, s_, true);
    transpose_array(n, scratch, values, q, q);
    vector_pass(n, values, scratch, true, false, s_, true);
    download(host, n, values);
  }
  void product(const uint32_t *a, size_t ac, const uint32_t *b, size_t bc,
               uint32_t *out, size_t limit, size_t n) {
    shape(n);
    if (!a || !b || !out || !ac || !bc || ac > n || bc > n || ac > n - bc + 1 ||
        !limit || limit > ac + bc - 1)
      throw std::runtime_error("GPU preprocessing convolution shape");
    for (size_t j = 0; j < ac; j++)
      if (a[j] >= (1u << s_))
        throw std::runtime_error("GPU product base field");
    for (size_t j = 0; j < bc; j++)
      if (b[j] >= (1u << s_))
        throw std::runtime_error("GPU product base field");
    grow(n, true);
    auto x = shards(n, 0), y = shards(n, 1);
    upload_base(a, ac, n, x);
    upload_base(b, bc, n, y);
    transform(n, x, false);
    transform(n, y, false);
    stage("GPU-pointwise", n, active(n), [&](unsigned d) {
      auto &c = *cards_[d];
      pointwise<<<(x.chunk + 255) / 256, 256>>>(
          x.data[d], y.data[d], x.chunk, s_, delta_,
          static_cast<uint32_t *>(c.logs.data),
          static_cast<uint32_t *>(c.powers.data));
    });
    transform(n, x, true);
    const auto scratch = shards(n, 2);
    stage("base-product-validation", limit, active(n), [&](unsigned d) {
      auto &c = *cards_[d];
      const size_t first = d * x.chunk,
                   take = first < limit ? std::min(x.chunk, limit - first) : 0;
      check(cudaMemsetAsync(c.flag.data, 0, 4));
      if (take)
        narrow<<<(take + 255) / 256, 256>>>(
            x.data[d], reinterpret_cast<uint32_t *>(scratch.data[d]), take, s_,
            static_cast<uint32_t *>(c.flag.data));
    });
    for (unsigned d = 0; d < active(n); d++) {
      auto &c = *cards_[d];
      check(cudaSetDevice(c.id));
      uint32_t bad = 0;
      check(cudaMemcpy(&bad, c.flag.data, 4, cudaMemcpyDeviceToHost));
      if (bad)
        throw std::runtime_error("GPU convolution escaped base field");
    }
    stage("base-product-download", limit, active(n), [&](unsigned d) {
      const size_t first = d * x.chunk,
                   take = first < limit ? std::min(x.chunk, limit - first) : 0;
      if (take)
        check(cudaMemcpyAsync(out + first, scratch.data[d], take * 4,
                              cudaMemcpyDeviceToHost));
    });
  }
  // Both output strategies use exactly the same upload and forward transform.
  // begin/read is serialized by the Go engine lock; no borrowed pointer survives.
  void spectrum_begin(const uint32_t *g, size_t count, size_t n) {
    spectrum_ready_ = 0;
    shape(n);
    if (!g || !count || count > n || n < 4)
      throw std::runtime_error("GPU spectrum shape");
    for (size_t j = 0; j < count; j++)
      if (g[j] >= (1u << s_))
        throw std::runtime_error("GPU spectrum coefficient");
    grow(n);
    auto values = shards(n, 0);
    upload_base(g, count, n, values);
    transform(n, values, false);
    spectrum_ready_ = n;
  }
  void spectrum_read(Word *out, size_t first, size_t count) {
    if (!out || !spectrum_ready_ || first > spectrum_ready_ || !count ||
        count > spectrum_ready_ - first)
      throw std::runtime_error("GPU spectrum read range/state");
    auto values = shards(spectrum_ready_, 0);
    for (size_t copied = 0; copied < count;) {
      const size_t position = first + copied;
      const unsigned d = unsigned(position / values.chunk);
      const size_t offset = position % values.chunk,
                   take = std::min(count - copied, values.chunk - offset);
      check(cudaSetDevice(cards_[d]->id));
      check(cudaMemcpy(out + copied, values.data[d] + offset,
                       take * sizeof(Word), cudaMemcpyDeviceToHost));
      copied += take;
    }
  }
  void spectrum(const uint32_t *g, size_t count, Word *out, size_t n) {
    if (!out || n != (size_t(1) << (2 * s_)))
      throw std::runtime_error("GPU spectrum shape");
    spectrum_begin(g, count, n);
    download(out, n, shards(n, 0));
  }
  void roots(const int64_t *input, size_t count, uint32_t *out) {
    const uint64_t order = (uint64_t(1) << (2 * s_)) - 1;
    if (!input || !out || !count || count > 65536)
      throw std::runtime_error("GPU root batch shape");
    for (size_t j = 0; j < count; j++)
      if (input[j] < 0 || uint64_t(input[j]) >= order ||
          (j && input[j] <= input[j - 1]))
        throw std::runtime_error("GPU root indices sorted/range");
    const unsigned participants = unsigned(std::min(cards_.size(), count));
    for (unsigned d = 0; d < participants; d++) {
      auto &c = *cards_[d];
      const size_t first = count * d / participants,
                   take = count * (d + 1) / participants - first;
      ensure(c, c.indices, take * 8);
      ensure(c, c.roots, take * 16);
      check(cudaSetDevice(c.id));
      check(cudaMemcpy(c.indices.data, input + first, take * 8,
                       cudaMemcpyHostToDevice));
      check(cudaMemset(c.flag.data, 0, 4));
      root_factors<<<(take + 127) / 128, 128>>>(
          static_cast<int64_t *>(c.indices.data), take,
          static_cast<uint32_t *>(c.roots.data), order, s_, delta_, omega_,
          static_cast<Word *>(c.jumps.data),
          static_cast<uint32_t *>(c.logs.data),
          static_cast<uint32_t *>(c.powers.data),
          static_cast<uint32_t *>(c.flag.data));
      check(cudaGetLastError());
    }
    sync(participants);
    for (unsigned d = 0; d < participants; d++) {
      auto &c = *cards_[d];
      const size_t first = count * d / participants,
                   take = count * (d + 1) / participants - first;
      check(cudaSetDevice(c.id));
      check(cudaMemcpy(out + 4 * first, c.roots.data, take * 16,
                       cudaMemcpyDeviceToHost));
      uint32_t bad = 0;
      check(cudaMemcpy(&bad, c.flag.data, 4, cudaMemcpyDeviceToHost));
      if (bad)
        throw std::runtime_error("GPU fixed root escaped base field");
    }
  }
  uint64_t good_count(unsigned t) {
    const uint64_t n = uint64_t(1) << (2 * s_);
    const size_t q = size_t(1) << s_, words = (n + 63) / 64;
    if (maximum_ != n || !t || t >= q - 1)
      throw std::runtime_error(
          "GPU exact enumeration requires complete domain/t");
    for (auto &c : cards_) {
      c->weights.clear();
      c->ternary.clear();
      c->good.clear();
      c->total.clear();
    }
    size_t classes = 1;
    for (unsigned j = 0; j < s_; j++)
      classes *= 3;
    good_active_ = unsigned(std::min(cards_.size(), words));
    good_count_ = 0;
    for (unsigned d = 0; d < good_active_; d++) {
      auto &c = *cards_[d];
      const size_t take = words / good_active_;
      // Admit every principal temporary together before allocating the large
      // table.
      admit(c, classes + q * 4 + take * 8 + 8);
      ensure(c, c.weights, classes);
      ensure(c, c.ternary, q * 4);
      ensure(c, c.good, take * 8);
      ensure(c, c.total, 8);
    }
    stage("GPU-good-weight-classes", classes, good_active_, [&](unsigned d) {
      auto &c = *cards_[d];
      weight_classes<<<(classes + 255) / 256, 256>>>(
          static_cast<unsigned char *>(c.weights.data), classes, s_, t);
      ternary_indices<<<(q + 255) / 256, 256>>>(
          static_cast<uint32_t *>(c.ternary.data), q, s_);
    });
    stage("GPU-good-expansion", n, good_active_, [&](unsigned d) {
      auto &c = *cards_[d];
      const size_t take = words / good_active_;
      check(cudaMemset(c.total.data, 0, 8));
      expand_good<<<(take + 255) / 256, 256>>>(
          static_cast<unsigned char *>(c.weights.data),
          static_cast<uint32_t *>(c.ternary.data),
          static_cast<uint64_t *>(c.good.data),
          static_cast<unsigned long long *>(c.total.data), d * take, take, n,
          s_);
    });
    for (unsigned d = 0; d < good_active_; d++) {
      auto &c = *cards_[d];
      uint64_t count = 0;
      check(cudaSetDevice(c.id));
      check(cudaMemcpy(&count, c.total.data, 8, cudaMemcpyDeviceToHost));
      good_count_ += count;
      c.weights.clear();
      c.ternary.clear();
      c.total.clear();
    }
    return good_count_;
  }
  void good_fill(int64_t *output, size_t count) {
    if (!output || !good_active_ || count != good_count_)
      throw std::runtime_error("GPU good-list state/shape");
    const size_t words = ((size_t(1) << (2 * s_)) + 63) / 64,
                 per_card = words / good_active_;
    size_t written = 0;
    std::vector<uint64_t> block(std::min<size_t>(per_card, 1 << 20));
    const auto began = Clock::now();
    auto last = began;
    for (unsigned d = 0; d < good_active_; d++) {
      auto &c = *cards_[d];
      check(cudaSetDevice(c.id));
      for (size_t first = 0; first < per_card; first += block.size()) {
        const size_t take = std::min(block.size(), per_card - first);
        check(cudaMemcpy(block.data(),
                         static_cast<uint64_t *>(c.good.data) + first, take * 8,
                         cudaMemcpyDeviceToHost));
        for (size_t j = 0; j < take; j++) {
          auto bits = block[j];
          while (bits) {
            if (written >= count)
              throw std::runtime_error("GPU good-list count mismatch");
            output[written++] = int64_t(64 * (d * per_card + first + j) +
                                        __builtin_ctzll(bits));
            bits &= bits - 1;
          }
        }
        if (Clock::now() - last >= std::chrono::seconds(30)) {
          event("GPU-good-list-download", "progress", written, count, began);
          last = Clock::now();
        }
      }
      c.good.clear();
    }
    good_active_ = 0;
    if (written != count)
      throw std::runtime_error("GPU good-list count mismatch");
    event("GPU-good-list-download", "completed", written, count, began);
  }
  void release_products() {
    for (auto &c : cards_) {
      c->y.clear();
      c->roots.clear();
      c->indices.clear();
    }
  }
  uint64_t owned() const {
    uint64_t bytes = 0;
    for (const auto &c : cards_)
      bytes += c->owned();
    return bytes;
  }
};
#include "resident_encoding.cuh"
#include "vector_stage.cuh"
#include "rank_storage.cuh"

struct Engine {
  using Narrow = Resident<uint32_t>;
  using Wide = Resident<uint64_t>;
  std::variant<std::unique_ptr<Narrow>, std::unique_ptr<Wide>> state;
  // Destroy the borrowing stage before its resident executor.
  std::variant<std::monostate, std::unique_ptr<VectorStage<uint32_t>>,
               std::unique_ptr<VectorStage<uint64_t>>> vectors;
  size_t vector_domain = 0, vector_capacity = 0;
};
template <class Function> int api(char *error, size_t capacity, Function fn) {
  try {
    CurrentDevice restore;
    fn();
    return 0;
  } catch (const std::exception &e) {
    if (error && capacity)
      std::snprintf(error, capacity, "%s", e.what());
    return 1;
  }
}
template <class Function> void visit(void *handle, Function fn) {
  if (!handle)
    throw std::runtime_error("GPU preparation handle");
  std::visit(
      [&](auto &state) {
        if (!state)
          throw std::runtime_error("GPU preparation state");
        try {
          fn(*state);
        } catch (...) {
          state->drain();
          throw;
        }
      },
      static_cast<Engine *>(handle)->state);
}
} // namespace
extern "C" int pg_create(unsigned s, uint32_t polynomial, uint32_t delta,
                         uint64_t omega, const uint32_t *logs, size_t lc,
                         const uint32_t *powers, size_t pc, const int *ids,
                         size_t count, size_t maximum, uint64_t cap,
                         void **output, char *error, size_t capacity) {
  if (output)
    *output = nullptr;
  return api(error, capacity, [&] {
    if (!output)
      throw std::runtime_error("GPU preparation output handle");
    auto e = std::make_unique<Engine>();
    if (s <= 16)
      e->state = std::make_unique<Engine::Narrow>(s, polynomial, delta, omega,
                                                  logs, lc, powers, pc, ids,
                                                  count, maximum, cap);
    else
      e->state =
          std::make_unique<Engine::Wide>(s, polynomial, delta, omega, logs, lc,
                                         powers, pc, ids, count, maximum, cap);
    *output = e.release();
  });
}
extern "C" int pg_apply(void *handle, void *values, size_t n, unsigned width,
                        int inverse, char *error, size_t capacity) {
  return api(error, capacity, [&] {
    if (inverse != 0 && inverse != 1)
      throw std::runtime_error("GPU transform direction");
    visit(handle, [&](auto &state) {
      using Word = std::conditional_t<
          std::is_same_v<std::decay_t<decltype(state)>, Engine::Narrow>,
          uint32_t, uint64_t>;
      if (width != sizeof(Word))
        throw std::runtime_error("GPU preparation value width");
      state.apply(static_cast<Word *>(values), n, inverse != 0);
    });
  });
}
extern "C" int pg_grid(void *handle, void *values, size_t n, unsigned width,
                       char *error, size_t capacity) {
  return api(error, capacity, [&] {
    visit(handle, [&](auto &state) {
      using Word = std::conditional_t<
          std::is_same_v<std::decay_t<decltype(state)>, Engine::Narrow>,
          uint32_t, uint64_t>;
      if (width != sizeof(Word))
        throw std::runtime_error("GPU preparation value width");
      state.grid(static_cast<Word *>(values), n);
    });
  });
}
extern "C" int pg_product(void *h, const uint32_t *a, size_t ac,
                          const uint32_t *b, size_t bc, uint32_t *out,
                          size_t limit, size_t n, char *error,
                          size_t capacity) {
  return api(error, capacity, [&] {
    visit(h, [&](auto &state) { state.product(a, ac, b, bc, out, limit, n); });
  });
}
extern "C" int pg_spectrum(void *h, const uint32_t *g, size_t count, void *out,
                           size_t n, unsigned width, char *error,
                           size_t capacity) {
  return api(error, capacity, [&] {
    visit(h, [&](auto &state) {
      using Word = std::conditional_t<
          std::is_same_v<std::decay_t<decltype(state)>, Engine::Narrow>,
          uint32_t, uint64_t>;
      if (width != sizeof(Word))
        throw std::runtime_error("GPU spectrum width");
      state.spectrum(g, count, static_cast<Word *>(out), n);
    });
  });
}
extern "C" int pg_spectrum_begin(void *h, const uint32_t *g, size_t count,
                                  size_t n, char *error, size_t capacity) {
  return api(error, capacity, [&] {
    visit(h, [&](auto &state) { state.spectrum_begin(g, count, n); });
  });
}
extern "C" int pg_spectrum_read(void *h, void *out, size_t first, size_t count,
                                 unsigned width, char *error, size_t capacity) {
  return api(error, capacity, [&] {
    visit(h, [&](auto &state) {
      using Word = std::conditional_t<
          std::is_same_v<std::decay_t<decltype(state)>, Engine::Narrow>,
          uint32_t, uint64_t>;
      if (width != sizeof(Word))
        throw std::runtime_error("GPU spectrum read width");
      state.spectrum_read(static_cast<Word *>(out), first, count);
    });
  });
}
extern "C" int pg_roots(void *h, const int64_t *input, size_t count,
                        uint32_t *out, char *error, size_t capacity) {
  return api(error, capacity, [&] {
    visit(h, [&](auto &state) { state.roots(input, count, out); });
  });
}
extern "C" int pg_good_count(void *h, unsigned t, uint64_t *count, char *error,
                             size_t capacity) {
  return api(error, capacity, [&] {
    if (!count)
      throw std::runtime_error("GPU good count pointer");
    visit(h, [&](auto &state) { *count = state.good_count(t); });
  });
}
extern "C" int pg_good_fill(void *h, int64_t *out, size_t count, char *error,
                            size_t capacity) {
  return api(error, capacity, [&] {
    visit(h, [&](auto &state) { state.good_fill(out, count); });
  });
}
extern "C" int pg_release_products(void *h, char *error, size_t capacity) {
  return api(error, capacity,
             [&] { visit(h, [&](auto &state) { state.release_products(); }); });
}
extern "C" uint64_t pg_owned_bytes(void *h) {
  uint64_t bytes = 0;
  if (h)
    visit(h, [&](auto &state) { bytes = state.owned(); });
  return bytes;
}
extern "C" void pg_destroy(void *h) {
  CurrentDevice restore;
  delete static_cast<Engine *>(h);
}

extern "C" int pg_set_code(void *handle, const uint32_t *p, size_t pc,
                           const void *spectrum, size_t n, unsigned width, int layout,
                           char *error, size_t capacity) {
  return api(error, capacity, [&] {
    visit(handle, [&](auto &state) {
      using Word = std::conditional_t<
          std::is_same_v<std::decay_t<decltype(state)>, Engine::Narrow>, uint32_t,
          uint64_t>;
      if (width != sizeof(Word))
        throw std::runtime_error("GPU encoding spectrum width");
      state.set_code(p, pc, static_cast<const Word *>(spectrum), n, layout);
    });
  });
}
extern "C" int pg_encode_code(void *handle, const uint32_t *message, size_t k,
                              uint32_t *word, size_t n, double *stats, char *error,
                              size_t capacity) {
  return api(error, capacity, [&] {
    visit(handle, [&](auto &state) { state.encode_code(message, k, word, n, stats); });
  });
}

namespace {
template <class Function> void visit_stage(void *handle, unsigned width,
    size_t domain, size_t local_capacity, Function fn) {
  visit(handle, [&](auto &state) {
    using Word = std::conditional_t<
        std::is_same_v<std::decay_t<decltype(state)>, Engine::Narrow>, uint32_t, uint64_t>;
    if (width != sizeof(Word)) throw std::runtime_error("GPU stage width");
    auto &engine = *static_cast<Engine *>(handle);
    if (!engine.vector_domain) {
      auto stage = std::make_unique<VectorStage<Word>>(state, domain, local_capacity);
      engine.vectors = std::move(stage);
      engine.vector_domain = domain;
      engine.vector_capacity = local_capacity;
    }
    if (engine.vector_domain != domain || engine.vector_capacity != local_capacity)
      throw std::runtime_error("GPU stage immutable shape");
    fn(*std::get<std::unique_ptr<VectorStage<Word>>>(engine.vectors));
  });
}
} // namespace
// Optional server-stage ABI, used only by a dedicated VectorPreparation owner.
extern "C" int pg_vector_stage(void *handle, void *values, size_t words,
    unsigned width, size_t domain, size_t local_capacity, int columns,
    int inverse, size_t first, size_t count, char *error, size_t capacity) {
  return api(error, capacity, [&] {
    if (!values || !words || (columns != 0 && columns != 1) || (inverse != 0 && inverse != 1))
      throw std::runtime_error("GPU vector stage arguments");
    visit_stage(handle, width, domain, local_capacity, [&](auto &stage) {
      using Word = typename std::decay_t<decltype(stage)>::word_type;
      stage.upload(static_cast<Word *>(values), words);
      stage.apply(columns != 0, inverse != 0, first, count);
      stage.download(static_cast<Word *>(values), words);
    });
  });
}
extern "C" int pg_taylor_degrees(void *handle, unsigned width, size_t domain,
    size_t local_capacity, unsigned level, uint64_t *out, size_t limit,
    size_t *count, char *error, size_t capacity) {
  return api(error, capacity, [&] {
    if (!out || !count) throw std::runtime_error("Taylor degree output");
    visit_stage(handle, width, domain, local_capacity, [&](auto &stage) {
      const auto terms = stage.degrees(level);
      if (terms.size() > limit) throw std::runtime_error("Taylor degree capacity");
      std::copy(terms.begin(), terms.end(), out); *count = terms.size();
    });
  });
}
extern "C" int pg_taylor_stage(void *handle, void *values, size_t words,
    unsigned width, size_t domain, size_t local_capacity, unsigned level,
    int quotient, size_t block, size_t first, size_t count, char *error, size_t capacity) {
  return api(error, capacity, [&] {
    if (!values || !words || (quotient != 0 && quotient != 1))
      throw std::runtime_error("GPU Taylor stage arguments");
    visit_stage(handle, width, domain, local_capacity, [&](auto &stage) {
      using Word = typename std::decay_t<decltype(stage)>::word_type;
      stage.upload(static_cast<Word *>(values), words);
      stage.taylor(level, quotient != 0, block, first, count);
      stage.download(static_cast<Word *>(values), words);
    });
  });
}

extern "C" int pg_stage_pointwise(void *handle, void *left, const void *right,
    size_t words, unsigned width, size_t domain, size_t local_capacity,
    char *error, size_t capacity) {
  return api(error,capacity,[&]{
    visit_stage(handle,width,domain,local_capacity,[&](auto &stage){
      using Word=typename std::decay_t<decltype(stage)>::word_type;
      stage.product(static_cast<Word *>(left),static_cast<const Word *>(right),words);
    });
  });
}

// Optional dedicated rank-storage owner. Destroy before its borrowed engine.
extern "C" int pg_rank_create(void *engine, size_t rows, size_t columns,
    unsigned rank, unsigned ranks, size_t tile, void **out, char *error, size_t cap) {
  if (out) *out = nullptr;
  return api(error, cap, [&] {
    if (!out) throw std::runtime_error("rank storage output");
    visit(engine, [&](auto &state) {
      using Word = std::conditional_t<
          std::is_same_v<std::decay_t<decltype(state)>, Engine::Narrow>, uint32_t, uint64_t>;
      auto arena = std::make_unique<RankStorage<Word>>(state, rows, columns, rank, ranks, tile);
      auto handle = std::make_unique<RankStorageHandle>(std::move(arena));
      *out = handle.release();
    });
  });
}
extern "C" int pg_rank_view(void *h, unsigned slot, size_t rows, size_t columns,
    int interleaved, char *error, size_t cap) {
  return api(error, cap, [&] {
    if (interleaved != 0 && interleaved != 1) throw std::runtime_error("rank storage layout");
    visit_rank(h, [&](auto &s) { s.view(slot, rows, columns, interleaved != 0); });
  });
}
extern "C" int pg_rank_transfer(void *h, unsigned slot, size_t offset, void *values,
    size_t words, int upload, char *error, size_t cap) {
  return api(error, cap, [&] {
    if (upload != 0 && upload != 1) throw std::runtime_error("rank storage direction");
    visit_rank(h, [&](auto &s) {
      using Word = typename std::decay_t<decltype(s)>::word_type;
      s.transfer(slot, offset, static_cast<Word *>(values), words, upload != 0);
    });
  });
}
extern "C" int pg_rank_local(void *h, unsigned source, unsigned target, char *error, size_t cap) {
  return api(error, cap, [&] { visit_rank(h, [&](auto &s) { s.local(source, target); }); });
}
extern "C" int pg_rank_tile(void *h, unsigned slot, unsigned peer, size_t offset,
    void *values, size_t words, int pack, char *error, size_t cap) {
  return api(error, cap, [&] {
    if (pack != 0 && pack != 1) throw std::runtime_error("rank storage direction");
    visit_rank(h, [&](auto &s) {
      using Word = typename std::decay_t<decltype(s)>::word_type;
      s.tile(slot, peer, offset, static_cast<Word *>(values), words, pack != 0);
    });
  });
}
extern "C" void pg_rank_destroy(void *h) {
  CurrentDevice restore;
  delete static_cast<RankStorageHandle *>(h);
}

extern "C" int pg_rank_vectors(void *h, unsigned slot, int columns, int inverse, char *error, size_t cap) {
  return api(error, cap, [&] {
    if ((columns != 0 && columns != 1) || (inverse != 0 && inverse != 1))
      throw std::runtime_error("rank vector axis/direction");
    visit_rank(h, [&](auto &state) { state.vectors(slot, columns != 0, inverse != 0); });
  });

}

extern "C" int pg_rank_degrees(void *h, unsigned level, uint64_t *out,
    size_t limit, size_t *count, char *error, size_t cap) {
  return api(error, cap, [&] {
    if (!out || !count) throw std::runtime_error("rank Taylor degree output");
    visit_rank(h, [&](auto &s) {
      const auto terms = s.degrees(level);
      if (terms.size() > limit) throw std::runtime_error("rank Taylor degree capacity");
      std::copy(terms.begin(), terms.end(), out); *count = terms.size();
    });
  });
}
extern "C" int pg_rank_taylor(void *h, unsigned slot, const void *snapshot,
    size_t words, size_t capacity, unsigned level, int quotient,
    size_t block, size_t first, size_t count, char *error, size_t cap) {
  return api(error, cap, [&] {
    if (quotient != 0 && quotient != 1) throw std::runtime_error("rank Taylor direction");
    visit_rank(h, [&](auto &s) {
      using Word = typename std::decay_t<decltype(s)>::word_type;
      s.taylor(slot, static_cast<const Word *>(snapshot), words, capacity, level,
          quotient != 0, block, first, count);
    });
  });
}

extern "C" int pg_rank_pointwise(void *h, unsigned left, unsigned right, char *error, size_t cap) {
  return api(error, cap, [&] { visit_rank(h, [&](auto &s) { s.pointwise(left, right); }); });
}

extern "C" int pg_rank_taylor_local(void *h,unsigned slot,const void *snapshot,
    size_t words,size_t capacity,unsigned level,int quotient,size_t block,
    size_t first,size_t count,char *error,size_t cap) {
  return api(error,cap,[&]{
    if(quotient!=0 && quotient!=1)throw std::runtime_error("rank Taylor direction");
    visit_rank(h,[&](auto&s){
      using Word=typename std::decay_t<decltype(s)>::word_type;
      s.taylor(slot,static_cast<const Word*>(snapshot),words,capacity,level,
          quotient!=0,block,first,count,true);
    });
  });
}
