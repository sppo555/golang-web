package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"myproject/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从cookie中获取token
		cookie, err := r.Cookie("token")
		if err != nil {
			logger.LogMessage(logger.WARN, "未找到token: %v", err)
			http.Error(w, "未授权", http.StatusUnauthorized)
			return
		}

		// 解析并验证token
		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			// 从环境变量获取密钥
			secretKey := os.Getenv("JWT_SECRET")
			if secretKey == "" {
				logger.LogMessage(logger.ERROR, "JWT_SECRET 环境变量未设置")
				return nil, fmt.Errorf("JWT_SECRET 环境变量未设置")
			}
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			logger.LogMessage(logger.WARN, "无效的token: %v", err)
			http.Error(w, "无效的token", http.StatusUnauthorized)
			return
		}

		// 提取user_id并将其存储在请求的上下文中
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if userID, ok := claims["user_id"].(float64); ok {
				// 将user_id存储到请求的上下文中
				ctx := context.WithValue(r.Context(), "user_id", int(userID))
				// 将更新后的请求上下文传递给下一个处理器
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		logger.LogMessage(logger.ERROR, "token中未找到有效的user_id")
		http.Error(w, "无效的token", http.StatusUnauthorized)
	}
}
