package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"myproject/pkg/database"
	"net/http"
	"strconv"
	"time"
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
	itemName := r.URL.Query().Get("item_name")
	priceStr := r.URL.Query().Get("price")
	overlaysStr := r.URL.Query().Get("overlays")

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return fmt.Errorf("无效的价格: %v", err)
	}

	overlays := overlaysStr == "true"

	dbPrice, err := getProductPrice(database.DB, itemName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("%s [ERROR] 未找到商品: %s", time.Now().Format(time.RFC3339), itemName)
			return fmt.Errorf("未找到商品: %s", itemName)
		}
		log.Printf("%s [ERROR] 查询价格时出错: %v", time.Now().Format(time.RFC3339), err)
		return fmt.Errorf("查询价格时出错: %v", err)
	}

	if overlays {
		err = updateProductPrice(database.DB, itemName, price)
		if err != nil {
			log.Printf("%s [ERROR] 更新价格时出错: %v", time.Now().Format(time.RFC3339), err)
			return fmt.Errorf("更新价格时出错: %v", err)
		}
		log.Printf("%s [INFO] 商品 %s 的价格已更新为 %.2f", time.Now().Format(time.RFC3339), itemName, price)
		dbPrice = price
	}

	response := map[string]interface{}{
		"item_name": itemName,
		"price":     dbPrice,
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}
