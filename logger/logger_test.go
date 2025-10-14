package logger

import (
	"bytes"
	"log"
	"os"
	"testing"
	"time"

	"github.com/sage-x-project/sage-payment-agent-for-demo/types"
)

func TestSetLogLevel(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectedLevel LogLevel
	}{
		{"Debug level", "debug", DEBUG},
		{"Info level", "info", INFO},
		{"Warn level", "warn", WARN},
		{"Error level", "error", ERROR},
		{"Attack level", "attack", ATTACK},
		{"Unknown level", "unknown", INFO},
		{"Empty string", "", INFO},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLogLevel(tt.level)
			if logLevel != tt.expectedLevel {
				t.Errorf("SetLogLevel(%s): got %v, want %v", tt.level, logLevel, tt.expectedLevel)
			}
		})
	}

	// Reset to default
	SetLogLevel("info")
}

func TestDebug(t *testing.T) {
	var buf bytes.Buffer
	debugLogger = log.New(&buf, "[DEBUG] ", log.Ldate|log.Ltime)

	logLevel = DEBUG
	Debug("Test debug message: %s", "hello")

	output := buf.String()
	if len(output) == 0 {
		t.Error("Debug() should produce output when log level is DEBUG")
	}
	if !bytes.Contains(buf.Bytes(), []byte("Test debug message: hello")) {
		t.Errorf("Debug() output doesn't contain expected message: %s", output)
	}

	// Test with higher log level
	buf.Reset()
	logLevel = INFO
	Debug("Should not appear")

	if buf.Len() > 0 {
		t.Error("Debug() should not produce output when log level is INFO")
	}

	// Reset
	debugLogger = log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime)
	logLevel = INFO
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer
	infoLogger = log.New(&buf, "[INFO] ", log.Ldate|log.Ltime)

	logLevel = INFO
	Info("Test info message: %d", 123)

	output := buf.String()
	if len(output) == 0 {
		t.Error("Info() should produce output when log level is INFO")
	}
	if !bytes.Contains(buf.Bytes(), []byte("Test info message: 123")) {
		t.Errorf("Info() output doesn't contain expected message: %s", output)
	}

	// Reset
	infoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
}

func TestWarn(t *testing.T) {
	var buf bytes.Buffer
	infoLogger = log.New(&buf, "[INFO] ", log.Ldate|log.Ltime)

	logLevel = WARN
	Warn("Test warn message")

	output := buf.String()
	if len(output) == 0 {
		t.Error("Warn() should produce output when log level is WARN")
	}
	if !bytes.Contains(buf.Bytes(), []byte("[WARN]")) {
		t.Errorf("Warn() output doesn't contain [WARN] prefix: %s", output)
	}

	// Reset
	infoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	logLevel = INFO
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	errorLogger = log.New(&buf, "[ERROR] ", log.Ldate|log.Ltime)

	logLevel = ERROR
	Error("Test error message")

	output := buf.String()
	if len(output) == 0 {
		t.Error("Error() should produce output when log level is ERROR")
	}
	if !bytes.Contains(buf.Bytes(), []byte("Test error message")) {
		t.Errorf("Error() output doesn't contain expected message: %s", output)
	}

	// Reset
	errorLogger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime)
}

func TestLogPaymentRequest(t *testing.T) {
	var buf bytes.Buffer
	infoLogger = log.New(&buf, "[INFO] ", log.Ldate|log.Ltime)
	debugLogger = log.New(&buf, "[DEBUG] ", log.Ldate|log.Ltime)

	logLevel = DEBUG

	req := &types.PaymentRequest{
		Amount:      100.0,
		Currency:    "USDC",
		Product:     "Sunglasses",
		Recipient:   "0x742d35Cc",
		Description: "Test payment",
	}

	LogPaymentRequest(req)

	if !bytes.Contains(buf.Bytes(), []byte("$100.00")) {
		t.Error("LogPaymentRequest should log amount")
	}
	if !bytes.Contains(buf.Bytes(), []byte("0x742d35Cc")) {
		t.Error("LogPaymentRequest should log recipient")
	}

	// Reset
	infoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	debugLogger = log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime)
	logLevel = INFO
}

