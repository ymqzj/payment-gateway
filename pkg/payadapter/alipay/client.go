package alipay

import "github.com/ascoders/alipay"

func NewAliClient(config Config) *alipay.Client {
    client := alipay.Client{
	Partner   : config.AppID, // 合作者ID
	Key       : config.AlipayPublicKey, // 合作者私钥
	ReturnUrl : config.NotifyURL, // 同步返回地址
	NotifyUrl : config.ReturnURL, // 网站异步返回地址
}

    return &client
}