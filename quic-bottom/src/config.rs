//! Configuration module for QUIC Bottom
//! 
//! Handles configuration loading and management

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::path::Path;

/// QUIC Bottom configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct QuicBottomConfig {
    /// Update interval in milliseconds
    pub update_interval: u64,
    
    /// HTTP API port for Go integration
    pub api_port: u16,
    
    /// Maximum data points for time series
    pub max_data_points: usize,
    
    /// Widget configuration
    pub widgets: WidgetConfig,
    
    /// Color theme
    pub colors: ColorConfig,
}

/// Widget-specific configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WidgetConfig {
    /// Latency widget settings
    pub latency: LatencyWidgetConfig,
    
    /// Throughput widget settings
    pub throughput: ThroughputWidgetConfig,
    
    /// Connection widget settings
    pub connections: ConnectionWidgetConfig,
    
    /// Network quality widget settings
    pub network: NetworkWidgetConfig,
}

/// Latency widget configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LatencyWidgetConfig {
    /// Enable latency widget
    pub enabled: bool,
    
    /// Maximum data points
    pub max_points: usize,
    
    /// Show percentiles
    pub show_percentiles: bool,
    
    /// Show jitter
    pub show_jitter: bool,
}

/// Throughput widget configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ThroughputWidgetConfig {
    /// Enable throughput widget
    pub enabled: bool,
    
    /// Maximum data points
    pub max_points: usize,
    
    /// Show average
    pub show_average: bool,
    
    /// Show maximum
    pub show_maximum: bool,
}

/// Connection widget configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConnectionWidgetConfig {
    /// Enable connection widget
    pub enabled: bool,
    
    /// Show handshake times
    pub show_handshake_times: bool,
    
    /// Show success rate
    pub show_success_rate: bool,
}

/// Network quality widget configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NetworkWidgetConfig {
    /// Enable network widget
    pub enabled: bool,
    
    /// Show packet loss graph
    pub show_loss_graph: bool,
    
    /// Show retransmit graph
    pub show_retransmit_graph: bool,
    
    /// Show congestion control
    pub show_congestion_control: bool,
}

/// Color configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ColorConfig {
    /// Primary color
    pub primary: String,
    
    /// Secondary color
    pub secondary: String,
    
    /// Accent color
    pub accent: String,
    
    /// Success color
    pub success: String,
    
    /// Warning color
    pub warning: String,
    
    /// Error color
    pub error: String,
}

impl Default for QuicBottomConfig {
    fn default() -> Self {
        Self {
            update_interval: 100,
            api_port: 8080,
            max_data_points: 1000,
            widgets: WidgetConfig::default(),
            colors: ColorConfig::default(),
        }
    }
}

impl Default for WidgetConfig {
    fn default() -> Self {
        Self {
            latency: LatencyWidgetConfig::default(),
            throughput: ThroughputWidgetConfig::default(),
            connections: ConnectionWidgetConfig::default(),
            network: NetworkWidgetConfig::default(),
        }
    }
}

impl Default for LatencyWidgetConfig {
    fn default() -> Self {
        Self {
            enabled: true,
            max_points: 1000,
            show_percentiles: true,
            show_jitter: true,
        }
    }
}

impl Default for ThroughputWidgetConfig {
    fn default() -> Self {
        Self {
            enabled: true,
            max_points: 1000,
            show_average: true,
            show_maximum: true,
        }
    }
}

impl Default for ConnectionWidgetConfig {
    fn default() -> Self {
        Self {
            enabled: true,
            show_handshake_times: true,
            show_success_rate: true,
        }
    }
}

impl Default for NetworkWidgetConfig {
    fn default() -> Self {
        Self {
            enabled: true,
            show_loss_graph: true,
            show_retransmit_graph: true,
            show_congestion_control: true,
        }
    }
}

impl Default for ColorConfig {
    fn default() -> Self {
        Self {
            primary: "blue".to_string(),
            secondary: "green".to_string(),
            accent: "yellow".to_string(),
            success: "green".to_string(),
            warning: "yellow".to_string(),
            error: "red".to_string(),
        }
    }
}

impl QuicBottomConfig {
    /// Load configuration from file
    pub fn load_from_file<P: AsRef<Path>>(path: P) -> Result<Self> {
        let content = std::fs::read_to_string(path)?;
        let config: QuicBottomConfig = toml::from_str(&content)?;
        Ok(config)
    }

    /// Save configuration to file
    pub fn save_to_file<P: AsRef<Path>>(&self, path: P) -> Result<()> {
        let content = toml::to_string_pretty(self)?;
        std::fs::write(path, content)?;
        Ok(())
    }

    /// Create default configuration file
    pub fn create_default_config<P: AsRef<Path>>(path: P) -> Result<()> {
        let config = QuicBottomConfig::default();
        config.save_to_file(path)?;
        Ok(())
    }
}
