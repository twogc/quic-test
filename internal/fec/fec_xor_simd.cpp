#include "fec_xor_simd.h"
#include <cstring>
#include <cstdlib>
#include <algorithm>
#include <stdexcept>

// Platform-specific SIMD headers
#ifdef __x86_64__
    #include <immintrin.h>  // AVX2, AVX-512
    #include <cpuid.h>
#elif defined(__aarch64__)
    #include <arm_neon.h>
#endif

// NUMA support for Linux
#ifdef __linux__
    #include <numaif.h>
    #include <numa.h>
#endif

// ============================================================================
// Feature detection (x86 and ARM)
// ============================================================================

#ifdef __x86_64__
struct CPUFeatures {
    bool has_avx2;
    bool has_avx512f;
    bool has_avx512bw;
};

static CPUFeatures detect_cpu_features() {
    CPUFeatures features = {false, false, false};

    unsigned int eax, ebx, ecx, edx;

    // Get CPUID leaf 1 (ECX bit 28 = AVX)
    if (__get_cpuid(1, &eax, &ebx, &ecx, &edx)) {
        features.has_avx2 = (ecx & (1 << 28)) != 0;  // AVX bit in ECX
    }

    // Get CPUID leaf 7 (EBX bits for AVX2, AVX-512)
    if (__get_cpuid_count(7, 0, &eax, &ebx, &ecx, &edx)) {
        features.has_avx2 = (ebx & (1 << 5)) != 0;   // AVX2 bit in EBX
        features.has_avx512f = (ebx & (1 << 16)) != 0;  // AVX-512F in EBX
        features.has_avx512bw = (ebx & (1 << 30)) != 0; // AVX-512BW in EBX
    }

    return features;
}
#elif defined(__aarch64__)
struct CPUFeatures {
    bool has_neon;
};

static CPUFeatures detect_cpu_features() {
    // ARM64 always has NEON in aarch64
    return {true};
}
#else
struct CPUFeatures {
    // Fallback: no SIMD
};

static CPUFeatures detect_cpu_features() {
    return {};
}
#endif

// ============================================================================
// AVX2 Implementation (32-byte SIMD width, baseline for x86_64)
// ============================================================================

#ifdef __x86_64__
void xor_packets_avx2(
    const uint8_t* packets[],
    size_t num_packets,
    size_t packet_size,
    uint8_t* repair
) {
    if (num_packets == 0 || packet_size == 0) {
        return;
    }

    const size_t simd_width = 32;  // AVX2 = 256 bits = 32 bytes
    const size_t prefetch_distance = 256;

    // Check alignment for stream stores (optional optimization)
    bool repair_aligned = (reinterpret_cast<uintptr_t>(repair) % 32 == 0);
    bool use_stream = repair_aligned && (packet_size >= 4096);

    size_t i = 0;

    // Main loop: unroll x4 (process 128 bytes per iteration)
    for (; i + simd_width * 4 <= packet_size; i += simd_width * 4) {
        // Prefetch for better cache behavior
        if (i + prefetch_distance < packet_size) {
            for (size_t p = 0; p < num_packets; p++) {
                _mm_prefetch(
                    reinterpret_cast<const char*>(packets[p] + i + prefetch_distance),
                    _MM_HINT_T0
                );
            }
        }

        // Load first 128 bytes from packet 0
        __m256i result0 = _mm256_loadu_si256(
            reinterpret_cast<const __m256i*>(packets[0] + i)
        );
        __m256i result1 = _mm256_loadu_si256(
            reinterpret_cast<const __m256i*>(packets[0] + i + 32)
        );
        __m256i result2 = _mm256_loadu_si256(
            reinterpret_cast<const __m256i*>(packets[0] + i + 64)
        );
        __m256i result3 = _mm256_loadu_si256(
            reinterpret_cast<const __m256i*>(packets[0] + i + 96)
        );

        // XOR all remaining packets
        for (size_t p = 1; p < num_packets; p++) {
            result0 = _mm256_xor_si256(result0, _mm256_loadu_si256(
                reinterpret_cast<const __m256i*>(packets[p] + i)
            ));
            result1 = _mm256_xor_si256(result1, _mm256_loadu_si256(
                reinterpret_cast<const __m256i*>(packets[p] + i + 32)
            ));
            result2 = _mm256_xor_si256(result2, _mm256_loadu_si256(
                reinterpret_cast<const __m256i*>(packets[p] + i + 64)
            ));
            result3 = _mm256_xor_si256(result3, _mm256_loadu_si256(
                reinterpret_cast<const __m256i*>(packets[p] + i + 96)
            ));
        }

        // Store results
        if (use_stream) {
            _mm256_stream_si256(
                reinterpret_cast<__m256i*>(repair + i),
                result0
            );
            _mm256_stream_si256(
                reinterpret_cast<__m256i*>(repair + i + 32),
                result1
            );
            _mm256_stream_si256(
                reinterpret_cast<__m256i*>(repair + i + 64),
                result2
            );
            _mm256_stream_si256(
                reinterpret_cast<__m256i*>(repair + i + 96),
                result3
            );
        } else {
            _mm256_storeu_si256(
                reinterpret_cast<__m256i*>(repair + i),
                result0
            );
            _mm256_storeu_si256(
                reinterpret_cast<__m256i*>(repair + i + 32),
                result1
            );
            _mm256_storeu_si256(
                reinterpret_cast<__m256i*>(repair + i + 64),
                result2
            );
            _mm256_storeu_si256(
                reinterpret_cast<__m256i*>(repair + i + 96),
                result3
            );
        }
    }

    // Remaining AVX2 iterations (handle 32-byte chunks)
    for (; i + simd_width <= packet_size; i += simd_width) {
        __m256i result = _mm256_loadu_si256(
            reinterpret_cast<const __m256i*>(packets[0] + i)
        );

        for (size_t p = 1; p < num_packets; p++) {
            result = _mm256_xor_si256(result, _mm256_loadu_si256(
                reinterpret_cast<const __m256i*>(packets[p] + i)
            ));
        }

        if (use_stream) {
            _mm256_stream_si256(reinterpret_cast<__m256i*>(repair + i), result);
        } else {
            _mm256_storeu_si256(reinterpret_cast<__m256i*>(repair + i), result);
        }
    }

    // Memory fence if using stream stores
    if (use_stream) {
        _mm_sfence();
    }

    // Scalar tail for remaining bytes
    for (; i < packet_size; i++) {
        repair[i] = packets[0][i];
        for (size_t p = 1; p < num_packets; p++) {
            repair[i] ^= packets[p][i];
        }
    }
}
#endif  // __x86_64__

