package internal

import (
	"errors"
	"time"
)

// TestConfig описывает параметры теста для клиента и сервера.
type TestConfig struct {
	Mode         string        // Режим работы: server | client | test
	Addr         string        // Адрес для подключения или прослушивания
	Streams      int           // Количество потоков на соединение
	Connections  int           // Количество соединений
	Duration     time.Duration // Длительность теста
	PacketSize   int           // Размер пакета (байт)
	Rate         int           // Частота отправки пакетов (в секунду)
	ReportPath   string        // Путь к файлу для отчета
	ReportFormat string        // Формат отчета: csv | md | json
	CertPath     string        // Путь к TLS-сертификату
	KeyPath      string        // Путь к TLS-ключу
	Pattern      string        // Шаблон данных: random | zeroes | increment
	NoTLS        bool          // Отключить TLS
	Prometheus   bool          // Экспортировать метрики Prometheus

	// --- Эмуляция плохих сетей ---
	EmulateLoss    float64       // вероятность потери пакета (0..1)
	EmulateLatency time.Duration // дополнительная задержка
	EmulateDup     float64       // вероятность дублирования пакета (0..1)

	// --- Профилирование и мониторинг ---
	PprofAddr string // Адрес для pprof (например, :6060)

	// --- SLA проверки ---
	SlaRttP95     time.Duration // SLA: максимальный RTT p95
	SlaLoss       float64       // SLA: максимальная потеря пакетов
	SlaThroughput float64       // SLA: минимальная пропускная способность (KB/s)
	SlaErrors     int64         // SLA: максимальное количество ошибок
	
	// --- QUIC тюнинг ---
	CongestionControl string        // Алгоритм управления перегрузкой: cubic, bbr, reno
	MaxIdleTimeout    time.Duration // Максимальное время простоя соединения
	HandshakeTimeout  time.Duration // Таймаут handshake
	KeepAlive         time.Duration // Интервал keep-alive
	MaxStreams        int64         // Максимальное количество потоков
	MaxStreamData     int64         // Максимальный размер данных потока
	Enable0RTT        bool          // Включить 0-RTT
	EnableKeyUpdate   bool          // Включить key update
	EnableDatagrams   bool          // Включить datagrams
	MaxIncomingStreams int64        // Максимальное количество входящих потоков
	MaxIncomingUniStreams int64     // Максимальное количество входящих unidirectional потоков
	
	// --- FEC (Forward Error Correction) ---
	FECEnabled    bool    // Включить Forward Error Correction
	FECRedundancy float64 // Уровень избыточности FEC (0.0-1.0, например 0.05 = 5%, 0.10 = 10%, 0.20 = 20%)
	
	// --- PQC (Post-Quantum Cryptography) ---
	PQCEnabled  bool   // Включить Post-Quantum Cryptography (симуляция)
	PQCAlgorithm string // PQC алгоритм: "ml-kem-512", "ml-kem-768", "dilithium-2", "hybrid", "baseline"

	// --- AI Routing ---
	AIEnabled    bool   // Включить AI-маршрутизацию
	AIServiceURL string // URL сервиса прогнозирования (например, http://localhost:5000)
}

// Validate проверяет корректность конфигурации
func (cfg *TestConfig) Validate() error {
	if cfg.Connections <= 0 {
		return errors.New("connections must be positive")
	}
	if cfg.Streams <= 0 {
		return errors.New("streams must be positive")
	}
	if cfg.Duration <= 0 {
		return errors.New("duration must be positive")
	}
	if cfg.PacketSize <= 0 {
		return errors.New("packet size must be positive")
	}
	if cfg.Rate <= 0 {
		return errors.New("rate must be positive")
	}
	if cfg.EmulateLoss < 0 || cfg.EmulateLoss > 1 {
		return errors.New("emulate loss must be between 0 and 1")
	}
	if cfg.EmulateDup < 0 || cfg.EmulateDup > 1 {
		return errors.New("emulate dup must be between 0 and 1")
	}
	if cfg.SlaLoss < 0 || cfg.SlaLoss > 1 {
		return errors.New("SLA loss must be between 0 and 1")
	}
	
	// Валидация QUIC параметров
	validCC := map[string]bool{
		"cubic": true, "bbr": true, "bbrv2": true, "bbrv3": true, "reno": true,
	}
	if cfg.CongestionControl != "" && !validCC[cfg.CongestionControl] {
		return errors.New("congestion control must be one of: cubic, bbr, bbrv2, bbrv3, reno")
	}
	if cfg.MaxIdleTimeout < 0 {
		return errors.New("max idle timeout must be non-negative")
	}
	if cfg.HandshakeTimeout < 0 {
		return errors.New("handshake timeout must be non-negative")
	}
	if cfg.KeepAlive < 0 {
		return errors.New("keep alive must be non-negative")
	}
	if cfg.MaxStreams < 0 {
		return errors.New("max streams must be non-negative")
	}
	if cfg.MaxStreamData < 0 {
		return errors.New("max stream data must be non-negative")
	}
	if cfg.MaxIncomingStreams < 0 {
		return errors.New("max incoming streams must be non-negative")
	}
	if cfg.MaxIncomingUniStreams < 0 {
		return errors.New("max incoming uni streams must be non-negative")
	}
	
	// Валидация FEC параметров
	if cfg.FECRedundancy < 0 || cfg.FECRedundancy > 1 {
		return errors.New("FEC redundancy must be between 0 and 1")
	}
	
	return nil
}
