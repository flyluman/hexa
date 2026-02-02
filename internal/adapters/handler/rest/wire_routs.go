package rest

import (
	"net/http"
)

func (h *RestHandler) WireRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /",
		chain(
			http.HandlerFunc(h.GetRoot),
			h.BaseMiddleware,
			h.RateLimit,
			h.VisitLog,
		),
	)

	mux.Handle("GET /health",
		chain(
			http.HandlerFunc(h.GetHealth),
			h.BaseMiddleware,
			h.RateLimit,
		),
	)

	mux.Handle("POST /query",
		chain(
			http.HandlerFunc(h.PostQuery),
			h.BaseMiddleware,
			h.RateLimit,
		),
	)

	mux.Handle("POST /messenger",
		chain(
			http.HandlerFunc(h.PostMessage),
			h.BaseMiddleware,
			h.RateLimit,
			h.VisitLog,
		),
	)

	mux.Handle("GET /whoami",
		chain(
			http.HandlerFunc(h.GetWhoAmI),
			h.BaseMiddleware,
			h.RateLimit,
			h.VisitLog,
		),
	)

	// mux.HandleFunc("GET /debug/pprof/", pprof.Index)
	// mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
	// mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
	// mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
	// mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)
	// mux.Handle("GET /debug/pprof/goroutine", pprof.Handler("goroutine"))
	// mux.Handle("GET /debug/pprof/heap", pprof.Handler("heap"))
	// mux.Handle("GET /debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	// mux.Handle("GET /debug/pprof/block", pprof.Handler("block"))
	// mux.Handle("GET /debug/pprof/mutex", pprof.Handler("mutex"))

	return mux
}
