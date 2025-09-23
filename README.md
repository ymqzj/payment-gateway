# 支付网关系统 (Payment Gateway)

一个支持微信、支付宝、银联的Go语言统一支付网关系统，采用适配器模式设计，支持多种支付场景，易于扩展和维护。

## 🌟 特性

- ✅ **多支付渠道支持**: 微信、支付宝、银联一键接入
- ✅ **统一API接口**: 统一的支付、查询、退款、关闭接口
- ✅ **多种支付场景**: APP支付、H5支付、公众号支付、扫码支付、PC支付
- ✅ **异步通知处理**: 统一处理各渠道的支付通知
- ✅ **配置驱动**: 基于YAML的配置管理，支持多环境
- ✅ **高可扩展性**: 适配器模式，新增渠道无需修改核心代码
- ✅ **完整测试**: 单元测试、集成测试全覆盖
- ✅ **生产就绪**: 包含错误处理、日志、监控、Docker化部署

## 📁 项目结构

```
payment-gateway/
├── cmd/                    # 应用程序入口
│   └── server/            # HTTP服务器
├── internal/              # 内部私有模块
│   └── payment/           # 核心支付网关
├── pkg/                   # 可复用公共模块
│   └── payadapter/        # 支付渠道适配器
│       ├── wechat/        # 微信支付
│       ├── alipay/        # 支付宝
│       └── unionpay/      # 银联支付
├── api/                   # HTTP API接口
│   └── v1/               # API版本1
├── configs/               # 配置文件
├── test/                  # 测试文件
├── scripts/               # 部署脚本
├── logs/                  # 日志文件
└── docs/                  # 文档
```

## 🚀 快速开始

### 1. 环境要求

- Go 1.21+
- Docker (可选)
- 各支付渠道的开发配置

### 2. 安装依赖

```bash
# 克隆项目
git clone https://github.com/ymqzj/payment-gateway.git
cd payment-gateway

# 安装依赖
make deps

# 初始化项目
make init
```

### 3. 配置项目

复制配置文件并修改为你的配置：

```bash
cp configs/dev.yaml configs/local.yaml
# 编辑 configs/local.yaml 填入你的配置
```

### 4. 运行项目

```bash
# 开发模式
make dev

# 生产模式
make prod

# 构建并运行
make build
./bin/payment-gateway
```

## 🛠️ 配置说明

### 微信支付配置

```yaml
wechat:
  app_id: "你的微信应用ID"
  mch_id: "你的商户号"
  api_key: "你的API密钥"
  cert_path: "./certs/wechat/apiclient_cert.pem"
  key_path: "./certs/wechat/apiclient_key.pem"
  cert_serial_no: "证书序列号"
  api_v3_key: "APIv3密钥"
  notify_url: "https://yourdomain.com/notify/wechat"
```

### 支付宝配置

```yaml
alipay:
  app_id: "你的支付宝应用ID"
  private_key: |
    -----BEGIN RSA PRIVATE KEY-----
    你的私钥内容
    -----END RSA PRIVATE KEY-----
  alipay_public_key: |
    -----BEGIN PUBLIC KEY-----
    支付宝公钥内容
    -----END PUBLIC KEY-----
  gateway_url: "https://openapi.alipay.com/gateway.do"
  charset: "UTF-8"
  sign_type: "RSA2"
  notify_url: "https://yourdomain.com/notify/alipay"
  return_url: "https://yourdomain.com/return/alipay"
```

### 银联配置

```yaml
unionpay:
  mer_id: "你的商户号"
  cert_path: "./certs/unionpay/acp_prod_sign.pfx"
  cert_pwd: "证书密码"
  private_key_path: "./certs/unionpay/private.key"
  public_key_path: "./certs/unionpay/public.key"
  gateway_url: "https://gateway.95516.com/gateway/api/"
  back_url: "https://yourdomain.com/notify/unionpay"
  front_url: "https://yourdomain.com/return/unionpay"
```

## 📋 API接口

### 统一支付接口

