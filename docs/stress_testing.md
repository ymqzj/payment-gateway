# Payment Gateway Stress Testing Guide

## Overview

This document provides instructions on how to perform stress testing on the payment gateway to evaluate its performance and stability under high load conditions.

## Prerequisites

Before running stress tests, ensure you have the following tools installed:

1. [hey](https://github.com/rakyll/hey) - HTTP load generator
2. [k6](https://k6.io/) - Modern load testing tool
3. Running payment gateway server

## Starting the Payment Gateway Server

Before running stress tests, start the payment gateway server:

```bash
# Build the server
go build -o payment-gateway cmd/server/main.go

# Run the server
./payment-gateway
```

Note: The server runs on port 8080 by default. You can change this in the config files.

## Running Stress Tests

### 1. Using hey (simple approach)

Run the provided shell script:

```bash
./scripts/stress_test.sh
```

This script will perform various stress tests including:
- Health check endpoint testing
- Get channels endpoint testing
- Payment requests with mixed channels (WeChat, Alipay, UnionPay)
- Query requests
- Mixed operations

### 2. Using k6 (advanced approach)

Run the k6 test script:

```bash
k6 run scripts/stress_test_k6.js
```

This script provides more detailed metrics and follows a progressive load pattern:
- Starts with 50 virtual users
- Gradually increases to 200 virtual users
- Maintains load at different levels
- Includes custom metrics and thresholds

## Test Scenarios

The stress tests cover the following scenarios:

1. **Health Check**: Tests the `/api/v1/health` endpoint
2. **Supported Channels**: Tests the `/api/v1/channels` endpoint
3. **Payment Processing**: Tests the `/api/v1/pay` endpoint with different payment channels
4. **Order Query**: Tests the `/api/v1/query` endpoint
5. **Mixed Operations**: Concurrently tests multiple endpoints

## Monitoring During Tests

While running stress tests, monitor:

1. **Server Resource Usage**:
   ```bash
   # Monitor CPU and memory usage
   top -p $(pgrep payment-gateway)
   ```

2. **Response Times**: Check for increased latency under load
3. **Error Rates**: Monitor for HTTP errors or application errors
4. **Throughput**: Track requests per second

## Interpreting Results

Key metrics to analyze:

1. **Requests Per Second (RPS)**: How many requests the system can handle
2. **Response Time**: Average and 95th percentile response times
3. **Error Rate**: Percentage of failed requests
4. **Resource Utilization**: CPU and memory usage of the server

## Recommendations

1. **Start Small**: Begin with low concurrency and gradually increase load
2. **Monitor Resources**: Keep an eye on CPU, memory, and network usage
3. **Test All Endpoints**: Ensure all API endpoints are tested, not just the most common ones
4. **Simulate Real Usage**: Use realistic request patterns and data
5. **Test Failure Scenarios**: Include tests with invalid data and error conditions