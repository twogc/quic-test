package fec

import (
	"encoding/binary"
	"sync"
	"time"
)

const (
	maxActiveGroups = 4096
	groupTTL        = 5 * time.Second
	maxSymbolLen    = 1500 // MTU limit
	maxPacketCount  = 255  // Reasonable upper limit
)

// Recovered представляет восстановленный пакет
type Recovered struct {
	PacketID uint64
	Data     []byte
}

// FECDecoder реализует декодирование FEC пакетов для восстановления потерянных данных
// Ограничение: XOR-FEC восстанавливает только 1 потерянный пакет на группу
type FECDecoder struct {
	groups     map[uint64]*FECGroup // Группы пакетов по groupID
	mu         sync.RWMutex
	metrics    *FECDecoderMetrics
}

// FECGroup представляет группу пакетов для FEC
type FECGroup struct {
	groupID     uint64
	createdAt   time.Time
	packetCount int           // k (data symbols)
	symbolLen   int           // aligned length in bytes (after padding)
	present     map[uint64]bool // packetID -> present flag
	packets     map[uint64][]byte // packetID -> packet data (padded)
	redundancy  []byte        // 1 parity symbol (XOR), same length as symbolLen
	received    int          // Количество полученных пакетов
}

// FECDecoderMetrics метрики декодера
type FECDecoderMetrics struct {
	PacketsReceived       int64 `json:"packets_received"`
	RepairPacketsReceived int64 `json:"repair_packets_received"` // Fixed: received, not sent
	PacketsRecovered      int64 `json:"packets_recovered"`
	RecoveryEvents        int64 `json:"recovery_events"`
	FailedRecoveries      int64 `json:"failed_recoveries"`
	GroupsActive          int64 `json:"groups_active"`
	GroupsEvicted         int64 `json:"groups_evicted"`
}

// NewFECDecoder создает новый FEC decoder
func NewFECDecoder() *FECDecoder {
	return &FECDecoder{
		groups:  make(map[uint64]*FECGroup),
		metrics: &FECDecoderMetrics{},
	}
}

// padTo нормализует длину пакета до symbolLen (padding нулями)
func padTo(data []byte, n int) []byte {
	if len(data) >= n {
		return data[:n]
	}
	out := make([]byte, n)
	copy(out, data)
	return out
}

// parseRedundancyHeader безопасно парсит заголовок FEC пакета
func parseRedundancyHeader(b []byte) (groupID uint64, packetCount int, payload []byte, ok bool) {
	if len(b) < 11 || b[0] != 0xFE || b[1] != 0xC0 {
		return 0, 0, nil, false
	}
	
	groupID = binary.LittleEndian.Uint64(b[2:10])
	packetCount = int(b[10])
	
	if packetCount <= 0 || packetCount > maxPacketCount {
		return 0, 0, nil, false
	}
	
	return groupID, packetCount, b[11:], true
}

// AddPacket добавляет обычный пакет в группу
// Возвращает true если был восстановлен пакет
func (d *FECDecoder) AddPacket(packet []byte, packetID uint64, groupID uint64) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	// Проверяем лимиты перед созданием новой группы
	if _, exists := d.groups[groupID]; !exists {
		if len(d.groups) >= maxActiveGroups {
			d.evictOldestGroup()
		}
	}
	
	// Получаем или создаем группу
	group, exists := d.groups[groupID]
	if !exists {
		group = &FECGroup{
			groupID:     groupID,
			createdAt:   time.Now(),
			packets:     make(map[uint64][]byte),
			present:     make(map[uint64]bool),
			packetCount: 0,
		}
		d.groups[groupID] = group
		d.metrics.GroupsActive = int64(len(d.groups))
	}
	
	// Нормализуем длину пакета
	if group.symbolLen == 0 {
		group.symbolLen = len(packet)
		if group.symbolLen > maxSymbolLen {
			group.symbolLen = maxSymbolLen
		}
	}
	
	sym := padTo(packet, group.symbolLen)
	group.packets[packetID] = sym
	group.present[packetID] = true
	group.received++
	d.metrics.PacketsReceived++
	
	// Проверяем, можно ли восстановить недостающие пакеты
	if group.redundancy != nil && group.received < group.packetCount {
		return d.tryRecover(group)
	}
	
	return false
}

