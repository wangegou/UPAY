package config

import (
	"log"

	"github.com/spf13/viper"
)

func Init() {
	viper.SetConfigName(".env") // 配置文件名（不带扩展名）
	viper.SetConfigType("env")  // 配置文件类型
	viper.AddConfigPath(".")    // 配置文件路径

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("无法读取配置文件: %v", err)
	}
}

// 获取本程序的名称
func GetAppName() string {
	return viper.GetString("app_name") // 返回本程序的名称
}

// 获取收银台的网址
func GetAppUri() string {
	return viper.GetString("app_uri") // 返回收银台的网址
}

// 获取API认证Token
func GetApiAuthToken() string {
	return viper.GetString("api_auth_token") // 返回API认证Token
}

// 获取http服务监听端口
func GetHttpListen() string {
	return viper.GetString("http_listen") // 返回http服务监听端口
}

// 获取静态资源文件目录

func GetStaticDir() string {
	return viper.GetString("static_path") // 返回静态资源文件目录
}

// 获取日志最大大小
func GetLogMaxSize() int {
	return viper.GetInt("log_max_size") // 返回日志最大大小
}

// 获取日志最大保存天数
func GetLogMaxAge() int {
	return viper.GetInt("log_max_age") // 返回日志最大保存天数
}

// 获取日志最大备份数
func GetLogMaxBackups() int {
	return viper.GetInt("log_max_backups") // 返回日志最大备份数
}

// 获取redis配置

// 获取redis 主机IP配置
func GetRedisHost() string {
	return viper.GetString("redis_host") // 返回redis配置
}

// 获取redis 端口配置
func GetRedisPort() int {
	return viper.GetInt("redis_port") // 返回redis端口
}

// 获取redis 密码配置
func GetRedisPasswd() string {
	return viper.GetString("redis_passwd") // 返回redis密码
}

// 获取redis 数据库编号配置
func GetRedisDb() int {
	return viper.GetInt("redis_db") // 返回redis数据库编号
}

// 获取机器人Apitoken
func GetTgBotToken() string {
	return viper.GetString("tg_bot_token") // 返回机器人Apitoken
}

// 获取管理员userid
func GetTgManage() string {
	return viper.GetString("tg_manage") // 返回管理员userid
}

// 获取订单过期时间
func GetOrderExpirationTime() int {
	time := viper.GetInt("order_expiration_time") // 返回订单过期时间
	if time <= 0 {
		return 10 // 如果订单过期时间为0，则返回10
	}
	return time
}

// 获取强制汇率
func GetForcedUsdtRate() float64 {
	rate := viper.GetFloat64("forced_usdt_rate") // 返回强制汇率
	if rate <= 0 {
		return 6.4 // 如果强制汇率为0，则返回1
	}
	return rate
}

// 获取钱包地址 和使用状态

func GetWalletAddress() string {
	return viper.GetString("wallet_address") // 返回钱包地址
}

// 获取钱包状态
func GetWalletStatus() int {
	return viper.GetInt("wallet_status") // 返回钱包状态
}

// 获取客服联系方式
func GetCustomerServiceContact() string {
	return viper.GetString("customer_service_contact") // 返回客服联系方式
}

// 获取Bark通知服务
func GetBarkKey() string {
	return viper.GetString("bark_key") // 返回Bark通知服务
}
