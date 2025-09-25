package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ymqzj/payment-gateway/configs"
	"github.com/ymqzj/payment-gateway/internal/payment"
	"github.com/ymqzj/payment-gateway/pkg/payadapter/alipay"
	"github.com/ymqzj/payment-gateway/pkg/payadapter/unionpay"
	"github.com/ymqzj/payment-gateway/pkg/payadapter/wechat"
)

func main() {
	// 加载配置
	cfg := configs.Load(configs.GetEnv())

	// 创建支付适配器
	wechatAdapter, err := wechat.NewAdapter(cfg)
	if err != nil {
		log.Printf("Failed to create wechat adapter: %v", err)
	}

	alipayAdapter, err := alipay.NewAdapter(cfg)
	if err != nil {
		log.Printf("Failed to create alipay adapter: %v", err)
	}

	unionpayAdapter, err := unionpay.NewAdapter(cfg)
	if err != nil {
		log.Printf("Failed to create unionpay adapter: %v", err)
	}

	// 创建支付网关
	var adapters []payment.PaymentAdapter
	if wechatAdapter != nil {
		adapters = append(adapters, wechatAdapter)
	}
	if alipayAdapter != nil {
		adapters = append(adapters, alipayAdapter)
	}
	if unionpayAdapter != nil {
		adapters = append(adapters, unionpayAdapter)
	}

	gateway := payment.NewPaymentGateway(adapters...)

	// 演示使用
	fmt.Println("🚀 支付网关启动成功")
	fmt.Printf("支持的渠道: %v\n", gateway.GetSupportedChannels())

	// 创建测试订单
	ctx := context.Background()

	// 微信支付测试
	if wechatAdapter != nil {
		fmt.Println("\n📱 测试微信支付...")
		payReq := &payment.UnifiedPayRequest{
			Channel:     payment.ChannelWechat,
			OutTradeNo:  "TEST_" + time.Now().Format("20060102150405"),
			TotalAmount: 0.01,
			Subject:     "测试支付1分钱",
			Scene:       payment.SceneApp,
			NotifyURL:   "https://example.com/notify",
		}

		resp, err := gateway.Pay(ctx, payReq)
		if err != nil {
			log.Printf("微信支付失败: %v", err)
		} else {
			fmt.Printf("✅ 微信支付参数准备完毕: %+v\n", resp.PayData)
		}
	}

	// 支付宝支付测试
	if alipayAdapter != nil {
		fmt.Println("\n💰 测试支付宝支付...")
		payReq := &payment.UnifiedPayRequest{
			Channel:     payment.ChannelAlipay,
			OutTradeNo:  "TEST_" + time.Now().Format("20060102150405"),
			TotalAmount: 0.01,
			Subject:     "测试支付1分钱",
			Scene:       payment.SceneApp,
			NotifyURL:   "https://example.com/notify",
		}

		resp, err := gateway.Pay(ctx, payReq)
		if err != nil {
			log.Printf("支付宝支付失败: %v", err)
		} else {
			fmt.Printf("✅ 支付宝支付参数准备完毕: %+v\n", resp.PayData)
		}
	}

	// 银联支付测试
	if unionpayAdapter != nil {
		fmt.Println("\n💳 测试银联支付...")
		payReq := &payment.UnifiedPayRequest{
			Channel:     payment.ChannelUnionPay,
			OutTradeNo:  "TEST_" + time.Now().Format("20060102150405"),
			TotalAmount: 0.01,
			Subject:     "测试支付1分钱",
			Scene:       payment.SceneApp,
			NotifyURL:   "https://example.com/notify",
		}

		resp, err := gateway.Pay(ctx, payReq)
		if err != nil {
			log.Printf("银联支付失败: %v", err)
		} else {
			fmt.Printf("✅ 银联支付参数准备完毕: %+v\n", resp.PayData)
		}
	}

	// 查询订单测试
	if wechatAdapter != nil {
		fmt.Println("\n🔍 测试订单查询...")
		queryReq := &payment.QueryRequest{
			Channel:    payment.ChannelWechat,
			OutTradeNo: "TEST_ORDER",
		}

		queryResp, err := gateway.Query(ctx, queryReq)
		if err != nil {
			log.Printf("查询失败: %v", err)
		} else {
			fmt.Printf("✅ 订单状态: %s\n", queryResp.TradeStatus)
		}
	}

	fmt.Println("\n🎉 支付网关测试完成")
}
