// +build cgo

package fec

/*
#cgo CFLAGS: -I. -mavx2 -mfma
#cgo LDFLAGS: -L. -Wl,-rpath,.

// Platform-specific LDFLAGS
#cgo linux LDFLAGS: -lfec_avx2
#cgo darwin LDFLAGS: -lfec_avx2

#include "fec_xor_simd.h"
#include <stdint.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// FECEncoderCXX is a high-performance FEC encoder using C++ SIMD
// It uses flat-slab memory layout to minimize CGO overhead
type FECEncoderCXX struct {
	ctx           *C.FECEncoderCtx
	slab          []byte        // C-managed flat buffer for packet data
	slabSize      int           // Current slab allocation size
	offsets       []uint32      // Packet offsets within slab
	repairBuffer  []byte        // C-managed repair packet buffer
	repairSize    int           // Current repair buffer size
	maxGroups     int           // Maximum groups to track
	mu            sync.Mutex    // Protect state
	initialized   bool
}

// RepairPacket is a repair packet output
type RepairPacket []byte

// NewFECEncoderCXX creates a new C++ optimized FEC encoder
// redundancy: ratio of repair packets to data packets (0.0-1.0)
// maxGroups: maximum number of groups to process in a batch
func NewFECEncoderCXX(redundancy float64, maxGroups int) *FECEncoderCXX {
	if redundancy <= 0 || redundancy > 1 {
		redundancy = 0.10
	}
	if maxGroups <= 0 {
		maxGroups = 1024
	}

	enc := &FECEncoderCXX{
		maxGroups: maxGroups,
	}

	// Create C++ context
	enc.ctx = C.fec_encoder_new(C.double(redundancy), C.uint32_t(maxGroups))
	if enc.ctx == nil {
		return nil
	}

	// Pre-allocate slab for packet data
	// Assume: maxGroups * 10 packets/group * 1200 bytes/packet
	enc.slabSize = maxGroups * 10 * 1200
	slabPtr := C.fec_alloc_slab(C.size_t(enc.slabSize))
	if slabPtr == nil {
		C.fec_encoder_free(enc.ctx)
		return nil
	}

	// Convert C pointer to Go slice
	enc.slab = unsafe.Slice((*byte)(slabPtr), enc.slabSize)
	enc.offsets = make([]uint32, 0, maxGroups*10)

	// Pre-allocate repair buffer
	enc.repairSize = maxGroups * 1200
	repairPtr := C.fec_alloc_repair_buffer(C.size_t(enc.repairSize))
	if repairPtr == nil {
		C.fec_free_slab(unsafe.Pointer(&enc.slab[0]))
		C.fec_encoder_free(enc.ctx)
		return nil
	}
	enc.repairBuffer = unsafe.Slice((*byte)(repairPtr), enc.repairSize)

	enc.initialized = true

	// Finalize function to clean up C resources
	runtime.SetFinalizer(enc, (*FECEncoderCXX).Close)

	return enc
}

// EncodeBatch encodes a batch of FEC groups in one CGO call
// groups: list of FEC batch groups, each with packets of same size
// Returns: repair packets (one per group), or error
func (e *FECEncoderCXX) EncodeBatch(groups []FECBatchGroup, packetSize int) ([]RepairPacket, error) {
	if !e.initialized {
		return nil, fmt.Errorf("encoder not initialized")
	}

	if len(groups) == 0 {
		return nil, nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Build flat slab: copy all packets into one buffer
	var offset uint32 = 0
	e.offsets = e.offsets[:0]

	for _, group := range groups {
		for _, pkt := range group.Packets {
			if offset+uint32(len(pkt)) > uint32(len(e.slab)) {
				// Resize slab if needed
				newSize := int(offset) + len(pkt) + (len(groups)-1)*packetSize*10
				if err := e.resizeSlab(newSize); err != nil {
					return nil, fmt.Errorf("failed to resize slab: %w", err)
				}
			}

			// Copy packet data
			copy(e.slab[offset:], pkt)
			e.offsets = append(e.offsets, offset)
			offset += uint32(len(pkt))
		}
	}

	// Ensure repair buffer is large enough
	repairNeeded := len(groups) * packetSize
	if repairNeeded > len(e.repairBuffer) {
		if err := e.resizeRepairBuffer(repairNeeded); err != nil {
			return nil, fmt.Errorf("failed to resize repair buffer: %w", err)
		}
	}

	// Call C++ batch encoder
	// Lock slab and offset pointers during C call
	ret := C.fec_encode_batch(
		e.ctx,
		(*C.uint8_t)(unsafe.Pointer(&e.slab[0])),
		(*C.uint32_t)(unsafe.Pointer(&e.offsets[0])),
		C.uint32_t(len(groups)),
		C.uint32_t(packetSize),
		(*C.uint8_t)(unsafe.Pointer(&e.repairBuffer[0])),
	)

	if ret != 0 {
		return nil, fmt.Errorf("C++ encoding failed with code %d", ret)
	}

	// Extract repair packets from buffer
	repairs := make([]RepairPacket, len(groups))
	for i := 0; i < len(groups); i++ {
		start := i * packetSize
		end := start + packetSize
		if end > len(e.repairBuffer) {
			end = len(e.repairBuffer)
		}

		// Copy repair packet (can't return direct pointer due to GC)
		repairs[i] = make([]byte, packetSize)
		copy(repairs[i], e.repairBuffer[start:end])
	}

	// Keep slices alive during C call (KeepAlive)
	runtime.KeepAlive(e.slab)
	runtime.KeepAlive(e.offsets)
	runtime.KeepAlive(e.repairBuffer)

	return repairs, nil
}

// resizeSlab expands the slab buffer if needed
func (e *FECEncoderCXX) resizeSlab(newSize int) error {
	if newSize <= len(e.slab) {
		return nil
	}

	// Allocate new slab
	newSlabPtr := C.fec_alloc_slab(C.size_t(newSize))
	if newSlabPtr == nil {
		return fmt.Errorf("failed to allocate slab of size %d", newSize)
	}

	newSlab := unsafe.Slice((*byte)(newSlabPtr), newSize)

	// Copy old data
	copy(newSlab, e.slab)

	// Free old slab
	C.fec_free_slab(unsafe.Pointer(&e.slab[0]))

	e.slab = newSlab
	e.slabSize = newSize

	return nil
}

// resizeRepairBuffer expands the repair buffer if needed
func (e *FECEncoderCXX) resizeRepairBuffer(newSize int) error {
	if newSize <= len(e.repairBuffer) {
		return nil
	}

	newRepairPtr := C.fec_alloc_repair_buffer(C.size_t(newSize))
	if newRepairPtr == nil {
		return fmt.Errorf("failed to allocate repair buffer of size %d", newSize)
	}

	newRepair := unsafe.Slice((*byte)(newRepairPtr), newSize)
	copy(newRepair, e.repairBuffer)

	C.fec_free_repair_buffer(unsafe.Pointer(&e.repairBuffer[0]))

	e.repairBuffer = newRepair
	e.repairSize = newSize

	return nil
}

// Close releases all C++ resources
func (e *FECEncoderCXX) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.initialized {
		return nil
	}

	if e.ctx != nil {
		C.fec_encoder_free(e.ctx)
		e.ctx = nil
	}

	if len(e.slab) > 0 {
		C.fec_free_slab(unsafe.Pointer(&e.slab[0]))
		e.slab = nil
	}

	if len(e.repairBuffer) > 0 {
		C.fec_free_repair_buffer(unsafe.Pointer(&e.repairBuffer[0]))
		e.repairBuffer = nil
	}

	e.initialized = false
	return nil
}

// FECBatchGroup represents a group of packets to encode in batch mode
// This is different from FECGroup in decoder.go which is for decoding
type FECBatchGroup struct {
	Packets [][]byte // Packet data
	Sizes   []uint32 // Packet sizes (optional, for variable-length packets)
}

