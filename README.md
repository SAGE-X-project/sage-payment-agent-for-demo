# SAGE Payment Agent (Demo)

**Status:** ✅ Production Ready
**Purpose:** Payment processing agent for SAGE protocol demonstration
**Version:** 1.0.0

---

## Overview

SAGE Payment Agent is a demonstration payment processing service that receives payment requests, validates them using RFC-9421 HTTP message signatures (when SAGE protocol is enabled), and simulates blockchain transactions.

This agent demonstrates the **security benefits of the SAGE protocol** by:
- ✅ Detecting message tampering when SAGE is ON
- ⚠️ Processing manipulated messages when SAGE is OFF

---

## Features

### Core Functionality
- **HTTP API Server** - Receives payment requests via REST API
- **SAGE Protocol Support** - RFC-9421 signature verification (ON/OFF toggle)
- **Transaction Simulation** - Simulates blockchain payment transactions
- **Attack Detection** - Logs detected manipulation attempts
- **Health Monitoring** - Health check and status endpoints

### Security Features (SAGE ON)
- 🔒 **Message Signature Verification** - RFC-9421 HTTP signatures
- 🔒 **DID Resolution** - Blockchain-based identity verification
- 🔒 **Replay Attack Prevention** - Timestamp and nonce validation
- 🔒 **Tampering Detection** - Cryptographic integrity checks

### Demo Mode (SAGE OFF)
- ⚠️ **No Signature Verification** - Processes any incoming request
- ⚠️ **Vulnerable to MITM** - Susceptible to man-in-the-middle attacks
- ⚠️ **No Integrity Checks** - Cannot detect message manipulation

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Payment Agent                             │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │   Handlers   │───▶│     SAGE     │───▶│ Transaction  │  │
│  │              │    │   Verifier   │    │  Simulator   │  │
│  └──────────────┘    └──────────────┘    └──────────────┘  │
│         │                    │                    │          │
│         │                    │                    │          │
│         ▼                    ▼                    ▼          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │    Logger    │    │    Config    │    │    Types     │  │
│  └──────────────┘    └──────────────┘    └──────────────┘  │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

---

## Installation

### Prerequisites
- Go 1.21 or higher
- Git

### Build from Source

```bash
cd sage-payment-agent-for-demo
go mod init github.com/sage-x-project/sage-payment-agent-for-demo
go build -o payment-agent
```

---

## Configuration

Configure the agent using environment variables:

### Server Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `AGENT_PORT` | `8091` | HTTP server port |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |

### SAGE Protocol Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `SAGE_ENABLED` | `true` | Enable SAGE protocol verification |
| `SAGE_STRICT_MODE` | `true` | Reject requests with invalid signatures |
| `BLOCKCHAIN_RPC_URL` | `http://localhost:8545` | Ethereum RPC endpoint |
| `CONTRACT_ADDRESS` | `0x...` | DID registry contract address |

### Transaction Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `TX_SIMULATION_MODE` | `true` | Simulate transactions (don't send real txs) |
| `TX_DELAY_MS` | `500` | Simulated transaction delay |
| `WALLET_ADDRESS` | `0x...` | Agent's wallet address |

---

## Usage

### Running the Agent

#### SAGE Protocol ON (Secure Mode)
```bash
export SAGE_ENABLED=true
export SAGE_STRICT_MODE=true
./payment-agent
```

Expected output:
```
[INFO] Starting SAGE Payment Agent
[INFO] SAGE Protocol: ENABLED (RFC-9421)
[INFO] Listening on :8091
```

#### SAGE Protocol OFF (Demo Attack Mode)
```bash
export SAGE_ENABLED=false
./payment-agent
```

Expected output:
```
[INFO] Starting SAGE Payment Agent
[WARN] SAGE Protocol: DISABLED (Vulnerable to attacks)
[INFO] Listening on :8091
```

---

## API Endpoints

### POST /payment
Process a payment request.

