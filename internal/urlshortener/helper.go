package urlshortener

import (
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"strings"
)

func normalizeURL(u string) (string, error) {
	u = strings.TrimSpace(u)
	if u == "" {
		return "", ErrInvalidURL
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return "", ErrInvalidURL
	}
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}
	if parsed.Host == "" {
		return "", ErrInvalidURL
	}
	return parsed.String(), nil
}

func sanitizeAlias(a string) string {
	a = strings.TrimSpace(a)
	a = strings.Trim(a, "/")
	return a
}

func generateShortCode(original string) string {
	sum := sha256.Sum256([]byte(original))
	encoded := base64.URLEncoding.EncodeToString(sum[:])
	if len(encoded) > 10 {
		return encoded[:10]
	}
	return encoded
}
