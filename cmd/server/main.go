package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	v1 "github.com/ymqzj/payment-gateway/api/v1"
	"github.com/ymqzj/payment-gateway/configs"
	"github.com/ymqzj/payment-gateway/internal/payment"
	"github.com/ymqzj/payment-gateway/pkg/payadapter/alipay"
	"github.com/ymqzj/payment-gateway/pkg/payadapter/unionpay"
	"github.com/ymqzj/payment-gateway/pkg/payadapter/wechat"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := configs.Load(configs.GetEnv())

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建支付适配器
	wechatAdapter, err := wechat.NewAdapter(cfg)
	if err != nil {
		log.Fatalf("Failed to create wechat adapter: %v", err)
	}

	// 因 alipay.NewAdapter 未定义，需检查是否缺少包引
	if err != nil {
		log.Fatalf("Failed to create alipay adapter: %v", err)
	}
	alipayAdapter, err := alipay.NewAdapter(cfg)
	if err != nil {
		log.Fatalf("Failed to create alipay adapter: %v", err)
	}

	unionpayAdapter, err := unionpay.NewAdapter(cfg)
	if err != nil {
		log.Fatalf("Failed to create unionpay adapter: %v", err)
	}

	// 创建支付网关
	gateway := payment.NewPaymentGateway(
		wechatAdapter,
		alipayAdapter,
		unionpayAdapter,
	)

	// 创建HTTP处理器
	handler := v1.NewPaymentHandler(gateway)

	// 创建Gin路由
	router := gin.Default()

	// 设置中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 设置路由
	v1 := router.Group("/api/v1")
	{
		v1.POST("/pay", handler.Pay)
		v1.POST("/query", handler.Query)
		v1.POST("/refund", handler.Refund)
		v1.POST("/close", handler.Close)
		v1.GET("/channels", handler.GetChannels)
		v1.GET("/health", handler.Health)

		// 通知接口
		v1.POST("/notify/:channel", handler.HandleNotify)
	}

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Server.Port),
		Handler: router,
	}

	// 启动服务器
	go func() {
		log.Printf("🚀 服务器启动在端口 %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🔄 服务器关闭中...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("服务器关闭失败:", err)
	}

	log.Println("✅ 服务器已关闭")
}
