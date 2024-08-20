package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"myproject/pkg/database"
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

	if r.Method != http.MethodPost {
		return fmt.Errorf("不支持的HTTP方法: %s", r.Method)
	}

	// 从上下文中获取用户ID
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		return fmt.Errorf("未找到用户ID")
	}

	var requestData struct {
		ItemName string   `json:"item_name" form:"item_name"`
		Price    *float64 `json:"price" form:"price"`
		Overlays bool     `json:"overlays" form:"overlays"`
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("读取请求体失败: %v", err)
		}

		if err := json.Unmarshal(body, &requestData); err != nil {
			log.Printf("[ERROR] JSON解析失败: %v, 原始数据: %s", err, string(body))
			return fmt.Errorf("解析JSON请求体失败: %v, 请确保发送的是有效的JSON数据", err)
		}
	} else {
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("解析表单数据失败: %v", err)
		}

		requestData.ItemName = r.FormValue("item_name")
		if priceStr := r.FormValue("price"); priceStr != "" {
			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				return fmt.Errorf("解析价格失败: %v", err)
			}
			requestData.Price = &price
		}
		requestData.Overlays, _ = strconv.ParseBool(r.FormValue("overlays"))
	}

	dbPrice, err := getProductPrice(database.DB, requestData.ItemName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[ERROR] 未找到商品: %s", requestData.ItemName)
			return fmt.Errorf("未找到商品: %s", requestData.ItemName)
		}
		log.Printf("[ERROR] 查询价格时出错: %v", err)
		return fmt.Errorf("查询价格时出错: %v", err)
	}

	if requestData.Overlays && requestData.Price != nil {
		err = updateProductPrice(database.DB, requestData.ItemName, *requestData.Price)
		if err != nil {
			log.Printf("[ERROR] 用户 %d 更新价格时出错: %v", userID, err)
			return fmt.Errorf("更新价格时出错: %v", err)
		}
		log.Printf("[INFO] 用户 %d 将商品 %s 的价格更新为 %.2f", userID, requestData.ItemName, *requestData.Price)
		dbPrice = *requestData.Price
	}

	response := map[string]interface{}{
		"item_name": requestData.ItemName,
		"price":     dbPrice,
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

func init() {
	// 禁用默认时间戳
	log.SetFlags(0)
}
