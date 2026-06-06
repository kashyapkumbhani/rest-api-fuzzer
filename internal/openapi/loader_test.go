package openapi

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesFirstServerAsBaseURL(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "openapi.yaml")
	err := os.WriteFile(specPath, []byte(`openapi: 3.0.3
info:
  title: Test
  version: 0.1.0
servers:
  - url: https://api.example.com/v1
paths:
  /health:
    get:
      responses:
        "200":
          description: ok
`), 0o600)
	if err != nil {
		t.Fatalf("write spec: %v", err)
	}

	spec, err := NewLoader().Load(context.Background(), specPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if spec.BaseURL != "https://api.example.com/v1" {
		t.Fatalf("base url = %q", spec.BaseURL)
	}
}
