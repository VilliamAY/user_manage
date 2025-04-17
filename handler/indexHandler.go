package handler

import (
	"exam/constant"
	"exam/middleware"
	"exam/server"
	"exam/utils"
	"html/template"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 获取统计数据
		stats, err := server.GetStatisticsData()
		if err != nil {
			middleware.OtherLog("获取统计数据失败")
			utils.ReturnJson(w, false, "获取统计数据失败", http.StatusInternalServerError)
			return
		}

		t, err := template.ParseFiles("view/index.html")
		if err != nil {
			middleware.OtherLog("加载静态页面失败")
			utils.ReturnJson(w, false, "加载静态页面失败", http.StatusInternalServerError)
			return
		}

		session, _ := constant.Store.Get(r, "session-name")
		nowAvatar, ok := session.Values["avatar"].(string)
		if !ok {
			nowAvatar = "/static/1.jpg" // 默认头像
		}

		data := struct {
			Stats  constant.StatsResponse
			Avatar string
		}{
			stats,
			nowAvatar,
		}

		err = t.Execute(w, data)
		if err != nil {
			middleware.OtherLog("传送stats数据失败")
			utils.ReturnJson(w, false, "传送stats数据失败", http.StatusInternalServerError)
		}
	}
}
