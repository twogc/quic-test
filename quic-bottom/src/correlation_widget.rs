//! Correlation widget for analyzing relationships between metrics
//! 
//! Shows correlation matrix between different QUIC metrics

use ratatui::{
    layout::{Constraint, Direction, Layout, Rect},
    style::{Color, Modifier, Style},
    text::{Line, Span},
    widgets::{Block, Borders, Paragraph, Table, Row, Cell},
    Frame,
};
use std::collections::HashMap;

/// Correlation data between two metrics
#[derive(Debug, Clone)]
pub struct CorrelationData {
    pub metric1: String,
    pub metric2: String,
    pub correlation: f64,
    pub significance: f64,
}

/// Correlation widget for metric analysis
pub struct CorrelationWidget {
    /// Correlation data
    pub correlations: Vec<CorrelationData>,
    
    /// Available metrics
    pub metrics: Vec<String>,
    
    /// Title
    pub title: String,
}

impl CorrelationWidget {
    pub fn new(title: String) -> Self {
        Self {
            correlations: Vec::new(),
            metrics: vec![
                "Latency".to_string(),
                "Throughput".to_string(),
                "Packet Loss".to_string(),
                "Retransmits".to_string(),
                "Connections".to_string(),
                "Errors".to_string(),
            ],
            title,
        }
    }

    /// Add correlation data
    pub fn add_correlation(&mut self, metric1: String, metric2: String, correlation: f64, significance: f64) {
        let data = CorrelationData {
            metric1,
            metric2,
            correlation,
            significance,
        };
        self.correlations.push(data);
    }

    /// Calculate correlation between two data series
    pub fn calculate_correlation(&self, data1: &[f64], data2: &[f64]) -> f64 {
        if data1.len() != data2.len() || data1.is_empty() {
            return 0.0;
        }

        let n = data1.len() as f64;
        let mean1 = data1.iter().sum::<f64>() / n;
        let mean2 = data2.iter().sum::<f64>() / n;

        let mut numerator = 0.0;
        let mut sum_sq1 = 0.0;
        let mut sum_sq2 = 0.0;

        for (x, y) in data1.iter().zip(data2.iter()) {
            let dx = x - mean1;
            let dy = y - mean2;
            numerator += dx * dy;
            sum_sq1 += dx * dx;
            sum_sq2 += dy * dy;
        }

        if sum_sq1 == 0.0 || sum_sq2 == 0.0 {
            return 0.0;
        }

        numerator / (sum_sq1 * sum_sq2).sqrt()
    }

    /// Get color for correlation strength
    fn get_correlation_color(&self, correlation: f64) -> Color {
        let abs_corr = correlation.abs();
        match abs_corr {
            x if x >= 0.8 => Color::Red,
            x if x >= 0.6 => Color::LightRed,
            x if x >= 0.4 => Color::Yellow,
            x if x >= 0.2 => Color::LightGreen,
            _ => Color::Green,
        }
    }

