package v1

import (
	"net/http"
	"time"

	"github.com/ymqzj/payment-gateway/internal/payment"

	"github.com/gin-gonic/gin"
)

// PaymentHandler 支付处理器
type PaymentHandler struct {
	gateway *payment.PaymentGateway
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler(gateway *payment.PaymentGateway) *PaymentHandler {
	return &PaymentHandler{
		gateway: gateway,
	}
}

// PayRequest 支付请求
type PayRequest struct {
	Channel     string  `json:"channel" binding:"required"`
	OutTradeNo  string  `json:"out_trade_no" binding:"required"`
	TotalAmount float64 `json:"total_amount" binding:"required,gt=0"`
	Subject     string  `json:"subject" binding:"required"`
	Scene       string  `json:"scene" binding:"required"`
	NotifyURL   string  `json:"notify_url" binding:"required"`
	ReturnURL   string  `json:"return_url,omitempty"`
	OpenID      string  `json:"openid,omitempty"`
	BuyerID     string  `json:"buyer_id,omitempty"`
	Attach      string  `json:"attach,omitempty"`
}

// PayResponse 支付响应
type PayResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// Pay 统一支付接口
func (h *PaymentHandler) Pay(c *gin.Context) {
	var req PayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, PayResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 转换渠道类型
	channel := payment.ChannelType(req.Channel)
	if !channel.IsValid() {
		c.JSON(http.StatusBadRequest, PayResponse{
			Code:    400,
			Message: "invalid channel",
		})
		return
	}

	// 转换场景类型
	scene := payment.PayScene(req.Scene)
	if !scene.IsValid() {
		c.JSON(http.StatusBadRequest, PayResponse{
			Code:    400,
			Message: "invalid scene",
		})
		return
	}

	// 构建支付请求
	payReq := &payment.UnifiedPayRequest{
		Channel:     channel,
		OutTradeNo:  req.OutTradeNo,
		TotalAmount: req.TotalAmount,
		Subject:     req.Subject,
		Scene:       scene,
		NotifyURL:   req.NotifyURL,
		ReturnURL:   req.ReturnURL,
		OpenID:      req.OpenID,
		Attach:      req.Attach,
	}

	// 调用支付网关
	resp, err := h.gateway.Pay(c.Request.Context(), payReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PayResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PayResponse{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"order_id":     resp.OrderID,
			"out_trade_no": resp.OutTradeNo,
			"pay_data":     resp.PayData,
			"qr_code":      resp.QRCode,
			"pay_url":      resp.PayURL,
			"channel":      resp.Channel,
		},
	})
}

// QueryRequest 查询请求
type QueryRequest struct {
	Channel    string `json:"channel" binding:"required"`
	OrderID    string `json:"order_id,omitempty"`
	OutTradeNo string `json:"out_trade_no,omitempty"`
}

// Query 订单查询接口
func (h *PaymentHandler) Query(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, PayResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	channel := payment.ChannelType(req.Channel)
	if !channel.IsValid() {
		c.JSON(http.StatusBadRequest, PayResponse{
			Code:    400,
			Message: "invalid channel",
		})
		return
	}

	queryReq := &payment.QueryRequest{
		Channel:    channel,
		OrderID:    req.OrderID,
		OutTradeNo: req.OutTradeNo,
	}

	resp, err := h.gateway.Query(c.Request.Context(), queryReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PayResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PayResponse{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"order_id":     resp.OrderID,
			"out_trade_no": resp.OutTradeNo,
			"trade_status": resp.TradeStatus,
			"total_amount": resp.TotalAmount,
			"pay_time":     resp.PayTime,
			"channel":      resp.Channel,
		},
	})
}

// RefundRequest 退款请求
type RefundRequest struct {
	Channel      string  `json:"channel" binding:"required"`
	OrderID      string  `json:"order_id,omitempty"`
	OutTradeNo   string  `json:"out_trade_no,omitempty"`
	OutRefundNo  string  `json:"out_refund_no" binding:"required"`
	RefundAmount float64 `json:"refund_amount" binding:"required,gt=0"`
	TotalAmount  float64 `json:"total_amount" binding:"required,gt=0"`
	RefundReason string  `json:"refund_reason,omitempty"`
}

// Refund 退款接口
func (h *PaymentHandler) Refund(c *gin.Context) {
	var req RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, PayResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	channel := payment.ChannelType(req.Channel)
	if !channel.IsValid() {
		c.JSON(http.StatusBadRequest, PayResponse{
			Code:    400,
			Message: "invalid channel",
		})
		return
	}

	refundReq := &payment.RefundRequest{
		Channel:      channel,
		OrderID:      req.OrderID,
		OutTradeNo:   req.OutTradeNo,
		OutRefundNo:  req.OutRefundNo,
		RefundAmount: req.RefundAmount,
		TotalAmount:  req.TotalAmount,
		RefundReason: req.RefundReason,
	}

	resp, err := h.gateway.Refund(c.Request.Context(), refundReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PayResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PayResponse{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"refund_id":     resp.RefundID,
			"out_refund_no": resp.OutRefundNo,
			"refund_amount": resp.RefundAmount,
			"refund_status": resp.RefundStatus,
			"refund_time":   resp.RefundTime,
			"channel":       resp.Channel,
		},
	})
}

// CloseRequest 关闭订单请求
type CloseRequest struct {
	Channel    string `json:"channel" binding:"required"`
	OrderID    string `json:"order_id,omitempty"`
	OutTradeNo string `json:"out_trade_no,omitempty"`
}

// Close 关闭订单接口
func (h *PaymentHandler) Close(c *gin.Context) {
	c.JSON(http.StatusOK, PayResponse{
		Code:    0,
		Message: "success",
	})
}

// GetChannels 获取支持的支付渠道
func (h *PaymentHandler) GetChannels(c *gin.Context) {
	channels := h.gateway.GetSupportedChannels()
	channelNames := make([]string, len(channels))
	for i, ch := range channels {
		channelNames[i] = string(ch)
	}

	c.JSON(http.StatusOK, PayResponse{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"channels": channelNames,
		},
	})
}

// Health 健康检查接口
func (h *PaymentHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, PayResponse{
		Code:    0,
		Message: "ok",
		Data: map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		},
	})
}

// HandleNotify 处理通知
func (h *PaymentHandler) HandleNotify(c *gin.Context) {
	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, PayResponse{
			Code:    400,
			Message: "channel parameter is required",
		})
		return
	}

	// 读取请求体
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, PayResponse{
			Code:    400,
			Message: "failed to read request body",
		})
		return
	}

	// 处理通知
	result, err := h.gateway.HandleNotify(c.Request.Context(), payment.ChannelType(channel), body)
	if err != nil {
		c.JSON(http.StatusBadRequest, PayResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, PayResponse{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"success":      result.Success,
			"out_trade_no": result.OutTradeNo,
			"total_amount": result.TotalAmount,
			"trade_status": result.TradeStatus,
			"channel":      result.Channel,
			"order_id":     result.OrderID,
			"pay_time":     result.PayTime,
		},
	})
}
