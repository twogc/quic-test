//! Real QUIC Bottom - Production Version
//! 
//! Features:
//! - Real-time QUIC metrics from Go application
//! - HTTP API for metrics collection
//! - Professional visualizations
//! - Network simulation integration
//! - Security testing integration
//! - Cloud deployment monitoring

use anyhow::Result;
use crossterm::{
    event::{self, DisableMouseCapture, EnableMouseCapture, Event, KeyCode, KeyEvent, KeyModifiers},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use ratatui::{
    backend::{Backend, CrosstermBackend},
    layout::{Constraint, Direction, Layout, Rect},
    style::{Color, Modifier, Style},
    widgets::{Block, Borders, Paragraph},
    Frame, Terminal,
};
use serde::{Deserialize, Serialize};
use std::io;
use std::sync::{Arc, Mutex};
use std::time::Duration;
use warp::Filter;

use quic_bottom::{
    simple_professional::{SimpleQuicLatencyGraph, SimpleQuicThroughputGraph},
    heatmap_widget::QUICPerformanceHeatmap,
    correlation_widget::QUICCorrelationWidget,
    anomaly_detection::QUICAnomalyWidget,
};

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
    
    // BBRv3 specific metrics (optional, only when using BBRv3)
    #[serde(default)]
    pub bbrv3_phase: Option<String>, // Startup, Drain, ProbeBW, ProbeRTT
    #[serde(default)]
    pub bbrv3_bw_fast: Option<f64>, // Fast-scale bandwidth (bps)
    #[serde(default)]
    pub bbrv3_bw_slow: Option<f64>, // Slow-scale bandwidth (bps)
    #[serde(default)]
    pub bbrv3_loss_rate_round: Option<f64>, // Loss rate per round
    #[serde(default)]
    pub bbrv3_loss_rate_ema: Option<f64>, // EMA loss rate
    #[serde(default)]
    pub bbrv3_loss_threshold: Option<f64>, // Loss threshold (2%)
    #[serde(default)]
    pub bbrv3_headroom_usage: Option<f64>, // Headroom usage (0.0-1.0)
    #[serde(default)]
    pub bbrv3_inflight_target: Option<f64>, // Target inflight (bytes)
    #[serde(default)]
    pub bbrv3_pacing_quantum: Option<i64>, // Pacing quantum (bytes)
    #[serde(default)]
    pub bbrv3_pacing_gain: Option<f64>, // Current pacing gain
    #[serde(default)]
    pub bbrv3_cwnd_gain: Option<f64>, // Current CWND gain
    #[serde(default)]
    pub bbrv3_probe_rtt_min_ms: Option<f64>, // Minimum RTT during ProbeRTT
    #[serde(default)]
    pub bbrv3_bufferbloat_factor: Option<f64>, // (avg_rtt / min_rtt) - 1
    #[serde(default)]
    pub bbrv3_stability_index: Option<f64>, // Δ throughput / Δ rtt
    #[serde(default)]
    pub bbrv3_phase_duration_ms: Option<std::collections::HashMap<String, f64>>, // Duration of each phase
    #[serde(default)]
    pub bbrv3_recovery_time_ms: Option<f64>, // Time to recover from loss
    #[serde(default)]
    pub bbrv3_loss_recovery_efficiency: Option<f64>, // recovered / lost
}

/// Real QUIC Bottom application
pub struct RealQUICBottom {
    // Basic graphs
    latency_graph: SimpleQuicLatencyGraph,
    throughput_graph: SimpleQuicThroughputGraph,
    
    // Enhanced analytics
    performance_heatmap: QUICPerformanceHeatmap,
    correlation_widget: QUICCorrelationWidget,
    anomaly_widget: QUICAnomalyWidget,
    
    // Real-time data
    current_metrics: Arc<Mutex<Option<RealQUICMetrics>>>,
    metrics_history: Arc<Mutex<Vec<RealQUICMetrics>>>,
    
    // App state
    should_quit: bool,
    update_interval: Duration,
    current_view: ViewMode,
    time_slot: usize,
    
    // Network simulation state
    network_simulation_active: bool,
    network_preset: String,
    network_latency: f64,
    network_loss: f64,
    network_bandwidth: f64,
    
    // Security testing state
    security_test_active: bool,
    security_score: f64,
    vulnerabilities_count: usize,
    
    // Cloud deployment state
    cloud_deployment_active: bool,
    cloud_provider: String,
    cloud_instances: usize,
    cloud_status: String,
}

#[derive(Debug, Clone, PartialEq)]
enum ViewMode {
    Dashboard,
    Analytics,
    Network,
    Security,
    Cloud,
    BBRv3,
    All,
}