```http
POST /api/v1/pay
Content-Type: application/json

{
  "channel": "wechat",
  "out_trade_no": "ORDER_20240101120000",
  "total_amount": 0.01,
  "subject": "商品描述",
  "scene": "app",
  "notify_url": "https://yourdomain.com/notify"
}
```

### 订单查询接口

```http
POST /api/v1/query
Content-Type: application/json

{
  "channel": "wechat",
  "out_trade_no": "ORDER_20240101120000"
}
```

### 退款接口

```http
POST /api/v1/refund
Content-Type: application/json

{
  "channel": "wechat",
  "out_trade_no": "ORDER_20240101120000",
  "out_refund_no": "REFUND_20240101120000",
  "refund_amount": 0.01,
  "total_amount": 0.01,
  "refund_reason": "退款原因"
}
```

### 关闭订单接口

```http
POST /api/v1/close
Content-Type: application/json

{
  "channel": "wechat",
  "out_trade_no": "ORDER_20240101120000"
}
```

### 获取支持渠道

```http
GET /api/v1/channels
```

## 🧪 测试

```bash
# 运行所有测试
make test

# 运行单元测试
make test-unit

# 生成测试覆盖率报告
make test-coverage

# 运行特定测试
go test -v ./test/wechat_test.go
```

## 🐳 Docker部署

### 构建镜像

```bash
make docker-build
```

### 运行容器

```bash
make docker-run
```

### Docker Compose

```yaml
version: '3.8'
services:
  payment-gateway:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENV=prod
    volumes:
      - ./configs:/app/configs
      - ./certs:/app/certs
      - ./logs:/app/logs
    restart: unless-stopped
```

## 🔧 开发指南

### 添加新支付渠道

1. 在 `pkg/payadapter/` 下创建新目录
2. 实现 `payment.Adapter` 接口
3. 注册到支付网关

示例：

```go
// 新渠道适配器
package newchannel

import (
    "context"
    "payment-gateway/internal/payment"
)

type Adapter struct {
    config *Config
}

func (a *Adapter) Pay(ctx context.Context, req *payment.UnifiedPayRequest) (*payment.UnifiedPayResponse, error) {
    // 实现支付逻辑
}

func (a *Adapter) GetChannel() payment.ChannelType {
    return payment.ChannelType("newchannel")
}
```

### 自定义通知处理器

```go
// 创建通知处理器
type MyNotifyProcessor struct{}

func (p *MyNotifyProcessor) Process(ctx context.Context, result *payment.NotifyResult) error {
    // 处理通知逻辑
    return nil
}

// 注册处理器
notifyManager.RegisterProcessor("my_processor", &MyNotifyProcessor{})
```

## 📊 监控和日志

### 日志配置

日志会自动写入 `logs/app.log`，支持以下级别：
- DEBUG
- INFO
- WARN
- ERROR

### 健康检查

```http
GET /api/v1/health
```

### 指标监控

项目预留了监控指标接口，可以集成Prometheus等监控系统。

## 🚨 错误处理

### 错误码说明

- `0`: 成功
- `1001-1999`: 客户端错误
- `2001-2999`: 渠道特定错误
- `3001-3999`: 签名错误
- `4001-4999`: 通知错误
- `5000-5999`: 服务器错误

### 常见错误处理

```go
// 错误处理示例
resp, err := gateway.Pay(ctx, req)
if err != nil {
    if payment.IsBusinessError(err) {
        // 业务错误，可以重试
    } else if payment.IsRetryable(err) {
        // 可重试错误
    } else {
        // 不可恢复错误
    }
}
```

## 📞 支持

如有问题，请提交 [GitHub Issue](https://github.com/your-repo/issues) 或联系维护者。

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

- [微信支付官方文档](https://pay.weixin.qq.com/wiki/)
- [支付宝开放平台](https://opendocs.alipay.com/)
- [中国银联开放平台](https://open.unionpay.com/)
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [Viper配置管理](https://github.com/spf13/viper)