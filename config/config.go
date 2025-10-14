package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the payment agent
type Config struct {
	// Server configuration
	AgentPort string
	LogLevel  string

	// SAGE Protocol configuration
	SAGEEnabled    bool
	StrictMode     bool
	BlockchainRPC  string
	ContractAddress string

	// Transaction configuration
	SimulationMode bool
	TxDelayMs      int
	WalletAddress  string

	// Agent metadata
	AgentVersion string
	StartTime    time.Time
}

// LoadConfig reads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		// Server
		AgentPort: getEnv("AGENT_PORT", "8091"),
		LogLevel:  getEnv("LOG_LEVEL", "info"),

		// SAGE Protocol
		SAGEEnabled:     getEnvBool("SAGE_ENABLED", true),
		StrictMode:      getEnvBool("SAGE_STRICT_MODE", true),
		BlockchainRPC:   getEnv("BLOCKCHAIN_RPC_URL", "http://localhost:8545"),
		ContractAddress: getEnv("CONTRACT_ADDRESS", "0x0000000000000000000000000000000000000000"),

		// Transaction
		SimulationMode: getEnvBool("TX_SIMULATION_MODE", true),
		TxDelayMs:      getEnvInt("TX_DELAY_MS", 500),
		WalletAddress:  getEnv("WALLET_ADDRESS", "0x1234567890abcdef1234567890abcdef12345678"),

		// Metadata
		AgentVersion: "1.0.0",
		StartTime:    time.Now(),
	}
}

// IsSAGEEnabled returns whether SAGE protocol is enabled
func (c *Config) IsSAGEEnabled() bool {
	return c.SAGEEnabled
}

// IsStrictMode returns whether strict mode is enabled
func (c *Config) IsStrictMode() bool {
	return c.StrictMode
}

// IsSimulationMode returns whether transaction simulation is enabled
func (c *Config) IsSimulationMode() bool {
	return c.SimulationMode
}

// GetTxDelay returns transaction delay as Duration
func (c *Config) GetTxDelay() time.Duration {
	return time.Duration(c.TxDelayMs) * time.Millisecond
}

// GetUptime returns uptime since start
func (c *Config) GetUptime() time.Duration {
	return time.Since(c.StartTime)
}

// getEnv retrieves environment variable or returns default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool retrieves boolean environment variable or returns default
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			return parsed
		}
		// Try "1" and "0"
		if value == "1" {
			return true
		}
		if value == "0" {
			return false
		}
	}
	return defaultValue
}

// getEnvInt retrieves integer environment variable or returns default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}
