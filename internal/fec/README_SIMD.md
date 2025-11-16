# FEC XOR SIMD - High-Performance Forward Error Correction

## Overview

This implementation provides high-performance Forward Error Correction (FEC) using XOR-based encoding with SIMD (Single Instruction Multiple Data) optimizations.

**Key Features:**
- ✅ **34-35x speedup** on ARM64 (NEON) vs scalar
- ✅ **AVX2 baseline** for x86_64 (wide compatibility)
- ✅ **AVX-512 support** with runtime dispatch (optional)
- ✅ **NEON support** for ARM64 (Graviton/Ampere/Apple Silicon)
- ✅ **Batch API** for minimal CGO overhead
- ✅ **Flat-slab memory** layout for cache efficiency
- ✅ **Bit-exact determinism** across all platforms

## Performance Results

### Benchmark Results (macOS ARM64 M2 Max)

```
Platform: ARM64 (Apple Silicon M2 Max)
Selected implementation: NEON

Test: 10 packets × 1200 bytes (standard QUIC)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Scalar (reference):   4.70 µs/group  →   20 Gbps
NEON (selected):      0.14 µs/group  →  702 Gbps
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Speedup: 34.43x

Test: 10 packets × 9000 bytes (jumbo frames)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Scalar (reference):  31.89 µs/group  →   22 Gbps
NEON (selected):      0.98 µs/group  →  731 Gbps
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Speedup: 32.41x

Test: 5 packets × 1200 bytes (small group)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Scalar (reference):   2.34 µs/group  →   20 Gbps
NEON (selected):      0.07 µs/group  →  729 Gbps
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Speedup: 35.57x
```

## Architecture

### Components

**C++ Layer (`fec_xor_simd.cpp`):**
- `xor_packets_avx2()` - AVX2 implementation (256-bit SIMD)
- `xor_packets_avx512()` - AVX-512 implementation (512-bit SIMD)
- `xor_packets_neon()` - ARM64 NEON implementation (128-bit SIMD)
- `xor_packets_scalar()` - Fallback scalar implementation
- `fec_select_xor_impl()` - Runtime feature detection and dispatch
- Batch API for CGO integration

**Go Layer (`fec_cgo.go`):**
- `FECEncoderCXX` - Wrapper around C++ encoder
- Flat-slab memory management
- Batch processing to minimize CGO overhead
- Automatic fallback to Go implementation if C++ unavailable

**Hybrid Layer (`encoder_hybrid.go`):**
- `HybridFECEncoder` - Automatic C++/Go selection
- Plug-and-play replacement for existing `FECEncoder`
- Maintains API compatibility

### Memory Layout

**Flat-Slab Architecture:**
```
┌─────────────────────────────────────┐
│ FEC Encoder Context                 │
└─────────────────────────────────────┘
           │
           ├─→ Flat Slab Buffer (C-managed)
           │   ┌────────────────────────────┐
           │   │ Packet 1 (1200 bytes)      │
           │   ├────────────────────────────┤
           │   │ Packet 2 (1200 bytes)      │
           │   ├────────────────────────────┤
           │   │ ...                         │
           │   └────────────────────────────┘
           │
           ├─→ Offset Table (Go-managed)
           │   ┌────────────────────────────┐
           │   │ [0, 1200, 2400, ...]       │
           │   └────────────────────────────┘
           │
           └─→ Repair Buffer (C-managed, 64-byte aligned)
               ┌────────────────────────────┐
               │ Repair packet output       │
               │ (aligned for stream stores)│
               └────────────────────────────┘
```

## Building

### macOS / Linux

```bash
cd cloudbridge/quic-test/internal/fec

# Build all variants
make all

# Run unit tests
make test

# Run performance benchmarks
make benchmark

# Clean
make clean
```

### Supported Platforms

| Platform | Architecture | SIMD | Status |
|----------|--------------|------|--------|
| Linux | x86_64 | AVX2, AVX-512 | ✅ Supported |
| Linux | ARM64 | NEON | ✅ Supported |
| macOS | x86_64 | AVX2 | ✅ Supported |
| macOS | ARM64 | NEON | ✅ Supported |
| Other | Any | Scalar | ✅ Fallback |

## Usage

### Option 1: Hybrid Encoder (Recommended)

```go
import "cloudbridge/quic-test/internal/fec"

// Create encoder (auto-selects C++ or Go)
encoder := fec.NewHybridFECEncoder(0.10)  // 10% redundancy

// Add packets
hasRedundancy, repair, err := encoder.AddPacket(packetData, packetID)
if hasRedundancy {
    // Send repair packet
}

// Check if using C++
if encoder.UseCXX() {
    println("Using high-performance C++ encoder")
}

// Clean up
encoder.Close()
```

### Option 2: C++ Encoder Directly

```go
// Create C++ encoder (will panic if C++ libs not available)
encoder := fec.NewFECEncoderCXX(0.10, 1024)
defer encoder.Close()

// Batch encode groups
groups := []fec.FECGroup{...}
repairs, err := encoder.EncodeBatch(groups, 1200)
```

