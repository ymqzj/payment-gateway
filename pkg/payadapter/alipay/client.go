// pkg/payadapter/alipay/client.go
package alipay

import (
	"log"

	"github.com/smartwalle/alipay/v3"
)

type Client struct {
	Client *alipay.Client
	Config *Config
}

func NewClient(conf *Config) *Client {
	client, err := alipay.New(conf.AppID, conf.PrivateKey, conf.IsSandbox)
	if err != nil {
		log.Fatal("❌ 初始化支付宝客户端失败:", err)
	}

	// 设置默认回调地址
	client.SetNotifyURL(conf.NotifyURL)
	client.SetReturnURL(conf.ReturnURL)
	client.SetCharset("utf-8")
	client.SetSignType(alipay.RSA2)

	// 设置支付宝公钥用于验签
	if conf.AlipayPublicKey != "" {
		client.LoadAliPayPublicKey(conf.AlipayPublicKey)
	}

	return &Client{
		Client: client,
		Config: conf,
	}
}
