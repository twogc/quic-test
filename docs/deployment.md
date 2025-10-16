# 2GC Network Protocol Suite - Deployment Guide

## Overview

This guide covers deployment options for the 2GC Network Protocol Suite, including standalone, distributed, and cloud-native deployments.

## Quick Start

### Prerequisites

- Go 1.25 or later
- Docker and Docker Compose (optional)
- Kubernetes cluster (for distributed deployment)

### Standalone Deployment

1. **Clone the repository**
```bash
git clone https://github.com/twogc/quic-test.git
cd quic-test
```

2. **Build the application**
```bash
go build -o quic-test main.go
```

3. **Run the server**
```bash
./quic-test --mode=server --addr=:9000
```

4. **Run the client**
```bash
./quic-test --mode=client --addr=localhost:9000
```

### Docker Deployment

1. **Build Docker image**
```bash
docker build -t quic-test .
```

2. **Run with Docker Compose**
```bash
docker-compose up -d
```

3. **Access the dashboard**
```bash
open http://localhost:9990
```

## Docker Configuration

### Dockerfile

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o quic-test main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/quic-test .
EXPOSE 9000 9990
CMD ["./quic-test", "--mode=server", "--addr=:9000"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  quic-test-server:
    build: .
    ports:
      - "9000:9000"
      - "9990:9990"
    environment:
      - QUIC_TEST_MODE=server
      - QUIC_TEST_ADDR=:9000
      - QUIC_TEST_PROMETHEUS=true
    volumes:
      - ./config:/app/config
      - ./reports:/app/reports
    command: ["./quic-test", "--mode=server", "--addr=:9000", "--prometheus"]

  quic-test-client:
    build: .
    depends_on:
      - quic-test-server
    environment:
      - QUIC_TEST_MODE=client
      - QUIC_TEST_ADDR=quic-test-server:9000
    command: ["./quic-test", "--mode=client", "--addr=quic-test-server:9000"]

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-storage:/var/lib/grafana
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./grafana/datasources:/etc/grafana/provisioning/datasources

volumes:
  grafana-storage:
```

## Kubernetes Deployment

### Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: quic-test
  labels:
    name: quic-test
```

### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: quic-test-config
  namespace: quic-test
data:
  config.yaml: |
    server:
      addr: ":9000"
      tls:
        enabled: true
        cert: "/etc/tls/tls.crt"
        key: "/etc/tls/tls.key"
    monitoring:
      prometheus:
        enabled: true
        port: 9990
    scenarios:
      - name: "wifi"
        rtt: "20ms"
        jitter: "5ms"
        loss: "0.1%"
        bandwidth: "100Mbps"
      - name: "lte"
        rtt: "50ms"
        jitter: "15ms"
        loss: "0.5%"
        bandwidth: "50Mbps"
```

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: quic-test-server
  namespace: quic-test
spec:
  replicas: 3
  selector:
    matchLabels:
      app: quic-test-server
  template:
    metadata:
      labels:
        app: quic-test-server
    spec:
      containers:
      - name: quic-test
        image: quic-test:latest
        ports:
        - containerPort: 9000
          name: quic
        - containerPort: 9990
          name: metrics
        env:
        - name: QUIC_TEST_MODE
          value: "server"
        - name: QUIC_TEST_ADDR
          value: ":9000"
        - name: QUIC_TEST_PROMETHEUS
          value: "true"
        volumeMounts:
        - name: config
          mountPath: /app/config
        - name: tls
          mountPath: /etc/tls
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 9990
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 9990
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: quic-test-config
      - name: tls
        secret:
          secretName: quic-test-tls
```

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: quic-test-server
  namespace: quic-test
spec:
  selector:
    app: quic-test-server
  ports:
  - name: quic
    port: 9000
    targetPort: 9000
    protocol: UDP
  - name: metrics
    port: 9990
    targetPort: 9990
    protocol: TCP
  type: ClusterIP
```

### Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: quic-test-ingress
  namespace: quic-test
  annotations:
    nginx.ingress.kubernetes.io/backend-protocol: "HTTP"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - quic-test.example.com
    secretName: quic-test-tls
  rules:
  - host: quic-test.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: quic-test-server
            port:
              number: 9990
```

## Production Deployment

### High Availability Setup

1. **Load Balancer Configuration**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: quic-test-lb
  namespace: quic-test
spec:
  type: LoadBalancer
  selector:
    app: quic-test-server
  ports:
  - name: quic
    port: 9000
    targetPort: 9000
    protocol: UDP
  - name: metrics
    port: 9990
    targetPort: 9990
    protocol: TCP
```

2. **Horizontal Pod Autoscaler**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: quic-test-hpa
  namespace: quic-test
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: quic-test-server
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Monitoring Setup

1. **Prometheus Configuration**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: quic-test
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
      evaluation_interval: 15s
    
    scrape_configs:
    - job_name: 'quic-test'
      static_configs:
      - targets: ['quic-test-server:9990']
      scrape_interval: 5s
      metrics_path: /metrics
```

2. **Grafana Dashboard**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboard
  namespace: quic-test
data:
  dashboard.json: |
    {
      "dashboard": {
        "title": "QUIC Test Dashboard",
        "panels": [
          {
            "title": "RTT p95",
            "type": "graph",
            "targets": [
              {
                "expr": "histogram_quantile(0.95, quic_rtt_seconds)",
                "legendFormat": "RTT p95"
              }
            ]
          }
        ]
      }
    }
```

### Security Configuration

1. **Network Policies**
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: quic-test-netpol
  namespace: quic-test
spec:
  podSelector:
    matchLabels:
      app: quic-test-server
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: quic-test
    ports:
    - protocol: UDP
      port: 9000
    - protocol: TCP
      port: 9990
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    ports:
    - protocol: TCP
      port: 9090
```

2. **Pod Security Policy**
```yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: quic-test-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
```

## Cloud Deployment

### AWS EKS

1. **Create EKS cluster**
```bash
eksctl create cluster --name quic-test --region us-west-2 --nodegroup-name workers --node-type t3.medium --nodes 3
```

2. **Deploy application**
```bash
kubectl apply -f k8s/
```

3. **Configure load balancer**
```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: quic-test-alb
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: nlb
spec:
  type: LoadBalancer
  selector:
    app: quic-test-server
  ports:
  - name: quic
    port: 9000
    targetPort: 9000
    protocol: UDP
EOF
```

### Google GKE

1. **Create GKE cluster**
```bash
gcloud container clusters create quic-test --zone us-central1-a --num-nodes 3
```

2. **Deploy application**
```bash
kubectl apply -f k8s/
```

3. **Configure ingress**
```bash
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: quic-test-ingress
  annotations:
    kubernetes.io/ingress.global-static-ip-name: quic-test-ip
spec:
  rules:
  - host: quic-test.example.com
    http:
      paths:
      - path: /*
        pathType: ImplementationSpecific
        backend:
          service:
            name: quic-test-server
            port:
              number: 9990
EOF
```

### Azure AKS

1. **Create AKS cluster**
```bash
az aks create --resource-group quic-test-rg --name quic-test --node-count 3 --enable-addons monitoring
```

2. **Deploy application**
```bash
kubectl apply -f k8s/
```

3. **Configure ingress**
```bash
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: quic-test-ingress
  annotations:
    kubernetes.io/ingress.class: nginx
spec:
  rules:
  - host: quic-test.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: quic-test-server
            port:
              number: 9990
EOF
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `QUIC_TEST_MODE` | Operation mode | `test` |
| `QUIC_TEST_ADDR` | Server address | `:9000` |
| `QUIC_TEST_PROMETHEUS` | Enable Prometheus | `false` |
| `QUIC_TEST_TLS` | Enable TLS | `true` |
| `QUIC_TEST_CERT` | TLS certificate path | `/etc/tls/tls.crt` |
| `QUIC_TEST_KEY` | TLS key path | `/etc/tls/tls.key` |
| `QUIC_TEST_LOG_LEVEL` | Log level | `info` |
| `QUIC_TEST_CONFIG` | Config file path | `/app/config/config.yaml` |

### Configuration File

```yaml
server:
  addr: ":9000"
  tls:
    enabled: true
    cert: "/etc/tls/tls.crt"
    key: "/etc/tls/tls.key"
  quic:
    max_idle_timeout: "30s"
    handshake_timeout: "10s"
    keep_alive: "30s"
    max_streams: 100
    max_stream_data: "1MB"
    enable_0rtt: true
    enable_key_update: true
    enable_datagrams: true

monitoring:
  prometheus:
    enabled: true
    port: 9990
    path: "/metrics"
  grafana:
    enabled: true
    port: 3000

scenarios:
  - name: "wifi"
    rtt: "20ms"
    jitter: "5ms"
    loss: "0.1%"
    bandwidth: "100Mbps"
  - name: "lte"
    rtt: "50ms"
    jitter: "15ms"
    loss: "0.5%"
    bandwidth: "50Mbps"

sla:
  rtt_p95: "50ms"
  loss_rate: "1%"
  throughput: "50Mbps"
  errors: 0
```

## Troubleshooting

### Common Issues

1. **Port conflicts**
```bash
# Check if ports are in use
netstat -tulpn | grep :9000
netstat -tulpn | grep :9990
```

2. **TLS certificate issues**
```bash
# Generate self-signed certificate
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
```

3. **Kubernetes deployment issues**
```bash
# Check pod status
kubectl get pods -n quic-test

# Check logs
kubectl logs -f deployment/quic-test-server -n quic-test

# Check events
kubectl get events -n quic-test
```

### Performance Tuning

1. **System limits**
```bash
# Increase file descriptor limits
ulimit -n 65536

# Increase network buffer sizes
echo 'net.core.rmem_max = 16777216' >> /etc/sysctl.conf
echo 'net.core.wmem_max = 16777216' >> /etc/sysctl.conf
sysctl -p
```

2. **Kubernetes resource limits**
```yaml
resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

### Monitoring and Alerting

1. **Prometheus alerts**
```yaml
groups:
- name: quic-test
  rules:
  - alert: HighRTT
    expr: histogram_quantile(0.95, quic_rtt_seconds) > 0.05
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High RTT detected"
      description: "RTT p95 is {{ $value }}s"
```

2. **Grafana dashboards**
```json
{
  "dashboard": {
    "title": "QUIC Test Performance",
    "panels": [
      {
        "title": "RTT Distribution",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, quic_rtt_seconds)",
            "legendFormat": "RTT p95"
          }
        ]
      }
    ]
  }
}
```