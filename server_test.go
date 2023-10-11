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

	v.SendCode("85254997909", "test012")
}
