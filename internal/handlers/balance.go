package handlers

import (
	"myproject/internal/services"
	"net/http"
)

// BalanceHandler 处理用户余额修改请求
func BalanceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只允许POST方法", http.StatusMethodNotAllowed)
		return
	}

	// 调用服务层处理余额修改逻辑
	err := services.HandleBalanceUpdate(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
