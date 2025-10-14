package sage

import (
	"net/http/httptest"
	"testing"

	"github.com/sage-x-project/sage-payment-agent-for-demo/config"
	"github.com/sage-x-project/sage-payment-agent-for-demo/types"
)

func TestNewVerifier(t *testing.T) {
	cfg := &config.Config{}
	verifier := NewVerifier(cfg)

	if verifier == nil {
		t.Fatal("NewVerifier() returned nil")
	}

	if verifier.config != cfg {
		t.Error("NewVerifier() didn't set config properly")
	}

	if verifier.stats == nil {
		t.Error("NewVerifier() didn't initialize stats")
	}
}

func TestVerifyRequest_SAGEDisabled(t *testing.T) {
	cfg := &config.Config{
		SAGEEnabled: false,
	}

	verifier := NewVerifier(cfg)
	req := httptest.NewRequest("POST", "/payment", nil)

	result, err := verifier.VerifyRequest(req)

	if err != nil {
		t.Errorf("VerifyRequest() should not return error when SAGE disabled: %v", err)
	}

	if result == nil {
		t.Fatal("VerifyRequest() returned nil result")
	}

	if result.Valid {
		t.Error("Result should be invalid when SAGE is disabled")
	}

	if result.ErrorMessage != "SAGE protocol disabled" {
		t.Errorf("ErrorMessage: got %s", result.ErrorMessage)
	}
}

func TestVerifyRequest_MissingHeaders(t *testing.T) {
	cfg := &config.Config{
		SAGEEnabled: true,
	}

	verifier := NewVerifier(cfg)
	req := httptest.NewRequest("POST", "/payment", nil)
	// No signature headers

	result, err := verifier.VerifyRequest(req)

	if err == nil {
		t.Error("VerifyRequest() should return error for missing headers")
	}

	if result.Valid {
		t.Error("Result should be invalid for missing headers")
	}

	if verifier.stats.RejectedRequests != 1 {
		t.Errorf("RejectedRequests: got %d, want 1", verifier.stats.RejectedRequests)
	}
}

func TestVerifyRequest_ValidSignature(t *testing.T) {
	cfg := &config.Config{
		SAGEEnabled: true,
		StrictMode:  true,
	}

	verifier := NewVerifier(cfg)
	req := httptest.NewRequest("POST", "/payment", nil)
	req.Header.Set("Signature-Input", `sig1=("@method" "@authority");keyid="did:sage:0x123";created=1690000000`)
	req.Header.Set("Signature", "sig1=:MEUCIQDx...:")
	req.Header.Set("Content-Digest", "sha-256=:X48E9...:")

	result, err := verifier.VerifyRequest(req)

	if err != nil {
		t.Errorf("VerifyRequest() failed: %v", err)
	}

	if !result.Valid {
		t.Error("Result should be valid for proper signature headers")
	}

	if verifier.stats.VerifiedRequests != 1 {
		t.Errorf("VerifiedRequests: got %d, want 1", verifier.stats.VerifiedRequests)
	}
}

func TestVerifyRequest_InvalidDID(t *testing.T) {
	cfg := &config.Config{
		SAGEEnabled: true,
	}

	verifier := NewVerifier(cfg)
	req := httptest.NewRequest("POST", "/payment", nil)
	req.Header.Set("Signature-Input", `sig1=("@method");keyid="invalid_did"`)
	req.Header.Set("Signature", "sig1=:MEUCIQDx...:")

	result, err := verifier.VerifyRequest(req)

	if err == nil {
		t.Error("VerifyRequest() should return error for invalid DID")
	}

	if result.Valid {
		t.Error("Result should be invalid for invalid DID")
	}

	if result.PublicKeyFound {
		t.Error("PublicKeyFound should be false for invalid DID")
	}
}

