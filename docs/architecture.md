# Architecture

Technical architecture of `quic-test`.

## Overview

```
┌─────────────────────────────────────────────────────────────┐
│                         quic-test                            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐         ┌──────────────┐                 │
│  │              │         │              │                 │
│  │  CLI Client  │◄────────┤  CLI Server  │                 │
│  │              │  QUIC   │              │                 │
│  └──────┬───────┘         └──────┬───────┘                 │
│         │                        │                          │
│         │                        │                          │
│  ┌──────▼────────────────────────▼───────┐                 │
│  │                                        │                 │
│  │         QUIC Core (quic-go)            │                 │
│  │                                        │                 │
│  │  ┌──────────┐  ┌──────────┐           │                 │
│  │  │  BBRv2   │  │  BBRv3   │           │                 │
│  │  │   CC     │  │   CC     │           │                 │
│  │  └──────────┘  └──────────┘           │                 │
│  │                                        │                 │
│  └────────────────┬───────────────────────┘                 │
│                   │                                          │
│  ┌────────────────▼───────────────────────┐                 │
│  │                                        │                 │
│  │      FEC (C++/AVX2 SIMD)               │                 │
│  │                                        │                 │
│  └────────────────┬───────────────────────┘                 │
│                   │                                          │
│  ┌────────────────▼───────────────────────┐                 │
│  │                                        │                 │
│  │      Metrics & Monitoring              │                 │
│  │                                        │                 │
│  │  ┌──────────┐  ┌──────────┐           │                 │
│  │  │Prometheus│  │   TUI    │           │                 │
│  │  │  Export  │  │ (bottom) │           │                 │
│  │  └──────────┘  └──────────┘           │                 │
│  │                                        │                 │
│  └────────────────────────────────────────┘                 │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Components

### 1. CLI Layer

**Location:** `cmd/quic-test/`, `cmd/quic-bottom/`

- Command-line interface
- Configuration parsing
- Mode selection (client/server)
- Output formatting

**Technologies:**
- Go standard library (`flag`, `os`)
- Cobra (CLI framework)

### 2. QUIC Core

**Location:** `internal/quic/`, `client/`, `server/`

- QUIC protocol implementation (via quic-go)
- Connection management
- Stream multiplexing
- 0-RTT resumption

**Technologies:**
- [quic-go](https://github.com/quic-go/quic-go) v0.40+
- TLS 1.3
- HTTP/3

### 3. Congestion Control

**Location:** `internal/congestion/`

Implements multiple congestion control algorithms:

- **BBRv2** (default, stable)
- **BBRv3** (experimental)
- **NewReno** (fallback)

**Key Features:**
- Bandwidth probing
- RTT measurement
- Packet pacing
- Loss recovery

### 4. Forward Error Correction (FEC)

**Location:** `internal/fec/`

High-performance XOR-based FEC with SIMD optimization.

**Implementation:**
- C++ core (`fec_xor_simd.cpp`)
- AVX2 SIMD instructions
- CGO bindings
- NUMA-aware memory allocation

**Performance:**
- ~10 GB/s throughput on modern CPUs
- <1μs latency overhead

### 5. Network Emulation

**Location:** `internal/network_simulation.go`

Emulates various network conditions:

**Profiles:**
- Mobile (4G/LTE)
- Satellite
- Fiber
- Custom

**Parameters:**
- RTT (Round-Trip Time)
- Bandwidth
- Packet loss
- Jitter

**Implementation:**
- Token bucket for bandwidth limiting
- Delay queue for RTT emulation
- Random drop for packet loss

### 6. Metrics & Monitoring

**Location:** `internal/metrics/`

**Prometheus Metrics:**
```
quic_rtt_seconds{quantile="0.5"}
quic_rtt_seconds{quantile="0.95"}
quic_rtt_seconds{quantile="0.99"}
quic_jitter_seconds
quic_throughput_bytes_per_second
quic_packet_loss_ratio
quic_connections_total
quic_streams_total
quic_handshake_duration_seconds
```

**HDR Histogram:**
- High Dynamic Range histograms
- Accurate percentile calculation
- Low memory overhead

**TUI (quic-bottom):**
- Real-time visualization
- Terminal UI with graphs
- Keyboard navigation

## Data Flow

### Client → Server (Measurement)

```
1. Client initiates QUIC connection
   └─> TLS 1.3 handshake
   └─> 0-RTT if session ticket available

