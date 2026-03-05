package payment

import (
	"net/http"

	"github.com/zzliekkas/flow/v2"
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

// RegisterFlowRoutes 注册支付相关路由到 Flow 引擎
func RegisterFlowRoutes(e *flow.Engine, manager *PaymentManager) {
	// 支付相关路由
	paymentGroup := e.Group("/payment")
	{
		// 创建支付
		paymentGroup.POST("/create", func(c *flow.Context) {
			var req struct {
				Provider string  `json:"provider"`
				Amount   float64 `json:"amount"`
				Currency string  `json:"currency"`
				OrderID  string  `json:"order_id"`
			}

			if err := c.BindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, flow.H{
					"error": "无效的请求参数",
				})
				return
			}

			provider := manager.GetProvider(req.Provider)
			if provider == nil {
				c.JSON(http.StatusBadRequest, flow.H{
					"error": "不支持的支付提供者",
				})
				return
			}

			// 创建支付
			result, err := provider.CreatePayment(&PaymentRequest{
				Amount:   req.Amount,
				Currency: req.Currency,
				OrderID:  req.OrderID,
			})

			if err != nil {
				c.JSON(http.StatusInternalServerError, flow.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, result)
		})

		// 查询支付状态
		paymentGroup.GET("/status/:provider/:order_id", func(c *flow.Context) {
			providerName := c.Param("provider")
			orderID := c.Param("order_id")

			provider := manager.GetProvider(providerName)
			if provider == nil {
				c.JSON(http.StatusBadRequest, flow.H{
					"error": "不支持的支付提供者",
				})
				return
			}

			status, err := provider.QueryPayment(orderID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, flow.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, status)
		})

		// 支付回调处理
		paymentGroup.POST("/notify/:provider", func(c *flow.Context) {
			providerName := c.Param("provider")

			provider := manager.GetProvider(providerName)
			if provider == nil {
				c.JSON(http.StatusBadRequest, flow.H{
					"error": "不支持的支付提供者",
				})
				return
			}

			// 处理支付回调
			if err := provider.HandleNotify(c.Request); err != nil {
				c.JSON(http.StatusInternalServerError, flow.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, flow.H{
				"message": "回调处理成功",
			})
		})

		// 退款
		paymentGroup.POST("/refund", func(c *flow.Context) {
			var req struct {
				Provider string  `json:"provider"`
				OrderID  string  `json:"order_id"`
				Amount   float64 `json:"amount"`
				Reason   string  `json:"reason"`
			}

			if err := c.BindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, flow.H{
					"error": "无效的请求参数",
				})
				return
			}

			provider := manager.GetProvider(req.Provider)
			if provider == nil {
				c.JSON(http.StatusBadRequest, flow.H{
					"error": "不支持的支付提供者",
				})
				return
			}

			result, err := provider.Refund(&RefundRequest{
				OrderID: req.OrderID,
				Amount:  req.Amount,
				Reason:  req.Reason,
			})

			if err != nil {
				c.JSON(http.StatusInternalServerError, flow.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, result)
		})
	}
}
