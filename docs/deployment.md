# QUIC Test Deployment Guide

## Overview

This guide covers various deployment options for the QUIC Test tool, from simple local development to production-ready containerized deployments.

## Prerequisites

- Go 1.25 or later
- Docker and Docker Compose (for containerized deployment)
- Make (for build automation)
- Git (for source code management)

## Local Development

### Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/your-org/quic-test.git
   cd quic-test
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Build the project**
   ```bash
   make build
   ```

4. **Run tests**
   ```bash
   make test
   ```

5. **Start the dashboard**
   ```bash
   make run-dashboard
   ```

### Development Setup

1. **Install development tools**
   ```bash
   make dev-setup
   ```

2. **Run linting**
   ```bash
   make lint
   ```

3. **Run security checks**
   ```bash
   make vuln
   ```

## Docker Deployment

### Single Container

1. **Build the Docker image**
   ```bash
   docker build -t quic-test:latest .
   ```

2. **Run the container**
   ```bash
   docker run -p 9990:9990 -p 9000:9000 quic-test:latest
   ```

### Docker Compose

1. **Start all services**
   ```bash
   docker-compose up -d
   ```

2. **View logs**
   ```bash
   docker-compose logs -f
   ```

3. **Stop services**
   ```bash
   docker-compose down
   ```

### Services Included

- **quic-test**: Main application
- **prometheus**: Metrics collection
- **grafana**: Metrics visualization
- **jaeger**: Distributed tracing (optional)

## Production Deployment

### Kubernetes

1. **Create namespace**
   ```bash
   kubectl create namespace quic-test
   ```

2. **Deploy application**
   ```bash
   kubectl apply -f k8s/
   ```

3. **Check deployment**
   ```bash
   kubectl get pods -n quic-test
   ```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `QUIC_SERVER_ADDR` | QUIC server address | `:9000` |
| `QUIC_DASHBOARD_ADDR` | Dashboard address | `:9990` |
| `QUIC_PROMETHEUS_CLIENT_PORT` | Prometheus client port | `2112` |
| `QUIC_PROMETHEUS_SERVER_PORT` | Prometheus server port | `2113` |
| `QUIC_PPROF_ADDR` | Profiling address | `:6060` |

### Resource Requirements

#### Minimum Requirements

- **CPU**: 1 core
- **Memory**: 512 MB
- **Storage**: 1 GB

#### Recommended Requirements

- **CPU**: 2 cores
- **Memory**: 2 GB
- **Storage**: 10 GB

#### High-Performance Requirements

- **CPU**: 4+ cores
- **Memory**: 8+ GB
- **Storage**: 50+ GB SSD

## Configuration

### Test Configuration

Create a configuration file `config.yaml`:

```yaml
server:
  addr: ":9000"
  prometheus: true
  pprof_addr: ":6060"

client:
  connections: 2
  streams: 4
  packet_size: 1200
  rate: 100
  duration: "30s"

dashboard:
  addr: ":9990"
  static_path: "./static"

metrics:
  prometheus:
    enabled: true
    port: 2112
  grafana:
    enabled: true
    port: 3000

network:
  emulate_loss: 0.01
  emulate_latency: "10ms"
  emulate_dup: 0.005
```

### Network Profiles

Configure network profiles in `profiles.yaml`:

```yaml
profiles:
  wifi:
    rtt: "20ms"
    jitter: "5ms"
    loss: 0.02
    bandwidth: 1000
    duplication: 0.01

  lte:
    rtt: "50ms"
    jitter: "15ms"
    loss: 0.05
    bandwidth: 2000
    duplication: 0.02

  datacenter:
    rtt: "1ms"
    jitter: "0.1ms"
    loss: 0.0001
    bandwidth: 100000
    duplication: 0.0001
```

## Monitoring and Observability

### Prometheus Configuration

Create `prometheus/prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'quic-test'
    static_configs:
      - targets: ['quic-test:2112', 'quic-test:2113']
    scrape_interval: 5s
    metrics_path: /metrics

  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
```

### Grafana Dashboards