// AddRedundancyPacket добавляет redundancy пакет (FEC repair packet)
// Возвращает true и список восстановленных пакетов если восстановление успешно
func (d *FECDecoder) AddRedundancyPacket(redundancyPacket []byte) (bool, []Recovered) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	// Безопасный парсинг заголовка
	groupID, packetCount, payload, ok := parseRedundancyHeader(redundancyPacket)
	if !ok {
		return false, nil
	}
	
	// Проверяем лимиты перед созданием новой группы
	if _, exists := d.groups[groupID]; !exists {
		if len(d.groups) >= maxActiveGroups {
			d.evictOldestGroup()
		}
	}
	
	// Получаем или создаем группу
	group, exists := d.groups[groupID]
	if !exists {
		group = &FECGroup{
			groupID:     groupID,
			createdAt:   time.Now(),
			packets:     make(map[uint64][]byte),
			present:     make(map[uint64]bool),
			packetCount: packetCount,
		}
		d.groups[groupID] = group
		d.metrics.GroupsActive = int64(len(d.groups))
	}
	
	// Проверяем согласованность packetCount
	if group.packetCount != 0 && group.packetCount != packetCount {
		// Конфликтующие значения - удаляем группу
		delete(d.groups, groupID)
		d.metrics.GroupsActive = int64(len(d.groups))
		return false, nil
	}
	
	group.packetCount = packetCount
	
	// Нормализуем длину redundancy
	if group.symbolLen == 0 {
		group.symbolLen = len(payload)
		if group.symbolLen > maxSymbolLen {
			group.symbolLen = maxSymbolLen
		}
	}
	
	group.redundancy = padTo(payload, group.symbolLen)
	d.metrics.RepairPacketsReceived++ // Fixed: received, not sent
	
	// Пытаемся восстановить недостающие пакеты
	if group.received < group.packetCount {
		recovered := d.tryRecover(group)
		if recovered {
			// Возвращаем список восстановленных пакетов
			var recoveredList []Recovered
			for packetID := uint64(0); packetID < uint64(group.packetCount); packetID++ {
				if !group.present[packetID] {
					// Это восстановленный пакет
					if data, exists := group.packets[packetID]; exists {
						recoveredList = append(recoveredList, Recovered{
							PacketID: packetID,
							Data:     data,
						})
					}
				}
			}
			return true, recoveredList
		}
	}
	
	return false, nil
}

// tryRecover пытается восстановить недостающие пакеты используя redundancy
// Ограничение: XOR-FEC восстанавливает только 1 потерянный пакет на группу
func (d *FECDecoder) tryRecover(group *FECGroup) bool {
	if group.redundancy == nil {
		return false
	}
	
	if group.received >= group.packetCount {
		// Все пакеты уже получены
		return false
	}
	
	// Находим недостающие пакеты
	missing := group.packetCount - group.received
	if missing == 0 {
		return false
	}
	
	// XOR-FEC ограничение: только 1 потеря на группу
	if missing == 1 {
		packetID, data, ok := d.recoverSingle(group)
		if ok {
			group.packets[packetID] = data
			group.present[packetID] = true
			group.received++
			d.metrics.RecoveryEvents++
			d.metrics.PacketsRecovered++
			return true
		}
		d.metrics.FailedRecoveries++
	} else {
		// Multiple losses unsupported in XOR FEC
		d.metrics.FailedRecoveries++
		// Log would be: "multiple losses unsupported in XOR FEC, missing=%d", missing
	}
	
	return false
}

// recoverSingle восстанавливает один недостающий пакет через XOR
// Возвращает (packetID, data, ok)
func (d *FECDecoder) recoverSingle(group *FECGroup) (uint64, []byte, bool) {
	if group.redundancy == nil || group.symbolLen == 0 {
		return 0, nil, false
	}
	
	// Находим недостающий packetID
	var missingID *uint64
	for i := uint64(0); i < uint64(group.packetCount); i++ {
		if !group.present[i] {
			missingID = &i
			break
		}
	}
	
	if missingID == nil {
		return 0, nil, false
	}
	
	// Восстанавливаем пакет: recovered = redundancy XOR (all other packets)
	out := make([]byte, group.symbolLen)
	copy(out, group.redundancy)
	
	for id, pkt := range group.packets {
		if id == *missingID {
			continue
		}
		for i := 0; i < group.symbolLen && i < len(pkt); i++ {
			out[i] ^= pkt[i]
		}
	}
	
	return *missingID, out, true
}

// GetMetrics возвращает метрики декодера
func (d *FECDecoder) GetMetrics() *FECDecoderMetrics {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	metrics := *d.metrics
	return &metrics
}

// ResetMetrics сбрасывает метрики
func (d *FECDecoder) ResetMetrics() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.metrics = &FECDecoderMetrics{}
}

// evictOldestGroup удаляет самую старую группу (LRU эвикция)
func (d *FECDecoder) evictOldestGroup() {
	if len(d.groups) == 0 {
		return
	}
	
	var oldestID uint64
	var oldestTime time.Time
	first := true
	
	for id, group := range d.groups {
		if first || group.createdAt.Before(oldestTime) {
			oldestID = id
			oldestTime = group.createdAt
			first = false
		}
	}
	
	if !first {
		delete(d.groups, oldestID)
		d.metrics.GroupsEvicted++
		d.metrics.GroupsActive = int64(len(d.groups))
	}
}

// CleanupGroups удаляет просроченные группы (старше TTL)
func (d *FECDecoder) CleanupGroups() {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	now := time.Now()
	for id, group := range d.groups {
		if now.Sub(group.createdAt) > groupTTL {
			delete(d.groups, id)
			d.metrics.GroupsEvicted++
		}
	}
	d.metrics.GroupsActive = int64(len(d.groups))
}


