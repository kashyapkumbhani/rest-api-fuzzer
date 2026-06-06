package fuzzer

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestGeneratorIsDeterministicForSameSeed(t *testing.T) {
	doc := testDocument()
	ops := collectOperations(doc)
	if len(ops) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(ops))
	}

	left := newGenerator(42, "https://api.example.com", map[string]string{"X-Test": "yes"})
	right := newGenerator(42, "https://api.example.com", map[string]string{"X-Test": "yes"})
	strategy := Strategy{ID: "long_string", Name: "Long string"}

	for i := 0; i < 10; i++ {
		a, err := left.build(ops[0], i, strategy)
		if err != nil {
			t.Fatalf("left build: %v", err)
		}
		b, err := right.build(ops[0], i, strategy)
		if err != nil {
			t.Fatalf("right build: %v", err)
		}
		if a.URL != b.URL || string(a.Body) != string(b.Body) {
			t.Fatalf("case %d differed:\n%s\n%s\n%s\n%s", i, a.URL, b.URL, a.Body, b.Body)
		}
	}
}

func TestGeneratorBuildsPathQueryHeadersAndJSONBody(t *testing.T) {
	doc := testDocument()
	op := collectOperations(doc)[0]
	gen := newGenerator(7, "https://api.example.com/v1", map[string]string{"Authorization": "Bearer test"})

	req, err := gen.build(op, 1, Strategy{ID: "valid_baseline", Name: "Valid baseline"})
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if req.Method != "POST" {
		t.Fatalf("method = %q", req.Method)
	}
	if req.URL == "" || req.URL == "https://api.example.com/v1/pets/{id}" {
		t.Fatalf("path parameter was not expanded: %q", req.URL)
	}
	if req.Headers["Authorization"] != "Bearer test" {
		t.Fatalf("static header missing: %#v", req.Headers)
	}
	if req.Headers["Content-Type"] != "application/json" {
		t.Fatalf("content type missing: %#v", req.Headers)
	}
	if len(req.Body) == 0 {
		t.Fatal("expected JSON body")
	}
}

func TestStrategiesExposeLongPublicFuzzerList(t *testing.T) {
	got := Strategies()
	if len(got) < 25 {
		t.Fatalf("expected at least 25 fuzzing strategies, got %d", len(got))
	}
	seen := make(map[string]bool, len(got))
	for _, strategy := range got {
		if strategy.ID == "" || strategy.Name == "" || strategy.Description == "" {
			t.Fatalf("strategy should be fully documented: %#v", strategy)
		}
		if seen[strategy.ID] {
			t.Fatalf("duplicate strategy id %q", strategy.ID)
		}
		seen[strategy.ID] = true
	}
}

func testDocument() *openapi3.T {
	return &openapi3.T{
		OpenAPI: "3.1.0",
		Paths: openapi3.NewPaths(openapi3.WithPath("/pets/{id}", &openapi3.PathItem{
			Post: &openapi3.Operation{
				Parameters: openapi3.Parameters{
					&openapi3.ParameterRef{Value: &openapi3.Parameter{
						Name: "id",
						In:   "path",
						Schema: &openapi3.SchemaRef{Value: openapi3.NewIntegerSchema().
							WithMin(1).
							WithMax(99)},
					}},
					&openapi3.ParameterRef{Value: &openapi3.Parameter{
						Name:   "include",
						In:     "query",
						Schema: &openapi3.SchemaRef{Value: openapi3.NewBoolSchema()},
					}},
				},
				RequestBody: &openapi3.RequestBodyRef{Value: &openapi3.RequestBody{
					Content: openapi3.Content{
						"application/json": &openapi3.MediaType{
							Schema: &openapi3.SchemaRef{Value: openapi3.NewObjectSchema().
								WithProperty("name", openapi3.NewStringSchema()).
								WithProperty("age", openapi3.NewIntegerSchema()).
								WithRequired([]string{"name"})},
						},
					},
				}},
				Responses: openapi3.NewResponses(openapi3.WithStatus(200, &openapi3.ResponseRef{
					Value: openapi3.NewResponse().WithDescription("ok"),
				})),
			},
		})),
	}
}
