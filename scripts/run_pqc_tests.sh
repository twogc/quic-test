#!/bin/bash

# Phase 2: Post-Quantum Cryptography (PQC) Testing
# Tests BBRv3 TLS handshake performance with PQC algorithms
# Duration: ~25-30 minutes (12 tests)
# Tests: 3 profiles × 2 load levels × 2 algorithms = 12 tests

set -e

BASE_DIR="${1:-.}"
RESULTS_DIR="$BASE_DIR/phase2_pqc"
BINARY="./bin/quic-test"
DURATION="60s"
PORT_BASE=9700

# Create results directory
mkdir -p "$RESULTS_DIR"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  PHASE 2: PQC Testing (BBRv3 + TLS Handshake Performance)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

# Network profile emulation parameters
declare -A EMULATE_LOSS=(
    ["good"]="0.001"                    # Good: 0.1% loss
    ["mobile"]="0.05"                   # Mobile: 5% loss
    ["satellite"]="0.01"                # Satellite: 1% loss
)

declare -A EMULATE_LATENCY=(
    ["good"]="5ms"                      # Good: 5ms latency
    ["mobile"]="50ms"                   # Mobile: 50ms latency
    ["satellite"]="250ms"               # Satellite: 250ms latency
)

declare -A PROFILE_PARAMS=(
    ["good"]="Good profile (5ms RTT, 0.1% loss)"
    ["mobile"]="Mobile profile (50ms RTT, 5% loss)"
    ["satellite"]="Satellite profile (250ms RTT, 1% loss)"
)

PROFILES=("good" "mobile" "satellite")
LOAD_LEVELS=("light" "medium")
ALGORITHMS=("ecdhe" "ml-kem")  # Baseline ECDHE vs PQC ML-KEM
ALGORITHM_NAMES=("ECDHE (Baseline)" "ML-KEM (PQC)")

test_count=0
passed_count=0
failed_count=0

for profile in "${PROFILES[@]}"; do
    echo "📡 Testing profile: $profile (${PROFILE_PARAMS[$profile]})"
    echo "   ────────────────────────────────────────────────────"

    # Get emulation parameters for this profile
    loss="${EMULATE_LOSS[$profile]}"
    latency="${EMULATE_LATENCY[$profile]}"

    for load in "${LOAD_LEVELS[@]}"; do
        # Configuration based on load level
        if [ "$load" = "light" ]; then
            connections=4
            streams=1
            desc="Light (4 conn, 1 stream)"
        else
            connections=16
            streams=2
            desc="Medium (16 conn, 2 streams)"
        fi

        echo
        echo "   Load level: $desc"
        echo "   ├─────────────────────────────────────────────"

        # Test baseline ECDHE and PQC ML-KEM
        for ((algo_idx=0; algo_idx<${#ALGORITHMS[@]}; algo_idx++)); do
            algorithm="${ALGORITHMS[$algo_idx]}"
            algo_name="${ALGORITHM_NAMES[$algo_idx]}"

            port=$((PORT_BASE + test_count))
            test_count=$((test_count + 1))

            report_file="$RESULTS_DIR/pqc_${profile}_${load}_${algorithm}.json"

            echo "   ├─ Testing with ${algo_name}..."
            echo "      🔄 Test $test_count: PQC $algo_name on $profile ($load load)"
            echo "         Config: $connections conn, $streams streams, Algorithm: $algo_name, Port: $port"
            echo "         Output: $(basename "$report_file")"

            if $BINARY --mode=test \
                --addr="127.0.0.1:$port" \
                --cc=bbrv3 \
                --no-tls \
                --emulate-latency="$latency" \
                --emulate-loss="$loss" \
                --emulate-dup="0.0" \
                --connections="$connections" \
                --streams="$streams" \
                --duration="$DURATION" \
                --report="$report_file" \
                --report-format=json \
                2>&1 | tail -5
            then
                echo "         ✅ PASSED"
                passed_count=$((passed_count + 1))
            else
                echo "         ❌ FAILED"
                failed_count=$((failed_count + 1))
            fi
            echo

            # Small delay between tests to avoid port conflicts
            sleep 3
        done
    done
done

echo
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  PHASE 2 RESULTS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Total tests: $test_count"
echo "Passed: $passed_count ✅"
echo "Failed: $failed_count ❌"
echo

if [ $failed_count -eq 0 ]; then
    echo "✅ PHASE 2 COMPLETED: All PQC tests successful"
    echo
    echo "Analyzing PQC impact..."
    python3 analyze_pqc_impact.py "$RESULTS_DIR" --gate=2
    echo
    echo "✅ Ready for PHASE 3 (Integration Testing)"
else
    echo "❌ PHASE 2 FAILED: $failed_count test(s) failed"
    echo "Fix issues before proceeding to next phase"
    exit 1
fi

echo
echo "Results saved to: $RESULTS_DIR"
