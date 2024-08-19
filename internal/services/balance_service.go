package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"myproject/pkg/database"

	"github.com/dgrijalva/jwt-go"
)

// HandleBalanceUpdate 處理餘額修改邏輯
func HandleBalanceUpdate(w http.ResponseWriter, r *http.Request) error {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		log.Printf("%s [ERROR] 未提供令牌", time.Now().Format(time.RFC3339))
		return errors.New("未提供令牌")
	}

	// 移除可能存在的 "Bearer " 前綴
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-secret-key"), nil // 請使用與生成令牌時相同的密鑰
	})

	if err != nil {
		log.Printf("%s [ERROR] 解析令牌時發生錯誤: %v", time.Now().Format(time.RFC3339), err)
		return errors.New("無效的令牌")
	}

	if !token.Valid {
		log.Printf("%s [ERROR] 令牌無效", time.Now().Format(time.RFC3339))
		return errors.New("令牌無效")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("%s [ERROR] 無效的令牌聲明", time.Now().Format(time.RFC3339))
		return errors.New("無效的令牌聲明")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		log.Printf("%s [ERROR] 無效的用戶ID", time.Now().Format(time.RFC3339))
		return errors.New("無效的用戶ID")
	}

	// 檢查令牌是否正確以及是否過期
	var expiresAtStr string
	err = database.DB.QueryRow("SELECT expires_at FROM users WHERE id = ? AND token = ?", int(userID), tokenString).Scan(&expiresAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("%s [ERROR] 用戶令牌不匹配或用戶不存在: %d", time.Now().Format(time.RFC3339), int(userID))
			return errors.New("無效的用戶或令牌")
		}
		log.Printf("%s [ERROR] 檢查用戶令牌時發生錯誤: %v", time.Now().Format(time.RFC3339), err)
		return errors.New("檢查用戶令牌時發生錯誤")
	}

	expiresAt, err := time.Parse("2006-01-02 15:04:05", expiresAtStr)
	if err != nil {
		log.Printf("%s [ERROR] 解析過期時間時發生錯誤: %v", time.Now().Format(time.RFC3339), err)
		return errors.New("解析過期時間時發生錯誤")
	}

	if time.Now().After(expiresAt) {
		log.Printf("%s [ERROR] 用戶令牌已過期: %d", time.Now().Format(time.RFC3339), int(userID))
		return errors.New("令牌已過期")
	}

	balanceStr := r.FormValue("balance")
	amountStr := r.FormValue("amount")

	log.SetFlags(0)
	log.Printf("%s [INFO] 嘗試更新用戶 %d 的餘額", time.Now().Format(time.RFC3339), int(userID))

	var newBalance float64

	if balanceStr != "" {
		var parseErr error
		newBalance, parseErr = strconv.ParseFloat(balanceStr, 64)
		if parseErr != nil {
			log.Printf("%s [ERROR] 無效的餘額格式: %s", time.Now().Format(time.RFC3339), balanceStr)
			return errors.New("無效的餘額格式")
		}
	} else if amountStr != "" {
		amount, parseErr := strconv.ParseFloat(amountStr, 64)
		if parseErr != nil {
			log.Printf("%s [ERROR] 無效的金額格式: %s", time.Now().Format(time.RFC3339), amountStr)
			return errors.New("無效的金額格式")
		}

		// 開始事務
		tx, txErr := database.DB.Begin()
		if txErr != nil {
			log.Printf("%s [ERROR] 開始事務時發生錯誤: %v", time.Now().Format(time.RFC3339), txErr)
			return errors.New("開始事務時發生錯誤")
		}
		defer tx.Rollback() // 如果提交失敗，回滾事務

		// 獲取當前餘額
		var currentBalance float64
		queryErr := tx.QueryRow("SELECT balance FROM users_balances WHERE user_id = ? FOR UPDATE", int(userID)).Scan(&currentBalance)
		if queryErr != nil {
			if queryErr == sql.ErrNoRows {
				log.Printf("%s [ERROR] 未找到用戶: %d", time.Now().Format(time.RFC3339), int(userID))
				return errors.New("未找到該用戶")
			}
			log.Printf("%s [ERROR] 檢查用戶餘額時發生錯誤: %v", time.Now().Format(time.RFC3339), queryErr)
			return errors.New("檢查用戶餘額時發生錯誤")
		}

		newBalance = currentBalance + amount

		// 更新users_balances表中的balance欄位
		result, execErr := tx.Exec("UPDATE users_balances SET balance = ? WHERE user_id = ?", newBalance, int(userID))
		if execErr != nil {
			log.Printf("%s [ERROR] 更新餘額時發生錯誤: %v", time.Now().Format(time.RFC3339), execErr)
			return errors.New("更新餘額時發生錯誤")
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			log.Printf("%s [WARN] 未找到用戶或餘額未變更: %d", time.Now().Format(time.RFC3339), int(userID))
			return errors.New("未找到該用戶或餘額未變更")
		}

		// 提交事務
		if commitErr := tx.Commit(); commitErr != nil {
			log.Printf("%s [ERROR] 提交事務時發生錯誤: %v", time.Now().Format(time.RFC3339), commitErr)
			return errors.New("提交事務時發生錯誤")
		}
	} else {
		log.Printf("%s [ERROR] 未提供餘額或金額", time.Now().Format(time.RFC3339))
		return errors.New("未提供餘額或金額")
	}

	log.Printf("%s [INFO] 用戶 %d 的餘額更新成功，新餘額為 %.2f", time.Now().Format(time.RFC3339), int(userID), newBalance)
	w.Write([]byte(fmt.Sprintf("餘額更新成功，新餘額為 %.2f", newBalance)))
	return nil
}
