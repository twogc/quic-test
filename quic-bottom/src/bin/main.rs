//! QUIC Bottom - Main entry point
//! 
//! A specialized version of bottom for monitoring QUIC protocol metrics

use anyhow::Result;
use clap::Parser;
use log::info;

// Modules are defined in lib.rs

use quic_bottom::app::QuicBottomApp;

#[derive(Parser)]
#[command(name = "quic-bottom")]
#[command(about = "QUIC Bottom - Real-time QUIC protocol monitor")]
#[command(version)]
struct Cli {
    /// Configuration file path
    #[arg(short, long, default_value = "~/.config/quic-bottom/config.toml")]
    config: Option<String>,
    
    /// Enable debug logging
    #[arg(short, long)]
    debug: bool,
    
    /// Update interval in milliseconds
    #[arg(short, long, default_value = "100")]
    interval: u64,
    
    /// HTTP API port for Go integration
    #[arg(long, default_value = "8080")]
    api_port: u16,
}

#[tokio::main]
async fn main() -> Result<()> {
    let cli = Cli::parse();
    
    // Initialize logging
    if cli.debug {
        env_logger::Builder::from_default_env()
            .filter_level(log::LevelFilter::Debug)
            .init();
    } else {
        env_logger::init();
    }
    
    info!("Starting QUIC Bottom v{}", env!("CARGO_PKG_VERSION"));
    info!("Debug mode: {}", cli.debug);
    info!("Update interval: {}ms", cli.interval);
    info!("API port: {}", cli.api_port);
    
    // Initialize metrics system
    quic_bottom::metrics::init_metrics()?;
    
    // Start HTTP API server for Go integration
    let api_port = cli.api_port;
    tokio::spawn(async move {
        if let Err(e) = start_api_server(api_port).await {
            log::error!("API server error: {}", e);
        }
    });
    
    // Create and run the application
    let mut app = QuicBottomApp::new(cli.interval).await?;
    app.run().await?;
    
    info!("QUIC Bottom stopped");
    Ok(())
}

async fn start_api_server(port: u16) -> Result<()> {
    // Используем create_api_routes из bridge.rs для поддержки POST /metrics
    let routes = quic_bottom::bridge::create_api_routes();
    
    info!("Starting API server on port {}", port);
    warp::serve(routes)
        .run(([127, 0, 0, 1], port))
        .await;
    
    Ok(())
}
