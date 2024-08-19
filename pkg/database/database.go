package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func init() {
	log.SetFlags(0)

	username := os.Getenv("TEST_DB_USERNAME")
	password := os.Getenv("TEST_DB_PASSWORD")
	dsn := fmt.Sprintf("%s:%s@tcp(34.97.41.216:3306)/test", username, password)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("%s 无法连接到数据库: %v", time.Now().Format(time.RFC3339), err)
		os.Exit(1)
	}

	if err = DB.Ping(); err != nil {
		log.Printf("%s 无法ping数据库: %v", time.Now().Format(time.RFC3339), err)
		// os.Exit(1)
	}

	log.Printf("%s 成功连接到数据库", time.Now().Format(time.RFC3339))
}
