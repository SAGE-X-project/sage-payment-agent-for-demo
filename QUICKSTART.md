# Quick Start Guide - Sepolia Testnet

This guide will help you get the SAGE Payment Agent up and running with the deployed Sepolia testnet contracts in under 5 minutes.

## Prerequisites

- Go 1.21 or higher
- An RPC endpoint (Infura, Alchemy, or your own node)
- Sepolia testnet ETH (get from faucet if needed)

---

## Step 1: Setup Configuration

1. **Copy the Sepolia configuration template:**
   ```bash
   cp .env.sepolia .env
   ```

2. **Edit `.env` and update the following:**

   **Required:**
   ```bash
   # Get your API key from:
   # - Infura: https://infura.io
   # - Alchemy: https://alchemy.com
   BLOCKCHAIN_RPC_URL=https://sepolia.infura.io/v3/YOUR_INFURA_PROJECT_ID

   # Or use Alchemy:
   # BLOCKCHAIN_RPC_URL=https://eth-sepolia.g.alchemy.com/v2/YOUR_ALCHEMY_API_KEY
   ```

   **Optional (for real transactions):**
   ```bash
   # Your agent's wallet address (only needed if TX_SIMULATION_MODE=false)
   WALLET_ADDRESS=0xYourWalletAddress
   ```

   **Contract addresses are already configured:**
   ```bash
   IDENTITY_REGISTRY_ADDRESS=0x5B0763c3649eee889966dF478a73e53Df0420C84
   VALIDATION_REGISTRY_ADDRESS=0x97291e2D3023d166878ed45BBD176F92E5Fda098
   REPUTATION_REGISTRY_ADDRESS=0xE953B278fd2378BA4987FE07f71575dd3353C9a8
   ```

---

## Step 2: Get Sepolia Test ETH (Optional)

Only needed if you plan to send real transactions (`TX_SIMULATION_MODE=false`):

- **Alchemy Faucet:** https://www.alchemy.com/faucets/ethereum-sepolia
- **Sepolia Faucet:** https://sepoliafaucet.com
- **Infura Faucet:** https://www.infura.io/faucet/sepolia

For testing, simulation mode is enabled by default, so you don't need test ETH initially.

---

## Step 3: Build and Run

**Option A: Using the convenience script (Recommended)**
```bash
./run-sepolia.sh
```

This script will:
- ✅ Check your configuration
- ✅ Verify RPC connection
- ✅ Confirm contract deployments
- ✅ Build the agent if needed
- ✅ Start the agent with Sepolia configuration

**Option B: Manual build and run**
```bash
# Build
go build -o payment-agent

# Load environment and run
export $(cat .env | xargs)
./payment-agent
```

---

## Step 4: Verify It's Working

In a new terminal:

```bash
# Check health
curl http://localhost:8091/health

# Check status
curl http://localhost:8091/status

# Test payment (simulation mode)
curl -X POST http://localhost:8091/payment \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100.0,
    "currency": "USDC",
    "product": "Test Product",
    "recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "transaction_id": "0x...",
  "amount": 100.0,
  "recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "status": "confirmed"
}
```

Or if SAGE verification is enabled:
```json
{
  "success": false,
  "error": "signature_verification_required",
  "message": "RFC-9421 signature required when SAGE is enabled"
}
```

---

## Step 5: Run Integration Tests

Run the automated test suite:

```bash
./test-sepolia.sh
```

This will test:
- ✅ Agent health and status endpoints
- ✅ Payment processing
- ✅ Error handling
- ✅ Contract deployments
- ✅ RPC connectivity

---

## Configuration Options

### Enable/Disable SAGE Protocol

**SAGE ON (Secure Mode - Default):**
```bash
export SAGE_ENABLED=true
export SAGE_STRICT_MODE=true
./payment-agent
```
- ✅ Requires RFC-9421 signatures
- ✅ Validates DID and public keys
- ✅ Detects message tampering

**SAGE OFF (Demo Attack Mode):**
```bash
export SAGE_ENABLED=false
./payment-agent
```
- ⚠️ No signature verification
- ⚠️ Processes any request
- ⚠️ Vulnerable to MITM attacks

### Switch Between Simulation and Real Transactions

**Simulation Mode (Default - Recommended for testing):**
```bash
export TX_SIMULATION_MODE=true
```
- Fast transaction processing
- No gas costs
- Perfect for testing

**Real Transaction Mode:**
```bash
export TX_SIMULATION_MODE=false
export WALLET_ADDRESS=0xYourWalletAddress
```
- Sends real transactions to Sepolia
- Requires test ETH
- Use with caution

---

## Deployed Contract Information

