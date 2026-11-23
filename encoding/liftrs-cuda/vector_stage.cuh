#pragma once
// Internal server component, included after Resident inside its namespace. Owns no
// host grid: the borrowed fresh executor owns one reusable device tile and plans.
#include "taylor_window.cuh"
// Shape checks and launches are shared by full snapshots and GPU-local views;
// only the source accessor changes. Sparse coefficients/recurrence stay shared.
template <class Word> size_t Resident<Word>::taylor_shape(const Matrix<Word> &metadata,
    size_t domain,size_t capacity,size_t loaded,unsigned level,bool quotient,
    size_t block,size_t first,size_t count) {
    if (level >= metadata.dimension - metadata.split) throw std::runtime_error("Taylor level");
    const auto &terms = metadata.taylor[level];
    const size_t half = size_t(1) << (metadata.split + level);
    size_t minimum = half, maximum = 0;
    for (unsigned t = 0; t < terms.count; t++) { const auto degree=terms.degree[t]; minimum = std::min(minimum, size_t(degree)); maximum = std::max(maximum, size_t(degree)); }
    const size_t begin = quotient ? 0 : minimum, end = quotient ? maximum : half;
    if (!terms.count || !minimum || maximum > half / 2 || block >= domain ||
        block % (2 * half) || !count || first < begin || first > end || count > end-first ||
        count > capacity / (terms.count+1) || loaded != count*(terms.count+1))
      throw std::runtime_error("Taylor window extent");
  return half;
}
template <class Word> template <class View>
void Resident<Word>::taylor_launch(Card &card,View view,size_t domain) {
  CurrentDevice restore;
  check(cudaSetDevice(card.id));
  const auto began=std::chrono::steady_clock::now();
  const char *phase=view.quotient?"stage-Taylor-quotient":"stage-Taylor-remainder";
  active_domain_=domain;
  if (!bounded_taylor_progress_) event(phase,"started",0,view.count,began);
  else if (!taylor_calls_) taylor_began_=began;
  try {
    window_taylor<<<(view.count+255)/256,256>>>(view,
        (view.block/(2*view.half))*view.half+view.first,
        static_cast<uint32_t*>(card.logs.data),static_cast<uint32_t*>(card.powers.data));
    check(cudaGetLastError());check(cudaDeviceSynchronize());
    if (!bounded_taylor_progress_) event(phase,"completed",view.count,view.count,began);
    else { ++taylor_calls_; taylor_coefficients_+=view.count; }
  } catch(...) {
    taylor_failed_=true;
    if (bounded_taylor_progress_ && liftrs::execution::progress_enabled())
      liftrs::execution::progress_record(phase,"gpu-preparation","failed",0,view.count,
        std::chrono::duration_cast<std::chrono::milliseconds>(std::chrono::steady_clock::now()-began).count(),"coefficients");
    else event(phase,"failed",0,view.count,began);
    throw;
  }
}
template <class Word> void Resident<Word>::taylor_stage(Card &card,Word *values,
    Matrix<Word>*plan,const Matrix<Word>&metadata,size_t domain,size_t capacity,
    size_t loaded,unsigned level,bool quotient,size_t block,size_t first,size_t count) {
  const size_t half=taylor_shape(metadata,domain,capacity,loaded,level,quotient,block,first,count);
  TaylorWindow<Word> view{values,plan,block,half,first,count,level,quotient};
  taylor_launch(card,view,domain);
}
template <class Word> class VectorStage {
  Resident<Word> &engine_;
  size_t domain_, capacity_, loaded_ = 0;
  unsigned split_ = 0;
  Matrix<Word> *plan_;
  Matrix<Word> metadata_{};
  Card &card() { return *engine_.cards_[0]; }
  void select() { check(cudaSetDevice(card().id)); }
public:
  using word_type = Word;
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
    plan_ = engine_.plan(card(), domain, &metadata_);
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
    try {
      engine_.vector_stage(card(), static_cast<Word *>(card().x.data), plan_, domain_,
          columns, inverse, len, first, count);
    } catch (...) { loaded_ = 0; throw; }
  }

  std::vector<uint64_t> degrees(unsigned level) const {
    if (level >= metadata_.dimension - metadata_.split)
      throw std::runtime_error("Taylor level");
    const auto &terms = metadata_.taylor[level];
    std::vector<uint64_t> out;
    for (unsigned t = 0; t < terms.count; t++) out.push_back(terms.degree[t]);
    return out;
  }
  // Snapshot layout: target first, then one equally sized source window per
  // sparse term. Reuses exactly the same source accessor/update as HostTiled.
  void taylor(unsigned level, bool quotient, size_t block, size_t first, size_t count) {
    try {
      engine_.taylor_stage(card(), static_cast<Word *>(card().x.data), plan_, metadata_,
          domain_, capacity_, loaded_, level, quotient, block, first, count);
    } catch (...) { loaded_ = 0; throw; }
  }

  // Pointwise spectrum multiplication reuses the maintained kernel. Allocate
  // only the second bounded tile on demand, preserving FFT-only memory usage.
  void product(Word *left, const Word *right, size_t words) {
    if (!left || !right || !words || words > capacity_)
      throw std::runtime_error("pointwise tile extent");
    const auto a=reinterpret_cast<uintptr_t>(left), b=reinterpret_cast<uintptr_t>(right);
    const size_t bytes=words*sizeof(Word);
    if (a<b+bytes && b<a+bytes) throw std::runtime_error("pointwise overlapping operands");
    const uint64_t limit = uint64_t(1) << (2 * engine_.s_);
    for (size_t j=0;j<words;j++)
      if (uint64_t(left[j])>=limit || uint64_t(right[j])>=limit)
        throw std::runtime_error("pointwise field value");
    CurrentDevice restore; select();
    engine_.ensure(card(), card().y, words*sizeof(Word));
    upload(left, words);
    check(cudaMemcpy(card().y.data,right,words*sizeof(Word),cudaMemcpyHostToDevice));
    try {
      pointwise<<<(words+255)/256,256>>>(static_cast<Word *>(card().x.data),
          static_cast<const Word *>(card().y.data),words,engine_.s_,engine_.delta_,
          static_cast<uint32_t *>(card().logs.data),static_cast<uint32_t *>(card().powers.data));
      check(cudaGetLastError());check(cudaDeviceSynchronize());download(left,words);
    } catch (...) { loaded_=0; throw; }
  }
  void download(Word *data, size_t words) {
    if (!data || !loaded_ || words != loaded_)
      throw std::runtime_error("vector stage download extent");
    CurrentDevice restore;
    select();
    check(cudaMemcpy(data, card().x.data, words * sizeof(Word), cudaMemcpyDeviceToHost));
  }
};
