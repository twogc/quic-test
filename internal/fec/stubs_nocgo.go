// +build !cgo

package fec

// FECEncoderCXX stub when CGO is not available
type FECEncoderCXX struct {
	initialized bool
}

// FECBatchGroup stub
type FECBatchGroup struct {
	Packets [][]byte // Packet data
	Sizes   []uint32 // Packet sizes (optional)
}

// RepairPacket stub
type RepairPacket []byte

// NewFECEncoderCXX returns nil when CGO is not available
func NewFECEncoderCXX(redundancy float64, maxGroups int) *FECEncoderCXX {
	return nil
}

// EncodeBatch stub - not implemented without CGO
func (e *FECEncoderCXX) EncodeBatch(groups []FECBatchGroup, packetSize int) ([]RepairPacket, error) {
	return nil, nil
}

// Close stub - no-op without CGO
func (e *FECEncoderCXX) Close() error {
	return nil
}

