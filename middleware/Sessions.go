package middleware

import (
	"fmt"
	"net/http"

	"exam/constant"
)

// AuthMiddleware 登录验证中间件
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取会话
		session, err := constant.Store.Get(r, "session-name")
		if err != nil {
			http.Error(w, fmt.Sprintf("无法获取会话: %v", err), http.StatusInternalServerError)
			return
		}

		// 判断是否登录
		auth, ok := session.Values["authenticated"].(bool)
		if !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}
