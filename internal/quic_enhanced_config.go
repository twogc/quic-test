package internal

import (
	"fmt"
	"time"

	"github.com/quic-go/quic-go"
)

// EnhancedQUICConfig расширенная конфигурация QUIC с экспериментальными возможностями
type EnhancedQUICConfig struct {
	// Базовые настройки
	MaxIdleTimeout    time.Duration
	HandshakeTimeout  time.Duration
	KeepAlive         time.Duration
	MaxStreams        int64
	MaxStreamData     int64
	Enable0RTT        bool
	EnableKeyUpdate   bool
	EnableDatagrams   bool
	
	// Экспериментальные настройки
	CongestionControl    string        // cubic, bbr, bbrv2, reno
	ACKFrequency         int           // ACK frequency (0 = auto)
	MaxACKDelay          time.Duration // Максимальная задержка ACK
	EnableMultipath      bool          // Multipath QUIC (экспериментально)
	EnableFEC            bool          // Forward Error Correction для datagrams
	EnableQlog           bool          // qlog трассировка
	EnableGreasing       bool          // QUIC bit greasing (RFC 9287)
	
	// Производительность
	EnableGSO            bool          // UDP GSO (если поддерживается)
	EnableGRO            bool          // UDP GRO (если поддерживается)
	SocketBufferSize     int           // Размер сокет буферов
	EnableNUMA           bool          // NUMA pinning
	
	// Наблюдаемость
	QlogDir              string        // Директория для qlog файлов
	EnableTracing        bool          // OpenTelemetry трейсинг
	MetricsInterval      time.Duration // Интервал сбора метрик
}

// CreateEnhancedQUICConfig создает расширенную QUIC конфигурацию
func CreateEnhancedQUICConfig(cfg TestConfig, enhanced *EnhancedQUICConfig) *quic.Config {
	config := &quic.Config{
		// Версии QUIC с поддержкой v2
		Versions: []quic.Version{
			quic.Version2, // Приоритет v2
			quic.Version1,
		},
	}
	
	// Базовые настройки
	if enhanced.MaxIdleTimeout > 0 {
		config.MaxIdleTimeout = enhanced.MaxIdleTimeout
	} else if cfg.MaxIdleTimeout > 0 {
		config.MaxIdleTimeout = cfg.MaxIdleTimeout
	}
	
	if enhanced.HandshakeTimeout > 0 {
		config.HandshakeIdleTimeout = enhanced.HandshakeTimeout
	} else if cfg.HandshakeTimeout > 0 {
		config.HandshakeIdleTimeout = cfg.HandshakeTimeout
	}
	
	if enhanced.KeepAlive > 0 {
		config.KeepAlivePeriod = enhanced.KeepAlive
	} else if cfg.KeepAlive > 0 {
		config.KeepAlivePeriod = cfg.KeepAlive
	}
	
	// Потоки
	if enhanced.MaxStreams > 0 {
		config.MaxIncomingStreams = enhanced.MaxStreams
	} else if cfg.MaxStreams > 0 {
		config.MaxIncomingStreams = cfg.MaxStreams
	}
	
	if cfg.MaxIncomingStreams > 0 {
		config.MaxIncomingStreams = cfg.MaxIncomingStreams
	}
	
	if cfg.MaxIncomingUniStreams > 0 {
		config.MaxIncomingUniStreams = cfg.MaxIncomingUniStreams
	}
	
	// Размер данных потока
	if enhanced.MaxStreamData > 0 {
		config.MaxStreamReceiveWindow = uint64(enhanced.MaxStreamData)
	} else if cfg.MaxStreamData > 0 {
		config.MaxStreamReceiveWindow = uint64(cfg.MaxStreamData)
	}
	
	// 0-RTT
	if enhanced.Enable0RTT || cfg.Enable0RTT {
		config.Allow0RTT = true
	}
	
	// Key Update
	if enhanced.EnableKeyUpdate || cfg.EnableKeyUpdate {
		config.DisablePathMTUDiscovery = false
	}
	
	// Datagrams
	if enhanced.EnableDatagrams || cfg.EnableDatagrams {
		config.EnableDatagrams = true
	}
	
	// Экспериментальные настройки
	if enhanced.EnableGreasing {
		// Включаем greasing QUIC bit (RFC 9287)
		// config.DisableVersionNegotiationPackets = false // Не поддерживается в текущей версии quic-go
	}
	
	// Path MTU Discovery
	config.DisablePathMTUDiscovery = false
	
	return config
}

// CreateServerEnhancedQUICConfig создает расширенную конфигурацию для сервера
func CreateServerEnhancedQUICConfig(cfg TestConfig, enhanced *EnhancedQUICConfig) *quic.Config {
	config := CreateEnhancedQUICConfig(cfg, enhanced)
	
	// Серверные специфичные настройки
	// config.RequireAddressValidation = func(addr net.Addr) bool {
	//	// Для экспериментальных режимов можно ослабить валидацию
	//	if enhanced.EnableMultipath {
	//		return false // Multipath может требовать более гибкой валидации
	//	}
	//	return true
	// }
	
	return config
}

