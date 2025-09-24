package alipay

import (
	"fmt"
	"github.com/ascoders/alipay"
	"github.com/ymqzj/payment-gateway/internal/payment"
)
type Client struct {
	alipay *alipay.Client
}
func (c *Client) PayApp(req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
	form := alipay.Form(alipay.Options{
	OrderId:  "123",	// 唯一订单号
	Fee:      99.8,		// 价格
	NickName: "翱翔大空",	// 用户昵称，支付页面显示用
	Subject:  "充值100",	// 支付描述，支付页面显示用
})

    payUrl, err := form.GetPayURL()
    if err != nil {
        return nil, fmt.Errorf("TradeAppPay failed: %w", err)
    }

    return &payment.UnifiedPayResponse{
        Code:    "0",
        Message: "success",
        OrderID: req.OutTradeNo,
        PayData: map[string]string{
            "orderString": payUrl.Encode(), // ⚡ 这里返回给前端 App SDK 直接使用
        },
    }, nil
}