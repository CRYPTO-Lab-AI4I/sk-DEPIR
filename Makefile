# Standard commands prefer C++ transforms. The shared encoder API itself is
# unchanged: callers still choose WithCPPTransforms explicitly.
GO ?= go
PYTHON ?= python3
CXX ?= c++
BACKEND ?= cpp
RUN_NAME ?= native-$(shell date -u +%Y%m%d-%H%M%S)
export CXX

.PHONY: test test-go legacy-rme encode
legacy-rme:
	$(MAKE) -C encoding/rme_legacy rme

# The full suites include legacy file-based tests, which require this binary.
test: legacy-rme
	@test "$$($(GO) env CGO_ENABLED)" = 1 || { echo "C++ tests require CGO_ENABLED=1; use make test-go for pure Go"; exit 1; }
	$(GO) test -tags liftrs_cpp ./...

test-go: legacy-rme
	CGO_ENABLED=0 $(GO) test ./...

# Runner builds the native tag and selects C++ unless BACKEND=go or both.
# Fresh names preserve previous profiles/results; s=13 preprocessing must exist.
encode:
	$(PYTHON) experiments/native-encoding/run.py --go $(GO) --backend $(BACKEND) --run-name $(RUN_NAME)
