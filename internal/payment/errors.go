package payment

import (
	"errors"
	"fmt"
)

// 定义错误类型
var (
	// 通用错误
	ErrInvalidChannel      = errors.New("invalid payment channel")
	ErrInvalidScene        = errors.New("invalid payment scene")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrMissingParameter    = errors.New("missing required parameter")
	ErrInvalidParameter    = errors.New("invalid parameter")
	ErrSystemError         = errors.New("system error")
	ErrNetworkError        = errors.New("network error")
	ErrTimeout             = errors.New("request timeout")
	
	// 业务错误
	ErrOrderNotFound       = errors.New("order not found")
	ErrOrderClosed         = errors.New("order closed")
	ErrOrderPaid           = errors.New("order already paid")
	ErrOrderExpired        = errors.New("order expired")
	ErrRefundNotAllowed    = errors.New("refund not allowed")
	ErrInsufficientBalance = errors.New("insufficient balance")
	
	// 渠道特定错误
	ErrWechatError         = errors.New("wechat pay error")
	ErrAlipayError         = errors.New("alipay error")
	ErrUnionPayError       = errors.New("unionpay error")
	
	// 签名错误
	ErrInvalidSignature    = errors.New("invalid signature")
	ErrSignatureFailed     = errors.New("signature failed")
	
	// 通知错误
	ErrInvalidNotify       = errors.New("invalid notify data")
	ErrNotifyVerifyFailed  = errors.New("notify verification failed")
)

// ErrorCode 错误码类型
type ErrorCode struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error 实现 error 接口
func (e *ErrorCode) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewErrorCode 创建新的错误码
func NewErrorCode(code, message string) *ErrorCode {
	return &ErrorCode{
		Code:    code,
		Message: message,
	}
}

// NewErrorCodeWithDetails 创建带详细信息的错误码
func NewErrorCodeWithDetails(code, message, details string) *ErrorCode {
	return &ErrorCode{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// 预定义错误码
var (
	// 成功
	Success = NewErrorCode("0", "success")
	
	// 客户端错误 (4xx)
	BadRequest          = NewErrorCode("400", "bad request")
	Unauthorized        = NewErrorCode("401", "unauthorized")
	Forbidden           = NewErrorCode("403", "forbidden")
	NotFound            = NewErrorCode("404", "not found")
	MethodNotAllowed    = NewErrorCode("405", "method not allowed")
	RequestTimeout      = NewErrorCode("408", "request timeout")
	Conflict            = NewErrorCode("409", "conflict")
	UnprocessableEntity = NewErrorCode("422", "unprocessable entity")
	
	// 服务器错误 (5xx)
	InternalServerError = NewErrorCode("500", "internal server error")
	ServiceUnavailable  = NewErrorCode("503", "service unavailable")
	GatewayTimeout      = NewErrorCode("504", "gateway timeout")
	
	// 业务错误码
	InvalidChannel      = NewErrorCode("1001", "invalid payment channel")
	InvalidScene        = NewErrorCode("1002", "invalid payment scene")
	InvalidAmount       = NewErrorCode("1003", "invalid amount")
	MissingParameter    = NewErrorCode("1004", "missing required parameter")
	InvalidParameter    = NewErrorCode("1005", "invalid parameter")
	OrderNotFound       = NewErrorCode("1006", "order not found")
	OrderClosed         = NewErrorCode("1007", "order closed")
	OrderPaid           = NewErrorCode("1008", "order already paid")
	OrderExpired        = NewErrorCode("1009", "order expired")
	RefundNotAllowed    = NewErrorCode("1010", "refund not allowed")
	InsufficientBalance = NewErrorCode("1011", "insufficient balance")
	
	// 渠道错误码
	WechatError   = NewErrorCode("2001", "wechat pay error")
	AlipayError   = NewErrorCode("2002", "alipay error")
	UnionPayError = NewErrorCode("2003", "unionpay error")
	
	// 签名错误码
	InvalidSignature = NewErrorCode("3001", "invalid signature")
	SignatureFailed  = NewErrorCode("3002", "signature failed")
	
	// 通知错误码
	InvalidNotify      = NewErrorCode("4001", "invalid notify data")
	NotifyVerifyFailed = NewErrorCode("4002", "notify verification failed")
)

// IsRetryable 判断错误是否可重试
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	
	// 网络错误、超时错误等可重试
	return errors.Is(err, ErrNetworkError) ||
		errors.Is(err, ErrTimeout) ||
		errors.Is(err, ErrSystemError)
}

// IsBusinessError 判断是否为业务错误
func IsBusinessError(err error) bool {
	if err == nil {
		return false
	}
	
	// 业务相关错误
	return errors.Is(err, ErrOrderNotFound) ||
		errors.Is(err, ErrOrderClosed) ||
		errors.Is(err, ErrOrderPaid) ||
		errors.Is(err, ErrOrderExpired) ||
		errors.Is(err, ErrRefundNotAllowed) ||
		errors.Is(err, ErrInsufficientBalance)
}