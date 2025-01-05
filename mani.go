package main

import (
	"U_PAY/bootstarp"
	"U_PAY/log"

	"go.uber.org/zap"
)

func main() {
	// 初始化各项服务
	bootstarp.Start()

	// fmt.Println("我是在bootstarp.go中打印的")

	// 测试md5加密

	/* a := "hello world"

	b := md5.Sum([]byte(a))

	c := fmt.Sprintf("%x", b)

	fmt.Printf("c: %v\n", c)

	sin := "5eb63bbbe01eeed093cb22bb8f5acdc3"

	if c == sin {
		fmt.Println("验证成功")
	} else {
		fmt.Println("验证失败")
	}
	*/

	defer func() {
		if err := recover(); err != nil {
			// 捕获异常
			log.Logger.Info("捕获异常:", zap.Any("err", err))
		}
	}()

	/* 作用
	   异常处理: 这段代码的主要作用是捕获运行时的恐慌（panic）。在 Go 中，恐慌通常表示程序遇到了无法恢复的错误，例如数组越界、空指针解引用等。
	   defer 语句: defer 语句会在包含它的函数返回时执行。无论函数是正常返回还是由于恐慌而提前返回，defer 中的代码都会被执行。
	   3. recover 函数: recover 是一个内置函数，用于从恐慌中恢复。它可以在 defer 函数中调用，以获取恐慌的值并防止程序崩溃。如果没有发生恐慌，recover 返回 nil。
	   输出错误信息: 如果捕获到恐慌，代码会打印出 "捕获异常:" 后跟恐慌的具体错误信息。这有助于调试和了解程序在运行时遇到的问题。 */

}
