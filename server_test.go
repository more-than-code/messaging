package messaging

import (
	"testing"

	"github.com/more-than-code/messaging/sms-vendor"
)

func TestGenerateVerificationCode(t *testing.T) {
	msg, err := templateToMessage("The code is {{.Code}}", "1234")

	if err != nil {
		t.Error(err)
	}

	t.Log(msg)
}

func TestSendSms(t *testing.T) {
	v, _ := sms.NewVendor()
	//85254997909
	//8108039627413
	// v.SendCode("8108039627413", "test0123")
	v.SendCodeNProduct("85254997909", "test012", "TestProduct")
}

func TestBytePluysSendSms(t *testing.T) {
	v, _ := sms.NewBytePlusVendor()
	//85254997909
	//8108039627413
	// v.SendCode("8108039627413", "test0123")
	v.SendCodeNProduct("8108039627413", "01234", "TestProduct")
}
