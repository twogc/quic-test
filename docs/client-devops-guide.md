# Client-Side DevOps Guide

**Project**: 2GC Network Protocol Suite  
**Date**: October 7, 2025  
**Target**: Client-Side DevOps Implementation  
**Version**: 1.0  

## Executive Summary

This guide provides DevOps recommendations for client-side implementation of the 2GC Network Protocol Suite, focusing on connection optimization, monitoring, and performance tuning for remote clients connecting to QUIC servers.

## Current Server Status

### Server Information
- **Server Address**: `212.233.79.160:9000` (UDP)
- **Prometheus Metrics**: `http://212.233.79.160:2113/metrics`
- **pprof Profiling**: `http://212.233.79.160:6060/debug/pprof/`
- **Protocol**: QUIC over UDP
- **TLS**: Disabled (for testing)
- **Monitoring**: Enabled

### Server Optimizations Applied
- ✅ **Max Connections**: 1000
- ✅ **Rate per Connection**: 20 pps (safe zone)
- ✅ **Connection Timeout**: 60 seconds
- ✅ **Handshake Timeout**: 10 seconds
- ✅ **Keep-Alive**: 30 seconds
- ✅ **Max Streams**: 100 per connection
- ✅ **Datagrams**: Enabled
- ✅ **0-RTT**: Enabled

## Client-Side DevOps Requirements

### 1. Network Configuration

#### UDP Buffer Optimization
```bash
# Apply UDP buffer optimizations
sudo sysctl -w net.core.rmem_max=134217728
sudo sysctl -w net.core.rmem_default=134217728
sudo sysctl -w net.core.wmem_max=134217728
sudo sysctl -w net.core.wmem_default=134217728

# UDP-specific optimizations
sudo sysctl -w net.ipv4.udp_mem="102400 873800 16777216"
sudo sysctl -w net.ipv4.udp_rmem_min=8192
sudo sysctl -w net.ipv4.udp_wmem_min=8192
```

#### Network Stack Tuning
```bash
# General network optimizations
sudo sysctl -w net.core.netdev_max_backlog=5000
sudo sysctl -w net.core.somaxconn=65535

# TCP optimizations (for fallback)
sudo sysctl -w net.ipv4.tcp_congestion_control=bbr
sudo sysctl -w net.ipv4.tcp_rmem="4096 87380 134217728"
sudo sysctl -w net.ipv4.tcp_wmem="4096 65536 134217728"
```

### 2. Process Limits

#### System Limits
```bash
# Increase file descriptor limits
ulimit -n 65536
ulimit -u 32768

# Permanent limits in /etc/security/limits.conf
echo "quic-client soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "quic-client hard nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "quic-client soft nproc 32768" | sudo tee -a /etc/security/limits.conf
echo "quic-client hard nproc 32768" | sudo tee -a /etc/security/limits.conf
```

### 3. Client Configuration

#### Optimal Client Settings
```yaml
client:
server:
  address: "212.233.79.160:9000"
    protocol: "quic"
    tls: false
  
  connections:
    max_connections: 10          # Multiple connections for high throughput
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
    health_check_interval: 30s
```

### 4. Connection Strategy

#### Multi-Connection Approach
```go
type ClientConfig struct {
    ServerAddress     string        `json:"server_address"`
    MaxConnections    int           `json:"max_connections"`    // 10
    RatePerConn       int           `json:"rate_per_connection"` // 15 pps
    ConnectionTimeout time.Duration `json:"connection_timeout"`
    HandshakeTimeout  time.Duration `json:"handshake_timeout"`
    KeepAlive         time.Duration `json:"keep_alive"`
    MaxStreams        int           `json:"max_streams_per_connection"`
}

// Connection pool for multiple connections
type ConnectionPool struct {
    connections []*quic.Connection
    current     int
    mu          sync.RWMutex
}

func (cp *ConnectionPool) GetConnection() *quic.Connection {
    cp.mu.Lock()
    defer cp.mu.Unlock()
    
    conn := cp.connections[cp.current]
    cp.current = (cp.current + 1) % len(cp.connections)
    return conn
}
```

### 5. Rate Limiting Implementation

