# quic-test Metrics Schema

Complete specification of all metrics exported by quic-test for monitoring, analysis, and machine learning integration.

## Metric Types

### Performance Metrics

Measures of network performance characteristics.

#### quic_latency_ms
- **Type:** Gauge
- **Unit:** Milliseconds
- **Description:** One-way latency measured from sender to receiver
- **Labels:**
  - `connection_id`: Unique connection identifier
  - `remote_addr`: Remote peer address (IP:port)
  - `protocol`: Protocol version (e.g., "quic_v1")
  - `tls_version`: TLS version (e.g., "1.3")
- **Example:**
  ```
  quic_latency_ms{connection_id="conn_001",remote_addr="192.168.1.100:9000"} 25.5
  ```

#### quic_rtt_ms
- **Type:** Gauge
- **Unit:** Milliseconds
- **Description:** Round-trip time from sender to receiver and back
- **Labels:** Same as quic_latency_ms
- **Notes:** RTT = 2 × one-way latency (approximately, in ideal network)

#### quic_jitter_ms
- **Type:** Gauge
- **Unit:** Milliseconds
- **Description:** Variation in RTT, calculated as standard deviation of recent RTT samples
- **Labels:** Same as quic_latency_ms
- **Calculation:**
  ```
  jitter = std_dev(rtt_measurements[last_30_samples])
  ```

#### quic_throughput_mbps
- **Type:** Gauge
- **Unit:** Megabits per second
- **Description:** Effective throughput (goodput) - useful data rate
- **Labels:** Same as quic_latency_ms
- **Notes:**
  - Does not include retransmitted packets
  - Excludes protocol overhead
  - Measured over 1-second intervals

#### quic_goodput_mbps
- **Type:** Gauge
- **Unit:** Megabits per second
- **Description:** Data throughput delivered to application
- **Labels:** Same as quic_latency_ms
- **Notes:** Similar to throughput but excludes all control messages

#### quic_packet_loss_rate
- **Type:** Gauge
- **Unit:** Fraction (0.0 to 1.0)
- **Description:** Ratio of lost packets to total packets transmitted
- **Labels:** Same as quic_latency_ms
- **Calculation:**
  ```
  loss_rate = lost_packets / (sent_packets + lost_packets)
  ```

#### quic_packet_loss_percent
- **Type:** Gauge
- **Unit:** Percentage (0 to 100)
- **Description:** Percentage of packets lost
- **Calculation:**
  ```
  loss_percent = loss_rate × 100
  ```

### Reliability Metrics

Measures of protocol reliability and error recovery.

#### quic_retransmits
- **Type:** Counter
- **Unit:** Count
- **Description:** Total number of retransmitted packets
- **Labels:**
  - `connection_id`: Unique connection identifier
  - `reason`: Reason for retransmission ("timeout", "dup_ack", "loss_detected")
- **Example:**
  ```
  quic_retransmits{connection_id="conn_001",reason="timeout"} 42
  ```

#### quic_lost_packets
- **Type:** Counter
- **Unit:** Count
- **Description:** Number of packets declared lost by loss detection algorithm
- **Labels:**
  - `connection_id`: Unique connection identifier
  - `loss_detection_method`: Method used ("threshold", "pto", "reno")
- **Notes:** May exceed retransmits if packets are not retransmitted

#### quic_duplicate_packets
- **Type:** Counter
- **Unit:** Count
- **Description:** Number of duplicate packets detected and discarded
- **Labels:**
  - `connection_id`: Unique connection identifier
  - `source`: Packet source ("network", "local_buffer")

#### quic_reordered_packets
- **Type:** Counter
- **Unit:** Count
- **Description:** Number of out-of-order packets received
- **Labels:**
  - `connection_id`: Unique connection identifier
  - `severity`: Reordering severity ("minor", "major")

#### quic_handshake_failures
- **Type:** Counter
- **Unit:** Count
- **Description:** Number of failed connection handshakes
- **Labels:**
  - `reason`: Failure reason ("tls_error", "timeout", "version_negotiation")

#### quic_connection_resets
- **Type:** Counter
- **Unit:** Count
- **Description:** Number of connection resets (abnormal terminations)
- **Labels:**
  - `error_code`: QUIC error code (numeric)

### Congestion Control Metrics

