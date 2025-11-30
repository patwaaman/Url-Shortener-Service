package errconst

import "errors"

var (
	ErrInvalidURL = errors.New("invalid url")
	ErrNotFound   = errors.New("url not found")
	ErrAliasTaken = errors.New("alias already taken")

	ErrMissingToken   = errors.New("missing bearer token")
	ErrTokenInvalid   = errors.New("invalid token")
	ErrTokenExpired   = errors.New("token expired")
	ErrTokenMalformed = errors.New("malformed token")
	ErrTokenSignature = errors.New("invalid token signature")

	ErrInvalidPayload    = errors.New("invalid payload")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrTokenGeneration   = errors.New("token generation failed")

	ErrRateLimitExceed = errors.New("rate limit exceed")

	ErrShortenUrlFailed     = errors.New("failed to shorten url")
	ErrShortCodeNotFound    = errors.New("short code not mapped to url")
	ErrResolveUrlFailed     = errors.New("failed to resolve code to url")
	ErrAnalyticsStatsFailed = errors.New("failed to get analytics stats")

	ErrInvalidToDate   = errors.New("invalid to date")
	ErrInvalidFromDate = errors.New("invalid from date")
)
