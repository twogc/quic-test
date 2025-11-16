#!/bin/bash

# Phase 1: FEC Impact Testing
# Tests BBRv3 with FEC at 5%, 10%, 20% redundancy on mobile & satellite profiles
# Duration: ~45-60 minutes (18 tests)

set -e

BASE_DIR="${1:-.}"
RESULTS_DIR="$BASE_DIR/phase1_fec"
BINARY="./bin/quic-test"
DURATION="60s"
PORT_BASE=9600

# Create results directory
mkdir -p "$RESULTS_DIR"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  PHASE 1: FEC Impact Testing (BBRv3 + FEC)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

# Network profile emulation parameters
declare -A EMULATE_LOSS=(
    ["mobile"]="0.05"                   # Mobile: 5% loss
    ["satellite"]="0.01"                # Satellite: 1% loss
)

declare -A EMULATE_LATENCY=(
    ["mobile"]="50ms"                   # Mobile: 50ms latency
    ["satellite"]="250ms"               # Satellite: 250ms latency
)

declare -A PROFILE_PARAMS=(
    ["mobile"]="Mobile profile (50ms RTT, 5% loss)"
    ["satellite"]="Satellite profile (250ms RTT, 1% loss)"
)

# FEC configurations
declare -A FEC_LEVELS=(
    ["none"]="0"
    ["light"]="0.05"
    ["moderate"]="0.10"
    ["heavy"]="0.20"
)

PROFILES=("mobile" "satellite")
LOAD_LEVELS=("light" "medium")
FEC_CONFIGS=("none" "light" "moderate" "heavy")

test_count=0
passed_count=0
failed_count=0

# Helper function to run a single test
run_test() {
    local profile=$1
    local fec_level=$2
    local fec_label=$3
    local load=$4
    local connections=$5
    local streams=$6

    port=$((PORT_BASE + test_count))
    test_count=$((test_count + 1))

    # Determine output file name
    if [ "$fec_label" = "none" ]; then
        report_file="$RESULTS_DIR/fec_${profile}_baseline_${load}.json"
    else
        report_file="$RESULTS_DIR/fec_${profile}_${fec_label}_${load}.json"
    fi

    loss="${EMULATE_LOSS[$profile]}"
    latency="${EMULATE_LATENCY[$profile]}"

    echo "   🔄 Test $test_count: $fec_label FEC on $profile ($load load)"
    echo "      Config: $connections conn, $streams streams, FEC=${fec_level}%, Port=$port"
    echo "      Output: $(basename "$report_file")"

    if $BINARY --mode=test \
        --addr="127.0.0.1:$port" \
        --cc=bbrv3 \
        --no-tls \
        --emulate-latency="$latency" \
        --emulate-loss="$loss" \
        --fec="$fec_level" \
        --connections="$connections" \
        --streams="$streams" \
        --duration="$DURATION" \
        --report="$report_file" \
        --report-format=json \
        2>&1 | tail -5
    then
        echo "      ✅ PASSED"
        passed_count=$((passed_count + 1))
    else
        echo "      ❌ FAILED"
        failed_count=$((failed_count + 1))
    fi
    echo

    # Small delay between tests to avoid port conflicts
    sleep 2
}

# Test execution loop
for profile in "${PROFILES[@]}"; do
    echo "📡 Testing profile: $profile (${PROFILE_PARAMS[$profile]})"
    echo "   ────────────────────────────────────────────────────"
    echo

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

        echo "   Load level: $desc"
        echo "   ├─ Testing baseline (no FEC)..."

        # Run baseline test (FEC = 0%)
        run_test "$profile" "0" "none" "$load" "$connections" "$streams"

        # Run FEC tests at different levels
        for fec_config in "light" "moderate" "heavy"; do
            fec_value="${FEC_LEVELS[$fec_config]}"
            echo "   ├─ Testing with FEC $fec_config (${fec_value}%)..."
            run_test "$profile" "$fec_value" "$fec_config" "$load" "$connections" "$streams"
        done
        echo
    done
done

# Phase 1 Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  PHASE 1 RESULTS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Total tests: $test_count"
echo "Passed: $passed_count ✅"
echo "Failed: $failed_count ❌"
echo

if [ $failed_count -eq 0 ]; then
    echo "✅ PHASE 1 COMPLETED: All tests passed"
    echo
    echo "Analyzing FEC impact..."
    if [ -f "analyze_fec_impact.py" ]; then
        python3 analyze_fec_impact.py "$RESULTS_DIR" --gate=1
    else
        echo "ℹ️  Note: analyze_fec_impact.py not found. Skipping detailed analysis."
        echo "   Create this script to analyze FEC metrics and validate Gate 1 criteria."
    fi
    echo
    echo "✅ Ready for PHASE 2 (PQC testing)"
else
    echo "❌ PHASE 1 FAILED: $failed_count test(s) failed"
    echo "Fix issues before proceeding to next phase"
    exit 1
fi

echo
echo "Results saved to: $RESULTS_DIR"
