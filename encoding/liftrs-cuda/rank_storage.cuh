#pragma once
// Included inside Resident's namespace. Storage moves canonical slot bits without
// arithmetic. Three owned arrays and one bounded transfer tile are cap-accounted.
struct RankView {
  size_t rows = 0, columns = 0, first = 0, count = 0;
  bool interleaved = false;
  __host__ __device__ size_t address(size_t row, size_t column) const {
    return interleaved ? column * count + row : row * columns + column;
  }
  size_t words() const { return count * columns; }
};
// Directly address owned coefficients; only remote sources fall back to packed
// windows. Pass barriers and operative-source trimming prevent target races.
template<class Word> struct LocalTaylorWindow : TaylorWindow<Word> {
  Word *owned;
  RankView layout;
  __device__ Word &operator[](size_t index) const {
    const size_t origin=layout.first*layout.columns,words=layout.count*layout.columns;
    if(index>=origin && index-origin<words) {
      const size_t local=index-origin;
      return owned[layout.address(local/layout.columns,local%layout.columns)];
    }
    return TaylorWindow<Word>::operator[](index);
  }
};
static std::pair<size_t, size_t> storage_band(size_t n, unsigned rank, unsigned ranks) {
  return {n / ranks * rank + std::min<size_t>(rank, n % ranks),
          n / ranks + (rank < n % ranks)};
}
template <class Word> __global__ void rank_transfer(Word *data, Word *tile,
    RankView view, size_t offset, size_t words, bool upload) {
  const size_t j = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  if (j >= words) return;
  const size_t index = offset + j;
  const size_t address = view.address(index / view.columns, index % view.columns);
  if (upload) data[address] = tile[j];
  else tile[j] = data[address];
}
template <class Word> __global__ void rank_local(const Word *input, Word *output,
    RankView src, RankView dst, size_t offset, size_t words) {
  const size_t j = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  if (j >= words) return;
  const size_t index = offset + j, row = index % src.count, column = index / src.count;
  output[dst.address(column, src.first + row)] = input[src.address(row, dst.first + column)];
}
template <class Word> __global__ void rank_pack(Word *data, Word *tile,
    RankView view, size_t peer_first, size_t peer_count, size_t offset,
    size_t words, bool pack) {
  const size_t j = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  if (j >= words) return;
  const size_t index = offset + j;
  // Send order is original-column first; receive order is the same wire order,
  // with the sender's row band becoming columns in the destination view.
  const size_t address = pack
      ? view.address(index % view.count, peer_first + index / view.count)
      : view.address(index / peer_count, peer_first + index % peer_count);
  if (pack) tile[j] = data[address];
  else data[address] = tile[j];
}
// Boundary check before field-table indexing; raw storage intentionally accepts
// all slot bits, whereas arithmetic must reject noncanonical values before writes.
template <class Word> __global__ void rank_validate(const Word *values, size_t words,
    uint64_t limit, uint32_t *bad) {
  const size_t j = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  if (j < words && uint64_t(values[j]) >= limit) atomicExch(bad, 1u);
}
template <class Word> class RankStorage {
  Resident<Word> &engine_;
  size_t rows_, columns_, capacity_, tile_;
  unsigned rank_, ranks_;
  RankView views_[3];
  // A true flag is a proof that every live slot word is canonical. Raw writes
  // invalidate it before launch; zero reset and successful field scans restore
  // it. Closed field arithmetic preserves it only after a successful launch.
  bool canonical_[3]{};
  Matrix<Word> metadata_{};
  Matrix<Word> *plan_ = nullptr;
  Card &card() { return *engine_.cards_[0]; }
  Word *data(unsigned slot) {
    if (slot >= 3) throw std::runtime_error("rank storage slot");
    return static_cast<Word *>((slot == 0 ? card().x : slot == 1 ? card().y : card().scratch).data);
  }
  void sync() { check(cudaGetLastError()); check(cudaDeviceSynchronize()); }
  void extent(unsigned slot, size_t offset, size_t words) {
    data(slot);
    if (!words || words > tile_ || offset > views_[slot].words() ||
        words > views_[slot].words() - offset)
      throw std::runtime_error("rank storage transfer extent");
  }
  // Raw slots accept all bits, so validate complete operands before field-table
  // indexing or any arithmetic write. Flag and launches reuse bounded storage.
  void validate(unsigned slot,const char *phase,bool base=false) {
    const size_t words=views_[slot].words();
    canonical_[slot]=false;
    check(cudaSetDevice(card().id));
    auto *bad = static_cast<uint32_t *>(card().flag.data);
    check(cudaMemset(bad, 0, sizeof(uint32_t)));
    const auto began = std::chrono::steady_clock::now();
    auto last = began;

    engine_.active_domain_ = rows_ * columns_;
    engine_.event(phase, "started", 0, words, began);
    for (size_t offset = 0; offset < words;) {
      const size_t take = std::min<size_t>(1u << 20, words - offset);
      rank_validate<<<(take + 255) / 256, 256>>>(data(slot) + offset, take,
          uint64_t(1) << (base ? engine_.s_ : 2 * engine_.s_), bad);
      sync(); offset += take;
      if (offset == words || std::chrono::steady_clock::now() - last >= std::chrono::seconds(30)) {
        engine_.event(phase, offset == words ? "completed" : "progress", offset, words, began);
        last = std::chrono::steady_clock::now();
      }
    }
    uint32_t invalid = 0;
    check(cudaMemcpy(&invalid, bad, sizeof(invalid), cudaMemcpyDeviceToHost));
    if (invalid) throw std::runtime_error("rank arithmetic field value");
    canonical_[slot]=true;
  }
  void prepare() {
    const size_t domain = rows_ * columns_;
    engine_.shape(domain);
    if (domain < 4) throw std::runtime_error("rank Taylor domain");
    unsigned dimension = 0;
    for (size_t n = domain; n > 1; n >>= 1) dimension++;
    const size_t cols = size_t(1) << (dimension / 2);
    if (columns_ != cols || rows_ != domain / cols)
      throw std::runtime_error("rank Taylor canonical split");
    check(cudaSetDevice(card().id));
    if (!plan_) plan_ = engine_.plan(card(), domain, &metadata_);
  }
public:
  using word_type = Word;
  RankStorage(Resident<Word> &engine, size_t rows, size_t columns,
      unsigned rank, unsigned ranks, size_t tile)
      : engine_(engine), rows_(rows), columns_(columns), tile_(tile), rank_(rank), ranks_(ranks) {
    if (!rows || !columns || rows > SIZE_MAX / columns || rows * columns > engine_.maximum_ ||
        !ranks || ranks > 64 || rank >= ranks || !tile || tile > (1u << 20) ||
        engine_.cards_.size() != 1 || engine_.encoding_k_)
      throw std::runtime_error("rank storage shape/rank/tile");
    auto &c = card();
    if (c.x.bytes || c.y.bytes || c.scratch.bytes || c.transport.bytes || c.windows.bytes)
      throw std::runtime_error("rank storage requires dedicated fresh engine");
    capacity_ = std::max(storage_band(rows, rank, ranks).second * columns,
                        storage_band(columns, rank, ranks).second * rows);
    if (capacity_ > SIZE_MAX / sizeof(Word)) throw std::runtime_error("rank storage byte extent");
    // Admit all arrays before allocating any of them; include plans/tables already
    // retained by the executor. No private allocation escapes OwnedBytes.
    const uint64_t bytes = uint64_t(capacity_) * sizeof(Word);
    if (bytes > UINT64_MAX / 3 || tile > SIZE_MAX / sizeof(Word) ||
        3 * bytes > UINT64_MAX - tile * sizeof(Word))
      throw std::runtime_error("rank storage total byte extent");
    engine_.admit(c, 3 * bytes + tile * sizeof(Word));
    check(cudaSetDevice(c.id));
    if (capacity_)
      for (auto *b : {&c.x, &c.y, &c.scratch}) engine_.ensure(c, *b, capacity_ * sizeof(Word));
    engine_.ensure(c, c.transport, tile * sizeof(Word));
    for (unsigned slot = 0; slot < 3; slot++) view(slot, rows, columns, false);
  }
  // Reset destroys only this slot's previous contents. Zero initialization makes
  // empty/tail destinations deterministic; partial schedules are not FFT results.
  void view(unsigned slot, size_t rows, size_t columns, bool interleaved) {
    data(slot);
    if (!((rows == rows_ && columns == columns_) || (rows == columns_ && columns == rows_)))
      throw std::runtime_error("rank storage immutable matrix shape");
    const auto band = storage_band(rows, rank_, ranks_);
    RankView next{rows, columns, band.first, band.second, interleaved};
    check(cudaSetDevice(card().id));
    canonical_[slot]=false;
    if (next.words()) check(cudaMemset(data(slot), 0, next.words() * sizeof(Word)));
    check(cudaDeviceSynchronize());
    views_[slot] = next;
    canonical_[slot]=true;
  }
  void transfer(unsigned slot, size_t offset, Word *values, size_t words, bool upload) {
    extent(slot, offset, words);
    if (!values) throw std::runtime_error("rank storage host pointer");
    check(cudaSetDevice(card().id));
    auto *tile = static_cast<Word *>(card().transport.data);
    if (upload) {canonical_[slot]=false;check(cudaMemcpy(tile, values, words * sizeof(Word), cudaMemcpyHostToDevice));}
    rank_transfer<<<(words + 255) / 256, 256>>>(data(slot), tile, views_[slot], offset, words, upload);
    sync();
    if (!upload) check(cudaMemcpy(values, tile, words * sizeof(Word), cudaMemcpyDeviceToHost));
  }
  void local(unsigned source, unsigned target) {
    data(source); data(target);
    if (source == target || views_[source].rows != views_[target].columns ||
        views_[source].columns != views_[target].rows)
      throw std::runtime_error("rank storage transpose shape/alias");
    check(cudaSetDevice(card().id));
    const bool valid=canonical_[source] && canonical_[target];
    canonical_[target]=false;
    const size_t total = views_[source].count * views_[target].count;
    const auto began = std::chrono::steady_clock::now();
    auto last = began;
    const char *phase = "rank-storage-local-transpose";
    engine_.event(phase, "started", 0, total, began);
    // Bound launches/progress even if a future admitted shard has billions of
    // local elements; the transfer paths are already bounded by tile_.
    for (size_t offset = 0; offset < total;) {
      const size_t take = std::min<size_t>(1u << 20, total - offset);
      rank_local<<<(take + 255) / 256, 256>>>(data(source), data(target), views_[source], views_[target], offset, take);
      sync();
      offset += take;
      if (offset == total || std::chrono::steady_clock::now() - last >= std::chrono::seconds(30)) {
        engine_.event(phase, offset == total ? "completed" : "progress", offset, total, began);
        last = std::chrono::steady_clock::now();
      }
    }
    if (!total) engine_.event(phase, "completed", 0, 0, began);
    canonical_[target]=valid;
  }
  // Reuse the same local-vector launcher as VectorStage. Shape/layout admit only
  // the canonical balanced matrix, with complete vectors and global affine rows.
  // Base mode selects the existing Fq tensor metadata and zero affine origin.
  void vectors(unsigned slot, bool columns, bool inverse, bool base=false) {
    data(slot);
    const size_t domain = rows_ * columns_;
    engine_.shape(domain);
    if (domain < 4) throw std::runtime_error("rank vector domain");
    unsigned dimension = 0;
    for (size_t n = domain; n > 1; n >>= 1) dimension++;
    const size_t cols = size_t(1) << (dimension / 2), rows = domain / cols;
    const auto &v = views_[slot];
    if (rows_ != rows || columns_ != cols || !v.interleaved ||
        v.rows != (columns ? cols : rows) || v.columns != (columns ? rows : cols))
      throw std::runtime_error("rank vector canonical shape/interleaved layout");
    if (base && (domain != (size_t(1) << (2*engine_.s_)) || !columns || inverse))
      throw std::runtime_error("rank base-grid vectors require full q by q forward shape");
    validate(slot,base ? "rank-base-field-validation" : "rank-vector-field-validation",base);
    // The existing tensor metadata uses Fq for BOTH axes and no affine row
    // origin. Extension-row/column plans and their recurrences stay unchanged.
    auto *metadata = base ? static_cast<Matrix<Word> *>(card().base.data)
                          : engine_.plan(card(), domain);
    canonical_[slot]=false;
    engine_.vector_stage(card(), data(slot), metadata, domain, columns, inverse,
        v.columns, v.first, v.count);
    canonical_[slot]=true;
  }
  // Both views must describe the same evaluations in the same physical order.
  // Right is read-only; third slot and reusable window buffers are untouched.
  void pointwise(unsigned left,unsigned right) {
    data(left);data(right);
    const auto &a=views_[left], &b=views_[right];
    if(left==right || a.rows!=b.rows || a.columns!=b.columns || a.first!=b.first ||
        a.count!=b.count || a.interleaved!=b.interleaved)
      throw std::runtime_error("resident pointwise alias/view");
    validate(left,"rank-pointwise-left-validation");
    validate(right,"rank-pointwise-right-validation");
    const auto began=std::chrono::steady_clock::now();auto last=began;
    const char *phase="rank-pointwise-product";
    engine_.event(phase,"started",0,a.words(),began);
    canonical_[left]=false;
    try {
      for(size_t offset=0;offset<a.words();) {
        const size_t take=std::min<size_t>(1u<<20,a.words()-offset);
        liftrs_cuda::preparation::pointwise<<<(take+255)/256,256>>>(data(left)+offset,
            data(right)+offset,take,engine_.s_,engine_.delta_,
            static_cast<uint32_t*>(card().logs.data),static_cast<uint32_t*>(card().powers.data));
        sync();offset+=take;
        if(offset==a.words() || std::chrono::steady_clock::now()-last>=std::chrono::seconds(30)) {
          engine_.event(phase,offset==a.words()?"completed":"progress",offset,a.words(),began);last=std::chrono::steady_clock::now();
        }
      }
      if(!a.words())engine_.event(phase,"completed",0,0,began);
      canonical_[left]=true;
    } catch(...) {engine_.event(phase,"failed",0,a.words(),began);throw;}
  }
  std::vector<uint64_t> degrees(unsigned level) {
    prepare();
    if (level >= metadata_.dimension - metadata_.split) throw std::runtime_error("rank Taylor level");
    const auto &terms = metadata_.taylor[level];
    std::vector<uint64_t> out;
    for (unsigned t = 0; t < terms.count; t++) out.push_back(terms.degree[t]);
    return out;
  }
  // Snapshot is borrowed/read-only. Only updated targets go directly to the
  // resident array; neither sources nor a complete host shard are written back.
  void taylor(unsigned slot, const Word *snapshot, size_t words, size_t capacity, unsigned level,
      bool quotient, size_t block, size_t first, size_t count,bool local_sources=false) {
    data(slot); prepare();
    if (!snapshot || !words || words > capacity || capacity > (1u << 20) || !count || count > words ||
        level >= metadata_.dimension - metadata_.split || block >= rows_ * columns_)
      throw std::runtime_error("rank Taylor snapshot extent");
    const uint64_t limit = uint64_t(1) << (2 * engine_.s_);
    for (size_t j = 0; j < words; j++)
      if (uint64_t(snapshot[j]) >= limit) throw std::runtime_error("rank Taylor field value");
    const size_t half = size_t(1) << (metadata_.split + level);
    if (first >= half || count > half - first || block % (2 * half))
      throw std::runtime_error("rank Taylor block/interval");
    const auto &v = views_[slot];
    const size_t target = block + (quotient ? half : 0) + first, origin = v.first * v.columns;
    if (v.rows != rows_ || v.columns != columns_ || target < origin ||
        target - origin > v.words() || count > v.words() - (target - origin))
      throw std::runtime_error("rank Taylor target owner/view");
    // Validate local source data once per raw-write generation, not once per
    // window. Remote snapshot values were checked above; arithmetic is closed.
    const size_t admitted_half=engine_.taylor_shape(metadata_,rows_*columns_,capacity,
        words,level,quotient,block,first,count);
    if(local_sources && !canonical_[slot])validate(slot,"rank-Taylor-local-validation");
    engine_.ensure(card(), card().windows, capacity * sizeof(Word));
    auto *window=static_cast<Word*>(card().windows.data);
    check(cudaMemcpy(window,snapshot,words*sizeof(Word),cudaMemcpyHostToDevice));
    const bool valid=canonical_[slot];canonical_[slot]=false;
    TaylorWindow<Word> packed{window,plan_,block,admitted_half,first,count,level,quotient};
    if(local_sources) {
      LocalTaylorWindow<Word> view{packed,data(slot),v};
      engine_.taylor_launch(card(),view,rows_*columns_);
    } else {
      engine_.taylor_launch(card(),packed,rows_*columns_);
      rank_transfer<<<(count+255)/256,256>>>(data(slot),window,v,target-origin,count,true);
      sync();
    }
    canonical_[slot]=valid;
  }
  void tile(unsigned slot, unsigned peer, size_t offset, Word *values, size_t words, bool pack) {
    data(slot);
    if (peer >= ranks_ || peer == rank_ || !values || !words || words > tile_)
      throw std::runtime_error("rank storage peer/tile");
    const auto &v = views_[slot];
    const auto band = storage_band(v.columns, peer, ranks_);
    const size_t total = v.count * band.second;
    if (offset > total || words > total - offset) throw std::runtime_error("rank storage peer extent");
    check(cudaSetDevice(card().id));
    auto *buffer = static_cast<Word *>(card().transport.data);
    if (!pack) {canonical_[slot]=false;check(cudaMemcpy(buffer, values, words * sizeof(Word), cudaMemcpyHostToDevice));}
    rank_pack<<<(words + 255) / 256, 256>>>(data(slot), buffer, v, band.first, band.second, offset, words, pack);
    sync();
    if (pack) check(cudaMemcpy(values, buffer, words * sizeof(Word), cudaMemcpyDeviceToHost));
  }
};
using RankStorageHandle = std::variant<std::unique_ptr<RankStorage<uint32_t>>, std::unique_ptr<RankStorage<uint64_t>>>;
template <class Function> void visit_rank(void *handle, Function fn) {
  if (!handle) throw std::runtime_error("rank storage handle");
  std::visit([&](auto &state) {
    if (!state) throw std::runtime_error("rank storage state");
    fn(*state);
  }, *static_cast<RankStorageHandle *>(handle));
}
