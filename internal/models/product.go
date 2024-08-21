// package models

// import "database/sql"

// type Product struct {
// 	ID    int
// 	Name  string
// 	Price float64
// }

// // GetProductPrice 根据商品ID获取价格
// func GetProductPrice(db *sql.DB, productID string) (float64, error) {
// 	var price float64
// 	err := db.QueryRow("SELECT price FROM products WHERE id = ?", productID).Scan(&price)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return price, nil
// }
