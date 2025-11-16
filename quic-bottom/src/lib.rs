//! QUIC Bottom - QUIC-specific TUI monitor based on bottom
//! 
//! This library provides QUIC-specific widgets and monitoring capabilities
//! for the bottom TUI framework.

pub mod app;
pub mod widgets;
pub mod metrics;
pub mod bridge;
pub mod config;
pub mod demo_data;
pub mod improved_layout;
// pub mod professional_graphs; // Temporarily disabled due to compilation errors
pub mod simple_professional;
pub mod heatmap_widget;
pub mod correlation_widget;
pub mod anomaly_detection;

// Re-export key types
pub use metrics::QUICMetrics;
pub use config::QuicBottomConfig;

use anyhow::Result;
use tokio::runtime::Runtime;

/// Initialize the QUIC Bottom application
pub fn init_quic_bottom() -> Result<()> {
    env_logger::init();
    log::info!("Initializing QUIC Bottom");
    Ok(())
}

/// Start the QUIC Bottom TUI
pub fn start_quic_bottom(interval_ms: u64) -> Result<()> {
    let rt = Runtime::new()?;
    rt.block_on(async {
        let mut app = app::QuicBottomApp::new(interval_ms).await?;
        app.run().await?;
        Ok::<(), anyhow::Error>(())
    })
}

/// FFI function to update QUIC metrics from Go
#[no_mangle]
pub extern "C" fn update_quic_metrics(
    latency: f64,
    throughput: f64,
    connections: i32,
    errors: i32,
    packet_loss: f64,
    retransmits: i32,
) -> i32 {
    log::debug!(
        "Updating QUIC metrics: latency={}, throughput={}, connections={}, errors={}, loss={}, retransmits={}",
        latency, throughput, connections, errors, packet_loss, retransmits
    );
    
    // Update global metrics state
    if let Err(e) = metrics::update_metrics(metrics::QUICMetrics {
        latency,
        throughput,
        connections,
        errors,
        packet_loss,
        retransmits,
        timestamp: chrono::Utc::now(),
    }) {
        log::error!("Failed to update metrics: {}", e);
        return -1;
    }
    
    0
}

/// FFI function to get current metrics
#[no_mangle]
pub extern "C" fn get_quic_metrics() -> *mut metrics::QUICMetrics {
    match metrics::get_current_metrics() {
        Some(metrics) => {
            let boxed = Box::new(metrics);
            Box::into_raw(boxed)
        }
        None => std::ptr::null_mut(),
    }
}

/// FFI function to free metrics memory
#[no_mangle]
pub extern "C" fn free_quic_metrics(ptr: *mut metrics::QUICMetrics) {
    if !ptr.is_null() {
        unsafe {
            let _ = Box::from_raw(ptr);
        }
    }
}
