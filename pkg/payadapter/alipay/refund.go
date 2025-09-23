// pkg/payadapter/alipay/refund.go
package alipay

import (
	"fmt"
	"strconv"
)

type UnifiedRefundRequest struct {
	OutTradeNo   string  // 原支付订单号
	OutRefundNo  string  // 商户退款单号（幂等）
	RefundAmount float64 // 退款金额
	Reason       string  // 退款原因
}

type UnifiedRefundResponse struct {
	Code     string
	Message  string
	RefundID string // 支付宝退款单号
	Status   string // SUCCESS / FAILED
}

func (c *Client) Refund(req *UnifiedRefundRequest) (*UnifiedRefundResponse, error) {
	p := c.Client.TradeRefund(nil)
	p.OutTradeNo = req.OutTradeNo
	p.RefundAmount = strconv.FormatFloat(req.RefundAmount, 'f', 2, 64)
	p.OutRequestNo = req.OutRefundNo // 幂等键
	p.RefundReason = req.Reason

	resp, err := p.Do()
	if err != nil {
		return &UnifiedRefundResponse{
			Code:    "1",
			Message: "退款请求失败: " + err.Error(),
		}, nil
	}

	if resp.Code != "10000" {
		return &UnifiedRefundResponse{
			Code:    "1",
			Message: fmt.Sprintf("支付宝退款失败: %s - %s", resp.Code, resp.Msg),
		}, nil
	}

	return &UnifiedRefundResponse{
		Code:     "0",
		Message:  "success",
		RefundID: resp.RefundFee,
		Status:   "SUCCESS",
	}, nil
}
