// pkg/payadapter/alipay/notify.go
package alipay

import (
	"github.com/ymqzj/payment-gateway/internal/payment"
)

func (c *Client) HandleNotify(data []byte) (*payment.NotifyResult, error) {
	// 解析并验签通知
	noti, err := c.Client.GetTradeNotification(data)
	if err != nil {
		logger.Errorf("handle notify failed, err: %v", err)
		return nil, err
	}

	result := &payment.NotifyResult{
		Success:     noti.TradeStatus == "TRADE_SUCCESS" || noti.TradeStatus == "TRADE_FINISHED",
		OutTradeNo:  noti.OutTradeNo,
		TotalAmount: noti.TotalAmount,
		TradeStatus: noti.TradeStatus,
	}

	return result, nil
}
