#!/bin/bash

# X-UI Frontend Testing Utility
# Tests API endpoints and validates frontend components

BASE_URL="http://localhost:54321"
BYPASS_HEADER="X-Dev-Bypass: dev-test-bypass-2024"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=== X-UI Frontend Testing Utility ==="
echo ""

# Test API endpoint and return success/failure
test_api() {
    local name="$1"
    local method="$2"
    local endpoint="$3"
    local data="$4"
    
    echo -n "Testing $name... "
    
    if [ -z "$data" ]; then
        response=$(curl -s -X "$method" "$BASE_URL$endpoint" \
            -H "$BYPASS_HEADER" \
            -H "Content-Type: application/json")
    else
        response=$(curl -s -X "$method" "$BASE_URL$endpoint" \
            -H "$BYPASS_HEADER" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    success=$(echo "$response" | jq -r '.success // empty' 2>/dev/null)
    
    if [ "$success" = "true" ]; then
        echo -e "${GREEN}✓${NC}"
        return 0
    else
        echo -e "${RED}✗${NC}"
        echo "  Response: $response"
        return 1
    fi
}

# Test HTML component presence
test_component() {
    local name="$1"
    local pattern="$2"
    local page="${3:-/xui/outbounds}"
    
    echo -n "Checking $name... "
    
    html=$(curl -s "$BASE_URL$page" -H "$BYPASS_HEADER")
    
    if echo "$html" | grep -q "$pattern"; then
        echo -e "${GREEN}✓${NC}"
        return 0
    else
        echo -e "${RED}✗${NC} (not found: $pattern)"
        return 1
    fi
}

echo "=== API Endpoint Tests ==="
test_api "Get Outbounds" "POST" "/xui/outbound/list"
test_api "Get Subscriptions" "POST" "/xui/subscription/list"
test_api "Get Settings" "POST" "/xui/setting/all"
echo ""

echo "=== Frontend Component Tests ==="
test_component "Auto-complete for groups" "a-auto-complete"
test_component "Subscription button" "openSubscriptionManager"
test_component "Group manager" "openGroupManager"
test_component "Add outbound button" "openAddOutbound"
test_component "Subscription modal" "subscriptionModal.visible"
test_component "Group edit modal" "groupEditModal.visible"
echo ""

echo "=== JavaScript Method Tests ==="
test_component "filterGroupOption method" "filterGroupOption"
test_component "getSubscriptions method" "getSubscriptions"
test_component "saveSubscription method" "saveSubscription"
test_component "refreshSubscription method" "refreshSubscription"
test_component "addNewGroup method" "addNewGroup"
test_component "saveGroup method" "saveGroup"
echo ""

echo "=== Computed Properties Tests ==="
test_component "uniqueGroups computed" "uniqueGroups()"
test_component "testedCount computed" "testedCount()"
test_component "filteredOutbounds computed" "filteredOutbounds()"
echo ""

echo "=== Translation Keys Tests ==="
test_component "testedOutbounds translation" "pages.outbounds.testedOutbounds"
test_component "autoRouteMessage translation" "pages.outbounds.autoRouteMessage"
echo ""

echo "=== Test Complete ==="
