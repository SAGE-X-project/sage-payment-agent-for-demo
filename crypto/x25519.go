package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/curve25519"
)

// X25519KeyPair represents an X25519 key pair for HPKE
type X25519KeyPair struct {
	PrivateKey []byte // 32 bytes
	PublicKey  []byte // 32 bytes
	keyPath    string
}

// GenerateX25519KeyPair generates a new X25519 key pair
func GenerateX25519KeyPair() (*X25519KeyPair, error) {
	privateKey := make([]byte, 32)
	if _, err := rand.Read(privateKey); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	// Generate public key from private key
	publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("failed to derive public key: %w", err)
	}

	return &X25519KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

// LoadX25519KeyPair loads X25519 key pair from file
func LoadX25519KeyPair(path string) (*X25519KeyPair, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	// Decode hex
	privateKey, err := hex.DecodeString(string(keyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}

	if len(privateKey) != 32 {
		return nil, errors.New("invalid X25519 private key length")
	}

	// Derive public key
	publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("failed to derive public key: %w", err)
	}

	return &X25519KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		keyPath:    path,
	}, nil
}

// SaveX25519KeyPair saves X25519 key pair to file
func (k *X25519KeyPair) Save(path string) error {
	if len(k.PrivateKey) != 32 {
		return errors.New("invalid private key")
	}

	// Create directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Encode to hex
	privateKeyHex := hex.EncodeToString(k.PrivateKey)

	// Write to file
	if err := os.WriteFile(path, []byte(privateKeyHex), 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	k.keyPath = path
	return nil
}

// GetPublicKeyHex returns public key as hex string
func (k *X25519KeyPair) GetPublicKeyHex() string {
	return hex.EncodeToString(k.PublicKey)
}

// GetPrivateKeyHex returns private key as hex string
func (k *X25519KeyPair) GetPrivateKeyHex() string {
	return hex.EncodeToString(k.PrivateKey)
}

// NewX25519KeyManager creates or loads X25519 key for HPKE
func NewX25519KeyManager(keyPath string) (*X25519KeyPair, error) {
	// Check if key exists
	statInfo, statErr := os.Stat(keyPath)
	if statErr == nil {
		// Load existing key
		keypair, err := LoadX25519KeyPair(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load X25519 key: %w", err)
		}
		fmt.Printf("✓ Loaded existing X25519 key from: %s\n", keyPath)
		return keypair, nil
	} else if os.IsNotExist(statErr) {
		// Generate new key
		keypair, err := GenerateX25519KeyPair()
		if err != nil {
			return nil, fmt.Errorf("failed to generate X25519 key: %w", err)
		}

		// Save key
		if err := keypair.Save(keyPath); err != nil {
			return nil, fmt.Errorf("failed to save X25519 key: %w", err)
		}

		fmt.Printf("✓ Generated and saved new X25519 key to: %s\n", keyPath)
		return keypair, nil
	}

	// If stat error is not IsNotExist, return error
	_ = statInfo // Avoid unused variable warning
	return nil, fmt.Errorf("failed to check key file: %w", statErr)
}
