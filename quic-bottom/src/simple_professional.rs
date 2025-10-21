//! Simplified professional graphs for QUIC metrics
//! 
//! Based on bottom's advanced capabilities but simplified for easier implementation

use ratatui::{
    layout::{Constraint, Direction, Layout, Rect},
    style::{Color, Modifier, Style},
    symbols::Marker,
    text::Span,
    widgets::{Block, Borders, Dataset, GraphType, Paragraph},
    Frame,
};
use std::collections::VecDeque;

/// Simplified professional time graph for QUIC metrics
pub struct SimpleProfessionalGraph {
    /// Historical data points
    pub data_points: VecDeque<f64>,
    
    /// Maximum number of data points to keep
    pub max_points: usize,
    
    /// Y-axis bounds
    pub y_bounds: (f64, f64),
    
    /// Graph style
    pub style: Style,
    
    /// Title
    pub title: String,
    
    /// Whether graph is selected
    pub is_selected: bool,
}

impl SimpleProfessionalGraph {
    pub fn new(title: String, max_points: usize) -> Self {
        Self {
            data_points: VecDeque::with_capacity(max_points),
            max_points,
            y_bounds: (0.0, 100.0),
            style: Style::default().fg(Color::Green),
            title,
            is_selected: false,
        }
    }

    /// Add new data point
    pub fn add_data_point(&mut self, value: f64) {
        self.data_points.push_back(value);
        
        // Keep only recent data
        while self.data_points.len() > self.max_points {
            self.data_points.pop_front();
        }
        
        // Update y bounds based on current data
        self.update_y_bounds();
    }

    /// Update Y-axis bounds based on current data
    fn update_y_bounds(&mut self) {
        if self.data_points.is_empty() {
            return;
        }
        
        let min_val = self.data_points.iter().fold(f64::INFINITY, |a, &b| a.min(b));
        let max_val = self.data_points.iter().fold(f64::NEG_INFINITY, |a, &b| a.max(b));
        
        // Add some padding
        let padding = (max_val - min_val) * 0.1;
        self.y_bounds = (min_val - padding, max_val + padding);
    }

    /// Get analytics for the current data
    pub fn get_analytics(&self) -> SimpleAnalytics {
        if self.data_points.is_empty() {
            return SimpleAnalytics::default();
        }
        
        let values: Vec<f64> = self.data_points.iter().cloned().collect();
        let current = *values.last().unwrap();
        let average = values.iter().sum::<f64>() / values.len() as f64;
        let min = values.iter().fold(f64::INFINITY, |a, &b| a.min(b));
        let max = values.iter().fold(f64::NEG_INFINITY, |a, &b| a.max(b));
        
        // Calculate percentiles
        let mut sorted_values = values.clone();
        sorted_values.sort_by(|a, b| a.partial_cmp(b).unwrap());
        let p50 = percentile(&sorted_values, 0.5);
        let p95 = percentile(&sorted_values, 0.95);
        let p99 = percentile(&sorted_values, 0.99);
        
        SimpleAnalytics {
            current,
            average,
            min,
            max,
            p50,
            p95,
            p99,
            data_points: values.len(),
        }
    }

