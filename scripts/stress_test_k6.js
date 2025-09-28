import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate } from 'k6/metrics';

// Custom metrics
const paymentRequests = new Counter('payment_requests');
const queryRequests = new Counter('query_requests');
const errorRate = new Rate('errors');

// Options for stress testing
export const options = {
  stages: [
    { duration: '30s', target: 50 },  // Ramp up to 50 users over 30s
    { duration: '1m', target: 50 },   // Stay at 50 users for 1 minute
    { duration: '30s', target: 100 }, // Ramp up to 100 users over 30s
    { duration: '1m', target: 100 },  // Stay at 100 users for 1 minute
    { duration: '30s', target: 200 }, // Ramp up to 200 users over 30s
    { duration: '1m', target: 200 },  // Stay at 200 users for 1 minute
    { duration: '30s', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500'], // 95% of requests should be below 500ms
    'errors': ['rate<0.1'], // Error rate should be less than 10%
  },
};

// Base URL for the API
const BASE_URL = 'http://localhost:8080/api/v1';

// Generate unique order number
function generateOrderNo(prefix) {
  return `${prefix}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

export default function () {
  // Randomly choose an endpoint to test
  const endpoints = ['health', 'channels', 'pay', 'query'];
  const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];

  let res;
  let success;

  switch (endpoint) {
    case 'health':
      res = http.get(`${BASE_URL}/health`);
      success = check(res, {
        'health status is 200': (r) => r.status === 200,
      });
      errorRate.add(!success);
      break;

    case 'channels':
      res = http.get(`${BASE_URL}/channels`);
      success = check(res, {
        'channels status is 200': (r) => r.status === 200,
      });
      errorRate.add(!success);
      break;

    case 'pay':
      const channels = ['wechat', 'alipay', 'unionpay'];
      const channel = channels[Math.floor(Math.random() * channels.length)];
      
      const payPayload = JSON.stringify({
        channel: channel,
        out_trade_no: generateOrderNo(`TEST_${channel.toUpperCase()}`),
        total_amount: 0.01,
        subject: 'Stress Test Payment',
        scene: 'app',
        notify_url: 'https://example.com/notify',
      });

      res = http.post(`${BASE_URL}/pay`, payPayload, {
        headers: { 'Content-Type': 'application/json' },
      });

      success = check(res, {
        'pay status is 200': (r) => r.status === 200,
      });
      paymentRequests.add(1);
      errorRate.add(!success);
      break;

    case 'query':
      const queryPayload = JSON.stringify({
        channel: 'wechat',
        out_trade_no: 'TEST_ORDER_12345',
      });

      res = http.post(`${BASE_URL}/query`, queryPayload, {
        headers: { 'Content-Type': 'application/json' },
      });

      success = check(res, {
        'query status is 200': (r) => r.status === 200,
      });
      queryRequests.add(1);
      errorRate.add(!success);
      break;
  }

  // Add a small delay to simulate real user behavior
  sleep(Math.random() * 2);
}