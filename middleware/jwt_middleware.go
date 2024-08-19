// package middleware

// import (
// 	"net/http"
// 	"os"
// 	"strings"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/golang-jwt/jwt/v4"
// 	"gorm.io/gorm"
// )

// func JWTMiddleware(db *gorm.DB) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		tokenString := c.GetHeader("Authorization")
// 		if tokenString == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供授权令牌"})
// 			c.Abort()
// 			return
// 		}

// 		// 移除可能的 "Bearer " 前缀
// 		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

// 		// 使用环境变量或配置文件中的密钥
// 		secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))

// 		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 			return secretKey, nil
// 		})

// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌", "details": err.Error()})
// 			c.Abort()
// 			return
// 		}

// 		claims, ok := token.Claims.(jwt.MapClaims)
// 		if !ok {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌声明"})
// 			c.Abort()
// 			return
// 		}

// 		userID := uint(claims["user_id"].(float64))

// 		var user struct {
// 			Token     string
// 			ExpiresAt time.Time
// 		}
// 		if err := db.Table("users").Select("token", "expires_at").Where("id = ?", userID).First(&user).Error; err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在", "details": err.Error()})
// 			c.Abort()
// 			return
// 		}

// 		if user.Token != tokenString {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌不匹配"})
// 			c.Abort()
// 			return
// 		}

// 		if time.Now().After(user.ExpiresAt) {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌已过期"})
// 			c.Abort()
// 			return
// 		}

// 		// 修改这里：将 user_id 存储为 float64 类型
// 		c.Set("user_id", float64(userID))
// 		c.Next()
// 	}
// }
