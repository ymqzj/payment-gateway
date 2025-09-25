package alipay

import (
	"context"
	"fmt"
	"net/http"

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
	// Create a fake HTTP request to satisfy the SDK API
	// In a real implementation, the data should be parsed properly
	req, err := http.NewRequest("POST", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Body = http.NoBody

	// 解析并验签通知
	noti, err := c.client.GetTradeNotification(req)
	if err != nil {
		return nil, fmt.Errorf("handle notify failed: %w", err)
	}

	// Convert string amount to float64
	var totalAmount float64
	fmt.Sscanf(noti.TotalAmount, "%f", &totalAmount)

	result := &payment.NotifyResult{
		Success:     noti.TradeStatus == alipay.TradeStatusSuccess,
		OutTradeNo:  noti.OutTradeNo,
		TotalAmount: totalAmount,
		TradeStatus: string(noti.TradeStatus),
		Channel:     payment.ChannelAlipay,
	}

	return result, nil
}

// GetChannel 获取渠道标识
func (c *Client) GetChannel() payment.ChannelType {
	return payment.ChannelAlipay
}

// Refund 退款接口
func (c *Client) Refund(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	// TODO: Implement actual Alipay refund logic
	// This is a placeholder implementation
	return &payment.RefundResponse{
		Code:         "0",
		Message:      "success",
		RefundID:     "refund_test_id",
		OutRefundNo:  req.OutRefundNo,
		RefundAmount: req.RefundAmount,
		RefundStatus: "SUCCESS",
		Channel:      payment.ChannelAlipay,
	}, nil
}

// Close 关闭订单接口
func (c *Client) Close(ctx context.Context, req *payment.CloseRequest) error {
	// TODO: Implement actual Alipay close order logic
	// This is a placeholder implementation
	return nil
}

// Query 查询订单
func (c *Client) Query(ctx context.Context, req *payment.QueryRequest) (*payment.QueryResponse, error) {
	// TODO: Implement actual Alipay query logic
	// This is a placeholder implementation
	return &payment.QueryResponse{
		Code:        "0",
		Message:     "success",
		OrderID:     req.OrderID,
		OutTradeNo:  req.OutTradeNo,
		TradeStatus: payment.TradeStatusSuccess,
		Channel:     payment.ChannelAlipay,
	}, nil
}