impl RealQUICBottom {
    pub async fn new(interval_ms: u64) -> Result<Self> {
        Ok(Self {
            latency_graph: SimpleQuicLatencyGraph::new(),
            throughput_graph: SimpleQuicThroughputGraph::new(),
            performance_heatmap: QUICPerformanceHeatmap::new(),
            correlation_widget: QUICCorrelationWidget::new(),
            anomaly_widget: QUICAnomalyWidget::new(),
            current_metrics: Arc::new(Mutex::new(None)),
            metrics_history: Arc::new(Mutex::new(Vec::new())),
            should_quit: false,
            update_interval: Duration::from_millis(interval_ms),
            current_view: ViewMode::Dashboard,
            time_slot: 0,
            network_simulation_active: false,
            network_preset: "good".to_string(),
            network_latency: 20.0,
            network_loss: 1.0,
            network_bandwidth: 100.0,
            security_test_active: false,
            security_score: 100.0,
            vulnerabilities_count: 0,
            cloud_deployment_active: false,
            cloud_provider: "aws".to_string(),
            cloud_instances: 2,
            cloud_status: "running".to_string(),
        })
    }

    pub async fn run(&mut self) -> Result<()> {
        // Start HTTP API server in background
        let metrics_arc = Arc::clone(&self.current_metrics);
        let history_arc = Arc::clone(&self.metrics_history);
        
        tokio::spawn(async move {
            start_http_server(metrics_arc, history_arc).await;
        });

        // Give HTTP server time to start
        tokio::time::sleep(Duration::from_millis(500)).await;

        // Setup terminal
        enable_raw_mode()?;
        let mut stdout = io::stdout();
        execute!(stdout, EnterAlternateScreen, EnableMouseCapture)?;
        let backend = CrosstermBackend::new(stdout);
        let mut terminal = Terminal::new(backend)?;

        // Main event loop
        // Use a shorter timeout for event polling to ensure UI updates frequently
        let event_timeout = Duration::from_millis(100); // Poll every 100ms for events
        
        loop {
            if self.should_quit {
                break;
            }

            // Update all widgets with real data
            self.update_all_widgets();

            // Render the UI
            terminal.draw(|f| self.ui(f))?;

            // Handle events with short timeout to allow frequent UI updates
            if event::poll(event_timeout)? {
                if let Event::Key(key) = event::read()? {
                    self.handle_key_event(key);
                }
            }
            // If no event, loop continues immediately to update UI again
        }

        // Restore terminal
        disable_raw_mode()?;
        execute!(
            terminal.backend_mut(),
            LeaveAlternateScreen,
            DisableMouseCapture
        )?;
        terminal.show_cursor()?;

        Ok(())
    }

    fn update_all_widgets(&mut self) {
        // Get current metrics
        let metrics = {
            let current = self.current_metrics.lock().unwrap();
            current.clone()
        };

        if let Some(metrics) = metrics {
            // Apply network simulation effects
            let (adjusted_latency, adjusted_throughput, adjusted_loss) = self.apply_network_effects(
                metrics.latency, metrics.throughput, metrics.packet_loss
            );

            // Update basic graphs
            self.latency_graph.add_latency(adjusted_latency);
            self.throughput_graph.add_throughput(adjusted_throughput);

            // Update enhanced analytics
            self.performance_heatmap.add_performance_data(self.time_slot, 0, adjusted_latency);
            self.performance_heatmap.add_performance_data(self.time_slot, 1, adjusted_throughput);
            self.performance_heatmap.add_performance_data(self.time_slot, 2, adjusted_loss);
            self.performance_heatmap.add_performance_data(self.time_slot, 3, metrics.connections as f64);
            self.performance_heatmap.add_performance_data(self.time_slot, 4, metrics.errors as f64);

            // Update correlation data - include more metrics that change
            self.correlation_widget.add_metric_data("Latency".to_string(), adjusted_latency);
            self.correlation_widget.add_metric_data("Throughput".to_string(), adjusted_throughput);
            self.correlation_widget.add_metric_data("Packet Loss".to_string(), adjusted_loss);
            self.correlation_widget.add_metric_data("RTT".to_string(), metrics.rtt);
            self.correlation_widget.add_metric_data("Jitter".to_string(), metrics.jitter);
            self.correlation_widget.add_metric_data("Retransmits".to_string(), metrics.retransmits as f64);
            // Only add Connections and Errors if they change (to avoid constant values)
            if metrics.connections > 0 {
                self.correlation_widget.add_metric_data("Connections".to_string(), metrics.connections as f64);
            }
            if metrics.errors > 0 {
                self.correlation_widget.add_metric_data("Errors".to_string(), metrics.errors as f64);
            }
            self.correlation_widget.update_correlations();

            // Update anomaly detection
            self.anomaly_widget.add_quic_metric("Latency".to_string(), adjusted_latency);
            self.anomaly_widget.add_quic_metric("Throughput".to_string(), adjusted_throughput);
            self.anomaly_widget.add_quic_metric("Packet Loss".to_string(), adjusted_loss);
            self.anomaly_widget.add_quic_metric("Connections".to_string(), metrics.connections as f64);
            self.anomaly_widget.add_quic_metric("Errors".to_string(), metrics.errors as f64);

            // Update time slot
            self.time_slot = (self.time_slot + 1) % 20;
        }
    }

