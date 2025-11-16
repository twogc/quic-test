#ifndef FEC_XOR_SIMD_H
#define FEC_XOR_SIMD_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

/**
 * FEC Encoder Context (opaque handle from Go perspective)
 */
typedef struct FECEncoderCtx FECEncoderCtx;

/**
 * Create a new FEC encoder context
 * @param redundancy Redundancy ratio (0.0-1.0)
 * @param max_groups Maximum number of groups to track
 * @return Opaque FEC encoder context
 */
FECEncoderCtx* fec_encoder_new(double redundancy, uint32_t max_groups);

/**
 * Allocate aligned slab for packet data
 * @param size Total size to allocate
 * @return Pointer to 64-byte aligned memory, or NULL on failure
 */
void* fec_alloc_slab(size_t size);

/**
 * Allocate NUMA-aware slab with hugepages (2MB) if available
 * @param size Total size to allocate
 * @param numa_node NUMA node ID (or -1 for default)
 * @return Pointer to 64-byte aligned memory, or NULL on failure
 */
void* fec_alloc_slab_numa(size_t size, int numa_node);

/**
 * Allocate repair buffer for XOR output (stream-store aligned)
 * @param size Size to allocate
 * @return Pointer to 64-byte aligned memory for stream stores
 */
void* fec_alloc_repair_buffer(size_t size);

/**
 * Free repair buffer
 * @param ptr Pointer returned by fec_alloc_repair_buffer
 */
void fec_free_repair_buffer(void* ptr);

/**
 * Batch encode N groups of packets
 *
 * High-level overview:
 * - Takes flat slab of all packet data + offsets table
 * - Computes XOR repair packets for each group
 * - Uses AVX2/AVX-512/NEON/Scalar based on CPU capabilities
 *
 * @param ctx FEC encoder context
 * @param slab Flat buffer containing all packet data (C-managed memory)
 * @param offsets Array of offsets marking packet boundaries in slab
 * @param num_groups Number of groups to encode
 * @param packet_size Size of each packet in bytes
 * @param repair_out Output buffer for repair packets (num_groups * packet_size bytes)
 * @return 0 on success, -1 on error
 */
int fec_encode_batch(
    FECEncoderCtx* ctx,
    const uint8_t* slab,
    const uint32_t* offsets,
    uint32_t num_groups,
    uint32_t packet_size,
    uint8_t* repair_out
);

/**
 * Free encoder context and resources
 * @param ctx FEC encoder context
 */
void fec_encoder_free(FECEncoderCtx* ctx);

/**
 * Free slab memory
 * @param ptr Pointer returned by fec_alloc_slab or fec_alloc_slab_numa
 */
void fec_free_slab(void* ptr);

/**
 * Runtime feature detection and dispatch
 * Internal use only - called during encoder initialization
 */
typedef void (*xor_impl_fn)(
    const uint8_t* packets[],
    size_t num_packets,
    size_t packet_size,
    uint8_t* repair
);

/**
 * Select best XOR implementation based on CPU features
 * @return Function pointer to appropriate XOR implementation
 */
xor_impl_fn fec_select_xor_impl(void);

// Internal implementations (exported for testing only)
void xor_packets_scalar(
    const uint8_t* packets[],
    size_t num_packets,
    size_t packet_size,
    uint8_t* repair
);

#ifdef __x86_64__
void xor_packets_avx2(
    const uint8_t* packets[],
    size_t num_packets,
    size_t packet_size,
    uint8_t* repair
);

void xor_packets_avx512(
    const uint8_t* packets[],
    size_t num_packets,
    size_t packet_size,
    uint8_t* repair
);
#endif

#ifdef __aarch64__
void xor_packets_neon(
    const uint8_t* packets[],
    size_t num_packets,
    size_t packet_size,
    uint8_t* repair
);
#endif

#ifdef __cplusplus
}
#endif

#endif  // FEC_XOR_SIMD_H
