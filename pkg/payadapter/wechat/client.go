package wechat

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/app"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/h5"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/ymqzj/payment-gateway/configs"
	"github.com/ymqzj/payment-gateway/internal/payment"
)

const (
	timeFormat = "2006-01-02T15:04:05-07:00"
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
	switch req.Scene {
	case payment.SceneApp:
		return c.appPay(ctx, req)
	case payment.SceneH5:
		return c.h5Pay(ctx, req)
	case payment.SceneJSAPI:
		return c.jsapiPay(ctx, req)
	case payment.SceneNative:
		return c.nativePay(ctx, req)
	default:
		return c.appPay(ctx, req)
	}
}

// Helper functions to replace utils.StringPtr and utils.Int64Ptr
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

// appPay 实现App支付
func (c *Client) appPay(ctx context.Context, req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
	svc := app.AppApiService{Client: c.client}
	resp, result, err := svc.Prepay(ctx,
		app.PrepayRequest{
			Appid:       stringPtr(c.config.AppID),
			Mchid:       stringPtr(c.config.MchID),
			Description: stringPtr(req.Subject),
			OutTradeNo:  stringPtr(req.OutTradeNo),
			NotifyUrl:   stringPtr(req.NotifyURL),
			Amount:      &app.Amount{Total: int64Ptr(int64(req.TotalAmount * 100))},
			SceneInfo:   &app.SceneInfo{PayerClientIp: stringPtr("127.0.0.1")},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("wechat app pay failed: %w", err)
	}

	if result.Response.StatusCode != 200 {
		return nil, fmt.Errorf("wechat app pay failed with status: %d", result.Response.StatusCode)
	}

	return &payment.UnifiedPayResponse{
		Code:       "0",
		Message:    "success",
		OrderID:    req.OutTradeNo,
		OutTradeNo: req.OutTradeNo,
		PayData: map[string]string{
			"prepay_id":  *resp.PrepayId,
			"partner_id": c.config.MchID,
			"appid":      c.config.AppID,
		},
		Channel: payment.ChannelWechat,
	}, nil
}

// h5Pay 实现H5支付
func (c *Client) h5Pay(ctx context.Context, req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
	svc := h5.H5ApiService{Client: c.client}
	resp, result, err := svc.Prepay(ctx,
		h5.PrepayRequest{
			Appid:       stringPtr(c.config.AppID),
			Mchid:       stringPtr(c.config.MchID),
			Description: stringPtr(req.Subject),
			OutTradeNo:  stringPtr(req.OutTradeNo),
			NotifyUrl:   stringPtr(req.NotifyURL),
			Amount:      &h5.Amount{Total: int64Ptr(int64(req.TotalAmount * 100))},
			SceneInfo: &h5.SceneInfo{
				PayerClientIp: stringPtr("127.0.0.1"),
				H5Info:        &h5.H5Info{Type: stringPtr("iOS")},
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("wechat h5 pay failed: %w", err)
	}

	if result.Response.StatusCode != 200 {
		return nil, fmt.Errorf("wechat h5 pay failed with status: %d", result.Response.StatusCode)
	}

	return &payment.UnifiedPayResponse{
		Code:       "0",
		Message:    "success",
		OrderID:    req.OutTradeNo,
		OutTradeNo: req.OutTradeNo,
		PayData: map[string]string{
			"pay_url": *resp.H5Url,
		},
		Channel: payment.ChannelWechat,
	}, nil
}

// jsapiPay 实现JSAPI支付
func (c *Client) jsapiPay(ctx context.Context, req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
	if req.OpenID == "" {
		return nil, fmt.Errorf("openid is required for jsapi pay")
	}

	svc := jsapi.JsapiApiService{Client: c.client}
	resp, result, err := svc.Prepay(ctx,
		jsapi.PrepayRequest{
			Appid:       stringPtr(c.config.AppID),
			Mchid:       stringPtr(c.config.MchID),
			Description: stringPtr(req.Subject),
			OutTradeNo:  stringPtr(req.OutTradeNo),
			NotifyUrl:   stringPtr(req.NotifyURL),
			Amount:      &jsapi.Amount{Total: int64Ptr(int64(req.TotalAmount * 100))},
			Payer:       &jsapi.Payer{Openid: stringPtr(req.OpenID)},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("wechat jsapi pay failed: %w", err)
	}

	if result.Response.StatusCode != 200 {
		return nil, fmt.Errorf("wechat jsapi pay failed with status: %d", result.Response.StatusCode)
	}

	return &payment.UnifiedPayResponse{
		Code:       "0",
		Message:    "success",
		OrderID:    req.OutTradeNo,
		OutTradeNo: req.OutTradeNo,
		PayData: map[string]string{
			"prepay_id": *resp.PrepayId,
			"appid":     c.config.AppID,
		},
		Channel: payment.ChannelWechat,
	}, nil
}

// nativePay 实现Native支付
func (c *Client) nativePay(ctx context.Context, req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
	svc := native.NativeApiService{Client: c.client}
	resp, result, err := svc.Prepay(ctx,
		native.PrepayRequest{
			Appid:       stringPtr(c.config.AppID),
			Mchid:       stringPtr(c.config.MchID),
			Description: stringPtr(req.Subject),
			OutTradeNo:  stringPtr(req.OutTradeNo),
			NotifyUrl:   stringPtr(req.NotifyURL),
			Amount:      &native.Amount{Total: int64Ptr(int64(req.TotalAmount * 100))},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("wechat native pay failed: %w", err)
	}

	if result.Response.StatusCode != 200 {
		return nil, fmt.Errorf("wechat native pay failed with status: %d", result.Response.StatusCode)
	}

	return &payment.UnifiedPayResponse{
		Code:       "0",
		Message:    "success",
		OrderID:    req.OutTradeNo,
		OutTradeNo: req.OutTradeNo,
		PayData: map[string]string{
			"code_url": *resp.CodeUrl,
		},
		Channel: payment.ChannelWechat,
		QRCode:  *resp.CodeUrl,
	}, nil
}

// HandleNotify 处理异步通知
func (c *Client) HandleNotify(ctx context.Context, data []byte) (*payment.NotifyResult, error) {
	// Parse notification data using WeChat Pay SDK
	// In a real implementation, this would parse the actual WeChat Pay notification
	transaction := &payments.Transaction{}

	// For demo purposes, we'll simulate parsing the data
	// In a real implementation, you would use the SDK's notification parser

	result := &payment.NotifyResult{
		Success:     transaction.TradeState != nil && *transaction.TradeState == "SUCCESS",
		OutTradeNo:  "",
		TotalAmount: 0,
		TradeStatus: "",
		Channel:     payment.ChannelWechat,
		OrderID:     "",
	}

	if transaction.OutTradeNo != nil {
		result.OutTradeNo = *transaction.OutTradeNo
	}

	if transaction.Amount != nil && transaction.Amount.Total != nil {
		result.TotalAmount = float64(*transaction.Amount.Total) / 100
	}

	if transaction.TradeState != nil {
		result.TradeStatus = *transaction.TradeState
	}

	if transaction.TransactionId != nil {
		result.OrderID = *transaction.TransactionId
	}

	// Parse pay time if available
	if transaction.SuccessTime != nil {
		// SuccessTime is a *string, need to parse it to *time.Time
		if t, err := time.Parse(timeFormat, *transaction.SuccessTime); err == nil {
			result.PayTime = &t
		}
	}

	return result, nil
}

// GetChannel 获取渠道标识
func (c *Client) GetChannel() payment.ChannelType {
	return payment.ChannelWechat
}

// Refund 退款接口
func (c *Client) Refund(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	svc := refunddomestic.RefundsApiService{Client: c.client}
	resp, result, err := svc.Create(ctx,
		refunddomestic.CreateRequest{
			OutTradeNo:  stringPtr(req.OutTradeNo),
			OutRefundNo: stringPtr(req.OutRefundNo),
			Reason:      stringPtr(req.RefundReason),
			Amount: &refunddomestic.AmountReq{
				Refund:   int64Ptr(int64(req.RefundAmount * 100)),
				Total:    int64Ptr(int64(req.TotalAmount * 100)),
				Currency: stringPtr("CNY"),
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("wechat refund failed: %w", err)
	}

	if result.Response.StatusCode != 200 {
		return nil, fmt.Errorf("wechat refund failed with status: %d", result.Response.StatusCode)
	}

	status := "PROCESSING"
	if resp.Status != nil {
		status = string(*resp.Status)
	}

	refundResp := &payment.RefundResponse{
		Code:         "0",
		Message:      "success",
		OutRefundNo:  req.OutRefundNo,
		RefundAmount: req.RefundAmount,
		RefundStatus: status,
		Channel:      payment.ChannelWechat,
	}

	if resp.RefundId != nil {
		refundResp.RefundID = *resp.RefundId
	}

	if resp.Amount != nil && resp.Amount.Refund != nil {
		refundResp.RefundAmount = float64(*resp.Amount.Refund) / 100
	}

	if resp.SuccessTime != nil {
		// SuccessTime is a *string, need to parse it to *time.Time
		if t, err := time.Parse(timeFormat, *resp.SuccessTime); err == nil {
			refundResp.RefundTime = &t
		}
	}

	return refundResp, nil
}

// Close 关闭订单接口
func (c *Client) Close(ctx context.Context, req *payment.CloseRequest) error {
	svc := payments.NativeApiService{Client: c.client}
	result, err := svc.CloseOrder(ctx,
		payments.CloseOrderRequest{
			OutTradeNo: stringPtr(req.OutTradeNo),
			Mchid:      stringPtr(c.config.MchID),
		},
	)

	if err != nil {
		return fmt.Errorf("wechat close order failed: %w", err)
	}

	if result.Response.StatusCode != 204 && result.Response.StatusCode != 200 {
		return fmt.Errorf("wechat close order failed with status: %d", result.Response.StatusCode)
	}

	return nil
}

// Query 查询订单
func (c *Client) Query(ctx context.Context, req *payment.QueryRequest) (*payment.QueryResponse, error) {
	svc := payments.NativeApiService{Client: c.client}
	resp, result, err := svc.QueryOrderByOutTradeNo(ctx,
		payments.QueryOrderByOutTradeNoRequest{
			OutTradeNo: stringPtr(req.OutTradeNo),
			Mchid:      stringPtr(c.config.MchID),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("wechat query order failed: %w", err)
	}

	if result.Response.StatusCode != 200 {
		return nil, fmt.Errorf("wechat query order failed with status: %d", result.Response.StatusCode)
	}

	status := payment.TradeStatusNotPay
	if resp.TradeState != nil {
		switch *resp.TradeState {
		case "SUCCESS":
			status = payment.TradeStatusSuccess
		case "REFUND":
			status = payment.TradeStatusRefund
		case "NOTPAY":
			status = payment.TradeStatusNotPay
		case "CLOSED":
			status = payment.TradeStatusClosed
		case "REVOKED":
			status = payment.TradeStatusRevoked
		case "USERPAYING":
			status = payment.TradeStatusUserPaying
		case "PAYERROR":
			status = payment.TradeStatusPayError
		}
	}

	var payTime *time.Time
	if resp.SuccessTime != nil {
		// SuccessTime is a *string, need to parse it to *time.Time
		if t, err := time.Parse(timeFormat, *resp.SuccessTime); err == nil {
			payTime = &t
		}
	}

	queryResp := &payment.QueryResponse{
		Code:        "0",
		Message:     "success",
		OutTradeNo:  req.OutTradeNo,
		TradeStatus: status,
		Channel:     payment.ChannelWechat,
	}

	if resp.TransactionId != nil {
		queryResp.OrderID = *resp.TransactionId
	}

	if resp.Amount != nil && resp.Amount.Total != nil {
		queryResp.TotalAmount = float64(*resp.Amount.Total) / 100
	}

	queryResp.PayTime = payTime

	return queryResp, nil
}
