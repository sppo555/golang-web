package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"myproject/pkg/database"
	"myproject/pkg/logger"
)

func init() {
	// 移除 log 包的默認前綴
	log.SetFlags(0)
}

// HandleBalanceUpdate 處理餘額修改邏輯
func HandleBalanceUpdate(w http.ResponseWriter, r *http.Request) error {
	// 假設 JWTMiddleware 已經驗證了 token 並將用戶 ID 存儲在上下文中
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		logger.LogMessage(logger.ERROR, "無法獲取用戶ID，Context值：%v", r.Context().Value("user_id"))
		return errors.New("無法獲取用戶ID")
	}
	logger.LogMessage(logger.DEBUG, "成功獲取用戶ID：%v", userID)

	balanceStr := r.FormValue("balance")
	amountStr := r.FormValue("amount")

	logger.LogMessage(logger.INFO, "嘗試更新用戶 %d 的餘額", int(userID))

	var newBalance float64

	// 開始事務
	tx, txErr := database.DB.Begin()
	if txErr != nil {
		logger.LogMessage(logger.ERROR, "開始事務時發生錯誤: %v", txErr)
		return errors.New("開始事務時發生錯誤")
	}
	defer tx.Rollback() // 如果提交失敗，回滾事務

	// 獲取當前餘額
	selectSQL := "SELECT balance FROM users_balances WHERE user_id = ? FOR UPDATE"
	logger.LogMessage(logger.DEBUG, "执行SQL: %s, %d, %s", selectSQL, userID, balanceStr)
	var currentBalance float64
	queryErr := tx.QueryRow(selectSQL, int(userID)).Scan(&currentBalance)
	if queryErr != nil {
		if queryErr == sql.ErrNoRows {
			logger.LogMessage(logger.ERROR, "未找到用戶: %d", int(userID))
			return errors.New("未找到該用戶")
		}
		logger.LogMessage(logger.ERROR, "檢查用戶餘額時發生錯誤: %v", queryErr)
		return errors.New("檢查用戶餘額時發生錯誤")
	}

	// 添加新的日誌記錄，顯示更新前的餘額
	logger.LogMessage(logger.INFO, "用戶 %d 的當前餘額為 %.2f", int(userID), currentBalance)

	if balanceStr != "" {
		var parseErr error
		newBalance, parseErr = strconv.ParseFloat(balanceStr, 64)
		if parseErr != nil {
			logger.LogMessage(logger.ERROR, "無效的餘額格式: %s，錯誤：%v", balanceStr, parseErr)
			return errors.New("無效的餘額格式")
		}
	} else if amountStr != "" {
		amount, parseErr := strconv.ParseFloat(amountStr, 64)
		if parseErr != nil {
			logger.LogMessage(logger.ERROR, "無效的金額格式: %s，錯誤：%v", amountStr, parseErr)
			return errors.New("無效的金額格式")
		}
		logger.LogMessage(logger.DEBUG, "解析金額成功：%v", amount)
		newBalance = currentBalance + amount
	} else {
		logger.LogMessage(logger.ERROR, "未提供餘額或金額，FormValue：%v", r.Form)
		return errors.New("未提供餘額或金額")
	}

	// 更新users_balances表中的balance欄位
	updateSQL := "UPDATE users_balances SET balance = ? WHERE user_id = ?"
	logger.LogMessage(logger.DEBUG, "執行SQL: %s, %d, %s", updateSQL, userID, balanceStr)
	result, execErr := tx.Exec(updateSQL, newBalance, int(userID))
	if execErr != nil {
		logger.LogMessage(logger.ERROR, "更新餘額時發生錯誤: %v", execErr)
		return errors.New("更新餘額時發生錯誤")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		logger.LogMessage(logger.WARN, "未找到用戶或餘額未變更: %d", int(userID))
		return errors.New("未找到該用戶或餘額未變更")
	}

	// 提交事務
	if commitErr := tx.Commit(); commitErr != nil {
		logger.LogMessage(logger.ERROR, "提交事務時發生錯誤: %v", commitErr)
		return errors.New("提交事務時發生錯誤")
	}

	logger.LogMessage(logger.INFO, "用戶 %d 的餘額更新成功，新餘額為 %.2f", int(userID), newBalance)
	w.Write([]byte(fmt.Sprintf("餘額更新成功，新餘額為 %.2f", newBalance)))
	return nil
}
