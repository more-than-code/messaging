package sms

type SmsVendor interface {
	SendCode(phoneNumber, code string) error
	SendCodeNProduct(phoneNumber, code, product string) error
}
