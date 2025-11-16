package internal

import (
	"fmt"
	"time"
)

// NetworkProfile описывает сетевой профиль
type NetworkProfile struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	RTT         time.Duration `json:"rtt"`
	Jitter      time.Duration `json:"jitter"`
	Loss        float64       `json:"loss"`
	Bandwidth   float64       `json:"bandwidth"` // KB/s
	Duplication float64       `json:"duplication"`
	Latency     time.Duration `json:"latency"`
}

// GetNetworkProfile возвращает предустановленный сетевой профиль
func GetNetworkProfile(name string) (*NetworkProfile, error) {
	profiles := map[string]NetworkProfile{
		"wifi": {
			Name:        "WiFi 802.11n",
			Description: "Стандартная WiFi сеть 802.11n",
			RTT:         20 * time.Millisecond,
			Jitter:      5 * time.Millisecond,
			Loss:        0.02, // 2%
			Bandwidth:   1000, // 1 MB/s
			Duplication: 0.01, // 1%
			Latency:     10 * time.Millisecond,
		},
		"wifi-5g": {
			Name:        "WiFi 802.11ac (5GHz)",
			Description: "Быстрая WiFi сеть 802.11ac на 5GHz",
			RTT:         10 * time.Millisecond,
			Jitter:      2 * time.Millisecond,
			Loss:        0.01, // 1%
			Bandwidth:   5000, // 5 MB/s
			Duplication: 0.005, // 0.5%
			Latency:     5 * time.Millisecond,
		},
		"lte": {
			Name:        "LTE 4G",
			Description: "Мобильная LTE сеть 4G",
			RTT:         50 * time.Millisecond,
			Jitter:      15 * time.Millisecond,
			Loss:        0.05, // 5%
			Bandwidth:   2000, // 2 MB/s
			Duplication: 0.02, // 2%
			Latency:     30 * time.Millisecond,
		},
		"lte-advanced": {
			Name:        "LTE Advanced",
			Description: "Продвинутая LTE сеть с агрегацией несущих",
			RTT:         30 * time.Millisecond,
			Jitter:      10 * time.Millisecond,
			Loss:        0.03, // 3%
			Bandwidth:   8000, // 8 MB/s
			Duplication: 0.015, // 1.5%
			Latency:     20 * time.Millisecond,
		},
		"5g": {
			Name:        "5G NR",
			Description: "Сеть 5G New Radio",
			RTT:         5 * time.Millisecond,
			Jitter:      1 * time.Millisecond,
			Loss:        0.001, // 0.1%
			Bandwidth:   50000, // 50 MB/s
			Duplication: 0.0005, // 0.05%
			Latency:     2 * time.Millisecond,
		},
		"satellite": {
			Name:        "Satellite Internet",
			Description: "Спутниковый интернет (геостационарная орбита)",
			RTT:         500 * time.Millisecond,
			Jitter:      50 * time.Millisecond,
			Loss:        0.01, // 1%
			Bandwidth:   500, // 500 KB/s
			Duplication: 0.005, // 0.5%
			Latency:     250 * time.Millisecond,
		},
		"satellite-leo": {
			Name:        "Satellite LEO",
			Description: "Спутниковый интернет низкой орбиты (Starlink)",
			RTT:         50 * time.Millisecond,
			Jitter:      10 * time.Millisecond,
			Loss:        0.02, // 2%
			Bandwidth:   10000, // 10 MB/s
			Duplication: 0.01, // 1%
			Latency:     25 * time.Millisecond,
		},
		"ethernet": {
			Name:        "Ethernet 1Gbps",
			Description: "Проводная Ethernet сеть 1 Гбит/с",
			RTT:         1 * time.Millisecond,
			Jitter:      100 * time.Microsecond,
			Loss:        0.0001, // 0.01%
			Bandwidth:   100000, // 100 MB/s
			Duplication: 0.0001, // 0.01%
			Latency:     500 * time.Microsecond,
		},
		"ethernet-10g": {
			Name:        "Ethernet 10Gbps",
			Description: "Проводная Ethernet сеть 10 Гбит/с",
			RTT:         100 * time.Microsecond,
			Jitter:      10 * time.Microsecond,
			Loss:        0.00001, // 0.001%
			Bandwidth:   1000000, // 1 GB/s
			Duplication: 0.00001, // 0.001%
			Latency:     50 * time.Microsecond,
		},
		"dsl": {
			Name:        "DSL",
			Description: "Цифровая абонентская линия",
			RTT:         30 * time.Millisecond,
			Jitter:      5 * time.Millisecond,
			Loss:        0.01, // 1%
			Bandwidth:   2000, // 2 MB/s
			Duplication: 0.005, // 0.5%
			Latency:     15 * time.Millisecond,
		},
		"cable": {
			Name:        "Cable Internet",
			Description: "Кабельный интернет",
			RTT:         15 * time.Millisecond,
			Jitter:      3 * time.Millisecond,
			Loss:        0.005, // 0.5%
			Bandwidth:   10000, // 10 MB/s
			Duplication: 0.002, // 0.2%
			Latency:     8 * time.Millisecond,
		},
		"fiber": {
			Name:        "Fiber Optic",
			Description: "Оптоволоконная связь",
			RTT:         2 * time.Millisecond,
			Jitter:      500 * time.Microsecond,
			Loss:        0.0001, // 0.01%
			Bandwidth:   100000, // 100 MB/s
			Duplication: 0.0001, // 0.01%
			Latency:     1 * time.Millisecond,
		},
		"mobile-3g": {
			Name:        "3G Mobile",
			Description: "Мобильная сеть 3G",
			RTT:         100 * time.Millisecond,
			Jitter:      30 * time.Millisecond,
			Loss:        0.1, // 10%
			Bandwidth:   500, // 500 KB/s
			Duplication: 0.05, // 5%
			Latency:     50 * time.Millisecond,
		},
		"edge": {
			Name:        "EDGE Mobile",
			Description: "Мобильная сеть EDGE (2.5G)",
			RTT:         200 * time.Millisecond,
			Jitter:      50 * time.Millisecond,
			Loss:        0.15, // 15%
			Bandwidth:   100, // 100 KB/s
			Duplication: 0.1, // 10%
			Latency:     100 * time.Millisecond,
		},
		"international": {
			Name:        "International Link",
			Description: "Международное соединение (Россия-Европа)",
			RTT:         80 * time.Millisecond,
			Jitter:      20 * time.Millisecond,
			Loss:        0.03, // 3%
			Bandwidth:   5000, // 5 MB/s
			Duplication: 0.01, // 1%
			Latency:     40 * time.Millisecond,
		},
		"datacenter": {
			Name:        "Data Center",
			Description: "Внутри дата-центра",
			RTT:         100 * time.Microsecond,
			Jitter:      10 * time.Microsecond,
			Loss:        0.000001, // 0.0001%
			Bandwidth:   10000000, // 10 GB/s
			Duplication: 0.000001, // 0.0001%
			Latency:     50 * time.Microsecond,
		},
	}
	
	profile, exists := profiles[name]
	if !exists {
		return nil, fmt.Errorf("сетевой профиль '%s' не найден", name)
	}
	
	return &profile, nil
}