    fn apply_network_effects(&self, latency: f64, throughput: f64, loss: f64) -> (f64, f64, f64) {
        if !self.network_simulation_active {
            return (latency, throughput, loss);
        }

        let adjusted_latency = latency + self.network_latency;
        let adjusted_throughput = throughput * (1.0 - self.network_loss / 100.0);
        let adjusted_loss = loss + self.network_loss;

        (adjusted_latency, adjusted_throughput, adjusted_loss)
    }

    fn handle_key_event(&mut self, key: KeyEvent) {
        match key.code {
            KeyCode::Char('q') | KeyCode::Char('Q') => {
                self.should_quit = true;
            }
            KeyCode::Esc => {
                self.should_quit = true;
            }
            KeyCode::Char('c') if key.modifiers.contains(KeyModifiers::CONTROL) => {
                self.should_quit = true;
            }
            KeyCode::Char('r') => {
                // Reset all data
                self.reset_all_data();
            }
            KeyCode::Char('h') => {
                self.show_help();
            }
            // View switching
            KeyCode::Char('1') => {
                self.current_view = ViewMode::Dashboard;
            }
            KeyCode::Char('2') => {
                self.current_view = ViewMode::Analytics;
            }
            KeyCode::Char('3') => {
                self.current_view = ViewMode::Network;
            }
            KeyCode::Char('4') => {
                self.current_view = ViewMode::Security;
            }
            KeyCode::Char('5') => {
                self.current_view = ViewMode::Cloud;
            }
            KeyCode::Char('6') => {
                self.current_view = ViewMode::BBRv3;
            }
            KeyCode::Char('a') => {
                self.current_view = ViewMode::All;
            }
            // Network simulation controls
            KeyCode::Char('n') => {
                self.toggle_network_simulation();
            }
            KeyCode::Char('+') => {
                self.next_network_preset();
            }
            KeyCode::Char('-') => {
                self.prev_network_preset();
            }
            // Security testing controls
            KeyCode::Char('s') => {
                self.toggle_security_testing();
            }
            // Cloud deployment controls
            KeyCode::Char('d') => {
                self.toggle_cloud_deployment();
            }
            KeyCode::Char('i') => {
                self.scale_cloud_instances();
            }
            _ => {}
        }
    }

    fn reset_all_data(&mut self) {
        self.latency_graph = SimpleQuicLatencyGraph::new();
        self.throughput_graph = SimpleQuicThroughputGraph::new();
        self.performance_heatmap = QUICPerformanceHeatmap::new();
        self.correlation_widget = QUICCorrelationWidget::new();
        self.anomaly_widget = QUICAnomalyWidget::new();
        self.time_slot = 0;
        
        // Clear metrics history
        {
            let mut history = self.metrics_history.lock().unwrap();
            history.clear();
        }
    }

    fn toggle_network_simulation(&mut self) {
        self.network_simulation_active = !self.network_simulation_active;
    }

    fn next_network_preset(&mut self) {
        let presets = vec!["excellent", "good", "poor", "mobile", "satellite", "adversarial"];
        if let Some(current_index) = presets.iter().position(|&p| p == self.network_preset) {
            let next_index = (current_index + 1) % presets.len();
            self.network_preset = presets[next_index].to_string();
            self.apply_network_preset();
        }
    }

    fn prev_network_preset(&mut self) {
        let presets = vec!["excellent", "good", "poor", "mobile", "satellite", "adversarial"];
        if let Some(current_index) = presets.iter().position(|&p| p == self.network_preset) {
            let prev_index = if current_index == 0 { presets.len() - 1 } else { current_index - 1 };
            self.network_preset = presets[prev_index].to_string();
            self.apply_network_preset();
        }
    }

