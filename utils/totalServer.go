package utils

import (
	"database/sql"
	"encoding/json"
	"exam/constant"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

// ConnectDB 连接数据库
func ConnectDB() error {
	var err error
	constant.DB, err = sql.Open("mysql", "root:123456@tcp(localhost:3306)/webexam?parseTime=true")
	if err != nil {
		return err
	}
	// 测试连接是否成功
	err = constant.DB.Ping()
	if err != nil {
		return err
	}
	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() {
	if constant.DB != nil {
		err := constant.DB.Close()
		if err != nil {
			log.Printf("关闭数据库失败: %v", err)
		} else {
			fmt.Println("数据库连接已关闭")
		}
	}
}

// HashedPassword 使用 bcrypt 加密密码
func HashedPassword(password string) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("加密失败：", err)
		return []byte{}, err
	}
	return hashedPassword, nil
}

// RespondWithJSON 返回json响应
func RespondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	// 将数据转换为JSON
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("无法编码响应数据"))
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}

// UpdateNewUsersStat 更新 login_stats 表中的 new_users 字段
func UpdateNewUsersStat() error {
	// 获取今天的日期
	today := time.Now().Format("2006-01-02")

	// 检查今天是否已经有记录
	var count int
	err := constant.DB.QueryRow("SELECT COUNT(*) FROM login_stats WHERE login_date = ?", today).Scan(&count)
	if err != nil {
		return fmt.Errorf("查询 login_stats 失败: %v", err)
	}

	if count == 0 {
		// 如果没有今天的记录，插入一条新记录
		_, err = constant.DB.Exec("INSERT INTO login_stats (login_date, new_users) VALUES (?, 1)", today)
		if err != nil {
			return fmt.Errorf("插入 login_stats 失败: %v", err)
		}
	} else {
		// 如果已经有今天的记录，更新 new_users 字段
		_, err = constant.DB.Exec("UPDATE login_stats SET new_users = new_users + 1 WHERE login_date = ?", today)
		if err != nil {
			return fmt.Errorf("更新 login_stats 失败: %v", err)
		}
	}

	return nil
}

// UpdateDeletedUsersStat 更新 login_stats 表中的 deleted_users 字段
func UpdateDeletedUsersStat() error {
	// 获取今天的日期
	today := time.Now().Format("2006-01-02")

	// 检查今天是否已经有记录
	var count int
	err := constant.DB.QueryRow("SELECT COUNT(*) FROM login_stats WHERE login_date = ?", today).Scan(&count)
	if err != nil {
		return fmt.Errorf("查询 login_stats 失败: %v", err)
	}

	if count == 0 {
		// 如果没有今天的记录，插入一条新记录
		_, err = constant.DB.Exec("INSERT INTO login_stats (login_date, deleted_users) VALUES (?, 1)", today)
		if err != nil {
			return fmt.Errorf("插入 login_stats 失败: %v", err)
		}
	} else {
		// 如果已经有今天的记录，更新 deleted_users 字段
		_, err = constant.DB.Exec("UPDATE login_stats SET deleted_users = deleted_users + 1 WHERE login_date = ?", today)
		if err != nil {
			return fmt.Errorf("更新 login_stats 失败: %v", err)
		}
	}

	return nil
}
