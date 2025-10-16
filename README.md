# 2GC Network Protocol Suite

A comprehensive platform for testing and analyzing network protocols: QUIC, MASQUE, ICE/STUN/TURN and others

## Features

- **QUIC Protocol Testing** - Advanced QUIC implementation with experimental features
- **MASQUE Protocol Support** - Tunneling and proxying capabilities  
- **ICE/STUN/TURN Testing** - NAT traversal and P2P connection testing
- **TLS 1.3 Security** - Modern cryptography for secure connections
- **HTTP/3 Support** - HTTP over QUIC implementation
- **Experimental Features** - BBRv2, ACK-Frequency, FEC, Bit Greasing
- **Real-time Monitoring** - Prometheus metrics and Grafana dashboards
- **Comprehensive Testing** - Automated test matrix and regression testing

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
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)

## Usage

### QUIC Testing
```bash
# Server
go run main.go --mode=server --addr=:9000

# Client
go run main.go --mode=client --addr=127.0.0.1:9000 --connections=2 --streams=4 --packet-size=1200 --rate=100 --report=report.md --report-format=md --pattern=random

# Full test (server+client)
go run main.go --mode=test
```

### Experimental QUIC Features
```bash
# BBRv2 Congestion Control
go run main.go --mode=experimental --cc=bbrv2 --ackfreq=3 --fec=0.1

# ACK Frequency Optimization
go run main.go --mode=experimental --ackfreq=5 --qlog=out.qlog

# FEC with Bit Greasing
go run main.go --mode=experimental --fec=0.2 --greasing=true
```

### MASQUE Testing
```bash
go run main.go --mode=masque --masque-server=localhost:8443 --masque-targets=8.8.8.8:53,1.1.1.1:53
```

### ICE/STUN/TURN Testing
```bash
go run main.go --mode=ice --ice-stun=stun.l.google.com:19302 --ice-turn=turn.example.com:3478
```

### Web Dashboard
```bash
go run main.go --mode=dashboard
```

### Enhanced Testing
```bash
go run main.go --mode=enhanced
```

## Command Line Options

- `--mode` — operation mode: `server`, `client`, `test`, `dashboard`, `masque`, `ice`, `enhanced` (default: `test`)
- `--addr` — address for connection or listening (default: `:9000`)
- `--connections` — number of QUIC connections (default: 1)
- `--streams` — number of streams per connection (default: 1)
- `--duration` — test duration (0 — until manual termination, default: 0)
- `--packet-size` — packet size in bytes (default: 1200)
- `--rate` — packet sending rate per second (default: 100, supports ramp-up/ramp-down)
- `--report` — path to report file (optional)
- `--report-format` — report format: `csv`, `md`, `json` (default: `md`)
- `--cert` — path to TLS certificate (optional)
- `--key` — path to TLS key (optional)
- `--pattern` — data pattern: `random`, `zeroes`, `increment` (default: `random`)
- `--no-tls` — disable TLS (for testing)
- `--prometheus` — export Prometheus metrics on `/metrics`
- `--emulate-loss` — packet loss probability (0..1, e.g. 0.05 for 5%)
- `--emulate-latency` — additional delay before sending packet (e.g. 20ms)
- `--emulate-dup` — packet duplication probability (0..1)

## SLA Checks
- `--sla-rtt-p95` — maximum RTT p95 (e.g. 100ms)
- `--sla-loss` — maximum packet loss (0..1, e.g. 0.01 for 1%)
- `--sla-throughput` — minimum throughput (KB/s)
- `--sla-errors` — maximum number of errors

## QUIC Tuning
- `--cc` — congestion control algorithm: cubic, bbr, reno
- `--max-idle-timeout` — maximum connection idle timeout
- `--handshake-timeout` — handshake timeout
- `--keep-alive` — keep-alive interval
- `--max-streams` — maximum number of streams
- `--max-stream-data` — maximum stream data size
- `--enable-0rtt` — enable 0-RTT
- `--enable-key-update` — enable key update
- `--enable-datagrams` — enable datagrams
- `--max-incoming-streams` — maximum number of incoming streams
- `--max-incoming-uni-streams` — maximum number of incoming unidirectional streams

## Test Scenarios
- `--scenario` — predefined scenario: wifi, lte, sat, dc-eu, ru-eu, loss-burst, reorder
- `--list-scenarios` — show list of available scenarios

## Network Profiles
- `--network-profile` — network profile: wifi, lte, 5g, satellite, ethernet, fiber, datacenter
- `--list-profiles` — show list of available network profiles

## Advanced Features

