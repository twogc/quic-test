# 2GC Network Protocol Suite - API Documentation

## Overview

The 2GC Network Protocol Suite provides comprehensive APIs for testing, monitoring, and managing network protocols. This document covers all available APIs, including REST endpoints, gRPC services, and WebSocket connections.

## REST API

### Base URL
```
http://localhost:9990/api/v1
```

### Authentication
All API endpoints require authentication via API key:
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" http://localhost:9990/api/v1/status
```

### Core Endpoints

#### 1. System Status

**GET /status**
Returns system health and status information.

```bash
curl http://localhost:9990/api/v1/status
```

Response:
```json
{
  "status": "healthy",
  "version": "1.2.3",
  "uptime": "2h30m15s",
  "components": {
    "quic": "active",
    "masque": "active",
    "ice": "active",
    "monitoring": "active"
  },
  "metrics": {
    "active_connections": 42,
    "total_tests": 156,
    "success_rate": 0.98
  }
}
```

#### 2. Test Management

**POST /tests/start**
Start a new test with specified parameters.

```bash
curl -X POST http://localhost:9990/api/v1/tests/start \
  -H "Content-Type: application/json" \
  -d '{
    "name": "performance_test",
    "scenario": "wifi",
    "duration": "5m",
    "connections": 10,
    "streams": 5,
    "packet_size": 1200,
    "rate": 100
  }'
