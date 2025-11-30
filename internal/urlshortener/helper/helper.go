package helper

import (
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"strings"
	errconst "url-shortener/internal/error"
)

func NormalizeURL(u string) (string, error) {
	u = strings.TrimSpace(u)
	if u == "" {
		return "", errconst.ErrInvalidURL
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return "", errconst.ErrInvalidURL
	}
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}
	if parsed.Host == "" {
		return "", errconst.ErrInvalidURL
	}
	return parsed.String(), nil
}

func SanitizeAlias(a string) string {
	a = strings.TrimSpace(a)
	a = strings.Trim(a, "/")
	return a
}

func GenerateShortCode(original string) string {
	sum := sha256.Sum256([]byte(original))
	encoded := base64.URLEncoding.EncodeToString(sum[:])
	if len(encoded) > 10 {
		return encoded[:10]
	}
	return encoded
}
