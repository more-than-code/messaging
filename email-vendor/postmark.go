package email

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/keighl/postmark"
	"github.com/kelseyhightower/envconfig"
)

type Vendor struct {
	cfg MailConfig
}

type MailConfig struct {
	PostmarkApiKey      string `envconfig:"POSTMARK_API_KEY"`
	PostmarkEmailSender string `envconfig:"POSTMARK_MAIL_SENDER"`
}

func NewVendor() (*Vendor, error) {
	var cfg MailConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	return &Vendor{cfg: cfg}, nil
}

func (v *Vendor) SendCodeFromPostmark(mailAddress, sub, msg string) error {
	subject := sub
	htmlContent := msg

	client := postmark.NewClient(v.cfg.PostmarkApiKey, "")

	email := postmark.Email{
		From:     v.cfg.PostmarkEmailSender,
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

func (v *Vendor) SendCodeFromPostmark2(mailAddress, sub, msg string) error {
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
	arg6_2 := fmt.Sprintf("{From: '%s', To: '%s', Subject: '%s', HtmlBody: '%s'}", v.cfg.PostmarkEmailSender, mailAddress, subject, htmlContent)

	cmd := exec.Command(prg, arg1, arg2_1, arg2_2, arg3_1, arg3_2, arg4_1, arg4_2, arg5_1, arg5_2, arg6_1, arg6_2)
	stdout, err := cmd.Output()

	if err != nil {
		log.Println(err.Error())
		return err
	}

	log.Println(string(stdout))

	return nil
}

func (v *Vendor) SendWithAttachment(mailAddress, sub, msg string, attachment postmark.Attachment) error {
	subject := sub
	htmlContent := msg

	client := postmark.NewClient(v.cfg.PostmarkApiKey, "")

	email := postmark.Email{
		From:        v.cfg.PostmarkEmailSender,
		To:          mailAddress,
		Subject:     subject,
		HtmlBody:    htmlContent,
		Attachments: []postmark.Attachment{attachment},
		Tag:         "attachment",
		TrackOpens:  true,
	}

	res, err := client.SendEmail(email)

	if err != nil {
		log.Println(err.Error())
		return err
	}

	log.Println(res)

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
