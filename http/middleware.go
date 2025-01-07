package http

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"U_PAY/config"
	"U_PAY/dto"
	"U_PAY/log"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// 中间件用来验证数据是否完整，最后把请求参数存储在上下文中

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Logger.Info("进入中间件")
		// 获取请求体
		var requestParams dto.RequestParams
		if err := c.ShouldBindJSON(&requestParams); err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Logger.Info("请求体参数绑定失败")
			c.Abort()
			return
		}
		log.Logger.Info("请求体参数绑定成功")
		// 对请求参数进行验证
		validate := validator.New() //创建一个验证器实例：
		if err := validate.Struct(requestParams); err != nil {
			//如果验证错误，则返回错误信息，并终止请求

			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Logger.Info("请求体参数验证失败", zap.String("error", err.Error()))
			c.Abort()
			return
		}
		log.Logger.Info("请求体参数验证成功")
		// 上面已经获取到了请求参数，我们也按照规则进行拼接字符串进行md5加密计算和传入的Signature值进行对比
		// 使用 fmt.Sprintf 生成查询字符串(拼接了api_auth_token)

		// 排序拼接

		// 签名生成：按规则拼接
		/* params := map[string]string{
			"amount":       fmt.Sprintf("%.2f", requestParams.Amount), // 保留两位小数
			"notify_url":   requestParams.NotifyURL,
			"order_id":     requestParams.OrderID,
			"redirect_url": requestParams.RedirectURL,
		} */

		params := []string{
			fmt.Sprintf("amount=%x", requestParams.Amount),
			fmt.Sprintf("notify_url=%s", requestParams.NotifyURL),
			fmt.Sprintf("order_id=%s", requestParams.OrderID),
			fmt.Sprintf("redirect_url=%s", requestParams.RedirectURL),
		}
		// 打印拼接的参数
		log.Logger.Info("拼接的参数", zap.Any("params", params))

		// 打印原字符串
		/*
			log.Logger.Info("金额:", zap.Float64("amount", requestParams.Amount))
			log.Logger.Info("通知URL:", zap.String("notify_url", requestParams.NotifyURL))
			log.Logger.Info("订单ID:", zap.String("order_id", requestParams.OrderID))
			log.Logger.Info("重定向URL:", zap.String("redirect_url", requestParams.RedirectURL)) */

		/* // 排序拼接
		var keys []string
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys) // 按键名排序
		*/
		// 排序参数
		sort.Strings(params)

		// 使用 strings.Join 连接排序后的参数
		signatureString := strings.Join(params, "&") + config.GetApiAuthToken()
		/* var queryString string
		for _, key := range keys {
			value := params[key]
			if value != "" { // 跳过空值
				queryString += fmt.Sprintf("%s=%s&", key, value)
			}
		} */
		// queryString = strings.TrimRight(queryString, "&") + config.GetApiAuthToken()

		log.Logger.Info("拼接的查询字符串", zap.String("queryString", signatureString))

		/* 		queryString := fmt.Sprintf("amount=%f&notify_url=%s&order_id=%s&redirect_url=%s%s",
		requestParams.Amount, requestParams.NotifyURL, requestParams.OrderID, requestParams.RedirectURL, config.GetApiAuthToken())
		*/
		// 打印一下传入的签名
		log.Logger.Info("传入的签名", zap.String("signature", requestParams.Signature))
		// 对拼接的字符串进行md5加密，并验证如果传入的签名和计算的签名一致，则继续执行下一个中间件或者处理函数

		Signature := fmt.Sprintf("%x", md5.Sum([]byte(signatureString)))
		log.Logger.Info("计算的签名", zap.String("Signature", Signature))
		if requestParams.Signature != Signature {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "签名验证失败"})
			log.Logger.Info("签名验证失败")
			c.Abort()
			return
		}
		log.Logger.Info("签名验证成功")
		// 继续执行下一个中间件或者处理函数

		// 将请求参数存储在上下文中
		c.Set("requestParams", requestParams)
		c.Next()
	}
}
