//! Enhanced Analytics QUIC Bottom
//! 
//! Features:
//! - Heatmaps for performance analysis
//! - Correlation analysis between metrics
//! - Anomaly detection
//! - Advanced visualizations

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
    heatmap_widget::QUICPerformanceHeatmap,
    correlation_widget::QUICCorrelationWidget,
    anomaly_detection::QUICAnomalyWidget,
};

/// Enhanced Analytics QUIC Bottom application
pub struct EnhancedAnalyticsQuicBottom {
    // Basic graphs
    latency_graph: SimpleQuicLatencyGraph,
    throughput_graph: SimpleQuicThroughputGraph,
    
    // Enhanced analytics
    performance_heatmap: QUICPerformanceHeatmap,
    correlation_widget: QUICCorrelationWidget,
    anomaly_widget: QUICAnomalyWidget,
    
    // Demo data
    demo_generator: DemoDataGenerator,
    
    // App state
    should_quit: bool,
    update_interval: Duration,
    current_view: ViewMode,
    time_slot: usize,
}

#[derive(Debug, Clone, PartialEq)]
enum ViewMode {
    Basic,
    Heatmap,
    Correlation,
    Anomaly,
    All,
}

impl EnhancedAnalyticsQuicBottom {
    pub async fn new(interval_ms: u64) -> Result<Self> {
        Ok(Self {
            latency_graph: SimpleQuicLatencyGraph::new(),
            throughput_graph: SimpleQuicThroughputGraph::new(),
            performance_heatmap: QUICPerformanceHeatmap::new(),
            correlation_widget: QUICCorrelationWidget::new(),
            anomaly_widget: QUICAnomalyWidget::new(),
            demo_generator: DemoDataGenerator::new(),
            should_quit: false,
            update_interval: Duration::from_millis(interval_ms),
            current_view: ViewMode::All,
            time_slot: 0,
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

            // Update all widgets with demo data
            self.update_all_widgets();

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

    fn update_all_widgets(&mut self) {
        // Generate demo data
        let (latency, throughput, connections, errors, packet_loss) = self.demo_generator.generate_next();

        // Update basic graphs
        self.latency_graph.add_latency(latency);
        self.throughput_graph.add_throughput(throughput);

        // Update enhanced analytics
        self.performance_heatmap.add_performance_data(self.time_slot, 0, latency);
        self.performance_heatmap.add_performance_data(self.time_slot, 1, throughput);
        self.performance_heatmap.add_performance_data(self.time_slot, 2, packet_loss as f64);
        self.performance_heatmap.add_performance_data(self.time_slot, 3, connections as f64);
        self.performance_heatmap.add_performance_data(self.time_slot, 4, errors as f64);

        // Update correlation data
        self.correlation_widget.add_metric_data("Latency".to_string(), latency);
        self.correlation_widget.add_metric_data("Throughput".to_string(), throughput);
        self.correlation_widget.add_metric_data("Packet Loss".to_string(), packet_loss as f64);
        self.correlation_widget.add_metric_data("Connections".to_string(), connections as f64);
        self.correlation_widget.add_metric_data("Errors".to_string(), errors as f64);
        self.correlation_widget.update_correlations();

        // Update anomaly detection
        self.anomaly_widget.add_quic_metric("Latency".to_string(), latency);
        self.anomaly_widget.add_quic_metric("Throughput".to_string(), throughput);
        self.anomaly_widget.add_quic_metric("Packet Loss".to_string(), packet_loss as f64);
        self.anomaly_widget.add_quic_metric("Connections".to_string(), connections as f64);
        self.anomaly_widget.add_quic_metric("Errors".to_string(), errors as f64);

        // Update time slot
        self.time_slot = (self.time_slot + 1) % 20;
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
                // Reset all data
                self.latency_graph = SimpleQuicLatencyGraph::new();
                self.throughput_graph = SimpleQuicThroughputGraph::new();
                self.performance_heatmap = QUICPerformanceHeatmap::new();
                self.correlation_widget = QUICCorrelationWidget::new();
                self.anomaly_widget = QUICAnomalyWidget::new();
                self.demo_generator = DemoDataGenerator::new();
                self.time_slot = 0;
            }
            KeyCode::Char('h') => {
                self.show_help();
            }
            KeyCode::Char('1') => {
                self.current_view = ViewMode::Basic;
            }
            KeyCode::Char('2') => {
                self.current_view = ViewMode::Heatmap;
            }
            KeyCode::Char('3') => {
                self.current_view = ViewMode::Correlation;
            }
            KeyCode::Char('4') => {
                self.current_view = ViewMode::Anomaly;
            }
            KeyCode::Char('a') => {
                self.current_view = ViewMode::All;
            }
            _ => {}
        }
    }

    fn show_help(&self) {
        println!("Enhanced Analytics QUIC Bottom Help:");
        println!("  q/ESC - Quit");
        println!("  r - Reset all data");
        println!("  h - Show this help");
        println!("  1 - Basic graphs view");
        println!("  2 - Performance heatmap view");
        println!("  3 - Correlation analysis view");
        println!("  4 - Anomaly detection view");
        println!("  a - All views (default)");
    }

    fn ui(&self, f: &mut Frame) {
        match self.current_view {
            ViewMode::Basic => self.render_basic_view(f),
            ViewMode::Heatmap => self.render_heatmap_view(f),
            ViewMode::Correlation => self.render_correlation_view(f),
            ViewMode::Anomaly => self.render_anomaly_view(f),
            ViewMode::All => self.render_all_view(f),
        }
    }

    fn render_basic_view(&self, f: &mut Frame) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Header
                Constraint::Min(0),    // Main content
                Constraint::Length(3), // Footer
            ])
            .split(f.area());

        self.render_header(f, chunks[0], "Basic QUIC Graphs");

        let main_chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Percentage(50), // Latency
                Constraint::Percentage(50), // Throughput
            ])
            .split(chunks[1]);

        self.latency_graph.render(f, main_chunks[0]);
        self.throughput_graph.render(f, main_chunks[1]);

        self.render_footer(f, chunks[2]);
    }

    fn render_heatmap_view(&self, f: &mut Frame) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Header
                Constraint::Min(0),    // Heatmap
                Constraint::Length(3), // Footer
            ])
            .split(f.area());

        self.render_header(f, chunks[0], "Performance Heatmap");
        self.performance_heatmap.render(f, chunks[1]);
        self.render_footer(f, chunks[2]);
    }

    fn render_correlation_view(&self, f: &mut Frame) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Header
                Constraint::Min(0),    // Correlation
                Constraint::Length(3), // Footer
            ])
            .split(f.area());

        self.render_header(f, chunks[0], "Metrics Correlation Analysis");
        self.correlation_widget.render(f, chunks[1]);
        self.render_footer(f, chunks[2]);
    }

    fn render_anomaly_view(&self, f: &mut Frame) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Header
                Constraint::Min(0),    // Anomalies
                Constraint::Length(3), // Footer
            ])
            .split(f.area());

        self.render_header(f, chunks[0], "Anomaly Detection");
        self.anomaly_widget.render(f, chunks[1]);
        self.render_footer(f, chunks[2]);
    }

    fn render_all_view(&self, f: &mut Frame) {
        let chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Length(3), // Header
                Constraint::Min(0),    // Main content
                Constraint::Length(3), // Footer
            ])
            .split(f.area());

        self.render_header(f, chunks[0], "Enhanced Analytics - All Views");

        let main_chunks = Layout::default()
            .direction(Direction::Horizontal)
            .constraints([
                Constraint::Percentage(50), // Left column
                Constraint::Percentage(50), // Right column
            ])
            .split(chunks[1]);

        let left_chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Percentage(50), // Latency
                Constraint::Percentage(50), // Throughput
            ])
            .split(main_chunks[0]);

        let right_chunks = Layout::default()
            .direction(Direction::Vertical)
            .constraints([
                Constraint::Percentage(50), // Heatmap
                Constraint::Percentage(50), // Anomaly
            ])
            .split(main_chunks[1]);

        self.latency_graph.render(f, left_chunks[0]);
        self.throughput_graph.render(f, left_chunks[1]);
        self.performance_heatmap.render(f, right_chunks[0]);
        self.anomaly_widget.render(f, right_chunks[1]);

        self.render_footer(f, chunks[2]);
    }

    fn render_header(&self, f: &mut Frame, area: Rect, title: &str) {
        let header_text = format!("Enhanced Analytics QUIC Bottom - {}", title);
        let header = Paragraph::new(header_text)
            .style(Style::default().fg(Color::White).add_modifier(Modifier::BOLD))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(header, area);
    }

    fn render_footer(&self, f: &mut Frame, area: Rect) {
        let footer_text = "Press 'q' to quit, 'r' to reset, 'h' for help, '1-4' for views, 'a' for all";
        let footer = Paragraph::new(footer_text)
            .style(Style::default().fg(Color::Yellow))
            .block(Block::default().borders(Borders::ALL));
        f.render_widget(footer, area);
    }
}

#[tokio::main]
async fn main() -> Result<()> {
    env_logger::init();
    
    println!("🚀 Starting Enhanced Analytics QUIC Bottom...");
    println!("📊 Advanced analytics with heatmaps, correlation, and anomaly detection!");
    println!("🎯 Professional visualizations based on bottom's capabilities!");
    println!("");
    println!("Features:");
    println!("  ✅ Performance heatmaps");
    println!("  ✅ Correlation analysis");
    println!("  ✅ Anomaly detection");
    println!("  ✅ Advanced visualizations");
    println!("  ✅ Interactive view switching");
    println!("");
    println!("Controls:");
    println!("  q/ESC - Quit");
    println!("  r - Reset data");
    println!("  h - Show help");
    println!("  1 - Basic graphs");
    println!("  2 - Performance heatmap");
    println!("  3 - Correlation analysis");
    println!("  4 - Anomaly detection");
    println!("  a - All views");
    println!("");
    
    let mut app = EnhancedAnalyticsQuicBottom::new(100).await?;
    app.run().await?;
    
    println!("✅ Enhanced Analytics QUIC Bottom completed!");
    Ok(())
}