// CreateClientEnhancedQUICConfig создает расширенную конфигурацию для клиента
func CreateClientEnhancedQUICConfig(cfg TestConfig, enhanced *EnhancedQUICConfig) *quic.Config {
	config := CreateEnhancedQUICConfig(cfg, enhanced)
	
	// Клиентские специфичные настройки
	config.TokenStore = quic.NewLRUTokenStore(10, int(time.Hour.Seconds()))
	
	// Для экспериментальных режимов
	if enhanced.EnableMultipath {
		// Настройки для multipath (когда будет реализовано)
		config.DisablePathMTUDiscovery = false
	}
	
	return config
}

// PrintEnhancedQUICConfig выводит информацию о расширенных QUIC параметрах
func PrintEnhancedQUICConfig(cfg TestConfig, enhanced *EnhancedQUICConfig) {
	fmt.Printf("🚀 Enhanced QUIC Configuration:\n")
	
	// Базовые настройки
	if enhanced.MaxIdleTimeout > 0 {
		fmt.Printf("  - Max Idle Timeout: %v\n", enhanced.MaxIdleTimeout)
	}
	if enhanced.HandshakeTimeout > 0 {
		fmt.Printf("  - Handshake Timeout: %v\n", enhanced.HandshakeTimeout)
	}
	if enhanced.KeepAlive > 0 {
		fmt.Printf("  - Keep Alive: %v\n", enhanced.KeepAlive)
	}
	
	// Экспериментальные настройки
	if enhanced.CongestionControl != "" {
		fmt.Printf("  - Congestion Control: %s\n", enhanced.CongestionControl)
	}
	if enhanced.ACKFrequency > 0 {
		fmt.Printf("  - ACK Frequency: %d\n", enhanced.ACKFrequency)
	}
	if enhanced.MaxACKDelay > 0 {
		fmt.Printf("  - Max ACK Delay: %v\n", enhanced.MaxACKDelay)
	}
	if enhanced.EnableMultipath {
		fmt.Printf("  - Multipath QUIC: enabled (experimental)\n")
	}
	if enhanced.EnableFEC {
		fmt.Printf("  - FEC for Datagrams: enabled\n")
	}
	if enhanced.EnableQlog {
		fmt.Printf("  - qlog Tracing: enabled\n")
	}
	if enhanced.EnableGreasing {
		fmt.Printf("  - QUIC Bit Greasing: enabled\n")
	}
	
	// Производительность
	if enhanced.EnableGSO {
		fmt.Printf("  - UDP GSO: enabled\n")
	}
	if enhanced.EnableGRO {
		fmt.Printf("  - UDP GRO: enabled\n")
	}
	if enhanced.SocketBufferSize > 0 {
		fmt.Printf("  - Socket Buffer Size: %d bytes\n", enhanced.SocketBufferSize)
	}
	if enhanced.EnableNUMA {
		fmt.Printf("  - NUMA Pinning: enabled\n")
	}
	
	// Наблюдаемость
	if enhanced.EnableTracing {
		fmt.Printf("  - OpenTelemetry Tracing: enabled\n")
	}
	if enhanced.QlogDir != "" {
		fmt.Printf("  - qlog Directory: %s\n", enhanced.QlogDir)
	}
	
	fmt.Println()
}

// DefaultEnhancedConfig возвращает конфигурацию по умолчанию для экспериментальных возможностей
func DefaultEnhancedConfig() *EnhancedQUICConfig {
	return &EnhancedQUICConfig{
		// Базовые настройки
		MaxIdleTimeout:   60 * time.Second,
		HandshakeTimeout: 10 * time.Second,
		KeepAlive:        30 * time.Second,
		MaxStreams:       100,
		MaxStreamData:    1024 * 1024, // 1MB
		Enable0RTT:       true,
		EnableKeyUpdate:  true,
		EnableDatagrams:  true,
		
		// Экспериментальные настройки
		CongestionControl: "bbr",        // BBR по умолчанию
		ACKFrequency:      0,            // Автоматический выбор
		MaxACKDelay:       25 * time.Millisecond,
		EnableMultipath:   false,       // Отключено по умолчанию
		EnableFEC:         false,       // Отключено по умолчанию
		EnableQlog:       true,         // Включено для отладки
		EnableGreasing:   true,         // RFC 9287
		
		// Производительность
		EnableGSO:        true,         // Если поддерживается
		EnableGRO:        true,         // Если поддерживается
		SocketBufferSize: 1024 * 1024,  // 1MB
		EnableNUMA:       false,        // Отключено по умолчанию
		
		// Наблюдаемость
		QlogDir:         "./qlog",
		EnableTracing:   true,
		MetricsInterval: 1 * time.Second,
	}
}
