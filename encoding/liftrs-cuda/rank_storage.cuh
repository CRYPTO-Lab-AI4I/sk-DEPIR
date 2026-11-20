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
template <class Word> class RankStorage {
  Resident<Word> &engine_;
  size_t rows_, columns_, capacity_, tile_;
  unsigned rank_, ranks_;
  RankView views_[3];
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
    if (c.x.bytes || c.y.bytes || c.scratch.bytes || c.transport.bytes)
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
    if (next.words()) check(cudaMemset(data(slot), 0, next.words() * sizeof(Word)));
    check(cudaDeviceSynchronize());
    views_[slot] = next;
  }
  void transfer(unsigned slot, size_t offset, Word *values, size_t words, bool upload) {
    extent(slot, offset, words);
    if (!values) throw std::runtime_error("rank storage host pointer");
    check(cudaSetDevice(card().id));
    auto *tile = static_cast<Word *>(card().transport.data);
    if (upload) check(cudaMemcpy(tile, values, words * sizeof(Word), cudaMemcpyHostToDevice));
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
    if (!pack) check(cudaMemcpy(buffer, values, words * sizeof(Word), cudaMemcpyHostToDevice));
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
