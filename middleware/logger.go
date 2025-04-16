package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// 初始化数据库操作日志文件
var dbLogFile *os.File
var dbLogger *log.Logger

// 初始化其他日志文件
var otherLogFile *os.File
var otherLogger *log.Logger

func init() {
	// 创建或打开数据库操作日志文件
	var err error
	dbLogFile, err = os.OpenFile("log/db_operations.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("无法打开数据库操作日志文件: %v", err)
	}
	dbLogger = log.New(dbLogFile, "[DB] ", log.Ldate|log.Ltime|log.Lshortfile)

	// 创建或打开其他日志文件
	otherLogFile, err = os.OpenFile("log/other.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("无法打开其他日志文件: %v", err)
	}
	otherLogger = log.New(otherLogFile, "[OTHER] ", log.Ldate|log.Ltime|log.Lshortfile)
}

// CloseLogFiles 关闭日志文件
func CloseLogFiles() {
	if dbLogFile != nil {
		dbLogFile.Close()
	}
	if otherLogFile != nil {
		otherLogFile.Close()
	}
}

// DBLog 记录数据库操作日志
func DBLog(message string) {
	dbLogger.Println(message)
}

// OtherLog 记录其他日志
func OtherLog(message string) {
	otherLogger.Println(message)
}

// LoggingMiddleware 中间件记录请求日志
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// 记录请求信息
		OtherLog(fmt.Sprintf("Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr))
		// 调用下一个处理程序
		next.ServeHTTP(w, r)
		// 记录响应时间
		OtherLog(fmt.Sprintf("Response time: %s for %s %s", time.Since(start), r.Method, r.URL.Path))
	})
}

// LogDBOperation 封装数据库操作日志记录函数
func LogDBOperation(operation, query string, args ...interface{}) {
	var logMsg string
	if len(args) > 0 {
		logMsg = fmt.Sprintf("%s: %s, args: %v", operation, query, args)
	} else {
		logMsg = fmt.Sprintf("%s: %s", operation, query)
	}
	DBLog(logMsg)
}
