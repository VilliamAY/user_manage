package handler

import (
	"encoding/json"
	"exam/constant"
	"exam/middleware"
	"exam/server"
	_ "exam/server"
	"exam/utils"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// DeleteUserHandler 删除用户的的路由处理
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	//从http请求的url路径中提取用户id
	path := r.URL.Path
	idStr := strings.TrimPrefix(path, "/api/users/delete/")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.OtherLog("无效的用户ID")
		http.Error(w, "无效的用户ID", http.StatusBadRequest)
		return
	}

	err = server.DeleteUser(id)
	if err != nil {
		middleware.OtherLog("删除用户失败")
		http.Error(w, "删除用户失败", http.StatusInternalServerError)
		return
	}

	err = utils.UpdateDeletedUsersStat()
	if err != nil {
		// 如果更新失败，记录错误日志
		middleware.OtherLog("更新deleted_users失败")
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "用户删除成功",
	})
}

// SearchUserHandler 搜索的的路由处理
func SearchUserHandler(w http.ResponseWriter, r *http.Request) {
	// 只处理GET请求
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取页码参数，默认为1
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// 每页显示8条记录
	pageSize := 8

	path := r.URL.Path
	username := strings.TrimPrefix(path, "/api/users/search/")

	users, totalCount, err := server.SearchUser(username, page, pageSize)
	if err != nil {
		middleware.OtherLog("搜索用户失败")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "搜索用户失败",
		})
		return
	}

	// 计算总页数
	totalPages := totalCount / pageSize
	if totalCount%pageSize != 0 {
		totalPages++
	}

	// 返回JSON响应
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"users":      users,
		"page":       page,
		"totalPages": totalPages,
		"totalCount": totalCount,
	})
}

// CreateUserHandler 创建用户的路由处理
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	//验证用户信息

	// 获取表单字段
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")
	role := r.FormValue("role")
	status := r.FormValue("status")

	password, avatarPath := CheckInfo(w, r, username, password, email, "")
	if avatarPath == "" {
		return
	}
	if password == "" {
		ReturnJson(w, false, "密码为空")
		return
	}

	err := server.CreateUser(username, password, email, role, status, avatarPath)
	if err != nil {
		middleware.OtherLog("创建用户失败")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "创建用户失败" + err.Error(),
		})
		return
	}

	err = utils.UpdateNewUsersStat()
	if err != nil {
		// 如果更新失败，记录错误日志
		middleware.OtherLog("更新new_users失败")
	}

	// 返回成功响应
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "创建用户成功",
		"avatar":  avatarPath,
	})
}

// UserList 显示用户列表
func UserList(w http.ResponseWriter, r *http.Request) {
	// 获取页码参数，默认为1
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// 每页显示8条记录
	pageSize := 8

	// 获取数据库数据（带分页）
	users, totalCount, startPage, endPage, err := server.GetAllUsersByAdmin(page, pageSize)
	if err != nil {
		middleware.OtherLog("获取用户列表失败")
		http.Error(w, "获取用户列表失败", http.StatusInternalServerError)
		return
	}

	// 计算总页数
	totalPages := (totalCount + pageSize - 1) / pageSize

	// 构造传递给模板的数据结构
	data := struct {
		Users      []constant.User
		Page       int
		TotalPages int
		TotalCount int
		StartPage  int
		EndPage    int
	}{
		Users:      users,
		Page:       page,
		TotalPages: totalPages,
		TotalCount: totalCount,
		StartPage:  startPage,
		EndPage:    endPage,
	}

	// 定义 seq 函数
	seq := func(start, end int) []int {
		var result []int
		for i := start; i <= end; i++ {
			result = append(result, i)
		}
		return result
	}

	// 定义 FuncMap 并注册 seq 函数
	funcMap := template.FuncMap{
		"seq": seq,
	}

	// 使用 FuncMap 创建模板
	t := template.New("userList.html").Funcs(funcMap)

	// 解析模板文件
	t, err = t.ParseFiles("view/userList.html")
	if err != nil {
		middleware.OtherLog("模板解析失败:")
		http.Error(w, "模板解析失败", http.StatusInternalServerError)
		return
	}

	// 渲染模板
	err = t.Execute(w, data)
	if err != nil {
		middleware.OtherLog("模板执行失败:")
		http.Error(w, "模板执行失败", http.StatusInternalServerError)
		return
	}
}

