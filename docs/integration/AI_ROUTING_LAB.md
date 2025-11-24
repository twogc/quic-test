# AI Routing Lab Integration Guide

## Overview

This document provides comprehensive instructions for integrating **quic-test** with **AI Routing Lab** to enable machine learning-powered route prediction and optimization.

## Architecture

### Data Flow

```
quic-test (Go)
    ↓       ↑
Prometheus  Predictions
Metrics     (Feedback Loop)
    ↓       ↑
AI Routing Lab Collector
    ↓       ↑
Metrics Database (JSON/CSV)
    ↓       ↑
ML Models (LSTM, XGBoost, etc.)
    ↓       ↑
Predictions API (FastAPI)
```

### System Components

| Component | Language | Purpose |
|-----------|----------|---------|
| quic-test | Go | QUIC protocol testing, metrics generation, **AI consumer** |
| AI Routing Lab Collector | Python | Ingests Prometheus metrics from quic-test |
| ML Models | Python (TensorFlow/PyTorch) | Trains and serves latency/jitter predictions |
| CloudBridge Relay | Go | Uses ML predictions for intelligent routing |

## AI Consumer Integration

`quic-test` now includes a built-in AI consumer that closes the loop:

1.  **Metrics Collection**: `quic-test` collects real-time metrics (RTT, Jitter, Loss).
2.  **Prediction Request**: Periodically sends these metrics to the AI Routing Lab Inference API.
3.  **Route Optimization**: Receives predictions (e.g., "High Latency Expected") and simulates route switching logic.

To enable this feature, use the `--ai-enabled` flag.

## Quick Start

### Step 1: Start quic-test with Prometheus Export

```bash
# Terminal 1: Start quic-test server with metrics and AI enabled
cd quic-test
./bin/quic-test --mode=server --addr=:9000 --prometheus-port 9090 --ai-enabled --ai-service-url http://localhost:5000
```

**Verify Prometheus endpoint:**
```bash
curl http://localhost:9090/metrics | head -20
```

Expected output includes metrics like:
```
quic_latency_ms{connection_id="0"} 25.5
quic_jitter_ms{connection_id="0"} 2.3
quic_throughput_mbps{connection_id="0"} 150.2
quic_packet_loss_rate{connection_id="0"} 0.001
quic_rtt_ms{connection_id="0"} 25.5
```

### Step 2: Configure AI Routing Lab

```bash
# Terminal 2: Clone and setup AI Routing Lab
cd ../ai-routing-lab
pip install -r requirements.txt

# Configure collector for quic-test metrics
cat > config/collectors/quic_test.yaml << 'EOF'
collector:
  type: prometheus
  name: quic_test
  prometheus_url: http://localhost:9090
  scrape_interval: 1  # seconds
  metrics:
    - quic_latency_ms
    - quic_jitter_ms
    - quic_throughput_mbps
    - quic_packet_loss_rate
    - quic_rtt_ms
    - quic_retransmits
    - quic_congestion_window
EOF
```

### Step 3: Start Data Collection

```bash
# Collect metrics for training dataset (run for 5-10 minutes)
python -m data.collectors.quic_test_collector \
  --config config/collectors/quic_test.yaml \
  --output-file data/raw/quic_test_metrics.json \
  --duration 600  # 10 minutes
```

### Step 4: Train ML Model

```bash
# Train LSTM predictor
python -m models.train_predictor \
  --input-file data/raw/quic_test_metrics.json \
  --model-type lstm \
  --output-model models/quic_lstm_predictor.pkl \
  --test-split 0.2 \
  --epochs 50 \
  --batch-size 32
```

### Step 5: Validate Predictions

```bash
# Generate test data
./bin/quic-test --mode=test --network-profile=mobile --duration=120s --report=test_metrics.json

# Validate model predictions
python -m models.validate_predictor \
  --model-file models/quic_lstm_predictor.pkl \
  --test-file data/raw/quic_test_metrics.json \
  --output-report validation_report.json
```

## Metrics Specification

### Available Metrics from quic-test

#### Performance Metrics

| Metric | Type | Unit | Description |
|--------|------|------|-------------|
| quic_latency_ms | Gauge | milliseconds | One-way latency |
| quic_rtt_ms | Gauge | milliseconds | Round-trip time |
| quic_jitter_ms | Gauge | milliseconds | RTT variation |
| quic_throughput_mbps | Gauge | Mbps | Goodput (useful data rate) |
| quic_packet_loss_rate | Gauge | fraction | Lost packets / total packets |

#### Reliability Metrics

