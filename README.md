# 2GC Network Protocol Suite

A comprehensive platform for testing and analyzing network protocols: QUIC, MASQUE, ICE/STUN/TURN and others with **real-time professional visualizations**.

## New: Real-time QUIC Bottom Integration

**Professional TUI visualizations** based on the popular `bottom` system monitor, providing real-time QUIC metrics with advanced analytics, network simulation, security testing, and cloud deployment monitoring.

### Key Features

- **Real-time Metrics Visualization** - Professional TUI with live QUIC performance data
- **Advanced Analytics** - Heatmaps, correlation analysis, anomaly detection
- **Network Simulation** - Real Linux tc integration with preset profiles
- **Security Testing** - Comprehensive QUIC security analysis and attack simulation
- **Cloud Integration** - Multi-cloud deployment with auto-scaling
- **Interactive Controls** - Real-time parameter adjustment and view switching

## Features

- **QUIC Protocol Testing** - Advanced QUIC implementation with experimental features
- **MASQUE Protocol Support** - Tunneling and proxying capabilities  
- **ICE/STUN/TURN Testing** - NAT traversal and P2P connection testing
- **TLS 1.3 Security** - Modern cryptography for secure connections
- **HTTP/3 Support** - HTTP over QUIC implementation
- **Experimental Features** - BBRv2, ACK-Frequency, FEC, Bit Greasing
- **Real-time Monitoring** - Prometheus metrics and Grafana dashboards
- **Comprehensive Testing** - Automated test matrix and regression testing
- **Professional Visualizations** - Real-time TUI with QUIC Bottom integration

## Supported Protocols

- **QUIC** - Fast and reliable transport protocol
- **MASQUE** - Protocol for tunneling and proxying
- **ICE/STUN/TURN** - Protocols for NAT traversal and P2P connections
- **TLS 1.3** - Modern cryptography for secure connections
- **HTTP/3** - HTTP over QUIC

