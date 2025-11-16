package masque

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// CapsuleTester тестирует HTTP Capsules функциональность (RFC 9297)
type CapsuleTester struct {
	logger *zap.Logger
	config *MASQUEConfig
}

// Capsule представляет HTTP Capsule
type Capsule struct {
	Type   uint64
	Length uint64
	Value  []byte
}

// Capsule types as defined in RFC 9297
const (
	CapsuleTypeDatagram = 0x00
	CapsuleTypeClose    = 0x01
)

// NewCapsuleTester создает новый Capsule тестер
func NewCapsuleTester(logger *zap.Logger, config *MASQUEConfig) *CapsuleTester {
	return &CapsuleTester{
		logger: logger,
		config: config,
	}
}

// TestDatagrams тестирует HTTP Datagrams
func (ct *CapsuleTester) TestDatagrams(ctx context.Context, testData []byte) (int64, int64, error) {
	ct.logger.Info("Testing HTTP Datagrams", zap.Int("data_len", len(testData)))

	sent := int64(0)
	received := int64(0)

	// В реальной реализации здесь было бы тестирование с реальным HTTP/3 stream
	// Для тестирования имитируем отправку и получение datagrams
	
	// Имитируем отправку datagrams
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			return sent, received, ctx.Err()
		default:
			// Имитируем отправку datagram
			sent++
			time.Sleep(10 * time.Millisecond) // Имитируем сетевую задержку
			
			// Имитируем получение datagram (с небольшой вероятностью потери)
			if i%10 != 9 { // 90% успешность
				received++
			}
		}
	}

	ct.logger.Info("HTTP Datagrams test completed",
		zap.Int64("sent", sent),
		zap.Int64("received", received))

	return sent, received, nil
}

// TestCapsules тестирует HTTP Capsules
func (ct *CapsuleTester) TestCapsules(ctx context.Context) (int64, int64, error) {
	ct.logger.Info("Testing HTTP Capsules")

	sent := int64(0)
	received := int64(0)

	// Тестируем отправку различных типов capsules
	capsules := []*Capsule{
		{
			Type:   CapsuleTypeDatagram,
			Length: 5,
			Value:  []byte("Hello"),
		},
		{
			Type:   CapsuleTypeDatagram,
			Length: 8,
			Value:  []byte("MASQUE!"),
		},
		{
			Type:   CapsuleTypeClose,
			Length: 0,
			Value:  nil,
		},
	}

	for _, capsule := range capsules {
		select {
		case <-ctx.Done():
			return sent, received, ctx.Err()
		default:
			// Имитируем отправку capsule
			if err := ct.sendCapsule(capsule); err != nil {
				ct.logger.Error("Failed to send capsule", zap.Error(err))
				continue
			}
			sent++

			// Имитируем получение capsule
			time.Sleep(5 * time.Millisecond)
			received++
		}
	}

	ct.logger.Info("HTTP Capsules test completed",
		zap.Int64("sent", sent),
		zap.Int64("received", received))

	return sent, received, nil
}

// TestThroughput тестирует пропускную способность
func (ct *CapsuleTester) TestThroughput(ctx context.Context, duration time.Duration) (float64, error) {
	ct.logger.Info("Testing throughput", zap.Duration("duration", duration))

	start := time.Now()
	bytesSent := int64(0)
	
	// Создаем тестовые данные
	testData := make([]byte, 1024)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	for time.Since(start) < duration {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			// Создаем datagram capsule
			capsule := &Capsule{
				Type:   CapsuleTypeDatagram,
				Length: uint64(len(testData)),
				Value:  testData,
			}

			// Имитируем отправку
			if err := ct.sendCapsule(capsule); err != nil {
				ct.logger.Error("Failed to send capsule", zap.Error(err))
				continue
			}

			bytesSent += int64(len(testData))
			time.Sleep(1 * time.Millisecond) // Имитируем сетевую задержку
		}
	}

	elapsed := time.Since(start)
	throughput := float64(bytesSent) / elapsed.Seconds() / (1024 * 1024) // MB/s

	ct.logger.Info("Throughput test completed",
		zap.Float64("throughput_mbps", throughput),
		zap.Int64("bytes_sent", bytesSent))

	return throughput, nil
}

