// pkg/payadapter/alipay/refund.go
package alipay

import (
	"context"
	"fmt"
	"strconv"

	"github.com/smartwalle/alipay/v3"
	"github.com/ymqzj/payment-gateway/internal/payment"
)

func (c *Client) Refund(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	var p = alipay.TradeRefund{}
	p.OutTradeNo = req.OutTradeNo
	p.RefundAmount = strconv.FormatFloat(req.RefundAmount, 'f', 2, 64)
	p.OutRequestNo = req.OutRefundNo // 幂等键
	p.RefundReason = req.RefundReason

	resp, err := p.RefundAmount(c.client)
	if err != nil {
		return &payment.RefundResponse{
			Code:    "1",
			Message: "退款请求失败: " + err.Error(),
		}, nil
	}

	if resp.Code != "10000" {
		return &payment.RefundResponse{
			Code:    "1",
			Message: fmt.Sprintf("支付宝退款失败: %s - %s", resp.Code, resp.Msg),
		}, nil
	}

	return &payment.RefundResponse{
		Code:         "0",
		Message:      "success",
		RefundID:     resp.RefundFee,
		OutRefundNo:  req.OutRefundNo,
		RefundAmount: req.RefundAmount,
		RefundStatus: "SUCCESS",
		Channel:      payment.ChannelAlipay,
	}, nil
}
