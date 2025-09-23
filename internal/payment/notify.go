package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// NotifyHandler 通知处理器接口
type NotifyHandler interface {
	// Handle 处理通知
	Handle(ctx context.Context, channel ChannelType, body []byte) (*NotifyResult, error)

	// Verify 验证通知签名
	Verify(ctx context.Context, channel ChannelType, body []byte) error

	// Respond 生成响应
	Respond(ctx context.Context, success bool, message string) ([]byte, error)
}

// NotifyManager 通知管理器
type NotifyManager struct {
	gateway    *PaymentGateway
	handlers   map[ChannelType]NotifyHandler
	mu         sync.RWMutex
	processors map[string]NotifyProcessor
}

// NotifyProcessor 通知处理器接口
type NotifyProcessor interface {
	Process(ctx context.Context, result *NotifyResult) error
}

// NewNotifyManager 创建通知管理器
func NewNotifyManager(gateway *PaymentGateway) *NotifyManager {
	return &NotifyManager{
		gateway:    gateway,
		handlers:   make(map[ChannelType]NotifyHandler),
		processors: make(map[string]NotifyProcessor),
	}
}

// RegisterHandler 注册通知处理器
func (nm *NotifyManager) RegisterHandler(channel ChannelType, handler NotifyHandler) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.handlers[channel] = handler
}

// RegisterProcessor 注册通知处理器
func (nm *NotifyManager) RegisterProcessor(name string, processor NotifyProcessor) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.processors[name] = processor
}

// HandleNotify 处理异步通知
func (nm *NotifyManager) HandleNotify(ctx context.Context, channel ChannelType, body []byte) (*NotifyResult, error) {
	// 获取适配器
	adapter, err := nm.gateway.GetAdapter(channel)
	if err != nil {
		return nil, fmt.Errorf("get adapter failed: %w", err)
	}

	// 处理通知
	result, err := adapter.HandleNotify(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("handle notify failed: %w", err)
	}

	// 执行通知处理器
	if err := nm.processNotify(ctx, result); err != nil {
		return nil, fmt.Errorf("process notify failed: %w", err)
	}

	return result, nil
}

// processNotify 处理通知结果
func (nm *NotifyManager) processNotify(ctx context.Context, result *NotifyResult) error {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	// 执行所有注册的处理器
	for name, processor := range nm.processors {
		if err := processor.Process(ctx, result); err != nil {
			return fmt.Errorf("processor %s failed: %w", name, err)
		}
	}

	return nil
}

// GetNotifyResponse 获取通知响应
func (nm *NotifyManager) GetNotifyResponse(ctx context.Context, channel ChannelType, success bool, message string) ([]byte, error) {
	// 获取处理器
	handler, exists := nm.handlers[channel]
	if !exists {
		return nil, ErrInvalidChannel
	}

	// 生成响应
	return handler.Respond(ctx, success, message)
}

// DefaultNotifyHandler 默认通知处理器
type DefaultNotifyHandler struct {
	gateway *PaymentGateway
}

// NewDefaultNotifyHandler 创建默认通知处理器
func NewDefaultNotifyHandler(gateway *PaymentGateway) *DefaultNotifyHandler {
	return &DefaultNotifyHandler{
		gateway: gateway,
	}
}

// Handle 处理通知
func (h *DefaultNotifyHandler) Handle(ctx context.Context, channel ChannelType, body []byte) (*NotifyResult, error) {
	return h.gateway.HandleNotify(ctx, channel, body)
}

// Verify 验证通知签名
func (h *DefaultNotifyHandler) Verify(ctx context.Context, channel ChannelType, body []byte) error {
	// 获取适配器
	adapter, err := h.gateway.GetAdapter(channel)
	if err != nil {
		return err
	}

	// 这里应该调用适配器的验证方法
	// 由于接口限制，暂时返回 nil
	_ = adapter
	return nil
}

// Respond 生成响应
func (h *DefaultNotifyHandler) Respond(ctx context.Context, success bool, message string) ([]byte, error) {
	response := map[string]interface{}{
		"code":    "SUCCESS",
		"message": "OK",
	}

	if !success {
		response["code"] = "FAIL"
		response["message"] = message
	}

	return json.Marshal(response)
}

