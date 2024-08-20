package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"myproject/pkg/database"
	"myproject/pkg/logger"
	"net/http"
	"strings"
	"time"
)

func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			logger.LogMessage(logger.WARN, "未提供授权令牌")
			http.Error(w, "未提供授权令牌", http.StatusUnauthorized)
			return
		}

		// 移除可能的 "Bearer " 前缀
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		logger.LogMessage(logger.DEBUG, "接收到的令牌: %s", tokenString)

		var user struct {
			ID        uint
			Token     string
			ExpiresAt string
		}

		// 执行数据库查询
		err := database.DB.QueryRow("SELECT id, token, expires_at FROM users WHERE token = ?", tokenString).Scan(&user.ID, &user.Token, &user.ExpiresAt)

		if err != nil {
			if err == sql.ErrNoRows {
				logger.LogMessage(logger.WARN, "数据库中未找到匹配的令牌")
				http.Error(w, "无效的令牌", http.StatusUnauthorized)
			} else {
				logger.LogMessage(logger.ERROR, "数据库查询错误: %v", err)
				http.Error(w, "数据库查询错误: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// 解析 expires_at
		expiresAt, err := time.Parse("2006-01-02 15:04:05", user.ExpiresAt)
		if err != nil {
			logger.LogMessage(logger.ERROR, "解析 expires_at 错误: %v", err)
			http.Error(w, "服务器内部错误", http.StatusInternalServerError)
			return
		}

		logger.LogMessage(logger.DEBUG, "查询结果 - ID: %d, Token: %s, ExpiresAt: %s", user.ID, user.Token, expiresAt.Format(time.RFC3339))

		// 验证令牌是否匹配
		if user.Token != tokenString {
			logger.LogMessage(logger.WARN, "令牌不匹配 - 数据库: %s, 请求: %s", user.Token, tokenString)
			http.Error(w, "令牌不匹配", http.StatusUnauthorized)
			return
		}

		// 检查令牌是否过期
		if time.Now().After(expiresAt) {
			logger.LogMessage(logger.WARN, "令牌已过期 - 过期时间: %v, 当前时间: %v", expiresAt, time.Now())
			http.Error(w, "令牌已过期", http.StatusUnauthorized)
			return
		}

		logger.LogMessage(logger.INFO, "验证成功 - 用户ID: %d", user.ID)

		// 将 user_id 存储在请求上下文中
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
