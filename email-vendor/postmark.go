package email

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/keighl/postmark"
)

type PostmarkVendor struct {
	cfg Config
}

func NewPostmarkVendor(cfg Config) (*PostmarkVendor, error) {
	if cfg.Provider != ProviderPostmark {
		return nil, errors.New("postmark vendor requires provider POSTMARK")
	}
	if cfg.APIKey == "" {
		return nil, errors.New("postmark api key is required")
	}
	if cfg.EmailSender == "" {
		return nil, errors.New("postmark sender is required")
	}

	return &PostmarkVendor{cfg: cfg}, nil
}

func (v *PostmarkVendor) SendCode(mailAddress, sub, msg string) error {
	log.Printf("postmark: sending code to %s with subject %s", mailAddress, sub)
	subject := sub
	htmlContent := msg

	client := postmark.NewClient(v.cfg.APIKey, "")

	email := postmark.Email{
		From:     v.cfg.EmailSender,
		To:       mailAddress,
		Subject:  subject,
		HtmlBody: htmlContent,
		// TextBody:   "",
		Tag:        "verification-code",
		TrackOpens: true,
	}

	res, err := client.SendEmail(email)

	if err != nil {
		log.Printf("postmark: failed to send code to %s: %v", mailAddress, err)
		return err
	}

	log.Printf("postmark: sent code to %s, message id %s", mailAddress, res.MessageID)

	return nil
}

func (v *PostmarkVendor) SendCodeFromPostmark2(mailAddress, sub, msg string) error {
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
	arg5_2 := fmt.Sprintf("X-Postmark-Server-Token: %s", v.cfg.APIKey)
	arg6_1 := "-d"
	arg6_2 := fmt.Sprintf("{From: '%s', To: '%s', Subject: '%s', HtmlBody: '%s'}", v.cfg.EmailSender, mailAddress, subject, htmlContent)

	cmd := exec.Command(prg, arg1, arg2_1, arg2_2, arg3_1, arg3_2, arg4_1, arg4_2, arg5_1, arg5_2, arg6_1, arg6_2)
	stdout, err := cmd.Output()

	if err != nil {
		log.Printf("postmark: curl fallback failed for %s: %v", mailAddress, err)
		return err
	}

	log.Printf("postmark: curl fallback response for %s: %s", mailAddress, string(stdout))

	return nil
}

func (v *PostmarkVendor) SendEmail(to, bcc, sub, msg string) error {
	log.Printf("postmark: sending email to %s, bcc %s", to, bcc)
	subject := sub
	htmlContent := msg

	client := postmark.NewClient(v.cfg.APIKey, "")

	email := postmark.Email{
		From:       v.cfg.EmailSender,
		To:         to,
		Bcc:        bcc,
		Subject:    subject,
		HtmlBody:   htmlContent,
		Tag:        "email",
		TrackOpens: true,
	}

	res, err := client.SendEmail(email)

	if err != nil {
		log.Printf("postmark: failed to send email to %s: %v", to, err)
		return err
	}

	log.Printf("postmark: sent email to %s, message id %s", to, res.MessageID)

	return nil
}

func (v *PostmarkVendor) SendEmailWithAttachment(to, bcc, sub, msg string, attachments []Attachment) error {
	log.Printf("postmark: sending email with %d attachments to %s", len(attachments), to)
	subject := sub
	htmlContent := msg

	client := postmark.NewClient(v.cfg.APIKey, "")

	pmAttachments := []postmark.Attachment{}
	for _, a := range attachments {
		// Postmark expects base64 encoded string, same as our internal format
		pmAttachments = append(pmAttachments, postmark.Attachment{
			Name:        a.Name,
			Content:     a.Content, // Already base64 encoded
			ContentType: a.ContentType,
		})
	}

	email := postmark.Email{
		From:        v.cfg.EmailSender,
		To:          to,
		Bcc:         bcc,
		Subject:     subject,
		HtmlBody:    htmlContent,
		Attachments: pmAttachments,
		Tag:         "attachment",
		TrackOpens:  true,
	}

	res, err := client.SendEmail(email)

	if err != nil {
		log.Printf("postmark: failed to send email with attachment to %s: %v", to, err)
		return err
	}

	log.Printf("postmark: sent email with attachment to %s, message id %s", to, res.MessageID)

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
