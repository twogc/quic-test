//! Simple test for BBRv3 API endpoint
//! This binary starts only the HTTP API server without TUI

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::sync::{Arc, Mutex};
use warp::Filter;

/// Real-time QUIC metrics from Go application
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RealQUICMetrics {
    pub timestamp: u64,
    pub latency: f64,
    pub throughput: f64,
    pub connections: i32,
    pub errors: i32,
    pub packet_loss: f64,
    pub retransmits: i32,
    pub jitter: f64,
    pub congestion_window: i32,
    pub rtt: f64,
    pub bytes_received: i64,
    pub bytes_sent: i64,
    pub streams: i32,
    pub handshake_time: f64,

    // BBRv3 specific metrics
    #[serde(default)]
    pub bbrv3_phase: Option<String>,
    #[serde(default)]
    pub bbrv3_bw_fast: Option<f64>,
    #[serde(default)]
    pub bbrv3_bw_slow: Option<f64>,
    #[serde(default)]
    pub bbrv3_loss_rate_round: Option<f64>,
    #[serde(default)]
    pub bbrv3_loss_rate_ema: Option<f64>,
    #[serde(default)]
    pub bbrv3_loss_threshold: Option<f64>,
    #[serde(default)]
    pub bbrv3_headroom_usage: Option<f64>,
    #[serde(default)]
    pub bbrv3_inflight_target: Option<f64>,
    #[serde(default)]
    pub bbrv3_pacing_quantum: Option<i64>,
    #[serde(default)]
    pub bbrv3_pacing_gain: Option<f64>,
    #[serde(default)]
    pub bbrv3_cwnd_gain: Option<f64>,
    #[serde(default)]
    pub bbrv3_probe_rtt_min_ms: Option<f64>,
    #[serde(default)]
    pub bbrv3_bufferbloat_factor: Option<f64>,
    #[serde(default)]
    pub bbrv3_stability_index: Option<f64>,
    #[serde(default)]
    pub bbrv3_phase_duration_ms: Option<std::collections::HashMap<String, f64>>,
    #[serde(default)]
    pub bbrv3_recovery_time_ms: Option<f64>,
    #[serde(default)]
    pub bbrv3_loss_recovery_efficiency: Option<f64>,
}

#[tokio::main]
async fn main() -> Result<()> {
    let current_metrics: Arc<Mutex<Option<RealQUICMetrics>>> = Arc::new(Mutex::new(None));
    let metrics_history: Arc<Mutex<Vec<RealQUICMetrics>>> = Arc::new(Mutex::new(Vec::new()));

    println!("Starting BBRv3 API Test Server...");
    println!("HTTP API listening on http://127.0.0.1:8080");
    println!("\nAvailable endpoints:");
    println!("  POST http://127.0.0.1:8080/api/metrics - Send metrics");
    println!("  GET  http://127.0.0.1:8080/health - Health check");
    println!("  GET  http://127.0.0.1:8080/api/current - Get current metrics");
    println!("\nTo test, run in another terminal:");
    println!("  curl -X POST http://127.0.0.1:8080/health");
    println!("  curl -X POST http://127.0.0.1:8080/api/metrics -H 'Content-Type: application/json' -d '{{...}}'");
    println!("\nPress Ctrl+C to stop.\n");

    // HTTP API routes
    let current_metrics_post = Arc::clone(&current_metrics);
    let metrics_filter = warp::path("api")
        .and(warp::path("metrics"))
        .and(warp::post())
        .and(warp::body::json())
        .map(move |metrics: RealQUICMetrics| {
            println!("\n📊 Received metrics:");
            println!("  Phase: {:?}", metrics.bbrv3_phase);
            println!("  Bandwidth Fast: {:?} bps", metrics.bbrv3_bw_fast);
            println!("  Bandwidth Slow: {:?} bps", metrics.bbrv3_bw_slow);
            println!("  Loss Rate: {:?}%", metrics.bbrv3_loss_rate_ema.map(|x| x * 100.0));
            println!("  Bufferbloat: {:?}", metrics.bbrv3_bufferbloat_factor);
            println!("  Stability Index: {:?}", metrics.bbrv3_stability_index);
            println!("  Pacing Gain: {:?}", metrics.bbrv3_pacing_gain);
            println!("  CWND Gain: {:?}", metrics.bbrv3_cwnd_gain);
            println!("  Recovery Time: {:?} ms", metrics.bbrv3_recovery_time_ms);
            println!("  Recovery Efficiency: {:?}", metrics.bbrv3_loss_recovery_efficiency);

            // Update current metrics
            {
                let mut current = current_metrics_post.lock().unwrap();
                *current = Some(metrics.clone());
            }

            warp::reply::json(&serde_json::json!({"status": "ok", "message": "BBRv3 metrics received"}))
        });

    let health_filter = warp::path("health")
        .map(|| {
            println!("✅ Health check");
            warp::reply::json(&serde_json::json!({"status": "healthy"}))
        });

    let current_metrics_get = Arc::clone(&current_metrics);
    let current_filter = warp::path("api")
        .and(warp::path("current"))
        .and(warp::get())
        .map(move || {
            let current = current_metrics_get.lock().unwrap();
            println!("📋 Getting current metrics");
            warp::reply::json(&*current)
        });

    let routes = metrics_filter
        .or(health_filter)
        .or(current_filter);

    warp::serve(routes)
        .run(([127, 0, 0, 1], 8080))
        .await;

    Ok(())
}
