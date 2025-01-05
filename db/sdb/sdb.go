package sdb

import (
	"U_PAY/config"
	"U_PAY/log"

	"gorm.io/driver/sqlite" // 引入 SQLite 驱动
	"gorm.io/gorm"          // 引入 GORM 库
)

var SDB *gorm.DB

// 订单状态
const (
	StatusWaitPay     = 1 // 等待支付
	StatusPaySuccess  = 2 // 支付成功
	StatusExpired     = 3 // 已过期
	CallBackConfirmOk = 1 // 回调已确认
	CallBackConfirmNo = 2 // 回调未确认
)

// 订单表
type Orders struct {
	gorm.Model
	TradeId            string  // UPAY订单号
	OrderId            string  // 客户交易id
	BlockTransactionId string  // 区块id
	Amount             float64 // 订单金额，保留4位小数
	ActualAmount       float64 // 订单实际需要支付的金额，保留4位小数
	Token              string  // 所属钱包地址
	Status             int     // 1：等待支付，2：支付成功，3：已过期

	NotifyUrl       string // 异步回调地址
	RedirectUrl     string // 同步回调地址
	CallbackNum     int    // 回调次数
	CallBackConfirm int    // 回调是否已确认 1是 2否
	StartTime       int64  // 订单开始时间（时间戳）
	ExpirationTime  int64  // 订单过期时间（时间戳）

}

// 钱包状态
const (
	TokenStatusEnable  = 1 // 钱包启用
	TokenStatusDisable = 2 // 钱包禁用
)

// 钱包地址表
type WalletAddress struct {
	gorm.Model
	Token  string // 钱包token
	Status int    // 1:启用 2:禁用

}

func Init() {
	// 初始化数据库连接
	db, err := gorm.Open(sqlite.Open("u_pay.db"), &gorm.Config{})
	if err != nil {
		log.Logger.Info("数据库连接失败")
	}
	log.Logger.Info("数据库连接成功")
	// 将数据库实例赋值给SDB 方便其他模块调用
	SDB = db
	// 自动迁移 - 根据模型创建或更新表结构
	err = db.AutoMigrate(&Orders{}, &WalletAddress{})
	if err != nil {
		log.Logger.Info("数据库迁移失败")
	}
	log.Logger.Info("数据库迁移成功")
	/* // 创建一个订单记录
	order := Orders{
		OrderId:         "1234567890",
		Amount:          100.00,
		ActualAmount:    100.00,
		Token:           "TXFchkPBXBkHGifcUdQdErhBzP5T6kRfyv",
		Status:          StatusWaitPay,
		NotifyUrl:       "https://www.baidu.com",
		RedirectUrl:     "https://www.baidu.com",
		CallbackNum:     0,
		CallBackConfirm: 1,
	}
	if db.Create(&order).Error != nil {
		log.Logger.Info("创建订单记录失败")
	} else {
		log.Logger.Info("创建订单记录成功")
	} */

	/* var wallet WalletAddress
	// 创建一个钱包地址 先查询钱包地址存在不存在，存在不在创建，不存在就创建
	if db.First(&wallet, "token = ?", "TXFchkPBXBkHGifcUdQdErhBzP5T6kRfyv").Error != nil {
		db.Create(&WalletAddress{Token: "TXFchkPBXBkHGifcUdQdErhBzP5T6kRfyv", Status: TokenStatusEnable})
		log.Logger.Info("创建钱包地址成功")
	} else {
		log.Logger.Info("钱包地址已存在，不创建")


	} */

	// 因为钱包地址是从.env中获取的，所以需要先获取到钱包地址，然后创建钱包地址记录
	// 先查询钱包地址是否存在，存在不在创建，不存在就创建

	var wallet WalletAddress
	if db.First(&wallet, "token = ?", config.GetWalletAddress()).Error != nil {
		WalletAddresses := WalletAddress{
			Token:  config.GetWalletAddress(),
			Status: config.GetWalletStatus(),
		}
		// 在数据库中的WalletAddress表中插入钱包地址记录
		err = db.Create(&WalletAddresses).Error
		if err != nil {
			log.Logger.Info("创建钱包地址记录失败")
		} else {
			log.Logger.Info("创建钱包地址记录成功")
		}
	} else {
		log.Logger.Info("钱包地址已存在，不创建")
	}

}