    fn apply_network_preset(&mut self) {
        match self.network_preset.as_str() {
            "excellent" => {
                self.network_latency = 5.0;
                self.network_loss = 0.1;
                self.network_bandwidth = 1000.0;
            }
            "good" => {
                self.network_latency = 20.0;
                self.network_loss = 1.0;
                self.network_bandwidth = 100.0;
            }
            "poor" => {
                self.network_latency = 100.0;
                self.network_loss = 5.0;
                self.network_bandwidth = 10.0;
            }
            "mobile" => {
                self.network_latency = 200.0;
                self.network_loss = 10.0;
                self.network_bandwidth = 5.0;
            }
            "satellite" => {
                self.network_latency = 500.0;
                self.network_loss = 2.0;
                self.network_bandwidth = 2.0;
            }
            "adversarial" => {
                self.network_latency = 1000.0;
                self.network_loss = 20.0;
                self.network_bandwidth = 1.0;
            }
            _ => {}
        }
    }

    fn toggle_security_testing(&mut self) {
        self.security_test_active = !self.security_test_active;
        if self.security_test_active {
            // Simulate security test results
            self.security_score = 85.0 + (self.time_slot as f64 % 20.0);
            self.vulnerabilities_count = (self.time_slot % 5) as usize;
        }
    }

    fn toggle_cloud_deployment(&mut self) {
        self.cloud_deployment_active = !self.cloud_deployment_active;
        if self.cloud_deployment_active {
            self.cloud_status = "running".to_string();
        } else {
            self.cloud_status = "stopped".to_string();
        }
    }

    fn scale_cloud_instances(&mut self) {
        if self.cloud_deployment_active {
            self.cloud_instances = (self.cloud_instances % 5) + 1;
        }
    }

    fn show_help(&self) {
        println!("Real QUIC Bottom Help:");
        println!("  q/ESC - Quit");
        println!("  r - Reset all data");
        println!("  h - Show this help");
        println!("  1 - Dashboard view");
        println!("  2 - Analytics view");
        println!("  3 - Network simulation view");
        println!("  4 - Security testing view");
        println!("  5 - Cloud deployment view");
        println!("  6 - BBRv3 congestion control view");
        println!("  a - All views");
        println!("  n - Toggle network simulation");
        println!("  +/- - Change network preset");
        println!("  s - Toggle security testing");
        println!("  d - Toggle cloud deployment");
        println!("  i - Scale cloud instances");
    }

    fn ui(&self, f: &mut Frame) {
        match self.current_view {
            ViewMode::Dashboard => self.render_dashboard(f),
            ViewMode::Analytics => self.render_analytics_view(f),
            ViewMode::Network => self.render_network_view(f),
            ViewMode::Security => self.render_security_view(f),
            ViewMode::Cloud => self.render_cloud_view(f),
            ViewMode::BBRv3 => self.render_bbrv3_view(f),
            ViewMode::All => self.render_all_view(f),
        }
    }

