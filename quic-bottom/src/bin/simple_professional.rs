//! Simple Professional QUIC Bottom with advanced analytics
//! 
//! Features:
//! - Professional time graphs with analytics
//! - Advanced metrics and percentiles
//! - Real-time data visualization
//! - Simplified implementation

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
    widgets::{Block, Borders, Paragraph},
    Frame, Terminal,
};
use std::io;
use tokio::time::Duration;

use quic_bottom::{
    demo_data::DemoDataGenerator,
    simple_professional::{SimpleQuicLatencyGraph, SimpleQuicThroughputGraph},
};

/// Simple Professional QUIC Bottom application
pub struct SimpleProfessionalQuicBottom {
    latency_graph: SimpleQuicLatencyGraph,
    throughput_graph: SimpleQuicThroughputGraph,
    demo_generator: DemoDataGenerator,
    should_quit: bool,
    update_interval: Duration,
}

impl SimpleProfessionalQuicBottom {
    pub async fn new(interval_ms: u64) -> Result<Self> {
        Ok(Self {
            latency_graph: SimpleQuicLatencyGraph::new(),
            throughput_graph: SimpleQuicThroughputGraph::new(),
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

            // Update graphs with demo data
            self.update_graphs();

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

    fn update_graphs(&mut self) {
        // Generate demo data
        let (latency, throughput, _, _, _) = self.demo_generator.generate_next();

        // Update graphs
        self.latency_graph.add_latency(latency);
        self.throughput_graph.add_throughput(throughput);
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
                // Reset data
                self.latency_graph = SimpleQuicLatencyGraph::new();
                self.throughput_graph = SimpleQuicThroughputGraph::new();
                self.demo_generator = DemoDataGenerator::new();
            }
            KeyCode::Char('h') => {
                // Show help
                println!("Simple Professional QUIC Bottom Help:");
                println!("  q/ESC - Quit");
                println!("  r - Reset data");
                println!("  h - Show this help");
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

        // Main content - two professional graphs
        let main_chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Percentage(50), // Latency graph
                Constraint::Percentage(50), // Throughput graph
            ])
            .split(chunks[1]);

        // Render professional graphs
        self.latency_graph.render(f, main_chunks[0]);
        self.throughput_graph.render(f, main_chunks[1]);

        // Footer
        self.render_footer(f, chunks[2]);
    }

    fn render_header(&self, f: &mut Frame, area: Rect) {
        let header_text = "Simple Professional QUIC Bottom - Advanced Analytics";
        let header = Paragraph::new(header_text)
            .style(Style::default().fg(Color::White).add_modifier(Modifier::BOLD))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(header, area);
    }

    fn render_footer(&self, f: &mut Frame, area: Rect) {
        let footer_text = "Press 'q' to quit, 'r' to reset, 'h' for help";
        let footer = Paragraph::new(footer_text)
            .style(Style::default().fg(Color::Yellow))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(footer, area);
    }
}

#[tokio::main]
async fn main() -> Result<()> {
    env_logger::init();
    
    println!("🚀 Starting Simple Professional QUIC Bottom...");
    println!("📊 Advanced analytics and professional graphs!");
    println!("🎯 Based on bottom's capabilities but simplified!");
    println!("");
    println!("Features:");
    println!("  ✅ Professional time graphs");
    println!("  ✅ Advanced analytics (P50, P95, P99)");
    println!("  ✅ Real-time data visualization");
    println!("  ✅ Simplified implementation");
    println!("");
    println!("Controls:");
    println!("  q/ESC - Quit");
    println!("  r - Reset data");
    println!("  h - Show help");
    println!("");
    
    let mut app = SimpleProfessionalQuicBottom::new(100).await?;
    app.run().await?;
    
    println!("✅ Simple Professional QUIC Bottom completed!");
    Ok(())
}
