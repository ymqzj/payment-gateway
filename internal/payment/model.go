package payment

import "time"

// UnifiedPayRequest 统一支付请求
type UnifiedPayRequest struct {
	Channel     ChannelType // "wechat", "alipay", "unionpay"
	OutTradeNo  string      // 商户订单号
	TotalAmount float64     // 金额（元）
	Subject     string      // 商品标题
	Body        string      // 商品描述
	NotifyURL   string      // 异步通知地址
	ReturnURL   string      // 同步跳转地址（H5用）
	Scene       PayScene    // 支付场景: "app", "h5", "jsapi", "native" 等
	OpenID      string      // 微信 JSAPI 必填
	Attach      string      // 附加数据
}

// UnifiedPayResponse 统一支付响应
type UnifiedPayResponse struct {
	Code       string // 0=成功，其他失败
	Message    string
	OrderID    string      // 渠道订单号
	OutTradeNo string      // 商户订单号
	PayData    interface{} // 支付所需数据：如 app 参数、跳转 url、二维码等
	Channel    ChannelType // 支付渠道
	QRCode     string      // 二维码链接
	PayURL     string      // 支付链接
}

// NotifyResult 通知结果
type NotifyResult struct {
	Success     bool
	OutTradeNo  string
	TotalAmount float64
	TradeStatus string
	Channel     ChannelType
	OrderID     string
	PayTime     *time.Time
}

// QueryRequest 查询请求
type QueryRequest struct {
	Channel    ChannelType
	OrderID    string
	OutTradeNo string
}

// QueryResponse 查询响应
type QueryResponse struct {
	Code        string
	Message     string
	OrderID     string
	OutTradeNo  string
	TradeStatus TradeStatus
	TotalAmount float64
	PayTime     *time.Time
	Channel     ChannelType
}

// RefundRequest 退款请求
type RefundRequest struct {
	Channel      ChannelType
	OrderID      string
	OutTradeNo   string
	OutRefundNo  string
	RefundAmount float64
	TotalAmount  float64
	RefundReason string
}

// RefundResponse 退款响应
type RefundResponse struct {
	Code         string
	Message      string
	RefundID     string
	OutRefundNo  string
	RefundAmount float64
	RefundStatus string
	RefundTime   *time.Time
	Channel      ChannelType
}

// CloseRequest 关闭订单请求
type CloseRequest struct {
	Channel    ChannelType
	OrderID    string
	OutTradeNo string
}
