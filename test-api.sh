#!/bin/bash

# X-UI API Testing Script with Authentication Bypass
# Requires XUI_DEV_MODE=true environment variable

BASE_URL="http://localhost:54321"
BYPASS_HEADER="X-Dev-Bypass: dev-test-bypass-2024"

echo "=== X-UI API Test Script ==="
echo "Base URL: $BASE_URL"
echo ""

# Test 1: Get all outbounds
echo "1. Testing GET /xui/outbound/list"
curl -s -X POST "$BASE_URL/xui/outbound/list" \
  -H "$BYPASS_HEADER" \
  -H "Content-Type: application/json" | jq '.' || echo "Failed"
echo ""

# Test 2: Get all subscriptions
echo "2. Testing GET /xui/subscription/list"
curl -s -X POST "$BASE_URL/xui/subscription/list" \
  -H "$BYPASS_HEADER" \
  -H "Content-Type: application/json" | jq '.' || echo "Failed"
echo ""

# Test 3: Get all settings
echo "3. Testing GET /xui/setting/all"
curl -s -X POST "$BASE_URL/xui/setting/all" \
  -H "$BYPASS_HEADER" \
  -H "Content-Type: application/json" | jq '.' || echo "Failed"
echo ""

# Test 4: Add a test outbound
echo "4. Testing POST /xui/outbound/add (creating test outbound)"
curl -s -X POST "$BASE_URL/xui/outbound/add" \
  -H "$BYPASS_HEADER" \
  -H "Content-Type: application/json" \
  -d '{
    "protocol": "vmess",
    "address": "test.example.com",
    "port": 443,
    "settings": "{\"vnext\":[{\"address\":\"test.example.com\",\"port\":443,\"users\":[{\"id\":\"test-uuid\",\"alterId\":0,\"security\":\"auto\"}]}]}",
    "streamSettings": "{\"network\":\"tcp\",\"security\":\"none\"}",
    "tag": "test-outbound-001",
    "remark": "Test Outbound",
    "groupName": "test-group"
  }' | jq '.' || echo "Failed"
echo ""

# Test 5: Test the HTML page loads
echo "5. Testing HTML page access /xui/outbounds"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/xui/outbounds" \
  -H "$BYPASS_HEADER")
echo "HTTP Status: $HTTP_CODE"
if [ "$HTTP_CODE" = "200" ]; then
  echo "✓ Page loads successfully"
else
  echo "✗ Page failed to load"
fi
echo ""

echo "=== Test Complete ==="