func TestLogTransaction(t *testing.T) {
	var buf bytes.Buffer
	infoLogger = log.New(&buf, "[INFO] ", log.Ldate|log.Ltime)
	debugLogger = log.New(&buf, "[DEBUG] ", log.Ldate|log.Ltime)

	logLevel = DEBUG

	tx := &types.TransactionResult{
		TxHash:      "0xabc123",
		From:        "0x111",
		To:          "0x222",
		Amount:      100.0,
		Status:      "confirmed",
		Timestamp:   time.Now(),
		BlockNumber: 12345,
	}

	LogTransaction(tx)

	if !bytes.Contains(buf.Bytes(), []byte("confirmed")) {
		t.Error("LogTransaction should log status")
	}
	if !bytes.Contains(buf.Bytes(), []byte("0xabc123")) {
		t.Error("LogTransaction should log tx hash")
	}

	// Reset
	infoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	debugLogger = log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime)
	logLevel = INFO
}

func TestLogVerification(t *testing.T) {
	var buf bytes.Buffer
	infoLogger = log.New(&buf, "[INFO] ", log.Ldate|log.Ltime)
	errorLogger = log.New(&buf, "[ERROR] ", log.Ldate|log.Ltime)

	// Test valid verification
	buf.Reset()
	result := &types.VerificationResult{
		Valid:     true,
		SignerDID: "did:sage:0x123",
	}
	LogVerification(result)

	if !bytes.Contains(buf.Bytes(), []byte("PASSED")) {
		t.Error("LogVerification should log PASSED for valid signature")
	}

	// Test invalid verification
	buf.Reset()
	result = &types.VerificationResult{
		Valid:        false,
		ErrorMessage: "Invalid signature",
	}
	LogVerification(result)

	if !bytes.Contains(buf.Bytes(), []byte("FAILED")) {
		t.Error("LogVerification should log FAILED for invalid signature")
	}

	// Reset
	infoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	errorLogger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime)
}

func TestLogAttack(t *testing.T) {
	var buf bytes.Buffer
	attackLogger = log.New(&buf, "[ATTACK] ", log.Ldate|log.Ltime)

	attack := &types.AttackLog{
		Timestamp:  time.Now(),
		AttackType: "price_manipulation",
		Changes: []types.Change{
			{
				Field:         "amount",
				ExpectedValue: 100.0,
				ActualValue:   10000.0,
			},
		},
		Blocked:  true,
		SourceIP: "192.168.1.1",
	}

	LogAttack(attack)

	if !bytes.Contains(buf.Bytes(), []byte("TAMPERING DETECTED")) {
		t.Error("LogAttack should log tampering header")
	}
	if !bytes.Contains(buf.Bytes(), []byte("price_manipulation")) {
		t.Error("LogAttack should log attack type")
	}
	if !bytes.Contains(buf.Bytes(), []byte("BLOCKED")) {
		t.Error("LogAttack should log blocked status")
	}

	// Reset
	attackLogger = log.New(os.Stdout, "[ATTACK] ", log.Ldate|log.Ltime)
}

func TestLogAttackSimple(t *testing.T) {
	var buf bytes.Buffer
	attackLogger = log.New(&buf, "[ATTACK] ", log.Ldate|log.Ltime)

	LogAttackSimple("Simple attack message: %s", "test")

	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("Simple attack message: test")) {
		t.Errorf("LogAttackSimple() output doesn't contain expected message: %s", output)
	}

	// Reset
	attackLogger = log.New(os.Stdout, "[ATTACK] ", log.Ldate|log.Ltime)
}

func TestLogSAGEEnabled(t *testing.T) {
	// Just test it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogSAGEEnabled() panicked: %v", r)
		}
	}()

	LogSAGEEnabled()
}

func TestLogSAGEDisabled(t *testing.T) {
	// Just test it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogSAGEDisabled() panicked: %v", r)
		}
	}()

	LogSAGEDisabled()
}

func TestLogStartup(t *testing.T) {
	// Just test it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogStartup() panicked: %v", r)
		}
	}()

	LogStartup("8091", true)
	LogStartup("8091", false)
}

func TestLogLevel_Hierarchy(t *testing.T) {
	if DEBUG >= INFO {
		t.Error("DEBUG should be less than INFO")
	}
	if INFO >= WARN {
		t.Error("INFO should be less than WARN")
	}
	if WARN >= ERROR {
		t.Error("WARN should be less than ERROR")
	}
	if ERROR >= ATTACK {
		t.Error("ERROR should be less than ATTACK")
	}
}
