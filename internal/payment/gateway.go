package payment

import (
	"errors"
	"fmt"

	"github.com/ymqzj/payment-gateway/pkg/payadapter/alipay"
	"github.com/ymqzj/payment-gateway/pkg/payadapter/unionpay"
	"github.com/ymqzj/payment-gateway/pkg/payadapter/wechat"
)

type PaymentGateway struct {
	WechatClient   *wechat.Client
	AlipayClient   *alipay.Client
	UnionpayClient *unionpay.Client
}

func NewPaymentGateway(wechatConf *wechat.Config, alipayConf *alipay.Config, unionpayConf *unionpay.Config) *PaymentGateway {
	return &PaymentGateway{
		WechatClient:   wechat.NewClient(wechatConf),
		AlipayClient:   alipay.NewClient(alipayConf),
		UnionpayClient: unionpay.NewClient(unionpayConf),
	}
}

func (g *PaymentGateway) Pay(req *UnifiedPayRequest) (*UnifiedPayResponse, error) {
	switch req.Channel {
	case "wechat":
		return g.WechatClient.Pay(req)
	case "alipay":
		return g.AlipayClient.Pay(req)
	case "unionpay":
		return g.UnionpayClient.Pay(req)
	default:
		return nil, errors.New("unsupported payment channel: " + req.Channel)
	}
}

// 验签 & 异步通知处理也可以统一在此层分发
func (g *PaymentGateway) HandleNotify(channel string, data []byte) (*NotifyResult, error) {
	switch channel {
	case "wechat":
		return g.WechatClient.HandleNotify(data)
	case "alipay":
		return g.AlipayClient.HandleNotify(data)
	case "unionpay":
		return g.UnionpayClient.HandleNotify(data)
	default:
		return nil, fmt.Errorf("unknown channel: %s", channel)
	}
}