// ============================================================================
// AVX-512 Implementation (64-byte SIMD width, high-performance variant)
// ============================================================================

#if defined(__x86_64__) && defined(__AVX512F__)
void xor_packets_avx512(
    const uint8_t* packets[],
    size_t num_packets,
    size_t packet_size,
    uint8_t* repair
) {
    if (num_packets == 0 || packet_size == 0) {
        return;
    }

    const size_t simd_width = 64;  // AVX-512 = 512 bits = 64 bytes
    const size_t prefetch_distance = 256;

    bool repair_aligned = (reinterpret_cast<uintptr_t>(repair) % 64 == 0);
    bool use_stream = repair_aligned && (packet_size >= 4096);

    size_t i = 0;

    // Main loop: unroll x4 (process 256 bytes per iteration)
    for (; i + simd_width * 4 <= packet_size; i += simd_width * 4) {
        if (i + prefetch_distance < packet_size) {
            for (size_t p = 0; p < num_packets; p++) {
                _mm_prefetch(
                    reinterpret_cast<const char*>(packets[p] + i + prefetch_distance),
                    _MM_HINT_T0
                );
            }
        }

        // Load from packet 0
        __m512i result0 = _mm512_loadu_si512(
            reinterpret_cast<const __m512i*>(packets[0] + i)
        );
        __m512i result1 = _mm512_loadu_si512(
            reinterpret_cast<const __m512i*>(packets[0] + i + 64)
        );
        __m512i result2 = _mm512_loadu_si512(
            reinterpret_cast<const __m512i*>(packets[0] + i + 128)
        );
        __m512i result3 = _mm512_loadu_si512(
            reinterpret_cast<const __m512i*>(packets[0] + i + 192)
        );

        // XOR all remaining packets
        for (size_t p = 1; p < num_packets; p++) {
            result0 = _mm512_xor_si512(result0, _mm512_loadu_si512(
                reinterpret_cast<const __m512i*>(packets[p] + i)
            ));
            result1 = _mm512_xor_si512(result1, _mm512_loadu_si512(
                reinterpret_cast<const __m512i*>(packets[p] + i + 64)
            ));
            result2 = _mm512_xor_si512(result2, _mm512_loadu_si512(
                reinterpret_cast<const __m512i*>(packets[p] + i + 128)
            ));
            result3 = _mm512_xor_si512(result3, _mm512_loadu_si512(
                reinterpret_cast<const __m512i*>(packets[p] + i + 192)
            ));
        }

        // Store results
        if (use_stream) {
            _mm512_stream_si512(
                reinterpret_cast<__m512i*>(repair + i),
                result0
            );
            _mm512_stream_si512(
                reinterpret_cast<__m512i*>(repair + i + 64),
                result1
            );
            _mm512_stream_si512(
                reinterpret_cast<__m512i*>(repair + i + 128),
                result2
            );
            _mm512_stream_si512(
                reinterpret_cast<__m512i*>(repair + i + 192),
                result3
            );
        } else {
            _mm512_storeu_si512(
                reinterpret_cast<__m512i*>(repair + i),
                result0
            );
            _mm512_storeu_si512(
                reinterpret_cast<__m512i*>(repair + i + 64),
                result1
            );
            _mm512_storeu_si512(
                reinterpret_cast<__m512i*>(repair + i + 128),
                result2
            );
            _mm512_storeu_si512(
                reinterpret_cast<__m512i*>(repair + i + 192),
                result3
            );
        }
    }

    // Remaining AVX-512 iterations
    for (; i + simd_width <= packet_size; i += simd_width) {
        __m512i result = _mm512_loadu_si512(
            reinterpret_cast<const __m512i*>(packets[0] + i)
        );

        for (size_t p = 1; p < num_packets; p++) {
            result = _mm512_xor_si512(result, _mm512_loadu_si512(
                reinterpret_cast<const __m512i*>(packets[p] + i)
            ));
        }

        if (use_stream) {
            _mm512_stream_si512(reinterpret_cast<__m512i*>(repair + i), result);
        } else {
            _mm512_storeu_si512(reinterpret_cast<__m512i*>(repair + i), result);
        }
    }

    // Memory fence
    if (use_stream) {
        _mm_sfence();
    }

    // Scalar tail
    for (; i < packet_size; i++) {
        repair[i] = packets[0][i];
        for (size_t p = 1; p < num_packets; p++) {
            repair[i] ^= packets[p][i];
        }
    }
}
#endif  // __x86_64__ && __AVX512F__

