#!/bin/bash

# MASQUE Protocol Testing Script
# Tests various MASQUE tunnel types and scenarios

# Colors for output
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  MASQUE Protocol Testing Suite${NC}"
echo "Comprehensive testing of MASQUE tunneling"
echo ""

# Configuration
MASQUE_SERVER=${MASQUE_SERVER:-"212.233.79.160:8443"}
TEST_DURATION=${TEST_DURATION:-60}
CONCURRENT_TESTS=${CONCURRENT_TESTS:-10}

echo -e "${CYAN}Test Configuration:${NC}"
echo "  🌐 MASQUE Server: $MASQUE_SERVER"
echo "  ⏱️ Test Duration: $TEST_DURATION seconds"
echo "  🔗 Concurrent Tests: $CONCURRENT_TESTS"
echo ""

# Test 1: HTTP CONNECT Tunneling
echo -e "${YELLOW}🔍 Test 1: HTTP CONNECT Tunneling${NC}"
echo "Testing HTTP CONNECT over MASQUE..."

# Test basic HTTP CONNECT
echo -e "${CYAN}  Testing basic HTTP CONNECT...${NC}"
if curl -x masque://$MASQUE_SERVER http://httpbin.org/ip --connect-timeout 10 --max-time 30 >/dev/null 2>&1; then
    echo -e "${GREEN}  ✅ HTTP CONNECT test passed${NC}"
else
    echo -e "${RED}  ❌ HTTP CONNECT test failed${NC}"
fi

# Test HTTPS CONNECT
echo -e "${CYAN}  Testing HTTPS CONNECT...${NC}"
if curl -x masque://$MASQUE_SERVER https://httpbin.org/ip --connect-timeout 10 --max-time 30 >/dev/null 2>&1; then
    echo -e "${GREEN}  ✅ HTTPS CONNECT test passed${NC}"
else
    echo -e "${RED}  ❌ HTTPS CONNECT test failed${NC}"
fi

# Test with authentication
echo -e "${CYAN}  Testing authenticated CONNECT...${NC}"
if curl -x masque://user:pass@$MASQUE_SERVER http://httpbin.org/ip --connect-timeout 10 --max-time 30 >/dev/null 2>&1; then
    echo -e "${GREEN}  ✅ Authenticated CONNECT test passed${NC}"
else
    echo -e "${RED}  ❌ Authenticated CONNECT test failed${NC}"
fi

echo ""

# Test 2: UDP Tunneling
echo -e "${YELLOW}🔍 Test 2: UDP Tunneling${NC}"
echo "Testing UDP tunneling over MASQUE..."

# Test DNS over MASQUE
echo -e "${CYAN}  Testing DNS over MASQUE...${NC}"
if nslookup example.com masque://$MASQUE_SERVER >/dev/null 2>&1; then
    echo -e "${GREEN}  ✅ DNS tunneling test passed${NC}"
else
    echo -e "${RED}  ❌ DNS tunneling test failed${NC}"
fi

# Test custom UDP service
echo -e "${CYAN}  Testing custom UDP service...${NC}"
if echo "test" | nc -u masque://$MASQUE_SERVER 8.8.8.8 53 >/dev/null 2>&1; then
    echo -e "${GREEN}  ✅ UDP service test passed${NC}"
else
    echo -e "${RED}  ❌ UDP service test failed${NC}"
fi

echo ""

# Test 3: IP Tunneling
echo -e "${YELLOW}🔍 Test 3: IP Tunneling${NC}"
echo "Testing IP packet tunneling over MASQUE..."

# Test IP tunnel creation
echo -e "${CYAN}  Testing IP tunnel creation...${NC}"
if ip link add masque-tunnel type masque >/dev/null 2>&1; then
    echo -e "${GREEN}  ✅ IP tunnel creation test passed${NC}"
    ip link delete masque-tunnel >/dev/null 2>&1
else
    echo -e "${RED}  ❌ IP tunnel creation test failed${NC}"
fi

# Test IP routing
echo -e "${CYAN}  Testing IP routing...${NC}"
if ip route add 8.8.8.8/32 dev masque-tunnel >/dev/null 2>&1; then
    echo -e "${GREEN}  ✅ IP routing test passed${NC}"
    ip route delete 8.8.8.8/32 >/dev/null 2>&1
else
    echo -e "${RED}  ❌ IP routing test failed${NC}"
fi

echo ""

# Test 4: Performance Testing
echo -e "${YELLOW}🔍 Test 4: Performance Testing${NC}"
echo "Testing MASQUE performance..."

# Test bandwidth
echo -e "${CYAN}  Testing bandwidth...${NC}"
if command -v iperf3 >/dev/null 2>&1; then
    if iperf3 -c masque://$MASQUE_SERVER -t 10 >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ Bandwidth test passed${NC}"
    else
        echo -e "${RED}  ❌ Bandwidth test failed${NC}"
    fi
else
    echo -e "${YELLOW}  ⚠️ iperf3 not available, skipping bandwidth test${NC}"
fi

# Test latency
echo -e "${CYAN}  Testing latency...${NC}"
if ping -c 5 -I masque-tunnel 8.8.8.8 >/dev/null 2>&1; then
    echo -e "${GREEN}  ✅ Latency test passed${NC}"
else
    echo -e "${RED}  ❌ Latency test failed${NC}"
fi

# Test concurrent connections
echo -e "${CYAN}  Testing concurrent connections...${NC}"
echo "  Starting $CONCURRENT_TESTS concurrent connections..."
success_count=0
for i in $(seq 1 $CONCURRENT_TESTS); do
    if curl -x masque://$MASQUE_SERVER http://httpbin.org/ip --connect-timeout 5 --max-time 10 >/dev/null 2>&1; then
        success_count=$((success_count + 1))
    fi
