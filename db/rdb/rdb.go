package rdb

import (
	"U_PAY/config"
	"U_PAY/log"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func Init() {
	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		// 基本连接配置
		Addr:     fmt.Sprintf("%s:%d", config.GetRedisHost(), config.GetRedisPort()), // Redis 地址
		Password: config.GetRedisPasswd(),                                            // Redis 密码
		DB:       config.GetRedisDb(),                                                // 数据库编号

		// 连接超时设置
		DialTimeout:  10 * time.Second, // 建立连接超时时间
		ReadTimeout:  30 * time.Second, // 读取超时时间
		WriteTimeout: 30 * time.Second, // 写入超时时间

		// 连接池设置
		PoolSize:        10,               // 连接池最大连接数
		MinIdleConns:    5,                // 最小空闲连接数
		PoolTimeout:     4 * time.Second,  // 从连接池获取连接的超时时间
		ConnMaxLifetime: 30 * time.Minute, // 连接的最大存活时间（替代 MaxConnAge）
		ConnMaxIdleTime: 5 * time.Minute,  // 空闲连接超时时间（替代 IdleTimeout）

		// 其他设置
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			// 连接建立时的回调函数
			return nil
		},
	})
	ctx := context.Background()
	RDB = rdb
	// defer rdb.Close()  在其他调用时最后关闭
	// 测试连接
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		// redis 连接失败写入日志
		log.Logger.Error("redis 连接失败")
		panic(err)
	}
	fmt.Println(pong)
	// 测试redis是否连接成功 写入日志
	log.Logger.Info("redis 连接成功")

	/* 	// 测试访问不存在的键
	   	_, err = rdb.Get(ctx, "520").Result()
	   	if err != nil {
	   		log.Logger.Info("redis 访问不存在的键")
	   	} */

}

// Close 优雅关闭 Redis 连接
