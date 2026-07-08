package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func A() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		fmt.Println("A: before")
		c.Next() // 放行给下一个中间件/处理器
		fmt.Println("A: after  | cost =", time.Since(start))
	}
}

func B() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("B: before")
		c.Next() // 放行
		fmt.Println("B: after")
	}
}

func main() {
	r := gin.New()
	r.Use(A(), B()) // 注册顺序：先 A 后 B

	r.GET("/hello", func(c *gin.Context) {
		fmt.Println("Handler: doing work")
		c.String(200, "hello")
	})

	// ❗用于测试 panic 恢复的路由
	//r.GET("/panic-nil", func(c *gin.Context) {
	//	var p *int
	//	fmt.Println(*p) // 故意 panic，验证 Recovery 中间件
	//})

	r.Run(":8080")
}
