package docs

import (
	"bytes"
	"os"
	"testing"
)

func TestEmbeddedOpenAPIMatchesRootDocument(t *testing.T) {
	rootDoc, err := os.ReadFile("../../openapi.yaml")
	if err != nil {
		t.Fatalf("read root openapi.yaml: %v", err)
	}

	if !bytes.Equal(rootDoc, OpenAPIYAML) {
		t.Fatal("embedded OpenAPI document differs from root openapi.yaml")
	}
}
