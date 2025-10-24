package auth

import "context"

// TokenProvider defines the interface for obtaining access tokens
type TokenProvider interface {
	GetAccessToken(ctx context.Context) (string, error)
}

