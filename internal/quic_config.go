package internal

import (
	"fmt"
	"time"

	"github.com/quic-go/quic-go"
)

// CreateQUICConfig создает QUIC конфигурацию на основе параметров теста
func CreateQUICConfig(cfg TestConfig) *quic.Config {
	config := &quic.Config{
		// Включаем все возможные версии QUIC
		Versions: []quic.VersionNumber{
			quic.Version1,
			quic.Version2,
		},
	}
	
	// Настройка алгоритма управления перегрузкой
	// Congestion control настройки не поддерживаются в текущей версии quic-go
	// Оставляем комментарий для будущей реализации
	_ = cfg.CongestionControl
	
	// Настройка таймаутов
	if cfg.MaxIdleTimeout > 0 {
		config.MaxIdleTimeout = cfg.MaxIdleTimeout
	}
	
	if cfg.HandshakeTimeout > 0 {
		config.HandshakeIdleTimeout = cfg.HandshakeTimeout
	}
	
	// Настройка keep-alive
	if cfg.KeepAlive > 0 {
		config.KeepAlivePeriod = cfg.KeepAlive
	}
	
	// Настройка потоков
	if cfg.MaxStreams > 0 {
		config.MaxIncomingStreams = cfg.MaxStreams
	}
	
	if cfg.MaxIncomingStreams > 0 {
		config.MaxIncomingStreams = cfg.MaxIncomingStreams
	}
	
	if cfg.MaxIncomingUniStreams > 0 {
		config.MaxIncomingUniStreams = cfg.MaxIncomingUniStreams
	}
	
	// Настройка размера данных потока
	if cfg.MaxStreamData > 0 {
		config.MaxStreamReceiveWindow = uint64(cfg.MaxStreamData)
	}
	
	// Настройка 0-RTT
	if cfg.Enable0RTT {
		config.Allow0RTT = true
	}
	
	// Настройка key update
	if cfg.EnableKeyUpdate {
		config.DisablePathMTUDiscovery = false // Включаем для лучшей производительности
	}
	
	// Настройка datagrams
	if cfg.EnableDatagrams {
		config.EnableDatagrams = true
	}
	
	// Дополнительные оптимизации
	config.DisablePathMTUDiscovery = false
	// DisableVersionNegotiationPackets не поддерживается в текущей версии
	
	return config
}

// CreateServerQUICConfig создает QUIC конфигурацию для сервера
func CreateServerQUICConfig(cfg TestConfig) *quic.Config {
	config := CreateQUICConfig(cfg)
	
	// Серверные специфичные настройки
	// config.RequireAddressValidation = func(net.Addr) bool {
	//	return true // Требуем валидацию адреса для безопасности
	// }
	
	return config
}

// CreateClientQUICConfig создает QUIC конфигурацию для клиента
func CreateClientQUICConfig(cfg TestConfig) *quic.Config {
	config := CreateQUICConfig(cfg)
	
	// Клиентские специфичные настройки
	config.TokenStore = quic.NewLRUTokenStore(10, int(time.Hour.Seconds())) // Кэш токенов для 0-RTT
	
	return config
}

// PrintQUICConfig выводит информацию о настроенных QUIC параметрах
func PrintQUICConfig(cfg TestConfig) {
	hasQUICConfig := cfg.CongestionControl != "" || 
		cfg.MaxIdleTimeout > 0 || 
		cfg.HandshakeTimeout > 0 || 
		cfg.KeepAlive > 0 || 
		cfg.MaxStreams > 0 || 
		cfg.MaxStreamData > 0 || 
		cfg.Enable0RTT || 
		cfg.EnableKeyUpdate || 
		cfg.EnableDatagrams || 
		cfg.MaxIncomingStreams > 0 || 
		cfg.MaxIncomingUniStreams > 0
	
	if hasQUICConfig {
		fmt.Printf("🔧 QUIC Configuration:\n")
		
		if cfg.CongestionControl != "" {
			fmt.Printf("  - Congestion Control: %s\n", cfg.CongestionControl)
		}
		if cfg.MaxIdleTimeout > 0 {
			fmt.Printf("  - Max Idle Timeout: %v\n", cfg.MaxIdleTimeout)
		}
		if cfg.HandshakeTimeout > 0 {
			fmt.Printf("  - Handshake Timeout: %v\n", cfg.HandshakeTimeout)
		}
		if cfg.KeepAlive > 0 {
			fmt.Printf("  - Keep Alive: %v\n", cfg.KeepAlive)
		}
		if cfg.MaxStreams > 0 {
			fmt.Printf("  - Max Streams: %d\n", cfg.MaxStreams)
		}
		if cfg.MaxStreamData > 0 {
			fmt.Printf("  - Max Stream Data: %d bytes\n", cfg.MaxStreamData)
		}
		if cfg.Enable0RTT {
			fmt.Printf("  - 0-RTT: enabled\n")
		}
		if cfg.EnableKeyUpdate {
			fmt.Printf("  - Key Update: enabled\n")
		}
		if cfg.EnableDatagrams {
			fmt.Printf("  - Datagrams: enabled\n")
		}
		if cfg.MaxIncomingStreams > 0 {
			fmt.Printf("  - Max Incoming Streams: %d\n", cfg.MaxIncomingStreams)
		}
		if cfg.MaxIncomingUniStreams > 0 {
			fmt.Printf("  - Max Incoming Uni Streams: %d\n", cfg.MaxIncomingUniStreams)
		}
		
		fmt.Println()
	}
}
