# UPAY 支付系统 API 文档

## 目录

- [概述](#概述)
- [认证方式](#认证方式)
- [接口列表](#接口列表)
  - [创建订单](#创建订单)
  - [查询订单状态](#查询订单状态)
- [错误码说明](#错误码说明)

## 概述

UPAY 是一个支付系统，提供订单创建和状态查询等功能。本文档详细说明了系统提供的 API 接口。

## 认证方式

系统使用签名认证机制，需要按照以下规则生成签名：

1. 将请求参数按照参数名 ASCII 码从小到大排序
2. 按照格式 `参数名=参数值` 拼接参数字符串，参数之间用 `&` 连接
3. 在字符串末尾拼接 API 认证 Token
4. 对最终字符串进行 MD5 加密，得到签名

### 签名生成示例

以创建订单接口为例：

原始请求参数：

```
order_id: "TEST123"
amount: 100.00
notify_url: "http://example.com/notify"
redirect_url: "http://example.com/redirect"

```

假定 API Token: "your_api_token"

1. 参数排序：

```
amount=100.00
notify_url=http://example.com/notify
order_id=TEST123
redirect_url=http://example.com/redirect
```

2. 参数拼接：

```
amount=100.00&notify_url=http://example.com/notify&order_id=TEST123&redirect_url=http://example.com/redirect
```

3. 拼接 API Token：

```
amount=100.00&notify_url=http://example.com/notify&order_id=TEST123&redirect_url=http://example.com/redirectyour_api_token
```

4. MD5 加密得到最终签名

注意事项：

- 参数值使用原始值，不需要进行 URL 编码
- 浮点数不保留末尾的 0，如 100.00 应该使用 100
- 签名计算时区分大小写
- 空值参数不参与签名

## 接口列表

### 创建订单

#### 接口说明

- 请求方式：POST
- 请求路径：/api/create_order

#### 请求参数

| 参数名       | 类型   | 必选 | 说明             |
| ------------ | ------ | ---- | ---------------- |
| order_id     | string | 是   | 商户订单号       |
| amount       | float  | 是   | 订单金额(CNY)    |
| notify_url   | string | 是   | 异步通知地址     |
| redirect_url | string | 是   | 支付完成跳转地址 |
| signature    | string | 是   | 请求签名         |

#### 响应参数

```json
{
  "status_code": 200, // 状态码
  "message": "success", // 响应信息
  "data": {
    // 响应数据
    "trade_id": "TP202312250001", // 平台订单号
    "order_id": "TEST123", // 商户订单号
    "amount": 100.0, // 订单金额(CNY)
    "actual_amount": 14.28, // 实际支付金额(USDT)
    "token": "TRX7uHHB...", // 收款钱包地址
    "expiration_time": 1703491234, // 订单过期时间(时间戳)
    "payment_url": "https://pay.upay.com/pay/TP202312250001" // 支付页面URL
  }
}
```

### 支付结果通知「异步通知」

#### 通知参数

```json
{
  "trade_id": "TP202312250001", // 平台订单号
  "order_id": "TEST123", // 商户订单号
  "amount": 100.0, // 订单金额(CNY)
  "actual_amount": 14.28, // 实际支付金额(USDT)
  "token": "TRX7uHHB...", // 收款钱包地址
  "block_transaction_id": "0x123...", // 区块链交易ID
  "status": 2, // 订单状态
  "signature": "md5..." // 通知签名
}
```

商户收到通知后，需返回字符串 "ok" 或 "success"，否则系统会重复发送通知，最多 5 次。

## 错误码说明

| 状态码 | 说明           |
| ------ | -------------- |
| 200    | 请求成功       |
| 400    | 请求参数错误   |
| 401    | 签名验证失败   |
| 500    | 服务器内部错误 |

### 常见错误说明

1. CNY 金额小于最低支付金额 0.01
2. USDT 金额小于最低支付金额 0.01
3. 没有可用的钱包地址
4. 签名验证失败
5. 创建订单记录失败

```

```

```

```
