package providers

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"

	notify "github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	jsapi "github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/zzliekkas/flow-payment"
)

// WechatPayProvider 微信支付提供者
type WechatPayProvider struct {
	client   *core.Client
	mchID    string
	apiKey   string
	certPath string
	keyPath  string
}

// NewWechatPayProvider 创建微信支付提供者
func NewWechatPayProvider(mchID, apiKey, certPath, keyPath string) *WechatPayProvider {
	ctx := context.Background()

	// 读取私钥
	keyPem, err := ioutil.ReadFile(keyPath)
	if err != nil {
		panic(fmt.Sprintf("读取微信支付私钥失败: %v", err))
	}
	block, _ := pem.Decode(keyPem)
	if block == nil {
		panic("解析微信支付私钥PEM失败")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(fmt.Sprintf("解析微信支付私钥失败: %v", err))
	}

	// 证书序列号（假设你有 certSerialNo 变量，实际应从证书文件读取）
	certSerialNo := "YOUR_CERT_SERIAL_NO"

	opts := []core.ClientOption{
		option.WithMerchantCredential(mchID, certSerialNo, privateKey),
		option.WithoutValidator(),
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		panic(fmt.Sprintf("初始化微信支付客户端失败: %v", err))
	}

	return &WechatPayProvider{
		client:   client,
		mchID:    mchID,
		apiKey:   apiKey,
		certPath: certPath,
		keyPath:  keyPath,
	}
}

// Name 实现PaymentProvider接口
func (p *WechatPayProvider) Name() string {
	return "wechat"
}

// CreatePayment 实现PaymentProvider接口
func (p *WechatPayProvider) CreatePayment(req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	svc := jsapi.JsapiApiService{Client: p.client}
	ctx := context.Background()

	// 创建统一下单
	resp, result, err := svc.PrepayWithRequestPayment(ctx,
		jsapi.PrepayRequest{
			Appid:       core.String("your_appid"),
			Mchid:       core.String(p.mchID),
			Description: core.String(req.Subject),
			OutTradeNo:  core.String(req.OrderID),
			NotifyUrl:   core.String("http://xxx.com/notify/wechat"),
			Amount: &jsapi.Amount{
				Total:    core.Int64(int64(req.Amount * 100)), // 转换为分
				Currency: core.String("CNY"),
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("创建微信支付订单失败: %v", err)
	}

	if result.Response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("创建微信支付订单失败: HTTP %d", result.Response.StatusCode)
	}

	return &payment.PaymentResponse{
		TradeNo:    *resp.PrepayId,
		OrderID:    req.OrderID,
		PaymentURL: "", // 微信支付不需要跳转URL
		Status:     "created",
		PaymentInfo: map[string]string{
			"prepay_id": *resp.PrepayId,
			"app_id":    "your_appid",
			"mch_id":    p.mchID,
		},
	}, nil
}

// QueryPayment 实现PaymentProvider接口
func (p *WechatPayProvider) QueryPayment(orderID string) (*payment.PaymentStatus, error) {
	svc := jsapi.JsapiApiService{Client: p.client}
	ctx := context.Background()

	resp, result, err := svc.QueryOrderByOutTradeNo(ctx,
		jsapi.QueryOrderByOutTradeNoRequest{
			OutTradeNo: core.String(orderID),
			Mchid:      core.String(p.mchID),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("查询微信支付订单失败: %v", err)
	}

	if result.Response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("查询微信支付订单失败: HTTP %d", result.Response.StatusCode)
	}

	var paidTime string
	if resp.SuccessTime != nil {
		paidTime = *resp.SuccessTime
	}
	return &payment.PaymentStatus{
		TradeNo:    *resp.TransactionId,
		OrderID:    orderID,
		Status:     *resp.TradeState,
		PaidAmount: float64(*resp.Amount.Total) / 100, // 转换为元
		PaidTime:   paidTime,
		PaymentInfo: map[string]string{
			"trade_state_desc": *resp.TradeStateDesc,
			"payer":            *resp.Payer.Openid,
			"trade_type":       *resp.TradeType,
		},
	}, nil
}

// HandleNotify 实现PaymentProvider接口
func (p *WechatPayProvider) HandleNotify(request *http.Request) error {
	ctx := context.Background()
	handler := notify.NewNotifyHandler(p.apiKey, nil)
	var transaction map[string]interface{}
	notifyReq, err := handler.ParseNotifyRequest(ctx, request, &transaction)
	if err != nil {
		return fmt.Errorf("解析微信支付回调通知失败: %v", err)
	}
	_ = notifyReq
	return nil
}

// Refund 实现PaymentProvider接口
func (p *WechatPayProvider) Refund(req *payment.RefundRequest) (*payment.RefundResponse, error) {
	svc := refunddomestic.RefundsApiService{Client: p.client}
	ctx := context.Background()

	resp, result, err := svc.Create(ctx,
		refunddomestic.CreateRequest{
			OutTradeNo:  core.String(req.OrderID),
			OutRefundNo: core.String(fmt.Sprintf("refund_%s_%d", req.OrderID, time.Now().Unix())),
			Reason:      core.String(req.Reason),
			Amount: &refunddomestic.AmountReq{
				Refund:   core.Int64(int64(req.Amount * 100)), // 转换为分
				Total:    core.Int64(int64(req.Amount * 100)), // 应与订单金额一致
				Currency: core.String("CNY"),
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("创建微信支付退款失败: %v", err)
	}

	if result.Response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("创建微信支付退款失败: HTTP %d", result.Response.StatusCode)
	}

	return &payment.RefundResponse{
		RefundID:   *resp.RefundId,
		OrderID:    req.OrderID,
		Amount:     float64(*resp.Amount.Refund) / 100, // 转换为元
		Status:     string(*resp.Status),
		RefundTime: resp.CreateTime.Format(time.RFC3339),
		RefundInfo: map[string]string{
			"out_refund_no":         *resp.OutRefundNo,
			"channel":               string(*resp.Channel),
			"user_received_account": *resp.UserReceivedAccount,
		},
	}, nil
}
