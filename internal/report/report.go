package report

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Report struct {
	Seed       int64         `json:"seed"`
	StartedAt  time.Time     `json:"started_at"`
	Duration   time.Duration `json:"duration"`
	Operations int           `json:"operations"`
	Requests   int           `json:"requests"`
	Findings   []Finding     `json:"findings"`
}

type Finding struct {
	Kind       string        `json:"kind"`
	Method     string        `json:"method"`
	Path       string        `json:"path"`
	StatusCode int           `json:"status_code,omitempty"`
	Duration   time.Duration `json:"duration"`
	Message    string        `json:"message"`
	Request    Request       `json:"request"`
	Response   string        `json:"response,omitempty"`
}

type Request struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

func (r Report) HasFailures() bool {
	for _, finding := range r.Findings {
		if finding.Kind == "server_error" || finding.Kind == "schema_violation" || finding.Kind == "request_error" {
			return true
		}
	}
	return false
}

func (r Report) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "REST API Fuzzer report\n")
	fmt.Fprintf(&b, "Seed: %d\n", r.Seed)
	fmt.Fprintf(&b, "Operations: %d\n", r.Operations)
	fmt.Fprintf(&b, "Requests: %d\n", r.Requests)
	fmt.Fprintf(&b, "Duration: %s\n", r.Duration.Truncate(time.Millisecond))
	fmt.Fprintf(&b, "Findings: %d\n\n", len(r.Findings))

	if len(r.Findings) == 0 {
		b.WriteString("No findings.\n")
		return b.String()
	}

	sort.SliceStable(r.Findings, func(i, j int) bool {
		if r.Findings[i].Kind != r.Findings[j].Kind {
			return r.Findings[i].Kind < r.Findings[j].Kind
		}
		if r.Findings[i].Path != r.Findings[j].Path {
			return r.Findings[i].Path < r.Findings[j].Path
		}
		return r.Findings[i].Method < r.Findings[j].Method
	})

	for _, finding := range r.Findings {
		fmt.Fprintf(&b, "[%s] %s %s", finding.Kind, finding.Method, finding.Path)
		if finding.StatusCode != 0 {
			fmt.Fprintf(&b, " -> HTTP %d", finding.StatusCode)
		}
		fmt.Fprintf(&b, " in %s\n", finding.Duration.Truncate(time.Millisecond))
		fmt.Fprintf(&b, "  %s\n", finding.Message)
		fmt.Fprintf(&b, "  %s\n", finding.Request.URL)
		if finding.Request.Body != "" {
			fmt.Fprintf(&b, "  body: %s\n", finding.Request.Body)
		}
		if finding.Response != "" {
			fmt.Fprintf(&b, "  response: %s\n", compact(finding.Response, 220))
		}
	}
	return b.String()
}

func compact(value string, max int) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) <= max {
		return value
	}
	return value[:max] + "..."
}
