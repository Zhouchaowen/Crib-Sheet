// client.go
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func main() {
	// 测试持久连接
	fmt.Println("测试非持久连接：")
	testKeepAlive(false)
}

func testKeepAlive(keepalive bool) {
	// 创建自定义 Transport
	tr := &http.Transport{
		DisableKeepAlives: !keepalive,
	}

	// 创建客户端
	client := &http.Client{Transport: tr}

	// 连续发送5个请求
	for i := 0; i < 5; i++ {
		start := time.Now()

		url := "http://localhost:8080"
		if !keepalive {
			url = "http://localhost:8080/close"
		}

		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("请求错误: %v\n", err)
			continue
		}

		_, _ = ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		fmt.Printf("请求 %d: 状态=%s, 耗时=%v, Connection=%s\n",
			i+1,
			resp.Status,
			time.Since(start),
			resp.Header.Get("Connection"))
	}
}
