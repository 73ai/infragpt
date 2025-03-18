package identityapi

import (
	"context"
	"encoding/json"
	"github.com/priyanshujain/infragpt/services/infragpt/identity"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/generic/httperrors"
	"io"
	"log/slog"
	"net/http"
	"time"
)

func NewHandler(svc identity.Service) http.Handler {
	h := &httpHandler{
		svc: svc,
	}
	h.init()
	return panicHandler(h)
}

type httpHandler struct {
	http.ServeMux
	svc identity.Service
}

func (h *httpHandler) init() {
	h.HandleFunc("POST /v1/identity/create-user", h.createUser())
	h.HandleFunc("POST /v1/identity/login", h.login())
	h.HandleFunc("POST /v1/identity/refresh-credentials", h.refreshCredentials())
	h.HandleFunc("POST /v1/identity/validate-credentials", h.validateCredentials())
	h.HandleFunc("POST /v1/identity/verify-email", h.verifyEmail())
	h.HandleFunc("POST /v1/identity/request-reset-password", h.requestResetPassword())
	h.HandleFunc("POST /v1/identity/validate-reset-password-token", h.validateResetPasswordToken())
	h.HandleFunc("POST /v1/identity/reset-password", h.resetPassword())
	h.HandleFunc("POST /v1/identity/user-sessions", h.userSessions())
	h.HandleFunc("POST /v1/identity/initiate-google-auth", h.initiateGoogleAuth())
	h.HandleFunc("POST /v1/identity/complete-google-auth", h.completeGoogleAuth())
}

func (h *httpHandler) createUser() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ApiHandlerFunc(func(ctx context.Context, x request) (response, error) {
			cmd := identity.CreateUserCommand{
				Email:    x.Email,
				Password: x.Password,
			}
			// get ip from request header
			cmd.IP = r.Header.Get("x-forwarded-for")
			cmd.UserAgent = r.UserAgent()
			cmd.IPCountryISO = r.Header.Get("cf-ipcountry")

			credentials, err := h.svc.CreateUser(ctx, cmd)
			if err != nil {
				slog.Error("error in create user", "err", err)
				return response{}, httperrors.From(err)
			}
			return response{
				AccessToken:  credentials.AccessToken,
				RefreshToken: credentials.RefreshToken,
			}, nil
		})(w, r)
	}
}

func (h *httpHandler) login() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ApiHandlerFunc(func(ctx context.Context, x request) (response, error) {
			cmd := identity.LoginCommand{
				Email:    x.Email,
				Password: x.Password,
			}
			// get ip from request header
			cmd.IP = r.Header.Get("x-forwarded-for")
			cmd.UserAgent = r.UserAgent()
			cmd.IPCountryISO = r.Header.Get("cf-ipcountry")

			credentials, err := h.svc.Login(ctx, cmd)
			if err != nil {
				slog.Error("error in login", "err", err)
				return response{}, err
			}
			return response{
				AccessToken:  credentials.AccessToken,
				RefreshToken: credentials.RefreshToken,
			}, nil
		})
	}
}

func (h *httpHandler) refreshCredentials() http.HandlerFunc {
	type request struct {
		RefreshToken string `json:"refresh_token"`
	}
	type response struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ApiHandlerFunc(func(ctx context.Context, x request) (response, error) {
			cmd := identity.RefreshCredentialsCommand{
				RefreshToken: x.RefreshToken,
			}

			credentials, err := h.svc.RefreshCredentials(ctx, cmd)
			if err != nil {
				slog.Error("error in refresh credentials", "err", err)
				return response{}, httperrors.New(http.StatusUnauthorized, "invalid_credentials",
					"Invalid credentials", nil)
			}
			return response{
				AccessToken:  credentials.AccessToken,
				RefreshToken: credentials.RefreshToken,
			}, nil
		})(w, r)
	}
}

func (h *httpHandler) validateCredentials() http.HandlerFunc {
	type request struct {
	}

	type response struct{}

	return func(w http.ResponseWriter, r *http.Request) {
		ApiHandlerFunc(func(ctx context.Context, x request) (response, error) {
			cmd := identity.ValidateCredentialsQuery{}
			cmd.AuthorizationHeader = r.Header.Get("authorization")

			err := h.svc.ValidateCredentials(ctx, cmd)
			if err != nil {
				slog.Error("error in validate credentials", "err", err)
				return response{}, httperrors.New(http.StatusUnauthorized, "invalid_credentials",
					"Invalid credentials", nil)
			}
			return response{}, nil
		})(w, r)
	}
}

func (h *httpHandler) verifyEmail() http.HandlerFunc {
	type request struct {
		VerificationID string `json:"verification_id"`
	}
	type response struct{}

	return func(w http.ResponseWriter, r *http.Request) {
		ApiHandlerFunc(func(ctx context.Context, x request) (response, error) {
			cmd := identity.VerifyEmailCommand{
				VerificationID: x.VerificationID,
			}

			err := h.svc.VerifyEmail(ctx, cmd)
			if err != nil {
				slog.Error("error in verify email", "err", err)
				return response{}, err
			}
			return response{}, nil
		})(w, r)
	}
}

