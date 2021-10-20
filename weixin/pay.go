package weixin

import (
	"context"

	"gitee.com/Rainkropy/frm-lib/http"
	"gitee.com/Rainkropy/frm-lib/log"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

type PayConf struct {

	// 商户号
	MchID string

	// 证书序列号
	MchCertificateSerialNumber string

	// V3秘钥
	MchAPIv3Key string

	// 秘钥证书路径
	MchKeyPemPath string

	// 请求client
	Client *core.Client
}

// 初始化
func (p *PayConf) Init() error {
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(p.MchKeyPemPath)
	if err != nil {
		log.Error().Str("lib", "weixin").Str("method", "PayConf.GetClient.LoadPrivateKeyWithPath").Err(err).Send()
		return err
	}

	ctx := context.Background()
	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(p.MchID, p.MchCertificateSerialNumber, mchPrivateKey, p.MchAPIv3Key),
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		log.Error().Str("lib", "weixin").Str("method", "PayConf.GetClient.NewClient").Err(err).Send()
		return err
	}
	p.Client = client
	return nil
}

// JSAPI 下单
func (p *PayConf) Jsapi(appID, desc, tradeNo, attach, notifyURL, openID string, amount int64) (*jsapi.PrepayWithRequestPaymentResponse, error) {
	svc := jsapi.JsapiApiService{Client: p.Client}
	// 得到prepay_id，以及调起支付所需的参数和签名
	resp, _, err := svc.PrepayWithRequestPayment(context.Background(),
		jsapi.PrepayRequest{
			Appid:       core.String(appID),
			Mchid:       core.String(p.MchID),
			Description: core.String(desc),
			OutTradeNo:  core.String(tradeNo),
			Attach:      core.String(attach),
			NotifyUrl:   core.String(notifyURL),
			Amount: &jsapi.Amount{
				Total: core.Int64(amount),
			},
			Payer: &jsapi.Payer{
				Openid: core.String(openID),
			},
		},
	)

	if err == nil {
		return resp, nil
	} else {
		return nil, err
	}
}

// 根据订单号查询支付信息
func (p *PayConf) QueryOrderByOutTradeNo(outTradeNo string) (*payments.Transaction, error) {
	svc := jsapi.JsapiApiService{Client: p.Client}
	resp, _, err := svc.QueryOrderByOutTradeNo(context.Background(),
		jsapi.QueryOrderByOutTradeNoRequest{
			OutTradeNo: core.String(outTradeNo),
			Mchid:      core.String(p.MchID),
		},
	)

	if err != nil {
		log.Error().Str("lib", "weixin").Str("method", "PayConf.QueryOrderByOutTradeNo.QueryOrderByOutTradeNo").Str("outTradeNo", outTradeNo).Err(err).Send()
		return nil, err
	} else {
		return resp, nil
	}
}

// 验证支付回调通知
func (p *PayConf) VerifyPayNotify(app *http.App) (*payments.Transaction, error) {

	certVisitor := downloader.MgrInstance().GetCertificateVisitor(p.MchID)
	handler := notify.NewNotifyHandler(p.MchAPIv3Key, verifiers.NewSHA256WithRSAVerifier(certVisitor))

	transaction := new(payments.Transaction)
	_, err := handler.ParseNotifyRequest(context.Background(), app.Request.Ctx.Request, transaction)
	// 如果验签未通过，或者解密失败
	if err != nil {
		log.Error().Str("lib", "weixin").Str("method", "PayConf.VerifyPayNotify.ParseNotifyRequest").Bytes("body", app.GetBody()).Err(err).Send()
		return nil, err
	}
	return transaction, nil
}

// 退款
func (p *PayConf) CreateRefund(outTradeNo, reason string, amount int64) (*refunddomestic.Refund, error) {
	svc := refunddomestic.RefundsApiService{Client: p.Client}
	resp, _, err := svc.Create(context.Background(),
		refunddomestic.CreateRequest{
			OutTradeNo:   core.String(outTradeNo),
			OutRefundNo:  core.String(outTradeNo),
			Reason:       core.String(reason),
			FundsAccount: refunddomestic.REQFUNDSACCOUNT_AVAILABLE.Ptr(),
			Amount: &refunddomestic.AmountReq{
				Currency: core.String("CNY"),
				Refund:   core.Int64(amount),
				Total:    core.Int64(amount),
			},
		},
	)

	if err != nil {
		log.Error().Str("lib", "weixin").Str("method", "PayConf.CreateRefund.Create").Str("outTradeNo", outTradeNo).Err(err).Send()
		return nil, err
	}
	return resp, nil
}

// 查询退款详情
func (p *PayConf) QueryByOutRefundNo(outTradeNo string) (*refunddomestic.Refund, error) {
	svc := refunddomestic.RefundsApiService{Client: p.Client}
	resp, _, err := svc.QueryByOutRefundNo(context.Background(),
		refunddomestic.QueryByOutRefundNoRequest{
			OutRefundNo: core.String(outTradeNo),
		},
	)
	if err != nil {
		log.Error().Str("lib", "weixin").Str("method", "PayConf.QueryByOutRefundNo").Str("outTradeNo", outTradeNo).Err(err).Send()
		return nil, err
	}
	return resp, nil
}