```

Response:
```json
{
  "test_id": "test_12345",
  "status": "started",
  "start_time": "2024-01-15T10:30:00Z",
  "estimated_duration": "5m",
  "monitoring_url": "http://localhost:9990/api/v1/tests/test_12345/metrics"
}
```

**GET /tests/{test_id}/status**
Get current status of a running test.

```bash
curl http://localhost:9990/api/v1/tests/test_12345/status
```

Response:
```json
{
  "test_id": "test_12345",
  "status": "running",
  "progress": 0.65,
  "elapsed_time": "3m15s",
  "remaining_time": "1m45s",
  "current_metrics": {
    "rtt_p95": "25ms",
    "throughput": "850Mbps",
    "loss_rate": "0.02%"
  }
}
```

**POST /tests/{test_id}/stop**
Stop a running test.

```bash
curl -X POST http://localhost:9990/api/v1/tests/test_12345/stop
```

Response:
```json
{
  "test_id": "test_12345",
  "status": "stopped",
  "stop_time": "2024-01-15T10:35:00Z",
  "final_metrics": {
    "rtt_p95": "23ms",
    "throughput": "920Mbps",
    "loss_rate": "0.01%"
  }
}
```

#### 3. Metrics and Monitoring

**GET /tests/{test_id}/metrics**
Get real-time metrics for a test.

```bash
curl http://localhost:9990/api/v1/tests/test_12345/metrics
```

Response:
```json
{
  "test_id": "test_12345",
  "timestamp": "2024-01-15T10:32:30Z",
  "metrics": {
    "latency": {
      "p50": "15ms",
      "p95": "25ms",
      "p99": "35ms",
      "p999": "50ms"
    },
    "throughput": {
      "current": "850Mbps",
      "average": "820Mbps",
      "peak": "950Mbps"
    },
    "packet_loss": {
      "rate": "0.02%",
      "total_packets": 125000,
      "lost_packets": 25
    },
    "connections": {
      "active": 10,
      "established": 10,
      "failed": 0
    },
    "streams": {
      "active": 50,
      "completed": 1250,
      "failed": 2
    }
  }
}
```

**GET /tests/{test_id}/metrics/history**
Get historical metrics for a test.

```bash
curl "http://localhost:9990/api/v1/tests/test_12345/metrics/history?from=2024-01-15T10:30:00Z&to=2024-01-15T10:35:00Z&interval=30s"
```

Response:
```json
{
  "test_id": "test_12345",
  "time_range": {
    "from": "2024-01-15T10:30:00Z",
    "to": "2024-01-15T10:35:00Z"
  },
  "interval": "30s",
  "data_points": [
    {
      "timestamp": "2024-01-15T10:30:00Z",
      "rtt_p95": "20ms",
      "throughput": "800Mbps",
      "loss_rate": "0.01%"
    },
    {
      "timestamp": "2024-01-15T10:30:30Z",
      "rtt_p95": "22ms",
      "throughput": "850Mbps",
      "loss_rate": "0.02%"
    }
  ]
}
```

#### 4. Test Scenarios

**GET /scenarios**
List available test scenarios.

```bash
curl http://localhost:9990/api/v1/scenarios
```

Response:
```json
{
  "scenarios": [
    {
      "id": "wifi",
      "name": "WiFi Network",
      "description": "Home WiFi network simulation",
      "parameters": {
        "rtt": "20ms",
        "jitter": "5ms",
        "loss": "0.1%",
        "bandwidth": "100Mbps"
      }
    },
    {
      "id": "lte",
      "name": "LTE Network",
      "description": "Mobile LTE network simulation",
      "parameters": {
        "rtt": "50ms",
        "jitter": "15ms",
        "loss": "0.5%",
        "bandwidth": "50Mbps"
      }
    }
  ]
}
```

**GET /scenarios/{scenario_id}**
Get detailed information about a specific scenario.

```bash
curl http://localhost:9990/api/v1/scenarios/wifi
```

Response:
```json
{
  "id": "wifi",
  "name": "WiFi Network",
  "description": "Home WiFi network simulation",
  "parameters": {
    "rtt": "20ms",
    "jitter": "5ms",
    "loss": "0.1%",
    "bandwidth": "100Mbps"
  },
  "expected_metrics": {
    "rtt_p95": "25-30ms",
    "throughput": "80-90Mbps",
    "loss_rate": "<0.2%"
  }
}
```

#### 5. Network Profiles

**GET /profiles**
List available network profiles.

```bash
curl http://localhost:9990/api/v1/profiles
```

Response:
```json
{
  "profiles": [
    {
      "id": "datacenter",
      "name": "Datacenter",
      "description": "Low-latency datacenter network",
      "characteristics": {
        "rtt": "1ms",
        "jitter": "0.1ms",
        "loss": "0%",
        "bandwidth": "10Gbps"
      }
    },
    {
      "id": "satellite",
      "name": "Satellite",
      "description": "High-latency satellite network",
      "characteristics": {
        "rtt": "600ms",
        "jitter": "50ms",
        "loss": "1%",
        "bandwidth": "10Mbps"
      }
    }
  ]
}
```

#### 6. Reports

**GET /tests/{test_id}/report**
Get test report in specified format.

```bash
curl "http://localhost:9990/api/v1/tests/test_12345/report?format=json"
```

Response:
```json
{
  "test_id": "test_12345",
  "test_name": "performance_test",
  "start_time": "2024-01-15T10:30:00Z",
  "end_time": "2024-01-15T10:35:00Z",
  "duration": "5m",
  "scenario": "wifi",
  "summary": {
    "status": "completed",
    "success_rate": 0.98,
    "sla_compliance": true
  },
  "metrics": {
    "latency": {
      "p50": "15ms",
      "p95": "25ms",
      "p99": "35ms"
    },
    "throughput": {
      "average": "820Mbps",
      "peak": "950Mbps"
    },
    "packet_loss": "0.02%"
  },
  "sla_results": {
    "rtt_p95": {
      "threshold": "50ms",
      "actual": "25ms",
      "passed": true
    },
    "loss_rate": {
      "threshold": "1%",
      "actual": "0.02%",
      "passed": true
    }
  }
}
```

## gRPC API

### Service Definition

```protobuf
syntax = "proto3";

package quictest.v1;

service TestService {
  rpc StartTest(StartTestRequest) returns (StartTestResponse);
  rpc StopTest(StopTestRequest) returns (StopTestResponse);
  rpc GetTestStatus(GetTestStatusRequest) returns (GetTestStatusResponse);
  rpc GetMetrics(GetMetricsRequest) returns (GetMetricsResponse);
  rpc StreamMetrics(StreamMetricsRequest) returns (stream MetricsUpdate);
}

message StartTestRequest {
  string name = 1;
  string scenario = 2;
  int32 duration_seconds = 3;
  int32 connections = 4;
  int32 streams = 5;
  int32 packet_size = 6;
  int32 rate = 7;
}

message StartTestResponse {
  string test_id = 1;
  string status = 2;
  string start_time = 3;
  string estimated_duration = 4;
}

message GetMetricsRequest {
  string test_id = 1;
  string metric_type = 2;
}

message GetMetricsResponse {
  string test_id = 1;
  string timestamp = 2;
  map<string, MetricValue> metrics = 3;
}

message MetricValue {
  oneof value {
    double number = 1;
    string text = 2;
    bool boolean = 3;
  }
}
```

### Usage Example

```go
package main