// ListNetworkProfiles возвращает список доступных сетевых профилей
func ListNetworkProfiles() []string {
	return []string{
		"wifi",
		"wifi-5g",
		"lte",
		"lte-advanced",
		"5g",
		"satellite",
		"satellite-leo",
		"ethernet",
		"ethernet-10g",
		"dsl",
		"cable",
		"fiber",
		"mobile-3g",
		"edge",
		"international",
		"datacenter",
	}
}

// PrintNetworkProfile выводит информацию о сетевом профиле
func PrintNetworkProfile(profile *NetworkProfile) {
	fmt.Printf("🌐 Network Profile: %s\n", profile.Name)
	fmt.Printf("📝 Description: %s\n", profile.Description)
	fmt.Printf("Characteristics:\n")
	fmt.Printf("  - RTT: %v\n", profile.RTT)
	fmt.Printf("  - Jitter: %v\n", profile.Jitter)
	fmt.Printf("  - Loss: %.2f%%\n", profile.Loss*100)
	fmt.Printf("  - Bandwidth: %.1f KB/s (%.1f MB/s)\n", profile.Bandwidth, profile.Bandwidth/1000)
	fmt.Printf("  - Duplication: %.2f%%\n", profile.Duplication*100)
	fmt.Printf("  - Latency: %v\n", profile.Latency)
	fmt.Println()
}

