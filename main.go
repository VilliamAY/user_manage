package main

import (
	"exam/handler"
	_ "exam/handler"
	"exam/middleware"
	"exam/utils"
	"fmt"
	"net/http"
)

func main() {
	//连接数据库
	err := utils.ConnectDB()
	if err != nil {
		fmt.Println(err)
	}
	defer utils.CloseDB()
	defer middleware.CloseLogFiles()
	//加载静态文件
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// 使用日志记录中间件
	mux := http.NewServeMux()
	mux.HandleFunc("/login", handler.Login)
	mux.HandleFunc("/register", handler.Register)
	mux.HandleFunc("/index", handler.Index)
	mux.HandleFunc("/userList", handler.UserList)
	mux.HandleFunc("/api/users/delete/{id}", handler.DeleteUserHandler)
	mux.HandleFunc("/api/users/create", handler.CreateUserHandler)
	mux.HandleFunc("/api/users/search/{username}", handler.SearchUserHandler)
	mux.HandleFunc("/api/users/update/{id}", handler.UpdateUserHandler)

	// 嵌套中间件
	handlerWithRecover := middleware.RecoverMiddleware(mux)
	loggedMux := middleware.LoggingMiddleware(handlerWithRecover)

	myServer := http.Server{Addr: ":8090", Handler: loggedMux}

	myServer.ListenAndServe()
}
