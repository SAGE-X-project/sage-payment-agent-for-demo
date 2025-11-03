package crypto

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// KeyManager manages agent's private keys
type KeyManager struct {
	privateKey     *ecdsa.PrivateKey
	publicKeyHex   string
	address        string
	keyStorePath   string
	agentDID       string
}

// NewKeyManager creates a new key manager
// If privateKeyPath exists, load existing key
// Otherwise, generate new key and save it
func NewKeyManager(privateKeyPath, agentDID string) (*KeyManager, error) {
	km := &KeyManager{
		keyStorePath: privateKeyPath,
		agentDID:     agentDID,
	}

	// Check if key file exists
	if _, err := os.Stat(privateKeyPath); err == nil {
		// Load existing key
		if err := km.LoadKey(privateKeyPath); err != nil {
			return nil, fmt.Errorf("failed to load existing key: %w", err)
		}
		fmt.Printf("✓ Loaded existing private key from: %s\n", privateKeyPath)
	} else if os.IsNotExist(err) {
		// Generate new key
		if err := km.GenerateNewKey(); err != nil {
			return nil, fmt.Errorf("failed to generate new key: %w", err)
		}

		// Save key
		if err := km.SaveKey(privateKeyPath); err != nil {
			return nil, fmt.Errorf("failed to save key: %w", err)
		}
		fmt.Printf("✓ Generated and saved new private key to: %s\n", privateKeyPath)
	} else {
		return nil, fmt.Errorf("failed to check key file: %w", err)
	}

	return km, nil
}

// ImportKeyFromHex imports a private key from hex string
func ImportKeyFromHex(privateKeyHex, savePath, agentDID string) (*KeyManager, error) {
	km := &KeyManager{
		keyStorePath: savePath,
		agentDID:     agentDID,
	}

	// Remove 0x prefix if exists
	if len(privateKeyHex) > 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	// Decode hex to bytes
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid hex private key: %w", err)
	}

	// Convert to ECDSA private key
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid ECDSA private key: %w", err)
	}

	km.privateKey = privateKey
	km.publicKeyHex = hex.EncodeToString(crypto.FromECDSAPub(&privateKey.PublicKey))
	km.address = crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// Save imported key
	if err := km.SaveKey(savePath); err != nil {
		return nil, fmt.Errorf("failed to save imported key: %w", err)
	}

	fmt.Printf("✓ Imported private key and saved to: %s\n", savePath)
	return km, nil
}

// GenerateNewKey generates a new ECDSA private key
func (km *KeyManager) GenerateNewKey() error {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	km.privateKey = privateKey
	km.publicKeyHex = hex.EncodeToString(crypto.FromECDSAPub(&privateKey.PublicKey))
	km.address = crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	return nil
}

// LoadKey loads private key from file
func (km *KeyManager) LoadKey(path string) error {
	// Read key file
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	// Decode hex
	privateKeyBytes, err := hex.DecodeString(string(keyBytes))
	if err != nil {
		return fmt.Errorf("failed to decode key: %w", err)
	}

	// Convert to ECDSA
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse key: %w", err)
	}

	km.privateKey = privateKey
	km.publicKeyHex = hex.EncodeToString(crypto.FromECDSAPub(&privateKey.PublicKey))
	km.address = crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	return nil
}

