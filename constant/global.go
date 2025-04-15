package constant

import "database/sql"

// DB 全局变量
var DB *sql.DB

// User 结构体
type User struct {
	Id        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	LastLogin string `json:"lastLogin"`
	Status    string `json:"status"`
	Avatar    string `json:"avatar"`
}

// LoginData 表示每日统计数据
type LoginData struct {
	Date  string `json:"date"` // 对应 `login_date`
	Count int    `json:"count"`
}

// StatsResponse 包含所有统计数据的响应结构
type StatsResponse struct {
	UserGrowthRate   float64     `json:"userGrowthRate"`
	TotalUsers       int         `json:"totalUsers"`
	MonthLogins      int         `json:"monthLogins"`
	LoginGrowthRate  float64     `json:"loginGrowthRate"`
	DeactivatedUsers int         `json:"deactivatedUsers"`
	DeactivatedRate  float64     `json:"deactivatedRate"`
	LoginTrend       []LoginData `json:"loginTrend"`
}
