package handler

import (
	"exam/middleware"
	"exam/server"
	"html/template"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 获取统计数据
		stats, err := server.GetStatisticsData()
		if err != nil {
			middleware.OtherLog("获取统计数据失败")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		t, err := template.ParseFiles("view/index.html")
		if err != nil {
			middleware.OtherLog("加载静态页面失败")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = t.Execute(w, stats)
		if err != nil {
			middleware.OtherLog("传送stats数据失败")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
