# Education

Laboratory materials for universities and educational institutions.

## Overview

`quic-test` is designed for teaching network protocols, congestion control, and performance analysis. This document provides ready-to-use lab materials.

## Lab #1: QUIC Basics

**Duration:** 2 hours  
**Level:** Undergraduate  
**Prerequisites:** Basic networking knowledge (TCP/IP, HTTP)

### Learning Objectives

1. Understand QUIC protocol fundamentals
2. Observe 0-RTT connection resumption
3. Analyze connection migration
4. Compare QUIC vs TCP performance

### Setup

```bash
# Install quic-test
docker pull mlanies/quic-test:latest

# Or build from source
git clone https://github.com/twogc/quic-test
cd quic-test && make build
```

### Exercise 1.1: First QUIC Connection

**Task:** Establish a QUIC connection and measure RTT.

```bash
# Terminal 1: Start server
docker run -p 4433:4433/udp mlanies/quic-test:latest --mode=server

# Terminal 2: Run client
docker run mlanies/quic-test:latest \
  --mode=client \
  --server=<server-ip>:4433 \
  --duration=30s
```

**Questions:**
1. What is the handshake duration?
2. How many round trips for connection establishment?
3. What is the average RTT?

### Exercise 1.2: 0-RTT Resumption

**Task:** Observe 0-RTT connection resumption.

```bash
# First connection (full handshake)
docker run mlanies/quic-test:latest \
  --mode=client \
  --server=<server-ip>:4433 \
  --duration=10s

# Second connection (0-RTT)
docker run mlanies/quic-test:latest \
  --mode=client \
  --server=<server-ip>:4433 \
  --duration=10s \
  --0rtt
```

**Questions:**
1. How much faster is 0-RTT connection?
2. What data is sent in 0-RTT?
3. What are security implications of 0-RTT?

### Exercise 1.3: Connection Migration

**Task:** Simulate connection migration (network change).

```bash
# Start long-running connection
docker run mlanies/quic-test:latest \
  --mode=client \
  --server=<server-ip>:4433 \
  --duration=300s \
  --migrate-after=60s
```

**Questions:**
1. Does connection survive network change?
2. What is the migration latency?
3. How does QUIC maintain connection ID?

### Lab Report Template

```markdown
# Lab #1: QUIC Basics

**Student:** [Name]
**Date:** [Date]

## Exercise 1.1: First Connection
- Handshake duration: ___ ms
- RTT: ___ ms
- Observations: ___

## Exercise 1.2: 0-RTT
- First connection: ___ ms
- 0-RTT connection: ___ ms
- Speedup: ___x
- Security concerns: ___

## Exercise 1.3: Migration
- Migration successful: Yes/No
- Migration latency: ___ ms
- Observations: ___

## Conclusion
[Your analysis]
```

---

## Lab #2: Congestion Control

**Duration:** 3 hours  
**Level:** Graduate  
**Prerequisites:** TCP congestion control (Reno, CUBIC)

### Learning Objectives

1. Understand BBRv2 congestion control
2. Compare BBRv2 vs BBRv3
3. Analyze performance under packet loss
4. Measure fairness in multi-flow scenarios

### Exercise 2.1: BBRv2 Basics

**Task:** Measure BBRv2 performance under various conditions.

```bash
# Low latency, low loss
docker run mlanies/quic-test:latest \
  --mode=client \
  --congestion=bbrv2 \
  --profile=fiber \
  --duration=60s

# High latency, high loss
docker run mlanies/quic-test:latest \
  --mode=client \
  --congestion=bbrv2 \
  --profile=satellite \
  --duration=60s
```

**Questions:**
1. How does BBRv2 estimate bandwidth?
2. What is the probe RTT phase?
3. How does BBRv2 handle packet loss?

### Exercise 2.2: BBRv2 vs BBRv3

**Task:** Compare BBRv2 and BBRv3 performance.

```bash
# BBRv2
docker run mlanies/quic-test:latest \
  --mode=client \
  --congestion=bbrv2 \
  --profile=mobile \
  --duration=120s \
  --output=json > bbrv2.json

# BBRv3
docker run mlanies/quic-test:latest \
  --mode=client \
  --congestion=bbrv3 \
  --profile=mobile \
  --duration=120s \
  --output=json > bbrv3.json

# Compare results
python scripts/compare_results.py bbrv2.json bbrv3.json
```

**Questions:**
1. Which performs better under packet loss?
2. Which has lower latency?
3. Which is more fair in multi-flow scenarios?

### Exercise 2.3: Fairness Analysis

**Task:** Measure fairness between multiple flows.

```bash
# Start 4 concurrent flows
for i in {1..4}; do
  docker run -d --name flow-$i \
    mlanies/quic-test:latest \
    --mode=client \
    --congestion=bbrv2 \
    --duration=180s \
    --prometheus-port=909$i
done

# Collect metrics
python scripts/analyze_fairness.py
```

**Questions:**
1. What is the Jain's fairness index?
2. Are flows equally sharing bandwidth?
3. How does BBR achieve fairness?

