//! Bridge module for Go integration
//! 
//! Provides FFI functions and HTTP API for communication with Go QUIC test

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::sync::{Arc, Mutex};
use tokio::sync::broadcast;
use warp::Filter;

use crate::metrics::{QUICMetrics, update_metrics, get_current_metrics};

/// HTTP API request structure
#[derive(Debug, Deserialize, Serialize)]
pub struct MetricsRequest {
    pub latency: f64,
    pub throughput: f64,
    pub connections: i32,
    pub errors: i32,
    pub packet_loss: f64,
    pub retransmits: i32,
}

/// HTTP API response structure
#[derive(Debug, Deserialize, Serialize)]
pub struct MetricsResponse {
    pub status: String,
    pub message: Option<String>,
    pub metrics: Option<QUICMetrics>,
}

/// Bridge state for Go integration
pub struct GoBridge {
    metrics_sender: broadcast::Sender<QUICMetrics>,
}

impl GoBridge {
    pub fn new() -> Self {
        let (tx, _) = broadcast::channel(1000);
        Self {
            metrics_sender: tx,
        }
    }

    /// Update metrics from Go
    pub fn update_metrics(&self, req: MetricsRequest) -> Result<()> {
        let metrics = QUICMetrics {
            latency: req.latency,
            throughput: req.throughput,
            connections: req.connections,
            errors: req.errors,
            packet_loss: req.packet_loss,
            retransmits: req.retransmits,
            timestamp: chrono::Utc::now(),
        };

        // Update global metrics
        update_metrics(metrics.clone())?;

        // Broadcast to subscribers
        let _ = self.metrics_sender.send(metrics);

        Ok(())
    }

    /// Get current metrics
    pub fn get_current_metrics(&self) -> Option<QUICMetrics> {
        get_current_metrics()
    }

    /// Subscribe to metrics updates
    pub fn subscribe(&self) -> broadcast::Receiver<QUICMetrics> {
        self.metrics_sender.subscribe()
    }
}

/// FFI function to update metrics from Go
#[no_mangle]
pub extern "C" fn update_quic_metrics_ffi(
    latency: f64,
    throughput: f64,
    connections: i32,
    errors: i32,
    packet_loss: f64,
    retransmits: i32,
) -> i32 {
    log::debug!(
        "FFI: Updating QUIC metrics: latency={}, throughput={}, connections={}, errors={}, loss={}, retransmits={}",
        latency, throughput, connections, errors, packet_loss, retransmits
    );
    
    let metrics = QUICMetrics {
        latency,
        throughput,
        connections,
        errors,
        packet_loss,
        retransmits,
        timestamp: chrono::Utc::now(),
    };

    match update_metrics(metrics) {
        Ok(_) => 0,
        Err(e) => {
            log::error!("FFI: Failed to update metrics: {}", e);
            -1
        }
    }
}

/// FFI function to get current metrics
#[no_mangle]
pub extern "C" fn get_quic_metrics_ffi() -> *mut QUICMetrics {
    match get_current_metrics() {
        Some(metrics) => {
            let boxed = Box::new(metrics);
            Box::into_raw(boxed)
        }
        None => std::ptr::null_mut(),
    }
}

/// FFI function to free metrics memory
#[no_mangle]
pub extern "C" fn free_quic_metrics_ffi(ptr: *mut QUICMetrics) {
    if !ptr.is_null() {
        unsafe {
            let _ = Box::from_raw(ptr);
        }
    }
}

/// Create HTTP API routes for Go integration
pub fn create_api_routes() -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    let metrics_update = warp::path("metrics")
        .and(warp::post())
        .and(warp::body::json())
        .map(|req: MetricsRequest| {
            // Update metrics
            let metrics = QUICMetrics {
                latency: req.latency,
                throughput: req.throughput,
                connections: req.connections,
                errors: req.errors,
                packet_loss: req.packet_loss,
                retransmits: req.retransmits,
                timestamp: chrono::Utc::now(),
            };

            match update_metrics(metrics) {
                Ok(_) => {
                    let response = MetricsResponse {
                        status: "ok".to_string(),
                        message: Some("Metrics updated successfully".to_string()),
                        metrics: None,
                    };
                    warp::reply::json(&response)
                }
                Err(e) => {
                    let response = MetricsResponse {
                        status: "error".to_string(),
                        message: Some(format!("Failed to update metrics: {}", e)),
                        metrics: None,
                    };
                    warp::reply::json(&response)
                }
            }
        });

    let metrics_get = warp::path("metrics")
        .and(warp::get())
        .map(|| {
            match get_current_metrics() {
                Some(metrics) => {
                    let response = MetricsResponse {
                        status: "ok".to_string(),
                        message: None,
                        metrics: Some(metrics),
                    };
                    warp::reply::json(&response)
                }
                None => {
                    let response = MetricsResponse {
                        status: "error".to_string(),
                        message: Some("No metrics available".to_string()),
                        metrics: None,
                    };
                    warp::reply::json(&response)
                }
            }
        });

    let health = warp::path("health")
        .and(warp::get())
        .map(|| {
            warp::reply::json(&serde_json::json!({
                "status": "ok",
                "service": "quic-bottom",
                "version": env!("CARGO_PKG_VERSION")
            }))
        });

    metrics_update.or(metrics_get).or(health)
}

/// Start HTTP API server
pub async fn start_api_server(port: u16) -> Result<()> {
    let routes = create_api_routes();
    
    log::info!("Starting HTTP API server on port {}", port);
    warp::serve(routes)
        .run(([127, 0, 0, 1], port))
        .await;
    
    Ok(())
}
