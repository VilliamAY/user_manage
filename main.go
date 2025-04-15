package main

import (
	"exam/handler"
	_ "exam/handler"
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
	//加载静态文件
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	myServer := http.Server{Addr: ":8090"}
	http.HandleFunc("/login", handler.Login)
	http.HandleFunc("/register", handler.Register)
	http.HandleFunc("/index", handler.Index)
	http.HandleFunc("/userList", handler.UserList)
	http.HandleFunc("/api/users/delete/{id}", handler.DeleteUserHandler)
	http.HandleFunc("/api/users/create", handler.CreateUserHandler)
	http.HandleFunc("/api/users/search/{username}", handler.SearchUserHandler)
	http.HandleFunc("/api/users/update/{id}", handler.UpdateUserHandler)
	//http.HandleFunc("/api/stats", handler.GetStats)
	myServer.ListenAndServe()
}
