package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"myproject/pkg/database"
)

// HandleBalanceUpdate 處理餘額修改邏輯
func HandleBalanceUpdate(w http.ResponseWriter, r *http.Request) error {
	// 假設 JWTMiddleware 已經驗證了 token 並將用戶 ID 存儲在上下文中
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		log.Printf("%s [ERROR] 無法獲取用戶ID，Context值：%v", time.Now().Format(time.RFC3339), r.Context().Value("user_id"))
		return errors.New("無法獲取用戶ID")
	}
	log.Printf("%s [DEBUG] 成功獲取用戶ID：%v", time.Now().Format(time.RFC3339), userID)

	balanceStr := r.FormValue("balance")
	amountStr := r.FormValue("amount")

	log.SetFlags(0)
	log.Printf("%s [INFO] 嘗試更新用戶 %d 的餘額", time.Now().Format(time.RFC3339), int(userID))

	var newBalance float64

	if balanceStr != "" {
		var parseErr error
		newBalance, parseErr = strconv.ParseFloat(balanceStr, 64)
		if parseErr != nil {
			log.Printf("%s [ERROR] 無效的餘額格式: %s，錯誤：%v", time.Now().Format(time.RFC3339), balanceStr, parseErr)
			return errors.New("無效的餘額格式")
		}
		log.Printf("%s [DEBUG] 解析餘額成功：%v", time.Now().Format(time.RFC3339), newBalance)
	} else if amountStr != "" {
		amount, parseErr := strconv.ParseFloat(amountStr, 64)
		if parseErr != nil {
			log.Printf("%s [ERROR] 無效的金額格式: %s，錯誤：%v", time.Now().Format(time.RFC3339), amountStr, parseErr)
			return errors.New("無效的金額格式")
		}
		log.Printf("%s [DEBUG] 解析金額成功：%v", time.Now().Format(time.RFC3339), amount)

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
			return errors.New("更���餘額時發生錯誤")
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
		log.Printf("%s [ERROR] 未提供餘額或金額，FormValue：%v", time.Now().Format(time.RFC3339), r.Form)
		return errors.New("未提供餘額或金額")
	}

	log.Printf("%s [INFO] 用戶 %d 的餘額更新成功，新餘額為 %.2f", time.Now().Format(time.RFC3339), int(userID), newBalance)
	w.Write([]byte(fmt.Sprintf("餘額更新成功，新餘額為 %.2f", newBalance)))
	return nil
}
