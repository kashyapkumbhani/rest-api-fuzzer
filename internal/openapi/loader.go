package openapi

import (
	"context"
	"fmt"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
)

type Spec struct {
	Doc     *openapi3.T
	BaseURL string
	Source  string
}

type Loader struct {
	loader *openapi3.Loader
}

func NewLoader() *Loader {
	return &Loader{loader: openapi3.NewLoader()}
}

func (l *Loader) Load(ctx context.Context, source string) (*Spec, error) {
	doc, err := l.load(source)
	if err != nil {
		return nil, err
	}
	if err := doc.Validate(ctx); err != nil {
		return nil, fmt.Errorf("validate OpenAPI document: %w", err)
	}

	baseURL := ""
	if len(doc.Servers) > 0 {
		baseURL = doc.Servers[0].URL
	}
	if baseURL == "" {
		baseURL = "http://127.0.0.1"
	}

	return &Spec{Doc: doc, BaseURL: baseURL, Source: source}, nil
}

func (l *Loader) load(source string) (*openapi3.T, error) {
	u, err := url.Parse(source)
	if err == nil && u.Scheme != "" && u.Host != "" {
		doc, err := l.loader.LoadFromURI(u)
		if err != nil {
			return nil, fmt.Errorf("load remote spec: %w", err)
		}
		return doc, nil
	}

	doc, err := l.loader.LoadFromFile(source)
	if err != nil {
		return nil, fmt.Errorf("load file spec: %w", err)
	}
	return doc, nil
}
