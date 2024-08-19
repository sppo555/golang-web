package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"myproject/pkg/database"
	"net/http"
	"strings"
	"time"
)

func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "未提供授权令牌", http.StatusUnauthorized)
			return
		}

		// 移除可能的 "Bearer " 前缀
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		// 打印接收到的令牌
		// log.Printf("接收到的令牌: %s", tokenString)

		var user struct {
			ID        uint
			Token     string
			ExpiresAt string // 使用 string 类型来接收日期时间字符串
		}

		// 执行数据库查询
		err := database.DB.QueryRow("SELECT id, token, expires_at FROM users WHERE token = ?", tokenString).Scan(&user.ID, &user.Token, &user.ExpiresAt)

		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("%s 数据库中未找到匹配的令牌", time.Now().Format(time.RFC3339))
				http.Error(w, "无效的令牌", http.StatusUnauthorized)
			} else {
				log.Printf("数据库查询错误: %v", err)
				http.Error(w, "数据库查询错误: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// 解析 expires_at
		expiresAt, err := time.Parse("2006-01-02 15:04:05", user.ExpiresAt)
		if err != nil {
			log.Printf("解析 expires_at 错误: %v", err)
			http.Error(w, "服务器内部错误", http.StatusInternalServerError)
			return
		}

		// 修改日志输出格式
		log.SetFlags(0)
		log.SetOutput(LogWriter{})

		// 打印查询结果
		// logMsg := fmt.Sprintf("查询结果 - ID: %d, Token: %s, ExpiresAt: %s", user.ID, user.Token, expiresAt.Format(time.RFC3339))
		// log.Println(logMsg)

		// 验证令牌是否匹配
		if user.Token != tokenString {
			log.Printf("令牌不匹配 - 数据库: %s, 请求: %s", user.Token, tokenString)
			http.Error(w, "令牌不匹配", http.StatusUnauthorized)
			return
		}

		// 检查令牌是否过期
		if time.Now().After(expiresAt) {
			log.Printf("令牌已过期 - 过期时间: %v, 当前时间: %v", expiresAt, time.Now())
			http.Error(w, "令牌已过期", http.StatusUnauthorized)
			return
		}

		log.Printf("验证成功 - 用户ID: %d", user.ID)

		// �� user_id 存储在请求上下文中
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", int(user.ID))
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// 辅助函数：用于发送JSON响应
func sendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// 添加自定义的 LogWriter
type LogWriter struct{}

func (writer LogWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().Format(time.RFC3339) + " " + string(bytes))
}