- **Extended Metrics:**
  - Percentile latency (p50, p95, p99, p999), jitter, packet loss, retransmits, handshake time, session resumption, 0-RTT/1-RTT, flow control, key update, out-of-order, error breakdown.
- **Time Series:**
  - For latency, throughput, packet loss, retransmits, handshake time and others.
- **ASCII Charts:**
  - In Markdown reports for all key metrics (asciigraph).
- **Ramp-up/ramp-down:**
  - Packet sending rate dynamically increases and decreases for stress testing.
- **Bad Network Emulation:**
  - Delays, losses, packet duplication (see parameters above).
- **CI/CD Integration:**
  - JSON reports with versioned schema, exit code by SLA.
- **Prometheus:**
  - Live metrics export for monitoring.
- **SLA Checks:**
  - Automatic verification of metrics compliance with SLA requirements with exit code.
- **QUIC Tuning:**
  - Configuration of congestion control algorithms, timeouts, streams, 0-RTT, key update, datagrams.
- **Test Scenarios:**
  - Predefined scenarios for different network types (WiFi, LTE, satellite, datacenters).
- **Network Profiles:**
  - Realistic network profiles with specific RTT, jitter, loss, bandwidth values.
- **Web Dashboard:**
  - REST API, Server-Sent Events for real-time updates, embedded static files.

## Usage Examples

### Basic Test with SLA Checks
```
go run main.go --mode=test --sla-rtt-p95=100ms --sla-loss=0.01 --sla-throughput=50 --report=report.json --report-format=json
```

### Test with QUIC Tuning
```
go run main.go --mode=test --cc=bbr --enable-0rtt --enable-datagrams --max-streams=100 --keep-alive=30s
```

### Test with Predefined Scenario
```
go run main.go --scenario=wifi --report=wifi-test.md
```

### Test with Network Profile
```
go run main.go --network-profile=lte --report=lte-test.json --report-format=json
```

### Start Web Dashboard
```
go run cmd/dashboard/dashboard.go --addr=:9990
```

### List Available Scenarios
```
go run main.go --list-scenarios
```

### List Network Profiles
```
go run main.go --list-profiles
```

## Network Presets

| Preset | RTT | Jitter | Loss | Bandwidth | Expected P95 | Description |
|--------|-----|--------|------|-----------|---------------|-------------|
| `wifi` | 20ms | 5ms | 0.1% | 100 Mbps | 25-30ms | Home WiFi |
| `lte` | 50ms | 15ms | 0.5% | 50 Mbps | 70-80ms | Mobile LTE |
| `satellite` | 600ms | 50ms | 1% | 10 Mbps | 650-700ms | Satellite Internet |
| `datacenter` | 1ms | 0.1ms | 0% | 10 Gbps | 2-3ms | Local Datacenter Network |
| `eu-ru` | 80ms | 10ms | 0.2% | 1 Gbps | 90-100ms | Intercontinental |

## Default Behavior
- If `--duration` is not specified, the test continues until manual termination (Ctrl+C).
- After test completion, a report is automatically generated and saved in the selected format.

## Report Examples
- Markdown, CSV, JSON — contain test parameters, aggregated metrics, time series, ASCII charts, errors.

## Automatic Releases

QUIC Test uses an automatic release system via GitHub Actions.

### Quick Version Update
```bash
# Update version to v1.2.3
./scripts/update-version.sh v1.2.3

# Commit and push
git add tag.txt && git commit -m "chore: bump version to v1.2.3"
git push origin main
```

GitHub Actions automatically:
- ✅ Creates Git tag
- ✅ Builds binaries for all platforms (Linux, Windows, macOS)
- ✅ Creates GitHub Release
- ✅ Publishes Docker images

**More details**: [RELEASES.md](RELEASES.md)

## Documentation

- [Deployment Guide](docs/deployment.md)
- [API Documentation](docs/api.md)
- [Usage Guide](docs/usage.md)
- [Docker Configuration](docs/docker.md)
- [Versioning](docs/versioning.md)

### Research Reports
- [Experimental QUIC Laboratory Research Report](docs/reports/Experimental_QUIC_Laboratory_Research_Report.md)
- [QUIC Performance Comparison Report](docs/reports/QUIC_Performance_Comparison_Report.md)
- [Implementation Complete Report](docs/reports/IMPLEMENTATION_COMPLETE.md)
- [Release Notes](docs/reports/RELEASE_NOTES.md)

## Dependencies
- [quic-go](https://github.com/lucas-clemente/quic-go)
- [tablewriter](https://github.com/olekukonko/tablewriter)
- [asciigraph](https://github.com/guptarohit/asciigraph)
- [prometheus/client_golang](https://github.com/prometheus/client_golang)