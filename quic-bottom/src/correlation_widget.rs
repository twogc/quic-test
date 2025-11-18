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
                "RTT".to_string(),
                "Jitter".to_string(),
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
    /// Returns Pearson correlation coefficient
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

        // Check for zero variance (constant values)
        if sum_sq1 == 0.0 || sum_sq2 == 0.0 {
            return 0.0;
        }

        let denominator = (sum_sq1 * sum_sq2).sqrt();
        if denominator == 0.0 {
            return 0.0;
        }

        let correlation = numerator / denominator;
        
        // Clamp to [-1, 1] range
        correlation.max(-1.0).min(1.0)
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
        // Show matrix even if correlations are empty (they're being recalculated)
        // Only show empty message if we truly have no data
        if self.correlations.is_empty() {
            // Show a brief message that correlations are being calculated
            let empty_text = "Recalculating correlations...\nPlease wait a moment";
            let empty_paragraph = Paragraph::new(empty_text)
                .style(Style::default().fg(Color::Yellow))
                .block(Block::default().borders(Borders::ALL).title("QUIC Metrics Correlation"));
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
        let entry = self.metric_data.entry(metric.clone()).or_insert_with(Vec::new);
        entry.push(value);
        
        // Keep only recent data (last 100 points)
        if entry.len() > 100 {
            entry.remove(0);
        }
    }
    
    /// Get number of data points for a metric
    pub fn get_data_points_count(&self, metric: &str) -> usize {
        self.metric_data.get(metric).map(|v| v.len()).unwrap_or(0)
    }
    
    /// Get total number of metrics with data
    pub fn get_metrics_count(&self) -> usize {
        self.metric_data.len()
    }

    /// Update correlations
    pub fn update_correlations(&mut self) {
        // Get all metrics that have data (need at least 3 points for meaningful correlation)
        let min_data_points = 3;
        let metrics: Vec<String> = self.metric_data.keys()
            .filter(|k| {
                let count = self.metric_data.get(*k).map(|v| v.len()).unwrap_or(0);
                count >= min_data_points
            })
            .cloned()
            .collect();
        
        // Only calculate if we have at least 2 metrics with enough data
        if metrics.len() < 2 {
            return;
        }
        
        // Calculate new correlations first, then replace old ones
        // This prevents flickering when correlations are temporarily empty
        let mut new_correlations = Vec::new();
        
        for i in 0..metrics.len() {
            for j in (i + 1)..metrics.len() {
                if let (Some(data1), Some(data2)) = (
                    self.metric_data.get(&metrics[i]),
                    self.metric_data.get(&metrics[j])
                ) {
                    // Only calculate correlation if we have enough data points
                    if data1.len() >= min_data_points && data2.len() >= min_data_points {
                        // Use the minimum length to ensure both series are aligned
                        // Use more recent data (last N points) for better correlation
                        let min_len = data1.len().min(data2.len()).min(50); // Use up to 50 points
                        let data1_slice = &data1[data1.len() - min_len..];
                        let data2_slice = &data2[data2.len() - min_len..];
                        
                        // Check if data has variance (not all values are the same)
                        let has_variance1 = data1_slice.iter().any(|&x| (x - data1_slice[0]).abs() > 0.001);
                        let has_variance2 = data2_slice.iter().any(|&x| (x - data2_slice[0]).abs() > 0.001);
                        
                        if has_variance1 && has_variance2 {
                            let correlation = self.correlation.calculate_correlation(data1_slice, data2_slice);
                            let significance = correlation.abs(); // Simplified significance
                            
                            // Only add if correlation is meaningful (not NaN or infinite)
                            if correlation.is_finite() {
                                new_correlations.push(CorrelationData {
                                    metric1: metrics[i].clone(),
                                    metric2: metrics[j].clone(),
                                    correlation,
                                    significance,
                                });
                            }
                        }
                    }
                }
            }
        }
        
        // Only replace old correlations if we found new ones
        // This prevents clearing correlations when recalculation fails
        if !new_correlations.is_empty() {
            self.correlation.correlations = new_correlations;
        }
    }

    /// Render the correlation widget
    pub fn render(&self, f: &mut Frame, area: Rect) {
        // Check if we have enough data before rendering
        let metrics_with_data: Vec<String> = self.metric_data.keys()
            .filter(|k| self.metric_data.get(*k).map(|v| v.len()).unwrap_or(0) >= 3)
            .cloned()
            .collect();
        
        // Only show status if we don't have enough data points yet
        // If we have correlations, show them even if they're temporarily empty during recalculation
        if metrics_with_data.len() < 2 {
            let data_counts: Vec<String> = self.metric_data.iter()
                .map(|(k, v)| format!("{}: {} pts", k, v.len()))
                .collect();
            
            let status_text = format!(
                "Collecting data...\nMetrics with 3+ points: {}/{}\n\nData points:\n{}",
                metrics_with_data.len(),
                self.metric_data.len(),
                data_counts.join("\n")
            );
            
            let status_paragraph = Paragraph::new(status_text)
                .style(Style::default().fg(Color::Yellow))
                .block(Block::default().borders(Borders::ALL).title("QUIC Metrics Correlation"));
            f.render_widget(status_paragraph, area);
            return;
        }
        
        // If we have enough data, always render the correlation matrix
        // Even if correlations are temporarily empty, they will be recalculated
        self.correlation.render(f, area);
    }
}
