package email

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type MailchimpVendor struct {
	cfg Config
}

func NewMailchimpVendor(cfg Config) (*MailchimpVendor, error) {
	if cfg.Provider != ProviderMailchimp {
		return nil, errors.New("mailchimp vendor requires provider MAILCHIMP")
	}
	if cfg.APIKey == "" {
		return nil, errors.New("mailchimp api key is required")
	}
	if cfg.EmailSender == "" {
		return nil, errors.New("mailchimp sender is required")
	}

	return &MailchimpVendor{cfg: cfg}, nil
}

// Structure for the request payload
type SendMessageRequest struct {
	Key     string `json:"key"` // Your Mailchimp Transactional API key
	Message struct {
		FromEmail   string               `json:"from_email"`
		Subject     string               `json:"subject"`
		To          []To                 `json:"to"`
		Text        string               `json:"text"`
		Html        string               `json:"html,omitempty"`
		BccAddress  string               `json:"bcc_address,omitempty"`
		Attachments []MandrillAttachment `json:"attachments,omitempty"`
	} `json:"message"`
}

type To struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Type  string `json:"type,omitempty"`
}

type MandrillAttachment struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

func (v *MailchimpVendor) SendEmail(to, bcc, sub, msg string) error {
	log.Printf("mailchimp: sending email to %s, bcc %s", to, bcc)
	return v.send(to, bcc, sub, msg, nil)
}

func (v *MailchimpVendor) SendCode(mailAddress, sub, msg string) error {
	log.Printf("mailchimp: sending code to %s with subject %s", mailAddress, sub)
	return v.send(mailAddress, "", sub, msg, nil)
}

func (v *MailchimpVendor) SendEmailWithAttachment(to, bcc, sub, msg string, attachments []Attachment) error {
	log.Printf("mailchimp: sending email with %d attachments to %s", len(attachments), to)
	return v.send(to, bcc, sub, msg, attachments)
}

func (v *MailchimpVendor) send(to, bcc, sub, msg string, attachments []Attachment) error {
	payload := SendMessageRequest{Key: v.cfg.APIKey}
	payload.Message.FromEmail = v.cfg.EmailSender
	payload.Message.Subject = sub
	payload.Message.Text = msg
	if msg != "" {
		payload.Message.Html = msg
	}
	payload.Message.To = []To{{
		Email: to,
		Name:  "Recipient",
		Type:  "to",
	}}
	if bcc != "" {
		payload.Message.BccAddress = bcc
	}

	if len(attachments) > 0 {
		for _, a := range attachments {
			payload.Message.Attachments = append(payload.Message.Attachments, MandrillAttachment{
				Type:    a.ContentType,
				Name:    a.Name,
				Content: a.Content,
			})
		}
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %v", err)
	}

	url := "https://mandrillapp.com/api/1.0/messages/send"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("mailchimp: request failed to %s: %v", to, err)
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("mailchimp: non-200 response for %s, status %s", to, resp.Status)
		return fmt.Errorf("failed to send email, HTTP status: %v", resp.Status)
	}

	log.Printf("mailchimp: message sent to %s successfully", to)
	return nil
}
