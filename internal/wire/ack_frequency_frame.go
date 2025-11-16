package wire

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// AckFrequencyFrame представляет кадр ACK_FREQUENCY (0xaf)
// Согласно draft-ietf-quic-ack-frequency-11
type AckFrequencyFrame struct {
	SequenceNumber         uint64
	AckElicitingThreshold  uint64
	RequestedMaxAckDelayMs uint64
	ReorderingThreshold    uint64
}

// ImmediateAckFrame представляет кадр IMMEDIATE_ACK (0x1f)
type ImmediateAckFrame struct{}

const (
	FrameTypeAckFrequency = 0xaf
	FrameTypeImmediateAck = 0x1f
)

// Write сериализует ACK_FREQUENCY кадр
func (f *AckFrequencyFrame) Write(b *bytes.Buffer) error {
	// Записываем тип кадра
	b.WriteByte(FrameTypeAckFrequency)

	// Записываем поля в порядке согласно спецификации
	if err := writeVarInt(b, f.SequenceNumber); err != nil {
		return err
	}
	if err := writeVarInt(b, f.AckElicitingThreshold); err != nil {
		return err
	}
	if err := writeVarInt(b, f.RequestedMaxAckDelayMs); err != nil {
		return err
	}
	if err := writeVarInt(b, f.ReorderingThreshold); err != nil {
		return err
	}

	return nil
}

// Length возвращает длину кадра в байтах
func (f *AckFrequencyFrame) Length() int {
	return 1 + // тип кадра
		varIntLen(f.SequenceNumber) +
		varIntLen(f.AckElicitingThreshold) +
		varIntLen(f.RequestedMaxAckDelayMs) +
		varIntLen(f.ReorderingThreshold)
}

// Write сериализует IMMEDIATE_ACK кадр
func (f *ImmediateAckFrame) Write(b *bytes.Buffer) error {
	b.WriteByte(FrameTypeImmediateAck)
	return nil
}

// Length возвращает длину кадра в байтах
func (f *ImmediateAckFrame) Length() int {
	return 1 // только тип кадра
}

// ParseAckFrequencyFrame парсит ACK_FREQUENCY кадр
func ParseAckFrequencyFrame(r *bytes.Reader) (*AckFrequencyFrame, error) {
	frame := &AckFrequencyFrame{}

	var err error
	frame.SequenceNumber, err = readVarInt(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read sequence number: %v", err)
	}

	frame.AckElicitingThreshold, err = readVarInt(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read ack eliciting threshold: %v", err)
	}

	frame.RequestedMaxAckDelayMs, err = readVarInt(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read requested max ack delay: %v", err)
	}

	frame.ReorderingThreshold, err = readVarInt(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read reordering threshold: %v", err)
	}

	return frame, nil
}

// ParseImmediateAckFrame парсит IMMEDIATE_ACK кадр
func ParseImmediateAckFrame(r *bytes.Reader) (*ImmediateAckFrame, error) {
	return &ImmediateAckFrame{}, nil
}

// writeVarInt записывает переменное целое число
func writeVarInt(b *bytes.Buffer, val uint64) error {
	if val < 0x40 {
		return b.WriteByte(byte(val))
	} else if val < 0x4000 {
		return binary.Write(b, binary.BigEndian, uint16(val|0x4000))
	} else if val < 0x40000000 {
		return binary.Write(b, binary.BigEndian, uint32(val|0x80000000))
	} else {
		return binary.Write(b, binary.BigEndian, uint64(val|0xc000000000000000))
	}
}

// readVarInt читает переменное целое число
func readVarInt(r *bytes.Reader) (uint64, error) {
	firstByte, err := r.ReadByte()
	if err != nil {
		return 0, err
	}

	if firstByte < 0x40 {
		return uint64(firstByte), nil
	} else if firstByte < 0x80 {
		// 2-byte encoding
		secondByte, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		return uint64(firstByte&0x3f)<<8 | uint64(secondByte), nil
	} else if firstByte < 0xc0 {
		// 4-byte encoding
		var val uint32
		if err := binary.Read(r, binary.BigEndian, &val); err != nil {
			return 0, err
		}
		return uint64(val & 0x3fffffff), nil
	} else {
		// 8-byte encoding
		var val uint64
		if err := binary.Read(r, binary.BigEndian, &val); err != nil {
			return 0, err
		}
		return val & 0x3fffffffffffffff, nil
	}
}

// varIntLen возвращает длину переменного целого числа в байтах
func varIntLen(val uint64) int {
	if val < 0x40 {
		return 1
	} else if val < 0x4000 {
		return 2
	} else if val < 0x40000000 {
		return 4
	} else {
		return 8
	}
}