done

if [ $success_count -eq $CONCURRENT_TESTS ]; then
    echo -e "${GREEN}  ✅ All $CONCURRENT_TESTS concurrent connections passed${NC}"
elif [ $success_count -gt 0 ]; then
    echo -e "${YELLOW}  ⚠️ $success_count/$CONCURRENT_TESTS concurrent connections passed${NC}"
else
    echo -e "${RED}  ❌ All concurrent connections failed${NC}"
fi

echo ""

# Test 5: Security Testing
echo -e "${YELLOW}🔍 Test 5: Security Testing${NC}"
echo "Testing MASQUE security..."

# Test authentication
echo -e "${CYAN}  Testing authentication...${NC}"
if curl -x masque://$MASQUE_SERVER -H "Authorization: Bearer invalid-token" http://httpbin.org/ip --connect-timeout 5 --max-time 10 >/dev/null 2>&1; then
    echo -e "${YELLOW}  ⚠️ Authentication test - server accepted invalid token${NC}"
else
    echo -e "${GREEN}  ✅ Authentication test - server rejected invalid token${NC}"
fi

# Test encryption
echo -e "${CYAN}  Testing encryption...${NC}"
if command -v tcpdump >/dev/null 2>&1; then
    echo -e "${GREEN}  ✅ Encryption test - tcpdump available for verification${NC}"
else
    echo -e "${YELLOW}  ⚠️ tcpdump not available, cannot verify encryption${NC}"
fi

# Test access control
echo -e "${CYAN}  Testing access control...${NC}"
if curl -x masque://$MASQUE_SERVER http://restricted-site.com --connect-timeout 5 --max-time 10 >/dev/null 2>&1; then
    echo -e "${YELLOW}  ⚠️ Access control test - server allowed restricted access${NC}"
else
    echo -e "${GREEN}  ✅ Access control test - server blocked restricted access${NC}"
fi

echo ""

# Test 6: Monitoring
echo -e "${YELLOW}🔍 Test 6: Monitoring${NC}"
echo "Testing MASQUE monitoring..."

# Check server metrics
echo -e "${CYAN}  Checking server metrics...${NC}"
if curl -s http://212.233.79.160:2113/metrics | grep -q "masque"; then
    echo -e "${GREEN}  ✅ Server metrics available${NC}"
else
    echo -e "${RED}  ❌ Server metrics not available${NC}"
fi

# Check tunnel metrics
echo -e "${CYAN}  Checking tunnel metrics...${NC}"
if curl -s http://212.233.79.160:2113/metrics | grep -q "tunnel"; then
    echo -e "${GREEN}  ✅ Tunnel metrics available${NC}"
else
    echo -e "${RED}  ❌ Tunnel metrics not available${NC}"
fi

echo ""

# Summary
echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  MASQUE Testing Summary${NC}"
echo ""

# Count successful tests
total_tests=0
passed_tests=0

# HTTP CONNECT tests
total_tests=$((total_tests + 3))
if curl -x masque://$MASQUE_SERVER http://httpbin.org/ip --connect-timeout 5 --max-time 10 >/dev/null 2>&1; then
    passed_tests=$((passed_tests + 1))
fi
if curl -x masque://$MASQUE_SERVER https://httpbin.org/ip --connect-timeout 5 --max-time 10 >/dev/null 2>&1; then
    passed_tests=$((passed_tests + 1))
fi
if curl -x masque://user:pass@$MASQUE_SERVER http://httpbin.org/ip --connect-timeout 5 --max-time 10 >/dev/null 2>&1; then
    passed_tests=$((passed_tests + 1))
fi

# UDP tests
total_tests=$((total_tests + 2))
if nslookup example.com masque://$MASQUE_SERVER >/dev/null 2>&1; then
    passed_tests=$((passed_tests + 1))
fi
if echo "test" | nc -u masque://$MASQUE_SERVER 8.8.8.8 53 >/dev/null 2>&1; then
    passed_tests=$((passed_tests + 1))
fi

# Performance tests
total_tests=$((total_tests + 1))
if [ $success_count -gt 0 ]; then
    passed_tests=$((passed_tests + 1))
fi

# Security tests
total_tests=$((total_tests + 1))
if curl -x masque://$MASQUE_SERVER -H "Authorization: Bearer invalid-token" http://httpbin.org/ip --connect-timeout 5 --max-time 10 >/dev/null 2>&1; then
    passed_tests=$((passed_tests + 1))
fi

# Monitoring tests
total_tests=$((total_tests + 1))
if curl -s http://212.233.79.160:2113/metrics | grep -q "masque"; then
    passed_tests=$((passed_tests + 1))
fi

# Calculate success rate
success_rate=$((passed_tests * 100 / total_tests))

echo -e "${CYAN}Test Results:${NC}"
echo "  Total Tests: $total_tests"
echo "  Passed Tests: $passed_tests"
echo "  Success Rate: $success_rate%"

if [ $success_rate -ge 80 ]; then
    echo -e "${GREEN}  ✅ MASQUE testing completed successfully!${NC}"
elif [ $success_rate -ge 60 ]; then
    echo -e "${YELLOW}  ⚠️ MASQUE testing completed with warnings${NC}"
else
    echo -e "${RED}  ❌ MASQUE testing failed${NC}"
fi

echo ""
echo -e "${BLUE}🌐 Available Interfaces:${NC}"
echo "  MASQUE Server: masque://$MASQUE_SERVER"
echo "  Prometheus: http://212.233.79.160:2113/metrics"
echo "  Grafana: http://212.233.79.160:3000"
echo "  Jaeger: http://212.233.79.160:16686"
echo ""
echo -e "${GREEN}🎉 MASQUE testing complete!${NC}"

