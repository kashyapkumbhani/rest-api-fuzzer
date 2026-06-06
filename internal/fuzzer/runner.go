package fuzzer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/kashyapkumbhani/rest-api-fuzzer/internal/openapi"
	"github.com/kashyapkumbhani/rest-api-fuzzer/internal/report"
	"github.com/valyala/fasthttp"
)

type Runner struct {
	spec   *openapi.Spec
	config Config
	client *fasthttp.Client
}

func New(spec *openapi.Spec, config Config) *Runner {
	if config.BaseURL == "" {
		config.BaseURL = spec.BaseURL
	}
	if config.CasesPerOperation == 0 {
		config.CasesPerOperation = 1
	}
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}
	if config.SlowThreshold == 0 {
		config.SlowThreshold = 750 * time.Millisecond
	}
	return &Runner{
		spec:   spec,
		config: config,
		client: &fasthttp.Client{
			ReadTimeout:  config.Timeout,
			WriteTimeout: config.Timeout,
		},
	}
}

func (r *Runner) Run(ctx context.Context) (report.Report, error) {
	started := time.Now()
	ops := collectOperations(r.spec.Doc)
	gen := newGenerator(r.config.Seed, r.config.BaseURL, r.config.Headers)
	out := report.Report{
		Seed:       r.config.Seed,
		StartedAt:  started,
		Operations: len(ops),
		Fuzzers:    reportFuzzers(),
	}

	for _, op := range ops {
		for _, strategy := range Strategies() {
			for i := 0; i < r.config.CasesPerOperation; i++ {
				select {
				case <-ctx.Done():
					return out, ctx.Err()
				default:
				}
				req, err := gen.build(op, i, strategy)
				if err != nil {
					out.Findings = append(out.Findings, r.finding("generation_error", op, req, 0, 0, err.Error(), ""))
					continue
				}
				out.Requests++
				findings := r.execute(op, req)
				out.Findings = append(out.Findings, findings...)
			}
		}
	}

	out.Duration = time.Since(started)
	return out, nil
}

func reportFuzzers() []report.Fuzzer {
	available := Strategies()
	out := make([]report.Fuzzer, len(available))
	for i, strategy := range available {
		out[i] = report.Fuzzer{
			ID:          strategy.ID,
			Name:        strategy.Name,
			Description: strategy.Description,
		}
	}
	return out
}

func (r *Runner) execute(op operation, generated generatedRequest) []report.Finding {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.Header.SetMethod(generated.Method)
	req.SetRequestURI(generated.URL)
	for name, value := range generated.Headers {
		req.Header.Set(name, value)
	}
	if len(generated.Body) > 0 {
		req.SetBody(generated.Body)
	}

	start := time.Now()
	err := r.client.DoTimeout(req, resp, r.config.Timeout)
	elapsed := time.Since(start)
	if err != nil {
		return []report.Finding{r.finding("request_error", op, generated, 0, elapsed, err.Error(), "")}
	}

	status := resp.StatusCode()
	body := string(resp.Body())
	var findings []report.Finding
	if status >= 500 {
		findings = append(findings, r.finding("server_error", op, generated, status, elapsed, "endpoint returned a 5xx response", body))
	}
	if elapsed >= r.config.SlowThreshold {
		findings = append(findings, r.finding("slow_response", op, generated, status, elapsed, fmt.Sprintf("response exceeded %s", r.config.SlowThreshold), body))
	}
	if status < 500 {
		if violation := r.responseViolation(op.Op, status, resp); violation != "" {
			findings = append(findings, r.finding("schema_violation", op, generated, status, elapsed, violation, body))
		}
	}
	return findings
}

func (r *Runner) responseViolation(op *openapi3.Operation, status int, resp *fasthttp.Response) string {
	if op == nil || op.Responses == nil {
		return ""
	}
	responseRef := op.Responses.Status(status)
	if responseRef == nil {
		responseRef = op.Responses.Default()
	}
	if responseRef == nil || responseRef.Value == nil {
		return fmt.Sprintf("status %d is not documented", status)
	}

	contentType := string(resp.Header.ContentType())
	if strings.Contains(contentType, "json") {
		mt := responseRef.Value.Content.Get("application/json")
		if mt == nil || mt.Schema == nil {
			return "JSON response is missing an application/json schema"
		}
		var value any
		if err := json.Unmarshal(resp.Body(), &value); err != nil {
			return fmt.Sprintf("response is not valid JSON: %v", err)
		}
		err := mt.Schema.Value.VisitJSON(value, openapi3.MultiErrors())
		if err != nil {
			return err.Error()
		}
	}
	return ""
}

func (r *Runner) finding(kind string, op operation, req generatedRequest, status int, elapsed time.Duration, message string, responseBody string) report.Finding {
	return report.Finding{
		Kind:       kind,
		Method:     op.Method,
		Path:       op.Path,
		Fuzzer:     req.Strategy.ID,
		StatusCode: status,
		Duration:   elapsed,
		Message:    message,
		Request: report.Request{
			URL:     req.URL,
			Headers: req.Headers,
			Body:    string(req.Body),
		},
		Response: responseBody,
	}
}
