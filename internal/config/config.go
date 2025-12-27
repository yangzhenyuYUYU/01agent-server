package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server         ServerConfig         `mapstructure:"server"`
	Database       DatabaseConfig       `mapstructure:"database"`
	Redis          RedisConfig          `mapstructure:"redis"`
	JWT            JWTConfig            `mapstructure:"jwt"`
	Log            LogConfig            `mapstructure:"log"`
	CORS           CORSConfig           `mapstructure:"cors"`
	Doubao         DoubaoConfig         `mapstructure:"doubao"`
	WxPay          WxPayConfig          `mapstructure:"wxpay"`
	Alipay         AlipayConfig         `mapstructure:"alipay"`
	Wechat         WechatConfig         `mapstructure:"wechat"`
	WechatGzh      WechatGzhConfig      `mapstructure:"wechatGzh"`
	WechatPlatform WechatPlatformConfig `mapstructure:"wechatPlatform"`
	SMS            SMSConfig            `mapstructure:"sms"`
	OSS            OSSConfig            `mapstructure:"oss"`
	Email          EmailConfig          `mapstructure:"email"`
	BP             BPConfig             `mapstructure:"bp"`
	Credits        CreditsConfig        `mapstructure:"credits"`
	VerifyCode     VerifyCodeConfig     `mapstructure:"verifyCode"`
	Themes         map[string]string    `mapstructure:"themes"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Type     string         `mapstructure:"type"` // "mysql" or "postgres"
	MySQL    MySQLConfig    `mapstructure:"mysql"`
	Postgres PostgresConfig `mapstructure:"postgres"`
}

type MySQLConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	Charset         string        `mapstructure:"charset"`
	ParseTime       bool          `mapstructure:"parseTime"`
	Loc             string        `mapstructure:"loc"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
}

type PostgresConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	TimeZone        string        `mapstructure:"timeZone"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
}

type RedisConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	PoolSize   int    `mapstructure:"poolSize"`
	MaxRetries int    `mapstructure:"maxRetries"`
}

type JWTConfig struct {
	Secret string        `mapstructure:"secret"`
	Expire time.Duration `mapstructure:"expire"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"maxSize"`
	MaxBackups int    `mapstructure:"maxBackups"`
	MaxAge     int    `mapstructure:"maxAge"`
	Compress   bool   `mapstructure:"compress"`
}

type CORSConfig struct {
	AllowOrigins     []string      `mapstructure:"allowOrigins"`
	AllowMethods     []string      `mapstructure:"allowMethods"`
	AllowHeaders     []string      `mapstructure:"allowHeaders"`
	ExposeHeaders    []string      `mapstructure:"exposeHeaders"`
	AllowCredentials bool          `mapstructure:"allowCredentials"`
	MaxAge           time.Duration `mapstructure:"maxAge"`
}

// 豆包AI配置
type DoubaoConfig struct {
	APIKey  string             `mapstructure:"apiKey"`
	BaseURL string             `mapstructure:"baseURL"`
	Timeout time.Duration      `mapstructure:"timeout"`
	Models  DoubaoModelsConfig `mapstructure:"models"`
}

type DoubaoModelsConfig struct {
	DefaultLLM        string `mapstructure:"defaultLLM"`
	DefaultMultimodal string `mapstructure:"defaultMultimodal"`
	DefaultSTT        string `mapstructure:"defaultSTT"`
	DefaultTTS        string `mapstructure:"defaultTTS"`
	DefaultTTSVoice   string `mapstructure:"defaultTTSVoice"`
}

// 微信支付配置
type WxPayConfig struct {
	AppID          string `mapstructure:"appID"`
	MchID          string `mapstructure:"mchID"`
	SerialNo       string `mapstructure:"serialNo"`
	PrivateKeyPath string `mapstructure:"privateKeyPath"`
	APIV3Key       string `mapstructure:"apiV3Key"`
	NotifyURL      string `mapstructure:"notifyURL"`
}

// 支付宝配置
type AlipayConfig struct {
	AppID      string `mapstructure:"appID"`
	PublicKey  string `mapstructure:"publicKey"`
	PrivateKey string `mapstructure:"privateKey"`
	NotifyURL  string `mapstructure:"notifyURL"`
}

// 微信小程序配置
type WechatConfig struct {
	AppID  string `mapstructure:"appID"`
	Secret string `mapstructure:"secret"`
}

// 微信公众号配置
type WechatGzhConfig struct {
	Token  string `mapstructure:"token"`
	AppID  string `mapstructure:"appID"`
	Secret string `mapstructure:"secret"`
}

// 微信三方平台配置
type WechatPlatformConfig struct {
	Token          string `mapstructure:"token"`
	EncodingAESKey string `mapstructure:"encodingAESKey"`
	AppID          string `mapstructure:"appID"`
	AppSecret      string `mapstructure:"appSecret"`
}

// 短信配置
type SMSConfig struct {
	SecretID   string `mapstructure:"secretID"`
	SecretKey  string `mapstructure:"secretKey"`
	SDKAppID   string `mapstructure:"sdkAppID"`
	SignName   string `mapstructure:"signName"`
	TemplateID string `mapstructure:"templateID"`
}

// OSS配置
type OSSConfig struct {
	AccessKeyID      string `mapstructure:"accessKeyID"`
	AccessKeySecret  string `mapstructure:"accessKeySecret"`
	InternalEndpoint string `mapstructure:"internalEndpoint"`
	Endpoint         string `mapstructure:"endpoint"`
	BucketName       string `mapstructure:"bucketName"`
}

// 邮件配置
type EmailConfig struct {
	Sender     string `mapstructure:"sender"`
	Password   string `mapstructure:"password"`
	SMTPServer string `mapstructure:"smtpServer"`
	SMTPPort   int    `mapstructure:"smtpPort"`
	SenderName string `mapstructure:"senderName"`
}

// BP文档配置
type BPConfig struct {
	DocumentPath      string `mapstructure:"documentPath"`
	EmailTemplatePath string `mapstructure:"emailTemplatePath"`
	WeixinQRPath      string `mapstructure:"weixinQRPath"`
	WebsiteURL        string `mapstructure:"websiteURL"`
}

// 积分配置
type CreditsConfig struct {
	Initial           int `mapstructure:"initial"`
	ArticleGeneration int `mapstructure:"articleGeneration"`
	ToCNY             int `mapstructure:"toCNY"`
}

// 验证码配置
type VerifyCodeConfig struct {
	Expire         int `mapstructure:"expire"`
	ResendInterval int `mapstructure:"resendInterval"`
}

var AppConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	AppConfig = &config
	return &config, nil
}