// ============================================================================
// ARM64 NEON Implementation (128-bit SIMD width)
// ============================================================================

#ifdef __aarch64__
void xor_packets_neon(
    const uint8_t* packets[],
    size_t num_packets,
    size_t packet_size,
    uint8_t* repair
) {
    if (num_packets == 0 || packet_size == 0) {
        return;
    }

    const size_t simd_width = 16;  // NEON = 128 bits = 16 bytes

    size_t i = 0;

    // Main loop: unroll x4 (process 64 bytes per iteration)
    for (; i + simd_width * 4 <= packet_size; i += simd_width * 4) {
        // Load from packet 0
        uint8x16_t result0 = vld1q_u8(packets[0] + i);
        uint8x16_t result1 = vld1q_u8(packets[0] + i + 16);
        uint8x16_t result2 = vld1q_u8(packets[0] + i + 32);
        uint8x16_t result3 = vld1q_u8(packets[0] + i + 48);

        // XOR all remaining packets
        for (size_t p = 1; p < num_packets; p++) {
            result0 = veorq_u8(result0, vld1q_u8(packets[p] + i));
            result1 = veorq_u8(result1, vld1q_u8(packets[p] + i + 16));
            result2 = veorq_u8(result2, vld1q_u8(packets[p] + i + 32));
            result3 = veorq_u8(result3, vld1q_u8(packets[p] + i + 48));
        }

        // Store results (use 4-wide store if available)
        vst1q_u8(repair + i, result0);
        vst1q_u8(repair + i + 16, result1);
        vst1q_u8(repair + i + 32, result2);
        vst1q_u8(repair + i + 48, result3);
    }

    // Remaining NEON iterations
    for (; i + simd_width <= packet_size; i += simd_width) {
        uint8x16_t result = vld1q_u8(packets[0] + i);

        for (size_t p = 1; p < num_packets; p++) {
            result = veorq_u8(result, vld1q_u8(packets[p] + i));
        }

        vst1q_u8(repair + i, result);
    }

    // Scalar tail
    for (; i < packet_size; i++) {
        repair[i] = packets[0][i];
        for (size_t p = 1; p < num_packets; p++) {
            repair[i] ^= packets[p][i];
        }
    }
}
#endif  // __aarch64__

// ============================================================================
// Scalar Fallback Implementation
// ============================================================================

void xor_packets_scalar(
    const uint8_t* packets[],
    size_t num_packets,
    size_t packet_size,
    uint8_t* repair
) {
    if (num_packets == 0 || packet_size == 0) {
        return;
    }

    for (size_t i = 0; i < packet_size; i++) {
        repair[i] = packets[0][i];
        for (size_t p = 1; p < num_packets; p++) {
            repair[i] ^= packets[p][i];
        }
    }
}

// ============================================================================
// Function Selection and Runtime Dispatch
// ============================================================================

static xor_impl_fn g_xor_impl = nullptr;

