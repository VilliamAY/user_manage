package server

import (
	"exam/constant"
	"exam/middleware"
	"fmt"
	"time"
)

//这里写数据库操作

func GetStatisticsData() (constant.StatsResponse, error) {
	var stats constant.StatsResponse
	var err error

	// 获取总用户数
	query := "SELECT COUNT(*) FROM users"
	middleware.LogDBOperation("执行查询", query)
	if err = constant.DB.QueryRow(query).Scan(&stats.TotalUsers); err != nil {
		return stats, fmt.Errorf("获得总用户数失败: %v", err)
	}

	// 获取本月登录人次
	query = `
		SELECT IFNULL(SUM(login_count), 0) 
		FROM login_stats 
		WHERE login_date >= DATE_FORMAT(NOW(), '%Y-%m-01')
	`
	middleware.LogDBOperation("执行查询", query)
	if err = constant.DB.QueryRow(query).Scan(&stats.MonthLogins); err != nil {
		return stats, fmt.Errorf("无法获取本月登录信息: %v", err)
	}

	// 获取注销用户数量
	query = "SELECT SUM(deleted_users) FROM login_stats"
	middleware.LogDBOperation("执行查询", query)
	if err = constant.DB.QueryRow(query).Scan(&stats.DeactivatedUsers); err != nil {
		return stats, fmt.Errorf("无法获取禁用的用户: %v", err)
	}

	// 获取登录增长率
	currentLoginQuery := `
		SELECT IFNULL(SUM(login_count), 0) 
		FROM login_stats 
		WHERE login_date >= DATE_FORMAT(NOW(), '%Y-%m-01')
	`
	lastLoginQuery := `
		SELECT IFNULL(SUM(login_count), 0) 
		FROM login_stats 
		WHERE login_date >= DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 1 MONTH), '%Y-%m-01')
		AND login_date < DATE_FORMAT(NOW(), '%Y-%m-01')
	`
	middleware.LogDBOperation("计算登录增长率", currentLoginQuery, lastLoginQuery)
	if stats.LoginGrowthRate, err = calculateGrowthRate(currentLoginQuery, lastLoginQuery); err != nil {
		return stats, fmt.Errorf("failed to calculate login growth rate: %v", err)
	}

	// 获取用户增长率
	currentUserQuery := `
		SELECT IFNULL(SUM(new_users), 0) 
		FROM login_stats 
		WHERE login_date >= DATE_FORMAT(NOW(), '%Y-%m-01')
	`
	lastUserQuery := `
		SELECT IFNULL(SUM(new_users), 0) 
		FROM login_stats 
		WHERE login_date >= DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 1 MONTH), '%Y-%m-01')
		AND login_date < DATE_FORMAT(NOW(), '%Y-%m-01')
	`
	middleware.LogDBOperation("计算用户增长率", currentUserQuery, lastUserQuery)
	if stats.UserGrowthRate, err = calculateGrowthRate(currentUserQuery, lastUserQuery); err != nil {
		return stats, fmt.Errorf("failed to calculate user growth rate: %v", err)
	}

	// 获取注销用户增长率
	currentDeactivatedQuery := `
		SELECT IFNULL(SUM(deleted_users), 0) 
		FROM login_stats 
		WHERE login_date >= DATE_FORMAT(NOW(), '%Y-%m-01')
	`
	lastDeactivatedQuery := `
		SELECT IFNULL(SUM(deleted_users), 0) 
		FROM login_stats 
		WHERE login_date >= DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 1 MONTH), '%Y-%m-01')
		AND login_date < DATE_FORMAT(NOW(), '%Y-%m-01')
	`
	middleware.LogDBOperation("正在计算停用率", currentDeactivatedQuery, lastDeactivatedQuery)
	if stats.DeactivatedRate, err = calculateGrowthRate(currentDeactivatedQuery, lastDeactivatedQuery); err != nil {
		return stats, fmt.Errorf("failed to calculate deactivated rate: %v", err)
	}

	// 获取登录趋势数据（过去30天）
	middleware.LogDBOperation("获取过去30天的登录趋势数据", "")
	if stats.LoginTrend, err = getDailyStats(30); err != nil {
		return stats, fmt.Errorf("failed to get login trend: %v", err)
	}

	return stats, nil
}

// getDailyStats 获取每日统计数据并补全缺失日期
func getDailyStats(days int) ([]constant.LoginData, error) {
	// 初始化结果数组
	var trend []constant.LoginData

	// 查询数据库中的数据
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	query := `
        SELECT login_date, login_count
        FROM login_stats
        WHERE login_date >= ? AND login_date <= ?
        ORDER BY login_date
    `
	middleware.LogDBOperation("执行查询", query, startDate, endDate)
	rows, err := constant.DB.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %v", err)
	}
	defer rows.Close()

	// 将数据库中的数据存储到 map 中，key 为日期，value 为访问量
	dbData := make(map[string]int)
	for rows.Next() {
		var rawDate string
		var count int
		if err = rows.Scan(&rawDate, &count); err != nil {
			middleware.LogDBOperation("无法扫描行", query, err)
			return nil, fmt.Errorf("扫描行失败: %v", err)
		}

		// 确保日期格式一致（如果数据库中的日期包含时间部分或 ISO 8601 格式）
		parsedDate, err := parseDate(rawDate) // 调用自定义解析函数
		if err != nil {
			middleware.LogDBOperation("日期解析失败:", err.Error())
			continue
		}
		formattedDate := parsedDate.Format("2006-01-02")
		dbData[formattedDate] = count
	}

	// 生成完整的日期范围，并补充缺失的日期为 0
	for i := days; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02") // 当前日期减去 i 天
		count := dbData[date]                                     // 如果存在该日期的数据，则取值；否则为 0
		trend = append(trend, constant.LoginData{
			Date:  date,
			Count: count,
		})
	}

	return trend, nil
}

// 自定义日期解析函数，支持多种日期格式
func parseDate(rawDate string) (time.Time, error) {
	// 尝试解析 ISO 8601 格式（如 2025-04-01T00:00:00Z）
	if parsed, err := time.Parse(time.RFC3339, rawDate); err == nil {
		return parsed, nil
	}

	// 尝试解析带时间部分的格式（如 2025-04-01 00:00:00）
	if parsed, err := time.Parse("2006-01-02 15:04:05", rawDate); err == nil {
		return parsed, nil
	}

	// 尝试解析纯日期格式（如 2025-04-01）
	if parsed, err := time.Parse("2006-01-02", rawDate); err == nil {
		return parsed, nil
	}

	// 如果所有尝试都失败，返回错误
	return time.Time{}, fmt.Errorf("无法解析日期: %s", rawDate)
}

// calculateGrowthRate 计算增长率
func calculateGrowthRate(currentQuery, lastQuery string) (float64, error) {
	var current, last int

	middleware.LogDBOperation("正在执行当前查询", currentQuery)
	if err := constant.DB.QueryRow(currentQuery).Scan(&current); err != nil {
		return 0, err
	}

	if err := constant.DB.QueryRow(lastQuery).Scan(&last); err != nil {
		return 0, err
	}

	if last == 0 {
		if current == 0 {
			return 0, nil
		}
		return 100, nil // 上月为0，本月有数据，增长率为100%
	}

	return float64(current-last) / float64(last) * 100, nil
}
