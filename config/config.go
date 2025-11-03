package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the payment agent
type Config struct {
	// Server configuration
	AgentPort      string
	AgentHost      string
	AgentPublicURL string // Public URL for agent discovery and contract registration
	LogLevel       string

	// SAGE Protocol configuration
	SAGEEnabled    bool
	StrictMode     bool
	BlockchainRPC  string
	ContractAddress string
	ChainID        int

	// Transaction configuration
	SimulationMode bool
	TxDelayMs      int
	WalletAddress  string

	// Agent identity & keys
	AgentDID          string
	PrivateKeyPath    string
	AgentPrivateKey   string
	KeystorePath      string
	KeystorePassword  string
	AgentJWKFile      string
	AgentKEMJWKFile   string

	// Agent registry
	AutoRegister      bool
	AgentName         string
	AgentDescription  string
	AgentCapabilities string

	// Agent metadata
	AgentVersion string
	StartTime    time.Time
}

// LoadConfig reads configuration from environment variables
func LoadConfig() *Config {
	agentPort := getEnv("AGENT_PORT", "8091")
	agentHost := getEnv("AGENT_HOST", "0.0.0.0")

	// Generate default public URL if not provided
	defaultPublicURL := "http://localhost:" + agentPort
	if agentHost != "0.0.0.0" && agentHost != "localhost" {
		defaultPublicURL = "http://" + agentHost + ":" + agentPort
	}

	return &Config{
		// Server
		AgentPort:      agentPort,
		AgentHost:      agentHost,
		AgentPublicURL: getEnv("AGENT_PUBLIC_URL", defaultPublicURL),
		LogLevel:       getEnv("LOG_LEVEL", "info"),

		// SAGE Protocol
		SAGEEnabled:     getEnvBool("SAGE_ENABLED", true),
		StrictMode:      getEnvBool("SAGE_STRICT_MODE", true),
		BlockchainRPC:   getEnv("BLOCKCHAIN_RPC_URL", "http://localhost:8545"),
		ContractAddress: getEnv("CONTRACT_ADDRESS", "0x0000000000000000000000000000000000000000"),
		ChainID:         getEnvInt("CHAIN_ID", 31337), // 31337 for local, 11155111 for sepolia

		// Transaction
		SimulationMode: getEnvBool("TX_SIMULATION_MODE", true),
		TxDelayMs:      getEnvInt("TX_DELAY_MS", 500),
		WalletAddress:  getEnv("WALLET_ADDRESS", "0x1234567890abcdef1234567890abcdef12345678"),

		// Agent identity & keys
		AgentDID:         getEnv("AGENT_DID", ""),
		PrivateKeyPath:   getEnv("PRIVATE_KEY_PATH", "./keys/agent.key"),
		AgentPrivateKey:  getEnv("AGENT_PRIVATE_KEY", ""),
		KeystorePath:     getEnv("KEYSTORE_PATH", ""),
		KeystorePassword: getEnv("KEYSTORE_PASSWORD", ""),
		AgentJWKFile:     getEnv("AGENT_JWK_FILE", "./keys/payment.jwk"),
		AgentKEMJWKFile:  getEnv("AGENT_KEM_JWK_FILE", "./keys/kem/payment.x25519.jwk"),

		// Agent registry
		AutoRegister:      getEnvBool("AUTO_REGISTER", true),
		AgentName:         getEnv("AGENT_NAME", "Official Payment Agent"),
		AgentDescription:  getEnv("AGENT_DESCRIPTION", "Blockchain-based secure payment processing agent"),
		AgentCapabilities: getEnv("AGENT_CAPABILITIES", "crypto_payment,stablecoin,rfc9421_signature,hpke_encryption"),

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
