//! QUIC-specific widgets for bottom TUI
//! 
//! Adapted from bottom's widget system for QUIC protocol monitoring

use ratatui::{
    backend::Backend,
    layout::{Constraint, Direction, Layout, Rect},
    style::{Color, Modifier, Style},
    text::{Line, Span},
    widgets::{Block, Borders, Gauge, Paragraph, Sparkline},
    Frame,
};
use std::collections::VecDeque;

use crate::metrics::{QUICMetrics, calculate_latency_percentiles, calculate_jitter};

/// QUIC Latency Widget - displays RTT, jitter, and percentiles
pub struct QUICLatencyWidget {
    data: VecDeque<f64>,
    max_points: usize,
}

impl QUICLatencyWidget {
    pub fn new(max_points: usize) -> Self {
        Self {
            data: VecDeque::with_capacity(max_points),
            max_points,
        }
    }

    pub fn update(&mut self, latency: f64) {
        self.data.push_back(latency);
        if self.data.len() > self.max_points {
            self.data.pop_front();
        }
    }

    pub fn render(&self, f: &mut Frame, area: Rect) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Title
                Constraint::Min(10),  // Graph
                Constraint::Length(4), // Stats
            ])
            .split(area);

        // Title
        let title = Paragraph::new("QUIC Latency (ms)")
            .style(Style::default().fg(Color::Yellow).add_modifier(Modifier::BOLD))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(title, chunks[0]);

        // Sparkline graph
        if !self.data.is_empty() {
            let sparkline = Sparkline::default()
                .data(&self.data.iter().map(|&x| x as u64).collect::<Vec<u64>>())
                .style(Style::default().fg(Color::Green))
                .block(Block::default().borders(Borders::ALL).title("Latency Graph"));
            f.render_widget(sparkline, chunks[2]);
        }

        // Stats
        if !self.data.is_empty() {
            let data_vec: Vec<f64> = self.data.iter().cloned().collect();
            let (p50, p95, p99) = calculate_latency_percentiles(&data_vec);
            let jitter = calculate_jitter(&data_vec);
            let current = self.data.back().unwrap_or(&0.0);
            
            let stats_text = format!(
                "Current: {:.2}ms | P50: {:.2}ms | P95: {:.2}ms | P99: {:.2}ms | Jitter: {:.2}ms",
                current, p50, p95, p99, jitter
            );
            
            let stats = Paragraph::new(stats_text)
                .style(Style::default().fg(Color::Cyan))
                .block(Block::default().borders(Borders::NONE));
            f.render_widget(stats, chunks[2]);
        }
    }
}

/// QUIC Throughput Widget - displays bandwidth and packet rates
pub struct QUICThroughputWidget {
    data: VecDeque<f64>,
    max_points: usize,
}

impl QUICThroughputWidget {
    pub fn new(max_points: usize) -> Self {
        Self {
            data: VecDeque::with_capacity(max_points),
            max_points,
        }
    }

    pub fn update(&mut self, throughput: f64) {
        self.data.push_back(throughput);
        if self.data.len() > self.max_points {
            self.data.pop_front();
        }
    }

    pub fn render(&self, f: &mut Frame, area: Rect) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Title
                Constraint::Min(10),  // Graph
                Constraint::Length(4), // Stats
            ])
            .split(area);

        // Title
        let title = Paragraph::new("QUIC Throughput (KB/s)")
            .style(Style::default().fg(Color::Blue).add_modifier(Modifier::BOLD))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(title, chunks[0]);

        // Sparkline graph
        if !self.data.is_empty() {
            let sparkline = Sparkline::default()
                .data(&self.data.iter().map(|&x| x as u64).collect::<Vec<u64>>())
                .style(Style::default().fg(Color::Magenta))
                .block(Block::default().borders(Borders::NONE));
            f.render_widget(sparkline, chunks[1]);
        }

        // Stats
        if !self.data.is_empty() {
            let current = self.data.back().unwrap_or(&0.0);
            let avg = self.data.iter().sum::<f64>() / self.data.len() as f64;
            let max = self.data.iter().fold(0.0f64, |a, &b| a.max(b));
            
            let stats_text = format!(
                "Current: {:.2} KB/s | Avg: {:.2} KB/s | Max: {:.2} KB/s",
                current, avg, max
            );
            
            let stats = Paragraph::new(stats_text)
                .style(Style::default().fg(Color::Cyan))
                .block(Block::default().borders(Borders::NONE));
            f.render_widget(stats, chunks[2]);
        }
    }
}

