package auth

import "errors"

var (
	ErrEmailAlreadyUsed   = errors.New("email already in use")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrTokenNotFound      = errors.New("token not found or expired")
	ErrTOTPRequired       = errors.New("totp code required")
	ErrInvalidTOTP        = errors.New("invalid totp code")
	ErrTOTPAlreadyEnabled = errors.New("totp already enabled")
	ErrTOTPNotSetup       = errors.New("totp not configured")
)