    /// Render the professional graph
    pub fn render(&self, f: &mut Frame, area: Rect) {
        if self.data_points.is_empty() {
            self.render_empty_state(f, area);
            return;
        }

        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Title
                Constraint::Min(8),    // Graph
                Constraint::Length(4), // Analytics
            ])
            .split(area);

        // Title
        self.render_title(f, chunks[0]);
        
        // Graph
        self.render_graph(f, chunks[1]);
        
        // Analytics
        self.render_analytics(f, chunks[2]);
    }

    fn render_title(&self, f: &mut Frame, area: Rect) {
        let title_style = if self.is_selected {
            Style::default().fg(Color::Yellow).add_modifier(Modifier::BOLD)
        } else {
            Style::default().fg(Color::White)
        };
        
        let title = Paragraph::new(self.title.clone())
            .style(title_style)
            .block(Block::default()
                .borders(Borders::ALL)
                .border_style(if self.is_selected { 
                    Style::default().fg(Color::Yellow) 
                } else { 
                    Style::default().fg(Color::Gray) 
                }));
        f.render_widget(title, area);
    }

    fn render_graph(&self, f: &mut Frame, area: Rect) {
        // Convert data to chart format
        let data: Vec<(f64, f64)> = self.data_points
            .iter()
            .enumerate()
            .map(|(i, &value)| (i as f64, value))
            .collect();

        if data.is_empty() {
            return;
        }

        let dataset = Dataset::default()
            .data(&data)
            .style(self.style)
            .graph_type(GraphType::Line)
            .marker(Marker::Braille);

        // Create chart with professional styling
        let chart = ratatui::widgets::Chart::new(vec![dataset])
            .block(Block::default()
                .borders(Borders::ALL)
                .title("Time Series")
                .title_style(Style::default().fg(Color::Cyan)))
            .x_axis(ratatui::widgets::Axis::default()
                .bounds([0.0, self.data_points.len() as f64])
                .labels(vec![
                    Span::styled("0", self.style),
                    Span::styled(format!("{}", self.data_points.len()), self.style),
                ]))
            .y_axis(ratatui::widgets::Axis::default()
                .bounds([self.y_bounds.0, self.y_bounds.1])
                .labels(vec![
                    Span::styled(format!("{:.1}", self.y_bounds.0), self.style),
                    Span::styled(format!("{:.1}", self.y_bounds.1), self.style),
                ]));

        f.render_widget(chart, area);
    }

    fn render_analytics(&self, f: &mut Frame, area: Rect) {
        let analytics = self.get_analytics();
        
        let analytics_text = format!(
            "Current: {:.2} | Avg: {:.2} | Min: {:.2} | Max: {:.2} | P50: {:.2} | P95: {:.2} | P99: {:.2}",
            analytics.current, analytics.average, analytics.min, analytics.max,
            analytics.p50, analytics.p95, analytics.p99
        );
        
        let analytics_paragraph = Paragraph::new(analytics_text)
            .style(Style::default().fg(Color::Cyan))
            .block(Block::default().borders(Borders::NONE));
        
        f.render_widget(analytics_paragraph, area);
    }

    fn render_empty_state(&self, f: &mut Frame, area: Rect) {
        let empty_text = "No data available yet...";
        let empty_paragraph = Paragraph::new(empty_text)
            .style(Style::default().fg(Color::Gray))
            .block(Block::default()
                .borders(Borders::ALL)
                .title(self.title.as_str()));
        
        f.render_widget(empty_paragraph, area);
    }
}

/// Analytics data for the graph
#[derive(Default, Debug)]
pub struct SimpleAnalytics {
    pub current: f64,
    pub average: f64,
    pub min: f64,
    pub max: f64,
    pub p50: f64,
    pub p95: f64,
    pub p99: f64,
    pub data_points: usize,
}

/// Calculate percentile
fn percentile(sorted_data: &[f64], p: f64) -> f64 {
    if sorted_data.is_empty() {
        return 0.0;
    }
    
    let index = (p * (sorted_data.len() - 1) as f64) as usize;
    sorted_data[index]
}

/// Professional QUIC Latency Graph
pub struct SimpleQuicLatencyGraph {
    graph: SimpleProfessionalGraph,
}

impl SimpleQuicLatencyGraph {
    pub fn new() -> Self {
        Self {
            graph: SimpleProfessionalGraph::new(
                "QUIC Latency (ms)".to_string(),
                100, // 100 data points
            ),
        }
    }

    pub fn add_latency(&mut self, latency: f64) {
        self.graph.add_data_point(latency);
    }

    pub fn render(&self, f: &mut Frame, area: Rect) {
        self.graph.render(f, area);
    }

    pub fn get_analytics(&self) -> SimpleAnalytics {
        self.graph.get_analytics()
    }
}

/// Professional QUIC Throughput Graph
pub struct SimpleQuicThroughputGraph {
    graph: SimpleProfessionalGraph,
}

impl SimpleQuicThroughputGraph {
    pub fn new() -> Self {
        Self {
            graph: SimpleProfessionalGraph::new(
                "QUIC Throughput (KB/s)".to_string(),
                100, // 100 data points
            ),
        }
    }

    pub fn add_throughput(&mut self, throughput: f64) {
        self.graph.add_data_point(throughput);
    }

    pub fn render(&self, f: &mut Frame, area: Rect) {
        self.graph.render(f, area);
    }

    pub fn get_analytics(&self) -> SimpleAnalytics {
        self.graph.get_analytics()
    }
}
