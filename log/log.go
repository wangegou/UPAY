package log

import (
	"U_PAY/config"
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func Init() {
	// 创建一个 Console 编码器，输出更易读的文本格式
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// 添加以下配置来显示调用者信息
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder // 显示调用者信息
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder   // 时间格式
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 创建一个日志核心，输出到文件
	fileCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   "logs/upay.log",
			MaxSize:    config.GetLogMaxSize(), // MB
			MaxBackups: config.GetLogMaxBackups(),
			MaxAge:     config.GetLogMaxAge(), // days
		}),
		zap.InfoLevel,
	)

	// 创建另一个日志核心，输出到标准输出
	consoleCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zap.InfoLevel,
	)

	// 使用 zapcore.NewTee 将两个核心组合起来
	log_zap := zap.New(zapcore.NewTee(fileCore, consoleCore),
		zap.AddCaller(),      // 添加调用者信息
		zap.AddCallerSkip(0), // 调整调用栈跳过的帧数
	)

	// 将 logger 设置为全局变量
	Logger = log_zap
	// 确保 logger 在退出时进行同步
	defer Logger.Sync()
}
