package rest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"portfolio/internal/core/domain"
	"strings"
	"testing"
)

type serviceSpy struct {
	postMessageCalls int
	lastMessage      *domain.Message
	postMessageErr   error
}

func (s *serviceSpy) PostMessage(_ context.Context, msg *domain.Message) error {
	s.postMessageCalls++
	s.lastMessage = msg
	return s.postMessageErr
}

func (s *serviceSpy) VisitLog(*domain.MetaReq) {}

func (s *serviceSpy) IPFetcher(context.Context, string) (*domain.MetaIP, error) {
	return nil, errors.New("not implemented")
}

func (s *serviceSpy) PostQuery(context.Context, *domain.Query) ([]*domain.Log, error) {
	return nil, errors.New("not implemented")
}

type loggerNoop struct{}

func (loggerNoop) Info(string, ...any)  {}
func (loggerNoop) Error(string, ...any) {}
func (loggerNoop) Debug(string, ...any) {}

func TestPostMessage(t *testing.T) {
	t.Run("success redirects and forwards request id", func(t *testing.T) {
		svc := &serviceSpy{}
		h := &RestHandler{svc: svc, logger: loggerNoop{}}

		req := httptest.NewRequest(http.MethodPost, "/messenger", strings.NewReader(`{"name":"a","email":"b","text":"c"}`))
		req = req.WithContext(withRequestID(req.Context(), "rid-123"))
		rr := httptest.NewRecorder()

		h.PostMessage(rr, req)

		res := rr.Result()
		if res.StatusCode != http.StatusSeeOther {
			t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusSeeOther)
		}
		if res.Header.Get("Location") != "https://flyluman.github.io" {
			t.Fatalf("Location header = %q", res.Header.Get("Location"))
		}
		if svc.postMessageCalls != 1 {
			t.Fatalf("PostMessage calls = %d, want 1", svc.postMessageCalls)
		}
		if svc.lastMessage == nil || svc.lastMessage.RequestID != "rid-123" {
			t.Fatalf("forwarded message = %+v", svc.lastMessage)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		svc := &serviceSpy{}
		h := &RestHandler{svc: svc, logger: loggerNoop{}}

		req := httptest.NewRequest(http.MethodPost, "/messenger", strings.NewReader(`{invalid`))
		rr := httptest.NewRecorder()

		h.PostMessage(rr, req)

		if rr.Result().StatusCode != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rr.Result().StatusCode, http.StatusInternalServerError)
		}
		if svc.postMessageCalls != 0 {
			t.Fatalf("PostMessage calls = %d, want 0", svc.postMessageCalls)
		}
	})

	t.Run("service error returns 500 without redirect", func(t *testing.T) {
		svc := &serviceSpy{postMessageErr: errors.New("db error")}
		h := &RestHandler{svc: svc, logger: loggerNoop{}}

		req := httptest.NewRequest(http.MethodPost, "/messenger", strings.NewReader(`{"name":"a","email":"b","text":"c"}`))
		rr := httptest.NewRecorder()

		h.PostMessage(rr, req)

		res := rr.Result()
		if res.StatusCode != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusInternalServerError)
		}
		if loc := res.Header.Get("Location"); loc != "" {
			t.Fatalf("unexpected redirect location: %q", loc)
		}
	})
}
