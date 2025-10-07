# Server Connection Information

**Date**: October 7, 2025  
**Status**: ✅ **ONLINE**  

## 🌐 Server Details

### Primary Endpoints
- **QUIC Server**: `212.233.79.160:9000` (UDP)
- **Prometheus Metrics**: `http://212.233.79.160:2113/metrics`
- **pprof Profiling**: `http://212.233.79.160:6060/debug/pprof/`

### Monitoring Interfaces
- **Prometheus UI**: `http://212.233.79.160:9090`
- **Grafana Dashboard**: `http://212.233.79.160:3000`
- **Jaeger Tracing**: `http://212.233.79.160:16686`

## 🔧 Server Configuration

### Applied Optimizations
- ✅ **Max Connections**: 1000
- ✅ **Rate per Connection**: 20 pps (safe zone)
- ✅ **Connection Timeout**: 60 seconds
- ✅ **Handshake Timeout**: 10 seconds
- ✅ **Keep-Alive**: 30 seconds
- ✅ **Max Streams**: 100 per connection
- ✅ **Datagrams**: Enabled
- ✅ **0-RTT**: Enabled
- ✅ **TLS**: Disabled (for testing)

### Network Configuration
- ✅ **UDP Port 9000**: Open
- ✅ **TCP Port 2113**: Open (Prometheus)
- ✅ **TCP Port 6060**: Open (pprof)
- ✅ **UFW Status**: All ports configured

## 🚀 Quick Connection Test

### 1. Basic Connectivity
```bash
# Test UDP connectivity
nc -u 212.233.79.160 9000

# Test with timeout
timeout 5 nc -u 212.233.79.160 9000

# Check server metrics
curl http://212.233.79.160:2113/metrics
```

### 2. Client Configuration
```yaml
server:
  address: "212.233.79.160:9000"
  protocol: "quic"
  tls: false

connections:
  max_connections: 10
  rate_per_connection: 15  # Safe zone
  connection_timeout: 60s
  handshake_timeout: 10s
  keep_alive: 30s

streams:
  max_streams_per_connection: 8
  stream_timeout: 30s

performance:
  enable_datagrams: true
  enable_0rtt: true
  congestion_control: "bbr"

monitoring:
  prometheus_port: 2112
  metrics_interval: 5s
```

## 📊 Performance Guidelines

### ✅ Safe Zone (Recommended)
- **Rate per Connection**: 15-20 pps
- **Connections**: 10-20 for high throughput
- **Streams per Connection**: 8-16
- **Expected Latency**: < 50ms
- **Expected Jitter**: < 10ms

### ⚠️ Critical Zone (AVOID)
- **Rate per Connection**: 26-35 pps
- **High Latency**: > 100ms
- **High Jitter**: > 50ms
- **Connection Errors**: > 1%

## 🔍 Troubleshooting

### Connection Issues
```bash
# Check network connectivity
ping 212.233.79.160

# Check UDP port
nc -u 212.233.79.160 9000

# Check server status
curl http://212.233.79.160:2113/metrics
```

### Performance Issues
```bash
# Check server rate
curl http://212.233.79.160:2113/metrics | grep rate

# Check server connections
curl http://212.233.79.160:2113/metrics | grep connections

# Check server errors
curl http://212.233.79.160:2113/metrics | grep errors
```

## 📋 Quick Commands

### Start Client
```bash
# Basic test
./quic-client --server="212.233.79.160:9000" --no-tls

# Performance test
./quic-client --server="212.233.79.160:9000" --connections=10 --rate=15 --duration=60s --no-tls --prometheus

# High throughput test
./quic-client --server="212.233.79.160:9000" --connections=20 --rate=15 --duration=300s --no-tls --prometheus
```

### Monitor Performance
```bash
# Live monitoring
watch -n 5 'curl -s http://localhost:2112/metrics | grep -E "(connections|rate|latency)"'

# Server monitoring
watch -n 5 'curl -s http://212.233.79.160:2113/metrics | grep -E "(connections|rate|errors)"'
```

## 🎯 Ready for Testing!

**Server is online and ready for connections!**

- ✅ **UDP Port 9000**: Open and accessible
- ✅ **Prometheus Metrics**: Available at http://212.233.79.160:2113/metrics
- ✅ **Monitoring**: All interfaces active
- ✅ **Optimizations**: Applied and tested

**Start testing with the commands above!**

