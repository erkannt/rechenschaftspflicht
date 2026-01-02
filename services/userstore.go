package services

import "strings"

var (
	allowedEmails = []string{
		"foo@example.com",
		"alice@example.com",
		"bob@example.com",
	}
)

func IsAllowedEmail(email string) bool {
	for _, e := range allowedEmails {
		if strings.EqualFold(e, email) {
			return true
		}
	}
	return false
}
