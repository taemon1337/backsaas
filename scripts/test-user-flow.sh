#!/bin/bash

# BackSaaS User Flow Integration Test
# Tests the complete user journey: register → create tenant → access dashboard

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
BASE_URL="http://localhost:8000"
TEST_EMAIL="test-$(date +%s)@example.com"
TEST_PASSWORD="password123"
TEST_FIRST_NAME="Test"
TEST_LAST_NAME="User"
TEST_COMPANY="Test Company $(date +%s)"
TEST_SLUG="test-company-$(date +%s)"

echo -e "${BLUE}🧪 BackSaaS User Flow Integration Test${NC}"
echo "=================================================="
echo "Testing complete user journey..."
echo ""

# Step 1: Register User
echo -e "${YELLOW}Step 1: Registering new user...${NC}"
echo "Email: $TEST_EMAIL"

REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/platform/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"firstName\": \"$TEST_FIRST_NAME\",
    \"lastName\": \"$TEST_LAST_NAME\",
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\"
  }")

# Check if registration was successful
if echo "$REGISTER_RESPONSE" | jq -e '.token' > /dev/null; then
  TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.token')
  USER_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.user.id')
  echo -e "${GREEN}✅ User registered successfully${NC}"
  echo "User ID: $USER_ID"
  echo "Token: ${TOKEN:0:20}..."
else
  echo -e "${RED}❌ Registration failed${NC}"
  echo "Response: $REGISTER_RESPONSE"
  exit 1
fi

echo ""

# Step 2: Create Tenant
echo -e "${YELLOW}Step 2: Creating tenant...${NC}"
echo "Company: $TEST_COMPANY"
echo "Slug: $TEST_SLUG"

TENANT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/platform/tenants" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"name\": \"$TEST_COMPANY\",
    \"slug\": \"$TEST_SLUG\",
    \"template\": \"crm\",
    \"description\": \"Automated test tenant\"
  }")

# Check if tenant creation was successful
if echo "$TENANT_RESPONSE" | jq -e '.id' > /dev/null; then
  TENANT_ID=$(echo "$TENANT_RESPONSE" | jq -r '.id')
  TENANT_SLUG=$(echo "$TENANT_RESPONSE" | jq -r '.slug')
  echo -e "${GREEN}✅ Tenant created successfully${NC}"
  echo "Tenant ID: $TENANT_ID"
  echo "Tenant Slug: $TENANT_SLUG"
else
  echo -e "${RED}❌ Tenant creation failed${NC}"
  echo "Response: $TENANT_RESPONSE"
  exit 1
fi

echo ""

# Step 3: Test Dashboard Access
echo -e "${YELLOW}Step 3: Testing dashboard access...${NC}"

# Test /ui endpoint
UI_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/ui")
if [ "$UI_STATUS" = "200" ]; then
  echo -e "${GREEN}✅ Dashboard accessible (HTTP $UI_STATUS)${NC}"
else
  echo -e "${RED}❌ Dashboard not accessible (HTTP $UI_STATUS)${NC}"
fi

# Test tenant-specific URL
TENANT_UI_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/ui?tenant=$TENANT_SLUG")
if [ "$TENANT_UI_STATUS" = "200" ]; then
  echo -e "${GREEN}✅ Tenant-specific dashboard accessible (HTTP $TENANT_UI_STATUS)${NC}"
else
  echo -e "${YELLOW}⚠️  Tenant-specific dashboard returned HTTP $TENANT_UI_STATUS${NC}"
fi

echo ""

# Step 4: Test API Endpoints
echo -e "${YELLOW}Step 4: Testing API endpoints...${NC}"

# Test slug availability check
SLUG_CHECK_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/platform/tenants/check-slug?slug=test-available-slug")

if [ "$SLUG_CHECK_STATUS" = "200" ]; then
  echo -e "${GREEN}✅ Slug check API working (HTTP $SLUG_CHECK_STATUS)${NC}"
else
  echo -e "${RED}❌ Slug check API failed (HTTP $SLUG_CHECK_STATUS)${NC}"
fi

# Test user's tenants list
TENANTS_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/platform/users/me/tenants")

if echo "$TENANTS_RESPONSE" | jq -e 'type == "array"' > /dev/null; then
  TENANT_COUNT=$(echo "$TENANTS_RESPONSE" | jq length)
  echo -e "${GREEN}✅ User tenants API working ($TENANT_COUNT tenants)${NC}"
  
  # Verify the created tenant is in the list
  if [ "$TENANT_COUNT" -gt 0 ]; then
    FOUND_TENANT=$(echo "$TENANTS_RESPONSE" | jq -r ".[0].id")
    if [ "$FOUND_TENANT" = "$TENANT_ID" ]; then
      echo -e "${GREEN}✅ Created tenant found in user's tenant list${NC}"
    else
      echo -e "${YELLOW}⚠️  Created tenant not found in list (expected: $TENANT_ID, found: $FOUND_TENANT)${NC}"
    fi
  fi
else
  echo -e "${RED}❌ User tenants API failed: $TENANTS_RESPONSE${NC}"
fi

echo ""

# Summary
echo -e "${BLUE}📊 Test Summary${NC}"
echo "=================================================="
echo -e "User Registration: ${GREEN}✅ PASS${NC}"
echo -e "Tenant Creation: ${GREEN}✅ PASS${NC}"
echo -e "Dashboard Access: $([ "$UI_STATUS" = "200" ] && echo -e "${GREEN}✅ PASS${NC}" || echo -e "${RED}❌ FAIL${NC}")"
echo -e "API Endpoints: $([ "$SLUG_CHECK_STATUS" = "200" ] && echo -e "${GREEN}✅ PASS${NC}" || echo -e "${RED}❌ FAIL${NC}")"

echo ""
echo -e "${BLUE}🔗 Test URLs:${NC}"
echo "Dashboard: $BASE_URL/ui"
echo "Tenant Dashboard: $BASE_URL/ui?tenant=$TENANT_SLUG"
echo "Admin Console: $BASE_URL/admin"

echo ""
echo -e "${BLUE}🔑 Test Credentials:${NC}"
echo "Email: $TEST_EMAIL"
echo "Password: $TEST_PASSWORD"
echo "JWT Token: ${TOKEN:0:30}..."

echo ""
echo -e "${GREEN}🎉 Integration test completed!${NC}"
