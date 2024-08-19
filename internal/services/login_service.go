package services

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"myproject/pkg/database"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// HandleLogin 處理用戶登錄邏輯
func HandleLogin(w http.ResponseWriter, r *http.Request) error {
	username := r.FormValue("username")
	password := r.FormValue("password")

	log.Printf("%s [INFO] 嘗試登錄用戶: %s", time.Now().Format(time.RFC3339), username)

	var userID int
	var storedHash string
	err := database.DB.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", username).Scan(&userID, &storedHash)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("%s [WARN] 用戶不存在: %s", time.Now().Format(time.RFC3339), username)
			return errors.New("用戶不存在")
		}
		log.Printf("%s [ERROR] 數據庫查詢錯誤: %v", time.Now().Format(time.RFC3339), err)
		log.Printf("%s [DEBUG] 用戶名: %s, 密碼: %s", time.Now().Format(time.RFC3339), username, password)
		return errors.New("服務器錯誤")
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		log.Printf("%s [WARN] 密碼不正確: %s", time.Now().Format(time.RFC3339), username)
		return errors.New("密碼不正確")
	}

	// 生成JWT
	expirationTime := time.Now().Add(time.Hour * 24)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     expirationTime.Unix(),
	})

	tokenString, err := token.SignedString([]byte("your-secret-key")) // 請替換為安全的密鑰
	if err != nil {
		log.Printf("%s [ERROR] 生成JWT錯誤: %v", time.Now().Format(time.RFC3339), err)
		return errors.New("生成令牌失敗")
	}

	// 將token和過期時間存入數據庫
	_, err = database.DB.Exec("UPDATE users SET token = ?, expires_at = ? WHERE id = ?", tokenString, expirationTime, userID)
	if err != nil {
		log.Printf("%s [ERROR] 存儲token錯誤: %v", time.Now().Format(time.RFC3339), err)
		return errors.New("存儲令牌失敗")
	}

	log.Printf("%s [INFO] 用戶登錄成功: %s (ID: %d)", time.Now().Format(time.RFC3339), username, userID)
	w.Write([]byte("登錄成功，令牌: " + tokenString))
	return nil
}
