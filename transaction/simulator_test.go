package transaction

import (
	"testing"
	"time"

	"github.com/sage-x-project/sage-payment-agent-for-demo/config"
	"github.com/sage-x-project/sage-payment-agent-for-demo/types"
)

func TestNewSimulator(t *testing.T) {
	cfg := &config.Config{}
	sim := NewSimulator(cfg)

	if sim == nil {
		t.Fatal("NewSimulator() returned nil")
	}

	if sim.config != cfg {
		t.Error("NewSimulator() didn't set config properly")
	}
}

func TestProcessPayment_Success(t *testing.T) {
	cfg := &config.Config{
		SimulationMode: true,
		TxDelayMs:      10, // Short delay for tests
		WalletAddress:  "0xAGENT",
	}

	sim := NewSimulator(cfg)

	req := &types.PaymentRequest{
		Amount:    100.0,
		Currency:  "USDC",
		Product:   "Sunglasses",
		Recipient: "0x742d35Cc",
	}

	start := time.Now()
	result, err := sim.ProcessPayment(req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ProcessPayment() failed: %v", err)
	}

	if result == nil {
		t.Fatal("ProcessPayment() returned nil result")
	}

	// Check transaction result fields
	if result.TxHash == "" {
		t.Error("TxHash should not be empty")
	}

	if result.From != "0xAGENT" {
		t.Errorf("From: got %s, want 0xAGENT", result.From)
	}

	if result.To != "0x742d35Cc" {
		t.Errorf("To: got %s, want 0x742d35Cc", result.To)
	}

	if result.Amount != 100.0 {
		t.Errorf("Amount: got %f, want 100.0", result.Amount)
	}

	if result.Status != "confirmed" {
		t.Errorf("Status: got %s, want confirmed", result.Status)
	}

	if result.GasUsed != 21000 {
		t.Errorf("GasUsed: got %d, want 21000", result.GasUsed)
	}

	if result.BlockNumber == 0 {
		t.Error("BlockNumber should not be 0")
	}

	// Check delay was applied
	if elapsed < 10*time.Millisecond {
		t.Error("Transaction delay was not applied")
	}
}

func TestProcessPayment_InvalidAmount(t *testing.T) {
	cfg := &config.Config{
		SimulationMode: true,
		TxDelayMs:      0,
	}

	sim := NewSimulator(cfg)

	tests := []struct {
		name   string
		amount float64
	}{
		{"Zero amount", 0.0},
		{"Negative amount", -100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &types.PaymentRequest{
				Amount:    tt.amount,
				Recipient: "0x742d35Cc",
			}

			_, err := sim.ProcessPayment(req)
			if err == nil {
				t.Error("ProcessPayment() should return error for invalid amount")
			}
		})
	}
}

func TestProcessPayment_EmptyRecipient(t *testing.T) {
	cfg := &config.Config{
		SimulationMode: true,
		TxDelayMs:      0,
	}

	sim := NewSimulator(cfg)

	req := &types.PaymentRequest{
		Amount:    100.0,
		Recipient: "",
	}

	_, err := sim.ProcessPayment(req)
	if err == nil {
		t.Error("ProcessPayment() should return error for empty recipient")
	}
}

func TestProcessPayment_InvalidRecipientFormat(t *testing.T) {
	cfg := &config.Config{
		SimulationMode: true,
		TxDelayMs:      0,
	}

	sim := NewSimulator(cfg)

	req := &types.PaymentRequest{
		Amount:    100.0,
		Recipient: "abc", // Too short
	}

	_, err := sim.ProcessPayment(req)
	if err == nil {
		t.Error("ProcessPayment() should return error for invalid recipient format")
	}
}

