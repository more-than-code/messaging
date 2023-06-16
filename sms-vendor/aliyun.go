package sms

import (
	"fmt"
	"log"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/kelseyhightower/envconfig"
)

type AliyunVendor struct {
	clientGlobe *sdk.Client
	cfg         AliyunConfig
}

type AliyunConfig struct {
	AppKeyId          string `envconfig:"SMS_ACCESS_KEY_ID"`
	AppKeySecret      string `envconfig:"SMS_ACCESS_KEY_SECRET"`
	SignName          string `envconfig:"SMS_SIGN_NAME"`
	TemplateCode      string `envconfig:"SMS_TEMPLATE_CODE"`
	AppGlobeKeyId     string `envconfig:"SMS_GLOBE_ACCESS_KEY_ID"`
	AppGlobeKeySecret string `envconfig:"SMS_GLOBE_ACCESS_KEY_SECRET"`
}

func CreateClientGlobe(accessKeyId string, accessKeySecret string) (*sdk.Client, error) {
	return sdk.NewClientWithAccessKey("ap-southeast-1", accessKeyId, accessKeySecret)
}

func (v *AliyunVendor) SendCodeGlobe(phoneNumber, msg string) error {
	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https"
	request.Domain = "dysmsapi.ap-southeast-1.aliyuncs.com"
	request.Version = "2018-05-01"
	request.ApiName = "SendMessageToGlobe"
	request.QueryParams["RegionId"] = "ap-southeast-1"
	request.QueryParams["To"] = phoneNumber
	request.QueryParams["Message"] = msg

	response, err := v.clientGlobe.ProcessCommonRequest(request)

	fmt.Print(response.GetHttpContentString())

	return err
}

func NewAliyunVendor() (*AliyunVendor, error) {
	var cfg AliyunConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	clientGlobe, err := CreateClientGlobe(*tea.String(cfg.AppGlobeKeyId), *tea.String(cfg.AppGlobeKeySecret))
	if err != nil {
		return nil, err
	}

	return &AliyunVendor{clientGlobe: clientGlobe, cfg: cfg}, nil
}
