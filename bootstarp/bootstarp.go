package bootstarp

import (
	"U_PAY/config"
	"U_PAY/cron"
	"U_PAY/db/rdb"
	"U_PAY/db/sdb"
	"U_PAY/http"
	"U_PAY/log"
	"U_PAY/mq"
)

func Start() {
	// 初始化配置
	config.Init()
	// 初始化日志记录器
	log.Init()
	// 初始化redis
	rdb.Init()

	/*
		ctx := context.Background()
		// rdb.RDB.Set(ctx, "test", "test", 0)

		res, err := rdb.RDB.Get(ctx, "test").Result()
		if err != nil {
			log.Logger.Error("redis 获取失败")
			panic(err)
		}
		fmt.Printf("test: %v\n", res)
		// 关闭redis链接
		rdb.RDB.Close()
		// 初始化sqlite数据库
		sdb.Init()

		// 插入数据
		 err = sdb.SDB.Create(&sdb.User{Name: "张三", Email: "zhangsan@example.com", Password: "123456"}).Error
		if err != nil {
			log.Logger.Error("sqlite 插入失败")
			panic(err)
		}
		log.Logger.Info("sqlite 插入成功")

		// 查询数据

		var user sdb.User
		if err := sdb.SDB.First(&user, "email = ?", "zhangsan@example.com").Error; err != nil {
			fmt.Printf("查询失败: %v\n", err)
			return
		}
		fmt.Printf("姓名: %v\n", user.Name)
		fmt.Printf("邮箱: %v\n", user.Email)
		fmt.Printf("密码: %v\n", user.Password) */
	// 初始化sqlite数据库
	sdb.Init()
	// 初始化cron定时任务
	go cron.Start()
	// 初始化http服务

	// 初始化mq任务队列服务
	mq.Start()
	/* // 初始化telegram机器人
	telegram.Init()
	// 启动telegram机器人
	go telegram.Start() */
	// 初始化http服务
	http.Init()

}
