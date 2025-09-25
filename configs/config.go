package configs

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

// Config 全局配置结构体
type Config struct {
	Wechat   WechatConfig   `mapstructure:"wechat"`
	Alipay   AlipayConfig   `mapstructure:"alipay"`
	UnionPay UnionPayConfig `mapstructure:"unionpay"`
	Server   ServerConfig   `mapstructure:"server"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Database DatabaseConfig `mapstructure:"database"`
}

// WechatConfig 微信支付配置
type WechatConfig struct {
	AppID        string `mapstructure:"app_id"`
	MchID        string `mapstructure:"mch_id"`
	APIKey       string `mapstructure:"api_key"`
	CertPath     string `mapstructure:"cert_path"`
	KeyPath      string `mapstructure:"key_path"`
	CertSerialNo string `mapstructure:"cert_serial_no"`
	APIV3Key     string `mapstructure:"api_v3_key"`
	NotifyURL    string `mapstructure:"notify_url"`
}

// AlipayConfig 支付宝配置
type AlipayConfig struct {
	AppID           string `mapstructure:"app_id"`
	PrivateKey      string `mapstructure:"private_key"`
	AlipayPublicKey string `mapstructure:"alipay_public_key"`
	GatewayURL      string `mapstructure:"gateway_url"`
	Charset         string `mapstructure:"charset"`
	SignType        string `mapstructure:"sign_type"`
	NotifyURL       string `mapstructure:"notify_url"`
	ReturnURL       string `mapstructure:"return_url"`
}

// UnionPayConfig 银联配置
type UnionPayConfig struct {
	MerID          string `mapstructure:"mer_id"`
	AppId          string `mapstructure:"app_id"`
	CertPath       string `mapstructure:"cert_path"`
	CertPwd        string `mapstructure:"cert_pwd"`
	PrivateKeyPath string `mapstructure:"private_key_path"`
	PublicKeyPath  string `mapstructure:"public_key_path"`
	Gateway        string `mapstructure:"gateway"`
	BackURL        string `mapstructure:"back_url"`
	FrontURL       string `mapstructure:"front_url"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	Charset      string `mapstructure:"charset"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// Load 加载配置
func Load(env string) *Config {
	config := &Config{}

	// 设置配置文件路径
	viper.SetConfigName(env)
	viper.SetConfigType("yaml")
	
	// 添加多个可能的配置文件路径
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")
	viper.AddConfigPath(".")

	// 读取环境变量
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// 解析配置
	if err := viper.Unmarshal(config); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	return config
}

// GetEnv 获取当前环境
func GetEnv() string {
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}
	return env
}