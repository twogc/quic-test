# QUIC Test Usage Guide

## Overview

This guide provides comprehensive instructions for using the QUIC Test tool to perform network performance testing, analyze results, and integrate with monitoring systems.

## Quick Start

### Basic Test

1. **Start the server**
   ```bash
   ./quic-test --mode=server --addr=:9000
   ```

2. **Run a client test**
   ```bash
   ./quic-test --mode=client --addr=127.0.0.1:9000 --connections=2 --streams=4
   ```

3. **Run a complete test (server + client)**
   ```bash
   ./quic-test --mode=test --connections=2 --streams=4 --duration=30s
   ```

### Web Dashboard

1. **Start the dashboard**
   ```bash
   ./dashboard --addr=:9990
   ```

2. **Open in browser**
   ```
   http://localhost:9990
   ```

## Command Line Interface

### Basic Commands

#### Server Mode
```bash
./quic-test --mode=server [options]
```

#### Client Mode
```bash
./quic-test --mode=client --addr=<server> [options]
```

#### Test Mode (Server + Client)
```bash
./quic-test --mode=test [options]
```

### Configuration Options

#### Network Configuration

| Option | Description | Default |
|--------|-------------|---------|
| `--addr` | Address to connect/listen | `:9000` |
| `--connections` | Number of QUIC connections | `1` |
| `--streams` | Number of streams per connection | `1` |
| `--packet-size` | Packet size in bytes | `1200` |
| `--rate` | Packets per second | `100` |
| `--duration` | Test duration | `0` (until stopped) |

#### Network Emulation

| Option | Description | Default |
|--------|-------------|---------|
| `--emulate-loss` | Packet loss probability (0-1) | `0` |
| `--emulate-latency` | Additional latency | `0` |
| `--emulate-dup` | Packet duplication probability (0-1) | `0` |

#### TLS Configuration

| Option | Description | Default |
|--------|-------------|---------|
| `--cert` | TLS certificate file | (auto-generated) |
| `--key` | TLS private key file | (auto-generated) |
| `--no-tls` | Disable TLS | `false` |

#### Monitoring

| Option | Description | Default |
|--------|-------------|---------|
| `--prometheus` | Enable Prometheus metrics | `false` |
| `--pprof-addr` | Profiling address | (disabled) |

#### Reporting

| Option | Description | Default |
|--------|-------------|---------|
| `--report` | Report file path | (stdout) |
| `--report-format` | Report format (json/csv/md) | `md` |

### Advanced Configuration

#### QUIC Tuning

| Option | Description | Default |
|--------|-------------|---------|
| `--cc` | Congestion control (cubic/bbr/reno) | (default) |
| `--max-idle-timeout` | Max idle timeout | (default) |
| `--handshake-timeout` | Handshake timeout | (default) |
| `--keep-alive` | Keep-alive interval | (default) |
| `--max-streams` | Max streams per connection | (default) |
| `--max-stream-data` | Max stream data size | (default) |
| `--enable-0rtt` | Enable 0-RTT | `false` |
| `--enable-key-update` | Enable key update | `false` |
| `--enable-datagrams` | Enable datagrams | `false` |

#### SLA Configuration

| Option | Description | Default |
|--------|-------------|---------|
| `--sla-rtt-p95` | Max RTT p95 | (disabled) |
| `--sla-loss` | Max packet loss | (disabled) |
| `--sla-throughput` | Min throughput (KB/s) | (disabled) |
| `--sla-errors` | Max error count | (disabled) |

## Test Scenarios

### Predefined Scenarios

#### WiFi Network
```bash
./quic-test --scenario=wifi
```

#### LTE Network
```bash
./quic-test --scenario=lte
```

#### Satellite Network
```bash
./quic-test --scenario=sat
```

#### Data Center
```bash
./quic-test --scenario=dc-eu
```

#### International Link
```bash
./quic-test --scenario=ru-eu
```

#### Loss Burst
```bash
./quic-test --scenario=loss-burst
```

