#!/bin/bash

# Start QUIC Bottom TUI
# =====================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Starting QUIC Bottom TUI${NC}"

# Check if Rust is installed
if ! command -v cargo &> /dev/null; then
    echo -e "${RED}❌ Rust/Cargo not found. Please install Rust first.${NC}"
    echo "Visit: https://rustup.rs/"
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "quic-bottom/Cargo.toml" ]; then
    echo -e "${RED}❌ QUIC Bottom project not found. Please run from project root.${NC}"
    exit 1
fi

# Build QUIC Bottom
echo -e "${YELLOW}🔨 Building QUIC Bottom...${NC}"
cd quic-bottom

# Check if dependencies are available
if [ ! -d "temp/bottom" ]; then
    echo -e "${YELLOW}⚠️  Bottom source not found. Copying from temp/bottom...${NC}"
    if [ -d "../temp/bottom" ]; then
        cp -r ../temp/bottom temp/
    else
        echo -e "${RED}❌ Bottom source not found in temp/bottom${NC}"
        exit 1
    fi
fi

# Build the project
echo -e "${YELLOW}📦 Building QUIC Bottom...${NC}"
cargo build --release

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Build successful${NC}"

# Check if config exists
if [ ! -f "config.toml" ]; then
    echo -e "${YELLOW}⚠️  Config file not found. Creating default config...${NC}"
    # Create default config if it doesn't exist
fi

# Start QUIC Bottom
echo -e "${GREEN}🎯 Starting QUIC Bottom TUI...${NC}"
echo -e "${BLUE}📊 QUIC Bottom will be available on:${NC}"
echo -e "   - TUI Interface: Terminal"
echo -e "   - HTTP API: http://localhost:8080"
echo -e "   - Health Check: http://localhost:8080/health"
echo -e "   - Metrics API: http://localhost:8080/metrics"
echo ""
echo -e "${YELLOW}💡 Tips:${NC}"
echo -e "   - Press 'q' to quit"
echo -e "   - Press 'r' to refresh"
echo -e "   - Press 'h' for help"
echo ""

# Run QUIC Bottom
./target/release/quic-bottom --api-port 8080 --interval 100

echo -e "${GREEN}✅ QUIC Bottom stopped${NC}"
