package payment

import (
	"net/http"
)

// 支付请求参数
type PaymentParams struct {
	Amount      int64  // 金额，单位分
	Currency    string // 货币类型，如CNY、USD
	Description string // 订单描述
	OrderID     string // 业务订单号
	NotifyURL   string // 支付回调通知地址
	ReturnURL   string // 支付完成后跳转地址
	// 可扩展更多字段
}

// 支付结果
type PaymentResult struct {
	Provider      string      // 支付渠道
	PaymentURL    string      // 跳转支付页面的URL（如有）
	TransactionID string      // 支付平台订单号
	Raw           interface{} // 原始返回数据
}

// 回调处理结果
type CallbackResult struct {
	OrderID       string
	TransactionID string
	Paid          bool
	Amount        int64
	Raw           interface{}
}

// 订单状态查询结果
type StatusResult struct {
	OrderID       string
	TransactionID string
	Status        string // 如: pending, paid, failed
	Raw           interface{}
}

// PaymentRequest 支付请求参数
type PaymentRequest struct {
	Amount   float64           `json:"amount"`
	Currency string            `json:"currency"`
	OrderID  string            `json:"order_id"`
	Subject  string            `json:"subject,omitempty"`
	Body     string            `json:"body,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// PaymentResponse 支付响应结果
type PaymentResponse struct {
	TradeNo     string            `json:"trade_no"`
	OrderID     string            `json:"order_id"`
	PaymentURL  string            `json:"payment_url,omitempty"`
	QRCodeURL   string            `json:"qrcode_url,omitempty"`
	Status      string            `json:"status"`
	PaymentInfo map[string]string `json:"payment_info,omitempty"`
}

// PaymentStatus 支付状态
type PaymentStatus struct {
	TradeNo     string            `json:"trade_no"`
	OrderID     string            `json:"order_id"`
	Status      string            `json:"status"`
	PaidAmount  float64           `json:"paid_amount"`
	PaidTime    string            `json:"paid_time,omitempty"`
	PaymentInfo map[string]string `json:"payment_info,omitempty"`
}

// RefundRequest 退款请求参数
type RefundRequest struct {
	OrderID  string            `json:"order_id"`
	Amount   float64           `json:"amount"`
	Reason   string            `json:"reason,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// RefundResponse 退款响应结果
type RefundResponse struct {
	RefundID   string            `json:"refund_id"`
	OrderID    string            `json:"order_id"`
	Amount     float64           `json:"amount"`
	Status     string            `json:"status"`
	RefundTime string            `json:"refund_time,omitempty"`
	RefundInfo map[string]string `json:"refund_info,omitempty"`
}

// 支付渠道统一接口
type PaymentProvider interface {
	// Name 返回支付提供者名称
	Name() string

	// CreatePayment 创建支付
	CreatePayment(req *PaymentRequest) (*PaymentResponse, error)

	// QueryPayment 查询支付状态
	QueryPayment(orderID string) (*PaymentStatus, error)

	// HandleNotify 处理支付回调
	HandleNotify(request *http.Request) error

	// Refund 退款
	Refund(req *RefundRequest) (*RefundResponse, error)
}
