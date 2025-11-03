#!/bin/bash

# SAGE Payment Agent - Sepolia Integration Tests
# Tests the payment agent against deployed Sepolia contracts

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
AGENT_URL=${AGENT_URL:-http://localhost:8091}
TIMEOUT=5

echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║       SAGE Payment Agent - Sepolia Integration Tests      ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Check if agent is running
echo -e "${CYAN}Checking if payment agent is running...${NC}"
if ! curl -s -f -m $TIMEOUT "$AGENT_URL/health" > /dev/null 2>&1; then
    echo -e "${RED}❌ Payment agent is not running at $AGENT_URL${NC}"
    echo ""
    echo "Please start the agent first:"
    echo "  ./run-sepolia.sh"
    exit 1
fi
echo -e "${GREEN}✅ Payment agent is running${NC}"
echo ""

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to run a test
run_test() {
    local test_name=$1
    local test_command=$2
    local expected_pattern=$3

    TESTS_RUN=$((TESTS_RUN + 1))
    echo -e "${CYAN}Test $TESTS_RUN: $test_name${NC}"

    local result
    result=$(eval "$test_command" 2>&1)
    local exit_code=$?

    if [ $exit_code -eq 0 ] && echo "$result" | grep -q "$expected_pattern"; then
        echo -e "${GREEN}✅ PASSED${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}❌ FAILED${NC}"
        echo -e "${YELLOW}Output:${NC}"
        echo "$result"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Test 1: Health Check
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
run_test "Health Check" \
    "curl -s -f -m $TIMEOUT $AGENT_URL/health" \
    '"status".*"healthy"'
echo ""

# Test 2: Status Endpoint
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
run_test "Status Endpoint" \
    "curl -s -f -m $TIMEOUT $AGENT_URL/status" \
    '"agent".*"sage-payment-agent"'
echo ""

# Test 3: Root Endpoint
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
run_test "Root Endpoint" \
    "curl -s -f -m $TIMEOUT $AGENT_URL/" \
    "SAGE Payment Agent"
echo ""

# Test 4: Payment Request (SAGE disabled - should succeed)
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}Test $((TESTS_RUN + 1)): Payment Request (No Signature)${NC}"
TESTS_RUN=$((TESTS_RUN + 1))

PAYMENT_RESPONSE=$(curl -s -X POST "$AGENT_URL/payment" \
    -H "Content-Type: application/json" \
    -d '{
        "amount": 100.0,
        "currency": "USDC",
        "product": "Test Product",
        "recipient": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
    }' 2>&1)

if echo "$PAYMENT_RESPONSE" | grep -q '"success".*true\|"transaction_id"'; then
    echo -e "${GREEN}✅ PASSED - Payment processed${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo -e "${YELLOW}Transaction ID:${NC} $(echo $PAYMENT_RESPONSE | grep -o '"transaction_id":"[^"]*"' | cut -d'"' -f4)"
elif echo "$PAYMENT_RESPONSE" | grep -q 'signature.*required\|verification.*failed'; then
    echo -e "${YELLOW}⚠️  EXPECTED BEHAVIOR - SAGE verification is enabled${NC}"
    echo "    Payment rejected due to missing signature (this is correct when SAGE_ENABLED=true)"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}❌ FAILED${NC}"
    echo -e "${YELLOW}Response:${NC}"
    echo "$PAYMENT_RESPONSE"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 5: Invalid Payment Request
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}Test $((TESTS_RUN + 1)): Invalid Payment Request${NC}"
TESTS_RUN=$((TESTS_RUN + 1))

INVALID_RESPONSE=$(curl -s -X POST "$AGENT_URL/payment" \
    -H "Content-Type: application/json" \
    -d '{
        "amount": -100.0,
        "currency": "USDC"
    }' 2>&1)

if echo "$INVALID_RESPONSE" | grep -q '"success".*false\|"error"\|400\|invalid'; then
    echo -e "${GREEN}✅ PASSED - Invalid request rejected${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}❌ FAILED - Invalid request should be rejected${NC}"
    echo -e "${YELLOW}Response:${NC}"
    echo "$INVALID_RESPONSE"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
echo ""

# Test 6: Content-Type Validation
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
run_test "Content-Type Validation" \
    "curl -s -X POST $AGENT_URL/payment -H 'Content-Type: text/plain' -d 'invalid' -w '%{http_code}'" \
    "40[0-9]"