---

## Lab #3: Performance Analysis

**Duration:** 3 hours  
**Level:** Undergraduate/Graduate  
**Prerequisites:** Basic statistics, network emulation

### Learning Objectives

1. Compare QUIC vs TCP performance
2. Analyze impact of packet loss
3. Measure benefits of multiplexing
4. Understand FEC (Forward Error Correction)

### Exercise 3.1: QUIC vs TCP

**Task:** Compare QUIC and TCP under identical conditions.

```bash
# QUIC + TCP comparison
docker run mlanies/quic-test:latest \
  --mode=client \
  --compare-tcp \
  --profile=mobile \
  --duration=120s \
  --output=json > comparison.json

# Visualize results
python scripts/visualize_comparison.py comparison.json
```

**Questions:**
1. Which has lower latency?
2. Which has higher throughput?
3. How does head-of-line blocking affect TCP?

### Exercise 3.2: Packet Loss Impact

**Task:** Measure performance degradation under packet loss.

```bash
# Test with increasing loss rates
for loss in 0 1 5 10; do
  docker run mlanies/quic-test:latest \
    --mode=client \
    --profile=custom \
    --loss=${loss}% \
    --duration=60s \
    --output=json > loss-${loss}.json
done

# Plot results
python scripts/plot_loss_impact.py loss-*.json
```

**Questions:**
1. At what loss rate does QUIC outperform TCP?
2. How does retransmission affect latency?
3. What is the optimal FEC redundancy?

### Exercise 3.3: FEC Analysis

**Task:** Analyze Forward Error Correction benefits.

```bash
# Without FEC
docker run mlanies/quic-test:latest \
  --mode=client \
  --profile=custom \
  --loss=10% \
  --duration=60s \
  --output=json > no-fec.json

# With FEC
docker run mlanies/quic-test:latest \
  --mode=client \
  --profile=custom \
  --loss=10% \
  --fec=true \
  --fec-redundancy=0.15 \
  --duration=60s \
  --output=json > with-fec.json

# Compare
python scripts/compare_fec.py no-fec.json with-fec.json
```

**Questions:**
1. How much does FEC reduce retransmissions?
2. What is the overhead of FEC?
3. What is the optimal redundancy ratio?

---

## Additional Resources

### Datasets

Pre-collected datasets for analysis:

```bash
# Download sample datasets
wget https://cloudbridge-research.ru/datasets/quic-test-samples.tar.gz
tar -xzf quic-test-samples.tar.gz
```

**Included:**
- `mobile-4g/` — 4G LTE measurements
- `satellite/` — Satellite link measurements
- `fiber/` — Fiber optic measurements
- `high-loss/` — High packet loss scenarios

### Analysis Scripts

```bash
# Clone analysis scripts
git clone https://github.com/twogc/quic-test-analysis
cd quic-test-analysis

# Install dependencies
pip install -r requirements.python
```

**Available scripts:**
- `compare_results.py` — Compare two test runs
- `visualize_comparison.py` — Generate comparison plots
- `plot_loss_impact.py` — Plot loss vs performance
- `analyze_fairness.py` — Calculate fairness metrics
- `compare_fec.py` — Analyze FEC benefits

### Lecture Slides

Download lecture slides (PDF):

- [QUIC Protocol Overview](https://cloudbridge-research.ru/lectures/quic-overview.pdf)
- [Congestion Control in QUIC](https://cloudbridge-research.ru/lectures/quic-cc.pdf)
- [Performance Analysis](https://cloudbridge-research.ru/lectures/quic-performance.pdf)

### Video Tutorials

- [Getting Started with quic-test](https://youtube.com/watch?v=...)
- [BBRv2 vs BBRv3 Explained](https://youtube.com/watch?v=...)
- [QUIC Performance Tuning](https://youtube.com/watch?v=...)

## Grading Rubric

### Lab Reports (100 points)

- **Methodology (20 points)**
  - Clear description of setup
  - Reproducible commands
  - Proper network configuration

- **Data Collection (20 points)**
  - Complete measurements
  - Proper duration
  - Multiple runs for statistical significance

- **Analysis (30 points)**
  - Correct interpretation of results
  - Statistical analysis
  - Comparison with theory

- **Presentation (20 points)**
  - Clear graphs and tables
  - Proper labeling
  - Professional formatting

- **Conclusions (10 points)**
  - Insights from data
  - Limitations discussed
  - Future work suggested

## Contact for Educators

Interested in using `quic-test` in your course?

- **Email:** education@cloudbridge-research.ru
- **Telegram:** @cloudbridge_edu
- **GitHub Discussions:** [Education](https://github.com/twogc/quic-test/discussions/categories/education)

We provide:
- Free access to cloud infrastructure
- Custom lab materials
- Technical support
- Guest lectures (online)

## See Also

- [CLI Reference](cli.md)
- [Architecture](architecture.md)
- [Case Studies](case-studies.md)
- [AI Integration](ai-routing-integration.md)
