package handlers

import (
	"myproject/internal/services"
	"net/http"
)

// LoginHandler 处理用户登录请求
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只允许POST方法", http.StatusMethodNotAllowed)
		return
	}

	// 调用服务层处理登录逻辑
	err := services.HandleLogin(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}
}
