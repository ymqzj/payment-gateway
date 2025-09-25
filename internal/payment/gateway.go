package payment

import (
	"context"
	"fmt"
)

// Define interfaces for the adapters to avoid import cycles
type PaymentAdapter interface {
	Pay(ctx context.Context, req *UnifiedPayRequest) (*UnifiedPayResponse, error)
	HandleNotify(ctx context.Context, data []byte) (*NotifyResult, error)
	GetChannel() ChannelType
	Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error)
	Query(ctx context.Context, req *QueryRequest) (*QueryResponse, error)
	Close(ctx context.Context, req *CloseRequest) error
}

type PaymentGateway struct {
	adapters map[ChannelType]PaymentAdapter
}

func NewPaymentGateway(adapters ...PaymentAdapter) *PaymentGateway {
	gateway := &PaymentGateway{
		adapters: make(map[ChannelType]PaymentAdapter),
	}

	for _, adapter := range adapters {
		gateway.adapters[adapter.GetChannel()] = adapter
	}

	return gateway
}

func (g *PaymentGateway) Pay(ctx context.Context, req *UnifiedPayRequest) (*UnifiedPayResponse, error) {
	adapter, exists := g.adapters[req.Channel]
	if !exists {
		return nil, fmt.Errorf("unsupported payment channel: %s", req.Channel)
	}

	return adapter.Pay(ctx, req)
}

func (g *PaymentGateway) HandleNotify(ctx context.Context, channel ChannelType, data []byte) (*NotifyResult, error) {
	adapter, exists := g.adapters[channel]
	if !exists {
		return nil, fmt.Errorf("unsupported payment channel: %s", channel)
	}

	return adapter.HandleNotify(ctx, data)
}

func (g *PaymentGateway) GetSupportedChannels() []ChannelType {
	channels := make([]ChannelType, 0, len(g.adapters))
	for channel := range g.adapters {
		channels = append(channels, channel)
	}
	return channels
}

// Refund method that was missing - this fixes the compilation error
func (g *PaymentGateway) Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	adapter, exists := g.adapters[req.Channel]
	if !exists {
		return nil, fmt.Errorf("unsupported payment channel: %s", req.Channel)
	}

	return adapter.Refund(ctx, req)
}

// Query method implementation
func (g *PaymentGateway) Query(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	adapter, exists := g.adapters[req.Channel]
	if !exists {
		return nil, fmt.Errorf("unsupported payment channel: %s", req.Channel)
	}

	return adapter.Query(ctx, req)
}

// Close method implementation
func (g *PaymentGateway) Close(ctx context.Context, req *CloseRequest) error {
	adapter, exists := g.adapters[req.Channel]
	if !exists {
		return fmt.Errorf("unsupported payment channel: %s", req.Channel)
	}

	return adapter.Close(ctx, req)
}
