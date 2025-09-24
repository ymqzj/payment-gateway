package payment

type UnifiedPayRequest struct {
	Channel     string  // "wechat", "alipay", "unionpay"
	OutTradeNo  string  // 商户订单号
	TotalAmount float64 // 金额（元）
	Subject     string  // 商品标题
	Body        string  // 商品描述
	NotifyURL   string  // 异步通知地址
	ReturnURL   string  // 同步跳转地址（H5用）
	Scene       string  // 支付场景: "app", "h5", "jsapi", "native" 等
	OpenID      string  // 微信 JSAPI 必填
}

// UnifiedPayResponse 统一支付响应
type UnifiedPayResponse struct {
	Code    string // 0=成功，其他失败
	Message string
	OrderID string      // 渠道订单号
	PayData interface{} // 支付所需数据：如 app 参数、跳转 url、二维码等
}

type NotifyResult struct {
	Success     bool
	OutTradeNo  string
	TotalAmount float64
	TradeStatus string
}