// SaveKey saves private key to file (hex encoded)
func (km *KeyManager) SaveKey(path string) error {
	if km.privateKey == nil {
		return errors.New("no private key to save")
	}

	// Create directory if not exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Convert key to bytes
	privateKeyBytes := crypto.FromECDSA(km.privateKey)

	// Encode to hex
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	// Write to file with restricted permissions
	if err := os.WriteFile(path, []byte(privateKeyHex), 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}

// GetPrivateKey returns the private key
func (km *KeyManager) GetPrivateKey() *ecdsa.PrivateKey {
	return km.privateKey
}

// GetPublicKeyHex returns public key as hex string
func (km *KeyManager) GetPublicKeyHex() string {
	return km.publicKeyHex
}

// GetAddress returns Ethereum address
func (km *KeyManager) GetAddress() string {
	return km.address
}

// GetDID returns agent DID
func (km *KeyManager) GetDID() string {
	if km.agentDID != "" {
		return km.agentDID
	}
	// Generate DID from address if not set
	return fmt.Sprintf("did:sage:ethereum:%s", km.address)
}

// SignMessage signs a message with the private key
func (km *KeyManager) SignMessage(message []byte) ([]byte, error) {
	if km.privateKey == nil {
		return nil, errors.New("no private key available")
	}

	// Hash the message
	hash := crypto.Keccak256Hash(message)

	// Sign
	signature, err := crypto.Sign(hash.Bytes(), km.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	return signature, nil
}

// SignHash signs a hash with the private key
func (km *KeyManager) SignHash(hash []byte) ([]byte, error) {
	if km.privateKey == nil {
		return nil, errors.New("no private key available")
	}

	signature, err := crypto.Sign(hash, km.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign hash: %w", err)
	}

	return signature, nil
}

// GenerateOwnershipProof generates ECDSA signature for X25519 key ownership
// This proves that the agent owns the X25519 key for HPKE
// Message format matches Solidity: keccak256(abi.encodePacked("SAGE X25519 Ownership:", keyData, chainid, contract, owner))
func (km *KeyManager) GenerateOwnershipProof(x25519PublicKey []byte, chainID int64, contractAddress string) ([]byte, error) {
	if len(x25519PublicKey) != 32 {
		return nil, errors.New("invalid X25519 public key length (expected 32 bytes)")
	}

	// Build message using abi.encodePacked format
	// "SAGE X25519 Ownership:" (string - raw bytes)
	// keyData (bytes - 32 bytes X25519 public key)
	// chainID (uint256 - 32 bytes big-endian)
	// contractAddress (address - 20 bytes)
	// owner (address - 20 bytes)

	message := []byte("SAGE X25519 Ownership:")

	// Append X25519 public key (32 bytes)
	message = append(message, x25519PublicKey...)

	// Encode chainID as uint256 (32 bytes, big-endian)
	chainIDBytes := make([]byte, 32)
	chainIDBig := big.NewInt(chainID)
	chainIDBig.FillBytes(chainIDBytes)
	message = append(message, chainIDBytes...)

	// Encode contract address (20 bytes)
	contractAddr := common.HexToAddress(contractAddress)
	message = append(message, contractAddr.Bytes()...)

	// Encode owner address (20 bytes)
	ownerAddr := common.HexToAddress(km.address)
	message = append(message, ownerAddr.Bytes()...)

	// Hash the message (this is messageHash in Solidity)
	messageHash := crypto.Keccak256Hash(message)

	// Add Ethereum signed message prefix: "\x19Ethereum Signed Message:\n32" + messageHash
	ethSignedMessage := append([]byte("\x19Ethereum Signed Message:\n32"), messageHash.Bytes()...)
	ethSignedHash := crypto.Keccak256Hash(ethSignedMessage)

	// Sign with private key
	signature, err := crypto.Sign(ethSignedHash.Bytes(), km.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign ownership proof: %w", err)
	}

	return signature, nil
}

// GenerateRegistrationSignature generates a signature for agent registration
// Message format matches Solidity: keccak256(abi.encodePacked("SAGE Agent Registration:", chainid, contract, owner))
func (km *KeyManager) GenerateRegistrationSignature(chainID int64, contractAddress string) ([]byte, error) {
	// Build message using abi.encodePacked format
	// "SAGE Agent Registration:" (string - raw bytes)
	// chainID (uint256 - 32 bytes big-endian)
	// contractAddress (address - 20 bytes)
	// owner (address - 20 bytes)

	message := []byte("SAGE Agent Registration:")

	// Encode chainID as uint256 (32 bytes, big-endian)
	chainIDBytes := make([]byte, 32)
	chainIDBig := big.NewInt(chainID)
	chainIDBig.FillBytes(chainIDBytes)
	message = append(message, chainIDBytes...)

	// Encode contract address (20 bytes)
	contractAddr := common.HexToAddress(contractAddress)
	message = append(message, contractAddr.Bytes()...)

	// Encode owner address (20 bytes)
	ownerAddr := common.HexToAddress(km.address)
	message = append(message, ownerAddr.Bytes()...)

	// Hash the message (this is messageHash in Solidity)
	messageHash := crypto.Keccak256Hash(message)

	// Add Ethereum signed message prefix: "\x19Ethereum Signed Message:\n32" + messageHash
	ethSignedMessage := append([]byte("\x19Ethereum Signed Message:\n32"), messageHash.Bytes()...)
	ethSignedHash := crypto.Keccak256Hash(ethSignedMessage)

	// Sign with private key
	signature, err := crypto.Sign(ethSignedHash.Bytes(), km.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign registration message: %w", err)
	}

	return signature, nil
}