### Option 3: Go Encoder (Original)

```go
encoder := fec.NewFECEncoder(0.10)
hasRedundancy, repair, err := encoder.AddPacket(packetData, packetID)
```

## Testing

### Unit Tests

```bash
make test
```

Validates:
- ✅ Bit-exact output across all implementations
- ✅ Correctness for various packet sizes (1200, 1234, 1500, 9000 bytes)
- ✅ Edge cases (empty packets, non-SIMD-aligned sizes)
- ✅ Memory alignment and allocation

### Performance Benchmarks

```bash
make benchmark
```

Benchmarks:
- Standard QUIC (10 packets × 1200 bytes)
- Jumbo frames (10 packets × 9000 bytes)
- Small groups (5 packets × 1200 bytes)

## Integration with CloudBridge

### Step 1: Build C++ Library

```bash
cd cloudbridge/quic-test/internal/fec
make all
```

### Step 2: Update Client Code

Replace calls to `NewFECEncoder()` with `NewHybridFECEncoder()`:

```go
// Before:
// encoder := fec.NewFECEncoder(redundancy)

// After:
encoder := fec.NewHybridFECEncoder(redundancy)
```

### Step 3: Monitor Usage

Check encoder type in metrics:

```go
metrics := encoder.GetMetrics()
if encoder.UseCXX() {
    // Log: "FEC acceleration enabled (C++ SIMD)"
} else {
    // Log: "FEC using Go implementation"
}
```

## Performance Tuning

### Memory Allocation

The encoder pre-allocates memory for:
- Slab: `maxGroups × 10 packets × 1200 bytes` (default: 120 MB)
- Repair buffer: `maxGroups × 1200 bytes` (default: 12 MB)

Adjust if needed:

```go
encoder := fec.NewFECEncoderCXX(redundancy, maxGroups)
```

### Stream Stores

For large packets (>4096 bytes), non-temporal memory stores are used to avoid cache pollution:

```cpp
const size_t stream_threshold = 4096;
bool use_stream = (packet_size >= stream_threshold);
```

This is automatic - no tuning required.

### Prefetching

Software prefetch is used for better cache behavior:

```cpp
const size_t prefetch_distance = 256;  // 2-3 L1 cache lines ahead
```

## Troubleshooting

### C++ Library Not Loading

If the hybrid encoder falls back to Go:

```bash
# Check if libraries were built
ls -la cloudbridge/quic-test/internal/fec/libfec_*.{so,dylib}

# Rebuild
make clean && make all

# Verify symbols
nm libfec_neon.dylib | grep fec_encode
```

### Performance Not as Expected

Profile to verify execution path:

```bash
# Check which implementation is selected
encoder := fec.NewHybridFECEncoder(0.10)
if encoder.UseCXX() {
    // Using C++
} else {
    // Using Go (check why)
}
```

### Memory Issues

Pre-allocated memory is automatically resized if needed. Monitor:

```go
metrics := encoder.GetMetrics()
println("Packets encoded:", metrics.PacketsEncoded)
println("Redundancy bytes:", metrics.RedundancyBytes)
```

## Files

| File | Purpose |
|------|---------|
| `fec_xor_simd.h` | C-ABI header (public API) |
| `fec_xor_simd.cpp` | C++ implementation (AVX2/AVX-512/NEON/Scalar) |
| `fec_cgo.go` | Go CGO bindings and memory management |
| `encoder_hybrid.go` | Hybrid C++/Go encoder wrapper |
| `Makefile` | Build system |
| `fec_test.cpp` | C++ unit tests |
| `fec_benchmark.cpp` | Performance benchmarks |
| `encoder.go` | Original Go encoder (unchanged) |
| `decoder.go` | Original Go decoder (unchanged) |

## Future Improvements

### P2 (Next Phase): Reed-Solomon FEC

Implement RS-FEC for k-loss recovery (vs XOR's 1-loss limit):

```
Expected: 5-10x speedup vs Go
Gate: fec_recovered > 0, residual_loss < 1%
Timeline: 1-2 months
```

### Advanced Optimizations

- [ ] AVX-512 frequency throttling detection
- [ ] NUMA-aware memory binding (Linux)
- [ ] Hugepages support (2MB pages)
- [ ] Multipath QUIC support
- [ ] Post-quantum resistant cipher integration

## References

- [ISA-L (Intel Storage Acceleration Library)](https://github.com/intel/isa-l) - RS-FEC reference
- [QUIC RFC 9000](https://tools.ietf.org/html/rfc9000) - Protocol spec
- [MASQUE RFC 9298](https://tools.ietf.org/html/rfc9298) - UDP proxy protocol
- [ARM NEON Intrinsics](https://developer.arm.com/architectures/instruction-sets/intrinsics/) - ARM SIMD docs

## License

Same as CloudBridge project

