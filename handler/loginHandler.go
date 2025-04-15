package handler

import (
	"encoding/json"
	"exam/constant"
	"exam/server"
	"exam/utils"
	"html/template"
	"log"
	"net/http"
	"time"
)

// Login 登录时http请求的处理函数
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 如果是 GET 请求，显示登录页面
		t, _ := template.ParseFiles("view/login.html")
		t.Execute(w, nil)
	} else if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		// 验证用户名和密码
		if ok, s := server.CheckNamePwd(username, password); !ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": s,
			})
			return
		}

		// 验证用户状态
		if _, ok := server.CheckStatus(username); !ok {
			// 用户已经被禁用
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "用户已被禁用",
			})
			return
		}

		// 更新最后登录时间
		err := server.UpdateLastLogin(username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "更新登录时间失败",
			})
			return
		}

		// 登录成功后更新今日登录统计
		today := time.Now().Format("2006-01-02")
		_, err = constant.DB.Exec(`
        INSERT INTO login_stats (login_date, login_count)
        VALUES (?, 1)
        ON DUPLICATE KEY UPDATE login_count = login_count + 1
    `, today)

		if err != nil {
			log.Printf("更新登录统计失败: %v", err)
		}

		// 登录成功，返回 JSON 响应
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"message":  "登录成功",
			"redirect": "/index",
		})
	}
}

func Register(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		// 如果是 GET 请求，显示登录页面
		t, _ := template.ParseFiles("view/login.html")
		t.Execute(w, nil)

	} else if r.Method == "POST" {

		//注册
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm-password")

		if s, ok := server.AddUser(username, email, password, confirmPassword); ok {
			err := utils.UpdateNewUsersStat()
			if err != nil {
				// 如果更新失败，记录错误日志
				log.Printf("更新 new_users 失败: %v", err)
			}

			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"message": "注册成功",
			})
		} else {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": s,
			})
			return
		}
	}
}
