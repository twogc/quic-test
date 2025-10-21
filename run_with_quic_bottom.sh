#!/bin/bash

# Run QUIC Test with QUIC Bottom Integration
# Features: Real-time metrics visualization with QUIC Bottom

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}🚀 QUIC Test with QUIC Bottom Integration${NC}"
echo -e "${CYAN}===========================================${NC}"
echo ""

echo -e "${BLUE}🎯 Features:${NC}"
echo -e "${GREEN}✅ Real-time QUIC metrics visualization${NC}"
echo -e "${GREEN}✅ Professional TUI with QUIC Bottom${NC}"
echo -e "${GREEN}✅ HTTP API for metrics collection${NC}"
echo -e "${GREEN}✅ Network simulation integration${NC}"
echo -e "${GREEN}✅ Security testing integration${NC}"
echo -e "${GREEN}✅ Cloud deployment monitoring${NC}"
echo ""

# Build QUIC Bottom first
echo -e "${BLUE}🔨 Building QUIC Bottom...${NC}"
cd quic-bottom
cargo build --release --bin quic-bottom-real

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Failed to build QUIC Bottom. Exiting.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ QUIC Bottom built successfully${NC}"
cd ..

# Build Go application
echo -e "${BLUE}🔨 Building QUIC Test application...${NC}"
go build -o bin/quic-test .

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Failed to build QUIC Test. Exiting.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ QUIC Test built successfully${NC}"
echo ""

# Show usage examples
echo -e "${BLUE}📋 Usage Examples:${NC}"
echo -e "${YELLOW}  Server with QUIC Bottom:${NC}"
echo -e "${CYAN}    ./bin/quic-test --mode=server --quic-bottom${NC}"
echo ""
echo -e "${YELLOW}  Client with QUIC Bottom:${NC}"
echo -e "${CYAN}    ./bin/quic-test --mode=client --addr=localhost:9000 --quic-bottom${NC}"
echo ""
echo -e "${YELLOW}  Test with QUIC Bottom:${NC}"
echo -e "${CYAN}    ./bin/quic-test --mode=test --quic-bottom --duration=30s${NC}"
echo ""

# Parse command line arguments
MODE="test"
ADDR=":9000"
DURATION=""
CONNECTIONS="1"
STREAMS="1"
PACKET_SIZE="1200"
RATE="100"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --mode)
            MODE="$2"
            shift 2
            ;;
        --addr)
            ADDR="$2"
            shift 2
            ;;
        --duration)
            DURATION="$2"
            shift 2
            ;;
        --connections)
            CONNECTIONS="$2"
            shift 2
            ;;
        --streams)
            STREAMS="$2"
            shift 2
            ;;
        --packet-size)
            PACKET_SIZE="$2"
            shift 2
            ;;
        --rate)
            RATE="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --mode MODE           Mode: server, client, test (default: test)"
            echo "  --addr ADDR           Address (default: :9000)"
            echo "  --duration DURATION   Test duration (default: infinite)"
            echo "  --connections N       Number of connections (default: 1)"
            echo "  --streams N           Number of streams (default: 1)"
            echo "  --packet-size SIZE    Packet size in bytes (default: 1200)"
            echo "  --rate RATE            Rate in packets per second (default: 100)"
            echo "  --help               Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Build command
CMD="./bin/quic-test --mode=$MODE --addr=$ADDR --connections=$CONNECTIONS --streams=$STREAMS --packet-size=$PACKET_SIZE --rate=$RATE --quic-bottom"

if [ -n "$DURATION" ]; then
    CMD="$CMD --duration=$DURATION"
fi

echo -e "${GREEN}🎯 Running QUIC Test with QUIC Bottom...${NC}"
echo -e "${YELLOW}Command: $CMD${NC}"
echo ""
echo -e "${YELLOW}QUIC Bottom will start automatically on port 8080${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop both applications${NC}"
echo ""

# Run the command
eval $CMD