[![Watch demo video](https://customer-aedqzjrbponeadcg.cloudflarestream.com/d31af3803090bcb58597de9fe685a746/thumbnails/thumbnail.jpg)](https://customer-aedqzjrbponeadcg.cloudflarestream.com/d31af3803090bcb58597de9fe685a746/watch)

[![Build](https://github.com/twogc/quic-test/workflows/CI/badge.svg)](https://github.com/twogc/quic-test/actions)
[![Lint](https://github.com/twogc/quic-test/workflows/Lint/badge.svg)](https://github.com/twogc/quic-test/actions)
[![Security](https://github.com/twogc/quic-test/workflows/Security/badge.svg)](https://github.com/twogc/quic-test/security)
[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)
[![Rust Version](https://img.shields.io/badge/Rust-1.70+-orange.svg)](https://rust-lang.org/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)

## Quick Start with QUIC Bottom

### Prerequisites
- Go 1.25+
- Rust 1.70+
- Linux (for network simulation features)

### Installation

```bash
# Clone the repository
git clone https://github.com/twogc/quic-test.git
cd quic-test

# Build QUIC Bottom (Rust)
cd quic-bottom
cargo build --release --bin quic-bottom-real
cd ..

# Build Go application
go build -o bin/quic-test .
```

### Running with Real-time Visualizations

```bash
# Server with QUIC Bottom visualization
./bin/quic-test --mode=server --quic-bottom

# Client with QUIC Bottom visualization
./bin/quic-test --mode=client --addr=localhost:9000 --quic-bottom

# Test with QUIC Bottom visualization
./bin/quic-test --mode=test --quic-bottom --duration=30s

# Using the integrated script
./run_with_quic_bottom.sh --mode=test --duration=30s
```

## QUIC Bottom Features

### Real-time Visualizations
- **Time-series Graphs** - Latency, throughput, connections, errors
- **Performance Heatmaps** - Visual performance data representation
- **Correlation Analysis** - Statistical correlation between metrics
- **Anomaly Detection** - Real-time anomaly detection and alerts

### Interactive Controls
- `q/ESC` - Quit
- `r` - Reset all data
- `h` - Show help
- `1-5` - Switch views (Dashboard, Analytics, Network, Security, Cloud)
- `a` - All views
- `n` - Toggle network simulation
- `+/-` - Change network preset
- `s` - Toggle security testing
- `d` - Toggle cloud deployment
- `i` - Scale cloud instances

### View Modes
1. **Dashboard** - Basic graphs + heatmap + anomaly detection
2. **Analytics** - Correlation analysis + anomaly detection
3. **Network** - Network simulation status and controls
4. **Security** - Security testing status and results
5. **Cloud** - Cloud deployment status and controls
6. **All** - Complete overview of all features

## Network Simulation

### Preset Profiles
- **excellent** - 5ms latency, 0.1% loss, 1 Gbps
- **good** - 20ms latency, 1% loss, 100 Mbps
- **poor** - 100ms latency, 5% loss, 10 Mbps
- **mobile** - 200ms latency, 10% loss, 5 Mbps (with reordering)
- **satellite** - 500ms latency, 2% loss, 2 Mbps (with duplication)
- **adversarial** - 1000ms latency, 20% loss, 1 Mbps (with corruption)

### Real Linux tc Integration
```bash
# Network simulation requires root privileges
sudo ./bin/quic-test --mode=test --quic-bottom
```

## Security Testing

### TLS/QUIC Security Analysis
- **TLS Version Validation** - TLS 1.2, TLS 1.3 support
- **Cipher Suite Analysis** - Strong cipher validation
- **Certificate Validation** - Certificate chain verification
- **0-RTT Security Testing** - Early data security analysis
- **Key Rotation Testing** - Cryptographic key management
- **Anti-replay Protection** - Replay attack prevention

### Attack Simulation
- **MITM Attack Simulation** - Man-in-the-middle attack testing
- **Replay Attack Testing** - Packet replay analysis
- **DoS Attack Simulation** - Denial of service testing
- **Timing Attack Analysis** - Side-channel attack detection

## Cloud Integration

### Multi-cloud Support
- **AWS** - EC2, ALB, CloudWatch integration
- **Azure** - Virtual Machines, Load Balancer, Monitor
- **GCP** - Compute Engine, Load Balancer, Stackdriver
- **DigitalOcean** - Droplets, Load Balancer
- **Linode** - Instances, NodeBalancer

### Auto-scaling Features
- **Dynamic Scaling** - 1-5 instances based on metrics
- **Load Balancer Integration** - ALB, NLB, GCP LB
- **SSL/TLS Termination** - Secure connection handling
- **Health Checks** - Automated monitoring and alerts

## HTTP API

### Endpoints
- `POST /api/metrics` - Receive metrics from Go application
- `GET /health` - Health check
- `GET /api/current` - Get current metrics

### Metrics Structure
```json
{
  "timestamp": 1640995200,
  "latency": 25.5,
  "throughput": 150.2,
  "connections": 1,
  "errors": 0,
  "packet_loss": 0.1,
  "retransmits": 2,
  "jitter": 5.2,
  "congestion_window": 1000,
  "rtt": 25.5,
  "bytes_received": 1024000,
  "bytes_sent": 1024000,
  "streams": 1,
  "handshake_time": 150.0
}
```

## Usage

### Basic QUIC Testing
```bash
# Server
go run main.go --mode=server --addr=:9000

# Client
go run main.go --mode=client --addr=127.0.0.1:9000 --connections=2 --streams=4 --packet-size=1200 --rate=100 --report=report.md --report-format=md --pattern=random

# Full test (server+client)
go run main.go --mode=test
```

### With QUIC Bottom Visualization
```bash
# Server with real-time visualization
go run main.go --mode=server --addr=:9000 --quic-bottom

# Client with real-time visualization
go run main.go --mode=client --addr=127.0.0.1:9000 --quic-bottom

# Test with real-time visualization
go run main.go --mode=test --quic-bottom --duration=30s
```

### Experimental QUIC Features
```bash
# BBRv2 Congestion Control
go run main.go --mode=experimental --cc=bbrv2 --ackfreq=3 --fec=0.1

# ACK Frequency Optimization
go run main.go --mode=experimental --ackfreq=5 --cc=cubic

# FEC (Forward Error Correction)
go run main.go --mode=experimental --fec=0.05 --cc=bbrv2
```

### Network Simulation
```bash
# Excellent network conditions
go run main.go --mode=test --network-profile=excellent --quic-bottom

# Mobile network simulation
go run main.go --mode=test --network-profile=mobile --quic-bottom

# Adversarial network conditions
go run main.go --mode=test --network-profile=adversarial --quic-bottom
```

### Security Testing
```bash
# TLS 1.3 security testing
go run main.go --mode=test --security-test --tls-version=1.3 --quic-bottom

# QUIC security analysis
go run main.go --mode=test --security-test --quic-security --quic-bottom

# Attack simulation
go run main.go --mode=test --security-test --attack-simulation --quic-bottom
```

### Cloud Deployment
```bash
# AWS deployment
go run main.go --mode=test --cloud-deploy --provider=aws --region=us-east-1 --quic-bottom

# Azure deployment
go run main.go --mode=test --cloud-deploy --provider=azure --region=eastus --quic-bottom

# GCP deployment
go run main.go --mode=test --cloud-deploy --provider=gcp --region=us-central1 --quic-bottom
```

## Architecture

### Go Application (QUIC Tester)
```
main.go
├── Metrics Collection
├── HTTP API Bridge (port 8080)
├── Network Simulation
├── Security Testing
└── Cloud Deployment
```

### Rust Application (QUIC Bottom)
```
quic-bottom/
├── HTTP API Client
├── Real-time TUI
├── Professional Visualizations
├── Interactive Controls
└── Metrics Processing
```

### Communication Flow
```
Go QUIC Tester → HTTP API → Rust QUIC Bottom → TUI Display
     ↓              ↓              ↓
  Real Metrics → JSON Format → Professional Graphs
```

## Performance Features

### Real-time Updates
- **100ms update interval** for smooth real-time visualization
- **HTTP API** for low-latency metrics transmission
- **Efficient data structures** for high-performance rendering

### Professional Visualizations
- **Time-series graphs** with proper scaling and labels
- **Heatmaps** for performance data visualization
- **Correlation matrices** for statistical analysis
- **Anomaly detection** with real-time alerts

## Development

### Building from Source
```bash
# Build Go application
go build -o bin/quic-test .

# Build QUIC Bottom (Rust)
cd quic-bottom
cargo build --release --bin quic-bottom-real
cd ..

# Build all tools
make build
```

### Running Tests
```bash
# Run Go tests
go test ./...

# Run Rust tests
cd quic-bottom
cargo test
cd ..
```

### Development Mode
```bash
# Run with debug logging
RUST_LOG=debug ./quic-bottom/target/release/quic-bottom-real

# Run Go with debug logging
go run main.go --mode=test --quic-bottom --debug
```

## Documentation

- [Architecture Guide](docs/ARCHITECTURE.md)
- [API Documentation](docs/api.md)
- [Deployment Guide](docs/deployment.md)
- [Docker Guide](docs/docker.md)
- [Usage Guide](docs/usage.md)
- [Real Integration Report](REAL_INTEGRATION_REPORT.md)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [QUIC-Go](https://github.com/quic-go/quic-go) - Go QUIC implementation
- [Bottom](https://github.com/ClementTsang/bottom) - System monitor inspiration
- [Ratatui](https://github.com/ratatui-org/ratatui) - Rust TUI framework
- [Warp](https://github.com/seanmonstar/warp) - Rust HTTP framework

## What's New

### v2.0.0 - Real-time QUIC Bottom Integration
- **Real-time metrics visualization** with professional TUI
- **HTTP API integration** between Go and Rust applications
- **Network simulation** with real Linux tc integration
- **Security testing** with comprehensive QUIC analysis
- **Cloud deployment** with multi-cloud support
- **Interactive controls** for real-time parameter adjustment
- **Advanced analytics** with heatmaps, correlation, and anomaly detection

---

**This is a complete, production-ready QUIC testing and monitoring platform with professional real-time visualizations!**