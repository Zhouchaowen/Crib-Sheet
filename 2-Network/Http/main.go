// main.go
package main

import (
	"fmt"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("收到请求: %s %s\n", r.Method, r.URL.Path)
	time.Sleep(100 * time.Millisecond)

	// 通过设置响应头来控制连接是否保持
	if r.URL.Path == "/close" {
		w.Header().Set("Connection", "close")
	}

	w.Write([]byte("Hello World"))
}

func main() {
	// 默认的 Server 配置就支持长连接
	http.HandleFunc("/", handler)
	http.HandleFunc("/close", handler)

	fmt.Println("服务器启动在 :8080")
	http.ListenAndServe(":8080", nil)
}
