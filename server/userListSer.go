package server

import (
	"exam/constant"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

//数据库操作

// GetAllUsersByAdmin 管理员获取数据库中所有用户
func GetAllUsersByAdmin(page, pageSize int) ([]constant.User, int, int, int, error) {
	// 先获取总记录数
	var totalCount int
	err := constant.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalCount)
	if err != nil {
		fmt.Println("获取总记录数失败")
		return nil, 0, 0, 0, err
	}

	// 计算总页数
	totalPages := (totalCount + pageSize - 1) / pageSize

	// 计算偏移量
	offset := (page - 1) * pageSize

	stmt, err := constant.DB.Prepare("SELECT * FROM users LIMIT ? OFFSET ?")
	if err != nil {
		fmt.Println("预处理失败")
		return nil, 0, 0, 0, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(pageSize, offset)
	if err != nil {
		fmt.Println("获取结果失败")
		return nil, 0, 0, 0, err
	}
	defer rows.Close()

	var users []constant.User
	for rows.Next() {
		var user constant.User
		var lastLoginTime time.Time

		err = rows.Scan(&user.Id, &user.Username, &user.Password, &user.Email, &user.Role, &lastLoginTime, &user.Status, &user.Avatar)
		if err != nil {
			fmt.Println("扫描行失败")
			return nil, 0, 0, 0, err
		}
		user.LastLogin = lastLoginTime.Format("2006-01-02 15:04:05")
		users = append(users, user)
	}

	// 计算分页范围
	startPage := getMax(1, page-2)
	endPage := getMin(totalPages, page+2)

	return users, totalCount, startPage, endPage, nil
}

// 辅助函数：计算最大值
func getMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// 辅助函数：计算最小值
func getMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SearchUser 根据用户名查询用户列表
func SearchUser(username string, page, pageSize int) ([]constant.User, int, error) {
	// 先获取总记录数
	var totalCount int
	err := constant.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username LIKE ?", "%"+username+"%").Scan(&totalCount)
	if err != nil {
		fmt.Println("获取总记录数失败")
		return nil, 0, err
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	stmt, err := constant.DB.Prepare("SELECT * FROM users WHERE username LIKE ? LIMIT ? OFFSET ?")
	if err != nil {
		fmt.Println("预处理失败")
		return nil, 0, err
	}
	defer stmt.Close()

	rows, err := stmt.Query("%"+username+"%", pageSize, offset)
	if err != nil {
		fmt.Println("获取结果失败")
		return nil, 0, err
	}
	defer rows.Close()

	var users []constant.User
	for rows.Next() {
		var user constant.User
		var lastLoginTime time.Time

		err = rows.Scan(&user.Id, &user.Username, &user.Password, &user.Email, &user.Role, &lastLoginTime, &user.Status, &user.Avatar)
		if err != nil {
			fmt.Println("扫描行失败")
			return nil, 0, err
		}
		user.LastLogin = lastLoginTime.Format("2006-01-02 15:04:05")
		users = append(users, user)
	}
	return users, totalCount, nil
}

// CreateUser 管理员新建用户
func CreateUser(username, password, email, role, status, avatar string) error {
	stmt, err := constant.DB.Prepare("insert into users(username,password,email,role,status,avatar)values (?,?,?,?,?,?)")
	if err != nil {
		fmt.Println("预处理失败")
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(username, password, email, role, status, avatar)
	if err != nil {
		fmt.Println("sql失败")
		return err
	}
	//获取结果
	count, err := result.RowsAffected()
	if err != nil {
		fmt.Println("获取结果失败")
		return err
	}
	if count > 0 {
		fmt.Println("新增成功")
	} else {
		fmt.Println("新增失败")
	}

	return nil
}

// DeleteUser 管理员删除用户
func DeleteUser(id int) error {
	stmt, err := constant.DB.Prepare("delete from users where id = ?")
	if err != nil {
		fmt.Println("预处理失败")
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(id)
	if err != nil {
		fmt.Println("获取结果失败")
		return err
	}
	count, _ := result.RowsAffected()
	if count > 0 {
		fmt.Println("删除成功")
	} else {
		fmt.Println("删除失败")
	}

	//删除用户头像
	//非默认头像才删除

	return nil
}

// UpdateUser 修改用户信息
func UpdateUser(username, password, email, role, status, avatarPath string, id int) error {
	// 开始事务
	tx, err := constant.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // 确保发生错误时回滚

	// 动态构建查询
	query := "UPDATE users SET username=?, email=?, role=?, status=?, avatar=?"
	args := []interface{}{username, email, role, status, avatarPath}

	// 如果有密码，添加到查询中
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		query = strings.Replace(query, "SET", "SET password=?,", 1)
		args = append([]interface{}{hashedPassword}, args...)
	}

	// 添加WHERE条件
	query += " WHERE id=?"
	args = append(args, id)

	// 执行
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err = stmt.Exec(args...); err != nil {
		return err
	}

	// 提交事务
	return tx.Commit()
}

// CheckUsernameExistsExcludingCurrent 检查用户名是否存在
func CheckUsernameExistsExcludingCurrent(username string, excludeId int) (bool, error) {
	var count int
	err := constant.DB.QueryRow(
		"SELECT COUNT(*) FROM users WHERE username = ? AND id != ?",
		username,
		excludeId,
	).Scan(&count)

	if err != nil {
		return false, err
	}
	return count > 0, nil
}
