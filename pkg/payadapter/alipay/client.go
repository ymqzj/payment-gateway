package alipay

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/smartwalle/alipay/v3"
	"github.com/ymqzj/payment-gateway/configs"
	"github.com/ymqzj/payment-gateway/internal/payment"
)

type Client struct {
	client *alipay.Client
	config *Config
}

// NewAdapter 创建支付宝适配器
func NewAdapter(cfg *configs.Config) (*Client, error) {
	config := NewConfig(cfg)

	client, err := alipay.New(config.AppID, config.PrivateKey, !config.IsSandbox)
	if err != nil {
		return nil, fmt.Errorf("failed to create alipay client: %w", err)
	}

	// Load Alipay public key for verification
	if err := client.LoadAliPayPublicKey(config.AlipayPublicKey); err != nil {
		return nil, fmt.Errorf("failed to load alipay public key: %w", err)
	}

	// Note: In the newer version of the SDK, we don't set global return/notify URLs
	// These are set per request instead

	return &Client{
		client: client,
		config: config,
	}, nil
}

// Pay 实现支付接口
func (c *Client) Pay(ctx context.Context, req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
	// 创建支付宝支付请求
	var payURL string
	var err error

	switch req.Scene {
	case "app":
		// App支付
		var p = alipay.TradeAppPay{}
		p.NotifyURL = c.config.NotifyURL
		p.Subject = req.Subject
		p.OutTradeNo = req.OutTradeNo
		p.TotalAmount = fmt.Sprintf("%.2f", req.TotalAmount)
		p.ProductCode = "QUICK_MSECURITY_PAY" // App支付固定值
		payURL, err = c.client.TradeAppPay(p)
		if err != nil {
			return nil, err
		}
	case "h5":
		// 手机网站支付
		var p = alipay.TradeWapPay{}
		p.NotifyURL = c.config.NotifyURL
		p.ReturnURL = req.ReturnURL
		p.Subject = req.Subject
		p.OutTradeNo = req.OutTradeNo
		p.TotalAmount = fmt.Sprintf("%.2f", req.TotalAmount)
		p.ProductCode = "QUICK_WAP_PAY" // 手机网站支付固定值
		result, err := c.client.TradeWapPay(p)
		if err != nil {
			return nil, err
		}
		payURL = result.String()
	case "pc":
		// 电脑网站支付
		var p = alipay.TradePagePay{}
		p.NotifyURL = c.config.NotifyURL
		p.ReturnURL = req.ReturnURL
		p.Subject = req.Subject
		p.OutTradeNo = req.OutTradeNo
		p.TotalAmount = fmt.Sprintf("%.2f", req.TotalAmount)
		p.ProductCode = "FAST_INSTANT_TRADE_PAY" // 电脑网站支付固定值
		result, err := c.client.TradePagePay(p)
		if err != nil {
			return nil, err
		}
		payURL = result.String()
	default:
		// 默认使用App支付
		var p = alipay.TradeAppPay{}
		p.NotifyURL = c.config.NotifyURL
		p.Subject = req.Subject
		p.OutTradeNo = req.OutTradeNo
		p.TotalAmount = fmt.Sprintf("%.2f", req.TotalAmount)
		p.ProductCode = "QUICK_MSECURITY_PAY" // App支付固定值
		payURL, err = c.client.TradeAppPay(p)
		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, fmt.Errorf("alipay pay failed: %w", err)
	}

	return &payment.UnifiedPayResponse{
		Code:       "0",
		Message:    "success",
		OrderID:    req.OutTradeNo,
		OutTradeNo: req.OutTradeNo,
		PayData: map[string]string{
			"pay_url": payURL,
		},
		Channel: payment.ChannelAlipay,
	}, nil
}

