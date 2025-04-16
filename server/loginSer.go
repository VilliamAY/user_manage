package server

import (
	"exam/constant"
	"exam/middleware"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// CheckUsernameExists 检查用户名是否存在,存在为true
func CheckUsernameExists(username string) (bool, error) {
	query := "SELECT COUNT(*) FROM users WHERE username = ?"
	stmt, err := constant.DB.Prepare(query)
	if err != nil {
		middleware.LogDBOperation("预处理失败", query, err)
		return false, fmt.Errorf("预处理失败：%w", err)
	}
	defer stmt.Close()

	var count int
	err = stmt.QueryRow(username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("查询数据库失败：%w", err)
	}

	return count > 0, nil
}

// CheckNamePwd 判断用户名和密码是否正确
func CheckNamePwd(username, password string) (bool, string) {
	exists, _ := CheckUsernameExists(username)

	if !exists {
		return false, "用户不存在"
	}

	query := "SELECT password FROM users WHERE username = ?"
	middleware.LogDBOperation("执行查询", query, username)
	var hashedPassword string
	err := constant.DB.QueryRow(query, username).Scan(&hashedPassword)
	if err != nil {
		return false, "查询数据库失败"
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false, "密码错误"
	}

	middleware.LogDBOperation("登录成功", "", username)
	return true, "登录成功"
}

// CheckStatus 检查用户是否被禁用
func CheckStatus(username string) (error, bool) {
	query := "SELECT status FROM users WHERE username = ?"
	middleware.LogDBOperation("执行查询", query, username)
	var status string
	err := constant.DB.QueryRow(query, username).Scan(&status)
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
	now := time.Now().Format("2006-01-02 15:04:05")
	query := "UPDATE users SET lastLogin = ? WHERE username = ?"
	middleware.LogDBOperation("Preparing query", query, now, username)
	stmt, err := constant.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	middleware.LogDBOperation("Executing query", query, now, username)
	_, err = stmt.Exec(now, username)
	if err != nil {
		middleware.LogDBOperation("Failed to execute query", query, err)
	}
	return err
}
