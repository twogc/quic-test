# 2GC Network Protocol Suite API Documentation

## Overview

The 2GC Network Protocol Suite API provides REST endpoints for managing network protocol performance tests (QUIC, MASQUE, ICE/STUN/TURN), monitoring metrics, and generating reports. The API is designed to be simple, RESTful, and easy to integrate with monitoring systems.

## Base URL

```
http://localhost:9990
```

## Authentication

Currently, the API does not require authentication. In production environments, consider implementing proper authentication mechanisms.

## Endpoints

### Status

Get the current status of the testing system.

**GET** `/status`

#### Response

```json
{
  "server": {
    "running": true
  },
  "client": {
    "running": false
  },
  "last_update": "2024-01-15T10:30:00Z"
}
```

#### Status Codes

- `200 OK` - Status retrieved successfully

---

### Run Test

Start a new network protocol performance test.

**POST** `/run-test`

#### Request Body

```json
{
  "mode": "test",
  "addr": ":9000",
  "connections": 2,
  "streams": 4,
  "duration": "30s",
  "packet_size": 1200,
  "rate": 100,
  "pattern": "random",
  "prometheus": true,
  "emulate_loss": 0.01,
  "emulate_latency": "10ms",
  "emulate_dup": 0.005
}
```

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `mode` | string | Yes | Test mode: "test", "server", "client" |
| `addr` | string | Yes | Address to connect to or listen on |
| `connections` | integer | No | Number of QUIC connections (default: 1) |
| `streams` | integer | No | Number of streams per connection (default: 1) |
| `duration` | string | No | Test duration (e.g., "30s", "5m") |
| `packet_size` | integer | No | Packet size in bytes (default: 1200) |
| `rate` | integer | No | Packets per second (default: 100) |
| `pattern` | string | No | Data pattern: "random", "zeroes", "increment" |
| `prometheus` | boolean | No | Enable Prometheus metrics export |
| `emulate_loss` | number | No | Packet loss probability (0-1) |
| `emulate_latency` | string | No | Additional latency (e.g., "10ms") |
| `emulate_dup` | number | No | Packet duplication probability (0-1) |

#### Response

```json
{
  "status": "started",
  "message": "Test started",
  "config": {
    "mode": "test",
    "addr": ":9000",
    "connections": 2,
    "streams": 4
  }
}
```

#### Status Codes

- `200 OK` - Test started successfully
- `400 Bad Request` - Invalid configuration
- `405 Method Not Allowed` - Wrong HTTP method

---

### Stop Test

Stop the currently running test.

**POST** `/stop-test`

#### Response

```json
{
  "status": "stopped",
  "message": "Test stopped"
}
```

#### Status Codes

- `200 OK` - Test stopped successfully
- `405 Method Not Allowed` - Wrong HTTP method

---

### Presets

Manage test presets (scenarios and network profiles).

#### Get Available Presets

**GET** `/presets`

#### Response

```json
{
  "scenarios": [
    "wifi",
    "lte",
    "sat",
    "dc-eu",
    "ru-eu",
    "loss-burst",
    "reorder"
  ],
  "profiles": [
    "wifi",
    "wifi-5g",
    "lte",
    "lte-advanced",
    "5g",
    "satellite",
    "satellite-leo",
    "ethernet",
    "ethernet-10g",
    "dsl",
    "cable",
    "fiber",
    "mobile-3g",
    "edge",
    "international",
    "datacenter"
  ]
}
```

#### Apply Preset

**POST** `/presets`

#### Request Body

```json
{
  "type": "scenario",
  "name": "wifi"
}
```

or

```json
{
  "type": "profile",
  "name": "lte"
}
```

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `type` | string | Yes | Preset type: "scenario" or "profile" |
| `name` | string | Yes | Preset name |

#### Response

```json
{
  "status": "applied",
  "message": "Preset wifi applied",
  "config": {
    "mode": "test",
    "addr": ":9000",
    "connections": 2,
    "streams": 4,
    "emulate_loss": 0.02,
    "emulate_latency": "10ms"
  }
}
```

#### Status Codes

- `200 OK` - Preset applied successfully
- `400 Bad Request` - Invalid preset type or name
- `405 Method Not Allowed` - Wrong HTTP method

---

### Metrics

Get current test metrics.

**GET** `/metrics`

#### Response

```json
{
  "Success": 100,
  "Errors": 5,
  "BytesSent": 1024000,
  "BytesReceived": 1024000,
  "LatencyAverage": 25.5,
  "ThroughputAverage": 1000.0,
  "PacketLoss": 0.01,
  "Latencies": [20.0, 25.0, 30.0, 35.0, 40.0],
  "TimeSeriesLatency": [
    {"Time": 1.0, "Value": 25.0},
    {"Time": 2.0, "Value": 26.0}
  ],
  "TimeSeriesThroughput": [
    {"Time": 1.0, "Value": 1000.0},
    {"Time": 2.0, "Value": 1050.0}
  ]
}
```

#### Status Codes

- `200 OK` - Metrics retrieved successfully

---

### Report

Generate test reports in various formats.

**GET** `/report?format={format}`

#### Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `format` | string | No | Report format: "json", "csv", "md" (default: "json") |

