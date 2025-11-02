package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/sage-x-project/sage-payment-agent-for-demo/config"
	"github.com/sage-x-project/sage-payment-agent-for-demo/logger"
	"github.com/sage-x-project/sage-payment-agent-for-demo/sage"
	"github.com/sage-x-project/sage-payment-agent-for-demo/transaction"
	"github.com/sage-x-project/sage-payment-agent-for-demo/types"
)

// PaymentHandler handles payment requests
type PaymentHandler struct {
	config     *config.Config
	verifier   *sage.Verifier
	simulator  *transaction.Simulator
	txStats    *TransactionStats
}

// TransactionStats tracks transaction statistics
type TransactionStats struct {
	Total      int64
	Successful int64
	Failed     int64
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(cfg *config.Config) *PaymentHandler {
	return &PaymentHandler{
		config:     cfg,
		verifier:   sage.NewVerifier(cfg),
		simulator:  transaction.NewSimulator(cfg),
		txStats:    &TransactionStats{},
	}
}

// HandlePayment handles POST /payment requests
func (h *PaymentHandler) HandlePayment(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		logger.Warn("Method not allowed: %s", r.Method)
		h.sendError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST method is allowed", nil)
		return
	}

	// Read request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed to read request body: %v", err)
		h.sendError(w, http.StatusBadRequest, "invalid_request", "Failed to read request body", nil)
		return
	}
	defer r.Body.Close()

	// Parse payment request - try contract format first, then fallback to legacy
	paymentReq, err := h.parsePaymentRequest(bodyBytes)
	if err != nil {
		logger.Error("Failed to parse request: %v", err)
		h.sendError(w, http.StatusBadRequest, "invalid_request", err.Error(), nil)
		return
	}

	logger.LogPaymentRequest(&paymentReq)

	// Verify SAGE signature if enabled
	if h.config.IsSAGEEnabled() {
		verifyResult, err := h.verifier.VerifyRequest(r)
		logger.LogVerification(verifyResult)

		if err != nil || h.verifier.ShouldReject(verifyResult) {
			h.txStats.Failed++
			h.sendError(w, http.StatusUnauthorized, "signature_verification_failed", "Invalid RFC-9421 signature", &types.ErrorDetails{
				SignatureValid: verifyResult.Valid,
				Reason:         verifyResult.ErrorMessage,
			})
			return
		}
	} else {
		logger.Warn("Processing unverified request (SAGE disabled)")
	}

	// Process payment transaction
	h.txStats.Total++
	txResult, err := h.simulator.ProcessPayment(&paymentReq)
	if err != nil {
		h.txStats.Failed++
		logger.Error("Payment processing failed: %v", err)
		h.sendError(w, http.StatusBadRequest, "payment_failed", err.Error(), nil)
		return
	}

	h.txStats.Successful++
	logger.LogTransaction(txResult)

	// Send success response
	response := types.PaymentResponse{
		Success:       true,
		TransactionID: txResult.TxHash,
		Amount:        paymentReq.Amount,
		Recipient:     paymentReq.Recipient,
		Status:        txResult.Status,
		Timestamp:     txResult.Timestamp,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleHealth handles GET /health requests
func (h *PaymentHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	response := types.HealthResponse{
		Status:      "healthy",
		SAGEEnabled: h.config.IsSAGEEnabled(),
		UptimeSeconds: int64(h.config.GetUptime().Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleStatus handles GET /status requests
func (h *PaymentHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	stats := h.verifier.GetStats()

	response := types.StatusResponse{
		Agent:   "sage-payment-agent",
		Version: h.config.AgentVersion,
		SAGEProtocol: types.SAGEStatus{
			Enabled:          h.config.IsSAGEEnabled(),
			StrictMode:       h.config.IsStrictMode(),
			VerifiedRequests: stats.VerifiedRequests,
			RejectedRequests: stats.RejectedRequests,
		},
		Transactions: types.TransactionStats{
			Total:      h.txStats.Total,
			Successful: h.txStats.Successful,
			Failed:     h.txStats.Failed,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// parsePaymentRequest parses request in contract format or legacy format
func (h *PaymentHandler) parsePaymentRequest(bodyBytes []byte) (types.PaymentRequest, error) {
	// Try contract format first (Root Agent -> Target Agent)
	var agentReq types.AgentRequest
	if err := json.Unmarshal(bodyBytes, &agentReq); err == nil && agentReq.Intent == "payment" {
		logger.Info("Received contract format request")
		return h.convertFromAgentRequest(&agentReq)
	}

	// Try AgentMessage format (envelope with metadata)
	var agentMsg map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &agentMsg); err == nil {
		if metadata, ok := agentMsg["metadata"].(map[string]interface{}); ok {
			// Check if this looks like a payment message with metadata
			if hasPaymentMetadata(metadata) {
				logger.Info("Received AgentMessage format request")
				return h.convertFromAgentMessage(metadata)
			}
		}
	}

	// Fallback to legacy format
	var paymentReq types.PaymentRequest
	if err := json.Unmarshal(bodyBytes, &paymentReq); err != nil {
		return types.PaymentRequest{}, err
	}
	logger.Info("Received legacy format request")
	return paymentReq, nil
}

// hasPaymentMetadata checks if metadata contains payment information
func hasPaymentMetadata(metadata map[string]interface{}) bool {
	// Check for payment-specific keys
	_, hasAmountKRW := metadata["amountKRW"]
	_, hasPaymentAmount := metadata["payment.amountKRW"]
	_, hasRecipient := metadata["recipient"]
	_, hasTo := metadata["to"]

	return hasAmountKRW || hasPaymentAmount || hasRecipient || hasTo
}

// convertFromAgentMessage converts AgentMessage metadata to PaymentRequest
func (h *PaymentHandler) convertFromAgentMessage(metadata map[string]interface{}) (types.PaymentRequest, error) {
	// Extract amount (KRW to float64 conversion)
	var amount float64
	if v, ok := metadata["amountKRW"]; ok {
		switch val := v.(type) {
		case float64:
			amount = val
		case int:
			amount = float64(val)
		case int64:
			amount = float64(val)
		}
	} else if v, ok := metadata["payment.amountKRW"]; ok {
		switch val := v.(type) {
		case float64:
			amount = val
		case int:
			amount = float64(val)
		case int64:
			amount = float64(val)
		}
	}

	// Extract recipient
	recipient, _ := metadata["recipient"].(string)
	if recipient == "" {
		recipient, _ = metadata["to"].(string)
	}
	if recipient == "" {
		recipient, _ = metadata["payment.to"].(string)
	}

	// Extract product/item
	product, _ := metadata["item"].(string)
	if product == "" {
		product, _ = metadata["payment.item"].(string)
	}

	// Extract method for description
	method, _ := metadata["method"].(string)
	if method == "" {
		method, _ = metadata["payment.method"].(string)
	}

	return types.PaymentRequest{
		Amount:      amount,
		Currency:    "KRW",
		Recipient:   recipient,
		Product:     product,
		Description: method,
		Metadata:    metadata,
	}, nil
}

// convertFromAgentRequest converts AgentRequest to PaymentRequest
func (h *PaymentHandler) convertFromAgentRequest(agentReq *types.AgentRequest) (types.PaymentRequest, error) {
	params := agentReq.Parameters

	// Extract payment parameters
	amount, _ := params["amount"].(float64)
	currency, _ := params["currency"].(string)
	recipient, _ := params["recipient"].(string)
	product, _ := params["product"].(string)
	description, _ := params["description"].(string)

	// Default currency
	if currency == "" {
		currency = "KRW"
	}

	return types.PaymentRequest{
		Amount:      amount,
		Currency:    currency,
		Recipient:   recipient,
		Product:     product,
		Description: description,
		Metadata:    map[string]interface{}{
			"requestId":   agentReq.Metadata.RequestID,
			"sourceAgent": agentReq.Metadata.SourceAgent,
			"timestamp":   agentReq.Metadata.Timestamp,
		},
	}, nil
}

// sendError sends an error response
func (h *PaymentHandler) sendError(w http.ResponseWriter, status int, errorCode, message string, details *types.ErrorDetails) {
	response := types.PaymentResponse{
		Success:   false,
		Error:     errorCode,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
