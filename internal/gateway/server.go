package gateway

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	Port           string
	IAMURL         string
	CompetitionURL string
	RealtimeURL    string
}

func NewServer(cfg Config) (*http.ServeMux, error) {
	mux := http.NewServeMux()

	iamURL, err := url.Parse(cfg.IAMURL)
	if err != nil {
		return nil, fmt.Errorf("invalid IAM URL: %w", err)
	}
	iamProxy := httputil.NewSingleHostReverseProxy(iamURL)

	compURL, err := url.Parse(cfg.CompetitionURL)
	if err != nil {
		return nil, fmt.Errorf("invalid competition URL: %w", err)
	}
	compProxy := httputil.NewSingleHostReverseProxy(compURL)

	realtimeURL, err := url.Parse(cfg.RealtimeURL)
	if err != nil {
		return nil, fmt.Errorf("invalid realtime URL: %w", err)
	}
	realtimeProxy := httputil.NewSingleHostReverseProxy(realtimeURL)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	// WebSocket Routes -> Realtime
	mux.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		r.Host = realtimeURL.Host
		realtimeProxy.ServeHTTP(w, r)
	})

	// IAM registers every route under /api/v1/iam. Accept that canonical prefix
	// and rewrite the two legacy prefixes used by older clients.
	forwardToIAM := func(w http.ResponseWriter, r *http.Request) {
		r.Host = iamURL.Host
		iamProxy.ServeHTTP(w, r)
	}
	rewriteToIAM := func(from, to string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = to + strings.TrimPrefix(r.URL.Path, from)
			forwardToIAM(w, r)
		}
	}

	mux.HandleFunc("/api/v1/iam/", forwardToIAM)
	mux.HandleFunc("/api/v1/auth/", rewriteToIAM("/api/v1/", "/api/v1/iam/"))
	mux.HandleFunc("/api/v1/users/", rewriteToIAM("/api/v1/users/", "/api/v1/iam/user/v1/"))

	// Service Routes -> Competition
	mux.HandleFunc("/api/v1/competition/", func(w http.ResponseWriter, r *http.Request) {
		r.Host = compURL.Host
		compProxy.ServeHTTP(w, r)
	})

	return mux, nil
}

func Start(cfg Config) error {
	mux, err := NewServer(cfg)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	log.Printf("Gateway listening on %s", srv.Addr)
	return srv.ListenAndServe()
}
