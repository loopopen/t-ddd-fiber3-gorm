package http

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/loopopen/t-ddd-fiber3-gorm/cmd/api/config"
	"github.com/loopopen/t-ddd-fiber3-gorm/docs"
)

func TestNewAppRegistersSwaggerRoutes(t *testing.T) {
	title := docs.SwaggerInfo.Title
	version := docs.SwaggerInfo.Version
	host := docs.SwaggerInfo.Host
	basePath := docs.SwaggerInfo.BasePath
	t.Cleanup(func() {
		docs.SwaggerInfo.Title = title
		docs.SwaggerInfo.Version = version
		docs.SwaggerInfo.Host = host
		docs.SwaggerInfo.BasePath = basePath
	})

	app := NewApp(
		&config.Env{Name: "test"},
		&config.Config{Swagger: config.Swagger{Host: "example.com", BasePath: "/api"}},
		nil,
		nil,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	uiResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/swagger", http.NoBody))
	if err != nil {
		t.Fatalf("request swagger UI: %v", err)
	}
	defer uiResp.Body.Close()
	if uiResp.StatusCode != http.StatusOK {
		t.Fatalf("swagger UI status = %d, want %d", uiResp.StatusCode, http.StatusOK)
	}

	docResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/swagger.json", http.NoBody))
	if err != nil {
		t.Fatalf("request swagger document: %v", err)
	}
	defer docResp.Body.Close()
	if docResp.StatusCode != http.StatusOK {
		t.Fatalf("swagger document status = %d, want %d", docResp.StatusCode, http.StatusOK)
	}
	body, err := io.ReadAll(docResp.Body)
	if err != nil {
		t.Fatalf("read swagger document: %v", err)
	}
	if !strings.Contains(string(body), `"host": "example.com"`) {
		t.Fatalf("swagger document does not contain configured host: %s", body)
	}
}
