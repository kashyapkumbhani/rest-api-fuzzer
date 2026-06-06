package fuzzer

import "time"

type Config struct {
	BaseURL           string
	Seed              int64
	CasesPerOperation int
	Timeout           time.Duration
	SlowThreshold     time.Duration
	Headers           map[string]string
}
