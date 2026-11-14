#pragma once
// Internal metadata shared by C ABI modules; never owns caller arrays.
#include "transform.hpp"

struct liftrs_plan {
  liftrs::Transform transform;
  std::vector<std::vector<liftrs::Term>>
      taylor_terms;       // Code-only maps prepared before encoding.
  size_t owned_bytes = 0; // Vector capacity payloads + descriptors, not allocator RSS.
};
