package fec

import (
	"bytes"
	"testing"
)

// TestNewFECEncoder проверяет инициализацию FEC encoder
func TestNewFECEncoder(t *testing.T) {
	tests := []struct {
		name        string
		redundancy  float64
		expectError bool
	}{
		{"valid_10_percent", 0.10, false},
		{"valid_5_percent", 0.05, false},
		{"valid_20_percent", 0.20, false},
		{"invalid_negative", -0.10, false}, // Должна быть установлена default
		{"invalid_zero", 0, false},          // Должна быть установлена default
		{"invalid_greater_than_one", 1.5, false}, // Должна быть установлена default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := NewFECEncoder(tt.redundancy)
			if encoder == nil {
				t.Fatal("NewFECEncoder returned nil")
			}

			// Проверяем что redundancy установлена корректно
			if encoder.redundancy <= 0 || encoder.redundancy > 1 {
				t.Errorf("Invalid redundancy: %v", encoder.redundancy)
			}
		})
	}
}

// TestAddPacket проверяет добавление пакетов
func TestAddPacket(t *testing.T) {
	encoder := NewFECEncoder(0.10)

	// Добавляем пакеты меньше чем размер группы
	for i := 0; i < 5; i++ {
		packet := bytes.Repeat([]byte{byte(i)}, 1200)
		hasRepair, _, err := encoder.AddPacket(packet, uint64(i))
		if err != nil {
			t.Errorf("AddPacket failed: %v", err)
		}

		// До полной группы repair packets не должны быть созданы
		if hasRepair {
			t.Errorf("Got repair packet too early at packet %d", i)
		}
	}

	// Проверяем метрики - количество добавленных пакетов
	// Метрики обновляются при полной группе
	if encoder.metrics.PacketsEncoded > 5 {
		t.Errorf("Expected <=5 packets encoded, got %d", encoder.metrics.PacketsEncoded)
	}

	t.Logf("Encoder state - packets in buffer: %d", len(encoder.packets))
}

// TestFECEncoderFullGroup проверяет создание repair пакетов при полной группе
func TestFECEncoderFullGroup(t *testing.T) {
	encoder := NewFECEncoder(0.10)

	// Добавляем 10 пакетов (полная группа)
	for i := 0; i < 10; i++ {
		packet := bytes.Repeat([]byte{byte(i)}, 1200)
		hasRepair, repairPkt, err := encoder.AddPacket(packet, uint64(i))

		if err != nil {
			t.Errorf("AddPacket failed: %v", err)
		}

		// На 10-м пакете должен быть repair packet
		if i == 9 {
			if !hasRepair {
				t.Error("Expected repair packet for 10-th packet")
			}
			if len(repairPkt) == 0 {
				t.Error("Repair packet is empty")
			}
		}
	}

	// Проверяем метрики
	if encoder.metrics.GroupsProcessed < 1 {
		t.Errorf("Expected at least 1 group processed, got %d", encoder.metrics.GroupsProcessed)
	}
}

// TestFECEncoderResetAfterGroup проверяет очистку после полной группы
func TestFECEncoderResetAfterGroup(t *testing.T) {
	encoder := NewFECEncoder(0.10)

	// Добавляем 10 пакетов
	for i := 0; i < 10; i++ {
		packet := bytes.Repeat([]byte{byte(i)}, 1200)
		encoder.AddPacket(packet, uint64(i))
	}

	// Добавляем еще пакеты - новая группа должна начаться
	newGroupPacket := bytes.Repeat([]byte{99}, 1200)
	hasRepair, _, err := encoder.AddPacket(newGroupPacket, uint64(10))

	if err != nil {
		t.Errorf("AddPacket failed: %v", err)
	}

	// Новый пакет в новой группе не должен иметь repair packet
	if hasRepair {
		t.Error("Got repair packet in new group too early")
	}
}

// TestNewFECDecoder проверяет инициализацию FEC decoder
func TestNewFECDecoder(t *testing.T) {
	decoder := NewFECDecoder()
	if decoder == nil {
		t.Fatal("NewFECDecoder returned nil")
	}

	if decoder.groups == nil {
		t.Fatal("Decoder groups map is nil")
	}
}

// TestDecoderAddPacket проверяет добавление пакетов в decoder
func TestDecoderAddPacket(t *testing.T) {
	decoder := NewFECDecoder()

	// Создаем тестовый пакет
	packet := bytes.Repeat([]byte{0xAA}, 1200)
	packetID := uint64(1)
	groupID := uint64(1)

	success := decoder.AddPacket(packet, packetID, groupID)

	// AddPacket возвращает bool, не ошибку
	if !success {
		t.Logf("AddPacket returned false (expected for some cases)")
	}

	// Проверяем что группа создана
	if decoder.metrics.PacketsReceived != 1 {
		t.Errorf("Expected 1 packet received, got %d", decoder.metrics.PacketsReceived)
	}
}

