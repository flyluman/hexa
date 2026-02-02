package rest

import (
	"net"
	"net/http"
	"portfolio/internal/core/domain"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

func extractIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	ip := r.RemoteAddr
	host, _, err := net.SplitHostPort(ip)
	if err != nil {
		return ip
	}
	return host
}

func (h *RestHandler) BaseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = ulid.Make().String()
		}

		w.Header().Set("X-Request-ID", id)

		ctx := withRequestID(r.Context(), id)
		next.ServeHTTP(w, r.WithContext(ctx))

		h.logger.Info("request completed",
			"request_id", id,
			"method", r.Method,
			"path", r.URL.Path,
			"latency", time.Since(start).String(),
		)
	})
}

func (h *RestHandler) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)

		if !h.rl.Allow(ip) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *RestHandler) VisitLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := getRequestID(r.Context())

		h.svc.VisitLog(&domain.MetaReq{
			RequestID: id,
			IP:        extractIP(r),
			Path:      r.URL.Path,
			Useragent: r.UserAgent(),
		})

		next.ServeHTTP(w, r)
	})
}

func chain(h http.Handler, m ...func(http.Handler) http.Handler) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}
