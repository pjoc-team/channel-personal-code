package sdk

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pjoc-team/base-service/pkg/logger"
	"github.com/pjoc-team/base-service/pkg/sign"
	"github.com/pjoc-team/base-service/pkg/util"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	CODE_SUCCESS int = 200
	// Params error start--------------------------------
	// Params validate error
	CODE_PARAMS_ERROR int = 100
	// Check sign error
	CODE_CHECK_SIGN_ERROR int = 101
	NO_AVAILABLE_DEVICE   int = 302
	// --------------------------------
	CODE_SYSTEM_ERROR int = 999
)

type PayRequest struct {
	Version         string `form:"version" json:"version,omitempty"`
	OutTradeNo      string `form:"out_trade_no" json:"out_trade_no,omitempty"  binding:"required"` // 业务订单号
	PayAmount       string `form:"pay_amount" json:"pay_amount,omitempty"  binding:"required"`     // 支付金额（分）
	Currency        string `form:"currency" json:"currency,omitempty"`                             // 币种
	NotifyUrl       string `form:"notify_url" json:"notify_url,omitempty" binding:"required"`      // 接收通知的地址，不能带参数（即：不能包含问号）
	ReturnUrl       string `form:"return_url" json:"return_url,omitempty"`                         // 支付后跳转的前端地址
	AppId           string `form:"app_id" json:"app_id,omitempty" binding:"required"`              // 系统给商户分配的app_id
	SignType        string `form:"sign_type" json:"sign_type,omitempty"`                           // 加密方法，RSA和MD5，默认RSA
	Sign            string `form:"sign" json:"sign,omitempty"  binding:"required"`                 // 签名
	OrderTime       string `form:"order_time" json:"order_time,omitempty"  binding:"required"`     // 业务方下单时间，时间格式: 年年年年-月月-日日 时时:分分:秒秒，例如: 2006-01-02 15:04:05
	UserIp          string `form:"user_ip" json:"user_ip,omitempty"`                               // 发起支付的用户ip
	UserId          string `form:"user_id" json:"user_id,omitempty"`                               // 用户在业务系统的id
	PayerAccount    string `form:"payer_account" json:"payer_account,omitempty"`                   // 支付者账号，可选
	ProductId       string `form:"product_id" json:"product_id,omitempty"`                         // 业务系统的产品id
	ProductName     string `form:"product_name" json:"product_name,omitempty"`                     // 商品名称
	ProductDescribe string `form:"product_describe" json:"product_describe,omitempty"`             // 商品描述
	Charset         string `form:"charset" json:"charset,omitempty"`                               // 参数编码，只允许utf-8编码；签名时一定要使用该编码获取字节然后再进行签名
	CallbackJson    string `form:"callback_json" json:"callback_json,omitempty"`                   // 回调业务系统时需要带上的字符串
	ExtJson         string `form:"ext_json" json:"ext_json,omitempty"`                             // 扩展json
	ChannelId       string `form:"channel_id" json:"channel_id,omitempty" binding:"required"`      // 渠道id（非必须），如果未指定method，系统会根据method来找到可用的channel_id
	ExpireTime      string `form:"expire_time" json:"expire_time"  binding:"required"`             // 订单过期时间（个人码必须设置过期时间，否则导致某个金额一致被限制）
}

