package tron

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

/* type TransferDetails struct {
	TokenAbbr     string
	TransactionID string
	Quant         float64
	FromAddress   string
	ToAddress     string
	FinalResult   string // 可以用来表示 API 请求是否成功
} */

// --- 为解析 JSON 响应定义的结构体 ---
type TokenInfo struct {
	Symbol   string `json:"symbol"`
	Address  string `json:"address"`
	Decimals int    `json:"decimals"`
	Name     string `json:"name"`
}

type TransactionData struct {
	TransactionID  string    `json:"transaction_id"`
	TokenInfo      TokenInfo `json:"token_info"`
	BlockTimestamp int64     `json:"block_timestamp"`
	From           string    `json:"from"`
	To             string    `json:"to"`
	Type           string    `json:"type"`
	Value          string    `json:"value"`
}

type Meta struct {
	At          int64             `json:"at"`
	Fingerprint string            `json:"fingerprint"`
	Links       map[string]string `json:"links"`
	PageSize    int               `json:"page_size"`
}

type ApiResponseGrid struct {
	Data    []TransactionData `json:"data"`
	Success bool              `json:"success"`
	Meta    Meta              `json:"meta"`
}

// --- 结束 JSON 结构体定义 ---

// 传入钱包地址
// 注意：startTime 和 endTime 参数当前未在此函数实现中使用
func GetTransactionsGrid(toAddress string, startTime int64, endTime int64) (TransferDetails, error) {
	// --- 开始修改 GetTransactions 函数 ---
	details := TransferDetails{} // 初始化返回的结构体

	// 1. 构造请求 URL
	// 注意：这里硬编码了合约地址和 limit=1，根据需要可以将其作为参数传入
	contractAddress := "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t" // USDT TRC20 合约地址
	limit := 1
	min_timestamp := startTime
	max_timestamp := endTime

	apiURL := fmt.Sprintf("https://api.trongrid.io/v1/accounts/%s/transactions/trc20?contract_address=%s&limit=%d&only_confirmed=true&min_block_timestamp=%v&max_block_timestamp=%v",
		toAddress, contractAddress, limit, min_timestamp, max_timestamp)

	// 2. 发送 HTTP GET 请求
	resp, err := http.Get(apiURL)
	if err != nil {
		return details, fmt.Errorf("请求 API 失败: %w", err)
	}
	defer resp.Body.Close()

	// 3. 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return details, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return details, fmt.Errorf("API 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	// 4. 解析 返回的JSON 数据到结构体
	var apiResponse ApiResponseGrid
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return details, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	// 5. 检查 API 返回是否成功以及是否有数据
	if !apiResponse.Success {
		details.FinalResult = "API request not successful"
		return details, fmt.Errorf("API 报告请求不成功")
	}
	details.FinalResult = "Success" // API 请求成功

	if len(apiResponse.Data) == 0 {
		details.FinalResult = "Success, but no transactions found"
		// 没有找到交易记录，返回空的 details 但没有错误
		return details, fmt.Errorf("没有找到交易记录")
	}

	// 检查block_timestamp	时间是否在指定范围内,且，建议方向是到toAddress的记录,且，代币是USDT
	if (apiResponse.Data[0].BlockTimestamp > startTime && apiResponse.Data[0].BlockTimestamp < endTime) && (apiResponse.Data[0].To == toAddress) && (apiResponse.Data[0].TokenInfo.Symbol == "USDT") {

		// 6. 提取第一条交易记录的信息
		transaction := apiResponse.Data[0]

		/* // 7. 转换金额
		valueInt, err := strconv.ParseInt(transaction.Value, 10, 64)
		if err != nil {
			return details, fmt.Errorf("转换交易金额 '%s' 失败: %w", transaction.Value, err)
		}
		// 根据 decimals 计算实际数量
		divisor := math.Pow10(transaction.TokenInfo.Decimals)
		quant := float64(valueInt) / divisor */

		// 8. 填充 TransferDetails 结构体
		// 注意：这里假设 TokenInfo 中包含了 Token 的缩写(代币的缩写USDT)
		details.TokenAbbr = transaction.TokenInfo.Symbol
		// 注意：这里假设 TransactionID 是一个字符串（哈希值）
		details.TransactionID = transaction.TransactionID
		// 注意：这里假设 Quant 是一个浮点数（金额）
		details.Quant = formatAmount(transaction.Value)
		// 转账来源地址
		details.FromAddress = transaction.From
		// 转账目标地址
		details.ToAddress = transaction.To
		// FinalResult 已在前面设置

		return details, nil
		// --- 结束修改 GetTransactions 函数 ---
	} else {
		details.FinalResult = "Success, but no transactions found"
		// 没有找到交易记录，返回空的 details ，和时间不符的错误
		details.FinalResult = "Transaction timestamp not in specified range"
		return details, fmt.Errorf("交易时间戳 %d 不在指定范围内 [%d, %d] 或者没有找到入账记录或者代币非USDT",
			apiResponse.Data[0].BlockTimestamp, startTime, endTime)
	}
}

/* func formatAmount(quant string) float64 {
	// 直接将字符串转为 float64 类型
	amount, err := strconv.ParseFloat(quant, 64)
	if err != nil {
		log.Printf("Error parsing amount: %v", err)
		return 0 // 如果转换失败，返回 0
	}

	// 使用 1e6 计算金额，转换为 float64 类型
	amountFloat := amount / 1e6 // 使用 1e6 来处理精度

	// 保留小数点后2位
	amountFloat = math.Round(amountFloat*100) / 100

	return amountFloat
}
*/
