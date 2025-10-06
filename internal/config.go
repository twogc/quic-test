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
	ReportPath   string        // Путь к файлу для отчёта
	ReportFormat string        // Формат отчёта: csv | md | json
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
	SlaRttP95 time.Duration // SLA: максимальный RTT p95
	SlaLoss   float64       // SLA: максимальная потеря пакетов
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
	return nil
}
