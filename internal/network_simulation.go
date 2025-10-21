package internal

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// NetworkSimulationConfig конфигурация для эмуляции сетевых условий
type NetworkSimulationConfig struct {
	// Latency simulation
	Latency     time.Duration `json:"latency"`
	Jitter      time.Duration `json:"jitter"`
	
	// Packet loss simulation
	PacketLoss  float64 `json:"packet_loss"` // 0.0 to 1.0
	
	// Bandwidth simulation
	Bandwidth   int64   `json:"bandwidth"`   // bytes per second
	
	// Network conditions
	Duplication float64 `json:"duplication"`  // packet duplication rate
	Reordering  bool    `json:"reordering"`  // enable packet reordering
	
	// Advanced conditions
	BurstLoss   bool    `json:"burst_loss"`  // burst packet loss
	Corruption  float64 `json:"corruption"` // packet corruption rate
	
	// Time-based conditions
	Duration    time.Duration `json:"duration"`
	StartTime   time.Time     `json:"start_time"`
}

// NetworkSimulator симулятор сетевых условий
type NetworkSimulator struct {
	config     NetworkSimulationConfig
	isActive   bool
	startTime  time.Time
	stopTime   time.Time
}

// NewNetworkSimulator создает новый симулятор сетевых условий
func NewNetworkSimulator(config NetworkSimulationConfig) *NetworkSimulator {
	return &NetworkSimulator{
		config:   config,
		isActive: false,
	}
}

// Start начинает эмуляцию сетевых условий
func (ns *NetworkSimulator) Start() error {
	if ns.isActive {
		return fmt.Errorf("network simulation is already active")
	}

	log.Printf("🌐 Starting network simulation: Latency=%v, Loss=%.2f%%, Bandwidth=%d bps", 
		ns.config.Latency, ns.config.PacketLoss*100, ns.config.Bandwidth)

	// Apply network conditions using Linux tc (traffic control)
	if err := ns.applyNetworkConditions(); err != nil {
		return fmt.Errorf("failed to apply network conditions: %v", err)
	}

	ns.isActive = true
	ns.startTime = time.Now()
	
	// Schedule automatic stop if duration is specified
	if ns.config.Duration > 0 {
		go func() {
			time.Sleep(ns.config.Duration)
			ns.Stop()
		}()
	}

	return nil
}

// Stop останавливает эмуляцию сетевых условий
func (ns *NetworkSimulator) Stop() error {
	if !ns.isActive {
		return fmt.Errorf("network simulation is not active")
	}

	log.Printf("🛑 Stopping network simulation")

	// Remove network conditions
	if err := ns.removeNetworkConditions(); err != nil {
		return fmt.Errorf("failed to remove network conditions: %v", err)
	}

	ns.isActive = false
	ns.stopTime = time.Now()
	
	log.Printf("✅ Network simulation completed. Duration: %v", ns.stopTime.Sub(ns.startTime))
	return nil
}

// IsActive возвращает статус симуляции
func (ns *NetworkSimulator) IsActive() bool {
	return ns.isActive
}

// GetConfig возвращает текущую конфигурацию
func (ns *NetworkSimulator) GetConfig() NetworkSimulationConfig {
	return ns.config
}

// UpdateConfig обновляет конфигурацию симуляции
func (ns *NetworkSimulator) UpdateConfig(newConfig NetworkSimulationConfig) error {
	if ns.isActive {
		// Stop current simulation
		if err := ns.Stop(); err != nil {
			return fmt.Errorf("failed to stop current simulation: %v", err)
		}
	}

	ns.config = newConfig
	return nil
}

// applyNetworkConditions применяет сетевые условия через Linux tc
func (ns *NetworkSimulator) applyNetworkConditions() error {
	// Get default network interface
	iface, err := ns.getDefaultInterface()
	if err != nil {
		return fmt.Errorf("failed to get default interface: %v", err)
	}

	// Build tc commands
	commands := ns.buildTcCommands(iface)
	
	// Execute commands
	for _, cmd := range commands {
		if err := ns.executeCommand(cmd); err != nil {
			log.Printf("Warning: Failed to execute command '%s': %v", cmd, err)
		}
	}

	return nil
}

// removeNetworkConditions удаляет сетевые условия
func (ns *NetworkSimulator) removeNetworkConditions() error {
	iface, err := ns.getDefaultInterface()
	if err != nil {
		return fmt.Errorf("failed to get default interface: %v", err)
	}

	// Remove all qdiscs
	cmd := fmt.Sprintf("tc qdisc del dev %s root", iface)
	return ns.executeCommand(cmd)
}

// getDefaultInterface получает интерфейс по умолчанию
func (ns *NetworkSimulator) getDefaultInterface() (string, error) {
	cmd := exec.Command("ip", "route", "show", "default")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("no default route found")
	}

	parts := strings.Fields(lines[0])
	if len(parts) < 5 {
		return "", fmt.Errorf("invalid route format")
	}

	return parts[4], nil
}

