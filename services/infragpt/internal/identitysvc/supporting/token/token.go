package token

import (
	"crypto/rsa"
	"fmt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt/identity"
	"golang.org/x/crypto/bcrypt"
)

type tokenManager struct {
	privateKey *rsa.PrivateKey
}

const RefreshTokenExpiry = 30 * 24 * time.Hour

func (t tokenManager) NewRefreshToken(sessionID string) (identitysvc.RefreshToken, error) {
	token := uuid.NewString()
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return identitysvc.RefreshToken{}, fmt.Errorf("hash token: %w", err)
	}

	jti := uuid.New()
	expiryAt := time.Now().Add(RefreshTokenExpiry)
	claims := jwt.RegisteredClaims{
		ID:        jti.String(),
		Subject:   sessionID,
		ExpiresAt: jwt.NewNumericDate(expiryAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(t.privateKey)
	if err != nil {
		return identitysvc.RefreshToken{}, fmt.Errorf("sign token: %w", err)
	}

	return identitysvc.RefreshToken{
		HashedToken: string(hashedToken),
		TokenID:     jti.String(),
		TokenString: tokenString,
		ExpiryAt:    expiryAt,
	}, nil
}

func (t tokenManager) NewAccessToken(sessionID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   sessionID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(t.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return tokenString, nil
}

func (t tokenManager) ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return "", fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return &t.privateKey.PublicKey, nil
		})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("validate token: %w; %w",
			err, identity.ErrInvalidRefreshToken)
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		if claims.ID == "" || claims.Subject == "" {
			return "", identity.ErrInvalidRefreshToken
		}
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			return "", identity.ErrRefreshTokenExpired
		}
		return claims.ID, nil
	} else {
		return "", identity.ErrInvalidRefreshToken
	}
}

func (t tokenManager) ValidateAccessToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return "", fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return &t.privateKey.PublicKey, nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("validate token: %w; %w",
			err, identity.ErrInvalidAccessToken)
	}

	var sessionID string
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		sessionID = claims.Subject
		slog.Info("claims.Subject is 2", "claims", token.Claims)
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			return "", identity.ErrAccessTokenExpired
		}
	} else {
		sessionID, err = token.Claims.GetSubject()
		if err != nil {
			return "", identity.ErrInvalidAccessToken
		}
		expiryTime, err := token.Claims.GetExpirationTime()
		if err != nil {
			return "", identity.ErrInvalidAccessToken
		}
		if expiryTime.Before(time.Now()) {
			return "", identity.ErrAccessTokenExpired
		}
	}

	return sessionID, nil
}

func NewManager(privateKey *rsa.PrivateKey) identitysvc.TokenManager {
	return &tokenManager{
		privateKey: privateKey,
	}
}
