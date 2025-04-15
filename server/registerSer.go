package server

import (
	"exam/constant"
	"exam/utils"
)

// AddUser 把用户存储在数据库中
func AddUser(username, email, password, confirmPassword string) (string, bool) {
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

	// 使用预处理语句防止 SQL 注入
	stmt, err := constant.DB.Prepare("INSERT INTO users (username, email, password) VALUES (?, ?, ?)")
	if err != nil {
		return "预处理失败", false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, email, hashedPassword)
	if err != nil {
		return "插入数据失败", false
	}

	return "", true
}
