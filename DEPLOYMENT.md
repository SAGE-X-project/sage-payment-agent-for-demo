# ERC8004 Contract Deployment - Sepolia Testnet

## Deployment Summary

**Date:** 2025-11-03
**Network:** Sepolia Testnet (Chain ID: 11155111)
**Total Gas Used:** 4,051,526
**Deployer:** [Your Deployer Address]

---

## Deployed Contracts

### 1. ERC8004ValidationRegistry

**Purpose:** Manages validator registration, staking, and consensus mechanisms for the SAGE protocol.

```
Address: 0x97291e2D3023d166878ed45BBD176F92E5Fda098
```

**Constructor Arguments:**
- `minStake`: 0.01 ETH (10000000000000000 wei)
- `minValidators`: 3
- `consensusThreshold`: 66% (66)

**Gas Used:** 2,042,163

**Verification Status:**
- ✅ Etherscan: Exact Match
- ✅ Blockscout: Verified

**Links:**
- Etherscan: https://sepolia.etherscan.io/address/0x97291e2D3023d166878ed45BBD176F92E5Fda098
- Blockscout: https://sepolia.blockscout.com/address/0x97291e2D3023d166878ed45BBD176F92E5Fda098

**Key Features:**
- Validator registration with ETH staking
- Minimum 3 validators required
- 66% consensus threshold for validation
- Stake slashing for malicious behavior
- Validator reputation tracking

---

### 2. ERC8004ReputationRegistry

**Purpose:** Tracks and manages validator reputation scores based on validation history.

```
Address: 0xE953B278fd2378BA4987FE07f71575dd3353C9a8
```

**Constructor Arguments:**
- `validationRegistry`: 0x97291e2D3023d166878ed45BBD176F92E5Fda098

**Gas Used:** 927,953

**Verification Status:**
- ✅ Etherscan: Exact Match
- ✅ Blockscout: Verified

**Links:**
- Etherscan: https://sepolia.etherscan.io/address/0xE953B278fd2378BA4987FE07f71575dd3353C9a8
- Blockscout: https://sepolia.blockscout.com/address/0xE953B278fd2378BA4987FE07f71575dd3353C9a8

**Key Features:**
- Reputation scoring system
- Linked to ValidationRegistry
- Rewards good validators
- Penalizes malicious validators
- Historical reputation tracking

---

### 3. ERC8004IdentityRegistry (DID Registry)

**Purpose:** Maps Decentralized Identifiers (DIDs) to public keys and metadata for agent identity verification.

```
Address: 0x5B0763c3649eee889966dF478a73e53Df0420C84
```

**Constructor Arguments:** None

**Gas Used:** 1,081,410

**Verification Status:**
- ✅ Etherscan: Exact Match
- ✅ Blockscout: Verified

**Links:**
- Etherscan: https://sepolia.etherscan.io/address/0x5B0763c3649eee889966dF478a73e53Df0420C84
- Blockscout: https://sepolia.blockscout.com/address/0x5B0763c3649eee889966dF478a73e53Df0420C84

**Key Features:**
- DID-to-public-key mapping
- Agent identity registration
- Public key resolution for signature verification
- Metadata storage
- Used by SAGE protocol for RFC-9421 signature verification

---

## Contract Interactions

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      SAGE Protocol Flow                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Payment Agent                                                    │
│       │                                                           │
│       ├──▶ ERC8004IdentityRegistry                              │
│       │    └── Resolve agent DID → public key                   │
│       │    └── Verify RFC-9421 signature                        │
│       │                                                           │
│       ├──▶ ERC8004ValidationRegistry                            │
│       │    └── Check validator status                            │
│       │    └── Verify consensus                                  │
│       │                                                           │
│       └──▶ ERC8004ReputationRegistry                            │
│            └── Query validator reputation                        │
│            └── Update reputation after validation                │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### Typical Flow

1. **Agent Registration**
   ```
   User → ERC8004IdentityRegistry.registerDID(did, publicKey, metadata)
   ```

2. **Validator Registration**
   ```
   Validator → ERC8004ValidationRegistry.registerValidator() + 0.01 ETH stake
   ```

3. **Message Signature Verification**
   ```
   Payment Agent → ERC8004IdentityRegistry.resolvePublicKey(did)
                → Verify RFC-9421 signature with public key
   ```

4. **Consensus Validation**
   ```
   Payment Agent → ERC8004ValidationRegistry.validateConsensus()
                → ERC8004ReputationRegistry.updateReputation()
   ```

---

## Configuration

### Environment Setup

1. **Copy the Sepolia configuration:**
   ```bash
   cp .env.sepolia .env
   ```

2. **Update RPC endpoint:**
   ```bash
   # Using Infura
   BLOCKCHAIN_RPC_URL=https://sepolia.infura.io/v3/YOUR_PROJECT_ID

   # Or using Alchemy
   BLOCKCHAIN_RPC_URL=https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY
   ```

3. **Set your wallet address:**
   ```bash
   WALLET_ADDRESS=0xYourAgentWalletAddress
   ```

### Get Sepolia Test ETH

You'll need Sepolia ETH to interact with the contracts:

- **Alchemy Faucet:** https://www.alchemy.com/faucets/ethereum-sepolia
- **Sepolia Faucet:** https://sepoliafaucet.com
- **Infura Faucet:** https://www.infura.io/faucet/sepolia

---

## Testing the Deployment

### 1. Check Contract Deployment

