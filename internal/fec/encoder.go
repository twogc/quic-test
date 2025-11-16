package fec

import (
	"fmt"
	"sync"
)

// FECEncoder реализует Forward Error Correction используя XOR-based схему
// Для группы из N пакетов создает M redundancy пакетов (M = N * redundancy)
type FECEncoder struct {
	redundancy float64
	groupSize  int           // Размер группы пакетов для FEC
	mu         sync.RWMutex
	packets    [][]byte      // Буфер пакетов текущей группы
	packetIDs  []uint64      // ID пакетов в группе
	groupID    uint64        // ID текущей группы
	metrics    *FECMetrics
}

// FECMetrics метрики FEC
type FECMetrics struct {
	PacketsEncoded    int64   `json:"packets_encoded"`
	RedundancyPackets int64   `json:"redundancy_packets"`
	RedundancyBytes   int64   `json:"redundancy_bytes"`
	GroupsProcessed   int64   `json:"groups_processed"`
}

// NewFECEncoder создает новый FEC encoder
func NewFECEncoder(redundancy float64) *FECEncoder {
	if redundancy <= 0 || redundancy > 1 {
		redundancy = 0.10 // Default 10%
	}
	
	// Группа из 10 пакетов - хороший баланс между latency и эффективностью
	groupSize := 10
	
	return &FECEncoder{
		redundancy: redundancy,
		groupSize:  groupSize,
		packets:    make([][]byte, 0, groupSize),
		packetIDs:  make([]uint64, 0, groupSize),
		metrics:    &FECMetrics{},
	}
}

// AddPacket добавляет пакет в текущую группу
// Возвращает true если группа заполнена и нужно создать redundancy пакеты
func (e *FECEncoder) AddPacket(packet []byte, packetID uint64) (bool, []byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	// Копируем пакет
	packetCopy := make([]byte, len(packet))
	copy(packetCopy, packet)
	
	e.packets = append(e.packets, packetCopy)
	e.packetIDs = append(e.packetIDs, packetID)
	
	// Если группа заполнена, создаем redundancy пакеты
	// Вероятность создания repair packet зависит от redundancy rate
	if len(e.packets) >= e.groupSize {
		// Используем redundancy rate как вероятность создания repair packet
		// Для rate=0.10: всегда создаем (1 repair на 10 packets)
		// Для rate=0.05: создаем с вероятностью 0.5 (0.5 repair на 10 packets в среднем)
		// Для rate=0.20: всегда создаем (можно расширить до 2 repair packets)
		
		shouldCreateRepair := true
		if e.redundancy < 0.10 {
			// Для rate < 10% используем вероятностный подход
			// Упрощенная реализация: создаем repair packet с вероятностью = rate/0.10
			// Например, rate=0.05 → вероятность 0.5
			// В реальной реализации можно использовать random, но для детерминированности
			// создаем каждый N-й раз, где N = 0.10 / rate
			// Для rate=0.05: создаем каждую 2-ю группу
			createInterval := int(0.10 / e.redundancy)
			if createInterval > 1 {
				shouldCreateRepair = (e.groupID % uint64(createInterval)) == 0
			}
		}
		
		var redundancyPacket []byte
		var err error
		
		if shouldCreateRepair {
			redundancyPacket, err = e.generateRedundancy()
			if err != nil {
				return false, nil, err
			}
			
			e.metrics.RedundancyPackets++
			e.metrics.RedundancyBytes += int64(len(redundancyPacket))
		}
		
		// Сбрасываем группу
		e.packets = e.packets[:0]
		e.packetIDs = e.packetIDs[:0]
		e.groupID++
		
		e.metrics.GroupsProcessed++
		e.metrics.PacketsEncoded += int64(e.groupSize)
		
		if shouldCreateRepair {
			return true, redundancyPacket, nil
		}
		
		return false, nil, nil
	}
	
	return false, nil, nil
}

// generateRedundancy создает redundancy пакет используя XOR всех пакетов в группе
func (e *FECEncoder) generateRedundancy() ([]byte, error) {
	if len(e.packets) == 0 {
		return nil, fmt.Errorf("no packets in group")
	}
	
	// Находим максимальный размер пакета
	maxSize := 0
	for _, p := range e.packets {
		if len(p) > maxSize {
			maxSize = len(p)
		}
	}
	
	if maxSize == 0 {
		return nil, fmt.Errorf("empty packets")
	}
	
	// Создаем redundancy пакет как XOR всех пакетов
	redundancy := make([]byte, maxSize)
	
	for i := 0; i < maxSize; i++ {
		var xor byte
		count := 0
		for _, p := range e.packets {
			if i < len(p) {
				xor ^= p[i]
				count++
			}
		}
		redundancy[i] = xor
	}
	
	// Добавляем заголовок FEC: [groupID(8)][packetCount(2)][FEC_MARKER(1)]
	header := make([]byte, 11)
	header[0] = 0xFE // FEC marker
	header[1] = 0xC0 // FEC marker continuation
	header[2] = byte(e.groupID)
	header[3] = byte(e.groupID >> 8)
	header[4] = byte(e.groupID >> 16)
	header[5] = byte(e.groupID >> 24)
	header[6] = byte(e.groupID >> 32)
	header[7] = byte(e.groupID >> 40)
	header[8] = byte(e.groupID >> 48)
	header[9] = byte(e.groupID >> 56)
	header[10] = byte(len(e.packets))
	
	// Объединяем заголовок и redundancy данные
	result := append(header, redundancy...)
	
	return result, nil
}

// GetMetrics возвращает метрики FEC
func (e *FECEncoder) GetMetrics() *FECMetrics {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	// Возвращаем копию метрик
	metrics := *e.metrics
	return &metrics
}

// ResetMetrics сбрасывает метрики
func (e *FECEncoder) ResetMetrics() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.metrics = &FECMetrics{}
}

// Flush принудительно создает redundancy пакет для оставшихся пакетов
func (e *FECEncoder) Flush() ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if len(e.packets) == 0 {
		return nil, nil
	}
	
	redundancy, err := e.generateRedundancy()
	if err != nil {
		return nil, err
	}
	
	// Сбрасываем группу
	e.packets = e.packets[:0]
	e.packetIDs = e.packetIDs[:0]
	e.groupID++
	
	e.metrics.GroupsProcessed++
	e.metrics.PacketsEncoded += int64(len(e.packets))
	e.metrics.RedundancyPackets++
	if redundancy != nil {
		e.metrics.RedundancyBytes += int64(len(redundancy))
	}
	
	return redundancy, nil
}