#### Packet Reordering
```bash
./quic-test --scenario=reorder
```

### Network Profiles

#### WiFi 802.11n
```bash
./quic-test --network-profile=wifi
```

#### WiFi 802.11ac (5GHz)
```bash
./quic-test --network-profile=wifi-5g
```

#### LTE 4G
```bash
./quic-test --network-profile=lte
```

#### LTE Advanced
```bash
./quic-test --network-profile=lte-advanced
```

#### 5G NR
```bash
./quic-test --network-profile=5g
```

#### Satellite Internet
```bash
./quic-test --network-profile=satellite
```

#### Satellite LEO (Starlink)
```bash
./quic-test --network-profile=satellite-leo
```

#### Ethernet 1Gbps
```bash
./quic-test --network-profile=ethernet
```

#### Ethernet 10Gbps
```bash
./quic-test --network-profile=ethernet-10g
```

#### DSL
```bash
./quic-test --network-profile=dsl
```

#### Cable Internet
```bash
./quic-test --network-profile=cable
```

#### Fiber Optic
```bash
./quic-test --network-profile=fiber
```

#### Mobile 3G
```bash
./quic-test --network-profile=mobile-3g
```

#### EDGE Mobile
```bash
./quic-test --network-profile=edge
```

#### International Link
```bash
./quic-test --network-profile=international
```

#### Data Center
```bash
./quic-test --network-profile=datacenter
```

## Web Dashboard Usage

### Dashboard Features

1. **Real-time Metrics**
   - Live connection status
   - Current throughput
   - Latency statistics
   - Error rates

2. **Test Management**
   - Start/stop tests
   - Configure test parameters
   - Apply presets

3. **Report Generation**
   - JSON reports
   - CSV exports
   - Markdown documentation

### API Endpoints

#### Status
```bash
curl http://localhost:9990/status
```

#### Start Test
```bash
curl -X POST http://localhost:9990/run-test \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "test",
    "connections": 2,
    "streams": 4,
    "duration": "30s"
  }'
```

#### Get Metrics
```bash
curl http://localhost:9990/metrics
```

#### Generate Report
```bash
curl "http://localhost:9990/report?format=json"
```

## Monitoring Integration

### Prometheus Metrics

Enable Prometheus metrics:

```bash
./quic-test --mode=test --prometheus
```

Access metrics:

```bash
curl http://localhost:2112/metrics
```

### Grafana Dashboard

1. **Import dashboard**
   ```bash
   curl -X POST \
     -H "Content-Type: application/json" \
     -d @grafana/dashboards/quic-test.json \
     http://admin:admin@localhost:3000/api/dashboards/db
   ```

2. **Configure data source**
   - URL: `http://prometheus:9090`
   - Access: Server (default)

### Key Metrics

#### Connection Metrics
- `quic_connections_current`: Current connections
- `quic_connections_total`: Total connections
- `quic_streams_current`: Current streams
- `quic_streams_total`: Total streams

#### Performance Metrics
- `quic_latency_seconds`: Request latency
- `quic_throughput_bytes_per_second`: Throughput
- `quic_packet_loss_rate`: Packet loss rate

#### Error Metrics
- `quic_errors_total`: Total errors
- `quic_retransmits_total`: Total retransmits
- `quic_handshakes_total`: Total handshakes

## Report Formats

### JSON Report

```bash
./quic-test --mode=test --report=report.json --report-format=json
```

Example output:
```json
{
  "config": {
    "mode": "test",
    "connections": 2,
    "streams": 4
  },
  "metrics": {
    "Success": 100,
    "Errors": 5,
    "LatencyAverage": 25.5,
    "ThroughputAverage": 1000.0
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### CSV Report

```bash
./quic-test --mode=test --report=report.csv --report-format=csv
```

Example output:
```csv
Parameter,Value
Mode,test
Connections,2
Streams,4
Success,100
Errors,5
Average Latency (ms),25.5
```

### Markdown Report

```bash
./quic-test --mode=test --report=report.md --report-format=md
```

Example output:
```markdown
# QUIC Test Report