#### Per-Connection Rate Limiting
```go
type RateLimiter struct {
    limit     rate.Limit
    burst     int
    limiters  map[string]*rate.Limiter
    mu        sync.RWMutex
}

func (rl *RateLimiter) Allow(connID string) bool {
    rl.mu.RLock()
    limiter, exists := rl.limiters[connID]
    rl.mu.RUnlock()
    
    if !exists {
        rl.mu.Lock()
        limiter = rate.NewLimiter(rl.limit, rl.burst)
        rl.limiters[connID] = limiter
        rl.mu.Unlock()
    }
    
    return limiter.Allow()
}
```

### 6. Monitoring Implementation

#### Client Metrics
```go
type ClientMetrics struct {
    ConnectionsTotal    prometheus.Counter
    ConnectionsActive   prometheus.Gauge
    PacketsSent        prometheus.Counter
    PacketsReceived    prometheus.Counter
    BytesSent          prometheus.Counter
    BytesReceived      prometheus.Counter
    LatencyHistogram   prometheus.Histogram
    JitterHistogram    prometheus.Histogram
    ErrorsTotal        prometheus.Counter
    RatePerConnection  prometheus.Gauge
}

func (cm *ClientMetrics) RecordLatency(latency time.Duration) {
    cm.LatencyHistogram.Observe(latency.Seconds())
}

func (cm *ClientMetrics) RecordJitter(jitter time.Duration) {
    cm.JitterHistogram.Observe(jitter.Seconds())
}

func (cm *ClientMetrics) RecordRate(rate float64) {
    cm.RatePerConnection.Set(rate)
}
```

### 7. Health Check System

#### Client Health Monitoring
```go
type HealthChecker struct {
    serverURL     string
    interval      time.Duration
    timeout       time.Duration
    metrics       *ClientMetrics
    lastCheck     time.Time
    healthy       bool
}

func (hc *HealthChecker) CheckHealth() error {
    ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
    defer cancel()
    
    // Check server connectivity
    conn, err := quic.DialAddr(ctx, hc.serverURL, &tls.Config{InsecureSkipVerify: true}, &quic.Config{})
    if err != nil {
        hc.healthy = false
        return err
    }
    defer conn.CloseWithError(0, "")
    
    hc.healthy = true
    hc.lastCheck = time.Now()
    return nil
}
```

## Deployment Scripts

### 1. Client Setup Script
```bash
#!/bin/bash
# client-setup.sh

echo "🔧 Setting up QUIC client environment..."

# Apply system optimizations
echo "📡 Applying network optimizations..."
sudo sysctl -w net.core.rmem_max=134217728
sudo sysctl -w net.core.wmem_max=134217728
sudo sysctl -w net.core.netdev_max_backlog=5000
sudo sysctl -w net.ipv4.tcp_congestion_control=bbr

# Set process limits
ulimit -n 65536
ulimit -u 32768

# Create client configuration
cat > client-config.yaml << EOF
client:
server:
  address: "212.233.79.160:9000"
    protocol: "quic"
    tls: false
  
  connections:
    max_connections: 10
    rate_per_connection: 15
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
    health_check_interval: 30s
EOF

echo "✅ Client setup complete!"
```

### 2. Client Test Script
```bash
#!/bin/bash
# client-test.sh

SERVER_IP=${QUIC_SERVER_IP:-"212.233.79.160"}
CONNECTIONS=${QUIC_CONNECTIONS:-10}
RATE=${QUIC_RATE:-15}
DURATION=${QUIC_DURATION:-60s}

echo "🚀 Starting QUIC client test..."
echo "  Server: $SERVER_IP:9000"
echo "  Connections: $CONNECTIONS"
echo "  Rate: $RATE pps (safe zone)"
echo "  Duration: $DURATION"

# Run client with optimized settings
./quic-client \
  --server="$SERVER_IP:9000" \
  --connections=$CONNECTIONS \
  --rate=$RATE \
  --duration=$DURATION \
  --no-tls \
  --prometheus \
  --monitoring

echo "✅ Client test complete!"
```