Measures of congestion control algorithm behavior.

#### quic_congestion_window
- **Type:** Gauge
- **Unit:** Bytes
- **Description:** Current congestion window size
- **Labels:**
  - `connection_id`: Unique connection identifier
  - `algorithm`: Congestion control algorithm ("bbr", "cubic", "reno")
- **Example:**
  ```
  quic_congestion_window{connection_id="conn_001",algorithm="bbr"} 131072
  ```

#### quic_bytes_in_flight
- **Type:** Gauge
- **Unit:** Bytes
- **Description:** Amount of unacknowledged data in the network
- **Labels:**
  - `connection_id`: Unique connection identifier
- **Formula:**
  ```
  bytes_in_flight = bytes_sent - bytes_acked
  ```

#### quic_slow_start_threshold
- **Type:** Gauge
- **Unit:** Bytes
- **Description:** Slow start threshold (ssthresh) for congestion control
- **Labels:**
  - `connection_id`: Unique connection identifier
- **Notes:** Relevant for CUBIC and Reno algorithms

#### quic_pacing_rate_mbps
- **Type:** Gauge
- **Unit:** Megabits per second
- **Description:** Rate at which packets are paced out
- **Labels:**
  - `connection_id`: Unique connection identifier
  - `algorithm`: Congestion control algorithm

#### quic_cc_recovery_mode
- **Type:** Gauge
- **Unit:** Enum (0=normal, 1=slow_start, 2=congestion_avoidance, 3=recovery)
- **Description:** Current congestion control state
- **Labels:**
  - `connection_id`: Unique connection identifier

#### quic_bbr_state
- **Type:** Gauge
- **Unit:** Enum
- **Description:** Current state of BBR congestion control
- **Labels:**
  - `connection_id`: Unique connection identifier
- **States:**
  - 0: STARTUP
  - 1: DRAIN
  - 2: PROBE_BW
  - 3: PROBE_RTT

### Connection Metrics

Measures of QUIC connection status and lifecycle.

#### quic_connections_active
- **Type:** Gauge
- **Unit:** Count
- **Description:** Number of currently active connections
- **Labels:**
  - `role`: Connection role ("server", "client")
  - `state`: Connection state ("handshake", "established", "closing")
- **Example:**
  ```
  quic_connections_active{role="server",state="established"} 42
  ```

#### quic_connections_established
- **Type:** Counter
- **Unit:** Count
- **Description:** Total number of connections successfully established
- **Labels:**
  - `role`: Connection role ("server", "client")

#### quic_connections_failed
- **Type:** Counter
- **Unit:** Count
- **Description:** Total number of failed connection attempts
- **Labels:**
  - `reason`: Failure reason ("timeout", "version_negotiation", "tls_error")

#### quic_connections_migrated
- **Type:** Counter
- **Unit:** Count
- **Description:** Number of connections that migrated to new paths
- **Labels:**
  - `migration_type`: Type of migration ("path_challenge", "nat_rebinding", "address_change")

#### quic_handshake_time_ms
- **Type:** Histogram
- **Unit:** Milliseconds
- **Description:** Duration of TLS 1.3 handshake
- **Labels:**
  - `role`: Connection role ("server", "client")
- **Buckets:** [1, 5, 10, 25, 50, 100, 250, 500, 1000]
- **Example:**
  ```
  quic_handshake_time_ms_bucket{role="client",le="100"} 152
  quic_handshake_time_ms_bucket{role="client",le="+Inf"} 200
  quic_handshake_time_ms_count{role="client"} 200
  quic_handshake_time_ms_sum{role="client"} 15000
  ```

#### quic_idle_timeout_ms
- **Type:** Gauge
- **Unit:** Milliseconds
- **Description:** Idle timeout value for this connection
- **Labels:**
  - `connection_id`: Unique connection identifier

### Stream Metrics

Measures related to QUIC streams within connections.

#### quic_streams_active
- **Type:** Gauge
- **Unit:** Count
- **Description:** Number of active streams in connection
- **Labels:**
  - `connection_id`: Unique connection identifier
  - `stream_type`: Type of stream ("bidi", "uni")

#### quic_streams_created
- **Type:** Counter
- **Unit:** Count
- **Description:** Total number of streams created in connection
- **Labels:**
  - `connection_id`: Unique connection identifier

