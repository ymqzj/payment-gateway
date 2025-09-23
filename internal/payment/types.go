package payment

// ChannelType 支付渠道类型
type ChannelType string

const (
	ChannelWechat   ChannelType = "wechat"
	ChannelAlipay   ChannelType = "alipay"
	ChannelUnionPay ChannelType = "unionpay"
)

// PayScene 支付场景
type PayScene string

const (
	SceneApp    PayScene = "app"
	SceneH5     PayScene = "h5"
	SceneJSAPI  PayScene = "jsapi"
	SceneNative PayScene = "native"
	ScenePC     PayScene = "pc"
)

// TradeStatus 交易状态
type TradeStatus string

const (
	TradeStatusSuccess    TradeStatus = "SUCCESS"    // 支付成功
	TradeStatusRefund     TradeStatus = "REFUND"     // 已退款
	TradeStatusNotPay     TradeStatus = "NOTPAY"     // 未支付
	TradeStatusClosed     TradeStatus = "CLOSED"     // 已关闭
	TradeStatusRevoked    TradeStatus = "REVOKED"    // 已撤销
	TradeStatusUserPaying TradeStatus = "USERPAYING" // 用户支付中
	TradeStatusPayError   TradeStatus = "PAYERROR"   // 支付失败
)

// Currency 货币类型
type Currency string

const (
	CurrencyCNY Currency = "CNY" // 人民币
	CurrencyUSD Currency = "USD" // 美元
)

// PayMethod 支付方式
type PayMethod string

const (
	PayMethodWechatPay  PayMethod = "WECHAT_PAY"
	PayMethodAlipay     PayMethod = "ALIPAY"
	PayMethodUnionPay   PayMethod = "UNION_PAY"
	PayMethodCreditCard PayMethod = "CREDIT_CARD"
	PayMethodDebitCard  PayMethod = "DEBIT_CARD"
)

// String 返回字符串表示
func (c ChannelType) String() string {
	return string(c)
}

// String 返回字符串表示
func (p PayScene) String() string {
	return string(p)
}

// String 返回字符串表示
func (t TradeStatus) String() string {
	return string(t)
}

// IsValid 检查渠道类型是否有效
func (c ChannelType) IsValid() bool {
	switch c {
	case ChannelWechat, ChannelAlipay, ChannelUnionPay:
		return true
	default:
		return false
	}
}

// IsValid 检查支付场景是否有效
func (p PayScene) IsValid() bool {
	switch p {
	case SceneApp, SceneH5, SceneJSAPI, SceneNative, ScenePC:
		return true
	default:
		return false
	}
}