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

	"github.com/bytedance/gopkg/util/logger"
	"github.com/ymqzj/payment-gateway/pkg/payadapter/unionpay"
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

type CreateOrderRequest struct {
	OutTradeNo  string  // 商户订单号
	TotalAmount float64 // 金额（元）
	Subject     string  // 商品标题
	Body        string  // 商品描述
	NotifyUrl   string  // 异步回调
	ReturnUrl   string  // 同步跳转
	TraceID     string  // 全链路追踪ID
}

type CreateOrderResponse struct {
	Tn         string `json:"tn"`      // 交易流水号，前端用它唤起支付
	OrderId    string `json:"orderId"` // 银联订单号
	ResultCode string `json:"result_code"`
	ResultMsg  string `json:"result_msg"`
}

func (c *Client) CreateOrder(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error) {
	log := logger.With(
		zap.String("trace_id", req.TraceID),
		zap.String("out_trade_no", req.OutTradeNo),
	)

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
	sign, err := unionpay.GenerateSign(params, c.PrivateKey)
	if err != nil {
		log.Error("生成签名失败", zap.Error(err))
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
		log.Error("请求银联网关失败", zap.Error(err))
		return nil, fmt.Errorf("request unionpay failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("读取响应体失败", zap.Error(err))
		return nil, err
	}

	var result map[string]string
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Error("解析银联响应失败", zap.ByteString("body", body), zap.Error(err))
		return nil, err
	}

	// 验签
	if !unionpay.VerifySign(result, result["signature"], c.PublicKey) {
		log.Error("银联响应验签失败", zap.Any("response", result))
		return nil, errors.New("verify signature failed")
	}

	if result["respCode"] != "00" {
		log.Error("银联下单失败",
			zap.String("respCode", result["respCode"]),
			zap.String("respMsg", result["respMsg"]),
		)
		return nil, fmt.Errorf("unionpay error: %s", result["respMsg"])
	}

	log.Info("银联下单成功", zap.String("tn", result["tn"]))

	return &CreateOrderResponse{
		Tn:         result["tn"],
		OrderId:    result["orderId"],
		ResultCode: result["respCode"],
		ResultMsg:  result["respMsg"],
	}, nil
}
