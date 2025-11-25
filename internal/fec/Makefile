# Makefile for FEC XOR SIMD C++ Library
# Supports: AVX2 (baseline), AVX-512, NEON (ARM64), Scalar fallback

.PHONY: all clean test benchmark

# Compiler detection
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Linux)
    OS := linux
else ifeq ($(UNAME_S),Darwin)
    OS := darwin
else
    OS := unknown
endif

ifeq ($(UNAME_M),x86_64)
    ARCH := amd64
else ifeq ($(UNAME_M),arm64)
    ARCH := arm64
else ifeq ($(UNAME_M),aarch64)
    ARCH := arm64
else
    ARCH := unknown
endif

# Compiler flags
CXX := clang++
CXXFLAGS := -O3 -std=c++17 -fPIC -fvisibility=hidden -Wall -Wextra
LDFLAGS := -shared -lstdc++

# Platform-specific flags
ifeq ($(OS),darwin)
    LDFLAGS += -undefined dynamic_lookup
endif

# Target libraries
TARGET_AVX2 := libfec_avx2.so
TARGET_AVX512 := libfec_avx512.so
TARGET_NEON := libfec_neon.so
TARGET_SCALAR := libfec_scalar.so

# On macOS, use .dylib instead of .so
ifeq ($(OS),darwin)
    TARGET_AVX2 := libfec_avx2.dylib
    TARGET_AVX512 := libfec_avx512.dylib
    TARGET_NEON := libfec_neon.dylib
    TARGET_SCALAR := libfec_scalar.dylib
    LDFLAGS := -dynamiclib -lstdc++
endif

# Default: build appropriate variant for current platform
ifeq ($(ARCH),amd64)
    # x86_64: Build AVX2 as baseline
    all: $(TARGET_AVX2) $(TARGET_SCALAR)

    $(TARGET_AVX2): fec_xor_simd.cpp fec_xor_simd.h
	@echo "Building AVX2 baseline (x86_64)..."
	$(CXX) $(CXXFLAGS) -mavx2 -o $@ $< $(LDFLAGS)

    $(TARGET_AVX512): fec_xor_simd.cpp fec_xor_simd.h
	@echo "Building AVX-512 variant (x86_64, experimental)..."
	$(CXX) $(CXXFLAGS) -mavx512f -mavx512bw -o $@ $< $(LDFLAGS)

else ifeq ($(ARCH),arm64)
    # ARM64: Build NEON + Scalar
    all: $(TARGET_NEON) $(TARGET_SCALAR)

    $(TARGET_NEON): fec_xor_simd.cpp fec_xor_simd.h
	@echo "Building NEON (ARM64)..."
	$(CXX) $(CXXFLAGS) -march=armv8-a+simd -o $@ $< $(LDFLAGS)

else
    # Unknown architecture: Scalar only
    all: $(TARGET_SCALAR)
endif

$(TARGET_SCALAR): fec_xor_simd.cpp fec_xor_simd.h
	@echo "Building scalar fallback..."
	$(CXX) $(CXXFLAGS) -DFEC_SCALAR_ONLY -o $@ $< $(LDFLAGS)

clean:
	rm -f libfec_*.so libfec_*.dylib *.o *.a
	rm -f fec_test fec_benchmark

# Testing (C++ unit tests)
test: fec_test
	./fec_test

fec_test: all fec_test.cpp fec_xor_simd.h
	@echo "Building C++ tests..."
	$(CXX) $(CXXFLAGS) -o $@ fec_test.cpp fec_xor_simd.cpp

# Benchmarking
benchmark: all fec_benchmark.cpp
	@echo "Building benchmark..."
	$(CXX) $(CXXFLAGS) -O3 -o fec_benchmark fec_benchmark.cpp fec_xor_simd.cpp
	@echo "Running benchmark..."
	./fec_benchmark

# Info
info:
	@echo "Platform: $(OS)-$(ARCH)"
	@echo "Compiler: $(CXX)"
	@echo "CXXFLAGS: $(CXXFLAGS)"
	@echo "Targets: all clean test benchmark info"