type NoticeRequest struct {
	Version         string `json:"version,omitempty"form:"version"form:"version"`
	OutTradeNo      string `json:"out_trade_no,omitempty"form:"out_trade_no"`         // 业务订单号
	PayAmount       string `json:"pay_amount,omitempty"form:"pay_amount"`             // 支付金额（分）
	Currency        string `json:"currency,omitempty"form:"currency"`                 // 币种
	ReturnUrl       string `json:"return_url,omitempty"form:"return_url"`             // 支付后跳转的前端地址
	AppId           string `json:"app_id,omitempty"form:"app_id"`                     // 系统给商户分配的app_id
	SignType        string `json:"sign_type,omitempty"form:"sign_type"`               // 加密方法，RSA和MD5，默认RSA
	Sign            string `json:"sign,omitempty"form:"sign"`                         // 签名
	OrderTime       string `json:"order_time,omitempty"form:"order_time"`             // 业务方下单时间，时间格式: 年年年年-月月-日日 时时:分分:秒秒，例如: 2006-01-02 15:04:05
	UserIp          string `json:"user_ip,omitempty"form:"user_ip"`                   // 发起支付的用户ip
	UserId          string `json:"user_id,omitempty"form:"user_id"`                   // 用户在业务系统的id
	PayerAccount    string `json:"payer_account,omitempty"form:"payer_account"`       // 支付者账号，可选
	ProductId       string `json:"product_id,omitempty"form:"product_id"`             // 业务系统的产品id
	ProductName     string `json:"product_name,omitempty"form:"product_name"`         // 商品名称
	ProductDescribe string `json:"product_describe,omitempty"form:"product_describe"` // 商品描述
	Charset         string `json:"charset,omitempty"form:"charset"`                   // 参数编码，只允许utf-8编码；签名时一定要使用该编码获取字节然后再进行签名
	CallbackJson    string `json:"callback_json,omitempty"form:"callback_json"`       // 回调业务系统时需要带上的字符串
	ExtJson         string `json:"ext_json,omitempty"form:"ext_json"`                 // 扩展json
	ChannelId       string `json:"channel_id,omitempty"form:"channel_id"`             // 渠道id（非必须），如果未指定method，系统会根据method来找到可用的channel_id
	Method          string `json:"method,omitempty"form:"method"`                     // 例如：二维码支付，银联支付等。
	FactAmt         string `json:"fact_amt,omitempty"form:"fact_amt"`                 // 实际金额
	FareAmt         string `json:"fare_amt,omitempty"form:"fare_amt"`                 // 手续费
	RandomAmt       string `json:"random_amt" form:"random_amt"`                      // 随机金额，用户实际需要支付的金额
	SuccessTime     string `json:"success_time" form:"success_time"`                  // 支付时间
}

type FormEncoder struct {
	RequestParamsCompacter  *sign.ParamsCompacter
	ResponseParamsCompacter *sign.ParamsCompacter
	NotifyParamsCompacter   *sign.ParamsCompacter
}

func NewFormEncoder() *FormEncoder {
	encoder := &FormEncoder{}
	compacter := sign.NewParamsCompacter(&PayRequest{}, "form", []string{"sign"}, true, "&", "=")
	reponseCompacter := sign.NewParamsCompacter(&PayResponse{}, "form", []string{"sign"}, true, "&", "=")
	notifyRequestCompacter := sign.NewParamsCompacter(&NoticeRequest{}, "form", []string{"sign"}, true, "&", "=")
	encoder.RequestParamsCompacter = &compacter
	encoder.ResponseParamsCompacter = &reponseCompacter
	encoder.NotifyParamsCompacter = &notifyRequestCompacter
	return encoder
}

func (encoder *FormEncoder) SendRequest(apiUrl string, request *PayRequest, key string) (response *PayResponse, err error) {
	signMessage, err := encoder.Sign(*request, key)
	if err != nil {
		err = fmt.Errorf("failed to generate signMessage! error: %v", err.Error())
		return
	}
	request.Sign = signMessage
	params, err := request.Encode()
	if err != nil {
		err = fmt.Errorf("failed to generate map! error: %v", err.Error())
		return
	}
	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}

	logger.Log.Infof("Send params: %v", values)

	client := http.Client{Timeout: time.Second * 10}
	resp, err := client.PostForm(apiUrl, values)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		logger.Log.Errorf("post to: %v with form: %v and returns error: %v", apiUrl, values, err.Error())
		return
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Errorf("post to: %v with form: %v and returns error: %v", apiUrl, values, err.Error())
		return
	}
	payResponse := &PayResponse{}
	err = json.Unmarshal(bytes, payResponse)
	if err != nil {
		logger.Log.Errorf("unmarshal json failed! json: %v returns error: %v", string(bytes), err.Error())
		return
	} else {
		logger.Log.Debugf("post to: %v with form: %v and returns response: %v", apiUrl, values, response)
	}
	return payResponse, err
}

func (encoder *FormEncoder) Sign(request PayRequest, key string) (string, error) {
	source := encoder.RequestParamsCompacter.ParamsToString(request)
	logger.Log.Infof("Encode request: %v to source: %v", request, source)
	signMsg, e := Md5([]byte(source), key)
	if e != nil {
		return "", e
	}
	return signMsg, nil
}

