package infragptapi

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/generic/httperrors"
	"io"
	"log/slog"
	"net/http"
)

func NewHandler(svc infragpt.Service) http.Handler {
	h := &httpHandler{
		svc: svc,
	}
	h.init()
	return panicHandler(h)
}

type httpHandler struct {
	http.ServeMux
	svc infragpt.Service
}

func (h *httpHandler) init() {
	h.HandleFunc("GET /slack", h.completeSlackAuthentication)
}

func (h *httpHandler) completeSlackAuthentication(w http.ResponseWriter, r *http.Request) {
	type request struct{}
	type response struct{}

	code := r.URL.Query().Get("code")

	ApiHandlerFunc(func(ctx context.Context, x request) (response, error) {
		err := h.svc.CompleteSlackAuthentication(ctx, infragpt.CompleteSlackAuthenticationCommand{
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
