#!/bin/bash

###############################################################################
# BBRv2 vs BBRv3 Test Suite with XML Output and Metrics Summary
#
# Purpose: Run comprehensive tests for BBRv2 and BBRv3 with custom metrics
#          Export results to XML and summarize to metrics file
#
# Usage: bash bbrv2_bbrv3_test_suite.sh [--scenarios good,mobile,satellite,highloss]
#
# Output:
#   - XML reports: test_results/phase0/reports/*.xml
#   - Metrics summary: test_results/phase0/METRICS_SUMMARY.txt
#   - CSV comparison: test_results/phase0/COMPARISON.csv
#
# Author: CloudBridge Labs
# Date: 2025-11-03
###############################################################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# Configuration
QUIC_BIN=\"./bin/quic-test\"
RESULTS_DIR=\"test_results/phase0\"
REPORTS_DIR=\"$RESULTS_DIR/reports\"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
METRICS_FILE=\"$REPORTS_DIR/METRICS_SUMMARY.txt\"
CSV_FILE=\"$REPORTS_DIR/COMPARISON.csv\"
XML_TEMPLATE_DIR=\"/tmp/bbr_xml_$$\"

# Default scenarios
SELECTED_SCENARIOS=\"good mobile satellite highloss\"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --scenarios)
            SELECTED_SCENARIOS=\"$2\"
            shift 2
            ;;
        *)
            echo \"Unknown option: $1\"
            exit 1
            ;;
    esac
done

# Logging functions
log() {
    echo -e \"${BLUE}[$(date '+%H:%M:%S')]${NC} $1\"
}

log_success() {
    echo -e \"${GREEN}✓${NC} $1\"
}

log_error() {
    echo -e \"${RED}✗${NC} $1\"
}

log_section() {
    echo
    echo -e \"${BOLD}${YELLOW}>>> $1${NC}\"
    echo
}

# Create directories
mkdir -p \"$REPORTS_DIR\" \"$XML_TEMPLATE_DIR\"