func TestValidateRequest(t *testing.T) {
	cfg := &config.Config{}
	sim := NewSimulator(cfg)

	tests := []struct {
		name      string
		req       *types.PaymentRequest
		shouldErr bool
	}{
		{
			"Valid request",
			&types.PaymentRequest{Amount: 100.0, Recipient: "0x742d35Cc"},
			false,
		},
		{
			"Zero amount",
			&types.PaymentRequest{Amount: 0.0, Recipient: "0x742d35Cc"},
			true,
		},
		{
			"Negative amount",
			&types.PaymentRequest{Amount: -10.0, Recipient: "0x742d35Cc"},
			true,
		},
		{
			"Empty recipient",
			&types.PaymentRequest{Amount: 100.0, Recipient: ""},
			true,
		},
		{
			"Short recipient",
			&types.PaymentRequest{Amount: 100.0, Recipient: "0x123"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sim.validateRequest(tt.req)
			if tt.shouldErr && err == nil {
				t.Error("validateRequest() should return error")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("validateRequest() should not return error: %v", err)
			}
		})
	}
}

func TestGenerateTxHash(t *testing.T) {
	cfg := &config.Config{}
	sim := NewSimulator(cfg)

	hash1 := sim.generateTxHash()
	hash2 := sim.generateTxHash()

	// Check format (0x + 64 hex chars)
	if len(hash1) != 66 { // 0x + 64 chars
		t.Errorf("TxHash length: got %d, want 66", len(hash1))
	}

	if hash1[:2] != "0x" {
		t.Error("TxHash should start with 0x")
	}

	// Check uniqueness
	if hash1 == hash2 {
		t.Error("generateTxHash() should generate unique hashes")
	}
}

func TestGenerateBlockNumber(t *testing.T) {
	cfg := &config.Config{}
	sim := NewSimulator(cfg)

	blockNum := sim.generateBlockNumber()

	if blockNum == 0 {
		t.Error("generateBlockNumber() should not return 0")
	}

	// Block number should be reasonable (timestamp-based)
	now := time.Now().Unix() / 12
	if blockNum < now-100 || blockNum > now+100 {
		t.Errorf("Block number seems unreasonable: %d", blockNum)
	}
}

func TestGetTransactionStatus_SimulationMode(t *testing.T) {
	cfg := &config.Config{
		SimulationMode: true,
	}

	sim := NewSimulator(cfg)

	status, err := sim.GetTransactionStatus("0xabc123")

	if err != nil {
		t.Errorf("GetTransactionStatus() failed: %v", err)
	}

	if status != "confirmed" {
		t.Errorf("Status: got %s, want confirmed", status)
	}
}

func TestGetTransactionStatus_RealMode(t *testing.T) {
	cfg := &config.Config{
		SimulationMode: false,
	}

	sim := NewSimulator(cfg)

	status, err := sim.GetTransactionStatus("0xabc123")

	if err == nil {
		t.Error("GetTransactionStatus() should return error in real mode")
	}

	if status != "pending" {
		t.Errorf("Status: got %s, want pending", status)
	}
}

func TestProcessPayment_NoDelay(t *testing.T) {
	cfg := &config.Config{
		SimulationMode: true,
		TxDelayMs:      0, // No delay
		WalletAddress:  "0xAGENT",
	}

	sim := NewSimulator(cfg)

	req := &types.PaymentRequest{
		Amount:    100.0,
		Recipient: "0x742d35Cc",
	}

	start := time.Now()
	_, err := sim.ProcessPayment(req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ProcessPayment() failed: %v", err)
	}

	// Should complete quickly with no delay
	if elapsed > 100*time.Millisecond {
		t.Errorf("Transaction took too long with no delay: %v", elapsed)
	}
}

func TestProcessPayment_WithAllFields(t *testing.T) {
	cfg := &config.Config{
		SimulationMode: true,
		TxDelayMs:      0,
		WalletAddress:  "0xAGENT",
	}

	sim := NewSimulator(cfg)

	req := &types.PaymentRequest{
		Amount:      100.0,
		Currency:    "USDC",
		Product:     "Sunglasses",
		Recipient:   "0x742d35Cc",
		Description: "Test payment",
		Metadata: map[string]interface{}{
			"orderId": "12345",
		},
	}

	result, err := sim.ProcessPayment(req)

	if err != nil {
		t.Fatalf("ProcessPayment() failed: %v", err)
	}

	if result == nil {
		t.Fatal("ProcessPayment() returned nil result")
	}

	// Metadata doesn't affect transaction processing
	if result.Status != "confirmed" {
		t.Errorf("Status: got %s, want confirmed", result.Status)
	}
}
