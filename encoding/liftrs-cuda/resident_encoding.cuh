// Host-side field validation streams the borrowed arrays without copying them.
// Report completed blocks so large setup/input checks do not become silent.
template <class Value>
void validate_encoding_values(const char *phase, const Value *values, size_t count,
                              uint64_t limit) {
  liftrs::execution::ProgressScope scope(phase);
  liftrs::execution::WorkProgress progress("validation-blocks",
                                           (count + 65535) / 65536);
  const bool log = liftrs::execution::progress_enabled();
  auto began = std::chrono::steady_clock::now();
  if (log)
    liftrs::execution::progress_record(phase, "validation-blocks", "started", 0,
                                       (count + 65535) / 65536, 0);
  for (size_t first = 0; first < count; first += 65536) {
    for (size_t i = first; i < std::min(count, first + 65536); i++)
      if (uint64_t(values[i]) >= limit) {
        if (log)
          liftrs::execution::progress_record(
              phase, "validation-blocks", "failed", first / 65536,
              (count + 65535) / 65536,
              std::chrono::duration_cast<std::chrono::milliseconds>(
                  std::chrono::steady_clock::now() - began)
                  .count());
        throw std::runtime_error("GPU encoding field value out of range");
      }
    progress.advance();
  }
  if (log)
    liftrs::execution::progress_record(
        phase, "validation-blocks", "completed", (count + 65535) / 65536,
        (count + 65535) / 65536,
        std::chrono::duration_cast<std::chrono::milliseconds>(
            std::chrono::steady_clock::now() - began)
            .count());
}

// Included once after Resident's shared execution helpers. This adapter owns no
// second transform implementation: it supplies buffers to the CPU/GPU common
// cyclic schedule. Workspaces/constants are accounted in the executor's Cards.
template <class Word>
void Resident<Word>::set_code(const uint32_t *p, size_t pc, const Word *spectrum,
                              size_t n, int layout) {
  shape(n);
  if (encoding_k_ || !p || !spectrum || pc < 2 || pc - 1 > n / 2 ||
      n != (size_t(1) << (2 * s_)) || (layout != 0 && layout != 1) || p[0] != 1)
    throw std::runtime_error("GPU encoding code/state/layout");
  // Only P[0:K] participates in the truncated product. Validate P[K] below,
  // then omit that unused endpoint from the copied device prefix constants.
  const size_t k = pc - 1;
  active_domain_ = n;
  validate_encoding_values("GPU-code-P-validation", p, pc, uint64_t(1) << s_);
  if constexpr (sizeof(Word) > 4)
    validate_encoding_values("GPU-code-spectrum-validation", spectrum, n,
                             uint64_t(1) << (2 * s_));
  else if (s_ < 16)
    validate_encoding_values("GPU-code-spectrum-validation", spectrum, n,
                             uint64_t(1) << (2 * s_));
  size_t l = 4;
  while (l < 2 * k - 1)
    l *= 2;
  const size_t arena = std::max(n, 2 * l), count = active(n), prefix_count = active(l),
               prefix_chunk = l / prefix_count;
  // Dedicated clients configure once. Admission of all live arrays/plans occurs
  // before principal allocations; metadata/runtime headroom remains additional.
  for (unsigned d = 0; d < count; d++) {
    auto &c = *cards_[d];
    if (c.x.bytes || c.y.bytes || c.scratch.bytes)
      throw std::runtime_error("GPU encoding requires unused executor workspace");
    plan(c, n);
    if (d < prefix_count)
      plan(c, l);
    const size_t first = d * prefix_chunk,
                 have = d < prefix_count && first < k
                            ? std::min(prefix_chunk, k - first)
                            : 0;
    const size_t jobs = (n / count + liftrs_cuda::encoding::canonical_chunk - 1) /
                        liftrs_cuda::encoding::canonical_chunk;
    admit(c, uint64_t(arena / count + n / count + n / count) * sizeof(Word) +
                 4 * uint64_t(have + jobs + 1));
  }
  const auto setup_began = Clock::now();
  auto setup_last = setup_began;
  size_t copied = 0;
  const size_t total_copy = k * 4 + n * sizeof(Word);
  event("GPU-code-setup-bytes", "started", 0, total_copy, setup_began);
  auto copy_constant = [&](void *destination, const void *source, size_t bytes) {
    // Logged runs expose completed transfer bytes without a second data pass.
    // The ordinary path retains one bulk copy; instrumentation cost is separate.
    const size_t tile =
        liftrs::execution::progress_enabled() ? size_t(256) << 20 : bytes;
    for (size_t first = 0; first < bytes; first += tile) {
      const size_t take = std::min(tile, bytes - first);
      check(cudaMemcpy(static_cast<unsigned char *>(destination) + first,
                       static_cast<const unsigned char *>(source) + first, take,
                       cudaMemcpyHostToDevice));
      copied += take;
      if (Clock::now() - setup_last >= std::chrono::seconds(30)) {
        event("GPU-code-setup-bytes", "progress", copied, total_copy, setup_began);
        setup_last = Clock::now();
      }
    }
  };
  for (unsigned d = 0; d < count; d++) {
    auto &c = *cards_[d];
    const size_t first = d * prefix_chunk,
                 have = d < prefix_count && first < k
                            ? std::min(prefix_chunk, k - first)
                            : 0;
    ensure(c, c.x, arena / count * sizeof(Word));
    ensure(c, c.scratch, n / count * sizeof(Word));
    if (have)
      ensure(c, c.code_p, have * 4);
    ensure(c, c.code_spectrum, n / count * sizeof(Word));
    ensure(c, c.parity,
           4 * ((n / count + liftrs_cuda::encoding::canonical_chunk - 1) /
                    liftrs_cuda::encoding::canonical_chunk +
                1));
    check(cudaSetDevice(c.id));
    if (have)
      copy_constant(c.code_p.data, p + first, have * 4);
    copy_constant(c.code_spectrum.data, spectrum + d * (n / count),
                  n / count * sizeof(Word));
  }
  event("GPU-code-setup-bytes", "completed", copied, total_copy, setup_began);
  encoding_k_ = k;
  encoding_l_ = l;
  local_vectors_ = layout != 0;
}

