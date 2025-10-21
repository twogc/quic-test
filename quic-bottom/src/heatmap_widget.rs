//! Heatmap widget for performance analysis
//! 
//! Shows performance metrics over time with color-coded intensity

use ratatui::{
    layout::{Constraint, Direction, Layout, Rect},
    style::{Color, Modifier, Style},
    text::{Line, Span},
    widgets::{Block, Borders, Paragraph},
    Frame,
};
use std::collections::VecDeque;

/// Heatmap data point
#[derive(Debug, Clone)]
pub struct HeatmapPoint {
    pub x: usize,
    pub y: usize,
    pub value: f64,
    pub timestamp: chrono::DateTime<chrono::Utc>,
}

/// Heatmap widget for performance visualization
pub struct HeatmapWidget {
    /// Data points for the heatmap
    pub data: VecDeque<HeatmapPoint>,
    
    /// Maximum number of data points
    pub max_points: usize,
    
    /// Heatmap dimensions
    pub width: usize,
    pub height: usize,
    
    /// Value range for color mapping
    pub min_value: f64,
    pub max_value: f64,
    
    /// Title
    pub title: String,
}

impl HeatmapWidget {
    pub fn new(title: String, width: usize, height: usize) -> Self {
        Self {
            data: VecDeque::with_capacity(width * height),
            max_points: width * height,
            width,
            height,
            min_value: 0.0,
            max_value: 100.0,
            title,
        }
    }

    /// Add a data point to the heatmap
    pub fn add_data_point(&mut self, x: usize, y: usize, value: f64) {
        let point = HeatmapPoint {
            x,
            y,
            value,
            timestamp: chrono::Utc::now(),
        };
        
        self.data.push_back(point);
        
        // Keep only recent data
        while self.data.len() > self.max_points {
            self.data.pop_front();
        }
        
        // Update value range
        self.update_value_range();
    }

    /// Update the value range for color mapping
    fn update_value_range(&mut self) {
        if self.data.is_empty() {
            return;
        }
        
        let values: Vec<f64> = self.data.iter().map(|p| p.value).collect();
        self.min_value = values.iter().fold(f64::INFINITY, |a, &b| a.min(b));
        self.max_value = values.iter().fold(f64::NEG_INFINITY, |a, &b| a.max(b));
    }

    /// Get color for a value based on the range
    fn get_color_for_value(&self, value: f64) -> Color {
        if self.max_value == self.min_value {
            return Color::Gray;
        }
        
        let normalized = (value - self.min_value) / (self.max_value - self.min_value);
        
        match normalized {
            x if x < 0.2 => Color::Green,
            x if x < 0.4 => Color::LightGreen,
            x if x < 0.6 => Color::Yellow,
            x if x < 0.8 => Color::LightRed,
            _ => Color::Red,
        }
    }

    /// Render the heatmap
    pub fn render(&self, f: &mut Frame, area: Rect) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Title
                Constraint::Min(0),    // Heatmap
                Constraint::Length(3), // Legend
            ])
            .split(area);

        // Title
        self.render_title(f, chunks[0]);
        
        // Heatmap
        self.render_heatmap(f, chunks[1]);
        
        // Legend
        self.render_legend(f, chunks[2]);
    }

    fn render_title(&self, f: &mut Frame, area: Rect) {
        let title = Paragraph::new(self.title.clone())
            .style(Style::default().fg(Color::White).add_modifier(Modifier::BOLD))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(title, area);
    }

    fn render_heatmap(&self, f: &mut Frame, area: Rect) {
        if self.data.is_empty() {
            let empty_text = "No data available yet...";
            let empty_paragraph = Paragraph::new(empty_text)
                .style(Style::default().fg(Color::Gray))
                .block(Block::default().borders(Borders::ALL));
            f.render_widget(empty_paragraph, area);
            return;
        }

        // Create a 2D grid for the heatmap
        let mut grid = vec![vec![0.0; self.width]; self.height];
        
        // Fill the grid with data
        for point in &self.data {
            if point.x < self.width && point.y < self.height {
                grid[point.y][point.x] = point.value;
            }
        }

        // Render the heatmap
        let mut lines = Vec::new();
        for y in 0..self.height {
            let mut line_spans = Vec::new();
            for x in 0..self.width {
                let value = grid[y][x];
                let color = self.get_color_for_value(value);
                let char = if value > 0.0 { "█" } else { " " };
                line_spans.push(Span::styled(char, Style::default().fg(color)));
            }
            lines.push(Line::from(line_spans));
        }

        let heatmap_paragraph = Paragraph::new(lines)
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(heatmap_paragraph, area);
    }

    fn render_legend(&self, f: &mut Frame, area: Rect) {
        let legend_text = format!(
            "Range: {:.1} - {:.1} | Green: Low | Yellow: Medium | Red: High",
            self.min_value, self.max_value
        );
        
        let legend = Paragraph::new(legend_text)
            .style(Style::default().fg(Color::Cyan))
            .block(Block::default().borders(Borders::NONE));
        
        f.render_widget(legend, area);
    }
}

/// Performance Heatmap for QUIC metrics
pub struct QUICPerformanceHeatmap {
    heatmap: HeatmapWidget,
    time_slots: usize,
    metric_slots: usize,
}

impl QUICPerformanceHeatmap {
    pub fn new() -> Self {
        Self {
            heatmap: HeatmapWidget::new(
                "QUIC Performance Heatmap".to_string(),
                20, // 20 time slots
                10, // 10 metric slots (latency, throughput, etc.)
            ),
            time_slots: 20,
            metric_slots: 10,
        }
    }

    /// Add performance data
    pub fn add_performance_data(&mut self, time_slot: usize, metric_slot: usize, value: f64) {
        self.heatmap.add_data_point(time_slot, metric_slot, value);
    }

    /// Render the performance heatmap
    pub fn render(&self, f: &mut Frame, area: Rect) {
        self.heatmap.render(f, area);
    }
}
