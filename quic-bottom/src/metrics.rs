//! QUIC metrics handling and data structures

use serde::{Deserialize, Serialize};
use std::sync::{Arc, Mutex, RwLock};
use std::collections::VecDeque;
use chrono::{DateTime, Utc};

/// QUIC-specific metrics
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct QUICMetrics {
    pub latency: f64,
    pub throughput: f64,
    pub connections: i32,
    pub errors: i32,
    pub packet_loss: f64,
    pub retransmits: i32,
    pub timestamp: DateTime<Utc>,
}

/// Time series data for graphs
#[derive(Debug, Clone)]
pub struct TimeSeriesData {
    pub latency: VecDeque<f64>,
    pub throughput: VecDeque<f64>,
    pub packet_loss: VecDeque<f64>,
    pub retransmits: VecDeque<i32>,
    pub max_points: usize,
}

impl TimeSeriesData {
    pub fn new(max_points: usize) -> Self {
        Self {
            latency: VecDeque::with_capacity(max_points),
            throughput: VecDeque::with_capacity(max_points),
            packet_loss: VecDeque::with_capacity(max_points),
            retransmits: VecDeque::with_capacity(max_points),
            max_points,
        }
    }

    pub fn add_data_point(&mut self, metrics: &QUICMetrics) {
        // Add new data points
        self.latency.push_back(metrics.latency);
        self.throughput.push_back(metrics.throughput);
        self.packet_loss.push_back(metrics.packet_loss);
        self.retransmits.push_back(metrics.retransmits);

        // Remove old data points if we exceed max_points
        if self.latency.len() > self.max_points {
            self.latency.pop_front();
        }
        if self.throughput.len() > self.max_points {
            self.throughput.pop_front();
        }
        if self.packet_loss.len() > self.max_points {
            self.packet_loss.pop_front();
        }
        if self.retransmits.len() > self.max_points {
            self.retransmits.pop_front();
        }
    }

    pub fn get_latency_data(&self) -> Vec<f64> {
        self.latency.iter().cloned().collect()
    }

    pub fn get_throughput_data(&self) -> Vec<f64> {
        self.throughput.iter().cloned().collect()
    }

    pub fn get_packet_loss_data(&self) -> Vec<f64> {
        self.packet_loss.iter().cloned().collect()
    }

    pub fn get_retransmits_data(&self) -> Vec<i32> {
        self.retransmits.iter().cloned().collect()
    }
}

/// Global metrics state
static METRICS_STATE: Mutex<Option<Arc<RwLock<QUICMetricsState>>>> = Mutex::new(None);

#[derive(Debug)]
struct QUICMetricsState {
    current: QUICMetrics,
    time_series: TimeSeriesData,
}

impl QUICMetricsState {
    fn new() -> Self {
        Self {
            current: QUICMetrics {
                latency: 0.0,
                throughput: 0.0,
                connections: 0,
                errors: 0,
                packet_loss: 0.0,
                retransmits: 0,
                timestamp: Utc::now(),
            },
            time_series: TimeSeriesData::new(1000), // Keep last 1000 data points
        }
    }

    fn update(&mut self, metrics: QUICMetrics) {
        self.current = metrics.clone();
        self.time_series.add_data_point(&metrics);
    }

    fn get_current(&self) -> QUICMetrics {
        self.current.clone()
    }

    fn get_time_series(&self) -> TimeSeriesData {
        self.time_series.clone()
    }
}

/// Initialize the metrics system
pub fn init_metrics() -> Result<(), anyhow::Error> {
    let state = Arc::new(RwLock::new(QUICMetricsState::new()));
    let mut global_state = METRICS_STATE.lock().unwrap();
    *global_state = Some(state);
    Ok(())
}

/// Update QUIC metrics
pub fn update_metrics(metrics: QUICMetrics) -> Result<(), anyhow::Error> {
    let global_state = METRICS_STATE.lock().unwrap();
    if let Some(state) = global_state.as_ref() {
        let mut state_guard = state.write().unwrap();
        state_guard.update(metrics);
    }
    Ok(())
}

/// Get current metrics
pub fn get_current_metrics() -> Option<QUICMetrics> {
    let global_state = METRICS_STATE.lock().unwrap();
    if let Some(state) = global_state.as_ref() {
        let state_guard = state.read().unwrap();
        Some(state_guard.get_current())
    } else {
        None
    }
}

/// Get time series data
pub fn get_time_series_data() -> Option<TimeSeriesData> {
    let global_state = METRICS_STATE.lock().unwrap();
    if let Some(state) = global_state.as_ref() {
        let state_guard = state.read().unwrap();
        Some(state_guard.get_time_series())
    } else {
        None
    }
}

/// Calculate percentiles for latency data
pub fn calculate_latency_percentiles(data: &[f64]) -> (f64, f64, f64) {
    if data.is_empty() {
        return (0.0, 0.0, 0.0);
    }

    let mut sorted_data = data.to_vec();
    sorted_data.sort_by(|a, b| a.partial_cmp(b).unwrap());

    let len = sorted_data.len();
    let p50_idx = (len as f64 * 0.5) as usize;
    let p95_idx = (len as f64 * 0.95) as usize;
    let p99_idx = (len as f64 * 0.99) as usize;

    let p50 = sorted_data[p50_idx.min(len - 1)];
    let p95 = sorted_data[p95_idx.min(len - 1)];
    let p99 = sorted_data[p99_idx.min(len - 1)];

    (p50, p95, p99)
}

/// Calculate jitter (standard deviation) for latency data
pub fn calculate_jitter(data: &[f64]) -> f64 {
    if data.is_empty() {
        return 0.0;
    }

    let mean = data.iter().sum::<f64>() / data.len() as f64;
    let variance = data.iter()
        .map(|x| (x - mean).powi(2))
        .sum::<f64>() / data.len() as f64;
    
    variance.sqrt()
}
