# QUIC Bottom

A specialized TUI monitor for QUIC protocol metrics, based on the excellent [bottom](https://github.com/ClementTsang/bottom) project.

## Features

- **Real-time QUIC monitoring** - Live metrics display
- **QUIC-specific widgets** - Latency, throughput, connections, network quality
- **HTTP API integration** - Seamless integration with Go QUIC test
- **Interactive TUI** - Modern terminal interface
- **Configurable** - Customizable widgets and colors

## Quick Start

### Prerequisites

- Rust 1.70+ (install from [rustup.rs](https://rustup.rs/))
- Go 1.25+ (for integration)

### Building

```bash
# Build QUIC Bottom
cd quic-bottom
cargo build --release

# Or use the provided script
../scripts/start-quic-bottom.sh
```

### Running

```bash
# Start QUIC Bottom TUI
./target/release/quic-bottom

# With custom options
./target/release/quic-bottom --api-port 8080 --interval 100 --debug
```

## Integration with Go QUIC Test

QUIC Bottom integrates seamlessly with the Go QUIC test project:

### 1. Start QUIC Bottom

```bash
# Terminal 1: Start QUIC Bottom
cd quic-bottom
./target/release/quic-bottom --api-port 8080
```

### 2. Run QUIC Test

```bash
# Terminal 2: Run QUIC test (will automatically send metrics to QUIC Bottom)
go run main.go --mode=test --connections=2 --streams=4
```

## Configuration

QUIC Bottom uses a TOML configuration file:

```toml
# config.toml
update_interval = 100
api_port = 8080
max_data_points = 1000

[widgets.latency]
enabled = true
max_points = 1000
show_percentiles = true
show_jitter = true

[widgets.throughput]
enabled = true
max_points = 1000
show_average = true
show_maximum = true

[colors]
primary = "blue"
secondary = "green"
accent = "yellow"
```

## Widgets

### Latency Widget
- Real-time RTT display
- Percentiles (P50, P95, P99)
- Jitter calculation
- Time series graph

### Throughput Widget
- Bandwidth monitoring
- Average and maximum values
- Time series graph

### Connection Widget
- Active/failed connections
- Success rate
- Handshake times

### Network Quality Widget
- Packet loss monitoring
- Retransmit tracking
- Congestion control display

## HTTP API

QUIC Bottom provides an HTTP API for integration:

### Endpoints

- `GET /health` - Health check
- `GET /metrics` - Get current metrics
- `POST /metrics` - Update metrics

### Example Usage

```bash
# Check health
curl http://localhost:8080/health

# Get current metrics
curl http://localhost:8080/metrics

# Update metrics
curl -X POST http://localhost:8080/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "latency": 10.5,
    "throughput": 1000.0,
    "connections": 2,
    "errors": 0,
    "packet_loss": 0.1,
    "retransmits": 5
  }'
```

## Keyboard Shortcuts

- `q` - Quit
- `r` - Refresh metrics
- `h` - Show help
- `Ctrl+C` - Quit

## Development

### Project Structure

```
quic-bottom/
├── src/
│   ├── bin/main.rs          # Main entry point
│   ├── lib.rs               # Library interface
│   ├── app/                 # Application logic
│   ├── widgets/             # QUIC-specific widgets
│   ├── metrics/             # Metrics handling
│   ├── bridge/              # Go integration
│   └── config/              # Configuration
├── config.toml              # Configuration file
└── Cargo.toml               # Dependencies
```

### Adding New Widgets

1. Create widget in `src/widgets/`
2. Add to `src/app/mod.rs`
3. Update configuration in `src/config.rs`

### Building for Development

```bash
# Debug build
cargo build

# Run with debug logging
RUST_LOG=debug cargo run

# Run tests
cargo test
```

## Troubleshooting

### Common Issues

1. **Build fails**: Ensure Rust 1.70+ is installed
2. **API connection fails**: Check if QUIC Bottom is running on port 8080
3. **No metrics displayed**: Verify Go integration is working

### Debug Mode

```bash
# Enable debug logging
RUST_LOG=debug ./target/release/quic-bottom

# Check API health
curl http://localhost:8080/health
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- Based on [bottom](https://github.com/ClementTsang/bottom) by Clement Tsang
- Inspired by [htop](https://github.com/htop-dev/htop) and [gotop](https://github.com/xxxserxxx/gotop)
- Built with [ratatui](https://github.com/ratatui-org/ratatui) TUI framework
