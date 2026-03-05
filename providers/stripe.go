package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/refund"
	"github.com/stripe/stripe-go/v74/webhook"

	"github.com/zzliekkas/flow-payment"
)

// StripeProvider Stripe支付提供者
type StripeProvider struct {
	apiKey        string
	webhookSecret string
}

// NewStripeProvider 创建Stripe支付提供者
func NewStripeProvider(apiKey, webhookSecret string) *StripeProvider {
	stripe.Key = apiKey
	return &StripeProvider{
		apiKey:        apiKey,
		webhookSecret: webhookSecret,
	}
}

// Name 实现PaymentProvider接口
func (p *StripeProvider) Name() string {
	return "stripe"
}

// CreatePayment 实现PaymentProvider接口
func (p *StripeProvider) CreatePayment(req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	// 创建 PaymentIntent
	params := &stripe.PaymentIntentParams{}
	params.AddMetadata("order_id", req.OrderID)

	if req.Metadata != nil {
		for k, v := range req.Metadata {
			params.AddMetadata(k, v)
		}
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("创建Stripe支付失败: %v", err)
	}

	return &payment.PaymentResponse{
		TradeNo:    pi.ID,
		OrderID:    req.OrderID,
		PaymentURL: pi.NextAction.RedirectToURL.URL, // 如果需要重定向
		Status:     string(pi.Status),
		PaymentInfo: map[string]string{
			"client_secret": pi.ClientSecret,
		},
	}, nil
}

// QueryPayment 实现PaymentProvider接口
func (p *StripeProvider) QueryPayment(orderID string) (*payment.PaymentStatus, error) {
	// 通过元数据查询 PaymentIntent
	params := &stripe.PaymentIntentListParams{}
	params.Filters.AddFilter("metadata[order_id]", "", orderID)

	iter := paymentintent.List(params)
	for iter.Next() {
		pi := iter.PaymentIntent()
		return &payment.PaymentStatus{
			TradeNo:    pi.ID,
			OrderID:    orderID,
			Status:     string(pi.Status),
			PaidAmount: float64(pi.Amount) / 100, // 转换为元
			PaidTime:   time.Unix(pi.Created, 0).Format(time.RFC3339),
			PaymentInfo: map[string]string{
				"payment_method": string(pi.PaymentMethodTypes[0]),
			},
		}, nil
	}

	return nil, fmt.Errorf("未找到订单: %s", orderID)
}

// HandleNotify 实现PaymentProvider接口
func (p *StripeProvider) HandleNotify(request *http.Request) error {
	payload, err := io.ReadAll(request.Body)
	if err != nil {
		return fmt.Errorf("读取请求体失败: %v", err)
	}

	event, err := webhook.ConstructEvent(payload, request.Header.Get("Stripe-Signature"), p.webhookSecret)
	if err != nil {
		return fmt.Errorf("验证Webhook签名失败: %v", err)
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			return fmt.Errorf("解析PaymentIntent失败: %v", err)
		}
		// 处理支付成功逻辑
		// TODO: 更新订单状态等

	case "payment_intent.payment_failed":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			return fmt.Errorf("解析PaymentIntent失败: %v", err)
		}
		// 处理支付失败逻辑
		// TODO: 更新订单状态等
	}

	return nil
}

// Refund 实现PaymentProvider接口
func (p *StripeProvider) Refund(req *payment.RefundRequest) (*payment.RefundResponse, error) {
	// 通过订单ID查找 PaymentIntent
	params := &stripe.PaymentIntentListParams{}
	params.Filters.AddFilter("metadata[order_id]", "", req.OrderID)

	iter := paymentintent.List(params)
	var pi *stripe.PaymentIntent
	for iter.Next() {
		pi = iter.PaymentIntent()
		break
	}
	if pi == nil {
		return nil, fmt.Errorf("未找到订单: %s", req.OrderID)
	}

	// 创建退款
	refundParams := &stripe.RefundParams{
		PaymentIntent: stripe.String(pi.ID),
		Amount:        stripe.Int64(int64(req.Amount * 100)), // 转换为分
	}
	refundParams.AddMetadata("order_id", req.OrderID)
	refundParams.AddMetadata("reason", req.Reason)
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			refundParams.AddMetadata(k, v)
		}
	}

	r, err := refund.New(refundParams)
	if err != nil {
		return nil, fmt.Errorf("创建退款失败: %v", err)
	}

	return &payment.RefundResponse{
		RefundID:   r.ID,
		OrderID:    req.OrderID,
		Amount:     float64(r.Amount) / 100, // 转换为元
		Status:     string(r.Status),
		RefundTime: time.Unix(r.Created, 0).Format(time.RFC3339),
		RefundInfo: map[string]string{
			"receipt_number": r.ReceiptNumber,
		},
	}, nil
}

func (p *StripeProvider) QueryStatus(ctx context.Context, orderID string) (payment.StatusResult, error) {
	if p.apiKey == "" {
		return payment.StatusResult{}, errors.New("Stripe API Key 未配置")
	}
	stripe.Key = p.apiKey
	// 这里只做演示，实际应根据 orderID 查询 payment_intent
	return payment.StatusResult{
		OrderID:       orderID,
		TransactionID: "",
		Status:        "unknown",
		Raw:           nil,
	}, nil
}