import (
    "context"
    "log"
    
    "google.golang.org/grpc"
    pb "github.com/twogc/quic-test/proto"
)

func main() {
    conn, err := grpc.Dial("localhost:9991", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    client := pb.NewTestServiceClient(conn)
    
    // Start test
    resp, err := client.StartTest(context.Background(), &pb.StartTestRequest{
        Name: "performance_test",
        Scenario: "wifi",
        DurationSeconds: 300,
        Connections: 10,
        Streams: 5,
        PacketSize: 1200,
        Rate: 100,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Test started: %s", resp.TestId)
}
```

## WebSocket API

### Connection

```javascript
const ws = new WebSocket('ws://localhost:9990/ws/metrics');
```

### Message Format

```json
{
  "type": "metrics_update",
  "test_id": "test_12345",
  "timestamp": "2024-01-15T10:32:30Z",
  "data": {
    "rtt_p95": "25ms",
    "throughput": "850Mbps",
    "loss_rate": "0.02%"
  }
}
```

### Event Types

- `metrics_update`: Real-time metrics update
- `test_status`: Test status change
- `error`: Error notification
- `heartbeat`: Connection keep-alive

### Usage Example

```javascript
const ws = new WebSocket('ws://localhost:9990/ws/metrics');

ws.onopen = function() {
    console.log('Connected to metrics stream');
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    
    switch(data.type) {
        case 'metrics_update':
            updateMetricsDisplay(data.data);
            break;
        case 'test_status':
            updateTestStatus(data.status);
            break;
        case 'error':
            handleError(data.error);
            break;
    }
};

function updateMetricsDisplay(metrics) {
    document.getElementById('rtt').textContent = metrics.rtt_p95;
    document.getElementById('throughput').textContent = metrics.throughput;
    document.getElementById('loss').textContent = metrics.loss_rate;
}
```

## Error Handling

### HTTP Status Codes

- `200 OK`: Successful request
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Missing or invalid API key
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict (e.g., test already running)
- `500 Internal Server Error`: Server error

### Error Response Format

```json
{
  "error": {
    "code": "INVALID_PARAMETERS",
    "message": "Invalid test parameters provided",
    "details": {
      "field": "duration",
      "value": "invalid",
      "expected": "positive integer"
    }
  }
}
```

### Common Error Codes

- `INVALID_PARAMETERS`: Invalid request parameters
- `TEST_NOT_FOUND`: Test ID not found
- `TEST_ALREADY_RUNNING`: Test already in progress
- `SCENARIO_NOT_FOUND`: Scenario not found
- `INSUFFICIENT_RESOURCES`: Insufficient system resources
- `NETWORK_ERROR`: Network connectivity issue

## Rate Limiting

API requests are rate-limited to prevent abuse:

- **Standard endpoints**: 100 requests per minute
- **Metrics endpoints**: 1000 requests per minute
- **WebSocket connections**: 10 concurrent connections

Rate limit headers are included in responses:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642248000
```

## Authentication

### API Key Authentication

All API requests require an API key in the Authorization header:

```bash
curl -H "Authorization: Bearer YOUR_API_KEY" http://localhost:9990/api/v1/status
```

### API Key Management

API keys can be generated and managed through the web dashboard or CLI:

```bash
# Generate new API key
./quic-test --generate-api-key --name "production"

# List API keys
./quic-test --list-api-keys

# Revoke API key
./quic-test --revoke-api-key --key-id "key_12345"
```

## SDKs and Libraries

### Go SDK

```go
import "github.com/twogc/quic-test/sdk/go"

client := quictest.NewClient("http://localhost:9990", "your-api-key")

test, err := client.StartTest(&quictest.TestConfig{
    Name: "performance_test",
    Scenario: "wifi",
    Duration: "5m",
    Connections: 10,
})
```

### Python SDK

```python
from quictest import Client

client = Client("http://localhost:9990", "your-api-key")

test = client.start_test(
    name="performance_test",
    scenario="wifi",
    duration="5m",
    connections=10
)
```

### JavaScript SDK

```javascript
import { QuicTestClient } from '@twogc/quic-test-sdk';

const client = new QuicTestClient('http://localhost:9990', 'your-api-key');

const test = await client.startTest({
    name: 'performance_test',
    scenario: 'wifi',
    duration: '5m',
    connections: 10
});
```