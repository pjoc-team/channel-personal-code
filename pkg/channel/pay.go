package channel

import (
	"context"
	"fmt"
	"github.com/pjoc-team/base-service/pkg/date"
	"github.com/pjoc-team/base-service/pkg/logger"
	"github.com/pjoc-team/base-service/pkg/url"
	"github.com/pjoc-team/channel-isppay/pkg/sdk"
	pb "github.com/pjoc-team/pay-proto/go"
	"time"
)

const META_KEY_CHANNEL = "channel"

func (svc *PersonalChannel) Pay(ctx context.Context, request *pb.ChannelPayRequest) (response *pb.ChannelPayResponse, err error) {
	platform, exists := pb.Method_value[request.Method]
	if !exists {
		err = fmt.Errorf("method: %v is illegal", request.Method)
		logger.Log.Errorf("illegal request: %v error: %v", request, err.Error())
		return
	}
	channelAccount := request.ChannelAccount
	if channelAccount == "" {
		err = fmt.Errorf("account not found, request: %v", request)
		return
	}
	account := svc.PersonalAccountConfigMap[channelAccount]
	method := pb.Method(platform)
	if method == pb.Method_QR_CODE {
		r := &sdk.PayRequest{}
		r, err = svc.BuildRequest(account, request)
		if err != nil {
			logger.Log.Errorf("Failed to build Request! error: %v", err.Error())
			return
		}
		domain := account.GatewayDomain
		path, exists := channelApiMap[method]
		if !exists {
			err = fmt.Errorf("could'nt found api of this method: %v", method)
			return
		}
		apiUrl := url.CompactUrl(domain, path, "")
		resp := &sdk.PayResponse{}
		resp, err = svc.encoder.SendRequest(apiUrl, r, account.Md5Key)
		if err != nil || resp == nil {
			err = fmt.Errorf("error happened when send request to channel! reason: response is null")
			logger.Log.Errorf("Failed to send request! req: %v error: %v", request, err)
			return
		} else if resp.Code != sdk.CODE_SUCCESS {
			logger.Log.Errorf("Failed to request! code is not success. response: %v", response)
			err = fmt.Errorf("error happened when send request to channel! reason: response is null")
			return
		}
		response = &pb.ChannelPayResponse{}
		data := make(map[string]string)
		response.Data = data
		response.Data["qrcode"] = resp.Result.QrCode
		response.ChannelOrderId = resp.Result.OrderId
		response.Data["amt"] = fmt.Sprintf("%d", resp.Result.PayAmount)
	}
	logger.Log.Infof("Channel request: %v and response: %v", request, response)
	return response, nil
}

func (svc *PersonalChannel) BuildRequest(account *PersonalAccount, request *pb.ChannelPayRequest) (r *sdk.PayRequest, err error) {
	r = &sdk.PayRequest{}
	r.AppId = account.AppId
	r.ProductName = request.Product.Name
	r.ProductDescribe = request.Product.Description
	r.ProductId = request.Product.Id
	r.PayAmount = fmt.Sprintf("%d", request.PayAmount)
	r.UserIp = request.UserIp
	r.NotifyUrl = request.NotifyUrl
	r.OutTradeNo = request.GatewayOrderId
	r.OrderTime = date.NowTime()
	r.ExpireTime = expireTime(account.QrCodeExpireSeconds)
	r.SignType = "MD5"
	r.ChannelId = request.Meta[META_KEY_CHANNEL]
	return
}

func expireTime(expireSeconds int) string {
	d := time.Now().Add(time.Duration(expireSeconds) * time.Second)
	return d.Format(date.TIME_FORMAT)
}
