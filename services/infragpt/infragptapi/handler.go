package infragptapi

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/identityapi"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/generic/httperrors"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc"
	"io"
	"log/slog"
	"net/http"
	"time"
)

func NewHandler(svc infragpt.Service, identityService infragpt.IdentityService, identityConfig identitysvc.Config) http.Handler {
	h := &httpHandler{
		svc: svc,
	}
	h.init(identityService, identityConfig)
	return corsHandler(loggingHandler(panicHandler(h)))
}

type httpHandler struct {
	http.ServeMux
	svc infragpt.Service
}

func (h *httpHandler) init(identityService infragpt.IdentityService, identityConfig identitysvc.Config) {
	h.HandleFunc("GET /slack", h.completeSlackAuthentication)
	h.HandleFunc("POST /reply", h.sendReply)

	// Identity API routes
	identityHandler := identityapi.NewHandler(identityService)
	clerkValidator := identityapi.NewClerkValidator(identityConfig.Clerk.WebhookSecret)
	tokenValidator := identityapi.NewClerkTokenValidator(identityConfig.Clerk.PublishableKey)

	// Webhook endpoint (no auth required)
	h.Handle("POST /webhooks/clerk", clerkValidator.VerifyWebhookSignature(http.HandlerFunc(identityHandler.HandleClerkWebhook)))

	// Protected API endpoints
	h.Handle("POST /api/v1/organizations/get", tokenValidator.RequireAuth(http.HandlerFunc(identityHandler.GetOrganization)))
	h.Handle("POST /api/v1/organizations/metadata/set", tokenValidator.RequireAuth(http.HandlerFunc(identityHandler.SetOrganizationMetadata)))
}

func (h *httpHandler) completeSlackAuthentication(w http.ResponseWriter, r *http.Request) {
	type request struct{}
	type response struct{}

	code := r.URL.Query().Get("code")

	ApiHandlerFunc(func(ctx context.Context, x request) (response, error) {
		err := h.svc.CompleteSlackIntegration(ctx, infragpt.CompleteSlackIntegrationCommand{
			BusinessID: uuid.New().String(),
			Code:       code,
		})
		if err != nil {
			slog.Error("error in complete slack authentication", "err", err)
			return response{}, err
		}
		return response{}, nil
	})(w, r)
}

func (h *httpHandler) sendReply(w http.ResponseWriter, r *http.Request) {
	type request struct {
		ConversationID string `json:"conversation_id"`
		Message        string `json:"message"`
	}
	type response struct{}

	ApiHandlerFunc(func(ctx context.Context, req request) (response, error) {
		err := h.svc.SendReply(ctx, infragpt.SendReplyCommand{
			ConversationID: req.ConversationID,
			Message:        req.Message,
		})
		if err != nil {
			slog.Error("error sending reply", "err", err)
			return response{}, err
		}
		return response{}, nil
	})(w, r)
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

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.body != nil && len(b) < 1024 { // Only capture small responses
		rw.body.Write(b)
	}
	return rw.ResponseWriter.Write(b)
}

func loggingHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Read and log request body
		var requestBody []byte
		if r.Body != nil {
			requestBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create response writer wrapper
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:          &bytes.Buffer{},
		}

		// Log incoming request
		slog.Info("Request started",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"content_length", r.ContentLength,
			"auth_header_present", r.Header.Get("Authorization") != "",
			"request_body", string(requestBody),
		)

		// Process request
		h.ServeHTTP(rw, r)

		// Log response
		duration := time.Since(start)
		slog.Info("Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status_code", rw.statusCode,
			"duration_ms", duration.Milliseconds(),
			"response_body", rw.body.String(),
		)
	})
}

func corsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
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
