package services

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"myproject/pkg/database"
	"myproject/pkg/logger"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// HandleLogin 處理用戶登錄邏輯
func HandleLogin(w http.ResponseWriter, r *http.Request) error {
	username := r.FormValue("username")
	password := r.FormValue("password")

	logger.LogMessage(logger.INFO, "開始處理用戶登錄請求: %s", username)
	logger.LogMessage(logger.DEBUG, "接收到的登錄數據 - 用戶名: %s, 密碼長度: %d", username, len(password))

	var userID int
	var storedHash string
	logger.LogMessage(logger.DEBUG, "執行數據庫查詢以獲取用戶信息")
	err := database.DB.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", username).Scan(&userID, &storedHash)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.LogMessage(logger.WARN, "用戶不存在: %s", username)
			logger.LogMessage(logger.DEBUG, "數據庫查詢未返回結果")
			return errors.New("用戶不存在")
		}
		logger.LogMessage(logger.ERROR, "數據庫查詢錯誤: %v", err)
		logger.LogMessage(logger.DEBUG, "數據庫查詢失敗 - 用戶名: %s, 錯誤: %v", username, err)
		return errors.New("服務器錯誤")
	}

	logger.LogMessage(logger.DEBUG, "成功從數據庫獲取用戶信息 - 用戶ID: %d", userID)

	logger.LogMessage(logger.DEBUG, "開始驗證用戶密碼")
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		logger.LogMessage(logger.WARN, "密碼驗證失敗: %s", username)
		logger.LogMessage(logger.DEBUG, "密碼不匹配 - 用戶ID: %d", userID)
		return errors.New("密碼不正確")
	}
	logger.LogMessage(logger.DEBUG, "密碼驗證成功 - 用戶ID: %d", userID)

	// 生成JWT
	logger.LogMessage(logger.DEBUG, "開始生成JWT令牌")
	expirationTime := time.Now().Add(time.Hour * 24)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
	})

	tokenString, err := token.SignedString([]byte("your-secret-key")) // 請替換為安全的密鑰
	if err != nil {
		logger.LogMessage(logger.ERROR, "生成JWT錯誤: %v", err)
		logger.LogMessage(logger.DEBUG, "JWT生成失敗 - 用戶ID: %d, 錯誤: %v", userID, err)
		return errors.New("生成令牌失敗")
	}
	logger.LogMessage(logger.DEBUG, "成功生成JWT令牌 - 用戶ID: %d", userID)

	// 將token和過期時間存入數據庫
	logger.LogMessage(logger.DEBUG, "開始將令牌存儲到數據庫")
	_, err = database.DB.Exec("UPDATE users SET token = ?, expires_at = ? WHERE id = ?", tokenString, expirationTime, userID)
	if err != nil {
		logger.LogMessage(logger.ERROR, "存儲token錯誤: %v", err)
		logger.LogMessage(logger.DEBUG, "數據庫更新失敗 - 用戶ID: %d, 錯誤: %v", userID, err)
		return errors.New("存儲令牌失敗")
	}
	logger.LogMessage(logger.DEBUG, "成功將令牌存儲到數據庫 - 用戶ID: %d", userID)

	logger.LogMessage(logger.INFO, "用戶登錄成功: %s (ID: %d)", username, userID)
	logger.LogMessage(logger.DEBUG, "登錄過程完成 - 用戶ID: %d, 令牌長度: %d", userID, len(tokenString))

	w.Write([]byte("登錄成功，令牌: " + tokenString))
	return nil
}