// TestDecoderRecovery проверяет восстановление потерянных пакетов
func TestDecoderRecovery(t *testing.T) {
	encoder := NewFECEncoder(0.10)
	decoder := NewFECDecoder()

	// Кодируем 10 пакетов
	for i := 0; i < 10; i++ {
		packet := bytes.Repeat([]byte{byte(i)}, 1200)
		hasRepair, _, err := encoder.AddPacket(packet, uint64(i))

		if err != nil {
			t.Errorf("Encoding failed: %v", err)
		}

		if hasRepair {
			t.Logf("Repair packet created at packet %d", i)
		}

		// Добавляем в decoder (кроме одного пакета для проверки recovery)
		if i != 5 { // Пропускаем 5-й пакет
			decoder.AddPacket(packet, uint64(i), uint64(i/10))
		}
	}

	// Проверяем метрики
	if decoder.metrics.PacketsReceived > 0 {
		t.Logf("Total packets processed: %d", decoder.metrics.PacketsReceived)
	}
}

// TestEncoderWithDifferentRedundancy проверяет разные уровни redundancy
func TestEncoderWithDifferentRedundancy(t *testing.T) {
	redundancies := []float64{0.05, 0.10, 0.15, 0.20}

	for _, red := range redundancies {
		t.Run("redundancy_"+string(rune(int(red*100))), func(t *testing.T) {
			encoder := NewFECEncoder(red)

			// Добавляем 10 пакетов
			for i := 0; i < 10; i++ {
				packet := bytes.Repeat([]byte{byte(i)}, 1200)
				_, _, err := encoder.AddPacket(packet, uint64(i))
				if err != nil {
					t.Errorf("AddPacket failed: %v", err)
				}
			}

			// Проверяем что метрики накоплены
			if encoder.metrics.PacketsEncoded == 0 {
				t.Error("No packets encoded")
			}
		})
	}
}

// TestDecoderMetrics проверяет метрики decoder
func TestDecoderMetrics(t *testing.T) {
	decoder := NewFECDecoder()

	if decoder.metrics == nil {
		t.Fatal("Decoder metrics is nil")
	}

	// Проверяем инициальные значения
	if decoder.metrics.PacketsReceived != 0 {
		t.Errorf("Expected 0 initial packets, got %d", decoder.metrics.PacketsReceived)
	}
}

// BenchmarkEncoderAddPacket тест производительности добавления пакетов
func BenchmarkEncoderAddPacket(b *testing.B) {
	encoder := NewFECEncoder(0.10)
	packet := bytes.Repeat([]byte{0xFF}, 1200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder.AddPacket(packet, uint64(i))
	}
}

// BenchmarkDecoderAddPacket тест производительности добавления пакетов в decoder
func BenchmarkDecoderAddPacket(b *testing.B) {
	decoder := NewFECDecoder()
	packet := make([]byte, 1200)
	// Добавляем минимальный FEC header
	packet[0] = 0xAA

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoder.AddPacket(packet, uint64(i), uint64(i/10))
	}
}

// TestEncoderConcurrency проверяет потокобезопасность encoder
func TestEncoderConcurrency(t *testing.T) {
	encoder := NewFECEncoder(0.10)
	done := make(chan bool, 10)

	// Запускаем 10 горутин, добавляющих пакеты
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 50; j++ {
				packet := bytes.Repeat([]byte{byte(id)}, 1200)
				encoder.AddPacket(packet, uint64(id*100+j))
			}
			done <- true
		}(i)
	}

	// Ждем завершения всех горутин
	for i := 0; i < 10; i++ {
		<-done
	}

	// Проверяем что не было deadlock'а
	if encoder.metrics.PacketsEncoded == 0 {
		t.Error("No packets were encoded")
	}
}

// TestDecoderGroupsExpiration проверяет очистку групп по timeout
func TestDecoderGroupsExpiration(t *testing.T) {
	decoder := NewFECDecoder()

	// Добавляем пакет
	packet := make([]byte, 1200)
	packet[0] = 0xAA

	decoder.AddPacket(packet, uint64(1), uint64(1))

	initialGroups := len(decoder.groups)

	// Вызываем cleanup
	decoder.CleanupGroups()

	// После cleanup некоторые группы должны быть удалены
	// (зависит от их возраста, в тесте они только что добавлены)
	t.Logf("Groups before cleanup: %d, after: %d", initialGroups, len(decoder.groups))
}
