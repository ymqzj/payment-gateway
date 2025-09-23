// pkg/payadapter/alipay/config.go
package alipay

type Config struct {
	AppID           string // 应用ID
	PrivateKey      string // 应用私钥（PKCS#1 或 PKCS#8 格式字符串）
	AlipayPublicKey string // 支付宝公钥（用于验签）
	NotifyURL       string // 默认异步通知地址
	ReturnURL       string // 默认同步跳转地址（H5用）
	IsSandbox       bool   // 是否沙箱环境
}
