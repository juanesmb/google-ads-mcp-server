package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

const (
	// TokenRefreshBuffer is the time before expiration when we'll refresh the token
	TokenRefreshBuffer = 5 * time.Minute
	GoogleAdsScope     = "https://www.googleapis.com/auth/adwords"
	TokenURI           = "https://oauth2.googleapis.com/token"
)

type TokenManager struct {
	config       *jwt.Config
	tokenSource  oauth2.TokenSource
	cachedToken  *oauth2.Token
	mu           sync.RWMutex
	clientEmail  string
	privateKey   []byte
	tokenRefresh time.Duration
}

type Config struct {
	ClientEmail string
	PrivateKey  string
	Scopes      []string
}

// NewTokenManagerFromServiceAccount creates a TokenManager from service account JSON
func NewTokenManagerFromServiceAccount(serviceAccountJSON []byte, scopes ...string) (*TokenManager, error) {
	if len(scopes) == 0 {
		scopes = []string{GoogleAdsScope}
	}

	jwtConfig, err := google.JWTConfigFromJSON(serviceAccountJSON, scopes...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service account JSON: %w", err)
	}

	tokenSource := jwtConfig.TokenSource(context.Background())

	tm := &TokenManager{
		config:       jwtConfig,
		tokenSource:  tokenSource,
		clientEmail:  jwtConfig.Email,
		privateKey:   jwtConfig.PrivateKey,
		tokenRefresh: TokenRefreshBuffer,
	}

	return tm, nil
}

// GetAccessToken returns a valid access token, refreshing if necessary
func (tm *TokenManager) GetAccessToken(ctx context.Context) (string, error) {
	tm.mu.RLock()
	token := tm.cachedToken
	tm.mu.RUnlock()

	// Check if we have a valid cached token
	if token != nil && token.Valid() && !tm.shouldRefresh(token) {
		return token.AccessToken, nil
	}

	// Need to refresh the token
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have refreshed)
	if tm.cachedToken != nil && tm.cachedToken.Valid() && !tm.shouldRefresh(tm.cachedToken) {
		return tm.cachedToken.AccessToken, nil
	}

	// Fetch a new token
	newToken, err := tm.tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to obtain access token: %w", err)
	}

	tm.cachedToken = newToken
	return newToken.AccessToken, nil
}

// shouldRefresh determines if the token should be refreshed
// Returns true if the token will expire within the refresh buffer time
func (tm *TokenManager) shouldRefresh(token *oauth2.Token) bool {
	if token == nil || token.Expiry.IsZero() {
		return true
	}
	return time.Until(token.Expiry) < tm.tokenRefresh
}

// InvalidateToken clears the cached token, forcing a refresh on next access
func (tm *TokenManager) InvalidateToken() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.cachedToken = nil
}

// GetTokenInfo returns information about the current token (for debugging/monitoring)
func (tm *TokenManager) GetTokenInfo() (hasToken bool, expiresIn time.Duration, isValid bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.cachedToken == nil {
		return false, 0, false
	}

	return true, time.Until(tm.cachedToken.Expiry), tm.cachedToken.Valid()
}

// VerifyCredentials validates that the credentials can obtain a token
func (tm *TokenManager) VerifyCredentials(ctx context.Context) error {
	_, err := tm.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("credential verification failed: %w", err)
	}
	return nil
}
