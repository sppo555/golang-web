package logger

import (
	"fmt"
	"os"
	"time"
)

// LogLevel 定義日誌級別
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var logLevelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

var currentLogLevel = func() LogLevel {
	level := os.Getenv("LOG_LEVEL")
	for l, name := range logLevelNames {
		if name == level {
			return l
		}
	}
	return INFO // 默認日誌級別
}()

// SetLogLevel 設置當前日誌級別
func SetLogLevel(level LogLevel) {
	currentLogLevel = level
}

// LogMessage 根據當前日誌級別輸出日誌
func LogMessage(level LogLevel, format string, v ...interface{}) {
	if level >= currentLogLevel {
		timestamp := time.Now().Format(time.RFC3339)
		levelName := logLevelNames[level]
		fmt.Printf("%s [%s] %s\n", timestamp, levelName, fmt.Sprintf(format, v...))
	}
}
