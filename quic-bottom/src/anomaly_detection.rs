//! Anomaly detection for QUIC metrics
//! 
//! Automatically detects anomalies in performance data

use ratatui::{
    layout::{Constraint, Direction, Layout, Rect},
    style::{Color, Modifier, Style},
    text::{Line, Span},
    widgets::{Block, Borders, Paragraph},
    Frame,
};
use std::collections::VecDeque;

/// Anomaly detection result
#[derive(Debug, Clone)]
pub struct AnomalyResult {
    pub metric: String,
    pub value: f64,
    pub expected_range: (f64, f64),
    pub severity: AnomalySeverity,
    pub timestamp: chrono::DateTime<chrono::Utc>,
    pub description: String,
}

/// Anomaly severity levels
#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub enum AnomalySeverity {
    Low,
    Medium,
    High,
    Critical,
}

impl AnomalySeverity {
    pub fn get_color(&self) -> Color {
        match self {
            AnomalySeverity::Low => Color::Yellow,
            AnomalySeverity::Medium => Color::LightRed,
            AnomalySeverity::High => Color::Red,
            AnomalySeverity::Critical => Color::Magenta,
        }
    }

    pub fn get_description(&self) -> &'static str {
        match self {
            AnomalySeverity::Low => "Low",
            AnomalySeverity::Medium => "Medium",
            AnomalySeverity::High => "High",
            AnomalySeverity::Critical => "Critical",
        }
    }
}

/// Anomaly detector for QUIC metrics
pub struct AnomalyDetector {
    /// Historical data for each metric
    pub metric_history: std::collections::HashMap<String, VecDeque<f64>>,
    
    /// Anomaly results
    pub anomalies: VecDeque<AnomalyResult>,
    
    /// Maximum number of anomalies to keep
    pub max_anomalies: usize,
    
    /// Detection sensitivity (0.0 to 1.0)
    pub sensitivity: f64,
}

impl AnomalyDetector {
    pub fn new(sensitivity: f64) -> Self {
        Self {
            metric_history: std::collections::HashMap::new(),
            anomalies: VecDeque::new(),
            max_anomalies: 100,
            sensitivity,
        }
    }

    /// Add metric data point
    pub fn add_data_point(&mut self, metric: String, value: f64) {
        // Add to history
        self.metric_history
            .entry(metric.clone())
            .or_insert_with(VecDeque::new)
            .push_back(value);
        
        // Keep only recent data (last 100 points)
        if let Some(history) = self.metric_history.get_mut(&metric) {
            while history.len() > 100 {
                history.pop_front();
            }
        }

        // Check for anomalies
        if let Some(anomaly) = self.detect_anomaly(&metric, value) {
            self.anomalies.push_back(anomaly);
            
            // Keep only recent anomalies
            while self.anomalies.len() > self.max_anomalies {
                self.anomalies.pop_front();
            }
        }
    }

    /// Detect anomaly in metric value
    fn detect_anomaly(&self, metric: &str, value: f64) -> Option<AnomalyResult> {
        let history = self.metric_history.get(metric)?;
        
        if history.len() < 10 {
            return None; // Need more data for detection
        }

        let data: Vec<f64> = history.iter().cloned().collect();
        let (mean, std_dev) = self.calculate_statistics(&data);
        
        // Z-score based detection
        let z_score = (value - mean) / std_dev;
        let threshold = 2.0 + (1.0 - self.sensitivity) * 2.0; // 2.0 to 4.0 based on sensitivity
        
        if z_score.abs() > threshold {
            let severity = self.determine_severity(z_score.abs());
            let expected_range = (mean - 2.0 * std_dev, mean + 2.0 * std_dev);
            let description = format!(
                "Z-score: {:.2}, Expected: {:.1}-{:.1}, Actual: {:.1}",
                z_score, expected_range.0, expected_range.1, value
            );
            
            Some(AnomalyResult {
                metric: metric.to_string(),
                value,
                expected_range,
                severity,
                timestamp: chrono::Utc::now(),
                description,
            })
        } else {
            None
        }
    }

    /// Calculate mean and standard deviation
    fn calculate_statistics(&self, data: &[f64]) -> (f64, f64) {
        if data.is_empty() {
            return (0.0, 0.0);
        }

        let mean = data.iter().sum::<f64>() / data.len() as f64;
        let variance = data.iter()
            .map(|x| (x - mean).powi(2))
            .sum::<f64>() / data.len() as f64;
        let std_dev = variance.sqrt();

        (mean, std_dev)
    }

    /// Determine anomaly severity based on z-score
    fn determine_severity(&self, z_score: f64) -> AnomalySeverity {
        match z_score {
            x if x >= 4.0 => AnomalySeverity::Critical,
            x if x >= 3.0 => AnomalySeverity::High,
            x if x >= 2.5 => AnomalySeverity::Medium,
            _ => AnomalySeverity::Low,
        }
    }