// TestLatency тестирует задержку
func (ct *CapsuleTester) TestLatency(ctx context.Context) (time.Duration, error) {
	ct.logger.Info("Testing latency")

	testData := []byte("ping")
	
	// Создаем datagram capsule
	capsule := &Capsule{
		Type:   CapsuleTypeDatagram,
		Length: uint64(len(testData)),
		Value:  testData,
	}

	start := time.Now()
	
	// Отправляем capsule
	if err := ct.sendCapsule(capsule); err != nil {
		return 0, fmt.Errorf("failed to send capsule: %v", err)
	}

	// Имитируем получение ответа
	time.Sleep(5 * time.Millisecond)
	
	latency := time.Since(start)

	ct.logger.Info("Latency test completed",
		zap.Duration("latency", latency))

	return latency, nil
}

// sendCapsule отправляет HTTP Capsule
func (ct *CapsuleTester) sendCapsule(capsule *Capsule) error {
	// В реальной реализации здесь была бы отправка через HTTP/3 stream
	// Для тестирования имитируем отправку
	
	ct.logger.Debug("Sending capsule",
		zap.Uint64("type", capsule.Type),
		zap.Uint64("length", capsule.Length))

	// Имитируем сетевую задержку
	time.Sleep(1 * time.Millisecond)
	
	return nil
}

// parseCapsule парсит HTTP Capsule из данных
func (ct *CapsuleTester) parseCapsule(data []byte) (*Capsule, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty capsule data")
	}

	// Парсим тип capsule
	if len(data) < 1 {
		return nil, fmt.Errorf("invalid capsule: too short")
	}

	// Простой парсинг для тестирования
	capsuleType := uint64(data[0])
	n := 1

	// Парсим длину
	if len(data) < n+1 {
		return nil, fmt.Errorf("invalid capsule: length too short")
	}

	capsuleLength := uint64(data[n])
	m := 1

	// Извлекаем значение
	valueStart := n + m
	if len(data) < valueStart+int(capsuleLength) {
		return nil, fmt.Errorf("invalid capsule: value too short")
	}

	value := make([]byte, capsuleLength)
	if capsuleLength > 0 {
		copy(value, data[valueStart:valueStart+int(capsuleLength)])
	}

	return &Capsule{
		Type:   capsuleType,
		Length: capsuleLength,
		Value:  value,
	}, nil
}

// writeCapsule записывает HTTP Capsule в буфер
func (ct *CapsuleTester) writeCapsule(capsule *Capsule) ([]byte, error) {
	// Простая реализация для тестирования
	totalSize := 1 + 1 + len(capsule.Value) // type + length + value

	// Создаем буфер
	buf := make([]byte, totalSize)
	offset := 0

	// Записываем тип (1 байт)
	buf[offset] = byte(capsule.Type)
	offset++

	// Записываем длину (1 байт)
	buf[offset] = byte(capsule.Length)
	offset++

	// Записываем значение
	if len(capsule.Value) > 0 {
		copy(buf[offset:], capsule.Value)
	}

	return buf, nil
}

// TestCapsuleFallback тестирует fallback с DATAGRAM на Capsules
func (ct *CapsuleTester) TestCapsuleFallback(ctx context.Context) error {
	ct.logger.Info("Testing Capsule fallback mechanism")

	// Имитируем ситуацию, когда DATAGRAM недоступен
	ct.logger.Info("Simulating DATAGRAM unavailability")
	
	// Переключаемся на Capsules
	ct.logger.Info("Falling back to HTTP Capsules")
	
	// Тестируем отправку через Capsules
	testData := []byte("fallback test")
	capsule := &Capsule{
		Type:   CapsuleTypeDatagram,
		Length: uint64(len(testData)),
		Value:  testData,
	}

	if err := ct.sendCapsule(capsule); err != nil {
		return fmt.Errorf("failed to send fallback capsule: %v", err)
	}

	ct.logger.Info("Capsule fallback test completed successfully")
	return nil
}

// Stop останавливает Capsule тестер
func (ct *CapsuleTester) Stop() error {
	ct.logger.Info("Stopping Capsule tester")
	return nil
}
