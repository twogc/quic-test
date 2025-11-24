package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"quic-test/internal"
)

func main() {
	var (
		preset     = flag.String("preset", "good", "Network simulation preset (excellent, good, poor, mobile, satellite, adversarial)")
		latency    = flag.Duration("latency", 20*time.Millisecond, "Network latency")
		jitter     = flag.Duration("jitter", 5*time.Millisecond, "Network jitter")
		loss       = flag.Float64("loss", 0.01, "Packet loss rate (0.0 to 1.0)")
		bandwidth  = flag.Int64("bandwidth", 100*1024*1024, "Bandwidth in bytes per second")
		duration   = flag.Duration("duration", 0, "Simulation duration (0 = infinite)")
		duplication = flag.Float64("duplication", 0, "Packet duplication rate")
		reordering = flag.Bool("reordering", false, "Enable packet reordering")
		burst      = flag.Bool("burst", false, "Enable burst packet loss")
		corruption = flag.Float64("corruption", 0, "Packet corruption rate")
	)
	flag.Parse()

	fmt.Println("QUIC Network Simulation")
	fmt.Println("==========================")

	// Create network simulation config
	config := internal.NetworkSimulationConfig{
		Latency:     *latency,
		Jitter:      *jitter,
		PacketLoss:  *loss,
		Bandwidth:   *bandwidth,
		Duration:    *duration,
		Duplication: *duplication,
		Reordering:  *reordering,
		BurstLoss:   *burst,
		Corruption:  *corruption,
	}

	// Apply preset if specified
	if *preset != "" {
		simulator := internal.NewNetworkSimulator(config)
		if err := simulator.ApplyPreset(*preset); err != nil {
			log.Fatalf("Failed to apply preset '%s': %v", *preset, err)
		}
		config = simulator.GetConfig()
		fmt.Printf("Applied preset: %s\n", *preset)
	}

	// Create simulator
	simulator := internal.NewNetworkSimulator(config)

	// Start simulation
	if err := simulator.Start(); err != nil {
		log.Fatalf("Failed to start network simulation: %v", err)
	}

	fmt.Printf("🚀 Network simulation started:\n")
	fmt.Printf("  Latency: %v\n", config.Latency)
	fmt.Printf("  Jitter: %v\n", config.Jitter)
	fmt.Printf("  Packet Loss: %.2f%%\n", config.PacketLoss*100)
	fmt.Printf("  Bandwidth: %d bps\n", config.Bandwidth)
	fmt.Printf("  Duration: %v\n", config.Duration)

	// Wait for simulation to complete or user interrupt
	if config.Duration > 0 {
		time.Sleep(config.Duration)
	} else {
		fmt.Println("Press Ctrl+C to stop simulation...")
		select {}
	}

	// Stop simulation
	if err := simulator.Stop(); err != nil {
		log.Printf("Failed to stop simulation: %v", err)
	}

	fmt.Println("✅ Network simulation completed")
}
