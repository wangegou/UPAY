<?php

// 定义 UPAY 支付插件类
class UPAY_plugin {
    // 插件的基本信息，包括名称、作者、支付类型和输入参数
    static public $info = [
        'name'        => 'UPAY', // 插件名称
        'showname'    => 'UPAY', // 展示名称
        'author'      => 'UPAY', // 作者
        'link'        => 'https://example.com', // 官方链接
        'types'       => ['USDT'], // 支持的支付类型（此处仅支持 USDT）
        'inputs' => [ // 插件需要的输入参数
            'appurl' => [ // API 接口地址
                'name' => 'API接口地址',
                'type' => 'input',
                'note' => '以http://或https://开头，末尾不要有斜线/', // API 地址格式要求
            ],
            'appkey' => [ // API Token，用于签名
                'name' => 'API Token',
                'type' => 'input',
                'note' => '输入UPAY的API Token',
            ],
            'appid' => [ // 应用 ID（在此插件中未使用）
                'name' => 'APP ID',
                'type' => 'input',
                'note' => '输入任意字符即可',
            ],
        ],
        'select' => null, // 预留的下拉选择框（当前未使用）
        'note' => '', // 预留的备注信息
        'bindwxmp' => false, // 是否绑定微信公众号
        'bindwxa' => false, // 是否绑定微信小程序
    ];

    /**
     * 处理支付提交
     * @return array 返回跳转到支付页面的 URL
     */
    static public function submit() {
        global $siteurl, $channel, $order, $sitename;

        // 检查订单支付类型是否为 USDT
        if (in_array($order['typename'], self::$info['types'])) {
            // 返回跳转类型的支付 URL
            return ['type' => 'jump', 'url' => '/pay/UPAY/' . TRADE_NO . '/?sitename=' . $sitename];
        }
    }

    /**
     * 移动端 API 支付调用
     * @return mixed 调用 UPAY 方法
     */
    static public function mapi() {
        global $order;
        
        // 检查订单支付类型是否为 USDT
        if (in_array($order['typename'], self::$info['types'])) {
            return self::UPAY($order['typename']);
        }
    }

    /**
     * 获取 API 接口 URL
     * @return string 返回处理后的 API 地址
     */
    static private function getApiUrl() {
        global $channel;
        
        // 获取用户设置的 API 地址
        $apiurl = $channel['appurl'];
        
        // 确保 API 地址末尾没有 '/'
        if (substr($apiurl, -1) == '/') {
            $apiurl = substr($apiurl, 0, -1);
        }
        return $apiurl;
    }

    /**
     * 发送 HTTP 请求到 UPAY API
     * @param string $url API 端点路径
     * @param array $param 发送的参数
     * @return array 返回 JSON 解析后的响应数据
     */
    static private function sendRequest($url, $param) {
        // 组合完整的 API URL
        $url = self::getApiUrl() . $url;
        
        // 将参数转换为 JSON
        $post = json_encode($param, JSON_UNESCAPED_UNICODE);
        
        // 发送 HTTP 请求
        $response = get_curl($url, $post, 0, 0, 0, 0, 0, ['Content-Type: application/json']);
        
        // 返回解析后的 JSON 响应
        return json_decode($response, true);
    }

    /**
     * 生成 API 签名
     * @param array $params 需要签名的参数
     * @param string $apiToken API Token
     * @return string 返回 MD5 签名字符串
     */
    static public function Sign($params, $apiToken) {
        // 按参数名进行 ASCII 排序
        ksort($params);
        $str = '';
        
        // 拼接参数
        foreach ($params as $k => $val) {
            if ($val !== '') {
                $str .= $k . '=' . $val . '&';
            }
        }
        
        // 在末尾拼接 API Token
        $str = rtrim($str, '&') . $apiToken;
        
        // 返回 MD5 加密的签名
        return md5($str);
    }

    /**
     * 创建 UPAY 订单
     * @return string 返回支付链接
     * @throws Exception 订单创建失败时抛出异常
     */
    static private function CreateOrder() {
        global $siteurl, $channel, $order, $conf;

        // 构造请求参数
        $param = [
            'order_id' => TRADE_NO, // 订单号
            'amount' => floatval($order['realmoney']), // 订单金额
            'notify_url' => $conf['localurl'] . 'pay/notify/' . TRADE_NO . '/', // 异步通知 URL
            'redirect_url' => $siteurl . 'pay/return/' . TRADE_NO . '/', // 同步跳转 URL
        ];

        // 生成签名
        $param['signature'] = self::Sign($param, $channel['appkey']);

        // 发送订单请求
        $result = self::sendRequest('/api/create_order', $param);

        // 处理响应数据
        if (isset($result["status_code"]) && $result["status_code"] == 200) {
            \lib\Payment::updateOrder(TRADE_NO, $result['data']);
            return $result['data']['payment_url'];
        } else {
            throw new Exception($result["message"] ?? '返回数据解析失败');
        }
    }

    /**
     * 执行 UPAY 订单创建流程
     * @return array 返回跳转支付的 URL 或错误信息
     */
    static public function UPAY() {
        try {
            $code_url = self::CreateOrder();
        } catch (Exception $ex) {
            return ['type' => 'error', 'msg' => 'UPAY创建订单失败！' . $ex->getMessage()];
        }
        return ['type' => 'jump', 'url' => $code_url];
    }

    /**
     * 处理 UPAY 异步通知
     * @return array 返回异步通知结果
     */
    static public function notify() {
        global $channel, $order;

        // 获取通知数据
        $resultJson = file_get_contents("php://input");
        $resultArr = json_decode($resultJson, true);

        // 获取签名并移除签名字段
        $Signature = $resultArr["signature"];
        unset($resultArr['signature']);

        // 计算本地签名
        $sign = self::Sign($resultArr, $channel['appkey']);

        // 校验签名是否正确
        if ($sign === $Signature) {
            $out_trade_no = $resultArr['order_id'];
            if ($out_trade_no == TRADE_NO && $resultArr['status'] == 2) {
                // 处理回调通知（订单支付成功）
                processNotify($order, $out_trade_no);
                return ['type' => 'html', 'data' => 'ok'];
            }
        }
        return ['type' => 'html', 'data' => 'fail'];
    }

    /**
     * 处理同步跳转返回
     * @return array 返回跳转页面
     */
    static public function return() {
        return ['type' => 'page', 'page' => 'return'];
    }
}
