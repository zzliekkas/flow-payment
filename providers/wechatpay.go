package providers

import (
	"context"

	"github.com/zzliekkas/flow-payment"
)

type WeChatPayProvider struct {
	mchID    string
	apiKey   string
	certPath string
	keyPath  string
}

func NewWeChatPayProvider(mchID, apiKey, certPath, keyPath string) *WeChatPayProvider {
	return &WeChatPayProvider{mchID: mchID, apiKey: apiKey, certPath: certPath, keyPath: keyPath}
}

func (p *WeChatPayProvider) Name() string { return "wechatpay" }

func (p *WeChatPayProvider) CreatePayment(ctx context.Context, params payment.PaymentParams) (payment.PaymentResult, error) {
	return payment.PaymentResult{
		Provider:      p.Name(),
		PaymentURL:    "https://wechat.com/pay/mock-url",
		TransactionID: "mock-tx-id",
		Raw:           nil,
	}, nil
}

func (p *WeChatPayProvider) HandleCallback(ctx context.Context, request interface{}) (payment.CallbackResult, error) {
	return payment.CallbackResult{
		OrderID:       "mock-order-id",
		TransactionID: "mock-tx-id",
		Paid:          true,
		Amount:        100,
		Raw:           nil,
	}, nil
}

func (p *WeChatPayProvider) QueryStatus(ctx context.Context, orderID string) (payment.StatusResult, error) {
	return payment.StatusResult{
		OrderID:       orderID,
		TransactionID: "mock-tx-id",
		Status:        "paid",
		Raw:           nil,
	}, nil
}