2. Client sends measurement packets
   └─> Timestamped packets
   └─> FEC redundancy (if enabled)
   └─> Congestion control (BBRv2/BBRv3)

3. Server receives and echoes packets
   └─> Measures RTT
   └─> Calculates jitter
   └─> Tracks packet loss

4. Client collects metrics
   └─> Updates HDR histograms
   └─> Exports to Prometheus
   └─> Displays in TUI

5. Connection closes gracefully
   └─> Final statistics
   └─> Report generation
```

### Metrics Export Flow

```
quic-test (client/server)
    │
    ├─> Prometheus HTTP endpoint (:9090/metrics)
    │   └─> Grafana dashboard
    │   └─> AI Routing Lab
    │
    ├─> JSON export
    │   └─> Post-processing scripts
    │
    └─> TUI (quic-bottom)
        └─> Real-time visualization
```

## Build System

### Go Build

```bash
# Standard build
go build -o quic-test cmd/quic-test/main.go

# Optimized build
go build -ldflags="-s -w" -o quic-test cmd/quic-test/main.go
```

### FEC Library Build

```bash
cd internal/fec
make clean
make

# Produces:
# - libfec_avx2.so (Linux)
# - libfec_avx2.dylib (macOS)
# - libfec_scalar.so (fallback)
```

**Makefile targets:**
- `all` — Build for current platform
- `clean` — Remove build artifacts
- `test` — Run C++ unit tests
- `benchmark` — Run performance benchmarks

### Docker Build

```dockerfile
# Multi-stage build
FROM golang:1.21 AS builder
RUN apt-get update && apt-get install -y clang libnuma-dev
COPY . /app
WORKDIR /app
RUN cd internal/fec && make
RUN go build -o quic-test cmd/quic-test/main.go

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y libnuma1
COPY --from=builder /app/quic-test /usr/local/bin/
COPY --from=builder /app/internal/fec/libfec_*.so /usr/local/lib/
ENTRYPOINT ["quic-test"]
```

## Performance Considerations

### Memory

- **Client:** ~50 MB baseline
- **Server:** ~100 MB baseline
- **Per connection:** ~1-2 MB
- **FEC buffers:** Configurable (default 10 MB)

### CPU

- **Idle:** <1% CPU
- **Active measurement:** 10-30% CPU (single core)
- **FEC encoding:** 50-100% CPU (with AVX2)

### Network

- **Minimum bandwidth:** 1 Mbps
- **Recommended:** 10+ Mbps
- **Maximum tested:** 10 Gbps

## Security

### TLS 1.3

- Mandatory for all QUIC connections
- Self-signed certificates for testing
- Let's Encrypt integration (planned)

### Authentication

- mTLS support (experimental)
- Token-based auth (planned)

### Isolation

- No root privileges required
- Unprivileged ports (>1024)
- Containerized deployment (Docker)

## Extensibility

### Plugin System (Planned)

```go
type Plugin interface {
    Init(config Config) error
    OnConnect(conn Connection) error
    OnPacket(packet Packet) error
    OnClose(conn Connection) error
}
```

### Custom Congestion Control

```go
type CongestionControl interface {
    OnPacketSent(packet Packet)
    OnPacketAcked(packet Packet)
    OnPacketLost(packet Packet)
    GetCongestionWindow() int
}
```

## Dependencies

### Core

- `github.com/quic-go/quic-go` — QUIC implementation
- `github.com/prometheus/client_golang` — Metrics
- `github.com/HdrHistogram/hdrhistogram-go` — Histograms

### UI

- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — Styling

### Build

- Go 1.21+
- clang/g++ (for FEC)
- libnuma-dev (Linux only)

## See Also

- [CLI Reference](cli.md)
- [Roadmap](roadmap.md)
- [Contributing](../CONTRIBUTING.md)
