package cron

// 设置定义任务检查数据库订单表中有未支付的订单，去请求tron的api查询是否支付成功，如果钱包和金额都正确，则将订单状态改为已支付

import (
	"U_PAY/config"
	"U_PAY/db/rdb"
	"U_PAY/db/sdb"
	"U_PAY/log"
	"U_PAY/notification"
	"U_PAY/tron"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"U_PAY/dto"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 在文件顶部定义全局HTTP客户端
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
	},
}

// 定义一个任务结构体 UsdtRateJob
type UsdtCheckJob struct{}

// 定义一个异步请求参数的结构体

/* type PaymentNotification struct {
	TradeID            string  `json:"trade_id"`
	OrderID            string  `json:"order_id"`
	Amount             float64 `json:"amount"`
	ActualAmount       float64 `json:"actual_amount"`
	Token              string  `json:"token"`
	BlockTransactionID string  `json:"block_transaction_id"`
	Signature          string  `json:"signature"`
	Status             int     `json:"status"`
} */

// 实现 cron.Job 接口的 Run 方法
func (j UsdtCheckJob) Run() {
	// 执行任务的具体逻辑

	// 查询订单表中状态为未支付的订单,可能存在多个订单，所以用切片来接受结果
	var order []sdb.Orders
	if err := sdb.SDB.Where("status = ?", sdb.StatusWaitPay).Find(&order).Error; err != nil {
		log.Logger.Info("订单查询失败", zap.Any("err", err))
		return
	}
	if len(order) == 0 {
		log.Logger.Info("没有未支付的订单")
		// 查询不到符合要求的订单，直接返回
		return

	}
	for _, v := range order {
		// 查询支付转账的情况，传入的参数是每个订单里面的钱包地址和查询的开始时间戳和结束时间戳
		// 返回的是一个结构体
		td := tron.GetTransactions(v.Token, v.StartTime, v.ExpirationTime)
		// 判断返回的结构体里面的金额是否等于订单的实际金额「这里需要判断的USDT的数量」

		if v.ActualAmount == td.Quant {
			// 使用 Transaction 简化事务处理
			err := sdb.SDB.Transaction(func(tx *gorm.DB) error {
				// 当金额相等的时候，则将订单状态改为已支付
				v.Status = sdb.StatusPaySuccess
				// 将订单的block_transaction_id设置为查询到的交易id「保存的是交易哈希值」
				v.BlockTransactionId = td.TransactionID
				// 先保存到数据库里面
				// sdb.SDB.Save(&v)
				// 事务回调函数内部始终使用 tx 参数进行操作，这是GORM事务的正确使用方式
				if err := tx.Save(&v).Error; err != nil {
					log.Logger.Info("更新数据库表失败", zap.Any("err", err))
					return err
				}

				// // 发送Bark通知|| 异步进程发送通知
				// go notification.Start(v)

				return nil
			})
			if err == nil {

				// 异步回调
				go j.processCallback(v)

			} else {
				log.Logger.Info("已经检查到了支付金额，但更新数据库表失败", zap.Any("err", err))
			}
			// if err != nil {
			// 	log.Logger.Info("已经检查到了支付金额，但更新数据库表失败", zap.Any("err", err))
			// }
		}
	}

}

func Start() {
	// 创建一个新的 Cron 调度器

	// 如果上一次任务还在运行，新的任务执行时间到了，则等待上一次任务完成后再执行
	c := cron.New(cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger)))

	// 每 5 秒执行一次 UsdtRateJob 任务
	_, err := c.AddJob("@every 5s", UsdtCheckJob{})
	if err != nil {
		log.Logger.Info("任务添加失败")
	}

	// 启动 Cron 调度器
	c.Start()

	// 保持主程序运行，确保任务执行
	select {}
}

// 发起异步 POST 请求
func sendAsyncPost(url string, notification dto.PaymentNotification_request) (string, error) {
	// 将结构体转换为 JSON 数据
	requestBody, err := json.Marshal(notification)
	if err != nil {
		fmt.Printf("JSON 序列化失败: %v\n", err)
		return "", err
	}

	// 创建 POST 请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	// client := &http.Client{Timeout: 10 * time.Second}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	// 读取响应
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		fmt.Println("发送成功，服务器返回 200 OK")

		// 读取服务器响应
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		if buf.String() == "ok" || buf.String() == "success" {
			fmt.Println("发送成功，服务器返回字符串 'ok' 或 'success")
			return "ok", nil

		}

	}

	return "", errors.New("异步回调失败")
}

