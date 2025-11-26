# CLI Reference

Complete command-line reference for `quic-test`.

## Global Flags

```bash
--mode string          Operation mode: client, server, test (default "client")
--server string        Server address (default "localhost:4433")
--duration duration    Test duration (default 30s)
--verbose             Enable verbose logging
--config string       Path to config file
```

## Client Mode

```bash
quic-test --mode=client [flags]
```

### Client Flags

```bash
--compare-tcp         Run parallel TCP test for comparison
--profile string      Network profile: mobile, satellite, fiber, custom
--streams int         Number of concurrent streams (default 1)
--data-size string    Amount of data to transfer (e.g., 10MB, 1GB)
--prometheus-port int Prometheus metrics port (default 9090)
```

### Examples

```bash
# Basic latency test
quic-test --mode=client --server=demo.quic.tech:4433 --duration=30s

# QUIC vs TCP comparison
quic-test --mode=client --compare-tcp --duration=60s

# Mobile network emulation
quic-test --mode=client --profile=mobile --duration=30s

# High-throughput test
quic-test --mode=client --streams=10 --data-size=1GB
```

## Server Mode

```bash
quic-test --mode=server [flags]
```

### Server Flags

```bash
--listen string       Listen address (default ":4433")
--cert string         TLS certificate path
--key string          TLS private key path
--dashboard          Enable web dashboard (port 8080)
--prometheus-port int Prometheus metrics port (default 9090)
```

### Examples

```bash
# Basic server
quic-test --mode=server

# Server with dashboard
quic-test --mode=server --dashboard

# Custom certificate
quic-test --mode=server --cert=server.crt --key=server.key
```

## Network Profiles

### Built-in Profiles

**Mobile (4G/LTE)**
- RTT: 50-150ms
- Bandwidth: 5-50 Mbps
- Loss: 0.1-2%
- Jitter: 10-30ms

**Satellite**
- RTT: 500-700ms
- Bandwidth: 1-10 Mbps
- Loss: 0.5-5%
- Jitter: 50-100ms

**Fiber**
- RTT: 1-10ms
- Bandwidth: 100-1000 Mbps
- Loss: 0-0.1%
- Jitter: 0.1-1ms

### Custom Profile

```bash
quic-test --mode=client \
  --profile=custom \
  --rtt=100ms \
  --bandwidth=10mbps \
  --loss=1% \
  --jitter=20ms
```

## TUI Monitor (quic-bottom)

```bash
quic-bottom [flags]
```

### Flags

```bash
--server string       Server to monitor (default "localhost:4433")
--refresh duration    Refresh interval (default 1s)
--prometheus string   Prometheus endpoint (default "http://localhost:9090")
```

### Keyboard Shortcuts

- `q` — Quit
- `r` — Refresh
- `↑/↓` — Scroll
- `h` — Help

## Configuration File

Create `quic-test.yaml`:

```yaml
mode: client
server: demo.quic.tech:4433
duration: 60s

client:
  compare_tcp: true
  profile: mobile
  streams: 5

metrics:
  prometheus_port: 9090
  export_interval: 5s

logging:
  level: info
  format: json
```

Use with:
```bash
quic-test --config=quic-test.yaml
```

## Environment Variables

```bash
QUIC_TEST_MODE=client
QUIC_TEST_SERVER=demo.quic.tech:4433
QUIC_TEST_DURATION=30s
QUIC_TEST_VERBOSE=true
```

## Exit Codes

- `0` — Success
- `1` — General error
- `2` — Configuration error
- `3` — Network error
- `4` — TLS error

## Metrics Export

### Prometheus

Metrics available at `http://localhost:9090/metrics`:

```
quic_rtt_seconds
quic_jitter_seconds
quic_throughput_bytes_per_second
quic_packet_loss_ratio
quic_connections_total
quic_streams_total
```

### JSON

```bash
quic-test --mode=client --output=json > results.json
```

### CSV

```bash
quic-test --mode=client --output=csv > results.csv
```

## Advanced Usage

### BBRv3 Testing

```bash
quic-test --mode=client --congestion=bbrv3 --duration=60s
```

### FEC (Forward Error Correction)

```bash
quic-test --mode=client --fec=true --fec-redundancy=0.1
```

### 0-RTT Resumption

```bash
# First connection
quic-test --mode=client --server=demo.quic.tech:4433

# Subsequent connections use 0-RTT
quic-test --mode=client --server=demo.quic.tech:4433 --0rtt
```

## Troubleshooting

### Connection Refused

```bash
# Check if server is running
quic-test --mode=server &

# Test connection
quic-test --mode=client --server=localhost:4433
```

### TLS Certificate Errors

```bash
# Use self-signed certificate
quic-test --mode=server --cert=server.crt --key=server.key

# Client: skip verification (testing only!)
quic-test --mode=client --insecure
```

### High Packet Loss

```bash
# Enable FEC
quic-test --mode=client --fec=true --fec-redundancy=0.2

# Increase timeout
quic-test --mode=client --timeout=60s
```

## See Also

- [Architecture](architecture.md)
- [Education](education.md)
- [Case Studies](case-studies.md)
