//! Professional graphs with analytics and historical data scrolling
//! 
//! Based on bottom's advanced time graph capabilities

use ratatui::{
    layout::{Constraint, Direction, Layout, Rect},
    style::{Color, Modifier, Style},
    symbols::Marker,
    text::{Line, Span},
    widgets::{Block, Borders, Dataset, GraphType, Paragraph},
    Frame,
};
use std::collections::VecDeque;
use std::time::Instant;

/// Professional time graph for QUIC metrics
pub struct ProfessionalTimeGraph {
    /// Historical data points
    pub data_points: VecDeque<(Instant, f64)>,
    
    /// Maximum number of data points to keep
    pub max_points: usize,
    
    /// Current time window (seconds)
    pub time_window: f64,
    
    /// Y-axis bounds
    pub y_bounds: (f64, f64),
    
    /// Graph style
    pub style: Style,
    
    /// Title
    pub title: String,
    
    /// Whether graph is selected
    pub is_selected: bool,
    
    /// Whether graph is expanded
    pub is_expanded: bool,
}

impl ProfessionalTimeGraph {
    pub fn new(title: String, max_points: usize, time_window: f64) -> Self {
        Self {
            data_points: VecDeque::with_capacity(max_points),
            max_points,
            time_window,
            y_bounds: (0.0, 100.0),
            style: Style::default().fg(Color::Green),
            title,
            is_selected: false,
            is_expanded: false,
        }
    }

    /// Add new data point
    pub fn add_data_point(&mut self, value: f64) {
        let now = Instant::now();
        self.data_points.push_back((now, value));
        
        // Keep only recent data within time window
        while let Some(&(time, _)) = self.data_points.front() {
            match now.duration_since(time) {
                Ok(duration) => {
                    if duration.as_secs_f64() > self.time_window {
                        self.data_points.pop_front();
                    } else {
                        break;
                    }
                }
                Err(_) => break,
            }
        }
        
        // Also limit by max_points
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
        
        let values: Vec<f64> = self.data_points.iter().map(|(_, v)| *v).collect();
        let min_val = values.iter().fold(f64::INFINITY, |a, &b| a.min(b));
        let max_val = values.iter().fold(f64::NEG_INFINITY, |a, &b| a.max(b));
        
        // Add some padding
        let padding = (max_val - min_val) * 0.1;
        self.y_bounds = (min_val - padding, max_val + padding);
    }

    /// Get analytics for the current data
    pub fn get_analytics(&self) -> GraphAnalytics {
        if self.data_points.is_empty() {
            return GraphAnalytics::default();
        }
        
        let values: Vec<f64> = self.data_points.iter().map(|(_, v)| *v).collect();
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
        
        // Calculate trend (simple linear regression)
        let trend = calculate_trend(&values);
        
        GraphAnalytics {
            current,
            average,
            min,
            max,
            p50,
            p95,
            p99,
            trend,
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
        let now = Instant::now();
        let time_start = -self.time_window;
        
        // Convert data to chart format
        let data: Vec<(f64, f64)> = self.data_points
            .iter()
            .filter_map(|(time, value)| {
                now.duration_since(*time).ok().map(|duration| {
                    let time_offset = duration.as_secs_f64();
                    (-time_offset, *value)
                })
            })
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
                .bounds([time_start, 0.0])
                .labels(vec![
                    Span::styled(format!("{:.0}s", -self.time_window), self.style),
                    Span::styled("0s", self.style),
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
            "Current: {:.2} | Avg: {:.2} | Min: {:.2} | Max: {:.2} | P50: {:.2} | P95: {:.2} | P99: {:.2} | Trend: {:.2}",
            analytics.current, analytics.average, analytics.min, analytics.max,
            analytics.p50, analytics.p95, analytics.p99, analytics.trend
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
pub struct GraphAnalytics {
    pub current: f64,
    pub average: f64,
    pub min: f64,
    pub max: f64,
    pub p50: f64,
    pub p95: f64,
    pub p99: f64,
    pub trend: f64,
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

/// Calculate trend using simple linear regression
fn calculate_trend(values: &[f64]) -> f64 {
    if values.len() < 2 {
        return 0.0;
    }
    
    let n = values.len() as f64;
    let x_mean = (n - 1.0) / 2.0;
    let y_mean = values.iter().sum::<f64>() / n;
    
    let mut numerator = 0.0;
    let mut denominator = 0.0;
    
    for (i, &y) in values.iter().enumerate() {
        let x = i as f64;
        numerator += (x - x_mean) * (y - y_mean);
        denominator += (x - x_mean).powi(2);
    }
    
    if denominator == 0.0 {
        0.0
    } else {
        numerator / denominator
    }
}

/// Professional QUIC Latency Graph
pub struct ProfessionalQuicLatencyGraph {
    graph: ProfessionalTimeGraph,
}

impl ProfessionalQuicLatencyGraph {
    pub fn new() -> Self {
        Self {
            graph: ProfessionalTimeGraph::new(
                "QUIC Latency (ms)".to_string(),
                1000,
                60.0, // 60 seconds window
            ),
        }
    }

    pub fn add_latency(&mut self, latency: f64) {
        self.graph.add_data_point(latency);
    }

    pub fn render(&self, f: &mut Frame, area: Rect) {
        self.graph.render(f, area);
    }

    pub fn get_analytics(&self) -> GraphAnalytics {
        self.graph.get_analytics()
    }
}

/// Professional QUIC Throughput Graph
pub struct ProfessionalQuicThroughputGraph {
    graph: ProfessionalTimeGraph,
}

impl ProfessionalQuicThroughputGraph {
    pub fn new() -> Self {
        Self {
            graph: ProfessionalTimeGraph::new(
                "QUIC Throughput (KB/s)".to_string(),
                1000,
                60.0, // 60 seconds window
            ),
        }
    }

    pub fn add_throughput(&mut self, throughput: f64) {
        self.graph.add_data_point(throughput);
    }

    pub fn render(&self, f: &mut Frame, area: Rect) {
        self.graph.render(f, area);
    }

    pub fn get_analytics(&self) -> GraphAnalytics {
        self.graph.get_analytics()
    }
}