xor_impl_fn fec_select_xor_impl() {
    if (g_xor_impl != nullptr) {
        return g_xor_impl;
    }

#ifdef __x86_64__
    CPUFeatures features = detect_cpu_features();

    if (features.has_avx512f && features.has_avx512bw) {
        // TODO: Check for frequency throttling
        // For now, prefer AVX2 to avoid throttling on some CPUs
        g_xor_impl = xor_packets_avx2;
    } else if (features.has_avx2) {
        g_xor_impl = xor_packets_avx2;
    } else {
        g_xor_impl = xor_packets_scalar;
    }
#elif defined(__aarch64__)
    // ARM64 always has NEON
    g_xor_impl = xor_packets_neon;
#else
    g_xor_impl = xor_packets_scalar;
#endif

    return g_xor_impl;
}

// ============================================================================
// Memory Management
// ============================================================================

void* fec_alloc_slab(size_t size) {
    void* ptr = nullptr;
    size_t aligned_size = (size + 63) & ~63UL;  // Round up to 64-byte boundary

#if defined(_POSIX_C_SOURCE) || defined(__unix__) || defined(__APPLE__)
    if (posix_memalign(&ptr, 64, aligned_size) != 0) {
        return nullptr;
    }
#elif defined(_WIN32)
    ptr = _aligned_malloc(aligned_size, 64);
#else
    ptr = malloc(aligned_size);
#endif

    return ptr;
}

void* fec_alloc_slab_numa(size_t size, int numa_node) {
    // Try NUMA-aware allocation if available
    // Fall back to regular alloc if not supported
#ifdef __linux__
    void* ptr = nullptr;
    size_t aligned_size = (size + 63) & ~63UL;

    if (posix_memalign(&ptr, 64, aligned_size) != 0) {
        return nullptr;
    }

    // Try to bind to NUMA node
    if (numa_node >= 0) {
        struct bitmask* mask = numa_allocate_nodemask();
        numa_bitmask_setbit(mask, numa_node);
        mbind(ptr, aligned_size, MPOL_BIND, mask->maskp, mask->size + 1, MPOL_MF_MOVE);
        numa_free_nodemask(mask);
    }

    return ptr;
#else
    // Fallback: use regular allocation
    return fec_alloc_slab(size);
#endif
}

void* fec_alloc_repair_buffer(size_t size) {
    return fec_alloc_slab(size);
}

void fec_free_repair_buffer(void* ptr) {
    free(ptr);
}

void fec_free_slab(void* ptr) {
#ifdef _WIN32
    _aligned_free(ptr);
#else
    free(ptr);
#endif
}

// ============================================================================
// FEC Encoder Context
// ============================================================================

struct FECEncoderCtx {
    double redundancy;
    uint32_t max_groups;
    xor_impl_fn impl;
};

FECEncoderCtx* fec_encoder_new(double redundancy, uint32_t max_groups) {
    auto ctx = new FECEncoderCtx();
    ctx->redundancy = (redundancy > 0 && redundancy <= 1.0) ? redundancy : 0.10;
    ctx->max_groups = max_groups > 0 ? max_groups : 1024;
    ctx->impl = fec_select_xor_impl();
    return ctx;
}

void fec_encoder_free(FECEncoderCtx* ctx) {
    if (ctx != nullptr) {
        delete ctx;
    }
}

// ============================================================================
// Batch Encoding API
// ============================================================================

int fec_encode_batch(
    FECEncoderCtx* ctx,
    const uint8_t* slab,
    const uint32_t* offsets,
    uint32_t num_groups,
    uint32_t packet_size,
    uint8_t* repair_out
) {
    if (ctx == nullptr || slab == nullptr || offsets == nullptr || repair_out == nullptr) {
        return -1;
    }

    if (num_groups == 0 || packet_size == 0) {
        return 0;
    }

    // For each group, extract packet pointers from slab and encode
    const uint8_t* packets[256];  // Max 256 packets per group (reasonable limit)

    for (uint32_t group_idx = 0; group_idx < num_groups; group_idx++) {
        // Number of packets in this group is inferred from offset spacing
        // For simplicity, assume all groups have same structure
        // In production, you'd encode the packet count in metadata

        uint32_t packets_in_group = 10;  // Default: 10 packets per group

        // Extract packet pointers from slab using offsets
        for (uint32_t p = 0; p < packets_in_group && p < 256; p++) {
            uint32_t offset = offsets[group_idx * packets_in_group + p];
            packets[p] = slab + offset;
        }

        // Encode this group
        uint8_t* repair_ptr = repair_out + (group_idx * packet_size);
        ctx->impl(packets, packets_in_group, packet_size, repair_ptr);
    }

    return 0;
}