func (encoder *FormEncoder) ParseNotify(request *http.Request) (*NoticeRequest, error) {
	notice := &NoticeRequest{}

	notice.Version = request.FormValue("version")
	notice.OutTradeNo = request.FormValue("out_trade_no")
	notice.PayAmount = request.FormValue("pay_amount")
	notice.Currency = request.FormValue("currency")
	notice.ReturnUrl = request.FormValue("return_url")
	notice.AppId = request.FormValue("app_id")
	notice.SignType = request.FormValue("sign_type")
	notice.Sign = request.FormValue("sign")
	notice.OrderTime = request.FormValue("order_time")
	notice.UserIp = request.FormValue("user_ip")
	notice.UserId = request.FormValue("user_id")
	notice.PayerAccount = request.FormValue("payer_account")
	notice.ProductId = request.FormValue("product_id")
	notice.ProductName = request.FormValue("product_name")
	notice.ProductDescribe = request.FormValue("product_describe")
	notice.Charset = request.FormValue("charset")
	notice.CallbackJson = request.FormValue("callback_json")
	notice.ExtJson = request.FormValue("ext_json")
	notice.ChannelId = request.FormValue("channel_id")
	notice.Method = request.FormValue("method")

	notice.FactAmt = request.FormValue("fact_amt")
	notice.FareAmt = request.FormValue("fare_amt")
	notice.RandomAmt = request.FormValue("random_amt")
	notice.SuccessTime = request.FormValue("success_time")
	return notice, nil
}

func (encoder *FormEncoder) CheckNotifySign(request NoticeRequest, key string) error {
	source := encoder.NotifyParamsCompacter.ParamsToString(request)
	//app_id=1&channel_id=WECHAT&fact_amt=0&fare_amt=0&order_time=2018-11-11 21:17:40&out_trade_no=201811112117407282660121223026580000083&pay_amount=100&product_describe=apple&product_name=apple&random_amt=153&sign_type=MD5&user_ip=127.0.0.1
	logger.Log.Infof("Build source: %v by request: %v", source, request)
	signMsg, e := Md5([]byte(source), key)
	if e != nil {
		return e
	}
	if !util.EqualsIgnoreCase(signMsg, request.Sign) {
		logger.Log.Errorf("check sign error! ours: %v actual: %v", signMsg, request.Sign)
		err := fmt.Errorf("check sign error")
		return err
	}
	return nil
}

func (entity *PayRequest) Encode() (params map[string]string, err error) {
	params = make(map[string]string)
	bytes, err := json.Marshal(entity)
	if err != nil {
		logger.Log.Errorf("Failed to encode! entity: %v error: %v", entity, err)
		return
	}
	err = json.Unmarshal(bytes, &params)
	if err != nil {
		logger.Log.Errorf("Failed to decode! entity: %v error: %v", entity, err)
	}
	return
}

func Md5(source []byte, key string) (string, error) {
	buffer := bytes.NewBuffer(source)
	buffer.Write([]byte(key))
	b := buffer.Bytes()
	sum := md5.Sum(b)
	s := hex.EncodeToString(sum[:])
	return s, nil
}

type PayResponse struct {
	ReturnResult
	Result *PayResult `json:"result"`
}

type PayResult struct {
	PayAmount uint32 `json:"pay_amount"`
	OrderId   string `json:"order_id"`
	QrCode    string `json:"qr_code"`
}

type ReturnResult struct {
	Code     int    `json:"code,omitempty"`
	Message  string `json:"message,omitempty"`
	Describe string `json:"describe,omitempty"`
}

var payRequestCompacter = sign.NewParamsCompacter(&PayRequest{}, "form", []string{"sign"}, true, "&", "=")

func (r *PayRequest) GetCompacter() sign.ParamsCompacter {
	return payRequestCompacter
}

func (r *PayRequest) ParamsToString(i interface{}) string {
	return payRequestCompacter.ParamsToString(i.(*PayRequest))
}

func (r *PayRequest) GetSign() string {
	return r.Sign
}

func (r *PayRequest) GetSignType() string {
	return r.SignType
}