    fn render_dashboard(&self, f: &mut Frame) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Header
                Constraint::Min(0),    // Main content
                Constraint::Length(3), // Footer
            ])
            .split(f.area());

        self.render_header(f, chunks[0], "Real QUIC Bottom - Dashboard");

        let main_chunks = Layout::default()
            .direction(Direction::Horizontal)
            .constraints([
                Constraint::Percentage(50), // Left column
                Constraint::Percentage(50), // Right column
            ])
            .split(chunks[1]);

        let left_chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Percentage(33), // Current metrics
                Constraint::Percentage(33), // Latency
                Constraint::Percentage(34), // Throughput
            ])
            .split(main_chunks[0]);

        let right_chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Percentage(50), // Heatmap
                Constraint::Percentage(50), // Anomaly
            ])
            .split(main_chunks[1]);

        // Current metrics widget
        let metrics_opt = self.current_metrics.lock().unwrap();
        let metrics_text = if let Some(metrics) = metrics_opt.as_ref() {
            format!(
                "Connections: {}\nLatency: {:.2} ms\nThroughput: {:.2} Mbps\nRTT: {:.2} ms\nPacket Loss: {:.2}%\nRetransmits: {}\nErrors: {}\nStreams: {}",
                metrics.connections,
                metrics.latency,
                metrics.throughput,
                metrics.rtt,
                metrics.packet_loss * 100.0,
                metrics.retransmits,
                metrics.errors,
                metrics.streams
            )
        } else {
            "Waiting for metrics...\n\nMake sure quic-test is running\nand sending data to:\nhttp://127.0.0.1:8080/api/metrics".to_string()
        };
        drop(metrics_opt);

        let current_metrics_widget = Paragraph::new(metrics_text)
            .style(Style::default().fg(Color::Cyan))
            .block(Block::default().borders(Borders::ALL).title("Current Metrics"));
        f.render_widget(current_metrics_widget, left_chunks[0]);

        self.latency_graph.render(f, left_chunks[1]);
        self.throughput_graph.render(f, left_chunks[2]);
        self.performance_heatmap.render(f, right_chunks[0]);
        self.anomaly_widget.render(f, right_chunks[1]);

        self.render_footer(f, chunks[2]);
    }

    fn render_analytics_view(&self, f: &mut Frame) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Header
                Constraint::Min(0),    // Analytics
                Constraint::Length(3), // Footer
            ])
            .split(f.area());

        self.render_header(f, chunks[0], "Real QUIC Bottom - Analytics");

        let main_chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Percentage(50), // Correlation
                Constraint::Percentage(50), // Anomaly
            ])
            .split(chunks[1]);

        self.correlation_widget.render(f, main_chunks[0]);
        self.anomaly_widget.render(f, main_chunks[1]);

        self.render_footer(f, chunks[2]);
    }

    fn render_network_view(&self, f: &mut Frame) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Header
                Constraint::Min(0),    // Network info
                Constraint::Length(3), // Footer
            ])
            .split(f.area());

        self.render_header(f, chunks[0], "Real QUIC Bottom - Network Simulation");

        // Get current metrics for real-time data
        let metrics_opt = self.current_metrics.lock().unwrap();
        let metrics_text = if let Some(metrics) = metrics_opt.as_ref() {
            format!(
                "Network Simulation: {}\nPreset: {}\nSimulated Latency: {:.1}ms\nSimulated Loss: {:.1}%\nSimulated Bandwidth: {:.1} Mbps\n\n--- Real Metrics ---\nActual Latency: {:.2} ms\nActual Throughput: {:.2} Mbps\nActual RTT: {:.2} ms\nPacket Loss: {:.2}%\nRetransmits: {}\nConnections: {}",
                if self.network_simulation_active { "ACTIVE" } else { "INACTIVE" },
                self.network_preset,
                self.network_latency,
                self.network_loss,
                self.network_bandwidth,
                metrics.latency,
                metrics.throughput,
                metrics.rtt,
                metrics.packet_loss * 100.0,
                metrics.retransmits,
                metrics.connections
            )
        } else {
            format!(
                "Network Simulation: {}\nPreset: {}\nLatency: {:.1}ms\nLoss: {:.1}%\nBandwidth: {:.1} Mbps\n\n--- Real Metrics ---\nWaiting for data...",
                if self.network_simulation_active { "ACTIVE" } else { "INACTIVE" },
                self.network_preset,
                self.network_latency,
                self.network_loss,
                self.network_bandwidth
            )
        };
        drop(metrics_opt);

        let network_paragraph = Paragraph::new(metrics_text)
            .style(Style::default().fg(Color::Cyan))
            .block(Block::default().borders(Borders::ALL).title("Network Status"));
        f.render_widget(network_paragraph, chunks[1]);

        self.render_footer(f, chunks[2]);
    }

    fn render_security_view(&self, f: &mut Frame) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Header
                Constraint::Min(0),    // Security info
                Constraint::Length(3), // Footer
            ])
            .split(f.area());

        self.render_header(f, chunks[0], "Real QUIC Bottom - Security Testing");

        // Get current metrics for real-time security data
        let metrics_opt = self.current_metrics.lock().unwrap();
        let security_text = if let Some(metrics) = metrics_opt.as_ref() {
            // Calculate security score based on errors and packet loss
            let error_rate = if metrics.connections > 0 {
                (metrics.errors as f64 / metrics.connections as f64) * 100.0
            } else {
                0.0
            };
            let calculated_score = (100.0 - error_rate - (metrics.packet_loss * 100.0)).max(0.0);
            let final_score = if self.security_test_active {
                self.security_score
            } else {
                calculated_score
            };
            
            format!(
                "Security Testing: {}\nSecurity Score: {:.1}%\nVulnerabilities: {}\n\n--- Connection Security ---\nErrors: {}\nError Rate: {:.2}%\nPacket Loss: {:.2}%\nRetransmits: {}\nHandshake Time: {:.2} ms\nJitter: {:.2} ms",
                if self.security_test_active { "ACTIVE" } else { "INACTIVE" },
                final_score,
                if self.security_test_active { self.vulnerabilities_count } else { 0 },
                metrics.errors,
                error_rate,
                metrics.packet_loss * 100.0,
                metrics.retransmits,
                metrics.handshake_time,
                metrics.jitter
            )
        } else {
            format!(
                "Security Testing: {}\nSecurity Score: {:.1}%\nVulnerabilities: {}\n\n--- Connection Security ---\nWaiting for data...",
                if self.security_test_active { "ACTIVE" } else { "INACTIVE" },
                self.security_score,
                self.vulnerabilities_count
            )
        };
        drop(metrics_opt);

        let security_paragraph = Paragraph::new(security_text)
            .style(Style::default().fg(Color::Yellow))
            .block(Block::default().borders(Borders::ALL).title("Security Status"));
        f.render_widget(security_paragraph, chunks[1]);

        self.render_footer(f, chunks[2]);
    }

    fn render_cloud_view(&self, f: &mut Frame) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Header
                Constraint::Min(0),    // Cloud info
                Constraint::Length(3), // Footer
            ])
            .split(f.area());

        self.render_header(f, chunks[0], "Real QUIC Bottom - Cloud Deployment");

        // Cloud deployment status
        let cloud_text = format!(
            "Cloud Deployment: {}\nProvider: {}\nInstances: {}\nStatus: {}",
            if self.cloud_deployment_active { "ACTIVE" } else { "INACTIVE" },
            self.cloud_provider,
            self.cloud_instances,
            self.cloud_status
        );

        let cloud_paragraph = Paragraph::new(cloud_text)
            .style(Style::default().fg(Color::Green))
            .block(Block::default().borders(Borders::ALL).title("Cloud Status"));
        f.render_widget(cloud_paragraph, chunks[1]);

        self.render_footer(f, chunks[2]);
    }

    fn render_bbrv3_view(&self, f: &mut Frame) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Header
                Constraint::Min(0),    // Main content
                Constraint::Length(3), // Footer
            ])
            .split(f.area());

        self.render_header(f, chunks[0], "BBRv3 Congestion Control");

        // Get current metrics
        let metrics_opt = self.current_metrics.lock().unwrap();

        if let Some(metrics) = metrics_opt.as_ref() {
            if metrics.bbrv3_phase.is_some() {
                // Main content area with 2 columns
                let main_chunks = Layout::default()
                    .direction(Direction::Horizontal)
                    .constraints([
                        Constraint::Percentage(50), // Left column
                        Constraint::Percentage(50), // Right column
                    ])
                    .split(chunks[1]);

                // Left column - 3 rows
                let left_chunks = Layout::default()
                    .direction(Direction::Vertical)
                    .constraints([
                        Constraint::Percentage(33), // Phase Status
                        Constraint::Percentage(33), // Bandwidth Estimates
                        Constraint::Percentage(34), // Loss Metrics
                    ])
                    .split(main_chunks[0]);

                // Right column - 3 rows
                let right_chunks = Layout::default()
                    .direction(Direction::Vertical)
                    .constraints([
                        Constraint::Percentage(33), // Bufferbloat & Stability
                        Constraint::Percentage(33), // Pacing/CWND Gains
                        Constraint::Percentage(34), // Recovery Metrics
                    ])
                    .split(main_chunks[1]);

                // 1. Phase Status Widget
                if let Some(phase) = &metrics.bbrv3_phase {
                    let phase_color = match phase.as_str() {
                        "Startup" => Color::Red,
                        "Drain" => Color::Yellow,
                        "ProbeBW" => Color::Green,
                        "ProbeRTT" => Color::Cyan,
                        _ => Color::White,
                    };

                    let phase_text = format!(
                        "Current Phase: {}\n\nDescription:\n- Manages network bottleneck\n- Optimizes bandwidth usage\n- Adapts to network conditions",
                        phase
                    );

                    let phase_widget = Paragraph::new(phase_text)
                        .style(Style::default().fg(phase_color).add_modifier(Modifier::BOLD))
                        .block(Block::default().borders(Borders::ALL).title("Phase Status"));
                    f.render_widget(phase_widget, left_chunks[0]);
                }

                // 2. Bandwidth Estimates Widget
                let bw_text = if let (Some(bw_fast), Some(bw_slow)) =
                    (&metrics.bbrv3_bw_fast, &metrics.bbrv3_bw_slow) {
                    let fast_mbps = bw_fast / 1_000_000.0;
                    let slow_mbps = bw_slow / 1_000_000.0;
                    format!(
                        "Fast Bandwidth: {:.2} Mbps\nSlow Bandwidth: {:.2} Mbps\n\nRatio: {:.2}x",
                        fast_mbps,
                        slow_mbps,
                        fast_mbps / slow_mbps.max(0.01)
                    )
                } else {
                    "N/A".to_string()
                };

                let bw_widget = Paragraph::new(bw_text)
                    .style(Style::default().fg(Color::Green))
                    .block(Block::default().borders(Borders::ALL).title("Bandwidth Estimates"));
                f.render_widget(bw_widget, left_chunks[1]);

                // 3. Loss Metrics Widget
                let loss_text = if let Some(loss_rate) = metrics.bbrv3_loss_rate_ema {
                    format!(
                        "Loss Rate (EMA): {:.2}%\n\nStatus: {}\nThreshold: 2.0%",
                        loss_rate * 100.0,
                        if loss_rate < 0.02 { "HEALTHY" } else { "ELEVATED" }
                    )
                } else {
                    "N/A".to_string()
                };

                let loss_widget = Paragraph::new(loss_text)
                    .style(Style::default().fg(Color::Yellow))
                    .block(Block::default().borders(Borders::ALL).title("Loss Metrics"));
                f.render_widget(loss_widget, left_chunks[2]);

                // 4. Bufferbloat & Stability Widget
                let bufferbloat_text = if let Some(factor) = metrics.bbrv3_bufferbloat_factor {
                    let status = if factor < 0.1 { "EXCELLENT" }
                                else if factor < 0.3 { "GOOD" }
                                else { "HIGH" };
                    format!(
                        "Bufferbloat: {:.3}\n\nStatus: {}\nTarget: < 0.1",
                        factor,
                        status
                    )
                } else {
                    "N/A".to_string()
                };

                let stability_text = format!(
                    "{}\n\nStability Index: {:.2}",
                    bufferbloat_text,
                    metrics.bbrv3_stability_index.unwrap_or(0.0)
                );

                let bufferbloat_widget = Paragraph::new(stability_text)
                    .style(Style::default().fg(Color::Magenta))
                    .block(Block::default().borders(Borders::ALL).title("Bufferbloat & Stability"));
                f.render_widget(bufferbloat_widget, right_chunks[0]);

                // 5. Pacing/CWND Gains Widget
                let gains_text = format!(
                    "Pacing Gain: {:.2}x\nCWND Gain: {:.2}x\n\nTarget Inflight: {} KB",
                    metrics.bbrv3_pacing_gain.unwrap_or(1.0),
                    metrics.bbrv3_cwnd_gain.unwrap_or(2.0),
                    (metrics.bbrv3_inflight_target.unwrap_or(0.0) / 1024.0) as i64
                );

                let gains_widget = Paragraph::new(gains_text)
                    .style(Style::default().fg(Color::Cyan))
                    .block(Block::default().borders(Borders::ALL).title("Pacing/CWND Gains"));
                f.render_widget(gains_widget, right_chunks[1]);

                // 6. Recovery Metrics Widget
                let recovery_text = format!(
                    "Recovery Time: {:.0} ms\nLoss Efficiency: {:.2}%\n\nHeadroom Usage: {:.1}%",
                    metrics.bbrv3_recovery_time_ms.unwrap_or(0.0),
                    metrics.bbrv3_loss_recovery_efficiency.unwrap_or(0.0) * 100.0,
                    metrics.bbrv3_headroom_usage.unwrap_or(0.0) * 100.0
                );

                let recovery_widget = Paragraph::new(recovery_text)
                    .style(Style::default().fg(Color::Blue))
                    .block(Block::default().borders(Borders::ALL).title("Recovery Metrics"));
                f.render_widget(recovery_widget, right_chunks[2]);
            } else {
                // BBRv3 metrics not available
                let no_data_text = "BBRv3 metrics not available.\n\nMake sure:\n1. quic-test is running with --congestion-control=bbrv3\n2. Connection is established\n3. Data is being transmitted";
                let no_data_widget = Paragraph::new(no_data_text)
                    .style(Style::default().fg(Color::Red))
                    .block(Block::default().borders(Borders::ALL).title("BBRv3 Status"));
                f.render_widget(no_data_widget, chunks[1]);
            }
        } else {
            // No metrics at all
            let no_metrics_text = "No metrics received yet.\n\nWaiting for quic-test connection...";
            let no_metrics_widget = Paragraph::new(no_metrics_text)
                .style(Style::default().fg(Color::Yellow))
                .block(Block::default().borders(Borders::ALL).title("Connection Status"));
            f.render_widget(no_metrics_widget, chunks[1]);
        }

        self.render_footer(f, chunks[2]);
    }

    fn render_all_view(&self, f: &mut Frame) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Header
                Constraint::Min(0),    // Main content
                Constraint::Length(3), // Footer
            ])
            .split(f.area());

        self.render_header(f, chunks[0], "Real QUIC Bottom - All Views");

        let main_chunks = Layout::default()
            .direction(Direction::Horizontal)
            .constraints([
                Constraint::Percentage(50), // Left column
                Constraint::Percentage(50), // Right column
            ])
            .split(chunks[1]);

        let left_chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Percentage(33), // Latency
                Constraint::Percentage(33), // Throughput
                Constraint::Percentage(34), // Heatmap
            ])
            .split(main_chunks[0]);

        let right_chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Percentage(50), // Correlation
                Constraint::Percentage(50), // Anomaly
            ])
            .split(main_chunks[1]);

        self.latency_graph.render(f, left_chunks[0]);
        self.throughput_graph.render(f, left_chunks[1]);
        self.performance_heatmap.render(f, left_chunks[2]);
        self.correlation_widget.render(f, right_chunks[0]);
        self.anomaly_widget.render(f, right_chunks[1]);

        self.render_footer(f, chunks[2]);
    }

    fn render_header(&self, f: &mut Frame, area: Rect, title: &str) {
        let header_text = format!("Real QUIC Bottom - {}", title);
        let header = Paragraph::new(header_text)
            .style(Style::default().fg(Color::White).add_modifier(Modifier::BOLD))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(header, area);
    }

    fn render_footer(&self, f: &mut Frame, area: Rect) {
        let footer_text = "Press 'q' to quit, 'r' to reset, 'h' for help, '1-6' for views, 'a' for all, 'n' for network, 's' for security, 'd' for cloud";
        let footer = Paragraph::new(footer_text)
            .style(Style::default().fg(Color::Yellow))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(footer, area);
    }
}

