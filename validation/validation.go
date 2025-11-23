package validation

import (
	"errors"
	"net/mail"
	"regexp"
)

// Pre-compile regex patterns for better performance
var (
	upperRegex    = regexp.MustCompile(`[A-Z]`)
	lowerRegex    = regexp.MustCompile(`[a-z]`)
	numberRegex   = regexp.MustCompile(`[0-9]`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}
	//parsing the email
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("invalid email format")
	}

	return nil
}
func ValidatePassword(password string) error {
	// Use pre-compiled regex patterns for better performance
	if !upperRegex.MatchString(password) || !lowerRegex.MatchString(password) || !numberRegex.MatchString(password) {
		return errors.New("password must contain uppercase, lowercase, and number")
	}
	return nil
}

// validate username
func ValidateUsername(username string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}
	if len(username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}
	if !usernameRegex.MatchString(username) {
		return errors.New("username can only contain letters, numbers, and underscores")
	}
	return nil
}