// HandleNotify 处理异步通知
func (c *Client) HandleNotify(ctx context.Context, data []byte) (*payment.NotifyResult, error) {
	// Create HTTP request from the actual notification data
	req, err := http.NewRequest("POST", "", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 解析并验签通知
	noti, err := c.client.GetTradeNotification(req)
	if err != nil {
		return nil, fmt.Errorf("handle notify failed: %w", err)
	}

	// Convert string amount to float64
	var totalAmount float64
	fmt.Sscanf(noti.TotalAmount, "%f", &totalAmount)

	// Parse pay time
	var payTime *time.Time
	if noti.GmtPayment != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", noti.GmtPayment); err == nil {
			payTime = &t
		}
	}

	result := &payment.NotifyResult{
		Success:     noti.TradeStatus == alipay.TradeStatusSuccess,
		OutTradeNo:  noti.OutTradeNo,
		TotalAmount: totalAmount,
		TradeStatus: string(noti.TradeStatus),
		Channel:     payment.ChannelAlipay,
		OrderID:     noti.TradeNo,
		PayTime:     payTime,
	}

	return result, nil
}

// GetChannel 获取渠道标识
func (c *Client) GetChannel() payment.ChannelType {
	return payment.ChannelAlipay
}

// Refund 退款接口
func (c *Client) Refund(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	var p = alipay.TradeRefund{}
	p.OutTradeNo = req.OutTradeNo
	p.RefundAmount = strconv.FormatFloat(req.RefundAmount, 'f', 2, 64)
	p.OutRequestNo = req.OutRefundNo // 幂等键
	p.RefundReason = req.RefundReason

	resp, err := c.client.TradeRefund(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("alipay refund failed: %w", err)
	}

	if resp.Code != "10000" {
		return &payment.RefundResponse{
			Code:    "1",
			Message: fmt.Sprintf("支付宝退款失败: %s - %s", resp.Code, resp.Msg),
		}, nil
	}

	refundAmount, _ := strconv.ParseFloat(resp.RefundFee, 64)

	// Using a simple time for the refund time
	now := time.Now()

	return &payment.RefundResponse{
		Code:         "0",
		Message:      "success",
		RefundID:     resp.TradeNo,
		OutRefundNo:  req.OutRefundNo,
		RefundAmount: refundAmount,
		RefundStatus: "SUCCESS",
		RefundTime:   &now,
		Channel:      payment.ChannelAlipay,
	}, nil
}

// Close 关闭订单接口
func (c *Client) Close(ctx context.Context, req *payment.CloseRequest) error {
	var p = alipay.TradeClose{}
	p.OutTradeNo = req.OutTradeNo

	_, err := c.client.TradeClose(ctx, p)
	if err != nil {
		return fmt.Errorf("alipay close order failed: %w", err)
	}

	return nil
}

// Query 查询订单
func (c *Client) Query(ctx context.Context, req *payment.QueryRequest) (*payment.QueryResponse, error) {
	var p = alipay.TradeQuery{}
	p.OutTradeNo = req.OutTradeNo

	resp, err := c.client.TradeQuery(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("alipay query order failed: %w", err)
	}

	if resp.Code != "10000" {
		return nil, fmt.Errorf("alipay query order failed: %s - %s", resp.Code, resp.Msg)
	}

	status := payment.TradeStatusNotPay
	switch resp.TradeStatus {
	case alipay.TradeStatusSuccess:
		status = payment.TradeStatusSuccess
	case "TRADE_CLOSED":
		status = payment.TradeStatusClosed
	case "WAIT_BUYER_PAY":
		status = payment.TradeStatusNotPay
	}

	var payTime *time.Time
	if resp.SendPayDate != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", resp.SendPayDate); err == nil {
			payTime = &t
		}
	}

	totalAmount, _ := strconv.ParseFloat(resp.TotalAmount, 64)

	return &payment.QueryResponse{
		Code:        "0",
		Message:     "success",
		OrderID:     resp.TradeNo,
		OutTradeNo:  req.OutTradeNo,
		TradeStatus: status,
		TotalAmount: totalAmount,
		PayTime:     payTime,
		Channel:     payment.ChannelAlipay,
	}, nil
}