// UpdateUserHandler 修改用户信息
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 获取表单字段
	id := r.FormValue("id")
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")
	role := r.FormValue("role")
	status := r.FormValue("status")

	//检查数据
	password, avatarPath := CheckInfo(w, r, username, password, email, id)
	if avatarPath == "" {
		return
	}

	numId, _ := strconv.Atoi(id)
	err := server.UpdateUser(username, password, email, role, status, avatarPath, numId)
	if err != nil {
		middleware.OtherLog("更新用户失败")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "更新用户失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应，包含更新后的头像路径
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "用户更新成功",
		"avatar":  avatarPath,
	})
}

// CheckInfo 验证用户信息
func CheckInfo(w http.ResponseWriter, r *http.Request, username, password, email string, id string) (string, string) {
	// 验证必填字段username，email
	if username == "" || email == "" {
		ReturnJson(w, false, "用户名和邮箱是必填项")
		return "", ""
	}

	// 检查用户名是否被其他用户占用
	if id != "" {
		numID, _ := strconv.Atoi(id)
		if ok, _ := server.CheckUsernameExistsExcludingCurrent(username, numID); ok {
			ReturnJson(w, false, "用户名已被其他用户使用")
			return "", ""
		}
	} else {
		// 新建用户时检查用户名是否已存在
		if ok, _ := server.CheckUsernameExists(username); ok {
			ReturnJson(w, false, "用户名已存在")
			return "", ""
		}
	}

	// 密码处理 - 只有在密码不为空时才加密
	if password != "" {
		hashedPassword, err := utils.HashedPassword(password)
		password = string(hashedPassword)
		if err != nil {
			ReturnJson(w, false, "用户名已存在")
			return "", ""
		}
	}

	// 头像上传处理
	avatarPath := "/static/1.jpg" // 默认头像路径
	file, header, err := r.FormFile("avatar")
	if err == nil {
		defer file.Close()

		// 验证文件类型
		allowedTypes := map[string]bool{
			"image/jpeg": true,
			"image/png":  true,
		}

		//获取文件格式
		fileType := header.Header.Get("Content-Type")
		if !allowedTypes[fileType] {
			ReturnJson(w, false, "只有jpg，png格式的图片可以当头像")
			return "", ""
		}

		// 限制文件大小 (2MB)
		if header.Size > 2<<20 {
			ReturnJson(w, false, "图片文件超过2MB了")
			return "", ""
		}

		// 生成唯一文件名
		//获取文件名
		ext := filepath.Ext(header.Filename)
		//拼接用户名时间为头像文件名
		newFilename := fmt.Sprintf("%s_%d%s", username, time.Now().UnixNano(), ext)
		avatarPath = filepath.Join("static/avatar", newFilename)

		// 创建目标文件
		dst, err := os.Create(avatarPath)
		if err != nil {
			ReturnJson(w, false, "Failed to create avatar file")
			return "", ""
		}
		defer dst.Close()

		// 复制文件内容
		if _, err := io.Copy(dst, file); err != nil {
			ReturnJson(w, false, "Failed to save avatar file")
			return "", ""
		}

		// 转换为web可访问路径
		avatarPath = "/" + filepath.ToSlash(avatarPath)
	}

	return password, avatarPath
}

// ReturnJson 返回json响应
func ReturnJson(w http.ResponseWriter, success bool, message string) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"message": message,
	})
}
