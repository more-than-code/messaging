package email

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
		FromEmail string `json:"from_email"`
		Subject   string `json:"subject"`
		To        []To   `json:"to"`
		Text      string `json:"text"`
	} `json:"message"`
}

type To struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (v *MailchimpVendor) SendEmail(to, bcc, sub, msg string) error {
	// Prepare the payload
	payload := SendMessageRequest{
		Key: v.cfg.APIKey,
	}

	// Set message details
	payload.Message.FromEmail = v.cfg.EmailSender
	payload.Message.Subject = sub
	payload.Message.Text = msg
	payload.Message.To = []To{
		{
			Email: to,
			Name:  "Recipient",
		},
	}

	// Convert payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %v", err)
	}

	// Define the API endpoint URL
	url := "https://mandrillapp.com/api/1.0/messages/send"

	// Make the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %v", err)
	}

	// Set content-type header
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send email, HTTP status: %v", resp.Status)
	}

	// If no errors, return nil (success)
	return nil
}

func (v *MailchimpVendor) SendCode(mailAddress, sub, msg string) error {
	return nil
}

func (v *MailchimpVendor) SendEmailWithAttachment(to, bcc, sub, msg string, attachments []Attachment) error {
	return nil
}
