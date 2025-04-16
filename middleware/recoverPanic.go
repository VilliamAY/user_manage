package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
)

// RecoverMiddleware 全局异常捕获中间件
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// 记录错误信息和堆栈跟踪信息
				log.Printf("[Recover] panic: %v\n%s", err, debug.Stack())

				// 构建错误响应
				errorResponse := struct {
					Code int    `json:"code"`
					Msg  string `json:"msg"`
				}{
					Code: http.StatusInternalServerError,
					Msg:  "服务器内部错误，请联系管理员",
				}

				// 设置响应头和状态码
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				// 返回错误响应
				response, err := json.Marshal(errorResponse)
				if err != nil {
					w.Write([]byte(`{"code":500,"msg":"服务器内部错误，无法返回错误信息"}`))
				} else {
					w.Write(response)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}
