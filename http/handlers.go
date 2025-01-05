package http

import (
	"U_PAY/config"
	"U_PAY/db/rdb"
	"U_PAY/db/sdb"
	"U_PAY/dto"
	"U_PAY/log"
	"U_PAY/mq"
	"U_PAY/notification"
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 启动服务器并实现优雅关闭
func runStopHTTP(r *gin.Engine) {
	// 创建自定义的 http.Server
	srv := &http.Server{
		Addr:    config.GetHttpListen(),
		Handler: r,
	}

	// 在新的 goroutine 中启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Logger.Error("listen: " + err.Error())
		}
	}()
	log.Logger.Info("服务器已启动，监听端口 :", zap.String("port", config.GetHttpListen()))

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Logger.Info("正在关闭服务器...")

	// 设置 5 秒的超时时间来优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Logger.Error("服务器关闭时发生错误: " + err.Error())
	}

	log.Logger.Info("服务器已成功关闭")
}

var gCreateTransactionLock sync.Mutex // 创建交易的锁
func CreateTransaction(c *gin.Context) {
	// 中间件验证成功后，这里就要开始创建订单的逻辑处理了；

	// 从上下文中获取请求参数
	requestParams, exists := c.Get("requestParams") //返回值的类型any 和 interface{} 是等价的，any 是 Go 1.18 引入的别名
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "中间件传递的请求参数未找到"})
		return
	}

	/* 	c.JSON(200, gin.H{
		 使用 requestParams.(RequestParams).OrderID 是因为 requestParams 是一个接口类型，您需要进行类型断言以访问其字段。
		如果 requestParams 已经是 RequestParams 类型，则可以直接使用 requestParams.OrderID，但在从上下文中获取时，通常需要进行类型断言。
		"交易记录": requestParams.(RequestParams).OrderID,
	}) */

	/* a := tron.GetTransactions("TXFchkPBXBkHGifcUdQdErhBzP5T6kRfyv")
	// 打印 代币名称
	fmt.Println(a.TokenAbbr)
	// 打印 交易哈希值
	fmt.Println(a.TransactionID)
	// 打印 金额
	fmt.Println(a.Quant)
	// 打印 付款地址
	fmt.Println(a.FromAddress)
	// 打印 收款地址
	fmt.Println(a.ToAddress)
	// 打印 交易结果
	fmt.Println(a.FinalResult) */

	// 获取可用的钱包地址和可用的USDT金额
	// 加锁
	gCreateTransactionLock.Lock()
	// 解锁
	defer gCreateTransactionLock.Unlock()

	// 将传入的金额进行十进制转换，并保留2位数
	amount := decimal.NewFromFloat(requestParams.(dto.RequestParams).Amount).Round(2)

	// 如果金额小于最低支付金额，则返回错误

	/* 	比较两个 Decimal 值：

	比较结果有以下几种：

	-1：表示第一个值小于第二个值
	0：表示两个值相等
	1：表示第一个值大于第二个值 */
	if amount.Cmp(decimal.NewFromFloat(CnyMinimumPaymentAmount)) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CNY金额小于最低支付金额0.01"})
		return
	}

	// 汇率  ；获取配置文件里的汇率设置，因为获取的是float64类型，所以需要进行十进制转换
	rate := decimal.NewFromFloat(config.GetForcedUsdtRate())
	// 计算换算为USDT的金额;保留2位小数
	//Div是除法，Round是四舍五入
	amount_usdt := amount.Div(rate).Round(2)

	// 如果换算后的金额小于最低支付金额，则返回错误
	if amount_usdt.Cmp(decimal.NewFromFloat(UsdtMinimumPaymentAmount)) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "USDT金额小于最低支付金额0.01"})
		return
	}

	// amount := requestParams.(RequestParams).Amount
	// 将金额和钱包地址进行拼接
	// fmt.Println(amount)
	// 先创建一个切片钱包地址的实例对象,用来存储查询到的钱包地址记录
	walletAddress := []sdb.WalletAddress{}
	err := sdb.SDB.Find(&walletAddress, "status=?", sdb.TokenStatusEnable).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "查询钱包地址失败"})
		return
	}

	// 判断是否获取到钱包地址，因为上面的查询用的是find，会把所有符合要求的记录存储在切片walletAddress的结构体实例中，所以通过判断切片的长度看否有数据
	if len(walletAddress) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有可用的钱包地址"})
		return
	}

	// 下面要判断一下当然的钱包地址和金额是否被占用，因为钱包地址金额金额必须绑定为唯一的的值，波场是通过不断检查钱包地址的最新的一笔入账USDT的金额来判断是否入账成功

	// 拼接钱包地址和和金额，作为Redis的键去查询，当有这个键时，说明这个钱包地址和金额已经被占用，需要更换钱包地址或者金额；所以这里应该是一个迭代循环

	var Wallet string
	var Amount decimal.Decimal
