package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateCSRFToken creates a CSRF token tied to the given session token.
func GenerateCSRFToken(sessionToken, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sessionToken))
	return hex.EncodeToString(mac.Sum(nil))
}

// ValidateCSRFToken checks whether the CSRF token matches the session token.
func ValidateCSRFToken(csrfToken, sessionToken, secret string) bool {
	expected := GenerateCSRFToken(sessionToken, secret)
	return hmac.Equal([]byte(csrfToken), []byte(expected))
}
