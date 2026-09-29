package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

// TestSchema guards the structure of the JSON schema, which isn't used by the app itself
// but helps users writing their data files in their editors.
func TestSchema(t *testing.T) {
	content, err := os.ReadFile("schema.json")
	if err != nil {
		t.Fatal(err)
	}

	var schema struct {
		Required   []string `json:"required"`
		AnyOf      []map[string][]string
		Properties map[string]map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(content, &schema); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if !reflect.DeepEqual(schema.Required, []string{"participants", "template"}) {
		t.Errorf("unexpected root `required`: %v", schema.Required)
	}
	if _, ok := schema.Properties["required"]; ok {
		t.Errorf("`required` must be a keyword at the root, not a property")
	}
	if schema.Properties["participants"]["minItems"] != float64(2) {
		t.Errorf("expected participants to have `minItems: 2`, got %v", schema.Properties["participants"])
	}

	expectedAnyOf := []map[string][]string{{"required": {"smtp"}}, {"required": {"mailgun"}}}
	if !reflect.DeepEqual(schema.AnyOf, expectedAnyOf) {
		t.Errorf("expected `smtp` or `mailgun` to be required, got %v", schema.AnyOf)
	}
}