/// QUIC Connection Status Widget - displays connection statistics
pub struct QUICConnectionWidget {
    active_connections: i32,
    failed_connections: i32,
    total_connections: i32,
    handshake_times: VecDeque<f64>,
}

impl QUICConnectionWidget {
    pub fn new() -> Self {
        Self {
            active_connections: 0,
            failed_connections: 0,
            total_connections: 0,
            handshake_times: VecDeque::with_capacity(100),
        }
    }

    pub fn update(&mut self, active: i32, failed: i32, total: i32) {
        self.active_connections = active;
        self.failed_connections = failed;
        self.total_connections = total;
    }

    pub fn add_handshake_time(&mut self, time: f64) {
        self.handshake_times.push_back(time);
        if self.handshake_times.len() > 100 {
            self.handshake_times.pop_front();
        }
    }

    pub fn render(&self, f: &mut Frame, area: Rect) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Title
                Constraint::Length(3), // Active connections
                Constraint::Length(3), // Failed connections
                Constraint::Length(3), // Success rate
                Constraint::Min(0),    // Handshake times
            ])
            .split(area);

        // Title
        let title = Paragraph::new("QUIC Connections")
            .style(Style::default().fg(Color::Green).add_modifier(Modifier::BOLD))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(title, chunks[0]);

        // Active connections
        let active_text = format!("Active: {}", self.active_connections);
        let active_style = if self.active_connections > 0 {
            Style::default().fg(Color::Green)
        } else {
            Style::default().fg(Color::Red)
        };
        let active = Paragraph::new(active_text)
            .style(active_style)
            .block(Block::default().borders(Borders::NONE));
        f.render_widget(active, chunks[1]);

        // Failed connections
        let failed_text = format!("Failed: {}", self.failed_connections);
        let failed_style = if self.failed_connections > 0 {
            Style::default().fg(Color::Red)
        } else {
            Style::default().fg(Color::Green)
        };
        let failed = Paragraph::new(failed_text)
            .style(failed_style)
            .block(Block::default().borders(Borders::NONE));
        f.render_widget(failed, chunks[2]);

        // Success rate
        let success_rate = if self.total_connections > 0 {
            (self.active_connections as f64 / self.total_connections as f64) * 100.0
        } else {
            0.0
        };
        let success_text = format!("Success Rate: {:.1}%", success_rate);
        let success_style = if success_rate >= 95.0 {
            Style::default().fg(Color::Green)
        } else if success_rate >= 80.0 {
            Style::default().fg(Color::Yellow)
        } else {
            Style::default().fg(Color::Red)
        };
        let success = Paragraph::new(success_text)
            .style(success_style)
            .block(Block::default().borders(Borders::NONE));
        f.render_widget(success, chunks[3]);

        // Handshake times sparkline
        if !self.handshake_times.is_empty() && chunks.len() > 4 {
            let sparkline = Sparkline::default()
                .data(&self.handshake_times.iter().map(|&x| x as u64).collect::<Vec<u64>>())
                .style(Style::default().fg(Color::Yellow))
                .block(Block::default().borders(Borders::NONE));
            f.render_widget(sparkline, chunks[4]);
        }
    }
}

