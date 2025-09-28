package stress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ymqzj/payment-gateway/internal/payment"
)

// StressTestConfig holds configuration for stress testing
type StressTestConfig struct {
	BaseURL         string
	ConcurrentUsers int
	Duration        time.Duration
}

// TestResult holds the results of a stress test
type TestResult struct {
	TotalRequests   int64
	SuccessRequests int64
	ErrorRequests   int64
	StartTime       time.Time
	EndTime         time.Time
}

// PaymentRequest represents a payment request
type PaymentRequest struct {
	Channel     string  `json:"channel"`
	OutTradeNo  string  `json:"out_trade_no"`
	TotalAmount float64 `json:"total_amount"`
	Subject     string  `json:"subject"`
	Scene       string  `json:"scene"`
	NotifyURL   string  `json:"notify_url"`
}

func main() {
	config := StressTestConfig{
		BaseURL:         "http://localhost:8080/api/v1",
		ConcurrentUsers: 100,
		Duration:        5 * time.Minute,
	}

	fmt.Printf("🚀 Starting stress test with %d concurrent users for %v\n", config.ConcurrentUsers, config.Duration)

	var wg sync.WaitGroup
	resultChan := make(chan TestResult, config.ConcurrentUsers)

	// Start concurrent users
	startTime := time.Now()
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			result := runUserTest(config, userID)
			resultChan <- result
		}(i)
	}

	// Wait for all users to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var totalResult TestResult
	totalResult.StartTime = startTime

	for result := range resultChan {
		totalResult.TotalRequests += result.TotalRequests
		totalResult.SuccessRequests += result.SuccessRequests
		totalResult.ErrorRequests += result.ErrorRequests
	}

	totalResult.EndTime = time.Now()
	duration := totalResult.EndTime.Sub(totalResult.StartTime)

	// Print results
	fmt.Println("\n📊 Stress Test Results:")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Total Requests: %d\n", totalResult.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", totalResult.SuccessRequests)
	fmt.Printf("Error Requests: %d\n", totalResult.ErrorRequests)
	fmt.Printf("Success Rate: %.2f%%\n", float64(totalResult.SuccessRequests)/float64(totalResult.TotalRequests)*100)
	fmt.Printf("Requests Per Second: %.2f\n", float64(totalResult.TotalRequests)/duration.Seconds())
}

func runUserTest(config StressTestConfig, userID int) TestResult {
	var result TestResult
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	endTime := time.Now().Add(config.Duration)
	channels := []string{string(payment.ChannelWechat), string(payment.ChannelAlipay), string(payment.ChannelUnionPay)}

	for time.Now().Before(endTime) {
		// Randomly choose a channel
		channel := channels[userID%len(channels)]

		// Create payment request
		paymentReq := PaymentRequest{
			Channel:     channel,
			OutTradeNo:  fmt.Sprintf("STRESS_%s_%d_%d", channel, userID, time.Now().UnixNano()),
			TotalAmount: 0.01,
			Subject:     "Stress Test Payment",
			Scene:       "app",
			NotifyURL:   "https://example.com/notify",
		}

		// Send payment request
		success := sendPaymentRequest(client, config.BaseURL, paymentReq)
		result.TotalRequests++
		if success {
			result.SuccessRequests++
		} else {
			result.ErrorRequests++
		}

		// Small delay to prevent overwhelming the server
		time.Sleep(100 * time.Millisecond)
	}

	return result
}

func sendPaymentRequest(client *http.Client, baseURL string, req PaymentRequest) bool {
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return false
	}

	httpReq, err := http.NewRequest("POST", baseURL+"/pay", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return false
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
