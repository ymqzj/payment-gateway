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

// Query method that was missing - this fixes the compilation error
// This is a temporary implementation that just returns an error
// A proper implementation would delegate to the appropriate adapter
func (g *PaymentGateway) Query(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	// This implementation fixes the compilation error but doesn't provide actual functionality
	// To properly implement this, we would need to add a Query method to the PaymentAdapter interface
	// and implement it in all adapter types
	return nil, fmt.Errorf("query method not yet implemented for channel: %s", req.Channel)
}