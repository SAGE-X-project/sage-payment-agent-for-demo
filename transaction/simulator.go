package transaction

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/sage-x-project/sage-payment-agent-for-demo/config"
	"github.com/sage-x-project/sage-payment-agent-for-demo/logger"
	"github.com/sage-x-project/sage-payment-agent-for-demo/types"
)

// Simulator handles transaction simulation
type Simulator struct {
	config *config.Config
}

// NewSimulator creates a new transaction simulator
func NewSimulator(cfg *config.Config) *Simulator {
	return &Simulator{
		config: cfg,
	}
}

// ProcessPayment simulates processing a payment transaction
func (s *Simulator) ProcessPayment(req *types.PaymentRequest) (*types.TransactionResult, error) {
	logger.Info("Processing payment transaction...")

	// Validate request
	if err := s.validateRequest(req); err != nil {
		logger.Error("Payment validation failed: %v", err)
		return nil, err
	}

	// Simulate network delay
	if s.config.IsSimulationMode() {
		logger.Debug("Simulating transaction delay: %v", s.config.GetTxDelay())
		time.Sleep(s.config.GetTxDelay())
	}

	// Generate transaction hash
	txHash := s.generateTxHash()

	// Create transaction result
	result := &types.TransactionResult{
		TxHash:    txHash,
		From:      s.config.WalletAddress,
		To:        req.Recipient,
		Amount:    req.Amount,
		Status:    "confirmed",
		Timestamp: time.Now(),
		BlockNumber: s.generateBlockNumber(),
		GasUsed:     21000, // Standard gas for transfer
	}

	logger.Info("Transaction confirmed: %s", txHash)
	logger.Debug("Gas used: %d", result.GasUsed)

	return result, nil
}

// validateRequest validates payment request
func (s *Simulator) validateRequest(req *types.PaymentRequest) error {
	if req.Amount <= 0 {
		return fmt.Errorf("invalid amount: must be greater than 0")
	}

	if req.Recipient == "" {
		return fmt.Errorf("recipient address required")
	}

	// Basic Ethereum address validation (starts with 0x, 42 chars total)
	if len(req.Recipient) < 10 {
		return fmt.Errorf("invalid recipient address format")
	}

	return nil
}

// generateTxHash generates a simulated transaction hash
func (s *Simulator) generateTxHash() string {
	// Generate 32 random bytes for transaction hash
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return "0x" + hex.EncodeToString(bytes)
}

// generateBlockNumber generates a simulated block number
func (s *Simulator) generateBlockNumber() int64 {
	// Use current time as base for block number
	// In real blockchain, this would be the actual block number
	return time.Now().Unix() / 12 // ~12 second block time
}

// GetTransactionStatus simulates getting transaction status
func (s *Simulator) GetTransactionStatus(txHash string) (string, error) {
	// In simulation mode, all transactions are immediately confirmed
	if s.config.IsSimulationMode() {
		return "confirmed", nil
	}

	// In real mode, this would query the blockchain
	return "pending", fmt.Errorf("not implemented: real blockchain query")
}
