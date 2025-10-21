#!/bin/bash

# Real QUIC Bottom Starter
# Features: Real-time QUIC metrics from Go application

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}🚀 Real QUIC Bottom${NC}"
echo -e "${CYAN}===================${NC}"
echo ""

echo -e "${BLUE}🎯 Real QUIC Bottom Features:${NC}"
echo -e "${GREEN}✅ Real-time QUIC metrics from Go application${NC}"
echo -e "${GREEN}✅ HTTP API for metrics collection${NC}"
echo -e "${GREEN}✅ Professional visualizations${NC}"
echo -e "${GREEN}✅ Network simulation integration${NC}"
echo -e "${GREEN}✅ Security testing integration${NC}"
echo -e "${GREEN}✅ Cloud deployment monitoring${NC}"
echo -e "${GREEN}✅ Interactive controls${NC}"
echo ""

echo -e "${BLUE}🌐 HTTP API Endpoints:${NC}"
echo -e "${YELLOW}  POST /api/metrics - Receive metrics from Go app${NC}"
echo -e "${YELLOW}  GET /health - Health check${NC}"
echo -e "${YELLOW}  GET /api/current - Get current metrics${NC}"
echo ""

echo -e "${BLUE}🎮 Interactive Controls:${NC}"
echo -e "${YELLOW}  q/ESC - Quit${NC}"
echo -e "${YELLOW}  r - Reset all data${NC}"
echo -e "${YELLOW}  h - Show help${NC}"
echo -e "${YELLOW}  1 - Dashboard view${NC}"
echo -e "${YELLOW}  2 - Analytics view${NC}"
echo -e "${YELLOW}  3 - Network simulation view${NC}"
echo -e "${YELLOW}  4 - Security testing view${NC}"
echo -e "${YELLOW}  5 - Cloud deployment view${NC}"
echo -e "${YELLOW}  a - All views${NC}"
echo -e "${YELLOW}  n - Toggle network simulation${NC}"
echo -e "${YELLOW}  +/- - Change network preset${NC}"
echo -e "${YELLOW}  s - Toggle security testing${NC}"
echo -e "${YELLOW}  d - Toggle cloud deployment${NC}"
echo -e "${YELLOW}  i - Scale cloud instances${NC}"
echo ""

echo -e "${CYAN}🚀 Starting Real QUIC Bottom...${NC}"
echo ""

# Navigate to the quic-bottom directory
cd quic-bottom

# Build the real QUIC Bottom binary
echo -e "${BLUE}🔨 Building real QUIC Bottom...${NC}"
cargo build --release --bin quic-bottom-real

# Check if the build was successful
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Build failed. Exiting.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Build completed${NC}"
echo ""

# Show usage examples
echo -e "${BLUE}📋 Usage Examples:${NC}"
echo -e "${YELLOW}  Basic run:${NC}"
echo -e "${CYAN}    ./target/release/quic-bottom-real${NC}"
echo ""
echo -e "${YELLOW}  With debug logging:${NC}"
echo -e "${CYAN}    RUST_LOG=debug ./target/release/quic-bottom-real${NC}"
echo ""

# Run the real QUIC Bottom application
echo -e "${GREEN}🎯 Running Real QUIC Bottom...${NC}"
echo -e "${YELLOW}Press 'q' to quit, 'h' for help${NC}"
echo -e "${YELLOW}HTTP API server will start on port 8080${NC}"
echo ""

./target/release/quic-bottom-real
