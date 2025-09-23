package wechat

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/wechatpay-apiv3/wechatpay-go/payments"

	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

func handleNotify(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("read body failed, err: %v", err)
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}

	// 验证签名
	verifyResult, err := utils.VerifySHA256WithRSA(r.Header.Get("Wechatpay-Serial"), r.Header.Get("Wechatpay-Signature"), r.Header.Get("Wechatpay-Timestamp"), r.Header.Get("Wechatpay-Nonce"), string(body), publicKey)
	if err != nil || !verifyResult {
		http.Error(w, "signature verify failed", http.StatusForbidden)
		return
	}

	// 解析通知内容
	var notify payments.Transaction
	if err := json.Unmarshal(body, &notify); err != nil {
		logger.Errorf("unmarshal notify failed, err: %v", err)
		http.Error(w, "unmarshal notify failed", http.StatusBadRequest)
		return
	}
	logger.Infof("handle notify, notify: %v", notify)
	// 处理通知逻辑
	// ...
}
