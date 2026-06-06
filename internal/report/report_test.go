package report

import "testing"

func TestHasFailuresIgnoresSlowResponses(t *testing.T) {
	r := Report{Findings: []Finding{{Kind: "slow_response"}}}
	if r.HasFailures() {
		t.Fatal("slow responses should be reported without forcing a non-zero exit")
	}
}

func TestHasFailuresReturnsTrueForServerError(t *testing.T) {
	r := Report{Findings: []Finding{{Kind: "server_error"}}}
	if !r.HasFailures() {
		t.Fatal("server errors should fail the run")
	}
}
