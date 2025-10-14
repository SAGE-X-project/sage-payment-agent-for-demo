package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear environment
	os.Clearenv()

	cfg := LoadConfig()

	if cfg.AgentPort != "8091" {
		t.Errorf("AgentPort: got %s, want 8091", cfg.AgentPort)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel: got %s, want info", cfg.LogLevel)
	}

	if !cfg.SAGEEnabled {
		t.Error("SAGEEnabled should be true by default")
	}

	if !cfg.StrictMode {
		t.Error("StrictMode should be true by default")
	}

	if !cfg.SimulationMode {
		t.Error("SimulationMode should be true by default")
	}
}

func TestLoadConfig_CustomValues(t *testing.T) {
	os.Setenv("AGENT_PORT", "9091")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("SAGE_ENABLED", "false")
	os.Setenv("SAGE_STRICT_MODE", "false")
	os.Setenv("TX_SIMULATION_MODE", "false")
	os.Setenv("TX_DELAY_MS", "1000")
	os.Setenv("WALLET_ADDRESS", "0xCustomWallet")
	defer os.Clearenv()

	cfg := LoadConfig()

	if cfg.AgentPort != "9091" {
		t.Errorf("AgentPort: got %s, want 9091", cfg.AgentPort)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel: got %s, want debug", cfg.LogLevel)
	}

	if cfg.SAGEEnabled {
		t.Error("SAGEEnabled should be false")
	}

	if cfg.StrictMode {
		t.Error("StrictMode should be false")
	}

	if cfg.SimulationMode {
		t.Error("SimulationMode should be false")
	}

	if cfg.TxDelayMs != 1000 {
		t.Errorf("TxDelayMs: got %d, want 1000", cfg.TxDelayMs)
	}

	if cfg.WalletAddress != "0xCustomWallet" {
		t.Errorf("WalletAddress: got %s, want 0xCustomWallet", cfg.WalletAddress)
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue string
		expected     string
	}{
		{"With env value", "custom", "default", "custom"},
		{"Without env value", "", "default", "default"},
		{"Empty string env", "", "fallback", "fallback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			if tt.envValue != "" {
				os.Setenv("TEST_KEY", tt.envValue)
			}

			result := getEnv("TEST_KEY", tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv: got %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{"True string", "true", false, true},
		{"False string", "false", true, false},
		{"1 string", "1", false, true},
		{"0 string", "0", true, false},
		{"Invalid string", "invalid", false, false},
		{"Empty string", "", true, true},
		{"Missing env", "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			if tt.envValue != "" {
				os.Setenv("TEST_BOOL", tt.envValue)
			}

			result := getEnvBool("TEST_BOOL", tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvBool: got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue int
		expected     int
	}{
		{"Valid int", "123", 0, 123},
		{"Negative int", "-456", 0, -456},
		{"Invalid int", "invalid", 999, 999},
		{"Empty string", "", 888, 888},
		{"Missing env", "", 777, 777},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			if tt.envValue != "" {
				os.Setenv("TEST_INT", tt.envValue)
			}

			result := getEnvInt("TEST_INT", tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvInt: got %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestConfig_IsSAGEEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"SAGE enabled", true},
		{"SAGE disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{SAGEEnabled: tt.enabled}
			if cfg.IsSAGEEnabled() != tt.enabled {
				t.Errorf("IsSAGEEnabled: got %v, want %v", cfg.IsSAGEEnabled(), tt.enabled)
			}
		})
	}
}

func TestConfig_IsStrictMode(t *testing.T) {
	tests := []struct {
		name   string
		strict bool
	}{
		{"Strict mode ON", true},
		{"Strict mode OFF", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{StrictMode: tt.strict}
			if cfg.IsStrictMode() != tt.strict {
				t.Errorf("IsStrictMode: got %v, want %v", cfg.IsStrictMode(), tt.strict)
			}
		})
	}
}

func TestConfig_IsSimulationMode(t *testing.T) {
	tests := []struct {
		name       string
		simulation bool
	}{
		{"Simulation ON", true},
		{"Simulation OFF", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{SimulationMode: tt.simulation}
			if cfg.IsSimulationMode() != tt.simulation {
				t.Errorf("IsSimulationMode: got %v, want %v", cfg.IsSimulationMode(), tt.simulation)
			}
		})
	}
}

func TestConfig_GetTxDelay(t *testing.T) {
	tests := []struct {
		name     string
		delayMs  int
		expected time.Duration
	}{
		{"500ms", 500, 500 * time.Millisecond},
		{"1000ms", 1000, 1 * time.Second},
		{"0ms", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{TxDelayMs: tt.delayMs}
			if cfg.GetTxDelay() != tt.expected {
				t.Errorf("GetTxDelay: got %v, want %v", cfg.GetTxDelay(), tt.expected)
			}
		})
	}
}

func TestConfig_GetUptime(t *testing.T) {
	startTime := time.Now().Add(-5 * time.Second)
	cfg := &Config{StartTime: startTime}

	uptime := cfg.GetUptime()
	if uptime < 5*time.Second {
		t.Errorf("GetUptime: got %v, want at least 5s", uptime)
	}
}

func TestLoadConfig_Integration(t *testing.T) {
	os.Setenv("AGENT_PORT", "8092")
	os.Setenv("SAGE_ENABLED", "true")
	os.Setenv("TX_DELAY_MS", "300")
	defer os.Clearenv()

	cfg := LoadConfig()

	// Verify all fields are populated
	if cfg.AgentPort == "" {
		t.Error("AgentPort should not be empty")
	}

	if cfg.AgentVersion == "" {
		t.Error("AgentVersion should not be empty")
	}

	if cfg.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}
}