### 3. Monitoring Script
```bash
#!/bin/bash
# client-monitor.sh

CLIENT_METRICS_URL="http://localhost:2112/metrics"
SERVER_METRICS_URL="http://212.233.79.160:2113/metrics"

echo "📊 Client Monitoring Dashboard"
echo "=============================="

while true; do
    echo "$(date): Client Status"
    
    # Check client metrics
    if curl -s $CLIENT_METRICS_URL >/dev/null 2>&1; then
        CONNECTIONS=$(curl -s $CLIENT_METRICS_URL | grep 'quic_client_connections_active' | awk '{print $2}')
        RATE=$(curl -s $CLIENT_METRICS_URL | grep 'quic_client_rate_per_connection' | awk '{print $2}')
        LATENCY=$(curl -s $CLIENT_METRICS_URL | grep 'quic_client_latency_seconds' | awk '{print $2}')
        
        echo "  Connections: $CONNECTIONS"
        echo "  Rate: $RATE pps"
        echo "  Latency: $LATENCY seconds"
        
        # Check for critical zone
        if (( $(echo "$RATE >= 26 && $RATE <= 35" | bc -l 2>/dev/null || echo "0") )); then
            echo "  🚨 WARNING: Entering critical zone ($RATE pps)"
        else
            echo "  ✅ Rate in safe zone"
        fi
    else
        echo "  ❌ Client metrics unavailable"
    fi
    
    # Check server metrics
    if curl -s $SERVER_METRICS_URL >/dev/null 2>&1; then
        SERVER_CONNECTIONS=$(curl -s $SERVER_METRICS_URL | grep 'quic_server_connections_total' | awk '{print $2}')
        SERVER_RATE=$(curl -s $SERVER_METRICS_URL | grep 'quic_server_rate_per_connection' | awk '{print $2}')
        
        echo "  Server Connections: $SERVER_CONNECTIONS"
        echo "  Server Rate: $SERVER_RATE pps"
    else
        echo "  ❌ Server metrics unavailable"
    fi
    
    echo "  ---"
    sleep 10
done
```

## Docker Deployment

### 1. Client Dockerfile
```dockerfile
# Dockerfile.client
FROM golang:1.25.1-alpine AS builder

RUN apk add --no-cache git make
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build-client

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget
RUN addgroup -g 1001 -S quic && \
    adduser -u 1001 -S quic -G quic

WORKDIR /app
COPY --from=builder /app/build/quic-client ./
COPY --from=builder /app/tag.txt ./

RUN chown -R quic:quic /app
USER quic

EXPOSE 2112

ENV QUIC_CLIENT_ADDR="212.233.79.160:9000"
ENV QUIC_CONNECTIONS=10
ENV QUIC_RATE=15
ENV QUIC_DURATION=60s
ENV QUIC_PROMETHEUS_CLIENT_PORT=2112
ENV QUIC_NO_TLS=true

CMD ./quic-client \
  --server=${QUIC_CLIENT_ADDR} \
  --connections=${QUIC_CONNECTIONS} \
  --rate=${QUIC_RATE} \
  --duration=${QUIC_DURATION} \
  --no-tls \
  --prometheus
```

### 2. Docker Compose for Client
```yaml
# docker-compose.client.yml
version: '3.8'

services:
  quic-client:
    build:
      context: .
      dockerfile: Dockerfile.client
    container_name: 2gc-network-client
    ports:
      - "2112:2112"
    environment:
      - QUIC_CLIENT_ADDR=212.233.79.160:9000
      - QUIC_CONNECTIONS=10
      - QUIC_RATE=15
      - QUIC_DURATION=60s
      - QUIC_PROMETHEUS_CLIENT_PORT=2112
      - QUIC_NO_TLS=true
    restart: unless-stopped
    networks:
      - 2gc-client-network

  prometheus-client:
    image: prom/prometheus:latest
    container_name: 2gc-client-prometheus
    ports:
      - "9091:9090"
    volumes:
      - ./prometheus-client.yml:/etc/prometheus/prometheus.yml
    networks:
      - 2gc-client-network

  grafana-client:
    image: grafana/grafana:latest
    container_name: 2gc-client-grafana
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-client-data:/var/lib/grafana
    networks:
      - 2gc-client-network

networks:
  2gc-client-network:
    driver: bridge

volumes:
  grafana-client-data:
```

## Performance Optimization

### 1. Connection Pooling
```go
type OptimizedClient struct {
    connectionPool *ConnectionPool
    rateLimiters   map[string]*rate.Limiter
    metrics        *ClientMetrics
    config         *ClientConfig
}

func (oc *OptimizedClient) Connect() error {
    // Create multiple connections for load distribution
    for i := 0; i < oc.config.MaxConnections; i++ {
        conn, err := oc.createConnection()
        if err != nil {
            return err
        }
        oc.connectionPool.AddConnection(conn)
    }
    return nil
}
```

