package providers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/smartwalle/alipay/v3"
	"github.com/zzliekkas/flow-payment"
)

// AlipayProvider 支付宝支付提供者
type AlipayProvider struct {
	client     *alipay.Client
	appID      string
	privateKey string
	publicKey  string
}

// NewAlipayProvider 创建支付宝支付提供者
func NewAlipayProvider(appID, privateKey, publicKey string) *AlipayProvider {
	client, err := alipay.New(appID, privateKey, false)
	if err != nil {
		panic(fmt.Sprintf("初始化支付宝客户端失败: %v", err))
	}

	// 加载支付宝公钥
	err = client.LoadAliPayPublicKey(publicKey)
	if err != nil {
		panic(fmt.Sprintf("加载支付宝公钥失败: %v", err))
	}

	return &AlipayProvider{
		client:     client,
		appID:      appID,
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

// Name 实现PaymentProvider接口
func (p *AlipayProvider) Name() string {
	return "alipay"
}

// CreatePayment 实现PaymentProvider接口
func (p *AlipayProvider) CreatePayment(req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	// 创建支付宝订单
	var pay = alipay.TradePagePay{}
	pay.NotifyURL = "http://xxx.com/notify/alipay"
	pay.ReturnURL = "http://xxx.com/return/alipay"
	pay.Subject = req.Subject
	pay.OutTradeNo = req.OrderID
	pay.TotalAmount = fmt.Sprintf("%.2f", req.Amount)
	pay.ProductCode = "FAST_INSTANT_TRADE_PAY"

	url, err := p.client.TradePagePay(pay)
	if err != nil {
		return nil, fmt.Errorf("创建支付宝订单失败: %v", err)
	}

	return &payment.PaymentResponse{
		TradeNo:    pay.OutTradeNo,
		OrderID:    req.OrderID,
		PaymentURL: url.String(),
		Status:     "created",
		PaymentInfo: map[string]string{
			"qr_code": url.String(), // 支付宝同步返回支付URL
		},
	}, nil
}

// QueryPayment 实现PaymentProvider接口
func (p *AlipayProvider) QueryPayment(orderID string) (*payment.PaymentStatus, error) {
	var query = alipay.TradeQuery{}
	query.OutTradeNo = orderID

	result, err := p.client.TradeQuery(query)
	if err != nil {
		return nil, fmt.Errorf("查询支付宝订单失败: %v", err)
	}

	if !result.IsSuccess() {
		return nil, fmt.Errorf("查询支付宝订单失败: %s", result.SubMsg)
	}

	amount := 0.0
	if result.TotalAmount != "" {
		fmt.Sscanf(result.TotalAmount, "%f", &amount)
	}

	return &payment.PaymentStatus{
		TradeNo:    result.TradeNo,
		OrderID:    result.OutTradeNo,
		Status:     string(result.TradeStatus),
		PaidAmount: amount,
		PaidTime:   result.SendPayDate,
		PaymentInfo: map[string]string{
			"buyer_id":  result.BuyerUserId,
			"seller_id": "",
			"trade_no":  result.TradeNo,
			"pay_method": func() string {
				if len(result.FundBillList) > 0 {
					return result.FundBillList[0].FundChannel
				}
				return ""
			}(),
		},
	}, nil
}

// HandleNotify 实现PaymentProvider接口
func (p *AlipayProvider) HandleNotify(request *http.Request) error {
	notification, err := p.client.GetTradeNotification(request)
	if err != nil {
		return fmt.Errorf("解析支付宝回调通知失败: %v", err)
	}

	if notification.TradeStatus == "TRADE_SUCCESS" {
		// 处理支付成功逻辑
		// TODO: 更新订单状态
	} else if notification.TradeStatus == "TRADE_CLOSED" {
		// 处理支付关闭逻辑
		// TODO: 更新订单状态
	}

	return nil
}

// Refund 实现PaymentProvider接口
func (p *AlipayProvider) Refund(req *payment.RefundRequest) (*payment.RefundResponse, error) {
	var refund = alipay.TradeRefund{}
	refund.OutTradeNo = req.OrderID
	refund.RefundAmount = fmt.Sprintf("%.2f", req.Amount)
	refund.RefundReason = req.Reason

	result, err := p.client.TradeRefund(refund)
	if err != nil {
		return nil, fmt.Errorf("创建支付宝退款失败: %v", err)
	}

	if !result.IsSuccess() {
		return nil, fmt.Errorf("创建支付宝退款失败: %s", result.SubMsg)
	}

	amount := 0.0
	if result.RefundFee != "" {
		fmt.Sscanf(result.RefundFee, "%f", &amount)
	}

	return &payment.RefundResponse{
		RefundID:   result.TradeNo,
		OrderID:    req.OrderID,
		Amount:     amount,
		Status:     "success",
		RefundTime: time.Now().Format(time.RFC3339),
		RefundInfo: map[string]string{
			"buyer_id":     result.BuyerUserId,
			"refund_fee":   result.RefundFee,
			"trade_no":     result.TradeNo,
			"out_trade_no": result.OutTradeNo,
		},
	}, nil
}
