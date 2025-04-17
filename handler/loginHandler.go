package handler

import (
	"encoding/json"
	"exam/constant"
	"exam/middleware"
	"exam/server"
	"exam/utils"
	"html/template"
	"log"
	"net/http"
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
			utils.ReturnJson(w, false, s, http.StatusBadRequest)
			return
		}

		// 验证用户状态
		if _, ok := server.CheckStatus(username); !ok {
			// 用户已经被禁用
			utils.ReturnJson(w, false, "用户已被禁用", http.StatusBadRequest)
			return
		}

		// 更新最后登录时间
		err := server.UpdateLastLogin(username)
		if err != nil {
			utils.ReturnJson(w, false, "更新登录时间失败", http.StatusInternalServerError)
			return
		}

		// 登录成功后更新今日登录统计
		err = server.UpdateLogins()

		if err != nil {
			middleware.OtherLog("更新登录统计失败" + err.Error())
		}

		//获取当前用户
		user := server.GetUser(username)

		// 验证成功后，存储用户角色
		session, _ := constant.Store.Get(r, "session-name")
		session.Values["authenticated"] = true
		session.Values["role"] = user.Role
		session.Values["avatar"] = user.Avatar
		session.Values["username"] = user.Username
		session.Save(r, w)

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

			utils.ReturnJson(w, true, "注册成功", http.StatusBadRequest)
		} else {
			middleware.OtherLog(s)
			utils.ReturnJson(w, false, s, http.StatusBadRequest)
			return
		}
	}
}