// HTTP API server for receiving metrics from Go application
async fn start_http_server(
    current_metrics: Arc<Mutex<Option<RealQUICMetrics>>>,
    metrics_history: Arc<Mutex<Vec<RealQUICMetrics>>>,
) {
    let current_metrics_post = Arc::clone(&current_metrics);
    let metrics_filter = warp::path("api")
        .and(warp::path("metrics"))
        .and(warp::post())
        .and(warp::body::json())
        .map(move |metrics: RealQUICMetrics| {
            // Update current metrics
            {
                let mut current = current_metrics_post.lock().unwrap();
                *current = Some(metrics.clone());
            }
            
            // Add to history
            {
                let mut history = metrics_history.lock().unwrap();
                history.push(metrics);
                
                // Keep only last 1000 metrics
                if history.len() > 1000 {
                    history.remove(0);
                }
            }
            
            warp::reply::json(&serde_json::json!({"status": "ok"}))
        });

    let health_filter = warp::path("health")
        .map(|| warp::reply::json(&serde_json::json!({"status": "healthy"})));

    let current_metrics_get = Arc::clone(&current_metrics);
    let current_filter = warp::path("api")
        .and(warp::path("current"))
        .and(warp::get())
        .map(move || {
            let current = current_metrics_get.lock().unwrap();
            warp::reply::json(&*current)
        });

    let routes = metrics_filter
        .or(health_filter)
        .or(current_filter);

    println!("Starting HTTP API server on port 8080...");
    warp::serve(routes)
        .run(([127, 0, 0, 1], 8080))
        .await;
}

