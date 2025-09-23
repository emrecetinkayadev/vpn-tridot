package billing

import "errors"

var (
	ErrUnsupportedProvider = errors.New("unsupported payment provider")
	ErrProviderDisabled    = errors.New("payment provider disabled")
	ErrPlanNotFound        = errors.New("plan not found")
	ErrInvalidWebhook      = errors.New("invalid webhook payload")
)
