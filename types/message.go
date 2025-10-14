package types

import "time"

// PaymentRequest represents an incoming payment request
type PaymentRequest struct {
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency,omitempty"`
	Product     string  `json:"product,omitempty"`
	Recipient   string  `json:"recipient"`
	Description string  `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PaymentResponse represents the payment processing result
type PaymentResponse struct {
	Success       bool      `json:"success"`
	TransactionID string    `json:"transaction_id,omitempty"`
	Amount        float64   `json:"amount,omitempty"`
	Recipient     string    `json:"recipient,omitempty"`
	Status        string    `json:"status,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	Error         string    `json:"error,omitempty"`
	Message       string    `json:"message,omitempty"`
	Details       *ErrorDetails `json:"details,omitempty"`
}

// ErrorDetails provides additional error information
type ErrorDetails struct {
	SignatureValid bool   `json:"signature_valid"`
	Reason         string `json:"reason"`
	Code           string `json:"code,omitempty"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status      string `json:"status"`
	SAGEEnabled bool   `json:"sage_enabled"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

// StatusResponse represents detailed status information
type StatusResponse struct {
	Agent         string          `json:"agent"`
	Version       string          `json:"version"`
	SAGEProtocol  SAGEStatus      `json:"sage_protocol"`
	Transactions  TransactionStats `json:"transactions"`
}

// SAGEStatus represents SAGE protocol status
type SAGEStatus struct {
	Enabled          bool  `json:"enabled"`
	StrictMode       bool  `json:"strict_mode"`
	VerifiedRequests int64 `json:"verified_requests"`
	RejectedRequests int64 `json:"rejected_requests"`
}

// TransactionStats represents transaction statistics
type TransactionStats struct {
	Total      int64 `json:"total"`
	Successful int64 `json:"successful"`
	Failed     int64 `json:"failed"`
}

// VerificationResult represents SAGE signature verification result
type VerificationResult struct {
	Valid         bool      `json:"valid"`
	SignerDID     string    `json:"signer_did,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	PublicKeyFound bool     `json:"public_key_found"`
}

// TransactionResult represents blockchain transaction result
type TransactionResult struct {
	TxHash    string    `json:"tx_hash"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"` // pending, confirmed, failed
	Timestamp time.Time `json:"timestamp"`
	BlockNumber int64   `json:"block_number,omitempty"`
	GasUsed     int64   `json:"gas_used,omitempty"`
}

// AttackLog represents detected attack attempts
type AttackLog struct {
	Timestamp     time.Time     `json:"timestamp"`
	AttackType    string        `json:"attack_type"`
	OriginalRequest  *PaymentRequest `json:"original_request,omitempty"`
	TamperedRequest  *PaymentRequest `json:"tampered_request,omitempty"`
	Changes       []Change      `json:"changes"`
	Blocked       bool          `json:"blocked"`
	SourceIP      string        `json:"source_ip,omitempty"`
}

// Change represents a detected change in the request
type Change struct {
	Field         string      `json:"field"`
	ExpectedValue interface{} `json:"expected_value"`
	ActualValue   interface{} `json:"actual_value"`
}
