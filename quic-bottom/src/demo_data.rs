//! Demo data generator for QUIC Bottom
//! 
//! Generates realistic QUIC metrics for demonstration

use std::collections::VecDeque;
use rand::Rng;

/// Demo data generator
pub struct DemoDataGenerator {
    latency_data: VecDeque<f64>,
    throughput_data: VecDeque<f64>,
    handshake_data: VecDeque<f64>,
    loss_data: VecDeque<f64>,
    retransmit_data: VecDeque<i32>,
    pub counter: u32,
}

impl DemoDataGenerator {
    pub fn new() -> Self {
        Self {
            latency_data: VecDeque::with_capacity(1000),
            throughput_data: VecDeque::with_capacity(1000),
            handshake_data: VecDeque::with_capacity(100),
            loss_data: VecDeque::with_capacity(1000),
            retransmit_data: VecDeque::with_capacity(1000),
            counter: 0,
        }
    }

    pub fn generate_next(&mut self) -> (f64, f64, f64, f64, i32) {
        let mut rng = rand::thread_rng();
        self.counter += 1;

        // Generate realistic QUIC metrics
        let base_latency = 10.0;
        let latency_variation = rng.gen_range(-5.0..15.0);
        let latency_trend = (self.counter as f64 * 0.1).sin() * 3.0;
        let latency = base_latency + latency_variation + latency_trend;

        let base_throughput = 1000.0;
        let throughput_variation = rng.gen_range(-200.0..500.0);
        let throughput_trend = (self.counter as f64 * 0.05).cos() * 200.0;
        let throughput = base_throughput + throughput_variation + throughput_trend;

        let handshake_time = rng.gen_range(50.0..200.0) + (self.counter as f64 * 0.2).sin() * 20.0;

        let packet_loss = if self.counter > 20 {
            rng.gen_range(0.0..2.0) + (self.counter as f64 * 0.1).sin() * 0.5
        } else {
            0.0
        };

        let retransmits = if self.counter > 25 {
            rng.gen_range(0..10) + ((self.counter as f64 * 0.15).sin() * 3.0) as i32
        } else {
            0
        };

        // Update data buffers
        self.latency_data.push_back(latency);
        self.throughput_data.push_back(throughput);
        self.handshake_data.push_back(handshake_time);
        self.loss_data.push_back(packet_loss);
        self.retransmit_data.push_back(retransmits);

        // Keep only last 1000 points
        if self.latency_data.len() > 1000 {
            self.latency_data.pop_front();
        }
        if self.throughput_data.len() > 1000 {
            self.throughput_data.pop_front();
        }
        if self.handshake_data.len() > 100 {
            self.handshake_data.pop_front();
        }
        if self.loss_data.len() > 1000 {
            self.loss_data.pop_front();
        }
        if self.retransmit_data.len() > 1000 {
            self.retransmit_data.pop_front();
        }

        (latency, throughput, handshake_time, packet_loss, retransmits)
    }

    pub fn get_latency_data(&self) -> &VecDeque<f64> {
        &self.latency_data
    }

    pub fn get_throughput_data(&self) -> &VecDeque<f64> {
        &self.throughput_data
    }

    pub fn get_handshake_data(&self) -> &VecDeque<f64> {
        &self.handshake_data
    }

    pub fn get_loss_data(&self) -> &VecDeque<f64> {
        &self.loss_data
    }

    pub fn get_retransmit_data(&self) -> &VecDeque<i32> {
        &self.retransmit_data
    }
}
