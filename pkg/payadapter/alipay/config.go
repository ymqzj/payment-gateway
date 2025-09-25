// pkg/payadapter/alipay/config.go
package alipay

import "github.com/ymqzj/payment-gateway/configs"

type Config struct {
	AppID           string // 应用ID
	PrivateKey      string // 应用私钥（PKCS#1 或 PKCS#8 格式字符串）
	AlipayPublicKey string // 支付宝公钥（用于验签）
	NotifyURL       string // 默认异步通知地址
	ReturnURL       string // 默认同步跳转地址（H5用）
	IsSandbox       bool   // 是否沙箱环境
}

// NewConfig 从全局配置创建支付宝配置
func NewConfig(config *configs.Config) *Config {
	return &Config{
		AppID:           config.Alipay.AppID,
		PrivateKey:      config.Alipay.PrivateKey,
		AlipayPublicKey: config.Alipay.AlipayPublicKey,
		NotifyURL:       config.Alipay.NotifyURL,
		ReturnURL:       config.Alipay.ReturnURL,
		IsSandbox:       false, // 可根据环境变量设置
	}
}