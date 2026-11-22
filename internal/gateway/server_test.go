package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewServerRoutes(t *testing.T) {
	var gotPath string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	mux, err := NewServer(Config{
		Port:           "8080",
		IAMURL:         upstream.URL,
		CompetitionURL: upstream.URL,
		RealtimeURL:    upstream.URL,
	})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	tests := []struct {
		path string
		want string
	}{
		{path: "/api/v1/iam/auth/login", want: "/api/v1/iam/auth/login"},
		{path: "/api/v1/auth/login", want: "/api/v1/iam/auth/login"},
		{path: "/api/v1/users/search", want: "/api/v1/iam/user/v1/search"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if rec.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want 204", rec.Code)
			}
			if gotPath != tt.want {
				t.Fatalf("upstream path = %q, want %q", gotPath, tt.want)
			}
		})
	}
}

func TestNewServerRejectsInvalidURL(t *testing.T) {
	if _, err := NewServer(Config{IAMURL: "://bad"}); err == nil {
		t.Fatal("expected an error for an invalid IAM URL")
	}
}
