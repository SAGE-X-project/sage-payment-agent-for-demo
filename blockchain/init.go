package blockchain

import (
	"fmt"
	"os"

	"github.com/sage-x-project/sage-payment-agent-for-demo/config"
	"github.com/sage-x-project/sage-payment-agent-for-demo/crypto"
)

// AgentInitializer handles agent startup and registration
type AgentInitializer struct {
	Config         *config.Config
	KeyManager     *crypto.KeyManager
	X25519Key      *crypto.X25519KeyPair
	RegistryClient *RegistryClient
}

// NewAgentInitializer creates a new agent initializer
func NewAgentInitializer(cfg *config.Config) (*AgentInitializer, error) {
	init := &AgentInitializer{
		Config: cfg,
	}

	// Step 1: Load or generate ECDSA key
	if err := init.initializeKeys(); err != nil {
		return nil, fmt.Errorf("failed to initialize keys: %w", err)
	}

	// Step 2: Connect to blockchain (if enabled)
	if cfg.BlockchainRPC != "" && cfg.ContractAddress != "0x0000000000000000000000000000000000000000" {
		if err := init.connectBlockchain(); err != nil {
			fmt.Printf("⚠️  Warning: Failed to connect to blockchain: %v\n", err)
			fmt.Println("   Agent will run without blockchain integration")
		}
	} else {
		fmt.Println("ℹ️  Blockchain integration disabled (no RPC URL or contract address)")
	}

	return init, nil
}

// initializeKeys loads or generates agent keys
func (init *AgentInitializer) initializeKeys() error {
	fmt.Println("\n🔑 Initializing Agent Keys...")

	// Option 1: Import from hex (if provided in env)
	if init.Config.AgentPrivateKey != "" {
		fmt.Println("  📥 Importing private key from environment variable...")
		km, err := crypto.ImportKeyFromHex(
			init.Config.AgentPrivateKey,
			init.Config.PrivateKeyPath,
			init.Config.AgentDID,
		)
		if err != nil {
			return fmt.Errorf("failed to import private key: %w", err)
		}
		init.KeyManager = km
	} else {
		// Option 2: Load existing or generate new
		km, err := crypto.NewKeyManager(init.Config.PrivateKeyPath, init.Config.AgentDID)
		if err != nil {
			return fmt.Errorf("failed to initialize key manager: %w", err)
		}
		init.KeyManager = km
	}

	fmt.Printf("  ✓ Agent Address: %s\n", init.KeyManager.GetAddress())
	fmt.Printf("  ✓ Agent DID: %s\n", init.KeyManager.GetDID())

	// Load or generate X25519 key for HPKE
	x25519Key, err := crypto.NewX25519KeyManager(init.Config.AgentKEMJWKFile)
	if err != nil {
		return fmt.Errorf("failed to initialize X25519 key: %w", err)
	}
	init.X25519Key = x25519Key

	fmt.Printf("  ✓ X25519 Public Key: %s\n", x25519Key.GetPublicKeyHex()[:16]+"...")

	return nil
}

// connectBlockchain connects to blockchain and initializes registry client
func (init *AgentInitializer) connectBlockchain() error {
	fmt.Println("\n🔗 Connecting to Blockchain...")
	fmt.Printf("  RPC URL: %s\n", init.Config.BlockchainRPC)
	fmt.Printf("  Contract: %s\n", init.Config.ContractAddress)
	fmt.Printf("  Chain ID: %d\n", init.Config.ChainID)

	registryClient, err := NewRegistryClient(
		init.Config.BlockchainRPC,
		init.Config.ContractAddress,
		int64(init.Config.ChainID),
		init.KeyManager,
		init.X25519Key,
	)
	if err != nil {
		return err
	}

	init.RegistryClient = registryClient
	fmt.Println("  ✓ Connected to blockchain successfully")

	return nil
}

// RegisterIfNeeded registers agent on blockchain if AUTO_REGISTER is enabled
func (init *AgentInitializer) RegisterIfNeeded() error {
	if !init.Config.AutoRegister {
		fmt.Println("\nℹ️  Auto-registration disabled")
		return nil
	}

	if init.RegistryClient == nil {
		fmt.Println("\n⚠️  Cannot register: no blockchain connection")
		return nil
	}

	// Check if already registered
	isRegistered, err := init.RegistryClient.IsRegistered()
	if err != nil {
		return fmt.Errorf("failed to check registration status: %w", err)
	}

	if isRegistered {
		fmt.Println("\n✓ Agent already registered on blockchain")
		return nil
	}

	// Register agent
	return init.RegistryClient.RegisterAgent(
		init.Config.AgentName,
		init.Config.AgentDescription,
		init.Config.AgentPublicURL,
		init.Config.AgentCapabilities,
	)
}

// GetAgentInfo returns formatted agent information
func (init *AgentInitializer) GetAgentInfo() string {
	info := "\n"
	info += "═══════════════════════════════════════════════════════\n"
	info += " SAGE Payment Agent - Identity Information\n"
	info += "═══════════════════════════════════════════════════════\n"
	info += fmt.Sprintf("  Name:         %s\n", init.Config.AgentName)
	info += fmt.Sprintf("  Version:      %s\n", init.Config.AgentVersion)
	info += fmt.Sprintf("  DID:          %s\n", init.KeyManager.GetDID())
	info += fmt.Sprintf("  Address:      %s\n", init.KeyManager.GetAddress())
	info += fmt.Sprintf("  Public URL:   %s\n", init.Config.AgentPublicURL)
	info += fmt.Sprintf("  Capabilities: %s\n", init.Config.AgentCapabilities)
	info += "═══════════════════════════════════════════════════════\n"
	return info
}

// Cleanup performs cleanup operations
func (init *AgentInitializer) Cleanup() {
	if init.RegistryClient != nil {
		init.RegistryClient.Close()
	}
}

// CheckEnvironmentSetup validates environment configuration
func CheckEnvironmentSetup() error {
	requiredEnvVars := []string{
		"AGENT_PORT",
	}

	missing := []string{}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}

	return nil
}
