package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEngineRoutesRegisteredWithoutRootTelemetryOrConfigRoutes(t *testing.T) {
	handler := Handler(Unimplemented{})

	tests := []struct {
		name   string
		method string
		path   string
		want   int
	}{
		{
			name:   "prefixed config route is registered",
			method: http.MethodGet,
			path:   "/engines/config/test-instance",
			want:   http.StatusNotImplemented,
		},
		{
			name:   "prefixed telemetry log get route is registered",
			method: http.MethodGet,
			path:   "/engines/telemetry/test-instance/log",
			want:   http.StatusNotImplemented,
		},
		{
			name:   "prefixed telemetry log post route is registered",
			method: http.MethodPost,
			path:   "/engines/telemetry/test-instance/log",
			want:   http.StatusNotImplemented,
		},
		{
			name:   "prefixed telemetry traces get route is registered",
			method: http.MethodGet,
			path:   "/engines/telemetry/test-instance/traces",
			want:   http.StatusNotImplemented,
		},
		{
			name:   "prefixed telemetry traces post route is registered",
			method: http.MethodPost,
			path:   "/engines/telemetry/test-instance/traces",
			want:   http.StatusNotImplemented,
		},
		{
			name:   "root config route is not registered",
			method: http.MethodGet,
			path:   "/config/test-instance",
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
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Fatalf("%s %s returned %d, want %d", tt.method, tt.path, rec.Code, tt.want)
			}
		})
	}
}
