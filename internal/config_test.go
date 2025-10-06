package internal

import (
	"testing"
	"time"
)

func TestTestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  TestConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: TestConfig{
				Mode:        "test",
				Addr:        ":9000",
				Connections: 1,
				Streams:     1,
				Duration:    time.Second,
				PacketSize:  1024,
				Rate:        100,
			},
			wantErr: false,
		},
		{
			name: "invalid connections",
			config: TestConfig{
				Mode:        "test",
				Addr:        ":9000",
				Connections: 0, // Invalid
				Streams:     1,
				Duration:    time.Second,
				PacketSize:  1024,
				Rate:        100,
			},
			wantErr: true,
		},
		{
			name: "invalid streams",
			config: TestConfig{
				Mode:        "test",
				Addr:        ":9000",
				Connections: 1,
				Streams:     0, // Invalid
				Duration:    time.Second,
				PacketSize:  1024,
				Rate:        100,
			},
			wantErr: true,
		},
		{
			name: "invalid packet size",
			config: TestConfig{
				Mode:        "test",
				Addr:        ":9000",
				Connections: 1,
				Streams:     1,
				Duration:    time.Second,
				PacketSize:  0, // Invalid
				Rate:        100,
			},
			wantErr: true,
		},
		{
			name: "invalid rate",
			config: TestConfig{
				Mode:        "test",
				Addr:        ":9000",
				Connections: 1,
				Streams:     1,
				Duration:    time.Second,
				PacketSize:  1024,
				Rate:        0, // Invalid
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("TestConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTestConfig_DefaultValues(t *testing.T) {
	config := TestConfig{}
	
	// Проверяем, что пустая конфигурация невалидна
	err := config.Validate()
	if err == nil {
		t.Error("Empty config should be invalid")
	}
	
	// Проверяем, что можем создать валидную конфигурацию
	validConfig := TestConfig{
		Mode:        "test",
		Addr:        ":9000",
		Connections: 1,
		Streams:     1,
		Duration:    time.Second,
		PacketSize:  1024,
		Rate:        100,
	}
	
	err = validConfig.Validate()
	if err != nil {
		t.Errorf("Valid config should not have errors: %v", err)
	}
}