# Check binary
if [ ! -f \"$QUIC_BIN\" ]; then
    log_error \"Binary not found: $QUIC_BIN\"
    exit 1
fi

log_section \"BBRv2 vs BBRv3 Comprehensive Test Suite\"

# Test scenarios with expected metrics baseline
declare -A SCENARIOS=(
    [good]=\"--emulate-latency=20ms --emulate-loss=0 RTT_20ms_Loss_0pct\"
    [mobile]=\"--emulate-latency=80ms --emulate-loss=0.01 RTT_80ms_Loss_1pct\"
    [satellite]=\"--emulate-latency=200ms --emulate-loss=0.05 RTT_200ms_Loss_5pct\"
    [highloss]=\"--emulate-latency=100ms --emulate-loss=0.10 RTT_100ms_Loss_10pct\"
)

# Initialize tracking
declare -A TEST_RESULTS
declare -a COMPLETED_TESTS
declare -a FAILED_TESTS

# Function to generate XML report
generate_xml_report() {
    local scenario=$1
    local cc=$2
    local params=$3
    local duration=$4
    local json_file=$5

    local xml_file=\"${REPORTS_DIR}/test_${scenario}_${cc}_${TIMESTAMP}.xml\"

    # Extract metrics from JSON (or use defaults)
    local throughput=$(jq -r '.metrics.throughput // \"N/A\"' \"$json_file\" 2>/dev/null || echo \"0\")
    local latency_p95=$(jq -r '.metrics.latency_p95 // \"N/A\"' \"$json_file\" 2>/dev/null || echo \"0\")
    local jitter=$(jq -r '.metrics.jitter // \"N/A\"' \"$json_file\" 2>/dev/null || echo \"0\")
    local loss_rate=$(jq -r '.metrics.loss_rate // \"N/A\"' \"$json_file\" 2>/dev/null || echo \"0\")
    local fairness=$(jq -r '.metrics.fairness // \"N/A\"' \"$json_file\" 2>/dev/null || echo \"0\")
    local recovery_time=$(jq -r '.metrics.recovery_time // \"N/A\"' \"$json_file\" 2>/dev/null || echo \"0\")

    cat > \"$xml_file\" <<EOF
<?xml version=\"1.0\" encoding=\"UTF-8\"?>
<test>
    <metadata>
        <timestamp>$(date -u +%Y-%m-%dT%H:%M:%SZ)</timestamp>
        <scenario>$scenario</scenario>
        <congestion_control>$cc</congestion_control>
        <duration>$duration</duration>
    </metadata>
    <network_conditions>
        <parameters>$params</parameters>
    </network_conditions>
    <metrics>
        <throughput unit=\"bytes/sec\">$throughput</throughput>
        <latency_p95 unit=\"ms\">$latency_p95</latency_p95>
        <jitter unit=\"ms\">$jitter</jitter>
        <loss_rate unit=\"percent\">$loss_rate</loss_rate>
        <fairness_jain unit=\"index\">$fairness</fairness_jain>
        <recovery_time unit=\"ms\">$recovery_time</recovery_time>
    </metrics>
    <source_data>
        <json_file>$json_file</json_file>
    </source_data>
</test>
EOF

    echo \"$xml_file\"
}

# Function to run single test
run_test() {
    local scenario=$1
    local cc=$2
    local params=\"${SCENARIOS[$scenario]%% *}\"
    local duration=\"60s\"

    local test_name=\"test_${scenario}_${cc}\"
    local json_file=\"${RESULTS_DIR}/${test_name}_${TIMESTAMP}.json\"

    log \"Testing: ${YELLOW}${test_name}${NC}\"

    # Run test
    if timeout 90 $QUIC_BIN --mode test --cc=$cc $params \
        --connections=4 \
        --duration=$duration \
        --report=\"$json_file\" \
        --report-format=json > /dev/null 2>&1; then

        log_success \"Completed: $test_name\"

        # Generate XML report
        local xml_file=$(generate_xml_report \"$scenario\" \"$cc\" \"$params\" \"$duration\" \"$json_file\")
        log_success \"Generated XML: $(basename $xml_file)\"

        COMPLETED_TESTS+=(\"$test_name\")
        return 0
    else
        log_error \"Failed: $test_name\"
        FAILED_TESTS+=(\"$test_name\")
        return 1
    fi
}

# Main test execution
log_section \"PHASE 0 TEST EXECUTION\"

total_tests=0
passed_tests=0
failed_tests=0

for scenario in $SELECTED_SCENARIOS; do
    if [ -z \"${SCENARIOS[$scenario]}\" ]; then
        log_error \"Unknown scenario: $scenario\"
        continue
    fi

    log_section \"Scenario: ${YELLOW}$scenario${NC}\"

    # Test BBRv2
    log \"[1/2] BBRv2...\"
    ((total_tests++))
    if run_test \"$scenario\" \"bbrv2\"; then
        ((passed_tests++))
    else
        ((failed_tests++))
    fi
    sleep 2

    # Test BBRv3
    log \"[2/2] BBRv3...\"
    ((total_tests++))
    if run_test \"$scenario\" \"bbrv3\"; then
        ((passed_tests++))
    else
        ((failed_tests++))
    fi
    sleep 2
done

# Generate metrics summary
log_section \"GENERATING METRICS SUMMARY\"

cat > \"$METRICS_FILE\" <<'EOF'
================================================================================
BBRv2 vs BBRv3 COMPREHENSIVE TEST RESULTS
================================================================================

Test Date: $(date)
Total Tests: ${total_tests}
Passed: ${passed_tests}
Failed: ${failed_tests}

================================================================================
METRICS SUMMARY (XML Generated)
================================================================================

Completed Tests:
EOF

for test in \"${COMPLETED_TESTS[@]}\"; do
    echo \"  ✓ $test\" >> \"$METRICS_FILE\"
done

echo >> \"$METRICS_FILE\"
echo \"Failed Tests:\" >> \"$METRICS_FILE\"
for test in \"${FAILED_TESTS[@]}\"; do
    echo \"  ✗ $test\" >> \"$METRICS_FILE\"
done

echo >> \"$METRICS_FILE\"
echo \"================================================================================\" >> \"$METRICS_FILE\"
echo \"EXPECTED IMPROVEMENTS (BBRv3 vs BBRv2)\" >> \"$METRICS_FILE\"
echo \"================================================================================\" >> \"$METRICS_FILE\"
echo \"\" >> \"$METRICS_FILE\"
echo \"Metric                     Target        Status\" >> \"$METRICS_FILE\"
echo \"--------                   ------        ------\" >> \"$METRICS_FILE\"
echo \"Throughput (RTT > 80ms)    +10-12%       Pending Analysis\" >> \"$METRICS_FILE\"
echo \"Jitter                     -40-50%       Pending Analysis\" >> \"$METRICS_FILE\"
echo \"Packet Loss Recovery       -60%          Pending Analysis\" >> \"$METRICS_FILE\"
echo \"Fairness (Jain)            ≥0.9          Pending Analysis\" >> \"$METRICS_FILE\"
echo \"Recovery Time              -60%          Pending Analysis\" >> \"$METRICS_FILE\"
echo \"\" >> \"$METRICS_FILE\"
echo \"================================================================================\" >> \"$METRICS_FILE\"
echo \"GENERATED REPORTS\" >> \"$METRICS_FILE\"
echo \"================================================================================\" >> \"$METRICS_FILE\"
echo \"\" >> \"$METRICS_FILE\"
echo \"XML Reports:\" >> \"$METRICS_FILE\"
ls -1 \"$REPORTS_DIR\"/*.xml 2>/dev/null | sed 's|.*|  - &|' >> \"$METRICS_FILE\"
echo \"\" >> \"$METRICS_FILE\"
echo \"JSON Raw Data:\" >> \"$METRICS_FILE\"
ls -1 \"$RESULTS_DIR\"/test_*.json 2>/dev/null | sed 's|.*|  - &|' >> \"$METRICS_FILE\"

# Generate CSV comparison
log_section \"GENERATING CSV COMPARISON\"

cat > \"$CSV_FILE\" << 'EOF'
Scenario,CC,Throughput_bytes_s,Latency_P95_ms,Jitter_ms,Loss_Rate_pct,Fairness_Jain,Recovery_Time_ms,Test_Status
EOF

for json in test_results/phase0/test_*.json; do
    if [ -f \"$json\" ]; then
        scenario=$(echo \"$(basename $json)\" | cut -d'_' -f2)
        cc=$(echo \"$(basename $json)\" | cut -d'_' -f3)

        throughput=$(jq -r '.metrics.throughput // \"N/A\"' \"$json\")
        latency=$(jq -r '.metrics.latency_p95 // \"N/A\"' \"$json\")
        jitter=$(jq -r '.metrics.jitter // \"N/A\"' \"$json\")
        loss=$(jq -r '.metrics.loss_rate // \"N/A\"' \"$json\")
        fairness=$(jq -r '.metrics.fairness // \"N/A\"' \"$json\")
        recovery=$(jq -r '.metrics.recovery_time // \"N/A\"' \"$json\")

        echo \"$scenario,$cc,$throughput,$latency,$jitter,$loss,$fairness,$recovery,OK\" >> \"$CSV_FILE\"
    fi
done

log_success \"CSV comparison saved: $CSV_FILE\"

# Summary
log_section \"TEST EXECUTION SUMMARY\"

log \"Total Tests: $total_tests\"
log_success \"Passed: $passed_tests\"
if [ $failed_tests -gt 0 ]; then
    log_error \"Failed: $failed_tests\"
else
    log_success \"Failed: 0\"
fi

log_section \"GENERATED FILES\"

echo \"Metrics Summary:\"
echo \"  → $METRICS_FILE\"
echo

echo \"CSV Comparison:\"
echo \"  → $CSV_FILE\"
echo

echo \"XML Reports:\"
ls -1 \"$REPORTS_DIR\"/*.xml 2>/dev/null | while read f; do
    echo \"  → $(basename $f)\"
done
echo

log_section \"NEXT STEPS\"

echo \"1. View metrics summary:\"
echo \"   cat $METRICS_FILE\"
echo

echo \"2. View CSV comparison:\"
echo \"   cat $CSV_FILE\"
echo

echo \"3. Analyze XML reports:\"
echo \"   ls -la $REPORTS_DIR/*.xml\"
echo

echo \"4. Extract and compare metrics:\"
echo \"   python3 scripts/analyze_phase0_results.py test_results/phase0/\"
echo

# Cleanup
rm -rf \"$XML_TEMPLATE_DIR\"

if [ $failed_tests -eq 0 ]; then
    log_success \"✅ Phase 0 testing completed successfully!\"
    exit 0
else
    log_error \"❌ Phase 0 testing completed with errors\"
    exit 1
fi
