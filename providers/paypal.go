package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/plutov/paypal/v4"
	"github.com/zzliekkas/flow-payment"
)

// PayPalProvider PayPal支付提供者
type PayPalProvider struct {
	client       *paypal.Client
	clientID     string
	clientSecret string
}

// NewPayPalProvider 创建PayPal支付提供者
func NewPayPalProvider(clientID, clientSecret string) *PayPalProvider {
	client, err := paypal.NewClient(clientID, clientSecret, paypal.APIBaseSandBox)
	if err != nil {
		panic(fmt.Sprintf("初始化PayPal客户端失败: %v", err))
	}
	return &PayPalProvider{
		client:       client,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

// Name 实现PaymentProvider接口
func (p *PayPalProvider) Name() string {
	return "paypal"
}

// CreatePayment 实现PaymentProvider接口
func (p *PayPalProvider) CreatePayment(req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	order, err := p.client.CreateOrder(context.Background(), paypal.OrderIntentCapture,
		[]paypal.PurchaseUnitRequest{
			{
				ReferenceID: req.OrderID,
				Amount: &paypal.PurchaseUnitAmount{
					Value:    fmt.Sprintf("%.2f", req.Amount),
					Currency: req.Currency,
				},
			},
		}, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("创建PayPal订单失败: %v", err)
	}

	var approveURL string
	for _, link := range order.Links {
		if link.Rel == "approve" {
			approveURL = link.Href
			break
		}
	}

	payerID := ""
	if order.Payer != nil {
		payerID = order.Payer.PayerID
	}

	return &payment.PaymentResponse{
		TradeNo:    order.ID,
		OrderID:    req.OrderID,
		PaymentURL: approveURL,
		Status:     string(order.Status),
		PaymentInfo: map[string]string{
			"payer_id": payerID,
		},
	}, nil
}

// QueryPayment 实现PaymentProvider接口
func (p *PayPalProvider) QueryPayment(orderID string) (*payment.PaymentStatus, error) {
	order, err := p.client.GetOrder(context.Background(), orderID)
	if err != nil {
		return nil, fmt.Errorf("查询PayPal订单失败: %v", err)
	}

	amount, _ := strconv.ParseFloat(order.PurchaseUnits[0].Amount.Value, 64)
	payerID := ""
	if order.Payer != nil {
		payerID = order.Payer.PayerID
	}

	return &payment.PaymentStatus{
		TradeNo:    order.ID,
		OrderID:    order.PurchaseUnits[0].ReferenceID,
		Status:     string(order.Status),
		PaidAmount: amount,
		PaidTime:   order.CreateTime.Format(time.RFC3339),
		PaymentInfo: map[string]string{
			"payer_id": payerID,
		},
	}, nil
}

// HandleNotify 实现PaymentProvider接口
func (p *PayPalProvider) HandleNotify(request *http.Request) error {
	webhookID := request.Header.Get("Paypal-Transmission-Id")
	if webhookID == "" {
		return fmt.Errorf("缺少PayPal Webhook ID")
	}

	var event map[string]interface{}
	if err := json.NewDecoder(request.Body).Decode(&event); err != nil {
		return fmt.Errorf("解析PayPal Webhook事件失败: %v", err)
	}
	eventType, _ := event["event_type"].(string)
	switch eventType {
	case "PAYMENT.CAPTURE.COMPLETED":
		// 处理支付成功
	case "PAYMENT.CAPTURE.DENIED":
		// 处理支付失败
	}
	return nil
}

// Refund 实现PaymentProvider接口
func (p *PayPalProvider) Refund(req *payment.RefundRequest) (*payment.RefundResponse, error) {
	order, err := p.client.GetOrder(context.Background(), req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("查询PayPal订单失败: %v", err)
	}

	var captureID string
	if len(order.PurchaseUnits) > 0 &&
		order.PurchaseUnits[0].Payments != nil &&
		len(order.PurchaseUnits[0].Payments.Captures) > 0 {
		captureID = order.PurchaseUnits[0].Payments.Captures[0].ID
	}
	if captureID == "" {
		return nil, fmt.Errorf("未找到支付捕获ID")
	}

	refund, err := p.client.RefundCapture(context.Background(), captureID, paypal.RefundCaptureRequest{
		Amount: &paypal.Money{
			Value:    fmt.Sprintf("%.2f", req.Amount),
			Currency: order.PurchaseUnits[0].Amount.Currency,
		},
		NoteToPayer: req.Reason,
	})
	if err != nil {
		return nil, fmt.Errorf("创建PayPal退款失败: %v", err)
	}

	amount, _ := strconv.ParseFloat(refund.Amount.Value, 64)

	return &payment.RefundResponse{
		RefundID:   refund.ID,
		OrderID:    req.OrderID,
		Amount:     amount,
		Status:     string(refund.Status),
		RefundTime: "",
		RefundInfo: map[string]string{},
	}, nil
}
