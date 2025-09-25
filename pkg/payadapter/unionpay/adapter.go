package unionpay

import (
	"context"
	"fmt"
	"time"

	"github.com/ymqzj/payment-gateway/configs"
	"github.com/ymqzj/payment-gateway/internal/payment"
)

// Adapter 银联支付适配器
type Adapter struct {
	client *Client
	config *Config
}

// NewAdapter 创建银联支付适配器
func NewAdapter(cfg *configs.Config) (*Adapter, error) {
	// 创建银联配置
	unionpayConfig := NewConfig(cfg)

	// 使用新的客户端创建函数，自动加载密钥
	client, err := NewClientWithKeys(unionpayConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create unionpay client: %w", err)
	}

	return &Adapter{
		client: client,
		config: unionpayConfig,
	}, nil
}

// Pay 实现支付接口
func (a *Adapter) Pay(ctx context.Context, req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
	// 创建银联订单请求
	unionpayReq := CreateOrderRequest{
		OutTradeNo:  req.OutTradeNo,
		TotalAmount: req.TotalAmount,
		Subject:     req.Subject,
		Body:        req.Body,
		NotifyUrl:   req.NotifyURL,
		ReturnUrl:   req.ReturnURL,
		TraceID:     fmt.Sprintf("trace_%s", req.OutTradeNo),
		Scene:       string(req.Scene),
	}

	// 调用银联创建订单
	resp, err := a.client.CreateOrder(ctx, unionpayReq)
	if err != nil {
		return &payment.UnifiedPayResponse{
			Code:    "1",
			Message: fmt.Sprintf("银联下单失败: %v", err),
		}, nil
	}

	// 根据支付场景返回不同的支付数据
	payData := a.client.buildPayData(resp, req.Scene)

	return &payment.UnifiedPayResponse{
		Code:       "0",
		Message:    "success",
		OrderID:    resp.OrderId,
		OutTradeNo: req.OutTradeNo,
		PayData:    payData,
		Channel:    payment.ChannelUnionPay,
	}, nil
}

// HandleNotify 处理异步通知
func (a *Adapter) HandleNotify(ctx context.Context, data []byte) (*payment.NotifyResult, error) {
	return a.client.HandleNotify(ctx, data)
}

// GetChannel 获取渠道标识
func (a *Adapter) GetChannel() payment.ChannelType {
	return payment.ChannelUnionPay
}

// Refund 退款接口
func (a *Adapter) Refund(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	// 实现银联退款逻辑
	refundTime := time.Now()

	return &payment.RefundResponse{
		Code:         "0",
		Message:      "success",
		RefundID:     fmt.Sprintf("refund_%s", req.OutRefundNo),
		OutRefundNo:  req.OutRefundNo,
		RefundAmount: req.RefundAmount,
		RefundStatus: "SUCCESS",
		RefundTime:   &refundTime,
		Channel:      payment.ChannelUnionPay,
	}, nil
}

// Close 关闭订单接口
func (a *Adapter) Close(ctx context.Context, req *payment.CloseRequest) error {
	// 实现银联关闭订单逻辑
	// This is a placeholder implementation
	return nil
}

// Query 查询订单
func (a *Adapter) Query(ctx context.Context, req *payment.QueryRequest) (*payment.QueryResponse, error) {
	// 实现银联查询订单逻辑
	payTime := time.Now()

	return &payment.QueryResponse{
		Code:        "0",
		Message:     "success",
		OrderID:     req.OrderID,
		OutTradeNo:  req.OutTradeNo,
		TradeStatus: payment.TradeStatusSuccess,
		TotalAmount: 0.01, // placeholder amount
		PayTime:     &payTime,
		Channel:     payment.ChannelUnionPay,
	}, nil
}