/// QUIC Network Quality Widget - displays packet loss, retransmits, and congestion control
pub struct QUICNetworkWidget {
    packet_loss: f64,
    retransmits: i32,
    congestion_control: String,
    loss_data: VecDeque<f64>,
    retransmit_data: VecDeque<i32>,
}

impl QUICNetworkWidget {
    pub fn new() -> Self {
        Self {
            packet_loss: 0.0,
            retransmits: 0,
            congestion_control: "Unknown".to_string(),
            loss_data: VecDeque::with_capacity(100),
            retransmit_data: VecDeque::with_capacity(100),
        }
    }

    pub fn update(&mut self, packet_loss: f64, retransmits: i32, cc: String) {
        self.packet_loss = packet_loss;
        self.retransmits = retransmits;
        self.congestion_control = cc;
        
        // Update time series data
        self.loss_data.push_back(packet_loss);
        self.retransmit_data.push_back(retransmits);
        
        if self.loss_data.len() > 100 {
            self.loss_data.pop_front();
        }
        if self.retransmit_data.len() > 100 {
            self.retransmit_data.pop_front();
        }
    }

    pub fn render(&self, f: &mut Frame, area: Rect) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Title
                Constraint::Length(3), // Packet loss
                Constraint::Length(3), // Retransmits
                Constraint::Length(3), // Congestion control
                Constraint::Min(0),    // Graphs
            ])
            .split(area);

        // Title
        let title = Paragraph::new("QUIC Network Quality")
            .style(Style::default().fg(Color::Red).add_modifier(Modifier::BOLD))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(title, chunks[0]);

        // Packet loss
        let loss_text = format!("Packet Loss: {:.2}%", self.packet_loss);
        let loss_style = if self.packet_loss < 1.0 {
            Style::default().fg(Color::Green)
        } else if self.packet_loss < 5.0 {
            Style::default().fg(Color::Yellow)
        } else {
            Style::default().fg(Color::Red)
        };
        let loss = Paragraph::new(loss_text)
            .style(loss_style)
            .block(Block::default().borders(Borders::NONE));
        f.render_widget(loss, chunks[1]);

        // Retransmits
        let retrans_text = format!("Retransmits: {}", self.retransmits);
        let retrans_style = if self.retransmits < 10 {
            Style::default().fg(Color::Green)
        } else if self.retransmits < 50 {
            Style::default().fg(Color::Yellow)
        } else {
            Style::default().fg(Color::Red)
        };
        let retrans = Paragraph::new(retrans_text)
            .style(retrans_style)
            .block(Block::default().borders(Borders::NONE));
        f.render_widget(retrans, chunks[2]);

        // Congestion control
        let cc_text = format!("CC Algorithm: {}", self.congestion_control);
        let cc = Paragraph::new(cc_text)
            .style(Style::default().fg(Color::Cyan))
            .block(Block::default().borders(Borders::NONE));
        f.render_widget(cc, chunks[3]);

        // Graphs
        if chunks.len() > 4 && !self.loss_data.is_empty() {
            let graph_chunks = Layout::default()
                .direction(Direction::Horizontal)
                .constraints([
                    Constraint::Percentage(50), // Loss graph
                    Constraint::Percentage(50), // Retransmit graph
                ])
                .split(chunks[4]);

            // Loss graph
            let loss_sparkline = Sparkline::default()
                .data(&self.loss_data.iter().map(|&x| x as u64).collect::<Vec<u64>>())
                .style(Style::default().fg(Color::Red))
                .block(Block::default().borders(Borders::NONE));
            f.render_widget(loss_sparkline, graph_chunks[0]);

            // Retransmit graph
            let retrans_sparkline = Sparkline::default()
                .data(&self.retransmit_data.iter().map(|&x| x as u64).collect::<Vec<u64>>())
                .style(Style::default().fg(Color::Yellow))
                .block(Block::default().borders(Borders::NONE));
            f.render_widget(retrans_sparkline, graph_chunks[1]);
        }
    }
}
