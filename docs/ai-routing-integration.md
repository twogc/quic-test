# AI Routing Lab Integration

Integration between `quic-test` and [AI Routing Lab](https://github.com/twogc/ai-routing-lab).

## Overview

`quic-test` provides network performance metrics that AI Routing Lab uses to train machine learning models for predicting optimal network routes.

```
┌──────────────┐         ┌───────────────┐         ┌─────────────────┐
│              │         │               │         │                 │
│  quic-test   ├────────►│  Prometheus   ├────────►│  AI Routing Lab │
│              │ metrics │               │ scrape  │                 │
└──────────────┘         └───────────────┘         └─────────────────┘
                                                            │
                                                            ▼
                                                    ┌───────────────┐
                                                    │  ML Models    │
                                                    │  (Route Pred) │
                                                    └───────────────┘
```

## Quick Start

### 1. Start quic-test with Prometheus export

```bash
# Server mode
docker run -p 4433:4433/udp -p 9090:9090 mlanies/quic-test:latest \
  --mode=server \
  --prometheus-port=9090

# Client mode (continuous testing)
docker run mlanies/quic-test:latest \
  --mode=client \
  --server=<target>:4433 \
  --duration=0 \
  --prometheus-port=9091
```

### 2. Configure AI Routing Lab

```yaml
# ai-routing-lab/config.yaml
data_sources:
  - type: prometheus
    url: http://localhost:9090
    metrics:
      - quic_rtt_seconds
      - quic_jitter_seconds
      - quic_throughput_bytes_per_second
      - quic_packet_loss_ratio
    scrape_interval: 5s
```

### 3. Train ML model

```bash
cd ai-routing-lab
python train.py --config=config.yaml --model=route_predictor
```

## Exported Metrics

### Core Metrics

```prometheus
# RTT (Round-Trip Time)
quic_rtt_seconds{quantile="0.5"}
quic_rtt_seconds{quantile="0.95"}
quic_rtt_seconds{quantile="0.99"}

# Jitter
quic_jitter_seconds

# Throughput
quic_throughput_bytes_per_second

# Packet Loss
quic_packet_loss_ratio

# Connection Stats
quic_connections_total
quic_streams_total
quic_handshake_duration_seconds
```

### Advanced Metrics

```prometheus
# Congestion Control
quic_congestion_window_bytes
quic_pacing_rate_bytes_per_second
quic_rtt_variance_seconds

# FEC (if enabled)
quic_fec_redundancy_ratio
quic_fec_recovered_packets_total
quic_fec_unrecoverable_packets_total
```

## Use Cases

### 1. Route Prediction

**Goal:** Predict which network route will have lowest latency.

**Data Collection:**
```bash
# Test multiple routes simultaneously
for route in route1 route2 route3; do
  docker run -d --name quic-test-$route \
    -p 909${route#route}:9090 \
    mlanies/quic-test:latest \
    --mode=client \
    --server=$route.example.com:4433 \
    --duration=0 \
    --prometheus-port=9090
done
```

**AI Routing Lab:**
```python
from ai_routing_lab import RoutePredictor

predictor = RoutePredictor()
predictor.train(
    prometheus_url="http://localhost:9090",
    features=["rtt", "jitter", "loss"],
    target="route_id"
)

# Predict best route
best_route = predictor.predict(current_conditions)
```

### 2. Anomaly Detection

**Goal:** Detect network anomalies in real-time.

**Data Collection:**
```bash
docker run -p 9090:9090 mlanies/quic-test:latest \
  --mode=server \
  --prometheus-port=9090
```

**AI Routing Lab:**
```python
from ai_routing_lab import AnomalyDetector

detector = AnomalyDetector()
detector.train(
    prometheus_url="http://localhost:9090",
    window_size="5m",
    threshold=0.95
)

# Detect anomalies
anomalies = detector.detect_realtime()
```

### 3. Congestion Control Optimization

**Goal:** Learn optimal BBR parameters for different network conditions.

**Data Collection:**
```bash
# Test BBRv2 with various parameters
for probe_bw in 0.25 0.5 0.75; do
  docker run mlanies/quic-test:latest \
    --mode=client \
    --congestion=bbrv2 \
    --bbr-probe-bw=$probe_bw \
    --prometheus-port=9090
done
```

**AI Routing Lab:**
```python
from ai_routing_lab import BBROptimizer

optimizer = BBROptimizer()
optimizer.train(
    prometheus_url="http://localhost:9090",
    parameters=["probe_bw", "probe_rtt_interval"],
    objective="throughput"
)

# Get optimal parameters
optimal_params = optimizer.optimize(network_conditions)
```

## Data Pipeline

### 1. Collection

```bash
# quic-test exports metrics
quic-test --mode=server --prometheus-port=9090
```

### 2. Storage

```yaml
# Prometheus config
scrape_configs:
  - job_name: 'quic-test'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 5s
```

### 3. Processing

```python
# AI Routing Lab collector
from ai_routing_lab.data import PrometheusCollector

collector = PrometheusCollector(
    url="http://localhost:9090",
    metrics=["quic_rtt_seconds", "quic_throughput_bytes_per_second"]
)

df = collector.collect(start_time="1h ago", end_time="now")
```

### 4. Training

```python
# Train model
from ai_routing_lab.models import RoutePredictor

model = RoutePredictor()
model.fit(df, target="route_latency")
model.save("route_predictor.pkl")
```

### 5. Inference

```python
# Real-time prediction
predictions = model.predict_realtime(
    prometheus_url="http://localhost:9090"
)
```

## Example: Complete Workflow

```bash
# 1. Start quic-test servers on multiple routes
docker-compose up -d

# 2. Collect data for 24 hours
sleep 86400

# 3. Train AI model
cd ai-routing-lab
python train.py --duration=24h --model=route_predictor

# 4. Deploy model
python deploy.py --model=route_predictor.pkl

# 5. Use predictions in production
curl http://localhost:8080/predict \
  -d '{"source": "client1", "destination": "server1"}'
```

## Performance

### Metrics Export Overhead

- **CPU:** <1% additional
- **Memory:** ~10 MB for histogram storage
- **Network:** ~1 KB/s per metric

### Scrape Interval

- **Recommended:** 5-10s
- **Minimum:** 1s (high overhead)
- **Maximum:** 60s (low resolution)

## Troubleshooting

### Metrics not appearing

```bash
# Check Prometheus endpoint
curl http://localhost:9090/metrics | grep quic_

# Check quic-test logs
docker logs <container-id>
```

### High cardinality

```bash
# Reduce label cardinality
quic-test --mode=server \
  --prometheus-port=9090 \
  --prometheus-labels=minimal
```

### Memory issues

```bash
# Reduce histogram buckets
quic-test --mode=server \
  --prometheus-port=9090 \
  --histogram-buckets=10
```

## See Also

- [AI Routing Lab Documentation](https://github.com/twogc/ai-routing-lab)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Architecture](architecture.md)
- [Case Studies](case-studies.md)