func TestExtractSignerDID(t *testing.T) {
	cfg := &config.Config{}
	verifier := NewVerifier(cfg)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"Valid DID",
			`sig1=("@method");keyid="did:sage:0x123";created=1690000000`,
			"did:sage:0x123",
		},
		{
			"Ethereum address",
			`sig1=("@method");keyid="0x742d35Cc";created=1690000000`,
			"0x742d35Cc",
		},
		{
			"No keyid",
			`sig1=("@method");created=1690000000`,
			"unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := verifier.extractSignerDID(tt.input)
			if result != tt.expected {
				t.Errorf("extractSignerDID(): got %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestResolveDID(t *testing.T) {
	cfg := &config.Config{}
	verifier := NewVerifier(cfg)

	tests := []struct {
		name     string
		did      string
		expected bool
	}{
		{"Valid DID", "did:sage:0x123", true},
		{"Valid Ethereum address", "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", true},
		{"Short Ethereum address", "0x742d35Cc", true},
		{"Invalid DID", "invalid", false},
		{"Empty DID", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := verifier.resolveDID(tt.did)
			if result != tt.expected {
				t.Errorf("resolveDID(%s): got %v, want %v", tt.did, result, tt.expected)
			}
		})
	}
}

func TestVerifyContentDigest(t *testing.T) {
	cfg := &config.Config{}
	verifier := NewVerifier(cfg)

	req := httptest.NewRequest("POST", "/payment", nil)

	tests := []struct {
		name     string
		digest   string
		expected bool
	}{
		{"Valid digest", "sha-256=:X48E9...:", true},
		{"Empty digest", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := verifier.verifyContentDigest(req, tt.digest)
			if result != tt.expected {
				t.Errorf("verifyContentDigest(): got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestShouldReject(t *testing.T) {
	tests := []struct {
		name       string
		sageEnabled bool
		strictMode bool
		valid      bool
		expected   bool
	}{
		{"SAGE disabled", false, true, false, false},
		{"Strict mode + invalid", true, true, false, true},
		{"Strict mode + valid", true, true, true, false},
		{"Non-strict + invalid", true, false, false, false},
		{"Non-strict + valid", true, false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				SAGEEnabled: tt.sageEnabled,
				StrictMode:  tt.strictMode,
			}

			verifier := NewVerifier(cfg)
			result := &types.VerificationResult{
				Valid: tt.valid,
			}

			shouldReject := verifier.ShouldReject(result)
			if shouldReject != tt.expected {
				t.Errorf("ShouldReject(): got %v, want %v", shouldReject, tt.expected)
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	cfg := &config.Config{
		SAGEEnabled: true,
	}

	verifier := NewVerifier(cfg)

	// Initially zero
	stats := verifier.GetStats()
	if stats.VerifiedRequests != 0 {
		t.Errorf("Initial VerifiedRequests: got %d, want 0", stats.VerifiedRequests)
	}
	if stats.RejectedRequests != 0 {
		t.Errorf("Initial RejectedRequests: got %d, want 0", stats.RejectedRequests)
	}

	// Verify a request
	req := httptest.NewRequest("POST", "/payment", nil)
	req.Header.Set("Signature-Input", `sig1=("@method");keyid="did:sage:0x123"`)
	req.Header.Set("Signature", "sig1=:MEUCIQDx...:")
	req.Header.Set("Content-Digest", "sha-256=:X48E9...:")

	verifier.VerifyRequest(req)

	// Check stats updated
	stats = verifier.GetStats()
	if stats.VerifiedRequests != 1 {
		t.Errorf("VerifiedRequests after verify: got %d, want 1", stats.VerifiedRequests)
	}
}

func TestVerifyRequest_NoContentDigest(t *testing.T) {
	cfg := &config.Config{
		SAGEEnabled: true,
	}

	verifier := NewVerifier(cfg)
	req := httptest.NewRequest("POST", "/payment", nil)
	req.Header.Set("Signature-Input", `sig1=("@method");keyid="did:sage:0x123"`)
	req.Header.Set("Signature", "sig1=:MEUCIQDx...:")
	// No Content-Digest header

	result, err := verifier.VerifyRequest(req)

	if err != nil {
		t.Errorf("VerifyRequest() should not fail without Content-Digest: %v", err)
	}

	// Should still pass if Content-Digest is optional
	if !result.Valid {
		t.Error("Result should be valid even without Content-Digest")
	}
}

func TestVerifyRequest_MultipleRequests(t *testing.T) {
	cfg := &config.Config{
		SAGEEnabled: true,
	}

	verifier := NewVerifier(cfg)

	// First request - valid
	req1 := httptest.NewRequest("POST", "/payment", nil)
	req1.Header.Set("Signature-Input", `sig1=("@method");keyid="did:sage:0x123"`)
	req1.Header.Set("Signature", "sig1=:MEUCIQDx...:")
	req1.Header.Set("Content-Digest", "sha-256=:X48E9...:")

	verifier.VerifyRequest(req1)

	// Second request - invalid
	req2 := httptest.NewRequest("POST", "/payment", nil)
	// Missing headers

	verifier.VerifyRequest(req2)

	// Check stats
	stats := verifier.GetStats()
	if stats.VerifiedRequests != 1 {
		t.Errorf("VerifiedRequests: got %d, want 1", stats.VerifiedRequests)
	}
	if stats.RejectedRequests != 1 {
		t.Errorf("RejectedRequests: got %d, want 1", stats.RejectedRequests)
	}
}
