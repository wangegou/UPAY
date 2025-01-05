package mq

import (
	"U_PAY/config"
	"U_PAY/db/sdb"
	"U_PAY/log"
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

var Client *asynq.Client

func Start() {
	// 获取redis地址
	addr := fmt.Sprintf("%s:%d", config.GetRedisHost(), config.GetRedisPort())

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: addr})
	Client = client

	go async_server_run()

}

// QueueOrderExpiration 订单过期任务的队列名称
const QueueOrderExpiration = "order:expiration"

func TaskOrderExpiration(payload string, expirationDuration time.Duration) {
	task := asynq.NewTask(QueueOrderExpiration, []byte(payload)) // 转换为字节切片
	// 将任务加入队列
	info, err := Client.Enqueue(task, asynq.ProcessIn(expirationDuration))
	if err != nil {
		log.Logger.Info("任务加入失败:" + err.Error())
	}
	log.Logger.Info("任务已加入队列:", zap.Any("info", info))
}

func async_server_run() {
	mux := asynq.NewServeMux()
	mux.HandleFunc(QueueOrderExpiration, handleCheckStatusCodeTask)
	// 获取redis地址
	addr := fmt.Sprintf("%s:%d", config.GetRedisHost(), config.GetRedisPort())
	server := asynq.NewServer(asynq.RedisClientOpt{Addr: addr}, asynq.Config{Concurrency: 10})
	if err := server.Run(mux); err != nil {
		log.Logger.Info("Error starting server:", zap.Any("err", err))
	}
}

// 处理过期任务
func handleCheckStatusCodeTask(ctx context.Context, t *asynq.Task) error {

	// 提取任务载荷传入的交易ID，根据ID去查一下订单记录里面的支付状态是否是待支付，如果是待支付，改为已过期
	// 订单过期后，需要解锁钱包地址和金额【从Redis里删除】
	payload := string(t.Payload())

	var order sdb.Orders

	err := sdb.SDB.First(&order, "trade_id = ?", payload).Error
	if err != nil {
		log.Logger.Info("订单查询失败")
		return err
	}

	if order.Status == sdb.StatusWaitPay {
		order.Status = sdb.StatusExpired
		sdb.SDB.Save(&order)
		log.Logger.Info(fmt.Sprintf("订单%v已设置为过期", order.TradeId))
	}

	// 订单过期后，需要解锁钱包地址和金额【从Redis里删除】

	/*

		这里不显性解锁也可以，因为在创建订单的时候，已经将钱包地址和金额锁定，并设置了过期时间，到期后自动删除


		address_amount := fmt.Sprintf("%s_%f", order.Token, order.ActualAmount)
		cx := context.Background()
		err = rdb.RDB.Del(cx, address_amount).Err()
		if err != nil {
			log.Logger.Info("钱包地址和金额解锁失败")
			return err
		} */

	return nil
}