#### quic_streams_reset
- **Type:** Counter
- **Unit:** Count
- **Description:** Number of streams reset by either peer
- **Labels:**
  - `connection_id`: Unique connection identifier
  - `direction`: Direction of reset ("local", "remote")

#### quic_stream_data_blocked
- **Type:** Counter
- **Unit:** Count
- **Description:** Number of STREAM_DATA_BLOCKED frames sent
- **Labels:**
  - `connection_id`: Unique connection identifier

#### quic_connections_blocked
- **Type:** Counter
- **Unit:** Count
- **Description:** Number of CONNECTION_BLOCKED frames sent
- **Labels:**
  - `connection_id`: Unique connection identifier

### Frame-Level Metrics

Detailed metrics about QUIC frame types.

#### quic_frames_sent
- **Type:** Counter
- **Unit:** Count
- **Description:** Number of specific frame type sent
- **Labels:**
  - `connection_id`: Unique connection identifier
  - `frame_type`: Frame type (ACK, CRYPTO, STREAM, PADDING, etc.)
- **Example:**
  ```
  quic_frames_sent{connection_id="conn_001",frame_type="ACK"} 5000
  quic_frames_sent{connection_id="conn_001",frame_type="STREAM"} 10000
  ```

#### quic_frames_received
- **Type:** Counter
- **Unit:** Count
- **Description:** Number of specific frame type received
- **Labels:**
  - `connection_id`: Unique connection identifier
  - `frame_type`: Frame type

#### quic_frames_processed_time_us
- **Type:** Histogram
- **Unit:** Microseconds
- **Description:** Time to process specific frame type
- **Labels:**
  - `frame_type`: Frame type
- **Buckets:** [10, 50, 100, 500, 1000, 5000, 10000]

### Data Transfer Metrics

Byte-level metrics for data transfer.

#### quic_bytes_sent
- **Type:** Counter
- **Unit:** Bytes
- **Description:** Total bytes sent (includes retransmissions and overhead)
- **Labels:**
  - `connection_id`: Unique connection identifier
  - `packet_type`: Packet type ("initial", "handshake", "1rtt", "0rtt")

#### quic_bytes_received
- **Type:** Counter
- **Unit:** Bytes
- **Description:** Total bytes received
- **Labels:**
  - `connection_id`: Unique connection identifier

#### quic_payload_bytes_sent
- **Type:** Counter
- **Unit:** Bytes
- **Description:** Bytes of actual payload data sent
- **Labels:**
  - `connection_id`: Unique connection identifier

#### quic_payload_bytes_received
- **Type:** Counter
- **Unit:** Bytes
- **Description:** Bytes of actual payload data received
- **Labels:**
  - `connection_id`: Unique connection identifier

#### quic_protocol_overhead_bytes
- **Type:** Gauge
- **Unit:** Bytes
- **Description:** Bytes of protocol overhead (headers, etc.)
- **Labels:**
  - `connection_id`: Unique connection identifier
  - `overhead_type`: Type of overhead ("header", "padding", "frame_header")

### Error Metrics

Metrics related to errors and exceptions.

#### quic_errors_total
- **Type:** Counter
- **Unit:** Count
- **Description:** Total number of errors
- **Labels:**
  - `error_type`: Type of error ("protocol_error", "transport_error", "crypto_error")
  - `error_code`: Numeric error code

#### quic_crypto_errors
- **Type:** Counter
- **Unit:** Count
- **Description:** Cryptographic errors
- **Labels:**
  - `error_details`: Error details ("decryption_failed", "key_update_failed", "bad_signature")

#### quic_timeout_errors
- **Type:** Counter
- **Unit:** Count
- **Description:** Connection timeout errors
- **Labels:**
  - `timeout_type`: Type of timeout ("idle", "pto", "handshake")

### Experimental Feature Metrics

Metrics for QUIC experimental features.

#### quic_fec_packets_sent
- **Type:** Counter
- **Unit:** Count
- **Description:** Number of FEC (Forward Error Correction) repair packets sent
- **Labels:**
  - `connection_id`: Unique connection identifier
  - `fec_scheme`: FEC scheme ("reed_solomon", "xor")