All contracts are deployed on **Sepolia Testnet** and verified on Etherscan:

### ERC8004 Identity Registry
- **Address:** `0x5B0763c3649eee889966dF478a73e53Df0420C84`
- **Purpose:** Maps DIDs to public keys for signature verification
- **Etherscan:** https://sepolia.etherscan.io/address/0x5B0763c3649eee889966dF478a73e53Df0420C84

### ERC8004 Validation Registry
- **Address:** `0x97291e2D3023d166878ed45BBD176F92E5Fda098`
- **Purpose:** Manages validator registration and consensus
- **Etherscan:** https://sepolia.etherscan.io/address/0x97291e2D3023d166878ed45BBD176F92E5Fda098

### ERC8004 Reputation Registry
- **Address:** `0xE953B278fd2378BA4987FE07f71575dd3353C9a8`
- **Purpose:** Tracks validator reputation scores
- **Etherscan:** https://sepolia.etherscan.io/address/0xE953B278fd2378BA4987FE07f71575dd3353C9a8

---

## Troubleshooting

### Issue: "Failed to connect to RPC endpoint"

**Solution:**
1. Check your `BLOCKCHAIN_RPC_URL` in `.env`
2. Verify your API key is correct
3. Test the connection:
   ```bash
   curl -X POST YOUR_RPC_URL \
     -H "Content-Type: application/json" \
     -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
   ```

### Issue: "Contract not found at address"

**Solution:**
1. Verify you're using Sepolia network (Chain ID: 11155111)
2. Check contract addresses in `.env` match:
   - Identity: `0x5B0763c3649eee889966dF478a73e53Df0420C84`
   - Validation: `0x97291e2D3023d166878ed45BBD176F92E5Fda098`
   - Reputation: `0xE953B278fd2378BA4987FE07f71575dd3353C9a8`

### Issue: "Port already in use"

**Solution:**
```bash
# Use different port
export AGENT_PORT=8092
./payment-agent

# Or kill existing process
lsof -ti:8091 | xargs kill
```

### Issue: "Signature verification always fails"

**Solution:**
1. Ensure agent DID is registered in Identity Registry
2. Check public key format is correct
3. Verify signature follows RFC-9421 format
4. For testing, you can disable SAGE:
   ```bash
   export SAGE_ENABLED=false
   ```

---

## Next Steps

1. **Explore the API:**
   - Read the full [API documentation](README.md#api-endpoints)
   - Test different payment scenarios

2. **Integration:**
   - Integrate with SAGE Gateway
   - Test with the infected gateway demo
   - Build your own client

3. **Deep Dive:**
   - Read [DEPLOYMENT.md](DEPLOYMENT.md) for detailed contract info
   - Explore the [sage/](sage/) directory for SAGE protocol implementation
   - Check [TEST_REPORT.md](TEST_REPORT.md) for test coverage

4. **Development:**
   - Run tests: `go test ./...`
   - Add custom handlers
   - Extend SAGE verification logic

---

## Useful Commands

```bash
# Build the agent
go build -o payment-agent

# Run with custom port
AGENT_PORT=8092 ./payment-agent

# Run with debug logging
LOG_LEVEL=debug ./payment-agent

# Run tests
go test ./... -v

# Check code coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Format code
go fmt ./...

# Run linter (if installed)
golangci-lint run
```

---

## Getting Help

- **Documentation:** See [README.md](README.md) for complete documentation
- **Contract Details:** See [DEPLOYMENT.md](DEPLOYMENT.md)
- **Test Reports:** See [TEST_REPORT.md](TEST_REPORT.md)
- **Issues:** [GitHub Issues](https://github.com/sage-x-project/sage/issues)

---

## Quick Reference

### Environment Variables
| Variable | Default | Description |
|----------|---------|-------------|
| `AGENT_PORT` | `8091` | HTTP server port |
| `SAGE_ENABLED` | `true` | Enable SAGE protocol verification |
| `TX_SIMULATION_MODE` | `true` | Simulate transactions |
| `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |

### API Endpoints
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/payment` | POST | Process payment request |
| `/health` | GET | Health check |
| `/status` | GET | Detailed status |
| `/` | GET | Agent info |

### Contract Addresses (Sepolia)
| Contract | Address |
|----------|---------|
| Identity Registry | `0x5B0763c3649eee889966dF478a73e53Df0420C84` |
| Validation Registry | `0x97291e2D3023d166878ed45BBD176F92E5Fda098` |
| Reputation Registry | `0xE953B278fd2378BA4987FE07f71575dd3353C9a8` |

---

**You're all set! 🚀**

The SAGE Payment Agent is now running and connected to Sepolia testnet contracts.