OuterLoop:
	for i := 0; i < 100; i++ {

		for _, v := range walletAddress {
			// 因为walletAddress是一个切片，这里循环就省略了下标用_代替，因为切片里的每个元素都是一个结构体（因为前面的find查询把返回的每条记录都存储在结构体实例中）
			// 获取钱包地址
			v_token := v.Token
			// 拼接钱包地址和和金额，作为Redis的键去查询
			address_amount := fmt.Sprintf("%s_%s", v.Token, amount_usdt.String())
			// 将address_amount作为键，去Redis里面查询
			cx := context.Background()
			_, err := rdb.RDB.Get(cx, address_amount).Result()
			// 如果获取到这个键。err会等于nil，说明这个钱包地址和金额已经被占用，需要更换钱包地址或者金额
			if err != nil {
				// 如果不等于nil，说明这个钱包地址和金额没有被占用，可以继续使用
				Wallet = v_token
				Amount = amount_usdt
				break OuterLoop
			}

		}

		// USDT金额递增
		amount_usdt = amount_usdt.Add(decimal.NewFromFloat(UsdtAmountPerIncrement))

	}
	log.Logger.Info(fmt.Sprintf("已经找到了可用的钱包地址和金额；钱包地址：%s,金额：%v", Wallet, amount_usdt))
	// fmt.Println(Wallet, Amount)
	// 获取可用的钱包地址和可用的USDT金额
	// Wallet, Amount := getAvailableWalletAddress(walletAddress, amount_usdt)

	// 准备订单所需的数据，创建订单记录
	order := sdb.Orders{
		TradeId:            GenerateUniqueTradeID(),
		OrderId:            requestParams.(dto.RequestParams).OrderID,
		BlockTransactionId: "NULL",
		Amount:             requestParams.(dto.RequestParams).Amount,
		ActualAmount:       amount_usdt.InexactFloat64(),
		Token:              Wallet,
		Status:             sdb.StatusWaitPay,
		NotifyUrl:          requestParams.(dto.RequestParams).NotifyURL,
		RedirectUrl:        requestParams.(dto.RequestParams).RedirectURL,
		CallbackNum:        0,
		CallBackConfirm:    0,
		StartTime:          carbon.Now().TimestampMilli(),
		ExpirationTime:     carbon.Now().AddMinutes(config.GetOrderExpirationTime()).TimestampMilli(),
	}
	// 用事务创建订单记录|保证创建订单记录的原子性（要么全部成功，要么全部失败）
	err = sdb.SDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			log.Logger.Info("创建订单记录失败：" + err.Error())
			return err
		}
		return nil
	})
	if err != nil {
		log.Logger.Info("创建订单记录失败：" + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建订单记录失败"})
		return
	}
	log.Logger.Info("创建订单记录成功")

	// 测试发送Bark通知
	notification.Start(order)

	// 获取配置文件中的订单过期时间并转换为分钟
	/* orderExpirationMinutes := viper.GetInt("order_expiration_time")
	if orderExpirationMinutes <= 0 {
		orderExpirationMinutes = 10 // 默认10分钟
	} */
	// 转换为 Duration（时间单位）
	expirationDuration := time.Duration(config.GetOrderExpirationTime()) * time.Minute
	log.Logger.Info(fmt.Sprintf("订单过期时间：%d分钟", config.GetOrderExpirationTime()))

	// 锁定钱包和金额
	address_amount := fmt.Sprintf("%s_%s", Wallet, Amount.String())
	cx := context.Background()
	err = rdb.RDB.Set(cx, address_amount, GenerateUniqueTradeID(), expirationDuration).Err()
	if err != nil {
		log.Logger.Error("锁定钱包和金额失败：" + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "锁定钱包和金额失败"})
		return
	}

	/* // 获取当前时间戳（毫秒）
	NowTime := carbon.Now().TimestampMilli()
	fmt.Println(NowTime) */

	// 创建任务队列；设置延迟执行时间，传入的order.TradeId作为payload[任务载荷]

	mq.TaskOrderExpiration(order.TradeId, expirationDuration)
	// 计算10分钟后的时间戳（毫秒）
	// ExpirationTime := carbon.Now().AddMinutes(config.GetOrderExpirationTime()).TimestampMilli()

	// 准备返回订单信息的数据
	orderInfo := dto.Response{
		StatusCode: 200,
		Message:    "success",
		Data: dto.Data{
			TradeID:        order.TradeId,
			OrderID:        order.OrderId,
			Amount:         decimal.NewFromFloat(order.Amount),
			ActualAmount:   decimal.NewFromFloat(order.ActualAmount),
			Token:          order.Token,
			ExpirationTime: order.ExpirationTime,
			PaymentURL:     fmt.Sprintf("%s%s%s", config.GetAppUri(), "/pay/checkout-counter/", order.TradeId),
		},
		// RequestID: "1234567890",
	}

	// 返回订单信息
	c.JSON(200, gin.H{
		"status_code": orderInfo.StatusCode,
		"message":     orderInfo.Message,
		"data":        gin.H{"trade_id": orderInfo.Data.TradeID, "order_id": orderInfo.Data.OrderID, "amount": orderInfo.Data.Amount, "actual_amount": orderInfo.Data.ActualAmount, "token": orderInfo.Data.Token, "expiration_time": orderInfo.Data.ExpirationTime, "payment_url": orderInfo.Data.PaymentURL},
		"requestId":   orderInfo.RequestID,
	})

}

