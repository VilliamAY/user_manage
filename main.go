package main

import (
	"exam/handler"
	"exam/middleware"
	"exam/utils"
	"fmt"
	"net/http"
)

func main() {
	// 连接数据库
	err := utils.ConnectDB()
	if err != nil {
		fmt.Println(err)
	}
	defer utils.CloseDB()
	defer middleware.CloseLogFiles()

	// 创建多路复用器
	mux := http.NewServeMux()

	// 静态文件路由（不经过中间件）
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// 动态路由
	dynamicMux := http.NewServeMux()
	dynamicMux.HandleFunc("/login", handler.Login)
	dynamicMux.HandleFunc("/register", handler.Register)
	dynamicMux.HandleFunc("/index", middleware.AuthMiddleware(handler.Index))
	dynamicMux.HandleFunc("/userList", middleware.AuthMiddleware(handler.UserList))
	dynamicMux.HandleFunc("/api/users/delete/{id}", middleware.AuthMiddleware(handler.DeleteUserHandler))
	dynamicMux.HandleFunc("/api/users/create", middleware.AuthMiddleware(handler.CreateUserHandler))
	dynamicMux.HandleFunc("/api/users/search/{username}", handler.SearchUserHandler)
	dynamicMux.HandleFunc("/api/users/update/{id}", handler.UpdateUserHandler)

	// 应用中间件到动态路由
	handlerWithRecover := middleware.RecoverMiddleware(dynamicMux)
	loggedMux := middleware.LoggingMiddleware(handlerWithRecover)

	// 将动态路由和静态文件路由合并
	mux.Handle("/", loggedMux)

	// 启动服务器
	myServer := http.Server{Addr: ":8090", Handler: mux}
	myServer.ListenAndServe()
}
