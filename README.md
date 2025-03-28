# UPAY 项目初衷

原项目采用的是 mysql 数据库，在日志里经常看到数据库链接被拒绝，又查不到原因，干脆就从零开始学习 go，阅读 EPUSDT 的代码，，开始重构，并采用 SQLite 数据库，使其更符合当前的开发需求。

# 项目简介

本项目是基于 EPUSDT 二次开发，没有改变 EPUSDT 的 API，只是将 EPUSDT 的代码结构进行了重构，使其更符合当前的开发需求。

原项目地址：https://github.com/assimon/epusdt/tree/master

## 项目特点

.
├── 支持私有化部署
├── 采用 GIN 框架，支持高并发
├── 支持多个钱包轮训
├── 支持 HTTP API 接入
├── 支持 BARK 通知
├── 不需要额外安装任何依赖
├── 直接运行脚本即可
├── 采用 SQLite 数据库，数据存储在本地，不需要额外安装数据库

# 快速开始

视频教程：https://www.youtube.com/watch?v=gR43V3aFhk0

1. 下载项目
2. 将 env. example 改为.env 并填写信息
3. 运行脚本

启动命令为：UPAY 所在目录+U_PAY 执行文件

与原项目的使用差别是：1 不需要初始化数据库，2，启动命令后面无需任何参数

## 流量限制需要配置 NGINX【可选配置】

- \# 真实 IP 解析配置（新增部分）
- set_real_ip_from 127.0.0.1;

\# 示例：Cloudflare 的 IPv4/IPv6 地址

- \# IPv4
  set_real_ip_from 173.245.48.0/20;
  set_real_ip_from 103.21.244.0/22;
  set_real_ip_from 103.22.200.0/22;
  set_real_ip_from 103.31.4.0/22;
  set_real_ip_from 141.101.64.0/18;
  set_real_ip_from 108.162.192.0/18;
  set_real_ip_from 190.93.240.0/20;
  set_real_ip_from 188.114.96.0/20;
  set_real_ip_from 197.234.240.0/22;
  set_real_ip_from 198.41.128.0/17;
  set_real_ip_from 162.158.0.0/15;
  set_real_ip_from 104.16.0.0/13;
  set_real_ip_from 104.24.0.0/14;
  set_real_ip_from 172.64.0.0/13;
  set_real_ip_from 131.0.72.0/22;

\# IPv6

set_real_ip_from 2400:cb00::/32;
set_real_ip_from 2606:4700::/32;
set_real_ip_from 2803:f800::/32;
set_real_ip_from 2405:b500::/32;
set_real_ip_from 2405:8100::/32;
set_real_ip_from 2a06:98c0::/29;
set_real_ip_from 2c0f:f248::/32;

- real_ip_header X-Forwarded-For;
- real_ip_recursive on;

- \# 全局代理标头（新增部分）
- proxy_set_header Host $host;
- proxy_set_header X-Real-IP $remote_addr;
- proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
- proxy_set_header X-Forwarded-Proto $scheme;

![UPAY服务器配置](https://github.com/wangegou/UPAY/blob/master/images/Server.nane%20upey.1oMpp.lcu.png)

# API 文档

https://github.com/wangegou/UPAY/blob/master/API.md

### 在线测试：

https://huojian.iosapp.icu/
下单时：支付方式选择 UPAY 即可

### 插件

独角数卡 ：插件支持原 EPUSDT 插件
独角数卡后台 ：秘钥里填写：http://127.0.0.1:8080/api/create_order

二次元|荔枝发卡 ：插件支持原 EPUSDT 插件
对接文档：https://blog.hi-kvm.com/archives/244.html

彩虹易支付：在本项目 plugins 目录下；

# 反馈

欢迎反馈问题，请在 GitHub 上提交问题，或者在项目中提交 PR。

电报：https://t.me/hellokvm
群组：https://t.me/UPAY_BUG
邮箱：8888@iosapp.icu

# 声明

UPAY 为开源的产品，仅用于学习交流使用！
不可用于任何违反中华人民共和国(含台湾省)或使用者所在地区法律法规的用途。
因为作者即本人仅完成代码的开发和开源活动(开源即任何人都可以下载使用或修改分发)，从未参与用户的任何运营和盈利活动。
且不知晓用户后续将程序源代码用于何种用途，故用户使用过程中所带来的任何法律责任即由用户自己承担
