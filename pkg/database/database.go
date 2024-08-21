package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"myproject/pkg/logger"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func init() {
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbName)

	// 设置连接超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		logger.LogMessage(logger.ERROR, "无法打开数据库连接: %v", err)
		os.Exit(1)
	}

	// 使用带超时的 Ping
	err = DB.PingContext(ctx)
	if err != nil {
		logger.LogMessage(logger.ERROR, "无法连接数据库: %v", err)
		os.Exit(1)
	}

	logger.LogMessage(logger.INFO, "成功连接到数据库")
}