## Test Configuration
| Parameter | Value |
|-----------|-------|
| Mode | test |
| Connections | 2 |
| Streams | 4 |

## Test Results
| Metric | Value |
|--------|-------|
| Successful Requests | 100 |
| Errors | 5 |
| Average Latency | 25.5 ms |
```

## SLA Testing

### Configure SLA Limits

```bash
./quic-test --mode=test \
  --sla-rtt-p95=100ms \
  --sla-loss=0.01 \
  --sla-throughput=50 \
  --sla-errors=10
```

### Exit Codes

- `0`: All SLA checks passed
- `1`: SLA violations detected
- `2`: Critical failures

### SLA Validation

The tool automatically validates:
- RTT percentiles (p95, p99)
- Packet loss rates
- Throughput thresholds
- Error counts

## Performance Tuning

### System Optimization

1. **Network buffers**
   ```bash
   echo 'net.core.rmem_max = 134217728' >> /etc/sysctl.conf
   echo 'net.core.wmem_max = 134217728' >> /etc/sysctl.conf
   sysctl -p
   ```

2. **File descriptors**
   ```bash
   ulimit -n 65536
   ```

### Application Tuning

1. **Goroutine limits**
   ```bash
   export GOMAXPROCS=4
   ```

2. **Memory limits**
   ```bash
   export GOGC=100
   ```

## Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   lsof -i :9000
   kill -9 <PID>
   ```

2. **Permission denied**
   ```bash
   chmod +x ./quic-test
   ```

3. **TLS certificate issues**
   ```bash
   ./quic-test --no-tls
   ```

### Debug Mode

Enable debug logging:

```bash
export LOG_LEVEL=debug
./quic-test --mode=test
```

### Profiling

Enable profiling:

```bash
./quic-test --mode=test --pprof-addr=:6060
```

Access profiling data:

```bash
go tool pprof http://localhost:6060/debug/pprof/profile
```

## Best Practices

### Test Design

1. **Start with small tests**
   - Low connection counts
   - Short durations
   - Simple scenarios

2. **Gradually increase complexity**
   - More connections
   - Longer durations
   - Complex scenarios

3. **Use appropriate network profiles**
   - Match your target environment
   - Test edge cases
   - Validate assumptions

### Monitoring

1. **Set up alerts**
   - High error rates
   - SLA violations
   - Resource exhaustion

2. **Regular testing**
   - Automated test runs
   - Performance regression detection
   - Capacity planning

### Documentation

1. **Record test results**
   - Save reports
   - Document findings
   - Track changes

2. **Share knowledge**
   - Team documentation
   - Best practices
   - Lessons learned

## Integration Examples

### CI/CD Pipeline

```yaml
# .github/workflows/quic-test.yml
name: QUIC Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run QUIC test
        run: |
          ./quic-test --mode=test \
            --sla-rtt-p95=100ms \
            --sla-loss=0.01 \
            --report=test-results.json \
            --report-format=json
      - name: Upload results
        uses: actions/upload-artifact@v2
        with:
          name: quic-test-results
          path: test-results.json
```

### Monitoring Integration

```bash
# Prometheus configuration
scrape_configs:
  - job_name: 'quic-test'
    static_configs:
      - targets: ['localhost:2112']
    scrape_interval: 5s
```

### Alerting Rules

```yaml
# prometheus-alerts.yml
groups:
  - name: quic-test
    rules:
      - alert: HighErrorRate
        expr: rate(quic_errors_total[5m]) > 0.1
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "High QUIC error rate detected"
```

## Support

### Getting Help

1. **Documentation**
   - README.md
   - API documentation
   - Deployment guide

2. **Community**
   - GitHub issues
   - Discussions
   - Slack channel

3. **Professional Support**
   - Enterprise support
   - Consulting services
   - Training programs