| Metric | Type | Unit | Description |
|--------|------|------|-------------|
| quic_retransmits | Counter | count | Number of retransmitted packets |
| quic_lost_packets | Counter | count | Unrecovered lost packets |
| quic_duplicate_packets | Counter | count | Duplicate packets received |
| quic_reordered_packets | Counter | count | Out-of-order packets |

#### Congestion Control Metrics

| Metric | Type | Unit | Description |
|--------|------|------|-------------|
| quic_congestion_window | Gauge | bytes | QUIC congestion window size |
| quic_bytes_in_flight | Gauge | bytes | Unacknowledged data in network |
| quic_slow_start_threshold | Gauge | bytes | SSThresh value |
| quic_cc_recovery_mode | Gauge | enum | Current recovery state |

#### Connection Metrics

| Metric | Type | Unit | Description |
|--------|------|------|-------------|
| quic_connections_active | Gauge | count | Active QUIC connections |
| quic_connections_established | Counter | count | Total connections established |
| quic_handshake_time_ms | Histogram | milliseconds | Initial handshake duration |
| quic_idle_timeout_ms | Gauge | milliseconds | Idle timeout setting |

### Metric Export Format

#### Prometheus Text Format
```
# HELP quic_latency_ms Latency in milliseconds
# TYPE quic_latency_ms gauge
quic_latency_ms{connection_id="conn_001",remote_addr="127.0.0.1:5000"} 25.5
quic_latency_ms{connection_id="conn_002",remote_addr="127.0.0.1:5001"} 23.1

# HELP quic_jitter_ms Jitter in milliseconds
# TYPE quic_jitter_ms gauge
quic_jitter_ms{connection_id="conn_001"} 2.3
quic_jitter_ms{connection_id="conn_002"} 1.8
```

#### JSON Export Format
```json
{
  "timestamp": 1700000000,
  "metrics": [
    {
      "name": "quic_latency_ms",
      "connection_id": "conn_001",
      "value": 25.5,
      "labels": {
        "remote_addr": "127.0.0.1:5000",
        "protocol": "quic",
        "tls_version": "1.3"
      }
    },
    {
      "name": "quic_jitter_ms",
      "connection_id": "conn_001",
      "value": 2.3,
      "labels": {
        "remote_addr": "127.0.0.1:5000"
      }
    }
  ]
}
```

## ML Model Integration

### Supported Model Types

#### LSTM (Long Short-Term Memory)

**Best for:** Time-series prediction with temporal dependencies

```bash
python -m models.train_predictor \
  --input-file data/raw/quic_test_metrics.json \
  --model-type lstm \
  --lookback-window 30 \
  --output-model models/quic_lstm.pkl
```

**Configuration:**
```yaml
model:
  type: lstm
  layers: 2
  hidden_units: 64
  dropout: 0.2
  activation: relu
  optimizer: adam
  loss: mse
```

#### XGBoost

**Best for:** Fast training and good accuracy with mixed feature types

```bash
python -m models.train_predictor \
  --input-file data/raw/quic_test_metrics.json \
  --model-type xgboost \
  --output-model models/quic_xgboost.pkl
```

**Configuration:**
```yaml
model:
  type: xgboost
  n_estimators: 100
  max_depth: 6
  learning_rate: 0.1
  subsample: 0.8
  colsample_bytree: 0.8
```

#### Gradient Boosting

**Best for:** Accurate predictions with feature importance analysis

```bash
python -m models.train_predictor \
  --input-file data/raw/quic_test_metrics.json \
  --model-type gradient_boosting \
  --output-model models/quic_gb.pkl
```

### Training Data Requirements

**Minimum dataset size:** 1000 samples (approximately 15-20 minutes of continuous testing)

**Optimal dataset size:** 10,000+ samples (2-3 hours of diverse network conditions)

**Data collection strategy:**
```bash
# Collect under various network conditions
for profile in excellent good poor mobile satellite adversarial; do
  echo "Collecting metrics for $profile network..."
  ./bin/quic-test --mode=test \
    --network-profile=$profile \
    --duration=300 \
    --report=metrics_${profile}.json
done

# Combine datasets
python -m data.combine_datasets \
  --input metrics_*.json \
  --output data/combined_metrics.json
```

## Advanced Configuration

### Multi-Path Prediction

```bash
# Train separate models for different network paths
python -m models.train_multi_path_predictor \
  --input-file data/combined_metrics.json \
  --output-dir models/multi_path/ \
  --paths 4 \
  --model-type lstm
```

### Real-time Model Serving

```bash
# Start model serving API
python -m api.predictor_service \
  --model-file models/quic_lstm_predictor.pkl \
  --port 5000 \
  --workers 4
```

**API Endpoint Example:**

