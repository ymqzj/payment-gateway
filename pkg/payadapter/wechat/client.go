package wechat

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/ymqzj/payment-gateway/configs"
	"github.com/ymqzj/payment-gateway/internal/payment"
)

type Client struct {
	client *core.Client
	config *Config
}

// NewAdapter 创建微信支付适配器
func NewAdapter(cfg *configs.Config) (*Client, error) {
	config := NewConfig(cfg)

	mchPrivateKey, err := loadPrivateKey(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("load private key failed: %w", err)
	}

	ctx := context.Background()
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(config.MchID, config.SerialNo, mchPrivateKey, config.APIv3Key),
	}

	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("init wechat pay client failed: %w", err)
	}

	return &Client{
		client: client,
		config: config,
	}, nil
}

func loadPrivateKey(file string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("decode private key error")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not rsa private key")
	}
	return rsaKey, nil
}

// Pay 实现支付接口
func (c *Client) Pay(ctx context.Context, req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
	// TODO: Implement actual WeChat Pay logic
	// This is a placeholder implementation
	return &payment.UnifiedPayResponse{
		Code:       "0",
		Message:    "success",
		OrderID:    req.OutTradeNo,
		OutTradeNo: req.OutTradeNo,
		PayData: map[string]string{
			"prepay_id": "wx_test_prepay_id",
		},
		Channel: payment.ChannelWechat,
	}, nil
}

// HandleNotify 处理异步通知
func (c *Client) HandleNotify(ctx context.Context, data []byte) (*payment.NotifyResult, error) {
	// TODO: Implement actual notification handling
	// This is a placeholder implementation
	return &payment.NotifyResult{
		Success:    true,
		OutTradeNo: "",
		Channel:    payment.ChannelWechat,
	}, nil
}

// GetChannel 获取渠道标识
func (c *Client) GetChannel() payment.ChannelType {
	return payment.ChannelWechat
}

// Refund 退款接口
func (c *Client) Refund(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	// TODO: Implement actual WeChat Pay refund logic
	// This is a placeholder implementation
	return &payment.RefundResponse{
		Code:         "0",
		Message:      "success",
		RefundID:     "refund_test_id",
		OutRefundNo:  req.OutRefundNo,
		RefundAmount: req.RefundAmount,
		RefundStatus: "SUCCESS",
		Channel:      payment.ChannelWechat,
	}, nil
}

// Close 关闭订单接口
func (c *Client) Close(ctx context.Context, req *payment.CloseRequest) error {
	// TODO: Implement actual WeChat Pay close order logic
	// This is a placeholder implementation
	return nil
}

// Query 查询订单
func (c *Client) Query(ctx context.Context, req *payment.QueryRequest) (*payment.QueryResponse, error) {
	// TODO: Implement actual WeChat Pay query logic
	// This is a placeholder implementation
	return &payment.QueryResponse{
		Code:        "0",
		Message:     "success",
		OrderID:     req.OrderID,
		OutTradeNo:  req.OutTradeNo,
		TradeStatus: payment.TradeStatusSuccess,
		Channel:     payment.ChannelWechat,
	}, nil
}