### 2. Stream Management
```go
type StreamManager struct {
    streams    map[string]*quic.Stream
    mu         sync.RWMutex
    maxStreams int
}

func (sm *StreamManager) GetStream(connID string) (*quic.Stream, error) {
    sm.mu.RLock()
    if stream, exists := sm.streams[connID]; exists {
        sm.mu.RUnlock()
        return stream, nil
    }
    sm.mu.RUnlock()
    
    // Create new stream
    stream, err := sm.createStream(connID)
    if err != nil {
        return nil, err
    }
    
    sm.mu.Lock()
    sm.streams[connID] = stream
    sm.mu.Unlock()
    
    return stream, nil
}
```

## Security Considerations

### 1. Authentication
```go
type AuthManager struct {
    token    string
    certFile string
    keyFile  string
}

func (am *AuthManager) Authenticate() error {
    // Implement authentication logic
    // Token-based, certificate-based, or OAuth2
    return nil
}
```

### 2. Encryption
```go
type EncryptionManager struct {
    tlsConfig *tls.Config
    certPool  *x509.CertPool
}

func (em *EncryptionManager) SetupTLS() error {
    // Configure TLS for secure connections
    // Certificate validation, cipher suites, etc.
    return nil
}
```

## Troubleshooting

### 1. Common Issues

#### Connection Timeouts
```bash
# Check network connectivity
ping 212.233.79.160
telnet 212.233.79.160 9000

# Check UDP connectivity
nc -u 212.233.79.160 9000
```

#### High Latency
```bash
# Check network path
traceroute 212.233.79.160

# Check for packet loss
ping -c 100 212.233.79.160
```

#### Rate Limiting Issues
```bash
# Check client rate
curl http://localhost:2112/metrics | grep rate

# Check server rate
curl http://212.233.79.160:2113/metrics | grep rate
```

### 2. Performance Tuning

#### Buffer Sizes
```bash
# Increase UDP buffers
sudo sysctl -w net.core.rmem_max=268435456
sudo sysctl -w net.core.wmem_max=268435456
```

#### Connection Limits
```bash
# Increase connection limits
ulimit -n 131072
ulimit -u 65536
```

## Monitoring and Alerting

### 1. Prometheus Alerts
```yaml
# prometheus-client-alerts.yml
groups:
- name: quic-client
  rules:
  - alert: QUICClientHighLatency
    expr: histogram_quantile(0.95, quic_client_latency_seconds) > 0.1
    for: 30s
    labels:
      severity: warning
    annotations:
      summary: "QUIC client high latency detected"
      description: "Client latency p95 is {{ $value }}s"
  
  - alert: QUICClientConnectionErrors
    expr: rate(quic_client_errors_total[5m]) > 0.01
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "QUIC client high error rate"
      description: "Client error rate is {{ $value }}"
```

### 2. Grafana Dashboard
```json
{
  "dashboard": {
    "title": "QUIC Client Dashboard",
    "panels": [
      {
        "title": "Active Connections",
        "type": "graph",
        "targets": [
          {
            "expr": "quic_client_connections_active",
            "legendFormat": "Active Connections"
          }
        ]
      },
      {
        "title": "Rate per Connection",
        "type": "graph",
        "targets": [
          {
            "expr": "quic_client_rate_per_connection",
            "legendFormat": "Rate (pps)"
          }
        ],
        "thresholds": [
          {
            "value": 26,
            "colorMode": "critical",
            "op": "gt"
          }
        ]
      }
    ]
  }
}
```

## Conclusion

Client-side DevOps implementation should focus on:

1. **Network Optimization**: UDP buffer tuning and congestion control
2. **Connection Management**: Multiple connections with rate limiting
3. **Monitoring**: Comprehensive metrics and alerting
4. **Security**: Authentication and encryption
5. **Performance**: Stream multiplexing and connection pooling

The key is to maintain connections in the safe zone (15-20 pps) while using multiple connections for high throughput requirements.

---

**Document Classification**: Client Implementation Guide  
**Distribution**: Client DevOps Team  
**Next Review**: November 2025
