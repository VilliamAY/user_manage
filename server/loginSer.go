package server

import (
	"exam/constant"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// CheckUsernameExists 检查用户名是否存在,存在为true
func CheckUsernameExists(username string) (bool, error) {
	// 使用预处理语句防止 SQL 注入
	stmt, err := constant.DB.Prepare("SELECT COUNT(*) FROM users WHERE username = ?")
	if err != nil {
		return false, fmt.Errorf("预处理失败：%w", err)
	}
	defer stmt.Close()

	// 查询数据库中是否有这个名字
	var count int
	err = stmt.QueryRow(username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("查询数据库失败：%w", err)
	}

	// 如果 count > 0，表示用户名已存在
	return count > 0, nil
}

// CheckNamePwd 判断用户名和密码是否正确
func CheckNamePwd(username, password string) (bool, string) {
	// 检查用户名是否存在
	exists, _ := CheckUsernameExists(username)

	if !exists {
		return false, "用户不存在"
	}

	// 密码是否正确
	var hashedPassword string
	err := constant.DB.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&hashedPassword)
	if err != nil {
		return false, "查询数据库失败"
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false, "密码错误"
	}

	return true, "登录成功"
}

// CheckStatus 检查用户是否被禁用
func CheckStatus(username string) (error, bool) {
	var status string
	err := constant.DB.QueryRow("SELECT status FROM users WHERE username = ?", username).Scan(&status)
	if err != nil {
		return err, false
	}
	if status == "启用" {
		return nil, true
	} else {
		return nil, false
	}
}

// UpdateLastLogin 更新用户最后登录时间
func UpdateLastLogin(username string) error {
	// 获取当前时间
	now := time.Now().Format("2006-01-02 15:04:05")

	// 准备SQL语句
	stmt, err := constant.DB.Prepare("UPDATE users SET lastLogin = ? WHERE username = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// 执行更新
	_, err = stmt.Exec(now, username)
	return err
}