// buildTcCommands строит команды tc для применения сетевых условий
func (ns *NetworkSimulator) buildTcCommands(iface string) []string {
	var commands []string

	// Base qdisc
	baseCmd := fmt.Sprintf("tc qdisc add dev %s root handle 1: htb default 30", iface)
	commands = append(commands, baseCmd)

	// Latency and jitter
	if ns.config.Latency > 0 {
		latencyMs := int(ns.config.Latency.Milliseconds())
		jitterMs := int(ns.config.Jitter.Milliseconds())
		
		netemCmd := fmt.Sprintf("tc qdisc add dev %s parent 1:30 handle 30: netem delay %dms", 
			iface, latencyMs)
		
		if jitterMs > 0 {
			netemCmd += fmt.Sprintf(" %dms", jitterMs)
		}
		
		commands = append(commands, netemCmd)
	}

	// Packet loss
	if ns.config.PacketLoss > 0 {
		lossPercent := ns.config.PacketLoss * 100
		lossCmd := fmt.Sprintf("tc qdisc change dev %s parent 1:30 handle 30: netem loss %.2f%%", 
			iface, lossPercent)
		commands = append(commands, lossCmd)
	}

	// Bandwidth limiting
	if ns.config.Bandwidth > 0 {
		bandwidthKbps := ns.config.Bandwidth / 1000
		bwCmd := fmt.Sprintf("tc class add dev %s parent 1: classid 1:1 htb rate %dkbit", 
			iface, bandwidthKbps)
		commands = append(commands, bwCmd)
		
		filterCmd := fmt.Sprintf("tc filter add dev %s parent 1: protocol ip prio 1 u32 match ip dst 0.0.0.0/0 flowid 1:1", 
			iface)
		commands = append(commands, filterCmd)
	}

	// Packet duplication
	if ns.config.Duplication > 0 {
		dupPercent := ns.config.Duplication * 100
		dupCmd := fmt.Sprintf("tc qdisc change dev %s parent 1:30 handle 30: netem duplicate %.2f%%", 
			iface, dupPercent)
		commands = append(commands, dupCmd)
	}

	// Packet reordering
	if ns.config.Reordering {
		reorderCmd := fmt.Sprintf("tc qdisc change dev %s parent 1:30 handle 30: netem reorder 25%% 50%%", 
			iface)
		commands = append(commands, reorderCmd)
	}

	// Burst loss
	if ns.config.BurstLoss {
		burstCmd := fmt.Sprintf("tc qdisc change dev %s parent 1:30 handle 30: netem loss random 10%%", 
			iface)
		commands = append(commands, burstCmd)
	}

	// Packet corruption
	if ns.config.Corruption > 0 {
		corruptPercent := ns.config.Corruption * 100
		corruptCmd := fmt.Sprintf("tc qdisc change dev %s parent 1:30 handle 30: netem corrupt %.2f%%", 
			iface, corruptPercent)
		commands = append(commands, corruptCmd)
	}

	return commands
}

// executeCommand выполняет команду
func (ns *NetworkSimulator) executeCommand(cmd string) error {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	command := exec.Command(parts[0], parts[1:]...)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %s, output: %s", err, string(output))
	}

	return nil
}

// GetSimulationStatus возвращает статус симуляции
func (ns *NetworkSimulator) GetSimulationStatus() map[string]interface{} {
	status := map[string]interface{}{
		"active":     ns.isActive,
		"start_time": ns.startTime,
		"stop_time":  ns.stopTime,
		"duration":   ns.stopTime.Sub(ns.startTime),
		"config":     ns.config,
	}

	if ns.isActive {
		status["elapsed"] = time.Since(ns.startTime)
	}

	return status
}

// PresetNetworkConditions предустановленные сетевые условия
var PresetNetworkConditions = map[string]NetworkSimulationConfig{
	"excellent": {
		Latency:    5 * time.Millisecond,
		Jitter:     1 * time.Millisecond,
		PacketLoss: 0.001, // 0.1%
		Bandwidth:  1000 * 1024 * 1024, // 1 Gbps
	},
	"good": {
		Latency:    20 * time.Millisecond,
		Jitter:     5 * time.Millisecond,
		PacketLoss: 0.01, // 1%
		Bandwidth:  100 * 1024 * 1024, // 100 Mbps
	},
	"poor": {
		Latency:    100 * time.Millisecond,
		Jitter:     20 * time.Millisecond,
		PacketLoss: 0.05, // 5%
		Bandwidth:  10 * 1024 * 1024, // 10 Mbps
	},
	"mobile": {
		Latency:    200 * time.Millisecond,
		Jitter:     50 * time.Millisecond,
		PacketLoss: 0.1, // 10%
		Bandwidth:  5 * 1024 * 1024, // 5 Mbps
		Reordering: true,
		BurstLoss:  true,
	},
	"satellite": {
		Latency:    500 * time.Millisecond,
		Jitter:     100 * time.Millisecond,
		PacketLoss: 0.02, // 2%
		Bandwidth:  2 * 1024 * 1024, // 2 Mbps
		Duplication: 0.01, // 1% duplication
	},
	"adversarial": {
		Latency:    1000 * time.Millisecond,
		Jitter:     200 * time.Millisecond,
		PacketLoss: 0.2, // 20%
		Bandwidth:  1 * 1024 * 1024, // 1 Mbps
		Reordering: true,
		BurstLoss:  true,
		Corruption: 0.05, // 5% corruption
	},
}

// ApplyPreset применяет предустановленные сетевые условия
func (ns *NetworkSimulator) ApplyPreset(presetName string) error {
	preset, exists := PresetNetworkConditions[presetName]
	if !exists {
		return fmt.Errorf("preset '%s' not found", presetName)
	}

	return ns.UpdateConfig(preset)
}

// ListPresets возвращает список доступных предустановок
func (ns *NetworkSimulator) ListPresets() []string {
	var presets []string
	for name := range PresetNetworkConditions {
		presets = append(presets, name)
	}
	return presets
}