// LoggingProcessor 日志处理器
type LoggingProcessor struct {
	logger Logger
}

// Logger 日志接口
type Logger interface {
	Info(ctx context.Context, msg string, fields ...interface{})
	Error(ctx context.Context, msg string, fields ...interface{})
	Warn(ctx context.Context, msg string, fields ...interface{})
}

// NewLoggingProcessor 创建日志处理器
func NewLoggingProcessor(logger Logger) *LoggingProcessor {
	return &LoggingProcessor{
		logger: logger,
	}
}

// Process 处理通知日志
func (p *LoggingProcessor) Process(ctx context.Context, result *NotifyResult) error {
	p.logger.Info(ctx, "processing notify",
		"channel", result.Channel,
		"order_id", result.OrderID,
		"out_trade_no", result.OutTradeNo,
		"trade_status", result.TradeStatus,
		"total_amount", result.TotalAmount,
		"pay_time", result.PayTime,
	)

	return nil
}

// OrderProcessor 订单处理器
type OrderProcessor struct {
	// 这里可以添加订单相关的依赖
}

// NewOrderProcessor 创建订单处理器
func NewOrderProcessor() *OrderProcessor {
	return &OrderProcessor{}
}

// Process 处理订单状态更新
func (p *OrderProcessor) Process(ctx context.Context, result *NotifyResult) error {
	// 这里可以实现订单状态更新逻辑
	// 例如：更新数据库、发送消息等

	// 示例：根据交易状态处理
	switch result.TradeStatus {
	case TradeStatusSuccess:
		// 处理支付成功
		return p.handlePaymentSuccess(ctx, result)
	case TradeStatusRefund:
		// 处理退款
		return p.handleRefund(ctx, result)
	case TradeStatusClosed:
		// 处理订单关闭
		return p.handleOrderClose(ctx, result)
	default:
		// 其他状态
		return nil
	}
}

func (p *OrderProcessor) handlePaymentSuccess(ctx context.Context, result *NotifyResult) error {
	// 实现支付成功处理逻辑
	// 例如：更新订单状态、发送通知、处理业务逻辑等
	return nil
}

func (p *OrderProcessor) handleRefund(ctx context.Context, result *NotifyResult) error {
	// 实现退款处理逻辑
	return nil
}

func (p *OrderProcessor) handleOrderClose(ctx context.Context, result *NotifyResult) error {
	// 实现订单关闭处理逻辑
	return nil
}

// MetricsProcessor 指标处理器
type MetricsProcessor struct {
	// 这里可以添加指标相关的依赖
}

// NewMetricsProcessor 创建指标处理器
func NewMetricsProcessor() *MetricsProcessor {
	return &MetricsProcessor{}
}

// Process 处理指标收集
func (p *MetricsProcessor) Process(ctx context.Context, result *NotifyResult) error {
	metrics := metrics.MustGet(ctx)

	// 示例：记录支付时间
	if result.PayTime != nil {
		latency := time.Since(*result.PayTime)
		_ = latency // 记录延迟指标
	}

	return nil
}

// NotifyResponse 通知响应
const (
	NotifyResponseSuccess = `{"code":"SUCCESS","message":"OK"}`
	NotifyResponseFail    = `{"code":"FAIL","message":"%s"}`
)

// HTTPHandler HTTP通知处理器
type HTTPHandler struct {
	notifyManager *NotifyManager
}

// NewHTTPHandler 创建HTTP处理器
func NewHTTPHandler(notifyManager *NotifyManager) *HTTPHandler {
	return &HTTPHandler{
		notifyManager: notifyManager,
	}
}

// ServeHTTP 处理HTTP通知
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 获取渠道
	channel := ChannelType(r.URL.Query().Get("channel"))
	if !channel.IsValid() {
		http.Error(w, "invalid channel", http.StatusBadRequest)
		return
	}

	// 读取请求体
	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	// 处理通知
	result, err := h.notifyManager.HandleNotify(ctx, channel, body)
	if err != nil {
		response := fmt.Sprintf(NotifyResponseFail, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(response))
		return
	}

	// 获取响应
	response, err := h.notifyManager.GetNotifyResponse(ctx, channel, true, "success")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(NotifyResponseFail, "internal error")))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)

	// 记录日志
	_ = result // 这里可以记录更详细的日志
}