**Request (SAGE OFF):**
```bash
curl -X POST http://localhost:8091/payment \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100.0,
    "currency": "USDC",
    "product": "Sunglasses",
    "recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
  }'
```

**Request (SAGE ON):**
```bash
curl -X POST http://localhost:8091/payment \
  -H "Content-Type: application/json" \
  -H "Signature-Input: sig1=(\"@method\" \"@authority\" \"@path\" \"content-digest\");created=1690000000" \
  -H "Signature: sig1=:MEUCIQDx...:" \
  -H "Content-Digest: sha-256=:X48E9...:" \
  -d '{
    "amount": 100.0,
    "currency": "USDC",
    "product": "Sunglasses",
    "recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
  }'
```

**Response (Success):**
```json
{
  "success": true,
  "transaction_id": "0x1234567890abcdef...",
  "amount": 100.0,
  "recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "status": "confirmed",
  "timestamp": "2025-10-14T02:00:00Z"
}
```

**Response (SAGE Verification Failed):**
```json
{
  "success": false,
  "error": "signature_verification_failed",
  "message": "Invalid RFC-9421 signature",
  "details": {
    "signature_valid": false,
    "reason": "Public key not found in DID registry"
  }
}
```

### GET /health
Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "sage_enabled": true,
  "uptime_seconds": 3600
}
```

### GET /status
Detailed status endpoint.

**Response:**
```json
{
  "agent": "sage-payment-agent",
  "version": "1.0.0",
  "sage_protocol": {
    "enabled": true,
    "strict_mode": true,
    "verified_requests": 1234,
    "rejected_requests": 5
  },
  "transactions": {
    "total": 1234,
    "successful": 1229,
    "failed": 5
  }
}
```

---

## Demo Scenarios

### Scenario 1: Normal Payment (SAGE ON)
```bash
# Gateway sends signed request → Payment Agent verifies → Transaction succeeds
SAGE_ENABLED=true ./payment-agent
```

Expected behavior:
- ✅ Signature verification passes
- ✅ Transaction processed
- ✅ Original amount/recipient used

### Scenario 2: Manipulated Payment (SAGE ON)
```bash
# Infected Gateway modifies request → Payment Agent detects tampering
SAGE_ENABLED=true ./payment-agent
```

Expected behavior:
- ❌ Signature verification fails
- ❌ Transaction rejected
- 🔒 Attack logged and blocked

### Scenario 3: Manipulated Payment (SAGE OFF)
```bash
# Infected Gateway modifies request → Payment Agent accepts
SAGE_ENABLED=false ./payment-agent
```

Expected behavior:
- ⚠️ No signature verification
- ✅ Transaction processed
- 💰 Modified amount/recipient used (ATTACK SUCCESS)

---

## Testing

### Run All Tests
```bash
go test ./...
```

### Run Tests with Coverage
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run Specific Package Tests
```bash
go test ./handlers -v
go test ./sage -v
```

**Target Coverage:** 90%+

---

## Logging

### Log Levels
- **DEBUG**: Detailed request/response data
- **INFO**: Normal operations
- **WARN**: SAGE disabled warnings
- **ERROR**: Processing errors
- **ATTACK**: Detected manipulation attempts

### Example Logs

**SAGE ON - Valid Request:**
```
[INFO] Incoming payment request: $100.0 to 0x742d35Cc
[DEBUG] Verifying RFC-9421 signature...
[INFO] Signature verification: PASSED
[INFO] Processing transaction...
[INFO] Transaction confirmed: 0xabc123...
```

**SAGE ON - Tampered Request:**
```
[INFO] Incoming payment request: $10000.0 to 0xATTACKER
[DEBUG] Verifying RFC-9421 signature...
[ERROR] Signature verification: FAILED
[ATTACK] TAMPERING DETECTED: Amount modified from $100 to $10000
[ATTACK] TAMPERING DETECTED: Recipient changed from 0x742d35Cc to 0xATTACKER
[ERROR] Request rejected: Invalid signature
```

**SAGE OFF - Tampered Request:**
```
[WARN] SAGE Protocol disabled - no verification
[INFO] Incoming payment request: $10000.0 to 0xATTACKER
[WARN] Processing unverified request
[INFO] Transaction confirmed: 0xdef456... (COMPROMISED!)
```

---

## Integration with Gateway

### Normal Flow (SAGE ON)
```
User → Gateway (signs) → Payment Agent (verifies) → Blockchain
         RFC-9421                RFC-9421              Real TX
