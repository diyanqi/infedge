package model

import "strings"

const emailParts = 2

// NormalizeEmail canonicalizes an email for identity checks while preserving
// the original address in User.Email for display and outbound mail.
func NormalizeEmail(raw string) string {
	email := strings.ToLower(strings.TrimSpace(raw))
	parts := strings.SplitN(email, "@", emailParts)
	if len(parts) != emailParts {
		return email
	}
	local, domain := parts[0], parts[1]
	if plus := strings.IndexByte(local, '+'); plus >= 0 {
		local = local[:plus]
	}
	if domain == "gmail.com" || domain == "googlemail.com" {
		local = strings.ReplaceAll(local, ".", "")
		if domain == "googlemail.com" {
			domain = "gmail.com"
		}
	}
	return local + "@" + domain
}
