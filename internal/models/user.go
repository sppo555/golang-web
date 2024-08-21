// package models

// import (
// 	"database/sql"

// 	"golang.org/x/crypto/bcrypt"
// )

// type User struct {
// 	ID       int
// 	Username string
// 	Password string
// 	Balance  float64
// }

// // GetUserByUsername 根据用户名获取用户
// func GetUserByUsername(db *sql.DB, username string) (*User, error) {
// 	var user User
// 	err := db.QueryRow("SELECT id, username, password FROM users WHERE username = ?", username).
// 		Scan(&user.ID, &user.Username, &user.Password)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &user, nil
// }

// // CheckPassword 验证密码
// func (u *User) CheckPassword(password string) bool {
// 	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
// 	return err == nil
// }

// // UpdateUserBalance 更新用户余额
// func UpdateUserBalance(db *sql.DB, userID string, newBalance float64) error {
// 	_, err := db.Exec("UPDATE user_balances SET balance = ? WHERE user_id = ?", newBalance, userID)
// 	return err
// }