func (h *httpHandler) requestResetPassword() http.HandlerFunc {
	type request struct {
		Email string `json:"email"`
	}
	type response struct{}

	return func(w http.ResponseWriter, r *http.Request) {
		ApiHandlerFunc(func(ctx context.Context, x request) (response, error) {
			cmd := identity.RequestResetPasswordCommand{
				Email: x.Email,
			}

			err := h.svc.RequestResetPassword(ctx, cmd)
			if err != nil {
				slog.Error("error in request reset password", "err", err)
				return response{}, err
			}
			return response{}, nil
		})(w, r)
	}
}

func (h *httpHandler) validateResetPasswordToken() http.HandlerFunc {
	type request struct {
		ResetID string `json:"reset_id"`
	}
	type response struct{}

	return func(w http.ResponseWriter, r *http.Request) {
		ApiHandlerFunc(func(ctx context.Context, x request) (response, error) {
			cmd := identity.ValidateResetPasswordTokenQuery{
				ResetID: x.ResetID,
			}

			err := h.svc.ValidateResetPasswordToken(ctx, cmd)
			if err != nil {
				slog.Error("error in validate reset password token", "err", err)
				return response{}, err
			}
			return response{}, nil
		})(w, r)
	}
}

func (h *httpHandler) resetPassword() http.HandlerFunc {
	type request struct {
		ResetID  string `json:"reset_id"`
		Password string `json:"password"`
	}
	type response struct{}

	return func(w http.ResponseWriter, r *http.Request) {
		ApiHandlerFunc(func(ctx context.Context, x request) (response, error) {
			cmd := identity.ResetPasswordCommand{
				ResetID:  x.ResetID,
				Password: x.Password,
			}

			err := h.svc.ResetPassword(ctx, cmd)
			if err != nil {
				slog.Error("error in reset password", "err", err)
				return response{}, err
			}
			return response{}, nil
		})(w, r)
	}
}

func (h *httpHandler) userSessions() http.HandlerFunc {
	type request struct {
	}

	type session struct {
		UserID         string `json:"user_id"`
		SessionID      string `json:"session_id"`
		UserAgent      string `json:"user_agent"`
		IPAddress      string `json:"ip_address"`
		IPCountryISO   string `json:"ip_country_iso"`
		LastActivityAt string `json:"last_activity_at"`
		IsExpired      bool   `json:"is_expired"`
	}

	type response struct {
		UserSessions []session `json:"sessions"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ApiHandlerFunc(func(ctx context.Context, x request) (response, error) {
			cmd := identity.UserSessionsQuery{}
			cmd.AuthorizationHeader = r.Header.Get("authorization")

			userSessions, err := h.svc.UserSessions(ctx, cmd)
			if err != nil {
				slog.Error("error in user sessions", "err", err)
				return response{}, err
			}

			var xs []session
			for _, userSession := range userSessions {
				xs = append(xs, session{
					UserID:         userSession.UserID,
					SessionID:      userSession.SessionID,
					UserAgent:      userSession.UserAgent,
					IPAddress:      userSession.IPAddress,
					IPCountryISO:   userSession.IPCountryISO,
					LastActivityAt: userSession.LastActivityAt.Format(time.RFC3339),
					IsExpired:      userSession.IsExpired,
				})
			}

			return response{
				UserSessions: xs,
			}, nil
		})(w, r)
	}
}

func (h *httpHandler) initiateGoogleAuth() http.HandlerFunc {
	type request struct {
	}
	type response struct {
		URL string `json:"url"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ApiHandlerFunc(func(ctx context.Context, x request) (response, error) {
			url, err := h.svc.InitiateGoogleAuth(ctx)
			if err != nil {
				slog.Error("error in initiate google auth", "err", err)
				return response{}, err
			}
			return response{
				URL: url,
			}, nil
		})(w, r)
	}
}

func (h *httpHandler) completeGoogleAuth() http.HandlerFunc {
	type request struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}
	type response struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ApiHandlerFunc(func(ctx context.Context, x request) (response, error) {
			cmd := identity.CompleteGoogleAuthCommand{
				Code:  x.Code,
				State: x.State,
			}

			credentials, err := h.svc.CompleteGoogleAuth(ctx, cmd)
			if err != nil {
				slog.Error("error in complete google auth", "err", err)
				return response{}, err
			}
			return response{
				AccessToken:  credentials.AccessToken,
				RefreshToken: credentials.RefreshToken,
			}, nil
		})(w, r)
	}
}

func ApiHandlerFunc[X any, Y any](api func(
	context.Context, X) (Y, error)) func(http.ResponseWriter, *http.Request) {
	const RequestIDHeader = "x-request-id"

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		request := new(X)
		bodyBytes, err := io.ReadAll(r.Body)

		json.Unmarshal(bodyBytes, request)

		w.Header().Set("Content-Type", "application/json")
		res, err := api(ctx, *request)
		if err != nil {
			slog.Error("error in api handler", "path", r.URL, "request", request, "err", err)
			var httpError = httperrors.From(err)
			w.WriteHeader(httpError.HttpStatus)
			_ = json.NewEncoder(w).Encode(httpError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	}
}

func panicHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("infragpt: panic while handling http request", "recover", r)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}
