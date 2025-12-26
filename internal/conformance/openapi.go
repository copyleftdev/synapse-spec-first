package conformance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// OpenAPIValidator validates HTTP responses against OpenAPI schemas
type OpenAPIValidator struct {
	schemas    map[string]*jsonschema.Schema
	compiler   *jsonschema.Compiler
	specPath   string
	components map[string]any
}

// NewOpenAPIValidator creates a validator from an OpenAPI spec
func NewOpenAPIValidator(specPath string) (*OpenAPIValidator, error) {
	v := &OpenAPIValidator{
		schemas:  make(map[string]*jsonschema.Schema),
		compiler: jsonschema.NewCompiler(),
		specPath: specPath,
	}

	if err := v.loadSpec(); err != nil {
		return nil, err
	}

	return v, nil
}

func (v *OpenAPIValidator) loadSpec() error {
	data, err := os.ReadFile(v.specPath)
	if err != nil {
		return fmt.Errorf("reading spec: %w", err)
	}

	var spec map[string]any
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return fmt.Errorf("parsing spec: %w", err)
	}

	// Load component schemas from referenced files
	baseDir := filepath.Dir(v.specPath)
	if err := v.loadComponentSchemas(baseDir); err != nil {
		return err
	}

	return nil
}

func (v *OpenAPIValidator) loadComponentSchemas(baseDir string) error {
	schemasDir := filepath.Join(baseDir, "components", "schemas")

	files, err := os.ReadDir(schemasDir)
	if err != nil {
		return fmt.Errorf("reading schemas dir: %w", err)
	}

	// First pass: add all schema resources
	schemaNames := []string{}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}
		if file.Name() == "_index.yaml" {
			continue
		}

		filePath := filepath.Join(schemasDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading schema file %s: %w", file.Name(), err)
		}

		var schemas map[string]any
		if err := yaml.Unmarshal(data, &schemas); err != nil {
			return fmt.Errorf("parsing schema file %s: %w", file.Name(), err)
		}

		for name, schema := range schemas {
			schemaMap, ok := schema.(map[string]any)
			if !ok {
				continue
			}

			// Convert to JSON Schema format
			jsonSchema := v.toJSONSchema(schemaMap)
			jsonBytes, err := json.Marshal(jsonSchema)
			if err != nil {
				continue
			}

			schemaID := fmt.Sprintf("synapse://schemas/%s", name)
			if err := v.compiler.AddResource(schemaID, bytes.NewReader(jsonBytes)); err != nil {
				return fmt.Errorf("adding schema %s: %w", name, err)
			}
			schemaNames = append(schemaNames, name)
		}
	}

	// Second pass: compile all schemas after all resources are added
	for _, name := range schemaNames {
		schemaID := fmt.Sprintf("synapse://schemas/%s", name)
		compiled, err := v.compiler.Compile(schemaID)
		if err != nil {
			return fmt.Errorf("compiling schema %s: %w", name, err)
		}
		v.schemas[name] = compiled
	}

	return nil
}

func (v *OpenAPIValidator) toJSONSchema(schema map[string]any) map[string]any {
	result := make(map[string]any)
	result["$schema"] = "https://json-schema.org/draft/2020-12/schema"

	for k, val := range schema {
		switch k {
		case "$ref":
			// Convert OpenAPI ref to our schema ID
			ref := val.(string)
			parts := strings.Split(ref, "/")
			schemaName := parts[len(parts)-1]
			result["$ref"] = fmt.Sprintf("synapse://schemas/%s", schemaName)
		case "properties":
			if props, ok := val.(map[string]any); ok {
				result["properties"] = v.convertProperties(props)
			}
		case "items":
			if items, ok := val.(map[string]any); ok {
				result["items"] = v.toJSONSchema(items)
			}
		case "allOf":
			if allOf, ok := val.([]any); ok {
				converted := make([]any, len(allOf))
				for i, item := range allOf {
					if itemMap, ok := item.(map[string]any); ok {
						converted[i] = v.toJSONSchema(itemMap)
					}
				}
				result["allOf"] = converted
			}
		default:
			result[k] = val
		}
	}

	return result
}

func (v *OpenAPIValidator) convertProperties(props map[string]any) map[string]any {
	result := make(map[string]any)
	for name, propDef := range props {
		if propMap, ok := propDef.(map[string]any); ok {
			result[name] = v.toJSONSchema(propMap)
		}
	}
	return result
}

// ValidateResponse validates an HTTP response against the expected schema
func (v *OpenAPIValidator) ValidateResponse(schemaName string, body []byte) error {
	schema, ok := v.schemas[schemaName]
	if !ok {
		return fmt.Errorf("schema not found: %s", schemaName)
	}

	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("parsing response body: %w", err)
	}

	if err := schema.Validate(data); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	return nil
}

// ValidateHandler validates that a handler's response conforms to the schema
func (v *OpenAPIValidator) ValidateHandler(
	handler http.HandlerFunc,
	method, path string,
	body io.Reader,
	expectedStatus int,
	responseSchema string,
) error {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != expectedStatus {
		return fmt.Errorf("expected status %d, got %d: %s", expectedStatus, rec.Code, rec.Body.String())
	}

	if responseSchema != "" && rec.Body.Len() > 0 {
		if err := v.ValidateResponse(responseSchema, rec.Body.Bytes()); err != nil {
			return fmt.Errorf("response validation failed for %s %s: %w", method, path, err)
		}
	}

	return nil
}

// ContractTestResult represents a single contract test result
type ContractTestResult struct {
	Endpoint    string
	Method      string
	Schema      string
	Passed      bool
	Error       string
	RequestBody string
	Response    string
}

// ContractTestSuite runs a suite of contract tests
type ContractTestSuite struct {
	validator *OpenAPIValidator
	results   []ContractTestResult
}

// NewContractTestSuite creates a new test suite
func NewContractTestSuite(specPath string) (*ContractTestSuite, error) {
	validator, err := NewOpenAPIValidator(specPath)
	if err != nil {
		return nil, err
	}

	return &ContractTestSuite{
		validator: validator,
		results:   make([]ContractTestResult, 0),
	}, nil
}

// RunTest runs a single contract test
func (s *ContractTestSuite) RunTest(
	ctx context.Context,
	client *http.Client,
	baseURL, method, path string,
	body []byte,
	expectedStatus int,
	responseSchema string,
) ContractTestResult {
	result := ContractTestResult{
		Endpoint:    path,
		Method:      method,
		Schema:      responseSchema,
		RequestBody: string(body),
	}

	url := baseURL + path
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		result.Error = fmt.Sprintf("creating request: %v", err)
		s.results = append(s.results, result)
		return result
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("executing request: %v", err)
		s.results = append(s.results, result)
		return result
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	result.Response = string(respBody)

	if resp.StatusCode != expectedStatus {
		result.Error = fmt.Sprintf("expected status %d, got %d", expectedStatus, resp.StatusCode)
		s.results = append(s.results, result)
		return result
	}

	if responseSchema != "" && len(respBody) > 0 {
		if err := s.validator.ValidateResponse(responseSchema, respBody); err != nil {
			result.Error = fmt.Sprintf("schema validation: %v", err)
			s.results = append(s.results, result)
			return result
		}
	}

	result.Passed = true
	s.results = append(s.results, result)
	return result
}

// Results returns all test results
func (s *ContractTestSuite) Results() []ContractTestResult {
	return s.results
}

// Summary returns a summary of test results
func (s *ContractTestSuite) Summary() (passed, failed int) {
	for _, r := range s.results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}
	return
}