Import the provided dashboard:

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d @grafana/dashboards/quic-test.json \
  http://admin:admin@localhost:3000/api/dashboards/db
```

### Logging Configuration

Configure structured logging:

```yaml
logging:
  level: "info"
  format: "json"
  output: "stdout"
  
  fields:
    service: "quic-test"
    version: "1.0.0"
    environment: "production"
```

## Security Considerations

### Network Security

1. **Firewall Rules**
   ```bash
   # Allow QUIC traffic
   ufw allow 9000/udp
   
   # Allow dashboard access
   ufw allow 9990/tcp
   
   # Allow Prometheus metrics
   ufw allow 2112/tcp
   ufw allow 2113/tcp
   ```

2. **TLS Configuration**
   ```yaml
   tls:
     enabled: true
     cert_file: "/etc/ssl/certs/quic-test.crt"
     key_file: "/etc/ssl/private/quic-test.key"
   ```

### Authentication

1. **API Authentication**
   ```yaml
   auth:
     enabled: true
     type: "bearer"
     token: "your-secret-token"
   ```

2. **Dashboard Authentication**
   ```yaml
   dashboard:
     auth:
       enabled: true
       username: "admin"
       password: "secure-password"
   ```

## Performance Tuning

### System Tuning

1. **Network Buffer Sizes**
   ```bash
   echo 'net.core.rmem_max = 134217728' >> /etc/sysctl.conf
   echo 'net.core.wmem_max = 134217728' >> /etc/sysctl.conf
   sysctl -p
   ```

2. **File Descriptor Limits**
   ```bash
   echo '* soft nofile 65536' >> /etc/security/limits.conf
   echo '* hard nofile 65536' >> /etc/security/limits.conf
   ```

### Application Tuning

1. **Goroutine Limits**
   ```go
   runtime.GOMAXPROCS(runtime.NumCPU())
   ```

2. **Memory Management**
   ```go
   runtime.GC()
   ```

## Troubleshooting

### Common Issues

1. **Port Already in Use**
   ```bash
   # Find process using port
   lsof -i :9000
   
   # Kill process
   kill -9 <PID>
   ```

2. **Permission Denied**
   ```bash
   # Check file permissions
   ls -la /path/to/quic-test
   
   # Fix permissions
   chmod +x /path/to/quic-test
   ```

3. **Memory Issues**
   ```bash
   # Check memory usage
   free -h
   
   # Monitor memory
   top -p $(pgrep quic-test)
   ```

### Debug Mode

Enable debug logging:

```bash
export LOG_LEVEL=debug
./quic-test --mode=test --debug
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

## Backup and Recovery

### Configuration Backup

```bash
# Backup configuration
tar -czf quic-test-config-$(date +%Y%m%d).tar.gz config/ profiles/

# Restore configuration
tar -xzf quic-test-config-20240115.tar.gz
```

### Data Backup

```bash
# Backup metrics data
docker exec quic-test-prometheus tar -czf /backup/prometheus-$(date +%Y%m%d).tar.gz /prometheus

# Backup Grafana data
docker exec quic-test-grafana tar -czf /backup/grafana-$(date +%Y%m%d).tar.gz /var/lib/grafana
```

## Updates and Maintenance

### Rolling Updates

1. **Update application**
   ```bash
   docker-compose pull
   docker-compose up -d
   ```

2. **Verify deployment**
   ```bash
   docker-compose ps
   curl http://localhost:9990/status
   ```

### Health Checks

```bash
# Application health
curl http://localhost:9990/status

# Metrics health
curl http://localhost:2112/metrics

# Dashboard health
curl http://localhost:9990/
```

## Support and Maintenance

### Log Analysis

```bash
# Application logs
docker-compose logs quic-test

# System logs
journalctl -u quic-test

# Error logs
grep -i error /var/log/quic-test.log
```

### Performance Monitoring

```bash
# CPU usage
top -p $(pgrep quic-test)

# Memory usage
ps aux | grep quic-test

# Network usage
netstat -i
```

### Maintenance Schedule

- **Daily**: Check logs for errors
- **Weekly**: Review performance metrics
- **Monthly**: Update dependencies
- **Quarterly**: Security audit
