package unionpay

import (
	"context"
	"fmt"

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
	// TODO: 实现银联异步通知处理
	// 这里需要解析银联的通知数据，验证签名，并返回处理结果

	// 暂时返回空结果，后续需要实现具体的通知处理逻辑
	return &payment.NotifyResult{
		Success:    false,
		OutTradeNo: "",
		Channel:    payment.ChannelUnionPay,
	}, fmt.Errorf("notify handling not implemented yet")
}

// GetChannel 获取渠道标识
func (a *Adapter) GetChannel() payment.ChannelType {
	return payment.ChannelUnionPay
}

// Refund 退款接口
func (a *Adapter) Refund(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	// TODO: Implement actual UnionPay refund logic
	// This is a placeholder implementation
	return &payment.RefundResponse{
		Code:         "0",
		Message:      "success",
		RefundID:     "refund_test_id",
		OutRefundNo:  req.OutRefundNo,
		RefundAmount: req.RefundAmount,
		RefundStatus: "SUCCESS",
		Channel:      payment.ChannelUnionPay,
	}, nil
}

// Close 关闭订单接口
func (a *Adapter) Close(ctx context.Context, req *payment.CloseRequest) error {
	// TODO: Implement actual UnionPay close order logic
	// This is a placeholder implementation
	return nil
}

// Query 查询订单
func (a *Adapter) Query(ctx context.Context, req *payment.QueryRequest) (*payment.QueryResponse, error) {
	// TODO: Implement actual UnionPay query logic
	// This is a placeholder implementation
	return &payment.QueryResponse{
		Code:        "0",
		Message:     "success",
		OrderID:     req.OrderID,
		OutTradeNo:  req.OutTradeNo,
		TradeStatus: payment.TradeStatusSuccess,
		Channel:     payment.ChannelUnionPay,
	}, nil
}
