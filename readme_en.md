# quic-test

Professional QUIC protocol testing platform for network engineers, researchers, and educators.

[![CI](https://github.com/twogc/quic-test/actions/workflows/pipeline.yml/badge.svg)](https://github.com/twogc/quic-test/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/twogc/quic-test)](https://goreportcard.com/report/github.com/twogc/quic-test)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/docker/v/mlanies/quic-test?label=docker)](https://hub.docker.com/r/mlanies/quic-test)

**English** | [Русский](readme.md)

## What is this?

`quic-test` is a tool for testing and analyzing QUIC protocol performance. Designed for educational and research purposes, with focus on reproducibility and detailed analytics.

**Key Features:**
- Measure latency, jitter, throughput for QUIC and TCP
- Emulate various network conditions (loss, delay, bandwidth)
- Real-time TUI visualization (`quic-bottom`)
- Prometheus metrics export
- Machine learning integration (AI Routing Lab)

## Quick Start

### Docker (recommended)

```bash
# Run client (performance test)
docker run mlanies/quic-test:latest --mode=client --server=demo.quic.tech:4433

# Run server
docker run -p 4433:4433/udp mlanies/quic-test:latest --mode=server
```

### From source

```bash
# Requirements: Go 1.21+, clang (for FEC)
git clone https://github.com/twogc/quic-test
cd quic-test

# Build FEC library
cd internal/fec && make && cd ../..

# Build
go build -o quic-test cmd/quic-test/main.go

# Run
./quic-test --mode=client --server=demo.quic.tech:4433
```

## Basic Usage

```bash
# Simple latency/throughput test
./quic-test --mode=client --server=localhost:4433 --duration=30s

# Compare QUIC vs TCP
./quic-test --mode=client --compare-tcp --duration=60s

# Emulate mobile network
./quic-test --profile=mobile --duration=30s

# TUI monitoring
quic-bottom --server=localhost:4433
```

## Architecture

```
quic-test/
├── cmd/
│   ├── quic-test/      # Main CLI
│   └── quic-bottom/    # TUI monitor
├── client/             # QUIC client
├── server/             # QUIC server
├── internal/
│   ├── quic/           # QUIC logic
│   ├── fec/            # Forward Error Correction (C++/AVX2)
│   ├── metrics/        # Prometheus metrics
│   └── congestion/     # BBRv2/BBRv3
└── docs/               # Documentation
```

**Details:** [docs/architecture.md](docs/architecture.md)

## Features

### Stable Features

- QUIC client/server (based on quic-go)
- RTT, jitter, throughput measurements
- Network profile emulation (mobile, satellite, fiber)
- TUI visualization (`quic-bottom`)
- Prometheus export
- BBRv2 congestion control

### Experimental Features

- BBRv3 congestion control
- Forward Error Correction (FEC) with AVX2
- MASQUE VPN testing
- TCP-over-QUIC tunneling
- ICE/STUN/TURN tests

### Planned Features (Roadmap)

- HTTP/3 load testing
- Automatic anomaly detection
- Multi-cloud deployment
- WebTransport support

**Full roadmap:** [docs/roadmap.md](docs/roadmap.md)

## Documentation

- **[CLI Reference](docs/cli.md)** — complete command reference
- **[Architecture](docs/architecture.md)** — detailed architecture
- **[Education](docs/education.md)** — lab materials for universities
- **[AI Integration](docs/ai-routing-integration.md)** — AI Routing Lab integration
- **[Case Studies](docs/case-studies.md)** — test results with methodology

## For Universities

Designed with education in mind. Includes ready-to-use lab materials:

- **Lab #1:** QUIC Basics — handshake, 0-RTT, connection migration
- **Lab #2:** Congestion Control — BBRv2 vs BBRv3 comparison
- **Lab #3:** Performance — QUIC vs TCP under various conditions

**Details:** [docs/education.md](docs/education.md)

## AI Routing Lab Integration

`quic-test` exports metrics to Prometheus, which are used in [AI Routing Lab](https://github.com/twogc/ai-routing-lab) for training route prediction models.

**Example:**
```bash
# Run with Prometheus export
./quic-test --mode=server --prometheus-port=9090

# AI Routing Lab collects metrics
curl http://localhost:9090/metrics
```

**Details:** [docs/ai-routing-integration.md](docs/ai-routing-integration.md)

## Development

```bash
# Run tests
go test ./...

# Linting
golangci-lint run

# Build Docker image
docker build -t quic-test .
```

## License

MIT License. See [LICENSE](LICENSE).

## Contacts

- **GitHub:** [twogc/quic-test](https://github.com/twogc/quic-test)
- **Blog:** [cloudbridge-research.ru](https://cloudbridge-research.ru)
- **Email:** info@cloudbridge-research.ru
- **Docker Hub:** [mlanies/quic-test](https://hub.docker.com/r/mlanies/quic-test)