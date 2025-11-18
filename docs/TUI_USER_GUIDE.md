# QUIC Bottom Real - TUI User Guide

## Overview

QUIC Bottom Real is a Terminal User Interface (TUI) application for real-time monitoring and visualization of QUIC protocol metrics. The application provides comprehensive analytics, network status monitoring, security metrics, and BBRv3 congestion control visualization.

## Table of Contents

1. [Installation and Setup](#installation-and-setup)
2. [Starting the Application](#starting-the-application)
3. [User Interface Overview](#user-interface-overview)
4. [View Modes](#view-modes)
5. [Keyboard Shortcuts](#keyboard-shortcuts)
6. [Metrics Reference](#metrics-reference)
7. [Troubleshooting](#troubleshooting)

## Installation and Setup

### Prerequisites

- Rust toolchain (1.70 or later)
- Compiled `quic-bottom-real` binary
- QUIC test application running and sending metrics to HTTP API

### Building the Application

```bash
cd quic-bottom
cargo build --release --bin quic-bottom-real
```

The binary will be located at `target/release/quic-bottom-real`.

## Starting the Application

### Basic Startup

```bash
cd quic-bottom
./target/release/quic-bottom-real
```

The application will start in TUI mode and begin listening on `http://127.0.0.1:8080` for metrics from the QUIC test application.

### Headless Mode

To run the application in headless mode (HTTP API only, no TUI):

```bash
./target/release/quic-bottom-real --headless
```

## User Interface Overview

The TUI consists of three main sections:

1. **Header**: Displays application title and current view mode
2. **Main Content Area**: Displays metrics, graphs, and visualizations based on selected view
3. **Footer**: Displays keyboard shortcuts and status information

The interface updates in real-time as metrics are received from the QUIC test application.

## View Modes

The application provides six distinct view modes, each optimized for specific analysis tasks.

### View 1: Dashboard

**Access**: Press `1` or default view on startup

**Contents**:
- Current Metrics Widget: Real-time display of key performance indicators
  - Connections count
  - Latency (ms)
  - Throughput (Mbps)
  - RTT (ms)
  - Packet Loss (%)
  - Retransmits count
  - Errors count
  - Streams count

- Latency Graph: Time-series visualization of latency measurements
- Throughput Graph: Time-series visualization of throughput measurements

**Use Cases**:
- Quick overview of system performance
- Monitoring basic connection health
- Identifying performance trends

### View 2: Analytics

**Access**: Press `2`

**Contents**:
- QUIC Metrics Correlation Matrix: Correlation coefficients between different metrics
  - Latency
  - Throughput
  - Packet Loss
  - RTT
  - Jitter
  - Retransmits
  - Connections
  - Errors

- Anomaly Detection Widget: Statistical analysis for identifying unusual patterns

**Correlation Interpretation**:
- Values range from -1.0 to 1.0
- Positive values indicate positive correlation (metrics increase together)
- Negative values indicate inverse correlation (one metric increases as the other decreases)
- Values close to 0 indicate weak or no correlation
- Color coding:
  - Red: Strong correlation (absolute value >= 0.8)
  - Yellow: Moderate correlation (absolute value >= 0.4)
  - Green: Weak correlation (absolute value < 0.4)

**Use Cases**:
- Understanding relationships between metrics
- Identifying performance bottlenecks
- Detecting anomalies in network behavior

### View 3: Network Status

**Access**: Press `3`

**Contents**:
- Network Simulation Status: Configuration of simulated network conditions
  - Simulation state (ACTIVE/INACTIVE)
  - Network preset name
  - Simulated latency (ms)
  - Simulated packet loss (%)
  - Simulated bandwidth (Mbps)

- Real Metrics: Actual measured network performance
  - Actual latency (ms)
  - Actual throughput (Mbps)
  - Actual RTT (ms)
  - Packet loss (%)
  - Retransmits count
  - Connections count

**Use Cases**:
- Comparing simulated vs actual network conditions
- Validating network simulation accuracy
- Understanding impact of network conditions on performance

### View 4: Security Status

**Access**: Press `4`

**Contents**:
- Security Testing Status: Configuration of security testing
  - Testing state (ACTIVE/INACTIVE)
  - Security score (%)
  - Vulnerabilities count

- Connection Security Metrics: Real-time security indicators
  - Errors count
  - Error rate (%)
  - Packet loss (%)
  - Retransmits count
  - Handshake time (ms)
  - Jitter (ms)

**Security Score Calculation**:
- Based on error rate and packet loss
- Formula: 100 - error_rate - (packet_loss * 100)
- Higher scores indicate better security posture

**Use Cases**:
- Monitoring connection security
- Identifying potential security issues
- Validating secure connection establishment

### View 5: Cloud Deployment

**Access**: Press `5`

**Contents**:
- Cloud Deployment Status: Configuration of cloud deployment simulation
  - Deployment state (ACTIVE/INACTIVE)
  - Cloud provider name
  - Number of instances
  - Deployment status

**Use Cases**:
- Simulating cloud deployment scenarios
- Testing multi-instance configurations
- Cloud-specific performance analysis

### View 6: BBRv3 Congestion Control

**Access**: Press `6`

**Contents**:
- Phase Status: Current BBRv3 operational phase
  - Startup: Initial bandwidth probing
  - Drain: Buffer draining phase
  - ProbeBW: Bandwidth probing phase
  - ProbeRTT: RTT probing phase

- Bandwidth Estimates: Dual-scale bandwidth estimation
  - Fast Bandwidth: Rapid response to changes (Mbps)
  - Slow Bandwidth: Stable long-term estimate (Mbps)
  - Ratio: Fast/Slow bandwidth ratio

- Loss Metrics: Packet loss analysis
  - Loss Rate (EMA): Exponential moving average of loss rate (%)
  - Status: HEALTHY (< 2%) or ELEVATED (>= 2%)
  - Threshold: 2.0%

- Bufferbloat and Stability: Network buffer analysis
  - Bufferbloat Factor: (avg_rtt / min_rtt) - 1
    - EXCELLENT: < 0.1
    - GOOD: 0.1 - 0.3
    - HIGH: > 0.3
  - Stability Index: Throughput stability metric

- Pacing and CWND Gains: Congestion control parameters
  - Pacing Gain: Current pacing rate multiplier
  - CWND Gain: Congestion window size multiplier
  - Target Inflight: Target bytes in flight (KB)

- Recovery Metrics: Loss recovery performance
  - Recovery Time: Time to recover from packet loss (ms)
  - Loss Recovery Efficiency: Percentage of lost packets recovered
  - Headroom Usage: Buffer headroom utilization (%)

**BBRv3 Phase Descriptions**:
- **Startup**: Rapidly increases sending rate to discover available bandwidth
- **Drain**: Reduces sending rate to drain network buffers
- **ProbeBW**: Periodically probes for additional bandwidth
- **ProbeRTT**: Periodically reduces sending rate to measure minimum RTT

**Use Cases**:
- Understanding BBRv3 algorithm behavior
- Optimizing congestion control parameters
- Analyzing network buffer conditions
- Monitoring bandwidth estimation accuracy

## Keyboard Shortcuts

### Navigation

| Key | Action |
|-----|--------|
| `1` | Switch to Dashboard view |
| `2` | Switch to Analytics view |
| `3` | Switch to Network Status view |
| `4` | Switch to Security Status view |
| `5` | Switch to Cloud Deployment view |
| `6` | Switch to BBRv3 Congestion Control view |
| `a` | Switch to All Views (combined view) |

### Application Control

| Key | Action |
|-----|--------|
| `q` | Quit application |
| `ESC` | Quit application |
| `r` | Reset all collected data |
| `h` | Show help information |

### Network Simulation (View 3)

| Key | Action |
|-----|--------|
| `n` | Toggle network simulation on/off |
| `+` | Switch to next network preset |
| `-` | Switch to previous network preset |

### Security Testing (View 4)

| Key | Action |
|-----|--------|
| `s` | Toggle security testing on/off |

### Cloud Deployment (View 5)

| Key | Action |
|-----|--------|
| `d` | Toggle cloud deployment on/off |
| `i` | Scale cloud instances (cycle through 1-5) |

## Metrics Reference

### Basic Metrics

**Latency**: Round-trip time for data packets, measured in milliseconds. Lower values indicate better performance.

**Throughput**: Data transfer rate, measured in Megabits per second (Mbps). Higher values indicate better performance.

**RTT (Round-Trip Time)**: Time taken for a packet to travel from source to destination and back, measured in milliseconds.

**Jitter**: Variation in latency between consecutive packets, measured in milliseconds. Lower values indicate more stable connection.

**Packet Loss**: Percentage of packets lost during transmission. Lower values indicate better network quality.

**Retransmits**: Number of packets that required retransmission. Lower values indicate better network reliability.

**Connections**: Number of active QUIC connections.

**Errors**: Number of connection or transmission errors encountered.

**Streams**: Number of active QUIC streams within connections.

**Handshake Time**: Time required to establish QUIC connection, measured in milliseconds.

### BBRv3 Specific Metrics

**BBRv3 Phase**: Current operational phase of the BBRv3 congestion control algorithm.

**Bandwidth Fast**: Fast-scale bandwidth estimate that responds quickly to network changes, measured in bits per second.

**Bandwidth Slow**: Slow-scale bandwidth estimate that provides stable long-term assessment, measured in bits per second.

**Loss Rate EMA**: Exponential moving average of packet loss rate, expressed as a decimal (e.g., 0.02 = 2%).

**Loss Threshold**: Maximum acceptable loss rate before BBRv3 reduces sending rate. Default: 2.0%.

**Bufferbloat Factor**: Measure of network buffer congestion. Calculated as (average_rtt / minimum_rtt) - 1. Values below 0.1 indicate excellent buffer management.

**Stability Index**: Metric indicating connection stability, calculated as change in throughput divided by change in RTT. Higher values indicate more stable connections.

**Pacing Gain**: Multiplier applied to the pacing rate. Values greater than 1.0 increase sending rate, values less than 1.0 decrease it.

**CWND Gain**: Multiplier applied to the congestion window size. Determines how much data can be in flight.

**Inflight Target**: Target number of bytes that should be in flight at any given time, measured in bytes.

**Recovery Time**: Time required to recover from packet loss events, measured in milliseconds.

**Loss Recovery Efficiency**: Percentage of lost packets that were successfully recovered through retransmission or FEC.

**Headroom Usage**: Percentage of available buffer headroom currently in use. Lower values provide more margin for traffic spikes.

## Troubleshooting

### No Metrics Displayed

**Symptoms**: All metrics show zero or "N/A"

**Possible Causes**:
1. QUIC test application is not running
2. QUIC test application is not sending metrics to the correct endpoint
3. Network connectivity issues between applications

**Solutions**:
1. Verify QUIC test application is running and configured to send metrics to `http://127.0.0.1:8080/api/metrics`
2. Check HTTP API health: `curl http://127.0.0.1:8080/health`
3. Verify firewall settings allow localhost connections

### BBRv3 Metrics Not Available

**Symptoms**: BBRv3 view shows "BBRv3 metrics not available"

**Possible Causes**:
1. QUIC test application not running with `--cc=bbrv3` flag
2. BBRv3 integration not properly initialized
3. Insufficient data collected yet

**Solutions**:
1. Ensure QUIC test application is started with `--cc=bbrv3` parameter
2. Check QUIC test application logs for BBRv3 initialization messages
3. Wait 10-15 seconds after starting test for metrics to accumulate

### Correlation Matrix Shows All Zeros

**Symptoms**: Correlation matrix displays 0.00 for all metric pairs

**Possible Causes**:
1. Metrics have constant values (no variance)
2. Insufficient data points collected
3. Metrics are not changing over time

**Solutions**:
1. Ensure QUIC test application is actively sending varying metrics
2. Wait for at least 3-5 data points per metric
3. Verify test parameters create varying network conditions

### Application Crashes or Freezes

**Symptoms**: TUI becomes unresponsive or exits unexpectedly

**Possible Causes**:
1. Terminal compatibility issues
2. Insufficient terminal size
3. Memory issues with large datasets

**Solutions**:
1. Ensure terminal supports ANSI escape codes
2. Resize terminal to at least 80x24 characters
3. Restart application and reset data with `r` key if needed

### HTTP API Not Responding

**Symptoms**: Cannot connect to HTTP API endpoint

**Possible Causes**:
1. Application not started
2. Port 8080 already in use
3. Firewall blocking connections

**Solutions**:
1. Verify application is running: `ps aux | grep quic-bottom-real`
2. Check if port is in use: `lsof -i :8080`
3. Verify localhost connectivity: `curl http://127.0.0.1:8080/health`

## Best Practices

1. **Start with Dashboard View**: Begin analysis with the Dashboard view to get an overview of system performance.

2. **Monitor BBRv3 Metrics**: Use BBRv3 view to understand congestion control behavior, especially when testing different network conditions.

3. **Analyze Correlations**: Use Analytics view to identify relationships between metrics that may indicate performance bottlenecks.

4. **Compare Simulated vs Actual**: Use Network Status view to validate that network simulation accurately represents real conditions.

5. **Regular Data Reset**: Use `r` key periodically to reset accumulated data and start fresh analysis sessions.

6. **Terminal Size**: Ensure terminal window is large enough (minimum 80x24) for proper display of all widgets and graphs.

7. **Long-Running Tests**: For extended test sessions, monitor memory usage and restart application if needed.

## API Integration

The application exposes an HTTP API for programmatic access to metrics:

- **GET /health**: Health check endpoint
- **GET /api/current**: Retrieve current metrics as JSON
- **POST /api/metrics**: Receive metrics from QUIC test application

For detailed API documentation, refer to the API specification document.

## Version Information

This guide applies to QUIC Bottom Real version 0.1.0 and later.

For updates and additional documentation, refer to the project repository.