template <class Word>
void Resident<Word>::encode_code(const uint32_t *message, size_t k, uint32_t *word,
                                 size_t n, double *stats) {
  if (!message || !word || !stats || !encoding_k_ || k != encoding_k_ || n != maximum_)
    throw std::runtime_error("GPU encoding input/state/shape");
  // Reject overlap before any upload or host output mutation. Sizes fit 64-bit
  // admission (s<=17); the C ABI caller owns these complete host spans.
  const auto input = reinterpret_cast<uintptr_t>(message),
             output = reinterpret_cast<uintptr_t>(word);
  if ((input <= output && output - input < k * 4) ||
      (output < input && input - output < n * 4))
    throw std::runtime_error("GPU encoding input/output overlap");
  auto began = Clock::now();
  std::fill(stats, stats + 10, 0);
  validate_encoding_values("GPU-code-message-validation", message, k,
                           uint64_t(1) << s_);
  const size_t l = encoding_l_;
  const unsigned count = active(n), prefix_count = active(l);
  auto p = shards(l, 0), m = p, full = shards(n, 2), scratch = shards(n, 0);
  for (unsigned d = 0; d < prefix_count; d++)
    m.data[d] += p.chunk;
  auto timed = [&](unsigned slot, auto fn) {
    auto start = Clock::now();
    fn();
    stats[slot] += std::chrono::duration<double>(Clock::now() - start).count();
  };
  auto prepare = [&] {
    timed(8, [&] {
      stage("GPU-encoding-prefix", l, prefix_count, [&](unsigned d) {
        auto &c = *cards_[d];
        size_t first = d * p.chunk, have = first < k ? std::min(p.chunk, k - first) : 0;
        widen<<<(p.chunk + 255) / 256, 256>>>(static_cast<uint32_t *>(c.code_p.data),
                                              p.data[d], p.chunk, have);
      });
    });
    timed(0, [&] { upload_base(message, k, l, m); });
  };
  auto transform_input = [&](liftrs::execution::Input input, bool inverse) {
    if (input == liftrs::execution::Input::full)
      transform(n, full, scratch, inverse, stats);
    else
      transform(l, input == liftrs::execution::Input::prefix_p ? p : m, shards(l, 2),
                inverse, stats);
  };
  auto prefix_product = [&] {
    timed(6, [&] {
      stage("GPU-encoding-prefix-product", l, prefix_count, [&](unsigned d) {
        auto &c = *cards_[d];
        pointwise<<<(p.chunk + 255) / 256, 256>>>(
            p.data[d], m.data[d], p.chunk, s_, delta_,
            static_cast<uint32_t *>(c.logs.data),
            static_cast<uint32_t *>(c.powers.data));
      });
    });
  };
  auto promote = [&] {
    timed(4, [&] {
      stage("GPU-encoding-promote", n, count, [&](unsigned d) {
        liftrs_cuda::encoding::promote<<<(full.chunk + 255) / 256, 256>>>(
            p, full.data[d], d * full.chunk, full.chunk, k);
      });
    });
  };
  auto spectrum_product = [&] {
    timed(6, [&] {
      stage("GPU-encoding-spectrum-product", n, count, [&](unsigned d) {
        auto &c = *cards_[d];
        pointwise<<<(full.chunk + 255) / 256, 256>>>(
            full.data[d], static_cast<Word *>(c.code_spectrum.data), full.chunk, s_,
            delta_, static_cast<uint32_t *>(c.logs.data),
            static_cast<uint32_t *>(c.powers.data));
      });
    });
  };
  // Final coefficients are in old scratch. The dead arena receives disjoint
  // canonical uint32 symbols, so no extra full device output array is allocated.
  Shards<uint32_t> canonical;
  canonical.chunk = scratch.chunk;
  canonical.shift = scratch.shift;
  for (unsigned d = 0; d < count; d++)
    canonical.data[d] = reinterpret_cast<uint32_t *>(scratch.data[d]);
  auto assemble = [&] {
    timed(9, [&] {
      const size_t jobs = (full.chunk + liftrs_cuda::encoding::canonical_chunk - 1) /
                          liftrs_cuda::encoding::canonical_chunk;
      const auto start = Clock::now();
      auto last = start;
      event("GPU-encoding-assembly", "started", 0, jobs * count, start);
      for (unsigned d = 0; d < count; d++) {
        check(cudaSetDevice(cards_[d]->id));
        check(cudaMemsetAsync(cards_[d]->flag.data, 0, 4));
      }
      constexpr size_t wave = 32768;
      for (size_t first_job = 0; first_job < jobs; first_job += wave) {
        size_t take = std::min(wave, jobs - first_job);
        for (unsigned d = 0; d < count; d++) {
          auto &c = *cards_[d];
          check(cudaSetDevice(c.id));
          liftrs_cuda::encoding::canonical<<<(take + 127) / 128, 128>>>(
              full.data[d], canonical, d * full.chunk, full.chunk, n, omega_, s_,
              delta_, static_cast<uint32_t *>(c.logs.data),
              static_cast<uint32_t *>(c.powers.data),
              static_cast<uint32_t *>(c.parity.data),
              static_cast<uint32_t *>(c.flag.data), first_job, take);
          check(cudaGetLastError());
        }
        sync(count);
        if (Clock::now() - last >= std::chrono::seconds(30)) {
          event("GPU-encoding-assembly", "progress", (first_job + take) * count,
                jobs * count, start);
          last = Clock::now();
        }
      }
      stage("GPU-encoding-parity", count, count, [&](unsigned d) {
        liftrs_cuda::encoding::reduce_parity<<<1, 256>>>(
            static_cast<uint32_t *>(cards_[d]->parity.data), jobs);
      });
      event("GPU-encoding-assembly", "completed", jobs * count, jobs * count, start);
    });
    timed(5, [&] {
      uint32_t zero = 0;
      for (unsigned d = 0; d < count; d++) {
        auto &c = *cards_[d];
        uint32_t bad = 0, partial = 0;
        check(cudaSetDevice(c.id));
        check(cudaMemcpy(&bad, c.flag.data, 4, cudaMemcpyDeviceToHost));
        if (bad)
          throw std::runtime_error("GPU encoding product escaped base field");
        check(cudaMemcpy(&partial,
                         static_cast<uint32_t *>(c.parity.data) +
                             (full.chunk + liftrs_cuda::encoding::canonical_chunk - 1) /
                                 liftrs_cuda::encoding::canonical_chunk,
                         4, cudaMemcpyDeviceToHost));
        zero ^= partial;
      }
      check(cudaSetDevice(cards_[0]->id));
      check(cudaMemcpy(canonical.data[0], &zero, 4, cudaMemcpyHostToDevice));
      stage("GPU-encoding-download", n, count, [&](unsigned d) {
        check(cudaMemcpyAsync(word + d * canonical.chunk, canonical.data[d],
                              canonical.chunk * 4, cudaMemcpyDeviceToHost));
      });
    });
  };
  liftrs::execution::encode_cyclic(prepare, transform_input, prefix_product, promote,
                                   spectrum_product, assemble);
  stats[7] = std::chrono::duration<double>(Clock::now() - began).count();
}
