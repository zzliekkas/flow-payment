package payment

import (
	"net/http"
)

// RegisterRoutes 注册支付相关路由
func RegisterRoutes(mux *http.ServeMux, manager *PaymentManager) {
	// Stripe 下单
	mux.HandleFunc("/pay/stripe", func(w http.ResponseWriter, r *http.Request) {
		// 这里只做演示，实际应解析请求参数
		provider := manager.Get("stripe")
		if provider == nil {
			http.Error(w, "stripe provider not found", http.StatusNotFound)
			return
		}
		// 示例参数
		req := &PaymentRequest{
			Amount:   100,
			Currency: "USD",
			OrderID:  "order123",
			Subject:  "Test Order",
			// 其他字段按需补充
		}
		result, err := provider.CreatePayment(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(result.PaymentURL)) // 实际应返回 JSON
	})

	// Stripe 回调
	mux.HandleFunc("/pay/callback/stripe", func(w http.ResponseWriter, r *http.Request) {
		provider := manager.Get("stripe")
		if provider == nil {
			http.Error(w, "stripe provider not found", http.StatusNotFound)
			return
		}
		// 这里只做演示，实际应解析 Stripe webhook
		if err := provider.HandleNotify(r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
}
