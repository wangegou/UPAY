package dto

import "github.com/shopspring/decimal"

// 定义一个异步通知请求参数的结构体

type PaymentNotification_request struct {
	TradeID            string  `json:"trade_id"`
	OrderID            string  `json:"order_id"`
	Amount             float64 `json:"amount"`
	ActualAmount       float64 `json:"actual_amount"`
	Token              string  `json:"token"`
	BlockTransactionID string  `json:"block_transaction_id"`
	Signature          string  `json:"signature"`
	Status             int     `json:"status"`
}

type Data struct {
	TradeID        string
	OrderID        string
	Amount         decimal.Decimal
	ActualAmount   decimal.Decimal
	Token          string
	ExpirationTime int64
	PaymentURL     string
}

// 定义返回的结构体|创建订单后返回的数据
type Response struct {
	StatusCode int
	Message    string
	Data       Data
	RequestID  string
}

// 定义模版所需数据视图模型
// 模版所需数据视图模型
type PaymentViewModel struct {
	TradeId                string  `json:"tradeId"`
	ActualAmount           float64 `json:"actualAmount"`
	Token                  string  `json:"token"`
	ExpirationTime         int64   `json:"expirationTime"`
	RedirectUrl            string  `json:"redirectUrl"` // 添加重定向URL
	AppName                string  `json:"appName"`
	CustomerServiceContact string  `json:"customerServiceContact"`
}

// RequestParams 用于存储请求参数
type RequestParams struct {
	OrderID     string  `json:"order_id" validate:"required"`
	Amount      float64 `json:"amount" validate:"required"`
	NotifyURL   string  `json:"notify_url" validate:"required,url"`
	RedirectURL string  `json:"redirect_url" validate:"required,url"`
	Signature   string  `json:"signature" validate:"required"`
}

// 定义字符图案
var Pattern = []string{
	"▗▖ ▗▖▗▄▄▄▖▗▖   ▗▖    ▗▄▖     ▗▖ ▗▖▗▄▄▖  ▗▄▖▗▖  ▗▖",
	"▐▌ ▐▌▐▌   ▐▌   ▐▌   ▐▌ ▐▌    ▐▌ ▐▌▐▌ ▐▌▐▌ ▐▌▝▚▞▘ ",
	"▐▛▀▜▌▐▛▀▀▘▐▌   ▐▌   ▐▌ ▐▌    ▐▌ ▐▌▐▛▀▘ ▐▛▀▜▌ ▐▌  ",
	"▐▌ ▐▌▐▙▄▄▖▐▙▄▄▖▐▙▄▄▖▝▚▄▞▘    ▝▚▄▞▘▐▌   ▐▌ ▐▌ ▐▌  ",
}
