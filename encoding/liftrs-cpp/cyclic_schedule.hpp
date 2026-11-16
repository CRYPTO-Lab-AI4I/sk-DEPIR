#pragma once
// Backend-independent cyclic encoding order. Callbacks join their work before
// returning; buffers may be reused only after their last completed operation.
#include "progress.hpp"

namespace liftrs::execution {
enum class Input { prefix_p, prefix_message, full };

// H=(P*M) mod X^K, Y=G*H, then phi(Y) and parity at coordinate zero.
// Every callback joins its workers before returning. That barrier is essential:
// transforms share scratch, and promotion deliberately reuses the prefix arena.
// Keeping this order here prevents resident/wide/file/disk implementations from
// drifting apart. Callbacks are inlined; no virtual calls or domain allocations.
template <class Prepare, class Transform, class PrefixProduct, class Promote,
          class SpectrumProduct, class Assemble>
void encode_cyclic(Prepare prepare, Transform transform, PrefixProduct prefix_product,
                   Promote promote, SpectrumProduct spectrum_product,
                   Assemble assemble) {
  EncodingProgress progress;
  progress.run("prepare-prefix", prepare);
  progress.run("prefix-p-forward", [&] { transform(Input::prefix_p, false); });
  progress.run("prefix-message-forward",
               [&] { transform(Input::prefix_message, false); });
  progress.run("prefix-product", prefix_product);
  progress.run("prefix-inverse", [&] { transform(Input::prefix_p, true); });
  progress.run("promote-prefix", promote);
  progress.run("full-forward", [&] { transform(Input::full, false); });
  progress.run("spectrum-product", spectrum_product);
  progress.run("full-inverse", [&] { transform(Input::full, true); });
  progress.run("canonical-assembly", assemble);
}

} // namespace liftrs::execution
