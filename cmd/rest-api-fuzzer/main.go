package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kashyapkumbhani/rest-api-fuzzer/internal/fuzzer"
	"github.com/kashyapkumbhani/rest-api-fuzzer/internal/openapi"
)

func main() {
	var cfg fuzzer.Config
	var specPath string
	var format string
	var headerList multiFlag

	flag.StringVar(&specPath, "spec", "", "Path or URL to an OpenAPI 3.x document")
	flag.StringVar(&cfg.BaseURL, "base-url", "", "Override the server URL from the OpenAPI document")
	flag.Int64Var(&cfg.Seed, "seed", 1337, "Deterministic random seed")
	flag.IntVar(&cfg.CasesPerOperation, "cases", 20, "Number of generated requests per operation")
	flag.DurationVar(&cfg.Timeout, "timeout", 5*time.Second, "Per-request timeout")
	flag.DurationVar(&cfg.SlowThreshold, "slow", 750*time.Millisecond, "Report responses slower than this threshold")
	flag.StringVar(&format, "format", "text", "Report format: text or json")
	flag.Var(&headerList, "header", "Static header to include, e.g. 'Authorization: Bearer token'")
	flag.Parse()

	if specPath == "" {
		fmt.Fprintln(os.Stderr, "missing required -spec")
		flag.Usage()
		os.Exit(2)
	}
	if cfg.CasesPerOperation <= 0 {
		fmt.Fprintln(os.Stderr, "-cases must be greater than zero")
		os.Exit(2)
	}

	cfg.Headers = parseHeaders(headerList)
	loader := openapi.NewLoader()
	spec, err := loader.Load(context.Background(), specPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load spec: %v\n", err)
		os.Exit(1)
	}

	runner := fuzzer.New(spec, cfg)
	report, err := runner.Run(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "fuzz: %v\n", err)
		os.Exit(1)
	}

	switch strings.ToLower(format) {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintf(os.Stderr, "write json: %v\n", err)
			os.Exit(1)
		}
	case "text":
		fmt.Print(report.Text())
	default:
		fmt.Fprintf(os.Stderr, "unsupported -format %q\n", format)
		os.Exit(2)
	}

	if report.HasFailures() {
		os.Exit(1)
	}
}

type multiFlag []string

func (m *multiFlag) String() string {
	return strings.Join(*m, ", ")
}

func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func parseHeaders(values []string) map[string]string {
	headers := make(map[string]string)
	for _, value := range values {
		name, body, ok := strings.Cut(value, ":")
		if !ok {
			continue
		}
		name = strings.TrimSpace(name)
		body = strings.TrimSpace(body)
		if name != "" {
			headers[name] = body
		}
	}
	return headers
}
