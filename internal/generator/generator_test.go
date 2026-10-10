package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"crudding/internal/types"
)

func TestGenerateAllTemplateCombinations(t *testing.T) {
	templates := os.DirFS(filepath.Join("..", "..", "templates"))
	architectures := []string{"monorepo", "camadas", "clean arch", "hexagonal", "4horse"}
	cases := []types.ProjectConfig{}
	for _, framework := range []string{"raw", "fastapi", "django", "flask"} {
		for _, architecture := range architectures {
			cases = append(cases, types.ProjectConfig{ProjectName: "Sample", Language: "python", Framework: framework, Architecture: architecture})
		}
	}
	for _, framework := range []string{"raw", "gin", "fiber", "chi"} {
		for _, architecture := range architectures {
			cases = append(cases, types.ProjectConfig{ProjectName: "Sample", Language: "go", Framework: framework, Architecture: architecture})
		}
	}
	for _, framework := range []string{"raw", "express", "nest", "fastify"} {
		for _, runtime := range []string{"js", "ts"} {
			for _, architecture := range architectures {
				cases = append(cases, types.ProjectConfig{ProjectName: "Sample", Language: "node", Runtime: runtime, Framework: framework, Architecture: architecture})
			}
		}
	}

	for index, config := range cases {
		t.Run(config.Language+"/"+config.Framework+"/"+config.Runtime+"/"+config.Architecture, func(t *testing.T) {
			destination := filepath.Join(t.TempDir(), "Sample")
			if err := Generate(destination, config, templates); err != nil {
				t.Fatalf("Generate() error = %v (case %d)", err, index)
			}
			if config.Language == "node" {
				contents, err := os.ReadFile(filepath.Join(destination, "package.json"))
				if err != nil {
					t.Fatal(err)
				}
				var parsed map[string]any
				if err := json.Unmarshal(contents, &parsed); err != nil {
					t.Fatalf("invalid package.json: %v", err)
				}
			}
		})
	}
}

func TestGenerateRefusesExistingDirectory(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "existing")
	if err := os.Mkdir(destination, 0755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(destination, "keep.txt")
	if err := os.WriteFile(marker, []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}
	config := types.ProjectConfig{ProjectName: "existing", Language: "go", Framework: "raw", Architecture: "camadas"}
	if err := Generate(destination, config, os.DirFS(filepath.Join("..", "..", "templates"))); err == nil {
		t.Fatal("Generate() should reject an existing directory")
	}
	content, err := os.ReadFile(marker)
	if err != nil || string(content) != "keep" {
		t.Fatalf("existing content changed: %q, %v", content, err)
	}
}

func TestGenerateRejectsUnsupportedOptions(t *testing.T) {
	templates := os.DirFS(filepath.Join("..", "..", "templates"))
	for _, config := range []types.ProjectConfig{
		{Language: "go", Framework: "express", Architecture: "camadas"},
		{Language: "node", Runtime: "ruby", Framework: "raw", Architecture: "camadas"},
		{Language: "python", Framework: "raw", Architecture: "unknown"},
	} {
		if err := Generate(filepath.Join(t.TempDir(), "project"), config, templates); err == nil || strings.Contains(err.Error(), "template") {
			t.Errorf("Generate(%+v) should reject unsupported options before rendering, got %v", config, err)
		}
	}
}