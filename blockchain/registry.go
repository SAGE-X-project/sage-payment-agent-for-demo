package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/sage-x-project/sage-payment-agent-for-demo/blockchain/bindings"
	"github.com/sage-x-project/sage-payment-agent-for-demo/crypto"
)

// RegistryClient manages agent registration on blockchain
type RegistryClient struct {
	client          *ethclient.Client
	contractAddress common.Address
	chainID         *big.Int
	keyManager      *crypto.KeyManager
	x25519Key       *crypto.X25519KeyPair
}

// NewRegistryClient creates a new registry client
func NewRegistryClient(
	rpcURL string,
	contractAddr string,
	chainID int64,
	keyManager *crypto.KeyManager,
	x25519Key *crypto.X25519KeyPair,
) (*RegistryClient, error) {
	// Connect to Ethereum node
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	actualChainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	if actualChainID.Int64() != chainID {
		return nil, fmt.Errorf("chain ID mismatch: expected %d, got %d", chainID, actualChainID.Int64())
	}

	return &RegistryClient{
		client:          client,
		contractAddress: common.HexToAddress(contractAddr),
		chainID:         big.NewInt(chainID),
		keyManager:      keyManager,
		x25519Key:       x25519Key,
	}, nil
}

// RegisterAgent registers agent on blockchain
// Step 1: Commit registration
// Step 2: Wait for minimum delay
// Step 3: Reveal and register
func (rc *RegistryClient) RegisterAgent(
	name string,
	description string,
	endpoint string,
	capabilities string,
) error {
	fmt.Println("\n🔗 Starting agent registration on blockchain...")

	// Step 1: Create commitment
	did := rc.keyManager.GetDID()
	saltHash := ethcrypto.Keccak256Hash([]byte(fmt.Sprintf("%s-%d", did, time.Now().UnixNano())))
	var salt [32]byte
	copy(salt[:], saltHash.Bytes())

	// Prepare keys for registration
	ecdsaPubKey := ethcrypto.FromECDSAPub(&rc.keyManager.GetPrivateKey().PublicKey)
	keys := [][]byte{ecdsaPubKey, rc.x25519Key.PublicKey}

	// Create commitment hash using ABI encoding (matching Solidity's abi.encode)
	// Contract expects: keccak256(abi.encode(params.did, params.keys, msg.sender, params.salt, block.chainid))
	bytes32Type, _ := abi.NewType("bytes32", "", nil)
	addressType, _ := abi.NewType("address", "", nil)
	stringType, _ := abi.NewType("string", "", nil)
	bytesArrayType, _ := abi.NewType("bytes[]", "", nil)
	uint256Type, _ := abi.NewType("uint256", "", nil)

	arguments := abi.Arguments{
		{Type: stringType},     // did
		{Type: bytesArrayType}, // keys
		{Type: addressType},    // msg.sender
		{Type: bytes32Type},    // salt
		{Type: uint256Type},    // chainid
	}

	agentAddress := common.HexToAddress(rc.keyManager.GetAddress())
	encoded, err := arguments.Pack(did, keys, agentAddress, salt, rc.chainID)
	if err != nil {
		return fmt.Errorf("failed to encode commitment data: %w", err)
	}

	commitHash := ethcrypto.Keccak256Hash(encoded)
	fmt.Printf("  📝 Commitment hash: %s\n", commitHash.Hex())

	// Create contract instance
	registry, err := bindings.NewAgentCardRegistry(rc.contractAddress, rc.client)
	if err != nil {
		return fmt.Errorf("failed to create contract instance: %w", err)
	}

	// Step 2: Send commit transaction
	auth, err := rc.createTransactor()
	if err != nil {
		return fmt.Errorf("failed to create transactor: %w", err)
	}

	// Set registration stake (0.01 ETH)
	auth.Value = big.NewInt(10000000000000000) // 0.01 ETH

	fmt.Println("  💰 Sending commitment transaction...")

	// Call CommitRegistration on the contract
	var commitHashBytes [32]byte
	copy(commitHashBytes[:], commitHash.Bytes())

	commitTx, err := registry.CommitRegistration(auth, commitHashBytes)
	if err != nil {
		return fmt.Errorf("failed to commit registration: %w", err)
	}

	fmt.Printf("  ✓ Commitment sent (tx: %s, stake: %s ETH)\n", commitTx.Hash().Hex(), auth.Value.String())

	// Wait for commit transaction to be mined
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	commitReceipt, err := bind.WaitMined(ctx, rc.client, commitTx)
	if err != nil {
		return fmt.Errorf("failed to wait for commit tx: %w", err)
	}
	if commitReceipt.Status != 1 {
		return fmt.Errorf("commit transaction failed")
	}
	fmt.Println("  ✓ Commitment transaction mined")

	// Step 3: Wait for minimum delay (shortened for testing)
	fmt.Println("  ⏳ Waiting for commit-reveal delay...")
	time.Sleep(2 * time.Second)

	// Step 4: Reveal and register
	fmt.Println("  📤 Revealing and registering agent...")

	// Generate signatures for each key
	// 1. ECDSA key signature (for key ownership verification)
	ecdsaSignature, err := rc.keyManager.GenerateRegistrationSignature(
		rc.chainID.Int64(),
		rc.contractAddress.Hex(),
	)
	if err != nil {
		return fmt.Errorf("failed to generate ECDSA signature: %w", err)
	}

	// 2. X25519 key ownership proof (ECDSA signature proving ownership of X25519 key)
	x25519Signature, err := rc.keyManager.GenerateOwnershipProof(
		rc.x25519Key.PublicKey,
		rc.chainID.Int64(),
		rc.contractAddress.Hex(),
	)
	if err != nil {
		return fmt.Errorf("failed to generate X25519 ownership proof: %w", err)
	}

	// Prepare registration parameters
	params := bindings.AgentCardStorageRegistrationParams{
		Did:          did,
		Name:         name,
		Description:  description,
		Endpoint:     endpoint,
		Capabilities: capabilities,
		Keys:         keys, // [ECDSA public key, X25519 public key]
		KeyTypes:     []uint8{0, 2}, // 0=ECDSA, 2=X25519
		Signatures:   [][]byte{ecdsaSignature, x25519Signature}, // One signature per key
		Salt:         salt, // [32]byte salt
	}

	// Create new transactor for register call (no value this time)
	auth2, err := rc.createTransactor()
	if err != nil {
		return fmt.Errorf("failed to create transactor for register: %w", err)
	}
	auth2.Value = nil // No ETH sent with register call

	// Call RegisterAgentWithParams on the contract
	registerTx, err := registry.RegisterAgentWithParams(auth2, params)
	if err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}

	fmt.Printf("  ✓ Registration transaction sent (tx: %s)\n", registerTx.Hash().Hex())

	// Wait for register transaction to be mined
	ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel2()

	registerReceipt, err := bind.WaitMined(ctx2, rc.client, registerTx)
	if err != nil {
		return fmt.Errorf("failed to wait for register tx: %w", err)
	}
	if registerReceipt.Status != 1 {
		return fmt.Errorf("register transaction failed")
	}

	fmt.Printf("  ✓ Agent registered successfully!\n")
	fmt.Printf("    DID: %s\n", did)
	fmt.Printf("    Address: %s\n", rc.keyManager.GetAddress())
	fmt.Printf("    Endpoint: %s\n", endpoint)

	// Store registration info for later use
	fmt.Println("\n  📋 Registration Details:")
	fmt.Printf("    Name: %s\n", name)
	fmt.Printf("    Description: %s\n", description)
	fmt.Printf("    Capabilities: %s\n", capabilities)
	fmt.Printf("    Keys: %d (ECDSA + X25519)\n", len(keys))
	fmt.Printf("    Signatures: %d (one per key)\n", len(params.Signatures))
	fmt.Printf("    Commit Tx: %s\n", commitTx.Hash().Hex())
	fmt.Printf("    Register Tx: %s\n", registerTx.Hash().Hex())

	return nil
}

// createTransactor creates an auth transactor for sending transactions
func (rc *RegistryClient) createTransactor() (*bind.TransactOpts, error) {
	privateKey := rc.keyManager.GetPrivateKey()

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, rc.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	// Get nonce
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	agentAddress := common.HexToAddress(rc.keyManager.GetAddress())
	nonce, err := rc.client.PendingNonceAt(ctx, agentAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasLimit = uint64(3000000) // 3M gas limit

	// Get suggested gas price
	gasPrice, err := rc.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}
	auth.GasPrice = gasPrice

	return auth, nil
}

// IsRegistered checks if agent is already registered
func (rc *RegistryClient) IsRegistered() (bool, error) {
	// In production, query the contract
	// For MVP, return false
	return false, nil
}

// Close closes the Ethereum client connection
func (rc *RegistryClient) Close() {
	if rc.client != nil {
		rc.client.Close()
	}
}
