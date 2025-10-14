package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPaymentRequest_Marshal(t *testing.T) {
	req := PaymentRequest{
		Amount:    100.0,
		Currency:  "USDC",
		Product:   "Sunglasses",
		Recipient: "0x742d35Cc",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Marshal produced empty data")
	}
}

func TestPaymentRequest_Unmarshal(t *testing.T) {
	jsonData := `{"amount":100.0,"currency":"USDC","recipient":"0x742d35Cc"}`

	var req PaymentRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if req.Amount != 100.0 {
		t.Errorf("Amount: got %f, want 100.0", req.Amount)
	}
}

func TestPaymentResponse_Marshal(t *testing.T) {
	resp := PaymentResponse{
		Success:       true,
		TransactionID: "0xabc123",
		Amount:        100.0,
		Timestamp:     time.Now(),
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Marshal produced empty data")
	}
}

func TestVerificationResult_Marshal(t *testing.T) {
	result := VerificationResult{
		Valid:     true,
		SignerDID: "did:sage:0x123",
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Marshal produced empty data")
	}
}

func TestTransactionResult_Marshal(t *testing.T) {
	tx := TransactionResult{
		TxHash:    "0xabc123",
		From:      "0x111",
		To:        "0x222",
		Amount:    100.0,
		Status:    "confirmed",
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Marshal produced empty data")
	}
}
