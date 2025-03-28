package http

import (
	"U_PAY/dto"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

/* type Data struct {
	trade_id        string
	order_id        string
	amount          decimal.Decimal
	actual_amount   decimal.Decimal
	token           string
	expiration_time int64
	payment_url     string
}

// 定义返回的结构体
type Response struct {
	statusCode int
	message    string
	data       Data
	requestId  string
} */

const ( // 定义常量
	CnyMinimumPaymentAmount  = 0.01 // cny最低支付金额
	UsdtMinimumPaymentAmount = 0.01 // usdt最低支付金额
	UsdtAmountPerIncrement   = 0.01 // usdt每次递增金额
	IncrementalMaximumNumber = 100  // 最大递增次数
)

func Init() {

	// 调用自定义的日志记录器
	// log.Logger.Info("我是在main.go中打印的")

	// 开始路由测试
	r := gin.Default()

	// 添加限流中间件
	r.Use(RateLimitMiddleware())

	// 配置 CORS 中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},                      // 允许的源
		AllowMethods:     []string{"GET", "POST"},            // 允许的方法
		AllowHeaders:     []string{"Origin", "Content-Type"}, // 允许的头
		ExposeHeaders:    []string{"Content-Length"},         // 可见的头
		AllowCredentials: true,                               // 允许携带凭据
		MaxAge:           10 * time.Minute,                   // 缓存时间
	}))

	// 修改静态文件和模板的配置
	r.Static("/static/js", "./static/js")   // 将 /static 路径映射到 ./static 目录
	r.Static("/static/css", "./static/css") // 将 /static 路径映射到 ./static 目录
	r.Static("/static/img", "./static/img") // 将 /static 路径映射到 ./static 目录
	r.LoadHTMLGlob("static/*.html")         // 直接使用 static 目录，不需要 ./

	// 定义根路径返回字符图案

	r.GET("/", func(c *gin.Context) {

		// 将字符图案拼接为一个字符串
		response := ""
		for _, line := range dto.Pattern {
			response += line + "\n"
		}

		// 返回字符图案
		c.String(http.StatusOK, response)
	})

	// 定义订单路由组
	r_api := r.Group("/api", AuthMiddleware())

	r_api.POST("/create_order", CreateTransaction)

	// 定义支付路由组
	r_pay := r.Group("/pay")
	// 返回支付页面【支付页面是静态页面，所以需要返回html文件】
	r_pay.GET("/checkout-counter/:trade_id", CheckoutCounter)

	// 检查订单状态
	r_pay.GET("/check-status/:trade_id", CheckOrderStatus)

	// 启动服务器并实现优雅关闭
	runStopHTTP(r)
}

// 获取可用的钱包地址和可用的USDT金额
/* func getAvailableWalletAddress(walletAddress []sdb.WalletAddress, amount_usdt decimal.Decimal) (string, decimal.Decimal) {

	Wallet, Amount := wallet_amount(walletAddress, amount_usdt)

	// 如果没有找到可用的钱包，尝试增加金额重试
	if Wallet == "" {
		maxRetries := 100
		increment := decimal.NewFromFloat(0.01)

		for i := 0; i < maxRetries; i++ {
			amount_usdt = amount_usdt.Add(increment)
			if Wallet, Amount = wallet_amount(walletAddress, amount_usdt); Wallet != "" {
				log.Logger.Info(fmt.Sprintf("第%d次重试，金额：%s", i+1, amount_usdt.String()))
				break
			}
		}
		return Wallet, Amount
	}

	return Wallet, Amount
}

func wallet_amount(walletAddress []sdb.WalletAddress, amount_usdt decimal.Decimal) (string, decimal.Decimal) {
	for _, v := range walletAddress {
		v_token := v.Token
		address_amount := fmt.Sprintf("%s_%s", v.Token, amount_usdt.String())

		// 将address_amount作为键，去Redis里面查询
		cx := context.Background()
		_, err := rdb.RDB.Get(cx, address_amount).Result()

		if err != nil {
			return v_token, amount_usdt
		}
	}
	return "", amount_usdt
} */
