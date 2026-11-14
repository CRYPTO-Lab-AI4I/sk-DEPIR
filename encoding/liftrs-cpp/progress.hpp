#pragma once

// Job-side progress is opt-in. Ordinary library/benchmark calls avoid log I/O
// and worker-counter updates. Logging reports work, never message/code values.
#include <atomic>
#include <chrono>
#include <cstddef>
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <ctime>

namespace liftrs::execution {
inline thread_local const char *encoding_phase = "standalone-transform";
inline bool progress_enabled() {
  const auto *value = std::getenv("LIFTRS_CPU_PROGRESS");
  return value && std::strcmp(value, "1") == 0;
}

class ProgressScope {
public:
  explicit ProgressScope(const char *phase) : previous_(encoding_phase) {
    encoding_phase = phase;
  }
  ~ProgressScope() {
    encoding_phase = previous_;
  }

private:
  const char *previous_;
};

inline void progress_record(const char *phase, const char *work, const char *event,
                            size_t completed, size_t total, uint64_t elapsed_ms,
                            const char *unit = nullptr) {
  char utc[32];
  const auto now = std::time(nullptr);
  std::tm value{};
  gmtime_r(&now, &value);
  std::strftime(utc, sizeof(utc), "%Y-%m-%dT%H:%M:%SZ", &value);
  char unit_json[80] = "";
  if (unit)
    std::snprintf(unit_json, sizeof(unit_json), ",\"unit\":\"%s\"", unit);
  std::fprintf(stderr,
               "{\"utc\":\"%s\",\"phase\":\"%s\",\"work\":\"%s\",\"event\":\"%s\","
               "\"completed\":%zu,\"total\":%zu,\"elapsed_ms\":%llu%s}\n",
               utc, phase, work, event, completed, total,
               static_cast<unsigned long long>(elapsed_ms), unit_json);
  std::fflush(stderr);
}

// advance() is called after joined/finished units, not on a timer heartbeat.
// A CAS throttles concurrent workers to one record per30s per operation. Clock
// reads and atomic updates are skipped entirely when instrumentation is disabled.
class WorkProgress {
public:
  WorkProgress(const char *work, size_t total, uint64_t interval_ms = 30000)
      : enabled_(progress_enabled()), phase_(encoding_phase), work_(work),
        total_(total), interval_(interval_ms) {
    if (enabled_) {
      start_ = now_ms();
      next_ = start_ + interval_;
    }
  }
  void advance(size_t count = 1) {
    if (!enabled_)
      return;
    auto complete = completed_.fetch_add(count, std::memory_order_relaxed) + count;
    const auto now = now_ms();
    auto next = next_.load(std::memory_order_relaxed);
    if (now >= next &&
        next_.compare_exchange_strong(next, now + interval_, std::memory_order_relaxed))
      progress_record(phase_, work_, "progress", complete, total_, now - start_);
  }

private:
  static uint64_t now_ms() {
    return std::chrono::duration_cast<std::chrono::milliseconds>(
               std::chrono::steady_clock::now().time_since_epoch())
        .count();
  }
  bool enabled_;
  const char *phase_, *work_;
  size_t total_;
  uint64_t start_ = 0, interval_;
  std::atomic<uint64_t> next_{0};
  std::atomic<size_t> completed_{0};
};

class EncodingProgress {
public:
  EncodingProgress() : enabled_(progress_enabled()) {}
  template <class Function> void run(const char *phase, Function function) {
    ProgressScope scope(phase);
    if (enabled_)
      emit(phase, "started");
    try {
      function();
    } catch (...) {
      if (enabled_)
        emit(phase, "failed");
      throw;
    }
    ++completed_;
    if (enabled_)
      emit(phase, "completed");
  }

private:
  void emit(const char *phase, const char *event) const {
    const auto elapsed = std::chrono::duration_cast<std::chrono::milliseconds>(
                             std::chrono::steady_clock::now() - start_)
                             .count();
    progress_record(phase, "encoding-operations", event, completed_, 10, elapsed);
  }
  bool enabled_;
  unsigned completed_ = 0;
  std::chrono::steady_clock::time_point start_ = std::chrono::steady_clock::now();
};
} // namespace liftrs::execution
