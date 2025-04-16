package middleware

import (
	"github.com/gorilla/sessions"
	"net/http"
)

// 定义一个会话存储
var store = sessions.NewCookieStore([]byte("your-secret-key"))

// 登录验证中间件
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取会话
		session, err := store.Get(r, "session-name")
		if err != nil {
			http.Error(w, "无法获取会话", http.StatusInternalServerError)
			return
		}

		// 检查用户是否已经登录
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// 用户已经登录，继续处理请求
		next(w, r)
	}
}
