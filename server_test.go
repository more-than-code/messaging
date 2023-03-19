package messaging

import "testing"

func TestGenerateVerificationCode(t *testing.T) {
	msg, err := templateToMessage("The code is {{.Code}}", "1234")

	if err != nil {
		t.Error(err)
	}

	t.Log(msg)
}
