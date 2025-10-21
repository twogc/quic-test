//! Simplified QUIC Bottom application
//! 
//! A minimal TUI application for QUIC monitoring

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
    widgets::{Block, Borders, Paragraph, Sparkline},
    Frame, Terminal,
};
use std::io;
use tokio::time::Duration;

use crate::{
    metrics::{get_current_metrics, init_metrics},
    widgets::{QUICConnectionWidget, QUICLatencyWidget, QUICNetworkWidget, QUICThroughputWidget},
    improved_layout::{create_improved_layout, render_spacer},
};

/// Main application state for QUIC Bottom
pub struct QuicBottomApp {
    latency_widget: QUICLatencyWidget,
    throughput_widget: QUICThroughputWidget,
    connection_widget: QUICConnectionWidget,
    network_widget: QUICNetworkWidget,
    should_quit: bool,
    update_interval: Duration,
}

impl QuicBottomApp {
    pub async fn new(interval_ms: u64) -> Result<Self> {
        // Initialize metrics system
        init_metrics()?;

        Ok(Self {
            latency_widget: QUICLatencyWidget::new(1000),
            throughput_widget: QUICThroughputWidget::new(1000),
            connection_widget: QUICConnectionWidget::new(),
            network_widget: QUICNetworkWidget::new(),
            should_quit: false,
            update_interval: Duration::from_millis(interval_ms),
        })
    }

    pub async fn run(&mut self) -> Result<()> {
        // Setup terminal
        enable_raw_mode()?;
        let mut stdout = io::stdout();
        execute!(stdout, EnterAlternateScreen, EnableMouseCapture)?;
        let backend = CrosstermBackend::new(stdout);
        let mut terminal = Terminal::new(backend)?;

        // Main event loop
        loop {
            if self.should_quit {
                break;
            }

            // Update widgets with latest metrics
            self.update_widgets();

            // Render the UI
            terminal.draw(|f| self.ui(f))?;

            // Handle events
            if event::poll(self.update_interval)? {
                if let Event::Key(key) = event::read()? {
                    self.handle_key_event(key);
                }
            }
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

    fn update_widgets(&mut self) {
        if let Some(metrics) = get_current_metrics() {
            // Update latency widget
            self.latency_widget.update(metrics.latency);

            // Update throughput widget
            self.throughput_widget.update(metrics.throughput);

            // Update connection widget
            self.connection_widget.update(
                metrics.connections,
                metrics.errors,
                metrics.connections + metrics.errors,
            );

            // Update network widget
            self.network_widget.update(
                metrics.packet_loss,
                metrics.retransmits,
                "BBRv2".to_string(), // TODO: Get actual CC algorithm
            );
        }
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
                // Refresh/reset metrics
                log::info!("Refreshing metrics");
            }
            KeyCode::Char('h') => {
                // Show help
                log::info!("Help: q/ESC to quit, r to refresh, h for help");
            }
            _ => {}
        }
    }

    fn ui(&self, f: &mut Frame) {
        let chunks = create_improved_layout(f.area());

        // Header
        self.render_header(f, chunks[0]);

        // Render widgets with better spacing
        self.latency_widget.render(f, chunks[1]);
        self.throughput_widget.render(f, chunks[2]);
        self.connection_widget.render(f, chunks[3]);
        self.network_widget.render(f, chunks[4]);

        // Footer
        self.render_footer(f, chunks[5]);
    }

    fn render_header(&self, f: &mut Frame, area: Rect) {
        let header_text = "QUIC Bottom - Real-time QUIC Protocol Monitor";
        let header = Paragraph::new(header_text)
            .style(Style::default().fg(Color::White).add_modifier(Modifier::BOLD))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(header, area);
    }

    fn render_footer(&self, f: &mut Frame, area: Rect) {
        let footer_text = "Press 'q' to quit, 'r' to refresh, 'h' for help";
        let footer = Paragraph::new(footer_text)
            .style(Style::default().fg(Color::Gray))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(footer, area);
    }
}