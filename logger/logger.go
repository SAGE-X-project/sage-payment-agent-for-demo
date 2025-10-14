package logger

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sage-x-project/sage-payment-agent-for-demo/types"
)

// LogLevel represents logging levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	ATTACK
)

var (
	debugLogger  = log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime)
	infoLogger   = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	warnLogger   = log.New(os.Stdout, "[WARN] ", log.Ldate|log.Ltime)
	errorLogger  = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime)
	attackLogger = log.New(os.Stdout, "[ATTACK] ", log.Ldate|log.Ltime)
	logLevel     = INFO
)

// SetLogLevel sets the minimum log level
func SetLogLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		logLevel = DEBUG
	case "info":
		logLevel = INFO
	case "warn", "warning":
		logLevel = WARN
	case "error":
		logLevel = ERROR
	case "attack":
		logLevel = ATTACK
	default:
		logLevel = INFO
	}
}

// Debug logs debug messages
func Debug(format string, v ...interface{}) {
	if logLevel <= DEBUG {
		debugLogger.Printf(format, v...)
	}
}

// Info logs informational messages
func Info(format string, v ...interface{}) {
	if logLevel <= INFO {
		infoLogger.Printf(format, v...)
	}
}

// Warn logs warning messages
func Warn(format string, v ...interface{}) {
	if logLevel <= WARN {
		infoLogger.Printf("[WARN] "+format, v...)
	}
}

// Error logs error messages
func Error(format string, v ...interface{}) {
	if logLevel <= ERROR {
		errorLogger.Printf(format, v...)
	}
}

// LogPaymentRequest logs incoming payment request
func LogPaymentRequest(req *types.PaymentRequest) {
	Info("Incoming payment request: $%.2f to %s", req.Amount, req.Recipient)
	if req.Product != "" {
		Debug("Product: %s", req.Product)
	}
	if req.Description != "" {
		Debug("Description: %s", req.Description)
	}
}

// LogTransaction logs transaction details
func LogTransaction(tx *types.TransactionResult) {
	Info("Transaction: %s", tx.Status)
	Info("TX Hash: %s", tx.TxHash)
	Debug("From: %s", tx.From)
	Debug("To: %s", tx.To)
	Debug("Amount: $%.2f", tx.Amount)
	if tx.BlockNumber > 0 {
		Debug("Block: %d", tx.BlockNumber)
	}
}

// LogVerification logs signature verification result
func LogVerification(result *types.VerificationResult) {
	if result.Valid {
		Info("Signature verification: PASSED")
		if result.SignerDID != "" {
			Debug("Signer DID: %s", result.SignerDID)
		}
	} else {
		Error("Signature verification: FAILED")
		if result.ErrorMessage != "" {
			Error("Reason: %s", result.ErrorMessage)
		}
	}
}

// LogAttack logs detected attack attempts
func LogAttack(attack *types.AttackLog) {
	attackLogger.Println("===== TAMPERING DETECTED =====")
	attackLogger.Printf("Type: %s", attack.AttackType)
	attackLogger.Printf("Timestamp: %s", attack.Timestamp.Format("2006-01-02 15:04:05"))

	if attack.SourceIP != "" {
		attackLogger.Printf("Source IP: %s", attack.SourceIP)
	}

	if len(attack.Changes) > 0 {
		attackLogger.Println("Detected Changes:")
		for _, change := range attack.Changes {
			attackLogger.Printf("  - Field: %s", change.Field)
			attackLogger.Printf("    Expected: %v", change.ExpectedValue)
			attackLogger.Printf("    Actual: %v", change.ActualValue)
		}
	}

	if attack.Blocked {
		attackLogger.Println("Status: BLOCKED")
	} else {
		attackLogger.Println("Status: NOT BLOCKED (SAGE disabled)")
	}
	attackLogger.Println("==============================")
}

// LogAttackSimple logs a simple attack message
func LogAttackSimple(format string, v ...interface{}) {
	attackLogger.Printf(format, v...)
}

// LogSAGEEnabled logs SAGE protocol banner
func LogSAGEEnabled() {
	fmt.Println("╔════════════════════════════════════════════════════╗")
	fmt.Println("║         SAGE PROTOCOL ENABLED (RFC-9421)           ║")
	fmt.Println("║                                                    ║")
	fmt.Println("║  ✓ Signature Verification: ON                      ║")
	fmt.Println("║  ✓ DID Resolution: ON                              ║")
	fmt.Println("║  ✓ Tampering Detection: ON                         ║")
	fmt.Println("║  ✓ Security: MAXIMUM                               ║")
	fmt.Println("╚════════════════════════════════════════════════════╝")
}

// LogSAGEDisabled logs SAGE disabled warning banner
func LogSAGEDisabled() {
	fmt.Println("╔════════════════════════════════════════════════════╗")
	fmt.Println("║       ⚠️  SAGE PROTOCOL DISABLED  ⚠️                ║")
	fmt.Println("║                                                    ║")
	fmt.Println("║  ✗ Signature Verification: OFF                     ║")
	fmt.Println("║  ✗ DID Resolution: OFF                             ║")
	fmt.Println("║  ✗ Tampering Detection: OFF                        ║")
	fmt.Println("║  ⚠️  VULNERABLE TO ATTACKS                          ║")
	fmt.Println("╚════════════════════════════════════════════════════╝")
}

// LogStartup logs startup information
func LogStartup(port string, sageEnabled bool) {
	Info("Starting SAGE Payment Agent")
	Info("Version: 1.0.0")
	Info("Port: %s", port)
	if sageEnabled {
		LogSAGEEnabled()
	} else {
		LogSAGEDisabled()
	}
	Info("Ready to process payments")
}
