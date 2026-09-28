package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var signedBodyHeaders = map[string]string{
	"Content-Digest":  "sha-256=:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=:",
	"Idempotency-Key": "550e8400-e29b-41d4-a716-446655440000",
}

func TestSelfOnlyEngineRoutesRegisteredWithoutLegacyRoutes(t *testing.T) {
	handler := Handler(Unimplemented{})

	tests := []struct {
		name    string
		method  string
		path    string
		headers map[string]string
		want    int
	}{
		{
			name:   "self config route is registered",
			method: http.MethodGet,
			path:   "/engines/config",
			want:   http.StatusNotImplemented,
		},
		{
			name:   "self telemetry log get route is registered",
			method: http.MethodGet,
			path:   "/engines/telemetry/log",
			want:   http.StatusNotImplemented,
		},
		{
			name:    "self telemetry log post route is registered",
			method:  http.MethodPost,
			path:    "/engines/telemetry/log",
			headers: signedBodyHeaders,
			want:    http.StatusNotImplemented,
		},
		{
			name:   "self telemetry traces get route is registered",
			method: http.MethodGet,
			path:   "/engines/telemetry/traces",
			want:   http.StatusNotImplemented,
		},
		{
			name:    "self telemetry traces post route is registered",
			method:  http.MethodPost,
			path:    "/engines/telemetry/traces",
			headers: signedBodyHeaders,
			want:    http.StatusNotImplemented,
		},
		{
			name:   "legacy instance config route is not registered",
			method: http.MethodGet,
			path:   "/engines/config/test-instance",
			want:   http.StatusNotFound,
		},
		{
			name:   "root telemetry log get route is not registered",
			method: http.MethodGet,
			path:   "/telemetry/test-instance/log",
			want:   http.StatusNotFound,
		},
		{
			name:   "root telemetry log post route is not registered",
			method: http.MethodPost,
			path:   "/telemetry/test-instance/log",
			want:   http.StatusNotFound,
		},
		{
			name:   "root telemetry traces get route is not registered",
			method: http.MethodGet,
			path:   "/telemetry/test-instance/traces",
			want:   http.StatusNotFound,
		},
		{
			name:   "root telemetry traces post route is not registered",
			method: http.MethodPost,
			path:   "/telemetry/test-instance/traces",
			want:   http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			for name, value := range tt.headers {
				req.Header.Set(name, value)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Fatalf("%s %s returned %d, want %d", tt.method, tt.path, rec.Code, tt.want)
			}
		})
	}
}

func TestBodyRoutesRequireDigestAndIdempotencyHeaders(t *testing.T) {
	handler := Handler(Unimplemented{})

	for _, path := range []string{
		"/engines/telemetry/log",
		"/engines/telemetry/traces",
	} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("POST %s returned %d, want %d", path, rec.Code, http.StatusBadRequest)
			}
		})
	}
}
