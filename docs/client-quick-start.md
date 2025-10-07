# QUIC Client Quick Start Guide

**Server Status**: ✅ **ONLINE**  
**Server Address**: `212.233.79.160:9000`  
**Protocol**: QUIC over UDP  
**TLS**: Disabled (for testing)  

## 🚀 Quick Connection

### 1. Basic Connection Test
```bash
# Test UDP connectivity to server
nc -u 212.233.79.160 9000

# Test with timeout
timeout 5 nc -u 212.233.79.160 9000
```

### 2. Client Configuration
```yaml
# client-config.yaml
server:
  address: "212.233.79.160:9000"
  protocol: "quic"
  tls: false

connections:
  max_connections: 10          # Multiple connections for throughput
  rate_per_connection: 15      # Safe zone (avoid 26-35 pps)
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

## 🔧 System Optimizations

### 1. UDP Buffer Tuning
```bash
# Apply UDP optimizations
sudo sysctl -w net.core.rmem_max=134217728
sudo sysctl -w net.core.rmem_default=134217728
sudo sysctl -w net.core.wmem_max=134217728
sudo sysctl -w net.core.wmem_default=134217728

# UDP-specific settings
sudo sysctl -w net.ipv4.udp_mem="102400 873800 16777216"
sudo sysctl -w net.ipv4.udp_rmem_min=8192
sudo sysctl -w net.ipv4.udp_wmem_min=8192
```

### 2. Process Limits
```bash
# Increase limits
ulimit -n 65536
ulimit -u 32768

# Permanent limits
echo "quic-client soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "quic-client hard nofile 65536" | sudo tee -a /etc/security/limits.conf
```

### 3. Network Stack
```bash
# General optimizations
sudo sysctl -w net.core.netdev_max_backlog=5000
sudo sysctl -w net.core.somaxconn=65535
sudo sysctl -w net.ipv4.tcp_congestion_control=bbr
```

## 📊 Monitoring

### 1. Server Metrics
```bash
# Check server status
curl http://212.233.79.160:2113/metrics

# Check server connections
curl http://212.233.79.160:2113/metrics | grep connections

# Check server rate
curl http://212.233.79.160:2113/metrics | grep rate
```

### 2. Client Metrics
```bash
# Start client with monitoring
./quic-client \
  --server="212.233.79.160:9000" \
  --connections=10 \
  --rate=15 \
  --duration=60s \
  --no-tls \
  --prometheus

# Check client metrics
curl http://localhost:2112/metrics
```

### 3. Web Interfaces
- **Prometheus Server**: http://212.233.79.160:9090
- **Grafana**: http://212.233.79.160:3000
- **Jaeger**: http://212.233.79.160:16686

## 🎯 Performance Guidelines

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

### 🚀 High Performance Strategy
- Use **multiple connections** (10-20)
- Keep **rate per connection low** (15 pps)
- Enable **datagrams and 0-RTT**
- Monitor **critical zone alerts**

## 🔍 Troubleshooting

### 1. Connection Issues
```bash
# Check network connectivity
ping 212.233.79.160

# Check UDP port
nc -u 212.233.79.160 9000

# Check firewall
sudo ufw status | grep 9000
```

### 2. Performance Issues
```bash
# Check rate
curl http://localhost:2112/metrics | grep rate

# Check latency
curl http://localhost:2112/metrics | grep latency

# Check errors
curl http://localhost:2112/metrics | grep errors
```

### 3. Server Issues
```bash
# Check server status
curl http://212.233.79.160:2113/metrics

# Check server logs
docker logs 2gc-network-server
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

## 🎉 Ready to Test!

**Server is online and ready for connections!**

- ✅ **UDP Port 9000**: Open
- ✅ **Prometheus**: http://212.233.79.160:2113/metrics
- ✅ **Monitoring**: Enabled
- ✅ **Optimizations**: Applied

**Start testing with the commands above!**
