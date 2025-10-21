//! Improved layout for QUIC Bottom with better spacing
//! 
//! This module provides better spacing between widgets to prevent them from "sticking together"

use ratatui::{
    layout::{Constraint, Direction, Layout, Rect},
    style::{Color, Modifier, Style},
    widgets::{Block, Borders, Paragraph, Sparkline},
    Frame,
};
use std::collections::VecDeque;

/// Improved layout with better spacing
pub fn create_improved_layout(area: Rect) -> Vec<Rect> {
    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Length(3), // Header
            Constraint::Length(1), // Spacer
            Constraint::Min(0),    // Main content
            Constraint::Length(1), // Spacer
            Constraint::Length(3), // Footer
        ])
        .split(area);

    // Main content with horizontal spacing
    let main_chunks = Layout::default()
        .direction(Direction::Horizontal)
        .constraints([
            Constraint::Percentage(45), // Left column
            Constraint::Length(3),      // Horizontal spacer
            Constraint::Percentage(45), // Right column
        ])
        .split(chunks[2]);

    // Left column with vertical spacing
    let left_chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Percentage(45), // Latency
            Constraint::Length(2),      // Vertical spacer
            Constraint::Percentage(45), // Throughput
        ])
        .split(main_chunks[0]);

    // Right column with vertical spacing
    let right_chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Percentage(45), // Connections
            Constraint::Length(2),      // Vertical spacer
            Constraint::Percentage(45), // Network
        ])
        .split(main_chunks[2]);

    vec![
        chunks[0],      // Header
        left_chunks[0], // Latency
        left_chunks[2], // Throughput
        right_chunks[0], // Connections
        right_chunks[2], // Network
        chunks[4],      // Footer
    ]
}

/// Render spacer between widgets
pub fn render_spacer(f: &mut Frame, area: Rect) {
    let spacer = Paragraph::new("")
        .block(Block::default().borders(Borders::NONE));
    f.render_widget(spacer, area);
}

/// Improved sparkline with better borders and spacing
pub fn render_improved_sparkline(
    f: &mut Frame,
    area: Rect,
    data: &VecDeque<f64>,
    title: &str,
    color: Color,
) {
    if !data.is_empty() {
        let sparkline = Sparkline::default()
            .data(&data.iter().map(|&x| x as u64).collect::<Vec<u64>>())
            .style(Style::default().fg(color))
            .block(Block::default()
                .borders(Borders::ALL)
                .title(title)
                .title_style(Style::default().fg(color).add_modifier(Modifier::BOLD)));
        f.render_widget(sparkline, area);
    } else {
        // Show placeholder when no data
        let placeholder = Paragraph::new("No data yet...")
            .style(Style::default().fg(Color::Gray))
            .block(Block::default()
                .borders(Borders::ALL)
                .title(title)
                .title_style(Style::default().fg(color).add_modifier(Modifier::BOLD)));
        f.render_widget(placeholder, area);
    }
}