#### quic_fec_packets_recovered
- **Type:** Counter
- **Unit:** Count
- **Description:** Number of packets recovered using FEC
- **Labels:**
  - `connection_id`: Unique connection identifier

#### quic_ack_delay_us
- **Type:** Histogram
- **Unit:** Microseconds
- **Description:** ACK delay measured
- **Labels:**
  - `connection_id`: Unique connection identifier
- **Buckets:** [100, 500, 1000, 5000, 10000, 25000]

## Export Formats

### Prometheus Text Format

Standard Prometheus exposition format used by the HTTP endpoint.

**Endpoint:** `http://<host>:<prometheus_port>/metrics`

**Example Output:**
```
# HELP quic_latency_ms Latency in milliseconds
# TYPE quic_latency_ms gauge
quic_latency_ms{connection_id="conn_001",remote_addr="192.168.1.100:9000"} 25.5

# HELP quic_jitter_ms Jitter in milliseconds
# TYPE quic_jitter_ms gauge
quic_jitter_ms{connection_id="conn_001"} 2.3

# HELP quic_throughput_mbps Throughput in Mbps
# TYPE quic_throughput_mbps gauge
quic_throughput_mbps{connection_id="conn_001"} 150.2

# HELP quic_retransmits Total retransmitted packets
# TYPE quic_retransmits counter
quic_retransmits{connection_id="conn_001",reason="timeout"} 42
```

### JSON Export Format

Structured format for programmatic access.

**Endpoint:** `http://<host>:<prometheus_port>/metrics?format=json`

**Example Output:**
```json
{
  "timestamp": 1700000000,
  "duration_seconds": 60,
  "metrics": [
    {
      "name": "quic_latency_ms",
      "type": "gauge",
      "unit": "milliseconds",
      "samples": [
        {
          "connection_id": "conn_001",
          "remote_addr": "192.168.1.100:9000",
          "value": 25.5,
          "timestamp": 1700000000
        }
      ]
    },
    {
      "name": "quic_throughput_mbps",
      "type": "gauge",
      "unit": "megabits_per_second",
      "samples": [
        {
          "connection_id": "conn_001",
          "value": 150.2,
          "timestamp": 1700000000
        }
      ]
    }
  ],
  "summary": {
    "total_connections": 1,
    "active_connections": 1,
    "total_bytes_sent": 1048576,
    "total_bytes_received": 1048576
  }
}
```

### CSV Export Format

Tabular format for data analysis.

**Endpoint:** `http://<host>:<prometheus_port>/metrics?format=csv`

**Example Output:**
```csv
timestamp,metric_name,connection_id,remote_addr,value,unit
1700000000,quic_latency_ms,conn_001,192.168.1.100:9000,25.5,milliseconds
1700000000,quic_jitter_ms,conn_001,192.168.1.100:9000,2.3,milliseconds
1700000000,quic_throughput_mbps,conn_001,192.168.1.100:9000,150.2,megabits_per_second
1700000000,quic_packet_loss_rate,conn_001,192.168.1.100:9000,0.001,fraction
```

## Usage Examples

### Querying with curl

```bash
# Get all metrics
curl http://localhost:9090/metrics

# Get metrics in JSON format
curl http://localhost:9090/metrics?format=json

# Get specific connection metrics
curl 'http://localhost:9090/metrics?connection_id=conn_001'

# Get metrics over time interval
curl 'http://localhost:9090/metrics?start_time=1700000000&end_time=1700001000'
```

### Integration with Prometheus

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'quic-test'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: '/metrics'
```

### Integration with Python (pandas)

```python
import pandas as pd
import requests

# Fetch metrics from quic-test
response = requests.get('http://localhost:9090/metrics?format=json')
data = response.json()

# Convert to DataFrame
metrics_list = []
for metric in data['metrics']:
    for sample in metric['samples']:
        row = {
            'metric': metric['name'],
            'unit': metric['unit'],
            **sample
        }
        metrics_list.append(row)

df = pd.DataFrame(metrics_list)
print(df.head())
```

## Retention Policy

- **In-memory retention:** Last 5 minutes of detailed metrics
- **File retention:** 1 hour (configurable)
- **Export retention:** Depends on external storage (Prometheus, etc.)

---

**Last Updated:** November 2025
**Version:** 1.0
**Status:** Production Ready
