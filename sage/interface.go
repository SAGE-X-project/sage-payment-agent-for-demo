package sage

import (
	"net/http"

	"github.com/sage-x-project/sage-payment-agent-for-demo/types"
)

// SignatureVerifier is the interface for SAGE signature verification
type SignatureVerifier interface {
	// VerifyRequest verifies the HTTP request signature
	VerifyRequest(r *http.Request) (*types.VerificationResult, error)

	// ShouldReject determines if request should be rejected
	ShouldReject(result *types.VerificationResult) bool

	// GetStats returns verification statistics
	GetStats() *VerificationStats
}