#### Response Formats

##### JSON Format (default)

```json
{
  "config": {
    "mode": "test",
    "addr": ":9000",
    "connections": 2,
    "streams": 4
  },
  "metrics": {
    "Success": 100,
    "Errors": 5,
    "LatencyAverage": 25.5
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

##### CSV Format

```csv
Parameter,Value
Mode,test
Address,:9000
Connections,2
Streams,4
Success,100
Errors,5
Average Latency (ms),25.5
```

##### Markdown Format

```markdown
# QUIC Test Report

**Generated:** 2024-01-15 10:30:00

## Test Configuration

| Parameter | Value |
|-----------|-------|
| Mode | test |
| Address | :9000 |
| Connections | 2 |
| Streams | 4 |

## Test Results

| Metric | Value |
|--------|-------|
| Successful Requests | 100 |
| Errors | 5 |
| Average Latency | 25.5 ms |
```

#### Status Codes

- `200 OK` - Report generated successfully
- `400 Bad Request` - Unsupported format

---

## Error Responses

All endpoints may return error responses in the following format:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": "Additional error details"
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| `INVALID_CONFIG` | Invalid test configuration |
| `TEST_ALREADY_RUNNING` | Test is already running |
| `TEST_NOT_RUNNING` | No test is currently running |
| `INVALID_PRESET` | Invalid preset type or name |
| `UNSUPPORTED_FORMAT` | Unsupported report format |

---

## Rate Limiting

Currently, the API does not implement rate limiting. In production environments, consider implementing appropriate rate limiting mechanisms.

---

## Examples

### Complete Test Workflow

1. **Check Status**
   ```bash
   curl -X GET http://localhost:9990/status
   ```

2. **Start Test with WiFi Scenario**
   ```bash
   curl -X POST http://localhost:9990/presets \
     -H "Content-Type: application/json" \
     -d '{"type": "scenario", "name": "wifi"}'
   ```

3. **Run Test**
   ```bash
   curl -X POST http://localhost:9990/run-test \
     -H "Content-Type: application/json" \
     -d '{
       "mode": "test",
       "addr": ":9000",
       "connections": 2,
       "streams": 4,
       "duration": "30s"
     }'
   ```

4. **Get Metrics**
   ```bash
   curl -X GET http://localhost:9990/metrics
   ```

5. **Generate Report**
   ```bash
   curl -X GET "http://localhost:9990/report?format=json"
   ```

6. **Stop Test**
   ```bash
   curl -X POST http://localhost:9990/stop-test
   ```

### Integration with Monitoring Systems

#### Prometheus Integration

The API supports Prometheus metrics export. Enable it by setting `prometheus: true` in the test configuration.

#### Grafana Dashboard

Use the following query to create Grafana dashboards:

```promql
# QUIC connections
quic_connections_current

# QUIC latency
histogram_quantile(0.95, rate(quic_latency_seconds_bucket[5m]))

# QUIC throughput
rate(quic_bytes_sent_total[5m])
```

---

## SDK Examples

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type TestConfig struct {
    Mode        string `json:"mode"`
    Addr        string `json:"addr"`
    Connections int    `json:"connections"`
    Streams     int    `json:"streams"`
    Duration    string `json:"duration"`
}

func main() {
    config := TestConfig{
        Mode:        "test",
        Addr:        ":9000",
        Connections: 2,
        Streams:     4,
        Duration:    "30s",
    }
    
    jsonData, _ := json.Marshal(config)
    
    resp, err := http.Post(
        "http://localhost:9990/run-test",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    fmt.Println("Test started")
}
```

### Python

```python
import requests
import json

# Start test
config = {
    "mode": "test",
    "addr": ":9000",
    "connections": 2,
    "streams": 4,
    "duration": "30s"
}

response = requests.post(
    "http://localhost:9990/run-test",
    json=config
)

print("Test started:", response.json())

# Get metrics
metrics = requests.get("http://localhost:9990/metrics")
print("Metrics:", metrics.json())

# Generate report
report = requests.get("http://localhost:9990/report?format=json")
print("Report:", report.json())
```

### JavaScript/Node.js

```javascript
const axios = require('axios');

async function runTest() {
    try {
        // Start test
        const config = {
            mode: 'test',
            addr: ':9000',
            connections: 2,
            streams: 4,
            duration: '30s'
        };
        
        const startResponse = await axios.post(
            'http://localhost:9990/run-test',
            config
        );
        console.log('Test started:', startResponse.data);
        
        // Wait for test to complete
        await new Promise(resolve => setTimeout(resolve, 30000));
        
        // Get metrics
        const metricsResponse = await axios.get(
            'http://localhost:9990/metrics'
        );
        console.log('Metrics:', metricsResponse.data);
        
        // Generate report
        const reportResponse = await axios.get(
            'http://localhost:9990/report?format=json'
        );
        console.log('Report:', reportResponse.data);
        
    } catch (error) {
        console.error('Error:', error.message);
    }
}

runTest();
```

---

## Changelog

### Version 1.0.0
- Initial API release
- Basic test management endpoints
- Metrics and reporting functionality
- Preset management for scenarios and profiles
- Support for JSON, CSV, and Markdown report formats
