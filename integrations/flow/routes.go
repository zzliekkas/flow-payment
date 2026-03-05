package flowintegration

import (
	"net/http"

	payment "github.com/zzliekkas/flow-payment"
	"github.com/zzliekkas/flow/v2"
)

// RegisterFlowRoutes registers payment routes on a Flow engine.
func RegisterFlowRoutes(e *flow.Engine, manager *payment.PaymentManager) {
	paymentGroup := e.Group("/payment")
	{
		paymentGroup.POST("/create", func(c *flow.Context) {
			var req struct {
				Provider string  `json:"provider"`
				Amount   float64 `json:"amount"`
				Currency string  `json:"currency"`
				OrderID  string  `json:"order_id"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, flow.H{"error": "无效的请求参数"})
				return
			}

			provider := manager.GetProvider(req.Provider)
			if provider == nil {
				c.JSON(http.StatusBadRequest, flow.H{"error": "不支持的支付提供者"})
				return
			}

			result, err := provider.CreatePayment(&payment.PaymentRequest{
				Amount:   req.Amount,
				Currency: req.Currency,
				OrderID:  req.OrderID,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, flow.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, result)
		})

		paymentGroup.GET("/status/:provider/:order_id", func(c *flow.Context) {
			providerName := c.Param("provider")
			orderID := c.Param("order_id")

			provider := manager.GetProvider(providerName)
			if provider == nil {
				c.JSON(http.StatusBadRequest, flow.H{"error": "不支持的支付提供者"})
				return
			}

			status, err := provider.QueryPayment(orderID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, flow.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, status)
		})

		paymentGroup.POST("/notify/:provider", func(c *flow.Context) {
			providerName := c.Param("provider")

			provider := manager.GetProvider(providerName)
			if provider == nil {
				c.JSON(http.StatusBadRequest, flow.H{"error": "不支持的支付提供者"})
				return
			}

			if err := provider.HandleNotify(c.Request); err != nil {
				c.JSON(http.StatusInternalServerError, flow.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, flow.H{"message": "回调处理成功"})
		})

		paymentGroup.POST("/refund", func(c *flow.Context) {
			var req struct {
				Provider string  `json:"provider"`
				OrderID  string  `json:"order_id"`
				Amount   float64 `json:"amount"`
				Reason   string  `json:"reason"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, flow.H{"error": "无效的请求参数"})
				return
			}

			provider := manager.GetProvider(req.Provider)
			if provider == nil {
				c.JSON(http.StatusBadRequest, flow.H{"error": "不支持的支付提供者"})
				return
			}

			result, err := provider.Refund(&payment.RefundRequest{
				OrderID: req.OrderID,
				Amount:  req.Amount,
				Reason:  req.Reason,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, flow.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, result)
		})
	}
}
