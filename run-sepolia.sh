#!/bin/bash

# SAGE Payment Agent - Sepolia Testnet Runner
# This script helps you run the payment agent with Sepolia testnet configuration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║         SAGE Payment Agent - Sepolia Testnet              ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠️  .env file not found${NC}"
    echo -e "${GREEN}Creating .env from .env.sepolia template...${NC}"

    if [ -f .env.sepolia ]; then
        cp .env.sepolia .env
        echo -e "${GREEN}✅ .env file created${NC}"
        echo ""
        echo -e "${YELLOW}⚠️  IMPORTANT: Please edit .env and update:${NC}"
        echo "   1. BLOCKCHAIN_RPC_URL with your Infura/Alchemy API key"
        echo "   2. WALLET_ADDRESS with your agent's wallet address"
        echo ""
        echo -e "${BLUE}Get Sepolia test ETH from:${NC}"
        echo "   - https://www.alchemy.com/faucets/ethereum-sepolia"
        echo "   - https://sepoliafaucet.com"
        echo ""
        read -p "Press Enter after updating .env to continue..."
    else
        echo -e "${RED}❌ .env.sepolia template not found${NC}"
        exit 1
    fi
fi

# Load environment variables
echo -e "${GREEN}Loading configuration from .env...${NC}"
export $(cat .env | grep -v '^#' | xargs)

# Validate required environment variables
REQUIRED_VARS=(
    "BLOCKCHAIN_RPC_URL"
    "IDENTITY_REGISTRY_ADDRESS"
    "VALIDATION_REGISTRY_ADDRESS"
    "REPUTATION_REGISTRY_ADDRESS"
)

MISSING_VARS=()
for var in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!var}" ]; then
        MISSING_VARS+=("$var")
    fi
done

if [ ${#MISSING_VARS[@]} -gt 0 ]; then
    echo -e "${RED}❌ Missing required environment variables:${NC}"
    for var in "${MISSING_VARS[@]}"; do
        echo "   - $var"
    done
    exit 1
fi

# Check if RPC URL is still using placeholder
if [[ "$BLOCKCHAIN_RPC_URL" == *"YOUR_INFURA_PROJECT_ID"* ]] || [[ "$BLOCKCHAIN_RPC_URL" == *"YOUR_ALCHEMY_API_KEY"* ]]; then
    echo -e "${RED}❌ Please update BLOCKCHAIN_RPC_URL in .env with your actual API key${NC}"
    exit 1
fi

# Display configuration
echo ""
echo -e "${BLUE}Configuration:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Network:           Sepolia Testnet (Chain ID: 11155111)"
echo "SAGE Protocol:     ${SAGE_ENABLED:-true}"
echo "Port:              ${AGENT_PORT:-8091}"
echo "Log Level:         ${LOG_LEVEL:-info}"
echo ""
echo "Contract Addresses:"
echo "  Identity Registry:   $IDENTITY_REGISTRY_ADDRESS"
echo "  Validation Registry: $VALIDATION_REGISTRY_ADDRESS"
echo "  Reputation Registry: $REPUTATION_REGISTRY_ADDRESS"
echo ""
echo "Wallet:            ${WALLET_ADDRESS}"
echo "Simulation Mode:   ${TX_SIMULATION_MODE:-true}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if binary exists
if [ ! -f payment-agent ]; then
    echo -e "${YELLOW}⚠️  payment-agent binary not found${NC}"
    echo -e "${GREEN}Building payment-agent...${NC}"
    go build -o payment-agent
    echo -e "${GREEN}✅ Build complete${NC}"
    echo ""
fi

# Test RPC connection
echo -e "${BLUE}Testing RPC connection...${NC}"
BLOCK_NUMBER=$(curl -s -X POST $BLOCKCHAIN_RPC_URL \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | \
    grep -o '"result":"[^"]*"' | cut -d'"' -f4)

if [ -z "$BLOCK_NUMBER" ]; then
    echo -e "${RED}❌ Failed to connect to RPC endpoint${NC}"
    echo "   Please check your BLOCKCHAIN_RPC_URL in .env"
    exit 1
fi

echo -e "${GREEN}✅ Connected to Sepolia (Block: $((16#${BLOCK_NUMBER:2})))${NC}"
echo ""

# Verify contracts are deployed
echo -e "${BLUE}Verifying contract deployments...${NC}"

verify_contract() {
    local address=$1
    local name=$2

    CODE=$(curl -s -X POST $BLOCKCHAIN_RPC_URL \
        -H "Content-Type: application/json" \
        -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getCode\",\"params\":[\"$address\",\"latest\"],\"id\":1}" | \
        grep -o '"result":"[^"]*"' | cut -d'"' -f4)

    if [ "$CODE" == "0x" ] || [ -z "$CODE" ]; then
        echo -e "${RED}❌ $name not found at $address${NC}"
        return 1
    else
        echo -e "${GREEN}✅ $name verified at $address${NC}"
        return 0
    fi
}

VERIFICATION_FAILED=0
verify_contract "$IDENTITY_REGISTRY_ADDRESS" "Identity Registry" || VERIFICATION_FAILED=1
verify_contract "$VALIDATION_REGISTRY_ADDRESS" "Validation Registry" || VERIFICATION_FAILED=1
verify_contract "$REPUTATION_REGISTRY_ADDRESS" "Reputation Registry" || VERIFICATION_FAILED=1

if [ $VERIFICATION_FAILED -eq 1 ]; then
    echo ""
    echo -e "${RED}❌ Contract verification failed${NC}"
    echo "   Please check contract addresses in .env"
    exit 1
fi

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ All checks passed! Starting payment agent...${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Start the agent
./payment-agent
