package google

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/domain"
	"golang.org/x/oauth2"
	"io"
	"net/http"
)

var ErrStateTokenRevoked = errors.New("state_token_revoked")

type Google struct {
	callbackPort         int
	oauthConfig          *oauth2.Config
	stateTokenRepository StateTokenRepository
}

var _ domain.GoogleAuthGateway = (*Google)(nil)

func (g Google) AuthURL(ctx context.Context) (string, error) {
	token, err := g.stateTokenRepository.NewStateToken(ctx)
	if err != nil {
		return "", fmt.Errorf("new state token: %w", err)
	}

	return g.oauthConfig.AuthCodeURL(token), nil
}

func (g Google) CompleteAuth(ctx context.Context, code, state string) (domain.UserProfile, error) {
	if err := g.stateTokenRepository.ValidateStateToken(ctx, state); err != nil {
		return domain.UserProfile{}, fmt.Errorf("validate state token: %w", err)
	}

	token, err := g.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return domain.UserProfile{}, fmt.Errorf("exchange token: %w", err)
	}

	// NOTE: We are not using the token.RefreshToken here as we don't need to refresh the token
	// It is not required for the current use case

	p, err := getUserProfileInfo(token.AccessToken)
	if err != nil {
		return domain.UserProfile{}, fmt.Errorf("get user info: %w", err)
	}
	return p, nil
}

func getUserProfileInfo(accessToken string) (domain.UserProfile, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return domain.UserProfile{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.UserProfile{}, fmt.Errorf("failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.UserProfile{}, fmt.Errorf("read body: %w", err)
	}

	var user = struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
	}{}
	if err := json.Unmarshal(body, &user); err != nil {
		return domain.UserProfile{}, fmt.Errorf("unmarshal: %w", err)
	}

	b64Picture := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte(user.Picture))

	return domain.UserProfile{
		Email:   user.Email,
		Name:    user.Name,
		Picture: b64Picture,
	}, nil
}
