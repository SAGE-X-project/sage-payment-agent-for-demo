package sage

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sage-x-project/sage-payment-agent-for-demo/config"
	"github.com/sage-x-project/sage-payment-agent-for-demo/logger"
	"github.com/sage-x-project/sage-payment-agent-for-demo/types"
)

// Verifier handles RFC-9421 signature verification
type Verifier struct {
	config *config.Config
	stats  *VerificationStats
}

// VerificationStats tracks verification statistics
type VerificationStats struct {
	VerifiedRequests int64
	RejectedRequests int64
}

// NewVerifier creates a new SAGE verifier
func NewVerifier(cfg *config.Config) *Verifier {
	return &Verifier{
		config: cfg,
		stats:  &VerificationStats{},
	}
}

// VerifyRequest verifies the HTTP request signature using RFC-9421
func (v *Verifier) VerifyRequest(r *http.Request) (*types.VerificationResult, error) {
	result := &types.VerificationResult{
		Timestamp: time.Now(),
	}

	// Check if SAGE is enabled
	if !v.config.IsSAGEEnabled() {
		logger.Warn("SAGE Protocol disabled - skipping verification")
		result.Valid = false
		result.ErrorMessage = "SAGE protocol disabled"
		return result, nil
	}

	logger.Debug("Verifying RFC-9421 signature...")

	// Check for required headers
	signatureInput := r.Header.Get("Signature-Input")
	signature := r.Header.Get("Signature")
	contentDigest := r.Header.Get("Content-Digest")

	if signatureInput == "" || signature == "" {
		v.stats.RejectedRequests++
		result.Valid = false
		result.ErrorMessage = "Missing required signature headers"
		logger.Error("Missing Signature or Signature-Input headers")
		return result, fmt.Errorf("missing signature headers")
	}

	// In a real implementation, we would:
	// 1. Parse Signature-Input to get signature parameters
	// 2. Reconstruct the signature base from request components
	// 3. Retrieve public key from DID registry (blockchain)
	// 4. Verify cryptographic signature
	//
	// For demo purposes, we'll do simplified validation

	// Extract signer DID from signature (simplified)
	signerDID := v.extractSignerDID(signatureInput)
	result.SignerDID = signerDID

	// Simulate DID resolution
	publicKeyFound := v.resolveDID(signerDID)
	result.PublicKeyFound = publicKeyFound

	if !publicKeyFound {
		v.stats.RejectedRequests++
		result.Valid = false
		result.ErrorMessage = "Public key not found in DID registry"
		logger.Error("DID resolution failed: %s", signerDID)
		return result, fmt.Errorf("DID not found: %s", signerDID)
	}

	// Verify content digest if present
	if contentDigest != "" {
		if !v.verifyContentDigest(r, contentDigest) {
			v.stats.RejectedRequests++
			result.Valid = false
			result.ErrorMessage = "Content digest verification failed"
			logger.Error("Content digest mismatch")
			return result, fmt.Errorf("content digest mismatch")
		}
	}

	// If we reach here, verification passed (in demo mode)
	v.stats.VerifiedRequests++
	result.Valid = true
	logger.Info("Signature verification: PASSED")

	return result, nil
}

// extractSignerDID extracts the signer's DID from Signature-Input header
func (v *Verifier) extractSignerDID(signatureInput string) string {
	// Simplified extraction - in real implementation would parse properly
	// Example: sig1=(\"@method\" \"@authority\");keyid=\"did:sage:0x123...\";created=1690000000

	// Look for keyid parameter
	if idx := strings.Index(signatureInput, "keyid="); idx != -1 {
		start := idx + 7 // Skip 'keyid="'
		end := strings.Index(signatureInput[start:], "\"")
		if end != -1 {
			return signatureInput[start : start+end]
		}
	}

	return "unknown"
}

// resolveDID simulates resolving a DID to get public key
func (v *Verifier) resolveDID(did string) bool {
	logger.Debug("Resolving DID: %s", did)

	// In a real implementation, this would:
	// 1. Connect to blockchain RPC
	// 2. Query DID registry contract
	// 3. Retrieve public key for the DID
	//
	// For demo purposes, we'll simulate success if DID looks valid

	// Check if DID format is valid (starts with "did:")
	if strings.HasPrefix(did, "did:") {
		logger.Debug("DID resolution: SUCCESS")
		return true
	}

	// For demo, also accept Ethereum addresses
	if strings.HasPrefix(did, "0x") && len(did) >= 10 {
		logger.Debug("DID resolution: SUCCESS (Ethereum address)")
		return true
	}

	logger.Debug("DID resolution: FAILED")
	return false
}

// verifyContentDigest verifies the Content-Digest header
func (v *Verifier) verifyContentDigest(r *http.Request, contentDigest string) bool {
	logger.Debug("Verifying content digest...")

	// In a real implementation, this would:
	// 1. Read the request body
	// 2. Compute SHA-256 hash
	// 3. Compare with Content-Digest header value
	//
	// For demo purposes, we'll accept any non-empty digest

	if contentDigest != "" {
		logger.Debug("Content digest verification: PASSED")
		return true
	}

	logger.Debug("Content digest verification: FAILED")
	return false
}

// ShouldReject determines if request should be rejected based on verification
func (v *Verifier) ShouldReject(result *types.VerificationResult) bool {
	// If SAGE is disabled, never reject
	if !v.config.IsSAGEEnabled() {
		return false
	}

	// In strict mode, reject invalid signatures
	if v.config.IsStrictMode() && !result.Valid {
		return true
	}

	// In non-strict mode, allow invalid signatures (but log them)
	return false
}

// GetStats returns verification statistics
func (v *Verifier) GetStats() *VerificationStats {
	return v.stats
}
