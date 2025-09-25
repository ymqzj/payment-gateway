package wechat

import "github.com/ymqzj/payment-gateway/configs"

type Config struct {
	AppID        string // 公众号/小程序/APP 的 appid
	MchID        string // 商户号
	APIv3Key     string // APIv3 密钥（在微信商户平台设置）
	SerialNo     string // 证书序列号
	PrivateKey   string // 商户私钥 (pem 格式内容 或 路径)
	CertFilePath string // 平台证书路径（用于回调验签，可选）
}

// NewConfig 从全局配置创建微信配置
func NewConfig(config *configs.Config) *Config {
	return &Config{
		AppID:      config.Wechat.AppID,
		MchID:      config.Wechat.MchID,
		APIv3Key:   config.Wechat.APIV3Key,
		SerialNo:   config.Wechat.CertSerialNo,
		PrivateKey: config.Wechat.KeyPath,
	}
}
