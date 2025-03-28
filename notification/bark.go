package notification

// 这里是bark的通知服务

import (
	"U_PAY/config"
	"U_PAY/db/sdb"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func sendBarkNotification(barkURL, title, body string) error {
	// 创建通知内容
	notification := map[string]string{
		"title": title,
		"body":  body,
	}

	// 将通知内容编码为 JSON
	jsonData, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	// 发送 POST 请求
	resp, err := http.Post(barkURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send notification, status code: %d", resp.StatusCode)
	}

	return nil
}

func Start(order sdb.Orders) {
	// 替换为你的 Bark URL
	barkURL := "https://api.day.app/" + config.GetBarkKey() // 你的 Bark 服务器 URL
	title := "UPAY 订单通知"
	// 将数据库中的数字翻译会自然语言
	var Status string
	if order.Status == 2 {
		Status = "支付成功"
	} else if order.Status == 3 {
		Status = "已过期"
	} else {
		Status = "等待支付"
	}

	var CallBackConfirm string
	if order.CallBackConfirm == sdb.CallBackConfirmOk {
		CallBackConfirm = "已回调"
	} else {
		CallBackConfirm = "未回调"
	}

	body := fmt.Sprintf("订单号:%s\n 支付金额%.2f\n支付状态:%s\n区块ID:%s\n回调状态：%s\n", order.TradeId, order.ActualAmount, Status, order.BlockTransactionId, CallBackConfirm)
	// body := "您的订单已成功创建！\n感谢您的购买！\n请查看您的订单详情。"

	// 发送通知
	err := sendBarkNotification(barkURL, title, body)
	if err != nil {
		fmt.Println("发送通知失败:", err)
	} else {
		fmt.Println("通知发送成功！")
	}
}
