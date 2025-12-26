package conformance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// AsyncAPIValidator validates event messages against AsyncAPI schemas
type AsyncAPIValidator struct {
	schemas  map[string]*jsonschema.Schema
	channels map[string]ChannelInfo
	compiler *jsonschema.Compiler
	specPath string
}

// ChannelInfo holds channel metadata
type ChannelInfo struct {
	Name        string
	Address     string
	Description string
	MessageName string
}

// NewAsyncAPIValidator creates a validator from an AsyncAPI spec
func NewAsyncAPIValidator(specPath string) (*AsyncAPIValidator, error) {
	v := &AsyncAPIValidator{
		schemas:  make(map[string]*jsonschema.Schema),
		channels: make(map[string]ChannelInfo),
		compiler: jsonschema.NewCompiler(),
		specPath: specPath,
	}

	if err := v.loadSpec(); err != nil {
		return nil, err
	}

	return v, nil
}

func (v *AsyncAPIValidator) loadSpec() error {
	data, err := os.ReadFile(v.specPath)
	if err != nil {
		return fmt.Errorf("reading spec: %w", err)
	}

	var spec map[string]any
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return fmt.Errorf("parsing spec: %w", err)
	}

	// Parse channels
	if channels, ok := spec["channels"].(map[string]any); ok {
		for name, chDef := range channels {
			if chMap, ok := chDef.(map[string]any); ok {
				info := ChannelInfo{
					Name:        name,
					Address:     getString(chMap, "address"),
					Description: getString(chMap, "description"),
				}
				v.channels[name] = info
			}
		}
	}

	// Parse component schemas - first pass: add all resources
	schemaNames := []string{}
	if components, ok := spec["components"].(map[string]any); ok {
		if schemas, ok := components["schemas"].(map[string]any); ok {
			for name, schemaDef := range schemas {
				if schemaMap, ok := schemaDef.(map[string]any); ok {
					jsonSchema := v.toJSONSchema(schemaMap)
					jsonBytes, err := json.Marshal(jsonSchema)
					if err != nil {
						continue
					}

					schemaID := fmt.Sprintf("synapse://asyncapi/%s", name)
					if err := v.compiler.AddResource(schemaID, bytes.NewReader(jsonBytes)); err != nil {
						return fmt.Errorf("adding schema %s: %w", name, err)
					}
					schemaNames = append(schemaNames, name)
				}
			}
		}
	}

	// Second pass: compile all schemas after all resources are added
	for _, name := range schemaNames {
		schemaID := fmt.Sprintf("synapse://asyncapi/%s", name)
		compiled, err := v.compiler.Compile(schemaID)
		if err != nil {
			return fmt.Errorf("compiling schema %s: %w", name, err)
		}
		v.schemas[name] = compiled
	}

	return nil
}

func (v *AsyncAPIValidator) toJSONSchema(schema map[string]any) map[string]any {
	result := make(map[string]any)
	result["$schema"] = "https://json-schema.org/draft/2020-12/schema"

	for k, val := range schema {
		switch k {
		case "$ref":
			ref := val.(string)
			// Extract schema name from ref like "#/components/schemas/OrderReceivedPayload"
			parts := strings.Split(ref, "/")
			schemaName := parts[len(parts)-1]
			result["$ref"] = fmt.Sprintf("synapse://asyncapi/%s", schemaName)
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

func (v *AsyncAPIValidator) convertProperties(props map[string]any) map[string]any {
	result := make(map[string]any)
	for name, propDef := range props {
		if propMap, ok := propDef.(map[string]any); ok {
			result[name] = v.toJSONSchema(propMap)
		}
	}
	return result
}

func getString(m map[string]any, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// ValidateMessage validates an event message against its schema
func (v *AsyncAPIValidator) ValidateMessage(schemaName string, payload []byte) error {
	schema, ok := v.schemas[schemaName]
	if !ok {
		return fmt.Errorf("schema not found: %s", schemaName)
	}

	var data any
	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("parsing message payload: %w", err)
	}

	if err := schema.Validate(data); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	return nil
}

// Channels returns all defined channels
func (v *AsyncAPIValidator) Channels() map[string]ChannelInfo {
	return v.channels
}

// EventTestResult represents a single event contract test result
type EventTestResult struct {
	Channel string
	Schema  string
	Passed  bool
	Error   string
	Payload string
}

// EventContractTestSuite runs a suite of event contract tests
type EventContractTestSuite struct {
	validator *AsyncAPIValidator
	results   []EventTestResult
}

// NewEventContractTestSuite creates a new event test suite
func NewEventContractTestSuite(specPath string) (*EventContractTestSuite, error) {
	validator, err := NewAsyncAPIValidator(specPath)
	if err != nil {
		return nil, err
	}

	return &EventContractTestSuite{
		validator: validator,
		results:   make([]EventTestResult, 0),
	}, nil
}

// ValidateEvent validates an event payload against a schema
func (s *EventContractTestSuite) ValidateEvent(channel, schema string, payload []byte) EventTestResult {
	result := EventTestResult{
		Channel: channel,
		Schema:  schema,
		Payload: string(payload),
	}

	if err := s.validator.ValidateMessage(schema, payload); err != nil {
		result.Error = err.Error()
	} else {
		result.Passed = true
	}

	s.results = append(s.results, result)
	return result
}

// Results returns all test results
func (s *EventContractTestSuite) Results() []EventTestResult {
	return s.results
}

// Summary returns a summary of test results
func (s *EventContractTestSuite) Summary() (passed, failed int) {
	for _, r := range s.results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}
	return
}

// Validator returns the underlying validator for direct access
func (s *EventContractTestSuite) Validator() *AsyncAPIValidator {
	return s.validator
}
