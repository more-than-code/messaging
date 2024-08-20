package sms

import (
	"fmt"
	"strings"

	"github.com/byteplus-sdk/byteplus-sdk-golang/service/sms"
	"github.com/kelseyhightower/envconfig"
	"github.com/volcengine/volc-sdk-golang/base"
)

type BytePlusVendor struct {
	cfg BytePlusConfig
}

type BytePlusConfig struct {
	AccessKey string `envconfig:"BYTEPLUS_ACCESSKEY"`
	SecretKey string `envconfig:"BYTEPLUS_SECRETKEY"`
	Account   string `envconfig:"BYTEPLUS_ACCOUNT"`
	Template  string `envconfig:"BYTEPLUS_TEMPLATE"`
}

func NewBytePlusVendor() (*BytePlusVendor, error) {
	var cfg BytePlusConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}

	return &BytePlusVendor{cfg: cfg}, nil
}

func (v *BytePlusVendor) SendCode(phoneNumber, code string) error {
	sms.DefaultInstance.Client.SetAccessKey(v.cfg.AccessKey)
	sms.DefaultInstance.Client.SetSecretKey(v.cfg.SecretKey)

	var tempId string
	if strings.HasPrefix(phoneNumber, "86") || strings.HasPrefix(phoneNumber, "+86") {
		tempId = v.cfg.Template
	} else {
		tempId = v.cfg.Template
	}

	req := &sms.SmsRequest{
		SmsAccount:    v.cfg.Account,
		TemplateID:    tempId,
		TemplateParam: fmt.Sprintf(`{"code": "%s"}`, code),
		PhoneNumbers:  phoneNumber,
		Tag:           "msgs",
	}
	result, statusCode, err := sms.DefaultInstance.Send(req)
	fmt.Printf("result = %+v\n", result)
	fmt.Printf("statusCode = %+v\n", statusCode)
	fmt.Printf("err = %+v\n", err)

	return err
}

func (v *BytePlusVendor) SendCodeNProduct(phoneNumber, code, product string) error {
	i18nInstance := sms.NewInstanceI18n(base.RegionApSingapore)
	i18nInstance.Client.SetAccessKey(v.cfg.AccessKey)
	i18nInstance.Client.SetSecretKey(v.cfg.SecretKey)

	req := &sms.SmsRequest{
		SmsAccount:    v.cfg.Account,
		TemplateID:    v.cfg.Template,
		TemplateParam: fmt.Sprintf(`{"code": "%s", "product": "%s"}`, code, product),
		PhoneNumbers:  phoneNumber,
		From:          "SaasifyEase",
		Tag:           "msgs",
	}
	result, statusCode, err := i18nInstance.Send(req)
	fmt.Printf("result = %+v\n", result)
	fmt.Printf("result.ResponseMetadata.Error = %+v\n", result.ResponseMetadata.Error)
	fmt.Printf("statusCode = %+v\n", statusCode)
	fmt.Printf("err = %+v\n", err)

	return err
}
