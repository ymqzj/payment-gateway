// pay.go
package unionpay

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ymqzj/payment-gateway/internal/payment"
	"go.uber.org/zap"
)

const (
	PROD_GATEWAY    = "https://gateway.unionpay.com/gateway/"
	SANDBOX_GATEWAY = "https://gateway.test.unionpay.com/gateway/"
)

type Client struct {
	MerId      string
	AppId      string
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey // 银联公钥，用于验签
	Gateway    string
	FrontUrl   string
	BackUrl    string
}

func NewClient(config *Config) *Client {
	gateway := SANDBOX_GATEWAY
	if config.Gateway == "prod" {
		gateway = PROD_GATEWAY
	}
	return &Client{
		MerId:      config.MerId,
		AppId:      config.AppId,
		PrivateKey: config.PrivateKey,
		PublicKey:  config.PublicKey,
		Gateway:    gateway,
		FrontUrl:   config.FrontUrl,
		BackUrl:    config.BackUrl,
	}
}

// NewClientWithKeys 创建新的客户端，自动从文件加载私钥和公钥
func NewClientWithKeys(config *Config) (*Client, error) {
	gateway := SANDBOX_GATEWAY
	if config.Gateway == "prod" {
		gateway = PROD_GATEWAY
	}

	// 加载私钥
	privateKey, err := LoadPrivateKeyFromFile(config.PrivateKeyPath, config.CertPwd)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	// 加载公钥
	publicKey, err := LoadPublicKeyFromFile(config.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	return &Client{
		MerId:      config.MerId,
		AppId:      config.AppId,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Gateway:    gateway,
		FrontUrl:   config.FrontUrl,
		BackUrl:    config.BackUrl,
	}, nil
}

type CreateOrderRequest struct {
	OutTradeNo  string  // 商户订单号
	TotalAmount float64 // 金额（元）
	Subject     string  // 商品标题
	Body        string  // 商品描述
	NotifyUrl   string  // 异步回调
	ReturnUrl   string  // 同步跳转
	TraceID     string  // 全链路追踪ID
	Scene       string  // 支付场景
}

type CreateOrderResponse struct {
	Tn         string `json:"tn"`      // 交易流水号，前端用它唤起支付
	OrderId    string `json:"orderId"` // 银联订单号
	ResultCode string `json:"result_code"`
	ResultMsg  string `json:"result_msg"`
}

func (c *Client) CreateOrder(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error) {
	// Create a simple logger
	log := zap.NewExample().Sugar()
	defer log.Sync()

	log.Info("开始创建银联支付订单")

	params := map[string]string{
		"version":      "5.1.0",
		"charset":      "UTF-8",
		"transType":    "01", // 消费
		"merId":        c.MerId,
		"appId":        c.AppId,
		"orderId":      req.OutTradeNo,
		"txnTime":      time.Now().Format("20060102150405"),
		"txnAmt":       fmt.Sprintf("%.0f", req.TotalAmount*100), // 单位：分
		"currencyCode": "156",
		"orderDesc":    req.Subject,
		"reqReserved":  req.Body,
		"backUrl":      c.BackUrl,
		"frontUrl":     c.FrontUrl,
		"channelType":  "07", // 07=移动端，08=PC
		"accessType":   "0",
		"customerInfo": "{}",
	}

	// 生成签名
	sign, err := GenerateSign(params, c.PrivateKey)
	if err != nil {
		log.Error("生成签名失败", err)
		return nil, fmt.Errorf("generate sign failed: %w", err)
	}
	params["signature"] = sign

	// POST 请求到银联
	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	resp, err := http.PostForm(c.Gateway+"api/PayTransReq.do", formData)
	if err != nil {
		log.Error("请求银联网关失败", err)
		return nil, fmt.Errorf("request unionpay failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("读取响应体失败", err)
		return nil, err
	}

	var result map[string]string
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Error("解析银联响应失败", string(body), err)
		return nil, err
	}

	// 验签
	if !VerifySign(result, result["signature"], c.PublicKey) {
		log.Error("银联响应验签失败", result)
		return nil, errors.New("verify signature failed")
	}

	if result["respCode"] != "00" {
		log.Error("银联下单失败",
			"respCode", result["respCode"],
			"respMsg", result["respMsg"],
		)
		return nil, fmt.Errorf("unionpay error: %s", result["respMsg"])
	}

	log.Info("银联下单成功", "tn", result["tn"])

	return &CreateOrderResponse{
		Tn:         result["tn"],
		OrderId:    result["orderId"],
		ResultCode: result["respCode"],
		ResultMsg:  result["respMsg"],
	}, nil
}

// Pay 实现支付接口
func (c *Client) Pay(ctx context.Context, req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
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
	resp, err := c.CreateOrder(ctx, unionpayReq)
	if err != nil {
		return &payment.UnifiedPayResponse{
			Code:    "1",
			Message: fmt.Sprintf("银联下单失败: %v", err),
		}, nil
	}

	// 根据支付场景返回不同的支付数据
	payData := c.buildPayData(resp, req.Scene)

	return &payment.UnifiedPayResponse{
		Code:       "0",
		Message:    "success",
		OrderID:    resp.OrderId,
		OutTradeNo: req.OutTradeNo,
		PayData:    payData,
		Channel:    payment.ChannelUnionPay,
	}, nil
}

// buildPayData 根据支付场景构建支付数据
func (c *Client) buildPayData(resp *CreateOrderResponse, scene payment.PayScene) interface{} {
	switch scene {
	case payment.SceneApp:
		// APP支付返回tn码
		return map[string]string{
			"tn": resp.Tn,
		}
	case payment.SceneH5, payment.ScenePC:
		// H5和PC支付返回支付URL
		return map[string]string{
			"pay_url": fmt.Sprintf("%s?tn=%s", c.Gateway, resp.Tn),
		}
	case payment.SceneNative:
		// 扫码支付返回二维码链接
		return map[string]string{
			"qr_code": fmt.Sprintf("%s?tn=%s", c.Gateway, resp.Tn),
		}
	default:
		// 默认返回tn码
		return map[string]string{
			"tn": resp.Tn,
		}
	}
}

// HandleNotify 处理异步通知
func (c *Client) HandleNotify(ctx context.Context, data []byte) (*payment.NotifyResult, error) {
	// 解析通知数据
	params := make(map[string]string)
	err := json.Unmarshal(data, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to parse notify data: %w", err)
	}

	// 验签
	signature := params["signature"]
	delete(params, "signature") // 验签时排除 signature 字段

	if !VerifySign(params, signature, c.PublicKey) {
		return nil, errors.New("invalid signature")
	}

	// 检查响应码
	respCode := params["respCode"]
	queryRespCode := params["queryRespCode"]

	// 构造返回结果
	result := &payment.NotifyResult{
		Channel: payment.ChannelUnionPay,
	}

	// 设置订单号
	if orderId := params["orderId"]; orderId != "" {
		result.OutTradeNo = orderId
	}

	// 设置银联订单号
	if tn := params["tn"]; tn != "" {
		result.OrderID = tn
	}

	// 解析支付金额
	if txnAmt := params["txnAmt"]; txnAmt != "" {
		// 银联金额单位是分，需要转换为元
		if amount, err := fmt.Sscanf(txnAmt, "%f", &result.TotalAmount); err == nil && amount == 1 {
			result.TotalAmount = result.TotalAmount / 100.0
		}
	}

	// 解析支付时间
	if payTimeStr := params["payTime"]; payTimeStr != "" {
		// 银联时间格式: YYYYMMDDHHMMSS
		if payTime, err := time.Parse("20060102150405", payTimeStr); err == nil {
			result.PayTime = &payTime
		}
	}

	// 判断支付是否成功
	if (respCode == "00" || respCode == "") && (queryRespCode == "00" || queryRespCode == "") {
		result.Success = true
		result.TradeStatus = string(payment.TradeStatusSuccess)
	} else {
		result.Success = false
		result.TradeStatus = string(payment.TradeStatusPayError)
	}

	return result, nil
}

// GetChannel 获取渠道标识
func (c *Client) GetChannel() payment.ChannelType {
	return payment.ChannelUnionPay
}
