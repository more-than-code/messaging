package sender

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/keighl/postmark"
	"github.com/kelseyhightower/envconfig"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type MailVendor struct {
	cfg MailConfig
}

type MailConfig struct {
	SendgridApiKey string `envconfig:"SENDGRID_API_KEY"`
	PostmarkApiKey string `envconfig:"POSTMARK_API_KEY"`
}

func NewMailVendor() (*MailVendor, error) {
	var cfg MailConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	return &MailVendor{cfg: cfg}, nil
}

func (v *MailVendor) SendCode(mailAddress, sub, msg string) error {
	from := mail.NewEmail("毛孩街 support", "support@mohiguide.com")
	subject := sub
	to := mail.NewEmail("毛孩街用戶", mailAddress)
	plainTextContent := msg
	htmlContent := fmt.Sprintf("<strong>%s</strong>", msg)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(v.cfg.SendgridApiKey)
	response, err := client.Send(message)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
	return err
}

func (v *MailVendor) SendCodeFromPostmark(mailAddress, sub, msg string) error {
	subject := sub
	htmlContent := msg

	client := postmark.NewClient(v.cfg.PostmarkApiKey, "")

	email := postmark.Email{
		From:     "support@mohiguide.com",
		To:       mailAddress,
		Subject:  subject,
		HtmlBody: htmlContent,
		// TextBody:   "",
		Tag:        "verification-code",
		TrackOpens: true,
	}

	res, err := client.SendEmail(email)

	if err != nil {
		log.Println(err.Error())
		return err
	}

	log.Println(res)

	return nil
}

func (v *MailVendor) SendCodeFromPostmark2(mailAddress, sub, msg string) error {
	subject := sub
	htmlContent := msg

	prg := "curl"

	arg1 := "https://api.postmarkapp.com/email"
	arg2_1 := "-X"
	arg2_2 := "POST"
	arg3_1 := "-H"
	arg3_2 := "Accept: application/json"
	arg4_1 := "-H"
	arg4_2 := "Content-Type: application/json"
	arg5_1 := "-H"
	arg5_2 := fmt.Sprintf("X-Postmark-Server-Token: %s", v.cfg.PostmarkApiKey)
	arg6_1 := "-d"
	arg6_2 := fmt.Sprintf("{From: 'support@mohiguide.com', To: '%s', Subject: '%s', HtmlBody: '%s'}", mailAddress, subject, htmlContent)

	cmd := exec.Command(prg, arg1, arg2_1, arg2_2, arg3_1, arg3_2, arg4_1, arg4_2, arg5_1, arg5_2, arg6_1, arg6_2)
	stdout, err := cmd.Output()

	if err != nil {
		log.Println(err.Error())
		return err
	}

	log.Println(string(stdout))

	return nil
}

type Mail struct {
	Sender  string
	To      []string
	CC      []string
	Bcc     []string
	Subject string
	Body    string
}

func ComposeMsg(mail Mail) string {
	// empty string
	msg := ""
	// set sender
	msg += fmt.Sprintf("From: %s\r\n", mail.Sender)
	// if more than 1 recipient
	if len(mail.To) > 0 {
		msg += fmt.Sprintf("Cc: %s\r\n", strings.Join(mail.CC, ";"))
	}

	// add subject
	msg += fmt.Sprintf("Subject: %s\r\n", mail.Subject)
	msg += "Content-Type: text/plain; charset=UTF-8\r\n\r\n"
	// add mail body
	msg += fmt.Sprintf("%s\r\n", mail.Body)
	return msg
}
