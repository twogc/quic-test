//! QUIC Bottom Demo - Shows dynamic graphs with test data
//! 
//! This demo version generates realistic QUIC metrics to showcase
//! the dynamic graphs and widgets

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

use quic_bottom::{
    demo_data::DemoDataGenerator,
    widgets::{QUICConnectionWidget, QUICLatencyWidget, QUICNetworkWidget, QUICThroughputWidget},
};

/// Demo application with test data
pub struct QuicBottomDemo {
    latency_widget: QUICLatencyWidget,
    throughput_widget: QUICThroughputWidget,
    connection_widget: QUICConnectionWidget,
    network_widget: QUICNetworkWidget,
    demo_generator: DemoDataGenerator,
    should_quit: bool,
    update_interval: Duration,
}

impl QuicBottomDemo {
    pub async fn new(interval_ms: u64) -> Result<Self> {
        Ok(Self {
            latency_widget: QUICLatencyWidget::new(1000),
            throughput_widget: QUICThroughputWidget::new(1000),
            connection_widget: QUICConnectionWidget::new(),
            network_widget: QUICNetworkWidget::new(),
            demo_generator: DemoDataGenerator::new(),
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

            // Update widgets with demo data
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
        // Generate demo data
        let (latency, throughput, handshake_time, packet_loss, retransmits) = 
            self.demo_generator.generate_next();

        // Update latency widget
        self.latency_widget.update(latency);

        // Update throughput widget
        self.throughput_widget.update(throughput);

        // Update connection widget
        let connections = 2 + (self.demo_generator.counter / 10) as i32;
        let errors = if self.demo_generator.counter > 20 { 
            (self.demo_generator.counter % 10) as i32 
        } else { 
            0 
        };
        self.connection_widget.update(connections, errors, connections + errors);
        self.connection_widget.add_handshake_time(handshake_time);

        // Update network widget
        self.network_widget.update(packet_loss, retransmits, "BBRv2".to_string());
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
                // Reset demo data
                self.demo_generator = DemoDataGenerator::new();
                self.latency_widget = QUICLatencyWidget::new(1000);
                self.throughput_widget = QUICThroughputWidget::new(1000);
                self.connection_widget = QUICConnectionWidget::new();
                self.network_widget = QUICNetworkWidget::new();
            }
            KeyCode::Char('h') => {
                // Show help
                println!("Help: q/ESC to quit, r to reset, h for help");
            }
            _ => {}
        }
    }

    fn ui(&self, f: &mut Frame) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Header
                Constraint::Min(0),    // Main content
                Constraint::Length(3), // Footer
            ])
            .split(f.area());

        // Header
        self.render_header(f, chunks[0]);

        // Main content
        let main_chunks = Layout::default()
            .direction(Direction::Horizontal)
            .constraints([
                Constraint::Percentage(50), // Left column
                Constraint::Percentage(50), // Right column
            ])
            .split(chunks[1]);

        // Left column
        let left_chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Percentage(50), // Latency
                Constraint::Percentage(50), // Throughput
            ])
            .split(main_chunks[0]);

        // Right column
        let right_chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Percentage(50), // Connections
                Constraint::Percentage(50), // Network
            ])
            .split(main_chunks[1]);

        // Render widgets
        self.latency_widget.render(f, left_chunks[0]);
        self.throughput_widget.render(f, left_chunks[1]);
        self.connection_widget.render(f, right_chunks[0]);
        self.network_widget.render(f, right_chunks[1]);

        // Footer
        self.render_footer(f, chunks[2]);
    }

    fn render_header(&self, f: &mut Frame, area: Rect) {
        let header_text = "QUIC Bottom DEMO - Dynamic Graphs with Test Data";
        let header = Paragraph::new(header_text)
            .style(Style::default().fg(Color::White).add_modifier(Modifier::BOLD))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(header, area);
    }

    fn render_footer(&self, f: &mut Frame, area: Rect) {
        let footer_text = "Press 'q' to quit, 'r' to reset, 'h' for help | DEMO MODE";
        let footer = Paragraph::new(footer_text)
            .style(Style::default().fg(Color::Yellow))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(footer, area);
    }
}

#[tokio::main]
async fn main() -> Result<()> {
    env_logger::init();
    
    println!("🚀 Starting QUIC Bottom DEMO with dynamic graphs...");
    println!("📊 This demo shows realistic QUIC metrics with live graphs!");
    println!("🎯 Watch the sparkline graphs update in real-time!");
    println!("");
    
    let mut demo = QuicBottomDemo::new(100).await?;
    demo.run().await?;
    
    println!("✅ QUIC Bottom DEMO completed!");
    Ok(())
}
