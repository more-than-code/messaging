package sms

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/volcengine/volc-sdk-golang/service/sms"
)

type Vendor struct {
	cfg VolcConfig
}

type VolcConfig struct {
	AccessKey  string `envconfig:"VOLC_ACCESSKEY"`
	SecretKey  string `envconfig:"VOLC_SECRETKEY"`
	Account    string `envconfig:"VOLC_ACCOUNT"`
	Sign       string `envconfig:"VOLC_SIGN"`
	Template   string `envconfig:"VOLC_TEMPLATE"`
	TemplateCn string `envconfig:"VOLC_TEMPLATE_CN"`
}

func NewVendor() (*Vendor, error) {
	var cfg VolcConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}

	return &Vendor{cfg: cfg}, nil
}

func (v *Vendor) SendCode(phoneNumber, code string) error {
	sms.DefaultInstance.Client.SetAccessKey(v.cfg.AccessKey)
	sms.DefaultInstance.Client.SetSecretKey(v.cfg.SecretKey)

	var tempId string
	if strings.HasPrefix(phoneNumber, "86") || strings.HasPrefix(phoneNumber, "+86") {
		tempId = v.cfg.TemplateCn
	} else {
		tempId = v.cfg.Template
	}

	req := &sms.SmsRequest{
		SmsAccount:    v.cfg.Account,
		Sign:          v.cfg.Sign,
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

func (v *Vendor) SendCodeNProduct(phoneNumber, code, product string) error {
	sms.DefaultInstance.Client.SetAccessKey(v.cfg.AccessKey)
	sms.DefaultInstance.Client.SetSecretKey(v.cfg.SecretKey)

	var tempId string
	if strings.HasPrefix(phoneNumber, "86") || strings.HasPrefix(phoneNumber, "+86") {
		tempId = v.cfg.TemplateCn
	} else {
		tempId = v.cfg.Template
	}

	req := &sms.SmsRequest{
		SmsAccount:    v.cfg.Account,
		Sign:          v.cfg.Sign,
		TemplateID:    tempId,
		TemplateParam: fmt.Sprintf(`{"code": "%s", "product": "%s"}`, code, product),
		PhoneNumbers:  phoneNumber,
		Tag:           "msgs",
	}
	result, statusCode, err := sms.DefaultInstance.Send(req)
	fmt.Printf("result = %+v\n", result)
	fmt.Printf("result.ResponseMetadata.Error = %+v\n", result.ResponseMetadata.Error)
	fmt.Printf("statusCode = %+v\n", statusCode)
	fmt.Printf("err = %+v\n", err)

	return err
}