```bash
# Install cast (from Foundry) if not already installed
curl -L https://foundry.paradigm.xyz | bash
foundryup

# Check ValidationRegistry
cast call 0x97291e2D3023d166878ed45BBD176F92E5Fda098 "minStake()(uint256)" \
  --rpc-url https://sepolia.infura.io/v3/YOUR_PROJECT_ID

# Should return: 10000000000000000 (0.01 ETH in wei)

# Check IdentityRegistry is deployed
cast code 0x5B0763c3649eee889966dF478a73e53Df0420C84 \
  --rpc-url https://sepolia.infura.io/v3/YOUR_PROJECT_ID

# Should return bytecode (not 0x)
```

### 2. Test Agent Integration

```bash
# Build the agent
go build -o payment-agent

# Set Sepolia configuration
export $(cat .env | xargs)

# Run the agent
./payment-agent
```

Expected output:
```
[INFO] Starting SAGE Payment Agent
[INFO] SAGE Protocol: ENABLED (RFC-9421)
[INFO] Network: Sepolia Testnet
[INFO] Identity Registry: 0x5B0763c3649eee889966dF478a73e53Df0420C84
[INFO] Validation Registry: 0x97291e2D3023d166878ed45BBD176F92E5Fda098
[INFO] Reputation Registry: 0xE953B278fd2378BA4987FE07f71575dd3353C9a8
[INFO] Listening on :8091
```

### 3. Test Health Check

```bash
curl http://localhost:8091/health
```

Expected response:
```json
{
  "status": "healthy",
  "sage_enabled": true,
  "network": "sepolia",
  "contracts": {
    "identity_registry": "0x5B0763c3649eee889966dF478a73e53Df0420C84",
    "validation_registry": "0x97291e2D3023d166878ed45BBD176F92E5Fda098",
    "reputation_registry": "0xE953B278fd2378BA4987FE07f71575dd3353C9a8"
  }
}
```

---

## Security Considerations

### Testnet Limitations

⚠️ **This is a TESTNET deployment:**

- ✅ Use for development and testing only
- ✅ Sepolia ETH has no real value
- ✅ Contracts can be redeployed if needed
- ❌ Do NOT use for production transactions
- ❌ Do NOT store real assets

### Production Deployment Checklist

When deploying to mainnet:

- [ ] Complete security audit
- [ ] Test all contract functions thoroughly
- [ ] Verify gas optimizations
- [ ] Set up monitoring and alerts
- [ ] Implement emergency pause mechanisms
- [ ] Configure multi-sig governance
- [ ] Test upgrade paths (if applicable)
- [ ] Document all admin functions
- [ ] Set appropriate access controls
- [ ] Verify all constructor parameters

---

## Troubleshooting

### Issue: RPC Connection Failed

**Error:** `failed to connect to RPC endpoint`

**Solution:**
```bash
# Test RPC connection
curl -X POST https://sepolia.infura.io/v3/YOUR_PROJECT_ID \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

# Should return current block number
```

### Issue: Contract Not Found

**Error:** `contract code not found at address`

**Solution:**
1. Verify you're using Sepolia network (Chain ID: 11155111)
2. Check contract address is correct
3. Confirm Sepolia ETH balance: https://sepolia.etherscan.io/address/YOUR_ADDRESS

### Issue: Signature Verification Fails

**Error:** `public key not found in DID registry`

**Solution:**
1. Ensure agent DID is registered in IdentityRegistry
2. Check public key is correctly formatted
3. Verify contract address in config matches deployed address

---

## Monitoring and Maintenance

### Block Explorers

- **Etherscan:** https://sepolia.etherscan.io
- **Blockscout:** https://sepolia.blockscout.com

### Monitor Contract Events

```bash
# Watch for validator registrations
cast logs "ValidatorRegistered(address,uint256)" \
  --address 0x97291e2D3023d166878ed45BBD176F92E5Fda098 \
  --rpc-url https://sepolia.infura.io/v3/YOUR_PROJECT_ID

# Watch for DID registrations
cast logs "DIDRegistered(string,address)" \
  --address 0x5B0763c3649eee889966dF478a73e53Df0420C84 \
  --rpc-url https://sepolia.infura.io/v3/YOUR_PROJECT_ID
```

### Contract State Queries

```bash
# Get validator count
cast call 0x97291e2D3023d166878ed45BBD176F92E5Fda098 \
  "getValidatorCount()(uint256)" \
  --rpc-url YOUR_RPC_URL

# Get validator reputation
cast call 0xE953B278fd2378BA4987FE07f71575dd3353C9a8 \
  "getReputation(address)(uint256)" \
  0xVALIDATOR_ADDRESS \
  --rpc-url YOUR_RPC_URL
```

---

## Next Steps

1. **Register Test Agents**
   - Create test DIDs
   - Register public keys in IdentityRegistry
   - Test signature verification

2. **Register Validators**
   - Set up validator nodes
   - Stake 0.01 ETH per validator
   - Test consensus mechanism

3. **Integration Testing**
   - Test full payment flow with SAGE protocol
   - Test attack scenarios (MITM, tampering)
   - Verify reputation updates

4. **Demo Preparation**
   - Configure infected gateway
   - Set up monitoring dashboard
   - Prepare demo scenarios

---

## Resources

### Documentation
- ERC8004 Spec: [Link to spec]
- SAGE Protocol: https://sage-protocol.dev
- RFC-9421 HTTP Signatures: https://www.rfc-editor.org/rfc/rfc9421

### Tools
- Foundry: https://book.getfoundry.sh
- Hardhat: https://hardhat.org
- Ethers.js: https://docs.ethers.org

### Support
- GitHub Issues: [Your repo issues link]
- Discord: [Your discord link]

---

**Deployment completed successfully! ✅**

All contracts are verified and ready for integration testing on Sepolia testnet.
