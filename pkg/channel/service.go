package channel

import (
	"context"
	"flag"
	"github.com/pjoc-team/base-service/pkg/logger"
	"github.com/pjoc-team/base-service/pkg/service"
	"github.com/pjoc-team/channel-isppay/pkg/sdk"
	"github.com/pjoc-team/etcd-config/config"
	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"net/url"
)

var channelApiMap = make(map[pb.Method]string)

func init() {
	channelApiMap[pb.Method_QR_CODE] = "/v1/personal/pay"
}

type PersonalAccount struct {
	GatewayDomain       string `json:"gateway_domain"`
	AppId               string `json:"app_id"`
	Md5Key              string `json:"md5_key"`
	QrCodeExpireSeconds int    `json:"qr_code_expire_seconds"`
}

type PersonalChannel struct {
	*service.Service
	ChannelApiMap            map[pb.Method]string
	PersonalAccountConfigMap map[string]*PersonalAccount
	encoder                  *sdk.FormEncoder
}

func (svc *PersonalChannel) Notify(ctx context.Context, request *pb.NotifyRequest) (response *pb.NotifyResponse, err error) {
	account, exists := svc.PersonalAccountConfigMap[request.PaymentAccount]
	if !exists {
		logger.Log.Errorf("could'nt found account: %v", request.PaymentAccount)
		return nil, status.Newf(codes.InvalidArgument, "could'nt found account: %v", request.PaymentAccount).Err()
	}

	httpRequest := request.Request
	if request.Type == pb.PayType_PAY {
		reqBody := string(httpRequest.Body)
		httpReq, _ := http.NewRequest(request.Request.Method.String(), request.Request.Url, nil)
		httpReq.Form, err = url.ParseQuery(reqBody)
		if err != nil || len(httpReq.Form) == 0 {
			return nil, status.New(codes.InvalidArgument, "req must be a form").Err()
		}

		notify, e := svc.encoder.ParseNotify(httpReq)
		if e != nil {
			if err != nil {
				logger.Log.Errorf("notify error! body: %v error: %v", reqBody, err.Error())
				return nil, status.Newf(codes.Internal, "msg %s", err.Error()).Err()
			}
		}
		err = svc.encoder.CheckNotifySign(*notify, account.Md5Key)
		if err != nil {
			logger.Log.Errorf("Check sign error! response: %v err: %v", reqBody, err.Error())
			return
		}

		response = &pb.NotifyResponse{}
		response.Status = pb.PayStatus_SUCCESS

		httpResponse := &pb.HTTPResponse{}
		httpResponse.Body = []byte("success")
		response.Response = httpResponse
		return response, nil
	} else {
		return nil, errors.New("not implements")
	}
}

func (svc *PersonalChannel) RegisterGrpc(gs *grpc.Server) {
	pb.RegisterPayChannelServer(gs, svc)
}

func initAccountDemo(configMap map[string]*PersonalAccount) {
	account := &PersonalAccount{}
	account.GatewayDomain = "http://asus.pjoc.pub:8889"
	account.Md5Key = "11"
	account.AppId = "1"
	account.QrCodeExpireSeconds = 120
	configMap[account.AppId] = account
}

func Init(svc *service.Service) {
	flag.Parse()
	channelService := &PersonalChannel{}
	channelService.Service = svc
	channelService.ChannelApiMap = channelApiMap
	channelService.PersonalAccountConfigMap = make(map[string]*PersonalAccount)
	channelService.encoder = sdk.NewFormEncoder()
	initAccountDemo(channelService.PersonalAccountConfigMap)
	config.Init(config.URL(svc.ConfigURI), config.WithDefault(&channelService.PersonalAccountConfigMap))

	channelService.StartGrpc(channelService.RegisterGrpc)
}
