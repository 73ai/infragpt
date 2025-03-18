package postgres

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/supporting/google"
	"time"
)

func (i *IdentityDB) ValidateStateToken(ctx context.Context, token string) error {
	stateToken, err := i.queries.StateToken(ctx, token)
	if err != nil {
		return fmt.Errorf("state token: %w", err)
	}

	if stateToken.Revoked {
		return google.ErrStateTokenRevoked
	}

	return nil
}

func (i *IdentityDB) NewStateToken(ctx context.Context) (string, error) {
	token, err := newStateToken()
	if err != nil {
		return "", fmt.Errorf("new state token: %w", err)
	}
	err = i.queries.CreateStateToken(ctx, CreateStateTokenParams{
		Token:     token,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	})
	if err != nil {
		return "", fmt.Errorf("create state token: %w", err)
	}
	return token, nil
}

func (i *IdentityDB) ExpireStateToken(ctx context.Context, token string) error {
	err := i.ExpireStateToken(
		ctx,
		token)
	if err != nil {
		return fmt.Errorf("expire state token: %w", err)
	}

	return nil
}

var _ google.StateTokenRepository = &IdentityDB{}

func newStateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