```bash
curl -X POST http://localhost:5000/predict \
  -H "Content-Type: application/json" \
  -d '{
    "metrics": {
      "recent_latency": [25.1, 25.5, 24.8, 25.2],
      "recent_loss": [0.001, 0.0, 0.001, 0.0],
      "connection_duration": 300,
      "packet_count": 10000
    }
  }' | jq .
```

**Response:**
```json
{
  "predictions": {
    "predicted_latency_ms": 25.3,
    "confidence": 0.92,
    "latency_range": {
      "min": 23.5,
      "max": 27.1
    }
  },
  "processing_time_ms": 2.3
}
```

### CloudBridge Relay Integration

Once models are trained and serving, integrate with CloudBridge Relay:

```bash
# Configure relay to use ML predictor
cat > /etc/cloudbridge/relay.yaml << 'EOF'
routing:
  mode: ml_enhanced
  predictor:
    service: http://localhost:5000
    timeout: 100ms
    fallback: latency_first
  paths:
    - name: path_0
      predictor_weight: 0.3
    - name: path_1
      predictor_weight: 0.3
    - name: path_2
      predictor_weight: 0.4
EOF

# Start relay with ML routing
./cloudbridge-relay --config /etc/cloudbridge/relay.yaml
```

## Troubleshooting

### Issue: No metrics appear in Prometheus

**Check 1:** Verify quic-test is running with --prometheus-port flag
```bash
ps aux | grep quic-test
```

**Check 2:** Verify port is accessible
```bash
curl http://localhost:9090/metrics
```

**Check 3:** Check quic-test logs
```bash
# Add debug logging
./bin/quic-test --mode=server --prometheus-port 9090 --debug
```

### Issue: Model accuracy is low

**Check 1:** Increase training dataset size
```bash
# Collect more data (minimum 30 minutes)
./bin/quic-test --mode=test --duration=1800 --report=training_data.json
```

**Check 2:** Try different model type
```bash
# XGBoost often works better with diverse network conditions
python -m models.train_predictor \
  --input-file data/raw/quic_test_metrics.json \
  --model-type xgboost \
  --output-model models/quic_xgboost.pkl
```

**Check 3:** Validate feature engineering
```bash
python -m data.analyze_features \
  --input-file data/raw/quic_test_metrics.json \
  --output-report feature_analysis.html
```

### Issue: Collector fails to connect to Prometheus

**Solution:** Verify connectivity and authentication
```bash
# Test Prometheus endpoint manually
curl -v http://localhost:9090/api/v1/query?query=quic_latency_ms

# Check firewall
sudo netstat -tlnp | grep 9090

# Check quic-test binding
./bin/quic-test --mode=server --addr=:9000 --prometheus-port 9090 --debug
```

## Performance Considerations

### Metrics Collection Overhead

- CPU overhead: < 2% on modern systems
- Memory usage: ~50 MB per 10,000 metrics in memory
- Disk I/O: ~1 MB per 1000 metrics stored

### Model Inference Performance

| Model Type | Latency (ms) | CPU Usage | Memory |
|-----------|-------------|-----------|--------|
| LSTM | 5-10 | Low | 200-500 MB |
| XGBoost | 1-3 | Low | 100-300 MB |
| Gradient Boosting | 2-5 | Low | 150-400 MB |

### Scaling Guidelines

For production deployments:

- **Small scale** (1-10 paths): Single model instance, LSTM or XGBoost
- **Medium scale** (10-50 paths): Multi-model setup, separate models per region
- **Large scale** (50+ paths): Distributed serving with load balancer

## Validation Checklist

- [ ] quic-test running with Prometheus metrics export
- [ ] AI Routing Lab collector successfully connecting to Prometheus
- [ ] Training dataset collected (minimum 1000 samples)
- [ ] ML model trained without errors
- [ ] Model validation metrics acceptable (R² > 0.85 for latency prediction)
- [ ] Model serving API operational
- [ ] CloudBridge Relay configured for ML-enhanced routing
- [ ] Predictions improving route selection quality over baseline

## References

- [AI Routing Lab Repository](https://github.com/twogc/ai-routing-lab)
- [quic-test Repository](https://github.com/twogc/quic-test)
- [CloudBridge Relay Documentation](https://github.com/twogc/cloudbridge-scalable-relay)
- [Prometheus Metrics Format](https://prometheus.io/docs/instrumenting/exposition_formats/)
- [LSTM for Time Series](https://colah.github.io/posts/2015-08-Understanding-LSTMs/)

## Support

For issues or questions:
- GitHub Issues: https://github.com/twogc/quic-test/issues
- Email: labs@cloudbridge.io
- Slack: #ai-routing-support (CloudBridge members)

---

**Last Updated:** November 2025
**Version:** 1.0
**Status:** Production Ready
