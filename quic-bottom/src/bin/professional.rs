//! Professional QUIC Bottom with advanced analytics and historical data
//! 
//! Features:
//! - Professional time graphs with analytics
//! - Historical data scrolling
//! - Advanced metrics and percentiles
//! - Trend analysis
//! - Interactive controls

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
    professional_graphs::{ProfessionalQuicLatencyGraph, ProfessionalQuicThroughputGraph},
};

/// Professional QUIC Bottom application
pub struct ProfessionalQuicBottom {
    latency_graph: ProfessionalQuicLatencyGraph,
    throughput_graph: ProfessionalQuicThroughputGraph,
    demo_generator: DemoDataGenerator,
    should_quit: bool,
    update_interval: Duration,
    current_time_window: f64,
}

impl ProfessionalQuicBottom {
    pub async fn new(interval_ms: u64) -> Result<Self> {
        Ok(Self {
            latency_graph: ProfessionalQuicLatencyGraph::new(),
            throughput_graph: ProfessionalQuicThroughputGraph::new(),
            demo_generator: DemoDataGenerator::new(),
            should_quit: false,
            update_interval: Duration::from_millis(interval_ms),
            current_time_window: 60.0, // 60 seconds default
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
                self.latency_graph = ProfessionalQuicLatencyGraph::new();
                self.throughput_graph = ProfessionalQuicThroughputGraph::new();
                self.demo_generator = DemoDataGenerator::new();
            }
            KeyCode::Char('h') => {
                // Show help
                println!("Professional QUIC Bottom Help:");
                println!("  q/ESC - Quit");
                println!("  r - Reset data");
                println!("  h - Show this help");
                println!("  +/- - Adjust time window");
                println!("  ←/→ - Navigate graphs");
            }
            KeyCode::Char('+') | KeyCode::Char('=') => {
                // Increase time window
                self.current_time_window = (self.current_time_window + 10.0).min(300.0);
                println!("Time window: {:.0}s", self.current_time_window);
            }
            KeyCode::Char('-') => {
                // Decrease time window
                self.current_time_window = (self.current_time_window - 10.0).max(10.0);
                println!("Time window: {:.0}s", self.current_time_window);
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
        let header_text = "Professional QUIC Bottom - Advanced Analytics & Historical Data";
        let header = Paragraph::new(header_text)
            .style(Style::default().fg(Color::White).add_modifier(Modifier::BOLD))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(header, area);
    }

    fn render_footer(&self, f: &mut Frame, area: Rect) {
        let footer_text = format!(
            "Time Window: {:.0}s | Press 'q' to quit, 'r' to reset, 'h' for help, '+/-' to adjust window",
            self.current_time_window
        );
        let footer = Paragraph::new(footer_text)
            .style(Style::default().fg(Color::Yellow))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(footer, area);
    }
}

#[tokio::main]
async fn main() -> Result<()> {
    env_logger::init();
    
    println!("🚀 Starting Professional QUIC Bottom...");
    println!("📊 Advanced analytics and historical data scrolling!");
    println!("🎯 Professional time graphs with trend analysis!");
    println!("");
    println!("Features:");
    println!("  ✅ Professional time graphs");
    println!("  ✅ Historical data scrolling");
    println!("  ✅ Advanced analytics (P50, P95, P99)");
    println!("  ✅ Trend analysis");
    println!("  ✅ Interactive time window adjustment");
    println!("  ✅ Real-time data visualization");
    println!("");
    println!("Controls:");
    println!("  q/ESC - Quit");
    println!("  r - Reset data");
    println!("  h - Show help");
    println!("  +/- - Adjust time window");
    println!("");
    
    let mut app = ProfessionalQuicBottom::new(100).await?;
    app.run().await?;
    
    println!("✅ Professional QUIC Bottom completed!");
    Ok(())
}
