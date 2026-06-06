package fuzzer

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type operation struct {
	Method string
	Path   string
	Op     *openapi3.Operation
}

type generatedRequest struct {
	Method  string
	Path    string
	URL     string
	Headers map[string]string
	Body    []byte
}

type generator struct {
	rand    *rand.Rand
	baseURL string
	headers map[string]string
}

func newGenerator(seed int64, baseURL string, headers map[string]string) *generator {
	return &generator{
		rand:    rand.New(rand.NewSource(seed)),
		baseURL: strings.TrimRight(baseURL, "/"),
		headers: headers,
	}
}

func collectOperations(doc *openapi3.T) []operation {
	var operations []operation
	for path, item := range doc.Paths.Map() {
		for method, op := range item.Operations() {
			if op == nil {
				continue
			}
			operations = append(operations, operation{
				Method: strings.ToUpper(method),
				Path:   path,
				Op:     op,
			})
		}
	}
	sort.Slice(operations, func(i, j int) bool {
		if operations[i].Path != operations[j].Path {
			return operations[i].Path < operations[j].Path
		}
		return operations[i].Method < operations[j].Method
	})
	return operations
}

func (g *generator) build(op operation, index int) (generatedRequest, error) {
	path := op.Path
	query := make(url.Values)
	headers := cloneMap(g.headers)

	for _, paramRef := range op.Op.Parameters {
		if paramRef == nil || paramRef.Value == nil {
			continue
		}
		param := paramRef.Value
		value := g.valueForSchema(param.Schema, index)
		text := scalarToString(value)
		switch param.In {
		case "path":
			path = strings.ReplaceAll(path, "{"+param.Name+"}", url.PathEscape(text))
		case "query":
			query.Add(param.Name, text)
		case "header":
			headers[param.Name] = text
		}
	}

	u := g.baseURL + path
	if encoded := query.Encode(); encoded != "" {
		u += "?" + encoded
	}

	var body []byte
	if op.Op.RequestBody != nil && op.Op.RequestBody.Value != nil {
		payload := g.requestBody(op.Op.RequestBody.Value, index)
		if payload != nil {
			encoded, err := json.Marshal(payload)
			if err != nil {
				return generatedRequest{}, fmt.Errorf("encode request body: %w", err)
			}
			body = encoded
			if _, ok := headers["Content-Type"]; !ok {
				headers["Content-Type"] = "application/json"
			}
		}
	}

	return generatedRequest{
		Method:  op.Method,
		Path:    op.Path,
		URL:     u,
		Headers: headers,
		Body:    body,
	}, nil
}

func (g *generator) requestBody(body *openapi3.RequestBody, index int) any {
	if body == nil || body.Content == nil {
		return nil
	}
	if mt := body.Content.Get("application/json"); mt != nil {
		return g.valueForSchema(mt.Schema, index)
	}
	for contentType, mt := range body.Content {
		if strings.Contains(contentType, "json") && mt != nil {
			return g.valueForSchema(mt.Schema, index)
		}
	}
	return nil
}

func (g *generator) valueForSchema(ref *openapi3.SchemaRef, index int) any {
	if ref == nil || ref.Value == nil {
		return g.mutatedString(index)
	}
	schema := ref.Value

	if len(schema.Enum) > 0 {
		return schema.Enum[g.rand.Intn(len(schema.Enum))]
	}
	if schema.Example != nil {
		return mutateExample(schema.Example, index)
	}
	if schema.Default != nil && index%4 == 0 {
		return schema.Default
	}

	switch schema.Type.Permits("object") {
	case true:
		obj := make(map[string]any)
		names := make([]string, 0, len(schema.Properties))
		for name := range schema.Properties {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			child := schema.Properties[name]
			obj[name] = g.valueForSchema(child, index)
		}
		for _, required := range schema.Required {
			if _, ok := obj[required]; !ok {
				obj[required] = g.mutatedString(index)
			}
		}
		if len(obj) == 0 && schema.AdditionalProperties.Schema != nil {
			obj["fuzz"] = g.valueForSchema(schema.AdditionalProperties.Schema, index)
		}
		return obj
	}

	if schema.Type.Permits("array") {
		count := 1 + g.rand.Intn(3)
		items := make([]any, count)
		for i := range items {
			items[i] = g.valueForSchema(schema.Items, index+i)
		}
		return items
	}
	if schema.Type.Permits("integer") {
		return g.integer(schema, index)
	}
	if schema.Type.Permits("number") {
		return float64(g.integer(schema, index)) + 0.25
	}
	if schema.Type.Permits("boolean") {
		return index%2 == 0
	}
	return g.string(schema, index)
}

func (g *generator) integer(schema *openapi3.Schema, index int) int64 {
	min := int64(-100)
	max := int64(100)
	if schema.Min != nil {
		min = int64(math.Ceil(*schema.Min))
	}
	if schema.Max != nil {
		max = int64(math.Floor(*schema.Max))
	}
	if max < min {
		max = min
	}
	switch index % 5 {
	case 0:
		return min
	case 1:
		return max
	case 2:
		return 0
	default:
		return min + int64(g.rand.Intn(int(max-min+1)))
	}
}

func (g *generator) string(schema *openapi3.Schema, index int) string {
	if schema.Pattern != "" && index%3 == 0 {
		return "pattern-" + strconv.Itoa(index)
	}
	if schema.Format == "email" {
		return fmt.Sprintf("fuzz-%d@example.com", index)
	}
	if schema.Format == "uuid" {
		return fmt.Sprintf("00000000-0000-4000-8000-%012d", int64(index)%1000000000000)
	}
	if schema.Format == "date" {
		return "2026-06-06"
	}
	if schema.Format == "date-time" {
		return "2026-06-06T00:00:00Z"
	}
	return g.mutatedString(index)
}

func (g *generator) mutatedString(index int) string {
	candidates := []string{
		"fuzz",
		"Fuzz Case " + strconv.Itoa(index),
		"",
		strings.Repeat("a", 64+g.rand.Intn(96)),
		"../etc/passwd",
		"' OR '1'='1",
		"<script>alert(1)</script>",
	}
	return candidates[index%len(candidates)]
}

func scalarToString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprint(typed)
	}
}

func mutateExample(value any, index int) any {
	switch typed := value.(type) {
	case string:
		if index%3 == 0 {
			return typed + "-fuzz"
		}
		return typed
	case float64:
		return typed + float64(index%5)
	case int:
		return typed + index%5
	case bool:
		return index%2 == 0
	case map[string]any:
		next := make(map[string]any, len(typed))
		for key, child := range typed {
			next[key] = mutateExample(child, index)
		}
		return next
	default:
		return value
	}
}

func cloneMap(source map[string]string) map[string]string {
	out := make(map[string]string, len(source))
	for key, value := range source {
		out[key] = value
	}
	return out
}