```

### Attack Flow (SAGE OFF)
```
User → Gateway (modifies) → Payment Agent (accepts) → Blockchain
         💰×100                  ⚠️ No check           Wrong TX!
```

---

## Development

### Project Structure
```
sage-payment-agent-for-demo/
├── main.go                 # Server entry point
├── go.mod                  # Go module definition
├── README.md               # This file
├── .gitignore              # Git ignore patterns
├── types/                  # Message type definitions
│   ├── message.go
│   └── message_test.go
├── config/                 # Configuration management
│   ├── config.go
│   └── config_test.go
├── logger/                 # Logging utilities
│   ├── logger.go
│   └── logger_test.go
├── handlers/               # HTTP request handlers
│   ├── payment.go
│   ├── health.go
│   └── *_test.go
├── sage/                   # SAGE protocol integration
│   ├── verifier.go         # RFC-9421 signature verification
│   ├── did.go              # DID resolution
│   └── *_test.go
└── transaction/            # Transaction simulation
    ├── simulator.go
    └── simulator_test.go
```

### Adding New Features

1. **Add new payment type:**
   - Update `types/message.go`
   - Add validation in `handlers/payment.go`
   - Add test cases

2. **Add new verification method:**
   - Implement in `sage/verifier.go`
   - Update configuration options
   - Add integration tests

---

## Troubleshooting

### Signature Verification Always Fails

**Problem:** All requests rejected even with valid signatures

**Solution:**
```bash
# Check DID registry connection
curl http://localhost:8545  # Ethereum RPC should respond

# Verify contract address
export CONTRACT_ADDRESS=0x... # Correct address

# Check logs for specific error
export LOG_LEVEL=debug
./payment-agent
```

### Agent Not Starting

**Problem:** Port already in use

**Solution:**
```bash
# Use different port
export AGENT_PORT=8092
./payment-agent

# Or kill existing process
lsof -ti:8091 | xargs kill
```

---

## Security Considerations

### Production Deployment

⚠️ **This is a DEMO implementation. For production use:**

1. **Use real private keys** - Store securely in HSM or key vault
2. **Enable TLS/HTTPS** - Encrypt all network traffic
3. **Implement rate limiting** - Prevent DoS attacks
4. **Add authentication** - API key or OAuth2
5. **Use real blockchain** - Connect to mainnet with proper key management
6. **Monitor transactions** - Set up alerts for unusual activity
7. **Regular audits** - Security and code reviews

### Demo Limitations

This implementation is for **demonstration purposes only**:
- ❌ Simplified signature verification
- ❌ Simulated blockchain transactions
- ❌ No rate limiting
- ❌ No authentication/authorization
- ❌ Limited error handling

---

## License

MIT License - See LICENSE file for details

---

## Related Projects

- **sage/** - Core SAGE protocol library (RFC-9421, DID, crypto)
- **sage-gateway-infected-for-demo/** - Man-in-the-middle attack simulator
- **sage-multi-agent/** - Multi-agent orchestration system
- **sage-fe/** - Frontend dashboard

---

## Support

For questions or issues:
- GitHub Issues: [sage-x-project/sage](https://github.com/sage-x-project/sage)
- Documentation: [SAGE Protocol Docs](https://sage-protocol.dev)

---

**Built for SAGE Open Source Competition 2025**
