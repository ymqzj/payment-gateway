package unionpay

import (
	"crypto/rsa"

	"github.com/ymqzj/payment-gateway/configs"
)

// Config 银联配置
type Config struct {
	MerId          string
	AppId          string
	CertPath       string
	CertPwd        string
	PrivateKeyPath string
	PublicKeyPath  string
	Gateway        string
	FrontUrl       string
	BackUrl        string
	PrivateKey     *rsa.PrivateKey
	PublicKey      *rsa.PublicKey
}

// NewConfig 从全局配置创建银联配置
func NewConfig(config *configs.Config) *Config {
	return &Config{
		MerId:          config.UnionPay.MerID,
		AppId:          config.UnionPay.AppId,
		CertPath:       config.UnionPay.CertPath,
		CertPwd:        config.UnionPay.CertPwd,
		PrivateKeyPath: config.UnionPay.PrivateKeyPath,
		PublicKeyPath:  config.UnionPay.PublicKeyPath,
		Gateway:        config.UnionPay.Gateway,
		BackUrl:        config.UnionPay.BackURL,
		FrontUrl:       config.UnionPay.FrontURL,
	}
}
