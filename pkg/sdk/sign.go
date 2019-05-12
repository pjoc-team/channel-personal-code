package sdk

type SignInterface interface {
	GetAppId() string
	//GetCompacter() sign.ParamsCompacter
	GetSign() string
	GetSignType() string
	ParamsToString(interface{}) string
}


