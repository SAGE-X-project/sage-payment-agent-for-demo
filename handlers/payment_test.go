package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sage-x-project/sage-payment-agent-for-demo/config"
	"github.com/sage-x-project/sage-payment-agent-for-demo/types"
)

func TestNewPaymentHandler(t *testing.T) {
	cfg := &config.Config{}
	handler := NewPaymentHandler(cfg)

	if handler == nil {
		t.Fatal("NewPaymentHandler() returned nil")
	}

	if handler.config != cfg {
		t.Error("NewPaymentHandler() didn't set config")
	}

	if handler.verifier == nil {
		t.Error("NewPaymentHandler() didn't initialize verifier")
	}

	if handler.simulator == nil {
		t.Error("NewPaymentHandler() didn't initialize simulator")
	}

	if handler.txStats == nil {
		t.Error("NewPaymentHandler() didn't initialize txStats")
	}
}

func TestHandlePayment_MethodNotAllowed(t *testing.T) {
	cfg := &config.Config{}
	handler := NewPaymentHandler(cfg)

	req := httptest.NewRequest("GET", "/payment", nil)
	w := httptest.NewRecorder()

	handler.HandlePayment(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status code: got %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandlePayment_InvalidJSON(t *testing.T) {
	cfg := &config.Config{
		SAGEEnabled: false,
	}
	handler := NewPaymentHandler(cfg)

	req := httptest.NewRequest("POST", "/payment", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	handler.HandlePayment(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlePayment_Success_SAGEDisabled(t *testing.T) {
	cfg := &config.Config{
		SAGEEnabled:    false,
		SimulationMode: true,
		TxDelayMs:      0,
		WalletAddress:  "0xAGENT",
	}
	handler := NewPaymentHandler(cfg)

	paymentReq := types.PaymentRequest{
		Amount:    100.0,
		Currency:  "USDC",
		Recipient: "0x742d35Cc",
	}
	body, _ := json.Marshal(paymentReq)

	req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandlePayment(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp types.PaymentResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp.Success {
		t.Error("Response success should be true")
	}

	if resp.TransactionID == "" {
		t.Error("TransactionID should not be empty")
	}

	if resp.Amount != 100.0 {
		t.Errorf("Amount: got %f, want 100.0", resp.Amount)
	}

	if resp.Status != "confirmed" {
		t.Errorf("Status: got %s, want confirmed", resp.Status)
	}
}

func TestHandlePayment_Success_SAGEEnabled(t *testing.T) {
	cfg := &config.Config{
		SAGEEnabled:    true,
		StrictMode:     false, // Non-strict to allow through
		SimulationMode: true,
		TxDelayMs:      0,
		WalletAddress:  "0xAGENT",
	}
	handler := NewPaymentHandler(cfg)

	paymentReq := types.PaymentRequest{
		Amount:    100.0,
		Recipient: "0x742d35Cc",
	}
	body, _ := json.Marshal(paymentReq)

	req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Signature-Input", `sig1=("@method");keyid="did:sage:0x123"`)
	req.Header.Set("Signature", "sig1=:MEUCIQDx...:")
	req.Header.Set("Content-Digest", "sha-256=:X48E9...:")
	w := httptest.NewRecorder()

	handler.HandlePayment(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code: got %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandlePayment_SAGEVerificationFailed(t *testing.T) {
	cfg := &config.Config{
		SAGEEnabled: true,
		StrictMode:  true, // Strict mode - reject invalid signatures
	}
	handler := NewPaymentHandler(cfg)

	paymentReq := types.PaymentRequest{
		Amount:    100.0,
		Recipient: "0x742d35Cc",
	}
	body, _ := json.Marshal(paymentReq)

	req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	// No signature headers

	w := httptest.NewRecorder()

	handler.HandlePayment(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status code: got %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var resp types.PaymentResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Success {
		t.Error("Response success should be false")
	}

	if resp.Error != "signature_verification_failed" {
		t.Errorf("Error code: got %s, want signature_verification_failed", resp.Error)
	}
}

func TestHandlePayment_InvalidPaymentAmount(t *testing.T) {
	cfg := &config.Config{
		SAGEEnabled:    false,
		SimulationMode: true,
		TxDelayMs:      0,
	}
	handler := NewPaymentHandler(cfg)

	paymentReq := types.PaymentRequest{
		Amount:    0.0, // Invalid amount
		Recipient: "0x742d35Cc",
	}
	body, _ := json.Marshal(paymentReq)

	req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.HandlePayment(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code: got %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp types.PaymentResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Success {
		t.Error("Response success should be false")
	}
}

func TestHandleHealth(t *testing.T) {
	cfg := &config.Config{
		SAGEEnabled: true,
	}
	handler := NewPaymentHandler(cfg)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.HandleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code: got %d, want %d", w.Code, http.StatusOK)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Content-Type should be application/json")
	}

	var resp types.HealthResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "healthy" {
		t.Errorf("Status: got %s, want healthy", resp.Status)
	}

	if !resp.SAGEEnabled {
		t.Error("SAGEEnabled should be true")
	}
}

func TestHandleStatus(t *testing.T) {
	cfg := &config.Config{
		SAGEEnabled: true,
		StrictMode:  true,
		AgentVersion: "1.0.0",
	}
	handler := NewPaymentHandler(cfg)

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	handler.HandleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code: got %d, want %d", w.Code, http.StatusOK)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Content-Type should be application/json")
	}

	var resp types.StatusResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Agent != "sage-payment-agent" {
		t.Errorf("Agent: got %s, want sage-payment-agent", resp.Agent)
	}

	if resp.Version != "1.0.0" {
		t.Errorf("Version: got %s, want 1.0.0", resp.Version)
	}

	if !resp.SAGEProtocol.Enabled {
		t.Error("SAGEProtocol.Enabled should be true")
	}

	if !resp.SAGEProtocol.StrictMode {
		t.Error("SAGEProtocol.StrictMode should be true")
	}
}

func TestHandleStatus_WithStats(t *testing.T) {
	cfg := &config.Config{
		SAGEEnabled:    false,
		SimulationMode: true,
		TxDelayMs:      0,
		WalletAddress:  "0xAGENT",
		AgentVersion:   "1.0.0",
	}
	handler := NewPaymentHandler(cfg)

	// Process a successful payment to update stats
	paymentReq := types.PaymentRequest{
		Amount:    100.0,
		Recipient: "0x742d35Cc",
	}
	body, _ := json.Marshal(paymentReq)

	req1 := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(body))
	w1 := httptest.NewRecorder()
	handler.HandlePayment(w1, req1)

	// Check status
	req2 := httptest.NewRequest("GET", "/status", nil)
	w2 := httptest.NewRecorder()
	handler.HandleStatus(w2, req2)

	var resp types.StatusResponse
	json.NewDecoder(w2.Body).Decode(&resp)

	if resp.Transactions.Total != 1 {
		t.Errorf("Total transactions: got %d, want 1", resp.Transactions.Total)
	}

	if resp.Transactions.Successful != 1 {
		t.Errorf("Successful transactions: got %d, want 1", resp.Transactions.Successful)
	}
}

func TestSendError(t *testing.T) {
	cfg := &config.Config{}
	handler := NewPaymentHandler(cfg)

	w := httptest.NewRecorder()
	details := &types.ErrorDetails{
		SignatureValid: false,
		Reason:         "test reason",
	}

	handler.sendError(w, http.StatusBadRequest, "test_error", "Test message", details)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code: got %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp types.PaymentResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Success {
		t.Error("Response success should be false")
	}

	if resp.Error != "test_error" {
		t.Errorf("Error code: got %s, want test_error", resp.Error)
	}

	if resp.Message != "Test message" {
		t.Errorf("Message: got %s, want Test message", resp.Message)
	}

	if resp.Details == nil {
		t.Fatal("Details should not be nil")
	}

	if resp.Details.Reason != "test reason" {
		t.Errorf("Details.Reason: got %s, want test reason", resp.Details.Reason)
	}
}

func TestHandlePayment_EmptyBody(t *testing.T) {
	cfg := &config.Config{
		SAGEEnabled: false,
	}
	handler := NewPaymentHandler(cfg)

	req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer([]byte{}))
	w := httptest.NewRecorder()

	handler.HandlePayment(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlePayment_TransactionStats(t *testing.T) {
	cfg := &config.Config{
		SAGEEnabled:    false,
		SimulationMode: true,
		TxDelayMs:      0,
		WalletAddress:  "0xAGENT",
	}
	handler := NewPaymentHandler(cfg)

	// Initial stats
	if handler.txStats.Total != 0 {
		t.Error("Initial total should be 0")
	}

	// Successful payment
	paymentReq := types.PaymentRequest{
		Amount:    100.0,
		Recipient: "0x742d35Cc",
	}
	body, _ := json.Marshal(paymentReq)

	req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.HandlePayment(w, req)

	if handler.txStats.Total != 1 {
		t.Errorf("Total: got %d, want 1", handler.txStats.Total)
	}

	if handler.txStats.Successful != 1 {
		t.Errorf("Successful: got %d, want 1", handler.txStats.Successful)
	}

	// Failed payment
	invalidReq := types.PaymentRequest{
		Amount:    0.0, // Invalid
		Recipient: "0x742d35Cc",
	}
	body2, _ := json.Marshal(invalidReq)

	req2 := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(body2))
	w2 := httptest.NewRecorder()

	handler.HandlePayment(w2, req2)

	if handler.txStats.Total != 2 {
		t.Errorf("Total: got %d, want 2", handler.txStats.Total)
	}

	if handler.txStats.Failed != 1 {
		t.Errorf("Failed: got %d, want 1", handler.txStats.Failed)
	}
}
