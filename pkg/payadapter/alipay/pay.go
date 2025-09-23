// pkg/payadapter/alipay/pay.go
package alipay

import (
	"fmt"
	"strconv"

	"github.com/ymqzj/payment-gateway/internal/payment"
)

var logger = logger.GetLogger()

func (c *Client) Pay(req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
	switch req.Scene {
	case payment.SceneApp:
		return c.payApp(req)
	case payment.SceneH5:
		return c.payH5(req)
	case payment.SceneJSAPI: // 小程序也走这个
		return c.payMiniProgram(req)
	case payment.SceneNative:
		return c.payPage(req) // PC 扫码支付
	default:
		logger.Errorf("unsupported scene: %s", req.Scene)
		return nil, fmt.Errorf("unsupported scene: %s", req.Scene)
	}
}

// APP 支付
func (c *Client) payApp(req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
	p := c.Client.TradeAppPay(nil)
	p.Subject = req.Subject
	p.OutTradeNo = req.OutTradeNo
	p.TotalAmount = strconv.FormatFloat(req.TotalAmount, 'f', 2, 64)
	p.NotifyURL = req.NotifyURL
	p.ProductCode = "QUICK_MSECURITY_PAY" // 固定值

	url, err := p.String()
	if err != nil {
		return &payment.UnifiedPayResponse{Code: "1", Message: "生成支付参数失败: " + err.Error()}, nil
	}

	return &payment.UnifiedPayResponse{
		Code:    "0",
		Message: "success",
		OrderID: req.OutTradeNo,
		PayData: map[string]string{
			"orderString": url, // 前端直接传给 AlipaySDK
		},
	}, nil
}

// H5 支付（手机浏览器）
func (c *Client) payH5(req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
	p := c.Client.TradeWapPay(nil)
	p.Subject = req.Subject
	p.OutTradeNo = req.OutTradeNo
	p.TotalAmount = strconv.FormatFloat(req.TotalAmount, 'f', 2, 64)
	p.NotifyURL = req.NotifyURL
	p.ReturnURL = req.ReturnURL
	p.ProductCode = "QUICK_WAP_WAY"

	url, err := p.String()
	if err != nil {
		return &payment.UnifiedPayResponse{Code: "1", Message: "生成H5支付链接失败: " + err.Error()}, nil
	}

	return &payment.UnifiedPayResponse{
		Code:    "0",
		Message: "success",
		OrderID: req.OutTradeNo,
		PayData: map[string]string{
			"pay_url": url, // 前端 window.location.href 跳转
		},
	}, nil
}

// 小程序支付（JSAPI）
func (c *Client) payMiniProgram(req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
	// 小程序需调用 “交易创建” 接口，返回 pay_data 给前端 my.tradePay
	p := c.Client.TradeCreate(nil)
	p.Subject = req.Subject
	p.OutTradeNo = req.OutTradeNo
	p.TotalAmount = strconv.FormatFloat(req.TotalAmount, 'f', 2, 64)
	p.NotifyURL = req.NotifyURL
	p.BuyerID = req.OpenID // 小程序用户的 alipay_user_id（不是 openid！）

	// 创建订单
	resp, err := p.Do()
	if err != nil {
		logger.Errorf("create trade failed, err: %v", err)
		return &payment.UnifiedPayResponse{Code: "1", Message: "创建小程序订单失败: " + err.Error()}, nil
	}

	if resp.Code != "10000" || resp.Msg != "Success" {
		logger.Errorf("create trade failed, err: %v, resp: %v", err, resp)
		return &payment.UnifiedPayResponse{
			Code:    "1",
			Message: fmt.Sprintf("支付宝返回错误: %s - %s", resp.Code, resp.Msg),
		}, nil
	}

	// 返回 trade_no 和前端拉起支付所需参数
	payData := map[string]interface{}{
		"trade_no": resp.TradeNo,
		"orderStr": resp.OrderStr, // 小程序 my.tradePay 需要的 orderStr
	}

	return &payment.UnifiedPayResponse{
		Code:    "0",
		Message: "success",
		OrderID: resp.TradeNo,
		PayData: payData,
	}, nil
}

// PC 扫码支付
func (c *Client) payPage(req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
	p := c.Client.TradePagePay(nil)
	p.Subject = req.Subject
	p.OutTradeNo = req.OutTradeNo
	p.TotalAmount = strconv.FormatFloat(req.TotalAmount, 'f', 2, 64)
	p.NotifyURL = req.NotifyURL
	p.ReturnURL = req.ReturnURL
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	form, err := p.BuildForm()
	if err != nil {
		logger.Errorf("create trade page failed, err: %v", err)
		return &payment.UnifiedPayResponse{Code: "1", Message: "生成扫码支付表单失败: " + err.Error()}, nil
	}

	return &payment.UnifiedPayResponse{
		Code:    "0",
		Message: "success",
		OrderID: req.OutTradeNo,
		PayData: map[string]string{
			"html_form": form, // 前端直接 document.write(form) 渲染
		},
	}, nil
}