    /// Get correlation strength description
    fn get_correlation_strength(&self, correlation: f64) -> &'static str {
        let abs_corr = correlation.abs();
        match abs_corr {
            x if x >= 0.8 => "Very Strong",
            x if x >= 0.6 => "Strong",
            x if x >= 0.4 => "Moderate",
            x if x >= 0.2 => "Weak",
            _ => "Very Weak",
        }
    }

    /// Render the correlation widget
    pub fn render(&self, f: &mut Frame, area: Rect) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Title
                Constraint::Min(0),    // Correlation matrix
                Constraint::Length(4), // Legend
            ])
            .split(area);

        // Title
        self.render_title(f, chunks[0]);
        
        // Correlation matrix
        self.render_correlation_matrix(f, chunks[1]);
        
        // Legend
        self.render_legend(f, chunks[2]);
    }

    fn render_title(&self, f: &mut Frame, area: Rect) {
        let title = Paragraph::new(self.title.clone())
            .style(Style::default().fg(Color::White).add_modifier(Modifier::BOLD))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(title, area);
    }

    fn render_correlation_matrix(&self, f: &mut Frame, area: Rect) {
        if self.correlations.is_empty() {
            let empty_text = "No correlation data available yet...";
            let empty_paragraph = Paragraph::new(empty_text)
                .style(Style::default().fg(Color::Gray))
                .block(Block::default().borders(Borders::ALL));
            f.render_widget(empty_paragraph, area);
            return;
        }

        // Create correlation matrix table
        let mut rows = Vec::new();
        
        // Header row
        let mut header_cells = vec![Cell::from("Metric").style(Style::default().fg(Color::Yellow))];
        for metric in &self.metrics {
            header_cells.push(Cell::from(metric.as_str()).style(Style::default().fg(Color::Yellow)));
        }
        rows.push(Row::new(header_cells));

        // Data rows
        for (i, metric1) in self.metrics.iter().enumerate() {
            let mut cells = vec![Cell::from(metric1.as_str()).style(Style::default().fg(Color::Yellow))];
            
            for (j, metric2) in self.metrics.iter().enumerate() {
                if i == j {
                    cells.push(Cell::from("1.00").style(Style::default().fg(Color::Green)));
                } else {
                    // Find correlation between these metrics
                    let correlation = self.correlations
                        .iter()
                        .find(|c| (c.metric1 == *metric1 && c.metric2 == *metric2) || 
                                 (c.metric1 == *metric2 && c.metric2 == *metric1))
                        .map(|c| c.correlation)
                        .unwrap_or(0.0);
                    
                    let color = self.get_correlation_color(correlation);
                    let formatted = format!("{:.2}", correlation);
                    cells.push(Cell::from(formatted).style(Style::default().fg(color)));
                }
            }
            
            rows.push(Row::new(cells));
        }

        let widths = vec![Constraint::Length(12); self.metrics.len() + 1];
        let table = Table::new(rows, widths)
            .block(Block::default().borders(Borders::ALL));

        f.render_widget(table, area);
    }

    fn render_legend(&self, f: &mut Frame, area: Rect) {
        let legend_text = "Correlation Strength: Red (Strong) | Yellow (Moderate) | Green (Weak)";
        let legend = Paragraph::new(legend_text)
            .style(Style::default().fg(Color::Cyan))
            .block(Block::default().borders(Borders::NONE));
        
        f.render_widget(legend, area);
    }
}

/// QUIC Metrics Correlation Widget
pub struct QUICCorrelationWidget {
    correlation: CorrelationWidget,
    metric_data: HashMap<String, Vec<f64>>,
}

impl QUICCorrelationWidget {
    pub fn new() -> Self {
        Self {
            correlation: CorrelationWidget::new("QUIC Metrics Correlation".to_string()),
            metric_data: HashMap::new(),
        }
    }

    /// Add metric data
    pub fn add_metric_data(&mut self, metric: String, value: f64) {
        self.metric_data.entry(metric.clone()).or_insert_with(Vec::new).push(value);
        
        // Keep only recent data (last 100 points)
        if let Some(data) = self.metric_data.get_mut(&metric) {
            if data.len() > 100 {
                data.remove(0);
            }
        }
    }

    /// Update correlations
    pub fn update_correlations(&mut self) {
        let metrics: Vec<String> = self.metric_data.keys().cloned().collect();
        
        for i in 0..metrics.len() {
            for j in (i + 1)..metrics.len() {
                if let (Some(data1), Some(data2)) = (
                    self.metric_data.get(&metrics[i]),
                    self.metric_data.get(&metrics[j])
                ) {
                    let correlation = self.correlation.calculate_correlation(data1, data2);
                    let significance = correlation.abs(); // Simplified significance
                    
                    self.correlation.add_correlation(
                        metrics[i].clone(),
                        metrics[j].clone(),
                        correlation,
                        significance,
                    );
                }
            }
        }
    }

    /// Render the correlation widget
    pub fn render(&self, f: &mut Frame, area: Rect) {
        self.correlation.render(f, area);
    }
}
