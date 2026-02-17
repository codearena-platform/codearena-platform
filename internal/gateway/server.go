package gateway

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Config struct {
	Port           string
	IAMURL         string
	CompetitionURL string
	RealtimeURL    string
}

func NewServer(cfg Config) *http.ServeMux {
	mux := http.NewServeMux()

	iamURL, _ := url.Parse(cfg.IAMURL)
	iamProxy := httputil.NewSingleHostReverseProxy(iamURL)

	compURL, _ := url.Parse(cfg.CompetitionURL)
	compProxy := httputil.NewSingleHostReverseProxy(compURL)

	realtimeURL, _ := url.Parse(cfg.RealtimeURL)
	realtimeProxy := httputil.NewSingleHostReverseProxy(realtimeURL)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	// WebSocket Routes -> Realtime
	mux.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		r.Host = realtimeURL.Host
		realtimeProxy.ServeHTTP(w, r)
	})

	// Service Routes -> IAM
	forwardToIAM := func(w http.ResponseWriter, r *http.Request) {
		r.Host = iamURL.Host
		iamProxy.ServeHTTP(w, r)
	}

	mux.HandleFunc("/api/v1/auth/", forwardToIAM)
	mux.HandleFunc("/api/v1/users/", forwardToIAM)

	// Service Routes -> Competition
	mux.HandleFunc("/api/v1/competition/", func(w http.ResponseWriter, r *http.Request) {
		r.Host = compURL.Host
		compProxy.ServeHTTP(w, r)
	})

	return mux
}

func Start(cfg Config) {
	mux := NewServer(cfg)
	log.Printf("Gateway listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