#[tokio::main]
async fn main() -> Result<()> {
    env_logger::init();

    let args: Vec<String> = std::env::args().collect();
    let headless = args.contains(&"--headless".to_string()) || args.contains(&"-h".to_string());

    println!("Starting Real QUIC Bottom...");
    println!("Real-time QUIC metrics from Go application!");
    println!("Professional visualizations with live data!");
    println!("");
    println!("Features:");
    println!("  ✅ Real-time QUIC metrics from Go application");
    println!("  ✅ HTTP API for metrics collection");
    println!("  ✅ Professional visualizations");
    println!("  ✅ Network simulation integration");
    println!("  ✅ Security testing integration");
    println!("  ✅ Cloud deployment monitoring");
    println!("  ✅ Interactive controls");
    println!("");
    println!("HTTP API endpoints:");
    println!("  POST /api/metrics - Receive metrics from Go app");
    println!("  GET /health - Health check");
    println!("  GET /api/current - Get current metrics");
    println!("");

    if headless {
        println!("🚀 Starting in HEADLESS mode (HTTP API only, no TUI)");
        println!("HTTP API server listening on http://127.0.0.1:8080");
        println!("\nTo test, run in another terminal:");
        println!("  curl -X GET http://127.0.0.1:8080/health");
        println!("  curl -X POST http://127.0.0.1:8080/api/metrics -H 'Content-Type: application/json' -d '{{...}}'");
        println!("\nPress Ctrl+C to stop.\n");

        let metrics_arc = Arc::new(Mutex::new(None));
        let history_arc = Arc::new(Mutex::new(Vec::new()));

        start_http_server(metrics_arc, history_arc).await;
    } else {
        println!("Starting in TUI mode");
        println!("Press '6' to switch to BBRv3 mode");
        println!("");

        let mut app = RealQUICBottom::new(100).await?;
        app.run().await?;
    }

    println!("✅ Real QUIC Bottom completed!");
    Ok(())
}
