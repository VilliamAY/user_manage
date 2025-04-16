package server

import (
	"exam/constant"
	"exam/middleware"
	"exam/utils"
)

// AddUser 把用户存储在数据库中
func AddUser(username, email, password, confirmPassword string) (string, bool) {
	// 记录检查用户名是否存在的操作
	middleware.LogDBOperation("Checking if username exists", "", username)
	// 检查用户名是否已存在
	if ok, _ := CheckUsernameExists(username); ok {
		return "用户名已存在", false
	}

	// 检查两次输入的密码是否一致
	if password != confirmPassword {
		return "两次输入的密码不一致", false
	}

	// 使用 bcrypt 加密密码
	hashedPassword, err := utils.HashedPassword(password)
	if err != nil {
		return "密码加密失败", false
	}

	query := "INSERT INTO users (username, email, password) VALUES (?, ?, ?)"
	// 记录预处理语句的操作
	middleware.LogDBOperation("Preparing query", query, username, email, hashedPassword)
	// 使用预处理语句防止 SQL 注入
	stmt, err := constant.DB.Prepare(query)
	if err != nil {
		return "预处理失败", false
	}
	defer stmt.Close()

	// 记录执行插入操作
	middleware.LogDBOperation("执行插入查询", query, username, email, hashedPassword)
	_, err = stmt.Exec(username, email, hashedPassword)
	if err != nil {
		return "插入数据失败", false
	}

	return "", true
}
