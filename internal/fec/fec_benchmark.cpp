#include "fec_xor_simd.h"
#include <chrono>
#include <iostream>
#include <vector>
#include <iomanip>
#include <cstring>
#include <random>

// ============================================================================
// Benchmark Utilities
// ============================================================================

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

struct BenchmarkResult {
    const char* name;
    double elapsed_ms;
    double throughput_gbps;
    double us_per_group;
};

BenchmarkResult benchmark_xor_impl(
    const char* name,
    xor_impl_fn impl,
    size_t num_packets,
    size_t packet_size,
    size_t num_iterations
) {
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
    auto elapsed = std::chrono::duration_cast<std::chrono::nanoseconds>(end - start);

    double elapsed_ms = elapsed.count() / 1e6;
    double elapsed_us = elapsed.count() / 1e3;
    double elapsed_sec = elapsed_us / 1e6;

    // Calculate metrics
    // Per iteration: num_packets * packet_size bytes processed
    size_t total_bytes = num_iterations * num_packets * packet_size;
    double throughput_gbps = (total_bytes * 8) / (elapsed_sec * 1e9);
    double us_per_group = elapsed_us / num_iterations;

    return {name, elapsed_ms, throughput_gbps, us_per_group};
}

void print_benchmark_header(size_t num_packets, size_t packet_size) {
    std::cout << "\n=== FEC XOR Benchmark ===" << std::endl;
    std::cout << "Test: " << num_packets << " packets × " << packet_size << " bytes" << std::endl;
    std::cout << "Data per group: " << (num_packets * packet_size / 1024) << " KB" << std::endl;
    std::cout << std::string(70, '-') << std::endl;
    std::cout << std::left
              << std::setw(25) << "Implementation"
              << std::setw(15) << "µs per group"
              << std::setw(15) << "Throughput"
              << std::setw(15) << "Total time" << std::endl;
    std::cout << std::string(70, '-') << std::endl;
}

void print_result(const BenchmarkResult& result) {
    std::cout << std::left
              << std::setw(25) << result.name
              << std::setw(15) << std::fixed << std::setprecision(2) << result.us_per_group
              << std::setw(15) << (std::string(result.name) == "Scalar" ? "" : std::to_string((int)result.throughput_gbps) + " Gbps")
              << std::setw(15) << (std::to_string((int)result.elapsed_ms) + " ms") << std::endl;
}

// ============================================================================
// Main Benchmark Runner
// ============================================================================

int main() {
    std::cout << "\n╔════════════════════════════════════════════╗" << std::endl;
    std::cout << "║  FEC XOR SIMD - Performance Benchmark      ║" << std::endl;
    std::cout << "╚════════════════════════════════════════════╝" << std::endl;

    std::cout << "\nPlatform: " <<
#ifdef __x86_64__
        "x86_64"
#elif defined(__aarch64__)
        "ARM64"
#else
        "Unknown"
#endif
    << std::endl;

    xor_impl_fn selected_impl = fec_select_xor_impl();
    std::cout << "Selected XOR impl: " <<
#ifdef __aarch64__
        (selected_impl == xor_packets_neon ? "NEON" : "Scalar")
#elif defined(__x86_64__)
        (selected_impl == xor_packets_avx2 ? "AVX2" :
         (selected_impl == xor_packets_avx512 ? "AVX-512" : "Scalar"))
#else
        "Scalar"
#endif
    << std::endl;

    // Test case 1: Standard QUIC packet size (1200 bytes, 10 packets)
    {
        print_benchmark_header(10, 1200);

        const size_t num_iterations = 10000;
        const size_t num_packets = 10;
        const size_t packet_size = 1200;

        BenchmarkResult reference = benchmark_xor_impl(
            "Scalar (reference)",
            xor_packets_scalar,
            num_packets,
            packet_size,
            num_iterations
        );
        print_result(reference);

        BenchmarkResult selected = benchmark_xor_impl(
            "Selected",
            selected_impl,
            num_packets,
            packet_size,
            num_iterations
        );
        print_result(selected);

        if (selected.throughput_gbps > 0) {
            double speedup = reference.us_per_group / selected.us_per_group;
            std::cout << "\n  Speedup: " << std::fixed << std::setprecision(2) << speedup << "x" << std::endl;
        }
    }

    // Test case 2: Larger packets
    {
        print_benchmark_header(10, 9000);

        const size_t num_iterations = 1000;
        const size_t num_packets = 10;
        const size_t packet_size = 9000;

        BenchmarkResult reference = benchmark_xor_impl(
            "Scalar (reference)",
            xor_packets_scalar,
            num_packets,
            packet_size,
            num_iterations
        );
        print_result(reference);

        BenchmarkResult selected = benchmark_xor_impl(
            "Selected",
            selected_impl,
            num_packets,
            packet_size,
            num_iterations
        );
        print_result(selected);

        if (selected.throughput_gbps > 0) {
            double speedup = reference.us_per_group / selected.us_per_group;
            std::cout << "\n  Speedup: " << std::fixed << std::setprecision(2) << speedup << "x" << std::endl;
        }
    }

    // Test case 3: Few packets
    {
        print_benchmark_header(5, 1200);

        const size_t num_iterations = 10000;
        const size_t num_packets = 5;
        const size_t packet_size = 1200;

        BenchmarkResult reference = benchmark_xor_impl(
            "Scalar (reference)",
            xor_packets_scalar,
            num_packets,
            packet_size,
            num_iterations
        );
        print_result(reference);

        BenchmarkResult selected = benchmark_xor_impl(
            "Selected",
            selected_impl,
            num_packets,
            packet_size,
            num_iterations
        );
        print_result(selected);

        if (selected.throughput_gbps > 0) {
            double speedup = reference.us_per_group / selected.us_per_group;
            std::cout << "\n  Speedup: " << std::fixed << std::setprecision(2) << speedup << "x" << std::endl;
        }
    }

    std::cout << std::string(70, '=') << std::endl;
    std::cout << "Benchmark complete!" << std::endl;

    return 0;
}

