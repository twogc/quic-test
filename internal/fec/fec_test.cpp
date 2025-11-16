#include "fec_xor_simd.h"
#include <cassert>
#include <cstring>
#include <iostream>
#include <vector>
#include <random>
#include <chrono>

// ============================================================================
// Test Utilities
// ============================================================================

void print_test_result(const char* test_name, bool passed) {
    if (passed) {
        std::cout << "✓ " << test_name << std::endl;
    } else {
        std::cout << "✗ " << test_name << std::endl;
    }
}

std::vector<uint8_t> generate_random_data(size_t size) {
    std::vector<uint8_t> data(size);
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<> dis(0, 255);

    for (size_t i = 0; i < size; i++) {
        data[i] = dis(gen);
    }

    return data;
}

// Reference scalar implementation for validation
void xor_packets_reference(
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
// Unit Tests
// ============================================================================

bool test_single_xor() {
    // Test XOR of two simple packets
    // 0x01 ^ 0x05 = 0x04
    // 0x02 ^ 0x06 = 0x04
    // 0x03 ^ 0x07 = 0x04
    // 0x04 ^ 0x08 = 0x0C
    const uint8_t pkt1[] = {0x01, 0x02, 0x03, 0x04};
    const uint8_t pkt2[] = {0x05, 0x06, 0x07, 0x08};

    uint8_t expected[] = {0x04, 0x04, 0x04, 0x0C};  // Corrected expected values
    uint8_t result[4];

    const uint8_t* packets[] = {pkt1, pkt2};
    xor_packets_reference(packets, 2, 4, result);

    bool passed = std::memcmp(result, expected, 4) == 0;
    print_test_result("test_single_xor", passed);
    return passed;
}

bool test_multiple_xor() {
    // Test XOR of multiple packets
    const uint8_t pkt1[] = {0xFF, 0xFF};
    const uint8_t pkt2[] = {0xFF, 0xFF};
    const uint8_t pkt3[] = {0xFF, 0xFF};

    uint8_t expected[] = {0xFF, 0xFF};  // 0xFF ^ 0xFF ^ 0xFF = 0xFF
    uint8_t result[2];

    const uint8_t* packets[] = {pkt1, pkt2, pkt3};
    xor_packets_reference(packets, 3, 2, result);

    bool passed = std::memcmp(result, expected, 2) == 0;
    print_test_result("test_multiple_xor", passed);
    return passed;
}

bool test_zero_packets() {
    uint8_t result[10] = {0};
    const uint8_t* packets[] = {};

    // Should not crash
    xor_packets_reference(packets, 0, 10, result);

    bool passed = true;
    print_test_result("test_zero_packets", passed);
    return passed;
}

bool test_bit_exact_vs_reference() {
    // Generate random test data
    const size_t packet_size = 1200;
    const size_t num_packets = 10;

    std::vector<std::vector<uint8_t>> packets;
    const uint8_t* packet_ptrs[num_packets];

    for (size_t i = 0; i < num_packets; i++) {
        packets.push_back(generate_random_data(packet_size));
        packet_ptrs[i] = packets[i].data();
    }

    // Reference implementation
    std::vector<uint8_t> reference(packet_size);
    xor_packets_reference(packet_ptrs, num_packets, packet_size, reference.data());

    // Test implementation (use feature detection)
    std::vector<uint8_t> test_result(packet_size);
    xor_impl_fn impl = fec_select_xor_impl();
    impl(packet_ptrs, num_packets, packet_size, test_result.data());

    // Compare
    bool passed = std::memcmp(reference.data(), test_result.data(), packet_size) == 0;
    print_test_result("test_bit_exact_vs_reference", passed);

    return passed;
}

bool test_non_multiple_of_simd() {
    // Test with packet sizes not multiple of SIMD width
    const size_t sizes[] = {1200, 1201, 1234, 1500, 1501};

    for (size_t size : sizes) {
        std::vector<std::vector<uint8_t>> packets;
        const uint8_t* packet_ptrs[10];

        for (int i = 0; i < 10; i++) {
            packets.push_back(generate_random_data(size));
            packet_ptrs[i] = packets[i].data();
        }

        std::vector<uint8_t> reference(size);
        xor_packets_reference(packet_ptrs, 10, size, reference.data());

        std::vector<uint8_t> test_result(size);
        xor_impl_fn impl = fec_select_xor_impl();
        impl(packet_ptrs, 10, size, test_result.data());

        if (std::memcmp(reference.data(), test_result.data(), size) != 0) {
            std::cout << "  Mismatch at size " << size << std::endl;
            print_test_result("test_non_multiple_of_simd", false);
            return false;
        }
    }

    print_test_result("test_non_multiple_of_simd", true);
    return true;
}

bool test_large_packet() {
    // Test with large packets
    const size_t packet_size = 9000;
    const size_t num_packets = 5;

    std::vector<std::vector<uint8_t>> packets;
    const uint8_t* packet_ptrs[num_packets];

    for (size_t i = 0; i < num_packets; i++) {
        packets.push_back(generate_random_data(packet_size));
        packet_ptrs[i] = packets[i].data();
    }

    std::vector<uint8_t> reference(packet_size);
    xor_packets_reference(packet_ptrs, num_packets, packet_size, reference.data());

    std::vector<uint8_t> test_result(packet_size);
    xor_impl_fn impl = fec_select_xor_impl();
    impl(packet_ptrs, num_packets, packet_size, test_result.data());

    bool passed = std::memcmp(reference.data(), test_result.data(), packet_size) == 0;
    print_test_result("test_large_packet", passed);
    return passed;
}

bool test_memory_allocation() {
    // Test memory allocation functions
    size_t size = 1024;

    void* slab = fec_alloc_slab(size);
    if (slab == nullptr) {
        print_test_result("test_memory_allocation", false);
        return false;
    }

    // Check alignment (should be 64-byte aligned)
    uintptr_t addr = reinterpret_cast<uintptr_t>(slab);
    bool aligned = (addr % 64) == 0;

    fec_free_slab(slab);

    print_test_result("test_memory_allocation", aligned);
    return aligned;
}

bool test_repair_buffer_allocation() {
    size_t size = 1200;

    void* repair = fec_alloc_repair_buffer(size);
    if (repair == nullptr) {
        print_test_result("test_repair_buffer_allocation", false);
        return false;
    }

    uintptr_t addr = reinterpret_cast<uintptr_t>(repair);
    bool aligned = (addr % 64) == 0;

    fec_free_repair_buffer(repair);

    print_test_result("test_repair_buffer_allocation", aligned);
    return aligned;
}

bool test_encoder_context() {
    FECEncoderCtx* ctx = fec_encoder_new(0.10, 1024);

    bool passed = (ctx != nullptr);

    if (ctx != nullptr) {
        fec_encoder_free(ctx);
    }

    print_test_result("test_encoder_context", passed);
    return passed;
}

// ============================================================================
// Performance Tests
// ============================================================================

struct BenchmarkResult {
    double elapsed_ms;
    double throughput_gbps;
    long iterations;
};

BenchmarkResult benchmark_xor(
    const char* impl_name,
    xor_impl_fn impl,
    size_t num_iterations
) {
    const size_t packet_size = 1200;
    const size_t num_packets = 10;

    // Prepare test data
    std::vector<std::vector<uint8_t>> packets;
    const uint8_t* packet_ptrs[num_packets];

    for (size_t i = 0; i < num_packets; i++) {
        packets.push_back(generate_random_data(packet_size));
        packet_ptrs[i] = packets[i].data();
    }

    std::vector<uint8_t> repair(packet_size);

    // Warm up
    for (size_t i = 0; i < 10; i++) {
        impl(packet_ptrs, num_packets, packet_size, repair.data());
    }

    // Benchmark
    auto start = std::chrono::high_resolution_clock::now();

    for (size_t i = 0; i < num_iterations; i++) {
        impl(packet_ptrs, num_packets, packet_size, repair.data());
    }

    auto end = std::chrono::high_resolution_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::milliseconds>(end - start);

    // Calculate throughput
    // Per iteration: 10 packets * 1200 bytes = 12000 bytes
    size_t total_bytes = num_iterations * num_packets * packet_size;
    double elapsed_ms = elapsed.count();
    double elapsed_sec = elapsed_ms / 1000.0;
    double throughput_gbps = (total_bytes * 8) / (elapsed_sec * 1e9);

    std::cout << "  " << impl_name << ": "
              << elapsed_ms << "ms, "
              << throughput_gbps << " Gbps" << std::endl;

    return {elapsed_ms, throughput_gbps, (long)num_iterations};
}

void run_benchmarks() {
    std::cout << "\n=== Performance Benchmarks ===" << std::endl;

    xor_impl_fn impl = fec_select_xor_impl();

    const size_t num_iterations = 10000;

    std::cout << "XOR Encoding (10 packets × 1200 bytes, " << num_iterations << " iterations):" << std::endl;
    benchmark_xor("Selected implementation", impl, num_iterations);

    std::cout << "\nReference scalar:" << std::endl;
    benchmark_xor("Scalar reference", xor_packets_scalar, num_iterations);
}

// ============================================================================
// Main Test Runner
// ============================================================================

int main() {
    std::cout << "=== FEC XOR SIMD Tests ===" << std::endl;
    std::cout << "Platform: " <<
#ifdef __x86_64__
        "x86_64"
#elif defined(__aarch64__)
        "ARM64"
#else
        "unknown"
#endif
    << std::endl;

    std::cout << "Selected XOR implementation: " <<
#ifdef __x86_64__
        (fec_select_xor_impl() == xor_packets_avx2 ? "AVX2" : "scalar")
#elif defined(__aarch64__)
        (fec_select_xor_impl() == xor_packets_neon ? "NEON" : "scalar")
#else
        "scalar"
#endif
    << std::endl << std::endl;

    bool all_passed = true;

    std::cout << "Running unit tests..." << std::endl;
    all_passed &= test_single_xor();
    all_passed &= test_multiple_xor();
    all_passed &= test_zero_packets();
    all_passed &= test_bit_exact_vs_reference();
    all_passed &= test_non_multiple_of_simd();
    all_passed &= test_large_packet();
    all_passed &= test_memory_allocation();
    all_passed &= test_repair_buffer_allocation();
    all_passed &= test_encoder_context();

    std::cout << "\n";
    if (all_passed) {
        std::cout << "All tests PASSED ✓" << std::endl;
    } else {
        std::cout << "Some tests FAILED ✗" << std::endl;
        return 1;
    }

    run_benchmarks();

    return 0;
}

