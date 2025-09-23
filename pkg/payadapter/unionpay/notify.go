// notify.go
package unionpay

import (
	"errors"
	"net/http"

	"github.com/bytedance/gopkg/util/logger"
	"go.uber.org/zap"
)

func (c *Client) HandleNotify(r *http.Request) (map[string]string, error) {
	log := logger.With(
		logger.String("trace_id", r.FormValue("traceId")),
		logger.String("out_trade_no", r.FormValue("outTradeNo")),
	)

	err := r.ParseForm()
	if err != nil {
		log.Error("解析表单失败", logger.Error(err))
		return nil, err
	}

	params := make(map[string]string)
	for k, v := range r.Form {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	// 验签
	signature := params["signature"]
	delete(params, "signature") // 验签时排除 signature 字段

	if !VerifySign(params, signature, c.PublicKey) {
		log.Error("银联回调验签失败", logger.Any("params", params))
		return nil, errors.New("invalid signature")
	}

	// 检查状态
	if params["respCode"] == "00" && params["queryRespCode"] == "00" {
		log.Info("银联支付成功回调",
			zap.String("orderId", params["orderId"]),
			zap.String("merOrderId", params["merOrderId"]),
			zap.String("txnAmt", params["txnAmt"]),
		)
	} else {
		log.Warn("银联支付未成功",
			zap.String("respCode", params["respCode"]),
			zap.String("respMsg", params["respMsg"]),
		)
	}

	return params, nil
}
