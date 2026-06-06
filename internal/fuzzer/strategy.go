package fuzzer

type Strategy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var strategies = []Strategy{
	{ID: "valid_baseline", Name: "Valid baseline", Description: "Generate schema-valid request values as a control case."},
	{ID: "min_boundary", Name: "Minimum boundary", Description: "Prefer minimum numeric values from schema constraints."},
	{ID: "max_boundary", Name: "Maximum boundary", Description: "Prefer maximum numeric values from schema constraints."},
	{ID: "zero_value", Name: "Zero value", Description: "Use zero-like numeric and scalar values."},
	{ID: "negative_number", Name: "Negative number", Description: "Probe handlers with negative numeric values."},
	{ID: "large_number", Name: "Large number", Description: "Probe overflow-prone handlers with very large numbers."},
	{ID: "decimal_precision", Name: "Decimal precision", Description: "Use high-precision decimal values for number schemas."},
	{ID: "empty_string", Name: "Empty string", Description: "Send empty strings through string fields and parameters."},
	{ID: "long_string", Name: "Long string", Description: "Send long but deterministic strings to probe length handling."},
	{ID: "unicode_string", Name: "Unicode string", Description: "Send non-ASCII text to probe encoding assumptions."},
	{ID: "sql_probe", Name: "SQL probe", Description: "Send SQL-shaped payload strings to find unsafe query builders."},
	{ID: "xss_probe", Name: "XSS probe", Description: "Send script-shaped payload strings to find unsafe reflection."},
	{ID: "path_traversal", Name: "Path traversal", Description: "Send traversal-shaped payload strings to path-like inputs."},
	{ID: "nullish_string", Name: "Nullish string", Description: "Send strings that often collide with null handling."},
	{ID: "boolean_true", Name: "Boolean true", Description: "Force boolean values to true."},
	{ID: "boolean_false", Name: "Boolean false", Description: "Force boolean values to false."},
	{ID: "enum_first", Name: "Enum first", Description: "Use the first documented enum value."},
	{ID: "enum_last", Name: "Enum last", Description: "Use the last documented enum value."},
	{ID: "required_only", Name: "Required only", Description: "Generate request bodies with required object fields only."},
	{ID: "extra_object_field", Name: "Extra object field", Description: "Add unexpected object properties to JSON bodies."},
	{ID: "empty_array", Name: "Empty array", Description: "Use empty arrays for array schemas."},
	{ID: "single_item_array", Name: "Single item array", Description: "Use one-item arrays for array schemas."},
	{ID: "large_array", Name: "Large array", Description: "Use larger arrays to probe loop and payload handling."},
	{ID: "duplicate_query", Name: "Duplicate query", Description: "Duplicate query parameters to probe parser ambiguity."},
	{ID: "encoded_slash", Name: "Encoded slash", Description: "Use encoded slash-like values in path parameters."},
	{ID: "missing_content_type", Name: "Missing content type", Description: "Send JSON bodies without a Content-Type header."},
	{ID: "invalid_json_body", Name: "Invalid JSON body", Description: "Send malformed JSON for operations that accept JSON."},
	{ID: "empty_json_body", Name: "Empty JSON body", Description: "Send an empty JSON object body."},
	{ID: "large_json_body", Name: "Large JSON body", Description: "Add a large deterministic payload field to JSON objects."},
	{ID: "case_flip_header", Name: "Case-flip header", Description: "Send generated headers with upper-case names."},
}

func Strategies() []Strategy {
	out := make([]Strategy, len(strategies))
	copy(out, strategies)
	return out
}
