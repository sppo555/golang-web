
package handlers

import (
    "net/http"
    "myproject/internal/services"
)

// PriceHandler 处理价格查询请求
func PriceHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
        return
    }

    // 调用服务层处理价格查询逻辑
    err := services.HandlePriceQuery(w, r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
