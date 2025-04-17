package handler

import (
	"exam/constant"
	"net/http"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := constant.Store.Get(r, "session-name")
	session.Options.MaxAge = -1 // 删除会话
	session.Save(r, w)

	// 禁用浏览器缓存
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
