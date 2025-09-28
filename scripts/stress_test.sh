#!/bin/bash

# Payment Gateway Stress Testing Script

# Server URL - modify as needed
BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"

echo "🚀 Starting Payment Gateway Stress Tests"

# Test 1: Health Check Endpoint
echo "🧪 Testing Health Check Endpoint"
hey -z 30s -c 50 $API_URL/health

# Test 2: Get Supported Channels
echo "📋 Testing Get Channels Endpoint"
hey -z 30s -c 50 $API_URL/channels

# Test 3: Payment Requests (Mixed Channels)
echo "💳 Testing Payment Requests"
# Prepare sample payment request data
cat > /tmp/pay_wechat.json <<EOF
{
  "channel": "wechat",
  "out_trade_no": "TEST_WX_$(date +%s%N)",
  "total_amount": 0.01,
  "subject": "Stress Test Payment",
  "scene": "app",
  "notify_url": "https://example.com/notify"
}
EOF

cat > /tmp/pay_alipay.json <<EOF
{
  "channel": "alipay",
  "out_trade_no": "TEST_ALI_$(date +%s%N)",
  "total_amount": 0.01,
  "subject": "Stress Test Payment",
  "scene": "app",
  "notify_url": "https://example.com/notify"
}
EOF

cat > /tmp/pay_unionpay.json <<EOF
{
  "channel": "unionpay",
  "out_trade_no": "TEST_UNION_$(date +%s%N)",
  "total_amount": 0.01,
  "subject": "Stress Test Payment",
  "scene": "app",
  "notify_url": "https://example.com/notify"
}
EOF

# Run concurrent payment requests
hey -z 30s -c 20 -m POST -H "Content-Type: application/json" -d @/tmp/pay_wechat.json $API_URL/pay &
hey -z 30s -c 20 -m POST -H "Content-Type: application/json" -d @/tmp/pay_alipay.json $API_URL/pay &
hey -z 30s -c 10 -m POST -H "Content-Type: application/json" -d @/tmp/pay_unionpay.json $API_URL/pay &

# Wait for all background processes to complete
wait

# Test 4: Query Requests
echo "🔍 Testing Query Requests"
# Prepare sample query request data
cat > /tmp/query.json <<EOF
{
  "channel": "wechat",
  "out_trade_no": "TEST_ORDER_12345"
}
EOF

hey -z 30s -c 50 -m POST -H "Content-Type: application/json" -d @/tmp/query.json $API_URL/query

# Test 5: Mixed Operations
echo "🔄 Testing Mixed Operations"
# Run mixed operations concurrently
hey -z 30s -c 10 -m POST -H "Content-Type: application/json" -d @/tmp/pay_wechat.json $API_URL/pay &
hey -z 30s -c 10 -m POST -H "Content-Type: application/json" -d @/tmp/query.json $API_URL/query &
hey -z 30s -c 5 -H "Content-Type: application/json" $API_URL/channels &
hey -z 30s -c 5 -H "Content-Type: application/json" $API_URL/health &

# Wait for all background processes to complete
wait

# Cleanup
rm -f /tmp/pay_wechat.json /tmp/pay_alipay.json /tmp/pay_unionpay.json /tmp/query.json

echo "✅ Stress Testing Completed"