// GenerateUniqueTradeID 生成唯一的交易ID（年月日时分秒纳秒 + 5位随机数）
func GenerateUniqueTradeID() string {
	now := time.Now()
	randomNum, _ := generateCryptoRandom(100000) // 使用安全随机数
	/* if err != nil {
		return ""
	}
	*/
	return fmt.Sprintf("%s%d%04d",
		now.Format("20060102150405"),
		now.Nanosecond()/1000, // 纳秒转换为微秒
		randomNum,
	)
}

// 生成加密安全的随机数
func generateCryptoRandom(max int64) (int64, error) {
	// 创建一个新的 big.Int，设置最大值
	maxBig := big.NewInt(max)

	// 使用 crypto/rand 生成随机数
	num, err := rand.Int(rand.Reader, maxBig) // 正确传递两个参数
	if err != nil {
		return 0, err
	}

	return num.Int64(), nil
}

func CheckoutCounter(c *gin.Context) {

	// 获取请求参数
	trade_id := c.Param("trade_id")

	// 获取订单信息
	order := sdb.Orders{}
	err := sdb.SDB.Find(&order, "trade_id=? and status=?", trade_id, sdb.StatusWaitPay).Error
	if err != nil {
		c.JSON(500, gin.H{"error": "获取订单信息失败"})
		return
	}

	// expirationMinutes := viper.GetInt("order_expiration_time")
	// 组装一下模版所需的参数
	viewModel := dto.PaymentViewModel{
		TradeId:                order.TradeId,
		ActualAmount:           order.ActualAmount,
		Token:                  order.Token,
		ExpirationTime:         order.ExpirationTime,
		RedirectUrl:            order.RedirectUrl,
		AppName:                config.GetAppName(),
		CustomerServiceContact: config.GetCustomerServiceContact(),
	}

	// 返回支付页面
	c.HTML(http.StatusOK, "index.html", viewModel)

}

// 检查订单状态
func CheckOrderStatus(c *gin.Context) {

	// 依据传入的路径参数【交易ID】，查询订单状态
	trade_id := c.Param("trade_id")

	// 查询订单状态
	order := sdb.Orders{}
	err := sdb.SDB.Find(&order, "trade_id=?", trade_id).Error
	if err != nil {
		c.JSON(500, gin.H{"message": "获取订单信息失败"})
		return
	}

	// 返回订单状态
	c.JSON(200, gin.H{"data": gin.H{"status": order.Status}})

}
