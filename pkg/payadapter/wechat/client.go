package wechat

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

type Client struct {
	Client *core.Client
	MchID  string
}

func NewClient(conf *Config) *Client {
	mchPrivateKey, err := loadPrivateKey(conf.PrivateKey)
	if err != nil {
		logger.Fatalf("加载商户私钥失败: %v", err)
	}

	ctx := context.Background()
	opts := []core.ClientOption{
		core.WithMerchant(conf.MchID, conf.SerialNo, mchPrivateKey),
	}

	// 如果提供了证书路径，自动下载并更新平台证书（用于回调验签）
	if conf.CertFilePath != "" {
		certificateDownloader := downloader.NewCertificateDownloader(
			downloader.WithValidInterval(6 * time.Hour), // 每6小时刷新
		)
		opts = append(opts, core.WithCertDownloader(certificateDownloader))
		verifier := verifiers.NewSHA256WithRSAVerifier()
		opts = append(opts, core.WithVerifier(verifier))
	}

	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		logger.Fatalf("初始化微信支付客户端失败: %v", err)
	}

	return &Client{
		Client: client,
		MchID:  conf.MchID,
	}
}

// 加载商户私钥（支持从字符串或文件路径加载）
func loadPrivateKey(keyOrPath string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(keyOrPath))
	if block != nil {
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}

	keyData, err := utils.LoadPrivateKeyWithPath(keyOrPath)
	if err != nil {
		return nil, fmt.Errorf("无法加载私钥: %w", err)
	}
	return x509.ParsePKCS1PrivateKey(keyData)
}
