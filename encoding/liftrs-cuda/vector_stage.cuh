#pragma once
// Internal server component, included after preparation.cu. This adapter owns no
// host grid: the borrowed fresh executor owns one reusable device tile and plans.
namespace {
template <class Word> class VectorStage {
  Resident<Word> &engine_;
  size_t domain_, capacity_, loaded_ = 0;
  unsigned split_ = 0;
  Matrix<Word> *plan_;
  Card &card() { return *engine_.cards_[0]; }
  void select() { check(cudaSetDevice(card().id)); }
public:
  // Global domain controls metadata/affine coordinates; capacity controls only
  // local storage. Single caller, one GPU, exclusive borrowing until destruction.
  VectorStage(Resident<Word> &engine, size_t domain, size_t capacity)
      : engine_(engine), domain_(domain), capacity_(capacity), plan_(nullptr) {
    static_assert(sizeof(size_t) == 8, "64-bit global coordinates required");
    engine_.shape(domain);
    if (domain < 4 || !capacity || capacity > domain ||
        capacity > SIZE_MAX / sizeof(Word) || engine_.cards_.size() != 1 ||
        engine_.encoding_k_)
      throw std::runtime_error("vector stage parameters");
    if (card().x.bytes || card().y.bytes || card().scratch.bytes)
      throw std::runtime_error("vector stage requires fresh storage");
    unsigned dimension = 0;
    for (auto n = domain; n > 1; n >>= 1) dimension++;
    split_ = dimension / 2;
    if (capacity < (size_t(1) << (dimension - split_)))
      throw std::runtime_error("vector stage cannot hold a complete vector");
    CurrentDevice restore;
    select();
    plan_ = engine_.plan(card(), domain);
    engine_.ensure(card(), card().x, capacity * sizeof(Word));
  }
  VectorStage(const VectorStage &) = delete;
  VectorStage &operator=(const VectorStage &) = delete;
  size_t length(bool columns) const {
    const size_t columns_count = size_t(1) << split_;
    return columns ? domain_ / columns_count : columns_count;
  }
  // Upload validates before mutation. No host pointer survives this synchronous
  // call; failed validation preserves the previous loaded tile for recovery.
  void upload(const Word *data, size_t words) {
    if (!data || !words || words > capacity_)
      throw std::runtime_error("vector stage upload extent");
    const uint64_t limit = uint64_t(1) << (2 * engine_.s_);
    for (size_t j = 0; j < words; j++)
      if (uint64_t(data[j]) >= limit)
        throw std::runtime_error("vector stage field value");
    CurrentDevice restore;
    select();
    loaded_ = 0; // A failed transfer must not expose a partially replaced tile.
    check(cudaMemcpy(card().x.data, data, words * sizeof(Word), cudaMemcpyHostToDevice));
    loaded_ = words;
  }
  // Local storage is [coefficient * count + vector]. Only the affine row point
  // uses first; local addresses always start at zero. Inverse can reuse the same
  // tile without a host transfer. No implicit transpose or Taylor conversion.
  void apply(bool columns, bool inverse, size_t first, size_t count) {
    const size_t len = length(columns), jobs = domain_ / len;
    if (!count || first > jobs || count > jobs - first ||
        count > capacity_ / len || count * len != loaded_)
      throw std::runtime_error("vector stage origin/extent");
    CurrentDevice restore;
    select();
    const auto began = std::chrono::steady_clock::now();
    auto last = began;
    const char *phase = columns
        ? (inverse ? "stage-columns-inverse-vectors" : "stage-columns-forward-vectors")
        : (inverse ? "stage-rows-inverse-vectors" : "stage-rows-forward-vectors");
    engine_.active_domain_ = domain_;
    engine_.event(phase, "started", 0, count, began);
    // Bound launches so progress names an increasing vector count even for large
    // admitted tiles; each thread still owns one complete vector.
    try {
      for (size_t offset = 0; offset < count;) {
        const size_t take = std::min<size_t>(4096, count - offset);
        local_vectors<<<(take + 127) / 128, 128>>>(static_cast<Word *>(card().x.data), plan_, columns,
            inverse, len, count, first, offset, take,
            static_cast<uint32_t *>(card().logs.data), static_cast<uint32_t *>(card().powers.data));
        check(cudaGetLastError());
        check(cudaDeviceSynchronize());
        offset += take;
        if (offset == count || std::chrono::steady_clock::now() - last >= std::chrono::seconds(30)) {
          engine_.event(phase, offset == count ? "completed" : "progress", offset, count, began);
          last = std::chrono::steady_clock::now();
        }
      }
    } catch (...) {
      loaded_ = 0;
      engine_.event(phase, "failed", 0, count, began);
      throw;
    }
  }
  void download(Word *data, size_t words) {
    if (!data || !loaded_ || words != loaded_)
      throw std::runtime_error("vector stage download extent");
    CurrentDevice restore;
    select();
    check(cudaMemcpy(data, card().x.data, words * sizeof(Word), cudaMemcpyDeviceToHost));
  }
};
} // namespace
