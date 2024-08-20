package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"myproject/pkg/database"
	"myproject/pkg/logger"
	"net/http"
	"strconv"
)

// updateProductPrice 更新商品价格
func updateProductPrice(db *sql.DB, itemName string, price float64) error {
	_, err := db.Exec("UPDATE users_price SET price = ? WHERE item_name = ?", price, itemName)
	return err
}

// getProductPrice 获取商品价格
func getProductPrice(db *sql.DB, itemName string) (float64, error) {
	var price float64
	err := db.QueryRow("SELECT price FROM users_price WHERE item_name = ?", itemName).Scan(&price)
	if err != nil {
		return 0, err
	}
	return price, nil
}

// HandlePriceQuery 处理价格查询逻辑
func HandlePriceQuery(w http.ResponseWriter, r *http.Request) error {
	logger.LogMessage(logger.INFO, "开始处理价格查询请求")

	if r.Method != http.MethodPost {
		logger.LogMessage(logger.WARN, "收到不支持的HTTP方法: %s", r.Method)
		return fmt.Errorf("不支持的HTTP方法: %s", r.Method)
	}

	// 从上下文中获取用户ID
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		logger.LogMessage(logger.ERROR, "未找到用户ID")
		return fmt.Errorf("未找到用户ID")
	}
	logger.LogMessage(logger.DEBUG, "处理用户ID: %d 的请求", userID)

	var requestData struct {
		ItemName string   `json:"item_name" form:"item_name"`
		Price    *float64 `json:"price" form:"price"`
		Overlays bool     `json:"overlays" form:"overlays"`
	}

	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.LogMessage(logger.ERROR, "读取请求体失败: %v", err)
			return fmt.Errorf("读取请求体失败: %v", err)
		}

		if err := json.Unmarshal(body, &requestData); err != nil {
			logger.LogMessage(logger.ERROR, "JSON解析失败: %v, 原始数据: %s", err, string(body))
			return fmt.Errorf("解析JSON请求体失败: %v, 请确保发送的是有效的JSON数据", err)
		}
		logger.LogMessage(logger.DEBUG, "成功解析JSON请求体")
	} else {
		if err := r.ParseForm(); err != nil {
			logger.LogMessage(logger.ERROR, "解析表单数据失败: %v", err)
			return fmt.Errorf("解析表单数据失败: %v", err)
		}

		requestData.ItemName = r.FormValue("item_name")
		if priceStr := r.FormValue("price"); priceStr != "" {
			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				logger.LogMessage(logger.ERROR, "解析价格失败: %v, 原始价格: %s", err, priceStr)
				return fmt.Errorf("解析价格失败: %v", err)
			}
			requestData.Price = &price
		}
		requestData.Overlays, _ = strconv.ParseBool(r.FormValue("overlays"))

	}

	logger.LogMessage(logger.DEBUG, "请求数据: 商品名=%s, 价格=%v, 是否覆盖=%v", requestData.ItemName, requestData.Price, requestData.Overlays)

	dbPrice, err := getProductPrice(database.DB, requestData.ItemName)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.LogMessage(logger.WARN, "未找到商品: %s", requestData.ItemName)
			return fmt.Errorf("未找到商品: %s", requestData.ItemName)
		}
		logger.LogMessage(logger.ERROR, "查询价格时出错: %v", err)
		return fmt.Errorf("查询价格时出错: %v", err)
	}
	logger.LogMessage(logger.DEBUG, "从数据库获取到的价格: %.2f", dbPrice)

	if requestData.Overlays && requestData.Price != nil {
		err = updateProductPrice(database.DB, requestData.ItemName, *requestData.Price)
		if err != nil {
			logger.LogMessage(logger.ERROR, "用户 %d 更新价格时出错: %v", userID, err)
			return fmt.Errorf("更新价格时出错: %v", err)
		}
		logger.LogMessage(logger.INFO, "用户 %d 将商品 %s 的价格从 %.2f 更新为 %.2f", userID, requestData.ItemName, dbPrice, *requestData.Price)
		dbPrice = *requestData.Price
	}

	response := map[string]interface{}{
		"item_name": requestData.ItemName,
		"price":     dbPrice,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		logger.LogMessage(logger.ERROR, "编码响应JSON时出错: %v", err)
		return fmt.Errorf("编码响应JSON时出错: %v", err)
	}

	logger.LogMessage(logger.INFO, "成功处理价格查询请求，返回价格: %.2f", dbPrice)
	return nil
}
