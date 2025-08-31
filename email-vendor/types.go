package email

type Attachment struct {
	// Name: attachment name
	Name string
	// Content: Base64 encoded attachment data
	Content string
	// ContentType: attachment MIME type
	ContentType string
	// ContentId: populate for inlining images with the images cid
	ContentID string `json:",omitempty"`
}

type EmailVendor interface {
	SendCode(mailAddress, sub, msg string) error
	SendEmail(to, bcc, sub, msg string) error
	SendEmailWithAttachment(to, bcc, sub, msg string, attachments []Attachment) error
}