// 生成签名
func generateSignature(data dto.PaymentNotification_request) string {
	// 按固定顺序拼接字段和值
	/* signatureString := fmt.Sprintf(
		"trade_id=%s&order_id=%s&amount=%.2f&actual_amount=%.2f&token=%s&block_transaction_id=%s&status=%d",
		data.TradeID,
		data.OrderID,
		data.Amount,
		data.ActualAmount,
		data.Token,
		data.BlockTransactionID,
		data.Status,
	) */

	// 按照 key=value 格式并按字典顺序排序参数
	params := []string{
		fmt.Sprintf("trade_id=%s", data.TradeID),
		fmt.Sprintf("order_id=%s", data.OrderID),
		fmt.Sprintf("amount=%g", data.Amount),
		fmt.Sprintf("actual_amount=%g", data.ActualAmount),
		fmt.Sprintf("token=%s", data.Token),
		fmt.Sprintf("block_transaction_id=%s", data.BlockTransactionID),
		fmt.Sprintf("status=%d", data.Status),
	}

	// 排序参数
	sort.Strings(params)

	// 使用 strings.Join 连接排序后的参数
	signatureString := strings.Join(params, "&") + config.GetApiAuthToken()
	// 打印拼接的参数
	log.Logger.Info("异步回调的拼接的参数", zap.Any("signatureString", signatureString))

	// 计算 MD5 哈希值
	hash := md5.Sum([]byte(signatureString))
	return hex.EncodeToString(hash[:]) // 转为十六进制字符串
}

// 解锁钱包地址和金额
func unlockWalletAddressAndAmount(v sdb.Orders) {
	// 解锁钱包地址和金额
	address_amount := fmt.Sprintf("%s_%f", v.Token, v.ActualAmount)
	cx := context.Background()
	err := rdb.RDB.Del(cx, address_amount).Err()
	if err != nil {
		log.Logger.Info("钱包地址和金额解锁失败", zap.Any("err", err))
		// return err
	}
}

func (j UsdtCheckJob) processCallback(v sdb.Orders) {
	// 解锁钱包地址和金额|| 异步进程解锁钱包地址和金额
	go unlockWalletAddressAndAmount(v)

	// 异步回调

	paymentNotification := dto.PaymentNotification_request{
		TradeID:            v.TradeId,
		OrderID:            v.OrderId,
		Amount:             v.Amount,
		ActualAmount:       v.ActualAmount,
		Token:              v.Token,
		BlockTransactionID: v.BlockTransactionId,
		Status:             v.Status,
	}
	// 生成签名
	signature := generateSignature(paymentNotification)
	paymentNotification.Signature = signature
	// 异步回调最大次数5次

	// 使用事务简化回调确认

	for i := 0; i < 5; i++ {
		ok, err := sendAsyncPost(v.NotifyUrl, paymentNotification)
		if ok == "ok" {
			v.CallBackConfirm = sdb.CallBackConfirmOk
			// sdb.SDB.Save(&v)
			// 事务回调函数内部始终使用 tx 参数进行操作，这是GORM事务的正确使用方式
			sdb.SDB.Save(&v)
			log.Logger.Info("已经确认订单支付成功，并把回调CallBackConfirm设置为1")

			break
		}
		if err != nil {

			log.Logger.Info("异步回调失败", zap.Any("err", err))
			// 回调次数+1
			// sdb.SDB.Model(&v).Update("callback_num", i+1)
			// sdb.SDB.Model(&v).UpdateColumn("callback_num", gorm.Expr("callback_num + 1"))
			// if err := sdb.SDB.Model(&v).UpdateColumn("callback_num", gorm.Expr("callback_num + ?", 1)).Error; err != nil {
			// 	log.Logger.Info("更新回调失败次数失败", zap.Any("err", err))
			// }
			if err := sdb.SDB.Model(&v).UpdateColumn("callback_num", gorm.Expr("callback_num + ?", 1)).Error; err != nil {
				log.Logger.Info("更新回调失败次数失败", zap.Any("err", err))
			}
			// 延迟0.5秒
			time.Sleep(500 * time.Millisecond)

			// 进入下次循环
			// continue
		}
	}
	// 发送Bark通知|| 异步进程发送通知
	go notification.Start(v)

}
