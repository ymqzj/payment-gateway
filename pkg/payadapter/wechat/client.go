package wechat
import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
)

type Client struct {
	Client *core.Client
	Appid  string
	MchID  string
}


func loadPrivateKey(file string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("decode private key error")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not rsa private key")
	}
	return rsaKey, nil
}

func NewClient(conf *Config) *Client {
	mchPrivateKey, err := loadPrivateKey(conf.PrivateKey)
	if err != nil {
		logger.Fatalf("加载商户私钥失败: %v", err)
	}

	ctx := context.Background()
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(conf.MchID, conf.SerialNo, mchPrivateKey, conf.APIv3Key),
	}

	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		logger.Fatalf("初始化微信支付客户端失败: %v", err)
	}

	return &Client{
		Client: client,
		Appid:  conf.AppID,
		MchID:  conf.MchID,
	}
}