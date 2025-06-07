package identityapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ClerkValidator struct {
	webhookSecret string
}

func NewClerkValidator(webhookSecret string) *ClerkValidator {
	return &ClerkValidator{
		webhookSecret: webhookSecret,
	}
}

func (cv *ClerkValidator) VerifyWebhookSignature(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signature := r.Header.Get("svix-signature")
		if signature == "" {
			http.Error(w, "Missing svix-signature header", http.StatusUnauthorized)
			return
		}

		timestamp := r.Header.Get("svix-timestamp")
		if timestamp == "" {
			http.Error(w, "Missing svix-timestamp header", http.StatusUnauthorized)
			return
		}

		msgId := r.Header.Get("svix-id")
		if msgId == "" {
			http.Error(w, "Missing svix-id header", http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		if !cv.verifySignature(signature, timestamp, msgId, body) {
			http.Error(w, "Invalid webhook signature", http.StatusUnauthorized)
			return
		}

		// Restore body for downstream handlers
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		next.ServeHTTP(w, r)
	})
}

func (cv *ClerkValidator) verifySignature(signature, timestamp, msgId string, body []byte) bool {
	// Check timestamp is not too old (5 minutes tolerance)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	
	now := time.Now().Unix()
	if now-ts > 300 { // 5 minutes
		return false
	}

	// Extract webhook secret (remove whsec_ prefix if present)
	secret := cv.webhookSecret
	if strings.HasPrefix(secret, "whsec_") {
		secret = strings.TrimPrefix(secret, "whsec_")
	}

	// Decode the base64 secret
	secretBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return false
	}

	// Parse signature header (format: "v1,g=signature1 v1,g=signature2")
	signatures := strings.Split(signature, " ")
	
	for _, sig := range signatures {
		if strings.HasPrefix(sig, "v1,") {
			sigPart := strings.TrimPrefix(sig, "v1,")
			expectedSig := cv.generateSignature(timestamp, msgId, body, secretBytes)
			
			// Decode the signature from the header
			headerSig, err := base64.StdEncoding.DecodeString(sigPart)
			if err != nil {
				continue
			}
			
			if hmac.Equal(headerSig, expectedSig) {
				return true
			}
		}
	}
	
	return false
}

func (cv *ClerkValidator) generateSignature(timestamp, msgId string, body []byte, secret []byte) []byte {
	// Svix payload format: id.timestamp.payload
	payload := msgId + "." + timestamp + "." + string(body)
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(payload))
	return h.Sum(nil)
}

type ClerkTokenValidator struct {
	publishableKey string
}

func NewClerkTokenValidator(publishableKey string) *ClerkTokenValidator {
	return &ClerkTokenValidator{
		publishableKey: publishableKey,
	}
}

type ClerkTokenClaims struct {
	UserID string
	OrgID  string
	Role   string
}

func (ctv *ClerkTokenValidator) ValidateClerkToken(ctx context.Context, token string) (*ClerkTokenClaims, error) {
	// TODO: Implement proper JWT validation with Clerk's public key
	// For now, return a mock validation
	if token == "" {
		return nil, fmt.Errorf("missing token")
	}

	// Mock validation - in production, validate with Clerk's JWT
	return &ClerkTokenClaims{
		UserID: "user_mock",
		OrgID:  "org_mock", 
		Role:   "admin",
	}, nil
}

func (ctv *ClerkTokenValidator) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		claims, err := ctv.ValidateClerkToken(r.Context(), token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), "clerk_claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (ctv *ClerkTokenValidator) RequireOnboarding(identityService interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("clerk_claims").(*ClerkTokenClaims)
			if !ok {
				http.Error(w, "Missing authentication context", http.StatusUnauthorized)
				return
			}

			// TODO: Check if onboarding is complete using identityService
			// For now, allow all requests
			_ = claims

			next.ServeHTTP(w, r)
		})
	}
}