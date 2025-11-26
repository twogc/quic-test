# Roadmap

Honest roadmap for `quic-test` development.

## Current Status (v0.9.x)

### ✅ Stable Features
- QUIC client/server implementation
- RTT, jitter, throughput measurements
- Network profile emulation
- Prometheus metrics export
- TUI visualization (`quic-bottom`)
- BBRv2 congestion control
- Docker support

### ⚗️ Experimental (use with caution)
- BBRv3 congestion control
- Forward Error Correction (FEC) with AVX2
- MASQUE VPN testing
- TCP-over-QUIC tunneling
- ICE/STUN/TURN tests

## v1.0.0 (Target: Q2 2025)

**Goal:** Production-ready release for educational and research use

### Must-Have
- [ ] Comprehensive test coverage (>80%)
- [ ] Full documentation
- [ ] Stable API
- [ ] Performance benchmarks
- [ ] Security audit
- [ ] Multi-platform binaries (Linux, macOS, Windows)

### Nice-to-Have
- [ ] Web dashboard UI
- [ ] Automated report generation
- [ ] Integration tests with real networks

## v1.1.0 (Target: Q3 2025)

**Goal:** HTTP/3 and WebTransport support

### Features
- [ ] HTTP/3 load testing
- [ ] WebTransport client/server
- [ ] gRPC-over-QUIC testing
- [ ] Advanced congestion control (CUBIC, Reno)

## v1.2.0 (Target: Q4 2025)

**Goal:** Cloud and automation

### Features
- [ ] Multi-cloud deployment (AWS, GCP, Azure)
- [ ] Kubernetes operator
- [ ] CI/CD integration (GitHub Actions, GitLab CI)
- [ ] Automated anomaly detection

## v2.0.0 (Target: 2026)

**Goal:** Enterprise features

### Features
- [ ] Distributed testing (multiple nodes)
- [ ] Real-time collaboration
- [ ] Advanced analytics (ML-based)
- [ ] Commercial support

## Experimental Features (No ETA)

These are research ideas, may or may not be implemented:

- **Quantum-resistant QUIC:** Testing with post-quantum cryptography
- **5G network emulation:** Specific 5G NR profiles
- **Satellite constellations:** Starlink/OneWeb emulation
- **AI-driven optimization:** Automatic parameter tuning

## Won't Implement

Features we explicitly decided NOT to implement:

- **Full HTTP/2 support:** Out of scope, use existing tools
- **VPN replacement:** MASQUE testing only, not production VPN
- **DDoS testing:** Ethical concerns, use specialized tools

## Community Requests

Vote for features on [GitHub Discussions](https://github.com/twogc/quic-test/discussions).

Top requests:
1. Windows native support (currently Docker only)
2. GUI application
3. Mobile app (iOS/Android)
4. Browser extension

## Contributing

Want to help? See [CONTRIBUTING.md](../CONTRIBUTING.md).

Priority areas:
- Documentation improvements
- Test coverage
- Bug fixes
- Performance optimization

## Versioning

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR:** Breaking changes
- **MINOR:** New features (backward compatible)
- **PATCH:** Bug fixes

## Release Schedule

- **Patch releases:** Monthly (bug fixes)
- **Minor releases:** Quarterly (new features)
- **Major releases:** Yearly (breaking changes)

## Deprecation Policy

- Features marked deprecated: 6 months notice
- Breaking changes: Only in major versions
- Security fixes: Immediate release

## See Also

- [Architecture](architecture.md)
- [Contributing](../CONTRIBUTING.md)
- [Changelog](../CHANGELOG.md)
