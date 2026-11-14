#pragma once
// Public code preparation only. It is not constant-time and sees no database.
// Base polynomial algorithms live in the shared Go layer; this executor owns
// width-selected transforms/products, batched roots and exact monomial discovery.
#include "online.hpp"
#include "prepare_plan.hpp"
#include <numeric>
#include <unordered_map>

namespace liftrs::preparation {

inline bool good_weights(const unsigned *weights, unsigned s, unsigned t) {
  // Lemma4.8: a POSITIVE reachable sum j*(2^s-1) makes the monomial bad.
  // Carry intervals implement that criterion without a q-sized DP per monomial.
  for (unsigned j = 1; j <= 2 * t; j++) {
    unsigned lo = j, hi = j;
    for (unsigned k = 0; k < s; k++) {
      lo = (lo + 1) / 2;
      hi = (hi + t * weights[k]) / 2;
      if (lo > hi)
        break;
    }
    if (lo <= j && j <= hi)
      return false;
  }
  return true;
}
inline bool good_pair(uint64_t index, unsigned s, unsigned t) {
  unsigned weights[20];
  for (unsigned k = 0; k < s; k++)
    weights[k] = unsigned((index >> k) & 1) + unsigned((index >> (s + k)) & 1);
  return good_weights(weights, s, t);
}

// No field-sized allocation at construction. Each native allocation is admitted
// against the explicit payload cap; metadata/vector capacity is included.
template <class Word> class Executor {
public:
  using Plan = TransformPlan<Word>;
  using Matrix = MatrixPlan<Word>;
  Executor(unsigned s, uint32_t polynomial, uint32_t delta, uint64_t omega,
           unsigned workers, size_t maximum, uint64_t cap)
      : workers_(workers), maximum_(maximum), cap_(cap), omega_(Word(omega)) {
    set_field(field_, s, polynomial, delta);
    if (!workers || workers > 64 || !maximum || (maximum & (maximum - 1)) ||
        maximum > field_.q || omega >= field_.q || !is_primitive(field_, omega_) ||
        !cap)
      throw std::runtime_error("invalid preparation parameters/limits");
    std::vector<Word> basis(s);
    for (unsigned j = 0; j < s; j++)
      basis[j] = Word(1) << j;
    base_ =
        build_transform<Plan>(s, polynomial, delta, basis.data(), basis.size(), true);
    auto inverse = field_power(field_, omega_, field_.q - 2);
    Word jump = 1;
    for (unsigned j = 0; j < 64; j++) {
      jumps_.push_back(constant_product(field_, jump));
      jump = field_.mul(jump, inverse);
    }
    admit(0);
  }
  unsigned width() const {
    return sizeof(Word);
  }
  uint64_t bytes() const {
    uint64_t value = sizeof(*this) + base_->owned_bytes;
    value += sizeof(Word) * (x_.capacity() + y_.capacity() + scratch_.capacity() +
                             columns_.capacity());
    value += 8 * good_.capacity() + weights_.capacity() + 4 * ternary_.capacity();
    value += jumps_.capacity() * sizeof(ProductT<Word>);
    for (const auto &p : jumps_)
      value += sizeof(Word) * p.rows.capacity();
    for (const auto &pair : matrices_)
      value += sizeof(Matrix) + pair.second->rows->owned_bytes +
               pair.second->columns->owned_bytes;
    return value;
  }
  void check_shape(size_t count) const {
    if (!count || (count & (count - 1)) || count > maximum_)
      throw std::runtime_error("preparation domain exceeds admitted limit");
  }
  void apply(Word *data, size_t n, bool inverse) {
    execution::ProgressScope scope(inverse ? "extension-interpolation"
                                           : "extension-evaluation");
    check_shape(n);
    if (!data)
      throw std::runtime_error("missing preparation buffer");
    for (size_t i = 0; i < n; i++)
      if (uint64_t(data[i]) >= field_.q)
        throw std::runtime_error("preparation value outside field");
    scratch(n);
    transform(data, n, inverse);
  }
  void base_grid(Word *grid, size_t n) {
    check_shape(n);
    if (!grid || n != field_.q)
      throw std::runtime_error("discovery requires a complete canonical grid");
    const size_t q = size_t(1) << field_.s;
    for (size_t i = 0; i < n; i++)
      if (uint64_t(grid[i]) >= q)
        throw std::runtime_error("discovery coefficient outside base field");
    const size_t active = std::min<size_t>(workers_, q), needed = active * q;
    if (columns_.size() < needed) {
      admit(needed * sizeof(Word));
      columns_ = std::vector<Word>(needed);
      admit(0);
    }
    execution::ProgressScope rows("discovery-base-rows");
    execution::parallel_jobs(
        q, workers_, [&](size_t row) { base_->transform.apply(grid + row * q, q); });
    execution::ProgressScope cols("discovery-base-columns");
    execution::WorkProgress progress("columns", q);
    execution::parallel_jobs(active, unsigned(active), [&](size_t worker) {
      auto *temporary = columns_.data() + worker * q;
      for (size_t col = worker; col < q; col += active) {
        for (size_t row = 0; row < q; row++)
          temporary[row] = grid[row * q + col];
        base_->transform.apply(temporary, q);
        for (size_t row = 0; row < q; row++)
          grid[row * q + col] = temporary[row];
        progress.advance();
      }
    });
  }
  void product(const uint32_t *a, size_t ac, const uint32_t *b, size_t bc,
               uint32_t *out, size_t limit, size_t n) {
    execution::ProgressScope scope("preparation-convolution");
    check_shape(n);
    if (!a || !b || !out || !ac || !bc || ac > n || bc > n || ac > n - bc + 1 ||
        !limit || limit > ac + bc - 1)
      throw std::runtime_error("preparation product shape");
    const uint64_t q = uint64_t(1) << field_.s;
    for (size_t i = 0; i < ac; i++)
      if (a[i] >= q)
        throw std::runtime_error("preparation base product input");
    for (size_t i = 0; i < bc; i++)
      if (b[i] >= q)
        throw std::runtime_error("preparation base product input");
    grow(n);
    const size_t chunk = 65536;
    execution::parallel_jobs((n + chunk - 1) / chunk, workers_, [&](size_t job) {
      for (size_t i = job * chunk; i < std::min(n, (job + 1) * chunk); i++) {
        x_[i] = i < ac ? a[i] : 0;
        y_[i] = i < bc ? b[i] : 0;
      }
    });
    transform(x_.data(), n, false);
    transform(y_.data(), n, false);
    execution::pointwise(field_, x_.data(), n, workers_, chunk,
                         [&](size_t i) { return y_[i]; });
    transform(x_.data(), n, true);
    std::atomic<bool> bad{false};
    execution::parallel_jobs((limit + chunk - 1) / chunk, workers_, [&](size_t job) {
      for (size_t i = job * chunk; i < std::min(limit, (job + 1) * chunk); i++) {
        if (uint64_t(x_[i]) >= q)
          bad.store(true, std::memory_order_relaxed);
        out[i] = uint32_t(x_[i]);
      }
    });
    if (bad.load())
      throw std::runtime_error("preparation product escaped base field");
  }
  void release_products() {
    drop(x_);
    drop(y_);
    drop(columns_);
  }
  void spectrum(const uint32_t *g, size_t count, Word *output, size_t n) {
    execution::ProgressScope scope("generator-spectrum");
    check_shape(n);
    if (!g || !output || count > n || n != field_.q)
      throw std::runtime_error("spectrum shape");
    const uint64_t q = uint64_t(1) << field_.s;
    for (size_t i = 0; i < count; i++)
      if (g[i] >= q)
        throw std::runtime_error("spectrum input outside base field");
    scratch(n);
    std::fill(output, output + n, Word(0));
    std::copy(g, g + count, output);
    transform(output, n, false);
  }
  void root_factors(const int64_t *indices, size_t count, uint32_t *output) {
    execution::ProgressScope scope("root-factor-batch");
    if (!indices || !output || !count)
      throw std::runtime_error("root batch buffers");
    const uint64_t order = field_.q - 1, q = uint64_t(1) << field_.s;
    for (size_t j = 0; j < count; j++)
      if (indices[j] < 0 || uint64_t(indices[j]) >= order ||
          (j && indices[j] <= indices[j - 1]))
        throw std::runtime_error("root coordinates must be sorted/distinct/in range");
    const size_t chunk = 2048;
    std::atomic<bool> bad{false};
    execution::parallel_jobs((count + chunk - 1) / chunk, workers_, [&](size_t job) {
      Word root = 0;
      uint64_t previous = 0;
      bool first = true;
      for (size_t j = job * chunk; j < std::min(count, (job + 1) * chunk); j++) {
        auto *factor = output + 4 * j;
        std::fill(factor, factor + 4, 0);
        const uint64_t exponent = uint64_t(indices[j]), partner = exponent * q % order;
        if (partner < exponent)
          continue;
        if (first) {
          root = field_power(field_, omega_, order - exponent);
          first = false;
        } else {
          auto gap = exponent - previous;
          root = gap < jumps_.size()
                     ? jumps_[gap].mul(field_, root)
                     : field_.mul(root, field_power(field_, omega_, order - gap));
        }
        previous = exponent;
        const uint32_t a = uint32_t(root & (q - 1)), b = uint32_t(root >> field_.s);
        if (partner == exponent) {
          if (b)
            bad.store(true, std::memory_order_relaxed);
          factor[0] = a;
          factor[1] = 1;
          factor[3] = 1;
        } else {
          // q-Frobenius sends u to u+1. Trace(a+b*u)=b and
          // norm=a²+a*b+delta*b²; no repeated q-power or inversion per root.
          factor[0] = field_.base_mul(a, a) ^ field_.base_mul(a, b) ^
                      field_.base_mul(field_.delta, field_.base_mul(b, b));
          factor[1] = b;
          factor[2] = 1;
          factor[3] = 2;
        }
      }
    });
    if (bad.load())
      throw std::runtime_error("fixed root escaped base field");
  }
  uint64_t count_good(unsigned t) {
    if (maximum_ != field_.q || !t || t >= (uint64_t(1) << field_.s) - 1)
      throw std::runtime_error("good enumeration requires full admitted domain/t");
    drop(good_);
    drop(weights_);
    drop(ternary_);
    size_t weight_count = 1;
    for (unsigned j = 0; j < field_.s; j++)
      weight_count *= 3;
    const size_t q = size_t(1) << field_.s, words = (field_.q + 63) / 64;
    admit(weight_count + q * 4 + words * 8);
    weights_ = std::vector<unsigned char>(weight_count);
    ternary_ = std::vector<uint32_t>(q);
    good_ = std::vector<uint64_t>(words);
    admit(0);
    execution::ProgressScope weights_scope("good-weight-vectors");
    const size_t chunk = 4096;
    execution::parallel_jobs(
        (weight_count + chunk - 1) / chunk, workers_, [&](size_t job) {
          for (size_t i = job * chunk; i < std::min(weight_count, (job + 1) * chunk);
               i++) {
            size_t value = i;
            unsigned w[20];
            for (unsigned k = 0; k < field_.s; k++) {
              w[k] = value % 3;
              value /= 3;
            }
            weights_[i] = good_weights(w, field_.s, t);
          }
        });
    // Paper Algorithm1: goodness depends only on the ternary bit-weight vector.
    // T(x)+T(y) indexes that vector, avoiding a carry DP for every grid point.
    uint32_t power = 1;
    for (unsigned bit = 0; bit < field_.s; bit++, power *= 3)
      for (size_t i = size_t(1) << bit; i < (size_t(1) << (bit + 1)); i++)
        ternary_[i] = power + ternary_[i - (size_t(1) << bit)];
    execution::ProgressScope expand("good-monomial-expansion");
    std::atomic<uint64_t> total{0};
    execution::parallel_jobs((words + chunk - 1) / chunk, workers_, [&](size_t job) {
      uint64_t local = 0;
      for (size_t i = job * chunk; i < std::min(words, (job + 1) * chunk); i++) {
        uint64_t bits = 0;
        for (unsigned j = 0; j < 64 && i * 64 + j < field_.q; j++) {
          const size_t index = i * 64 + j, x = index & (q - 1), y = index >> field_.s;
          if (weights_[size_t(ternary_[x]) + ternary_[y]])
            bits |= uint64_t(1) << j;
        }
        good_[i] = bits;
        local += __builtin_popcountll(bits);
      }
      total.fetch_add(local, std::memory_order_relaxed);
    });
    good_count_ = total.load();
    drop(weights_);
    drop(ternary_);
    return good_count_;
  }
  void fill_good(int64_t *out, size_t count) {
    if (!out || good_.empty() || count != good_count_)
      throw std::runtime_error("good enumeration state/count");
    size_t position = 0;
    execution::WorkProgress progress("bitset-words", good_.size());
    for (size_t i = 0; i < good_.size(); i++) {
      auto bits = good_[i];
      while (bits) {
        out[position++] = int64_t(64 * i + __builtin_ctzll(bits));
        bits &= bits - 1;
      }
      progress.advance();
    }
    if (position != count)
      throw std::runtime_error("good enumeration count mismatch");
    drop(good_);
  }

private:
  template <class Value> static void drop(std::vector<Value> &v) {
    std::vector<Value>().swap(v);
  }
  void admit(uint64_t extra) const {
    auto current = bytes();
    if (current > cap_ || extra > cap_ - current)
      throw std::runtime_error("native preparation byte cap exceeded");
  }
  Matrix &matrix(size_t n) {
    auto it = matrices_.find(n);
    if (it != matrices_.end())
      return *it->second;
    unsigned dim = 0;
    for (auto v = n; v > 1; v >>= 1)
      dim++;
    auto plan =
        build_matrix<Matrix>(field_.s, field_.polynomial, field_.delta, dim, dim / 2);
    matrices_.emplace(n, std::move(plan));
    admit(0);
    return *matrices_.at(n);
  }
  void scratch(size_t n) {
    if (scratch_.size() >= n)
      return;
    drop(scratch_);
    admit(n * sizeof(Word));
    scratch_ = std::vector<Word>(n);
    admit(0);
  }
  void grow(size_t n) {
    if (x_.size() >= n && y_.size() >= n) {
      scratch(n);
      return;
    }
    release_products();
    drop(scratch_);
    admit(3 * uint64_t(n) * sizeof(Word));
    x_ = std::vector<Word>(n);
    y_ = std::vector<Word>(n);
    scratch_ = std::vector<Word>(n);
    admit(0);
  }
  void transform(Word *data, size_t n, bool inverse) {
    if (n == 1)
      return;
    if (n == 2) {
      base_->transform.apply(data, n, inverse);
      return;
    }
    auto &plan = matrix(n);
    uint64_t stages[3];
    execution::matrix(false, n, plan.rows.get(), plan.columns.get(), data,
                      scratch_.data(), n, plan.split, workers_, inverse, stages);
  }
  FieldT<Word> field_;
  unsigned workers_;
  size_t maximum_;
  uint64_t cap_;
  Word omega_;
  std::unique_ptr<Plan> base_;
  std::vector<ProductT<Word>> jumps_;
  std::unordered_map<size_t, std::unique_ptr<Matrix>> matrices_;
  std::vector<Word> x_, y_, scratch_, columns_;
  std::vector<uint64_t> good_;
  std::vector<unsigned char> weights_;
  std::vector<uint32_t> ternary_;
  uint64_t good_count_ = 0;
};
} // namespace liftrs::preparation
