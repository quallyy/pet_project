package proxy

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// New creates a reverse proxy to the given upstream URL.
func New(upstreamURL string) http.Handler {
	target, err := url.Parse(upstreamURL)
	if err != nil {
		slog.Error("invalid upstream URL", "url", upstreamURL, "err", err)
		panic(err)
	}

	p := httputil.NewSingleHostReverseProxy(target)

	// return JSON on proxy failure instead of default HTML error page
	p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		slog.Error("proxy error", "upstream", upstreamURL, "path", r.URL.Path, "err", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"upstream service unavailable"}`))
	}

	return p
}