// pkg/payadapter/wechat/pay.go
package wechat

import (
	"context"
	"crypto/rsa"
	"fmt"
	"strconv"
	"time"

	"github.com/google/martian/log"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/app"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/refund"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"github.com/ymqzj/payment-gateway/internal/payment"
)

func (c *Client) Pay(req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
	switch req.Scene {
	case payment.SceneApp:
		return c.payApp(req)
	case payment.SceneJSAPI:
		return c.payJSAPI(req)
	default:
		return nil, fmt.Errorf("unsupported scene: %s", req.Scene)
	}
}

func (c *Client) payApp(req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
	ctx := context.Background()
	svc := app.AppApiService{Client: c.Client}

	total, _ := strconv.ParseInt(strconv.FormatFloat(req.TotalAmount*100, 'f', 0, 64), 10, 64)

	resp, result, err := svc.PrepayWithRequestPayment(ctx, app.PrepayRequest{
		Appid:       core.String(c.MchID),
		Mchid:       core.String(c.MchID),
		Description: core.String(req.Subject),
		OutTradeNo:  core.String(req.OutTradeNo),
		NotifyUrl:   core.String(req.NotifyURL),
		Amount: &app.Amount{
			Total:    core.Int64(total),
			Currency: core.String("CNY"),
		},
	})

	if err != nil {
		return &payment.UnifiedPayResponse{
			Code:    "1",
			Message: fmt.Sprintf("微信App支付下单失败: %v, 原始响应: %s", err, result.Response.Body),
		}, nil
	}

	payData := map[string]interface{}{
		"appid":     *resp.Appid,
		"prepayid":  resp.PrepayId,
		"noncestr":  resp.NonceStr,
		"timestamp": resp.TimeStamp,
		"package":   "Sign=WXPay",
		"sign":      resp.Signature,
	}

	return &payment.UnifiedPayResponse{
		Code:    "0",
		Message: "success",
		OrderID: *resp.PrepayId, // 注意：此处不是 TransactionId（那是支付后才有）
		PayData: payData,
	}, nil
}

func (c *Client) payJSAPI(req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
	if req.OpenID == "" {
		return nil, fmt.Errorf("openid is required for JSAPI payment")
	}

	ctx := context.Background()
	svc := jsapi.JsapiApiService{Client: c.Client}

	total, _ := strconv.ParseInt(strconv.FormatFloat(req.TotalAmount*100, 'f', 0, 64), 10, 64)

	resp, result, err := svc.Prepay(ctx, jsapi.PrepayRequest{
		Appid:       core.String(c.Config.AppID),
		Mchid:       core.String(c.Config.MchID),
		Description: core.String(req.Subject),
		OutTradeNo:  core.String(req.OutTradeNo),
		NotifyUrl:   core.String(req.NotifyURL),
		Amount: &jsapi.Amount{
			Total:    core.Int64(total),
			Currency: core.String("CNY"),
		},
		Payer: &jsapi.Payer{
			Openid: core.String(req.OpenID),
		},
	})

	if err != nil {
		return &payment.UnifiedPayResponse{
			Code:    "1",
			Message: fmt.Sprintf("微信JSAPI支付下单失败: %v, 原始响应: %s", err, result.Response.Body),
		}, nil
	}

	// 构造给前端的支付参数（小程序或H5拉起支付用）
	paySign, err := c.generateJSSDKSignature(resp.PrepayId)
	if err != nil {
		return &payment.UnifiedPayResponse{
			Code:    "1",
			Message: "生成 paySign 失败: " + err.Error(),
		}, nil
	}

	payData := map[string]interface{}{
		"appId":     c.Config.AppID,
		"timeStamp": strconv.FormatInt(time.Now().Unix(), 10),
		"nonceStr":  resp.NonceStr,
		"package":   "prepay_id=" + resp.PrepayId,
		"signType":  "RSA",
		"paySign":   paySign,
	}

	return &payment.UnifiedPayResponse{
		Code:    "0",
		Message: "success",
		OrderID: resp.PrepayId,
		PayData: payData,
	}, nil
}

// generateJSSDKSignature 生成小程序/JSAPI 支付签名
func (c *Client) generateJSSDKSignature(prepayID string) (string, error) {
	message := fmt.Sprintf("%s\n%s\n%s\n", c.Config.AppID, prepayID, time.Now().Unix())
	signature, err := utils.SignSHA256WithRSA(message, c.mchPrivateKey())
	if err != nil {
		return "", err
	}
	return signature, nil
}

// 私钥辅助方法
func (c *Client) mchPrivateKey() *rsa.PrivateKey {
	key, _ := loadPrivateKey(c.Config.PrivateKey)
	return key
}

func queryOrder(client *core.Client, outTradeNo string) {
	svc := payments.PaymentsApiService{Client: client}
	resp, result, err := svc.QueryOrderByOutTradeNo(
		context.Background(),
		outTradeNo,
		c.MchID,
	)

	if err != nil {
		log.Errorf("query order err: %+v", err)
		return
	}

	if result.Response.StatusCode != 200 {
		fmt.Printf("query failed, statusCode=%d\n", result.Response.StatusCode)
		return
	}

	fmt.Printf("订单状态：%s\n", *resp.TradeState)
}

func refundOrder(client *core.Client) {
	svc := refund.RefundApiService{Client: client}
	resp, result, err := svc.CreateRefund(
		context.Background(),
		refund.CreateRefundRequest{
			TransactionId: core.String("your_wx_trade_id"),
			OutRefundNo:   core.String("REFUND_20240601_001"),
			Reason:        core.String("用户取消"),
			Amount: &refund.Amount{
				Refund:   core.Int64(1),
				Total:    core.Int64(1),
				Currency: core.String("CNY"),
			},
			NotifyUrl: core.String("https://yourdomain.com/refund_notify"),
		},
	)

	if err != nil {
		log.Errorf("refund err: %+v", err)
		return
	}

	if result.Response.StatusCode != 200 {
		fmt.Printf("refund failed, statusCode=%d\n", result.Response.StatusCode)
		return
	}

	log.Infof("refund success, refund id: %s", *resp.RefundId)
}
