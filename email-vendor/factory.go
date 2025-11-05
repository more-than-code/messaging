package email

import "fmt"

// NewVendor creates an EmailVendor implementation based on provider config.
func NewVendor(cfg Config) (EmailVendor, error) {
	if cfg.Provider == "" {
		return nil, fmt.Errorf("email provider is required")
	}

	switch cfg.Provider {
	case ProviderPostmark:
		return NewPostmarkVendor(cfg)
	case ProviderMailchimp:
		return NewMailchimpVendor(cfg)
	default:
		return nil, fmt.Errorf("unsupported email provider: %s", cfg.Provider)
	}
}
