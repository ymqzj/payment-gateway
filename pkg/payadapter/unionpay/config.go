package unionpay

import (
	"crypto/rsa"

	"github.com/ymqzj/payment-gateway/configs"
)

// Config 银联配置
type Config struct {
	MerId      string
	AppId      string
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	Gateway    string
	FrontUrl   string
	BackUrl    string
}

// NewConfig 从全局配置创建银联配置
func NewConfig(config *configs.Config) *Config {
	return &Config{
		MerId:      config.UnionPay.MerID,
		AppId:      config.UnionPay.AppId,
		PrivateKey: config.UnionPay.PrivateKey,
		PublicKey:  config.UnionPay.PublicKey,
		Gateway:    config.UnionPay.Gateway,
		BackUrl:    config.UnionPay.BackURL,
		FrontUrl:   config.UnionPay.FrontURL,
	}
}
