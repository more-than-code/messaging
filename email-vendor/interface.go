package email

import "github.com/keighl/postmark"

type EmailVendor interface {
	SendCode(mailAddress, sub, msg string) error
	SendEmailWithAttachment(to, bcc, sub, msg string, attachments []postmark.Attachment) error
}
