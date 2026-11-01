package docs

import (
	"slices"
	"testing"

	"github.com/goccy/go-yaml"

	"calculator-api/internal/domain"
)

type openAPI struct {
	OpenAPI    string                    `yaml:"openapi"`
	Paths      map[string]map[string]any `yaml:"paths"`
	Components struct {
		Schemas struct {
			Operation struct {
				Enum []string `yaml:"enum"`
			} `yaml:"Operation"`
		} `yaml:"schemas"`
	} `yaml:"components"`
}

func loadSpec(t *testing.T) openAPI {
	t.Helper()
	var doc openAPI
	if err := yaml.Unmarshal(spec, &doc); err != nil {
		t.Fatalf("openapi.yaml is not valid YAML: %v", err)
	}
	return doc
}

// The spec is hand-written, so guard it against drifting from the code.
func TestSpecMatchesDomainOperations(t *testing.T) {
	doc := loadSpec(t)

	var want []string
	for _, op := range domain.Operations {
		want = append(want, string(op))
	}
	got := doc.Components.Schemas.Operation.Enum

	if !slices.Equal(got, want) {
		t.Errorf("Operation enum = %v, want %v (update openapi.yaml)", got, want)
	}
}

func TestSpecDocumentsRoutes(t *testing.T) {
	doc := loadSpec(t)

	if doc.OpenAPI == "" {
		t.Error("missing openapi version")
	}
	for path, method := range map[string]string{"/api/calculate": "post", "/api/health": "get"} {
		if _, ok := doc.Paths[path][method]; !ok {
			t.Errorf("spec does not document %s %s", method, path)
		}
	}
}