    /// Get recent anomalies
    pub fn get_recent_anomalies(&self, count: usize) -> Vec<AnomalyResult> {
        self.anomalies
            .iter()
            .rev()
            .take(count)
            .cloned()
            .collect()
    }

    /// Get anomaly count by severity
    pub fn get_anomaly_counts(&self) -> std::collections::HashMap<AnomalySeverity, usize> {
        let mut counts = std::collections::HashMap::new();
        
        for anomaly in &self.anomalies {
            *counts.entry(anomaly.severity.clone()).or_insert(0) += 1;
        }
        
        counts
    }
}

/// Anomaly detection widget
pub struct AnomalyWidget {
    detector: AnomalyDetector,
    title: String,
}

impl AnomalyWidget {
    pub fn new(title: String, sensitivity: f64) -> Self {
        Self {
            detector: AnomalyDetector::new(sensitivity),
            title,
        }
    }

    /// Add metric data
    pub fn add_metric_data(&mut self, metric: String, value: f64) {
        self.detector.add_data_point(metric, value);
    }

    /// Render the anomaly widget
    pub fn render(&self, f: &mut Frame, area: Rect) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Title
                Constraint::Min(0),    // Anomaly list
                Constraint::Length(3), // Summary
            ])
            .split(area);

        // Title
        self.render_title(f, chunks[0]);
        
        // Anomaly list
        self.render_anomalies(f, chunks[1]);
        
        // Summary
        self.render_summary(f, chunks[2]);
    }

    fn render_title(&self, f: &mut Frame, area: Rect) {
        let title = Paragraph::new(self.title.clone())
            .style(Style::default().fg(Color::White).add_modifier(Modifier::BOLD))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(title, area);
    }

    fn render_anomalies(&self, f: &mut Frame, area: Rect) {
        let recent_anomalies = self.detector.get_recent_anomalies(10);
        
        if recent_anomalies.is_empty() {
            let no_anomalies = Paragraph::new("No anomalies detected")
                .style(Style::default().fg(Color::Green))
                .block(Block::default().borders(Borders::ALL));
            f.render_widget(no_anomalies, area);
            return;
        }

        let mut lines = Vec::new();
        for anomaly in recent_anomalies {
            let severity_color = anomaly.severity.get_color();
            let severity_text = anomaly.severity.get_description();
            
            let line = Line::from(vec![
                Span::styled(
                    format!("[{}] ", severity_text),
                    Style::default().fg(severity_color).add_modifier(Modifier::BOLD)
                ),
                Span::styled(
                    format!("{}: {:.2} ", anomaly.metric, anomaly.value),
                    Style::default().fg(Color::White)
                ),
                Span::styled(
                    format!("({})", anomaly.description),
                    Style::default().fg(Color::Gray)
                ),
            ]);
            lines.push(line);
        }

        let anomalies_paragraph = Paragraph::new(lines)
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(anomalies_paragraph, area);
    }

    fn render_summary(&self, f: &mut Frame, area: Rect) {
        let counts = self.detector.get_anomaly_counts();
        let total_anomalies = counts.values().sum::<usize>();
        
        let summary_text = if total_anomalies == 0 {
            "✅ No anomalies detected".to_string()
        } else {
            format!(
                "⚠️  {} anomalies: Critical: {} | High: {} | Medium: {} | Low: {}",
                total_anomalies,
                counts.get(&AnomalySeverity::Critical).unwrap_or(&0),
                counts.get(&AnomalySeverity::High).unwrap_or(&0),
                counts.get(&AnomalySeverity::Medium).unwrap_or(&0),
                counts.get(&AnomalySeverity::Low).unwrap_or(&0),
            )
        };
        
        let summary = Paragraph::new(summary_text)
            .style(Style::default().fg(Color::Cyan))
            .block(Block::default().borders(Borders::NONE));
        
        f.render_widget(summary, area);
    }
}

/// QUIC Anomaly Detection Widget
pub struct QUICAnomalyWidget {
    anomaly: AnomalyWidget,
}

impl QUICAnomalyWidget {
    pub fn new() -> Self {
        Self {
            anomaly: AnomalyWidget::new("QUIC Anomaly Detection".to_string(), 0.7),
        }
    }

    /// Add QUIC metric data
    pub fn add_quic_metric(&mut self, metric: String, value: f64) {
        self.anomaly.add_metric_data(metric, value);
    }

    /// Render the anomaly widget
    pub fn render(&self, f: &mut Frame, area: Rect) {
        self.anomaly.render(f, area);
    }
}