// ApplyNetworkProfile применяет сетевой профиль к конфигурации теста
func ApplyNetworkProfile(cfg *TestConfig, profile *NetworkProfile) {
	cfg.EmulateLoss = profile.Loss
	cfg.EmulateLatency = profile.Latency
	cfg.EmulateDup = profile.Duplication
	
	// Адаптируем параметры теста под профиль
	if profile.Bandwidth < 1000 { // Медленная сеть
		cfg.Rate = 50
		cfg.Connections = 1
		cfg.Streams = 2
	} else if profile.Bandwidth < 10000 { // Средняя сеть
		cfg.Rate = 100
		cfg.Connections = 2
		cfg.Streams = 4
	} else { // Быстрая сеть
		cfg.Rate = 200
		cfg.Connections = 4
		cfg.Streams = 8
	}
	
	// Адаптируем размер пакета под RTT
	if profile.RTT > 100*time.Millisecond {
		cfg.PacketSize = 800 // Меньшие пакеты для высоких задержек
	} else if profile.RTT < 10*time.Millisecond {
		cfg.PacketSize = 1400 // Большие пакеты для низких задержек
	}
}

// GetProfileRecommendations возвращает рекомендации по настройке QUIC для профиля
func GetProfileRecommendations(profile *NetworkProfile) []string {
	var recommendations []string
	
	// Рекомендации по алгоритму управления перегрузкой
	if profile.RTT > 100*time.Millisecond {
		recommendations = append(recommendations, "Use BBR congestion control for high latency networks")
	} else if profile.Loss > 0.05 {
		recommendations = append(recommendations, "Use CUBIC congestion control for lossy networks")
	} else {
		recommendations = append(recommendations, "Use default congestion control for stable networks")
	}
	
	// Рекомендации по таймаутам
	if profile.RTT > 200*time.Millisecond {
		recommendations = append(recommendations, "Increase handshake timeout to 30s for satellite links")
		recommendations = append(recommendations, "Increase max idle timeout to 5 minutes")
	}
	
	// Рекомендации по потокам
	if profile.Bandwidth > 100000 { // Очень быстрая сеть
		recommendations = append(recommendations, "Enable 0-RTT for faster connection establishment")
		recommendations = append(recommendations, "Enable key update for long-lived connections")
		recommendations = append(recommendations, "Enable datagrams for real-time applications")
	}
	
	// Рекомендации по размеру окна
	if profile.RTT > 50*time.Millisecond {
		recommendations = append(recommendations, "Increase stream receive window size")
		recommendations = append(recommendations, "Enable flow control optimization")
	}
	
	return recommendations
}

// PrintProfileRecommendations выводит рекомендации по настройке
func PrintProfileRecommendations(profile *NetworkProfile) {
	recommendations := GetProfileRecommendations(profile)
	if len(recommendations) > 0 {
		fmt.Printf("💡 Recommendations for %s:\n", profile.Name)
		for i, rec := range recommendations {
			fmt.Printf("  %d. %s\n", i+1, rec)
		}
		fmt.Println()
	}
}
