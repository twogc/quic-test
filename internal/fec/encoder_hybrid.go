package fec

import (
	"fmt"
	"sync"
)

// HybridFECEncoder uses C++ SIMD for XOR encoding when available, falls back to Go
type HybridFECEncoder struct {
	// Hybrid encoder state
	redundancy    float64
	groupSize     int
	useCXX        bool      // Whether to use C++ implementation
	cxxEncoder    *FECEncoderCXX
	goEncoder     *FECEncoder

	// Buffering state
	packets       [][]byte
	packetIDs     []uint64
	groupID       uint64

	mu            sync.Mutex
	metrics       *FECMetrics
}

// NewHybridFECEncoder creates an encoder that uses C++ if available, else Go
func NewHybridFECEncoder(redundancy float64) *HybridFECEncoder {
	if redundancy <= 0 || redundancy > 1 {
		redundancy = 0.10
	}

	groupSize := 10

	enc := &HybridFECEncoder{
		redundancy: redundancy,
		groupSize:  groupSize,
		packets:    make([][]byte, 0, groupSize),
		packetIDs:  make([]uint64, 0, groupSize),
		metrics:    &FECMetrics{},
		useCXX:     false,
	}

	// Try to create C++ encoder
	cxxEnc := NewFECEncoderCXX(redundancy, 1024)
	if cxxEnc != nil && cxxEnc.initialized {
		enc.cxxEncoder = cxxEnc
		enc.useCXX = true
	} else {
		// Fallback to Go encoder
		enc.goEncoder = NewFECEncoder(redundancy)
		enc.useCXX = false
	}

	return enc
}

// AddPacket adds a packet to the encoder
// Returns (needsRedundancy, redundancyPacket, error)
func (e *HybridFECEncoder) AddPacket(packet []byte, packetID uint64) (bool, []byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Copy packet data
	packetCopy := make([]byte, len(packet))
	copy(packetCopy, packet)

	e.packets = append(e.packets, packetCopy)
	e.packetIDs = append(e.packetIDs, packetID)

	// Check if group is full
	if len(e.packets) >= e.groupSize {
		return e.generateRedundancy()
	}

	return false, nil, nil
}

// generateRedundancy creates repair packet using either C++ or Go
func (e *HybridFECEncoder) generateRedundancy() (bool, []byte, error) {
	if e.useCXX && e.cxxEncoder != nil {
		return e.generateRedundancyCXX()
	}
	return e.generateRedundancyGo()
}

// generateRedundancyCXX uses C++ SIMD encoder
func (e *HybridFECEncoder) generateRedundancyCXX() (bool, []byte, error) {
	if len(e.packets) == 0 {
		return false, nil, fmt.Errorf("no packets in group")
	}

	// Get max packet size
	maxSize := 0
	for _, p := range e.packets {
		if len(p) > maxSize {
			maxSize = len(p)
		}
	}

	if maxSize == 0 {
		return false, nil, fmt.Errorf("empty packets")
	}

	// Create FECBatchGroup for C++ encoder
	group := FECBatchGroup{
		Packets: e.packets,
		Sizes:   make([]uint32, len(e.packets)),
	}

	for i, p := range e.packets {
		group.Sizes[i] = uint32(len(p))
	}

	// Encode using C++
	repairs, err := e.cxxEncoder.EncodeBatch([]FECBatchGroup{group}, maxSize)
	if err != nil {
		return false, nil, fmt.Errorf("C++ encoding failed: %w", err)
	}

	if len(repairs) == 0 {
		return false, nil, fmt.Errorf("no repair packet generated")
	}

	repair := repairs[0]

	// Create FEC packet with header
	result := e.createFECPacket(repair, len(e.packets), maxSize)

	// Reset group
	e.packets = e.packets[:0]
	e.packetIDs = e.packetIDs[:0]
	e.groupID++

	e.metrics.GroupsProcessed++
	e.metrics.PacketsEncoded += int64(e.groupSize)
	e.metrics.RedundancyPackets++
	e.metrics.RedundancyBytes += int64(len(result))

	return true, result, nil
}

// generateRedundancyGo uses pure Go XOR encoder
func (e *HybridFECEncoder) generateRedundancyGo() (bool, []byte, error) {
	if len(e.packets) == 0 {
		return false, nil, fmt.Errorf("no packets in group")
	}

	// Create temporary Go encoder with same redundancy
	tempEncoder := &FECEncoder{
		redundancy: e.redundancy,
		packets:    e.packets,
		groupID:    e.groupID,
	}

	// Get repair packet
	repair, err := tempEncoder.generateRedundancy()
	if err != nil {
		return false, nil, err
	}

	// Reset group
	e.packets = e.packets[:0]
	e.packetIDs = e.packetIDs[:0]
	e.groupID++

	e.metrics.GroupsProcessed++
	e.metrics.PacketsEncoded += int64(e.groupSize)
	e.metrics.RedundancyPackets++
	e.metrics.RedundancyBytes += int64(len(repair))

	return true, repair, nil
}

// createFECPacket creates FEC packet with header
func (e *HybridFECEncoder) createFECPacket(repairData []byte, packetCount, maxSize int) []byte {
	// Header: [groupID(8)][packetCount(2)][FEC_MARKER(1)]
	header := make([]byte, 11)
	header[0] = 0xFE
	header[1] = 0xC0
	header[2] = byte(e.groupID)
	header[3] = byte(e.groupID >> 8)
	header[4] = byte(e.groupID >> 16)
	header[5] = byte(e.groupID >> 24)
	header[6] = byte(e.groupID >> 32)
	header[7] = byte(e.groupID >> 40)
	header[8] = byte(e.groupID >> 48)
	header[9] = byte(e.groupID >> 56)
	header[10] = byte(packetCount)

	result := append(header, repairData...)
	return result
}

// GetMetrics returns current metrics
func (e *HybridFECEncoder) GetMetrics() *FECMetrics {
	e.mu.Lock()
	defer e.mu.Unlock()

	metrics := *e.metrics
	return &metrics
}

// ResetMetrics resets metrics counters
func (e *HybridFECEncoder) ResetMetrics() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.metrics = &FECMetrics{}
}

// Flush flushes remaining packets
func (e *HybridFECEncoder) Flush() ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.packets) == 0 {
		return nil, nil
	}

	_, repair, err := e.generateRedundancy()
	return repair, err
}

// UseCXX returns true if C++ encoder is being used
func (e *HybridFECEncoder) UseCXX() bool {
	return e.useCXX
}

// Close cleans up resources
func (e *HybridFECEncoder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.cxxEncoder != nil {
		return e.cxxEncoder.Close()
	}
	return nil
}