echo ""

# Test 7: 404 Not Found
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
run_test "404 Not Found" \
    "curl -s -o /dev/null -w '%{http_code}' $AGENT_URL/nonexistent" \
    "404"
echo ""

# Contract Verification Tests (if RPC is configured)
if [ -f .env ]; then
    source .env

    if [ ! -z "$BLOCKCHAIN_RPC_URL" ] && [[ "$BLOCKCHAIN_RPC_URL" != *"YOUR_"* ]]; then
        echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${CYAN}Blockchain Contract Verification${NC}"
        echo ""

        # Test 8: Identity Registry
        echo -e "${CYAN}Test $((TESTS_RUN + 1)): Identity Registry Contract${NC}"
        TESTS_RUN=$((TESTS_RUN + 1))

        if [ ! -z "$IDENTITY_REGISTRY_ADDRESS" ]; then
            CODE=$(curl -s -X POST $BLOCKCHAIN_RPC_URL \
                -H "Content-Type: application/json" \
                -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getCode\",\"params\":[\"$IDENTITY_REGISTRY_ADDRESS\",\"latest\"],\"id\":1}" | \
                grep -o '"result":"[^"]*"' | cut -d'"' -f4)

            if [ "$CODE" != "0x" ] && [ ! -z "$CODE" ]; then
                echo -e "${GREEN}✅ PASSED - Contract deployed at $IDENTITY_REGISTRY_ADDRESS${NC}"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                echo -e "${RED}❌ FAILED - Contract not found${NC}"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            fi
        else
            echo -e "${YELLOW}⚠️  SKIPPED - IDENTITY_REGISTRY_ADDRESS not set${NC}"
        fi
        echo ""

        # Test 9: Validation Registry
        echo -e "${CYAN}Test $((TESTS_RUN + 1)): Validation Registry Contract${NC}"
        TESTS_RUN=$((TESTS_RUN + 1))

        if [ ! -z "$VALIDATION_REGISTRY_ADDRESS" ]; then
            CODE=$(curl -s -X POST $BLOCKCHAIN_RPC_URL \
                -H "Content-Type: application/json" \
                -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getCode\",\"params\":[\"$VALIDATION_REGISTRY_ADDRESS\",\"latest\"],\"id\":1}" | \
                grep -o '"result":"[^"]*"' | cut -d'"' -f4)

            if [ "$CODE" != "0x" ] && [ ! -z "$CODE" ]; then
                echo -e "${GREEN}✅ PASSED - Contract deployed at $VALIDATION_REGISTRY_ADDRESS${NC}"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                echo -e "${RED}❌ FAILED - Contract not found${NC}"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            fi
        else
            echo -e "${YELLOW}⚠️  SKIPPED - VALIDATION_REGISTRY_ADDRESS not set${NC}"
        fi
        echo ""

        # Test 10: Reputation Registry
        echo -e "${CYAN}Test $((TESTS_RUN + 1)): Reputation Registry Contract${NC}"
        TESTS_RUN=$((TESTS_RUN + 1))

        if [ ! -z "$REPUTATION_REGISTRY_ADDRESS" ]; then
            CODE=$(curl -s -X POST $BLOCKCHAIN_RPC_URL \
                -H "Content-Type: application/json" \
                -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getCode\",\"params\":[\"$REPUTATION_REGISTRY_ADDRESS\",\"latest\"],\"id\":1}" | \
                grep -o '"result":"[^"]*"' | cut -d'"' -f4)

            if [ "$CODE" != "0x" ] && [ ! -z "$CODE" ]; then
                echo -e "${GREEN}✅ PASSED - Contract deployed at $REPUTATION_REGISTRY_ADDRESS${NC}"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                echo -e "${RED}❌ FAILED - Contract not found${NC}"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            fi
        else
            echo -e "${YELLOW}⚠️  SKIPPED - REPUTATION_REGISTRY_ADDRESS not set${NC}"
        fi
        echo ""
    fi
fi

# Summary
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "Tests Run:    $TESTS_RUN"
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
else
    echo "Tests Failed: $TESTS_FAILED"
fi
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}✅ All tests passed!${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    exit 0
else
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${RED}❌ Some tests failed${NC}"
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    exit 1
fi
