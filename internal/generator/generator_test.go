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
	templates := os.DirFS(filepath.Join("..", ".."))
	legacyArchitectures := []string{"monorepo", "camadas", "clean arch", "hexagonal", "4horse"}
	golangArchitectures := []string{"flat", "layered", "modular", "hexagonal", "clean arch"}

	cases := []types.ProjectConfig{}
	for _, framework := range []string{"raw", "fastapi", "django", "flask"} {
		for _, architecture := range legacyArchitectures {
			cases = append(cases, types.ProjectConfig{ProjectName: "Sample", Language: "python", Framework: framework, Architecture: architecture})
		}
	}
	for _, framework := range []string{"raw", "gin", "fiber", "chi"} {
		for _, architecture := range golangArchitectures {
			cases = append(cases, types.ProjectConfig{ProjectName: "Sample", Language: "go", Framework: framework, Architecture: architecture})
		}
	}
	for _, framework := range []string{"raw", "express", "nest", "fastify"} {
		for _, runtime := range []string{"js", "ts"} {
			for _, architecture := range legacyArchitectures {
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

			readmePath := filepath.Join(destination, "README.md")
			readmeContent, err := os.ReadFile(readmePath)
			if err != nil {
				t.Fatalf("README.md não encontrado para %+v: %v", config, err)
			}
			if len(readmeContent) == 0 {
				t.Fatalf("README.md está vazio para %+v", config)
			}

			if config.Language == "go" {
				for _, requiredFile := range []string{"go.mod", "go.sum", ".gitignore"} {
					if _, err := os.Stat(filepath.Join(destination, requiredFile)); err != nil {
						t.Errorf("arquivo obrigatório %s não encontrado em %+v", requiredFile, config)
					}
				}

				switch config.Architecture {
				case "flat":
					for _, f := range []string{"main.go", "parse.go", "parse_test.go", "report.go"} {
						if _, err := os.Stat(filepath.Join(destination, f)); err != nil {
							t.Errorf("arquivo flat %s ausente: %v", f, err)
						}
					}
				case "layered":
					for _, f := range []string{"cmd/api/main.go", "internal/handlers/handler.go", "internal/services/service.go", "internal/repositories/repository.go", "internal/models/item.go"} {
						if _, err := os.Stat(filepath.Join(destination, f)); err != nil {
							t.Errorf("arquivo layered %s ausente: %v", f, err)
						}
					}
				case "modular":
					for _, f := range []string{"cmd/api/main.go", "cmd/worker/main.go", "internal/order/order.go", "internal/order/handler.go", "internal/order/service.go", "internal/order/repository.go", "internal/user/user.go", "internal/user/handler.go", "internal/user/repository.go", "internal/platform/database/database.go", "internal/platform/auth/auth.go"} {
						if _, err := os.Stat(filepath.Join(destination, f)); err != nil {
							t.Errorf("arquivo modular %s ausente: %v", f, err)
						}
					}
				case "hexagonal":
					for _, f := range []string{"cmd/rest/main.go", "cmd/grpc/main.go", "internal/domain/product.go", "internal/domain/service.go", "internal/ports/repository.go", "internal/ports/service.go", "internal/adapters/http/handler.go", "internal/adapters/grpc/server.go", "internal/adapters/postgres/repository.go"} {
						if _, err := os.Stat(filepath.Join(destination, f)); err != nil {
							t.Errorf("arquivo hexagonal %s ausente: %v", f, err)
						}
					}
				case "clean arch":
					for _, f := range []string{"cmd/api/main.go", "internal/domain/order.go", "internal/domain/errors.go", "internal/application/create_order.go", "internal/application/dto.go", "internal/ports/repository.go", "internal/infrastructure/persistence/repository.go", "internal/infrastructure/http/handler.go", "internal/infrastructure/messaging/producer.go", "pkg/validator/validator.go", "migrations/000001_init.sql"} {
						if _, err := os.Stat(filepath.Join(destination, f)); err != nil {
							t.Errorf("arquivo clean arch %s ausente: %v", f, err)
						}
					}
				}
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
	config := types.ProjectConfig{ProjectName: "existing", Language: "go", Framework: "raw", Architecture: "flat"}
	if err := Generate(destination, config, os.DirFS(filepath.Join("..", ".."))); err == nil {
		t.Fatal("Generate() should reject an existing directory")
	}
	content, err := os.ReadFile(marker)
	if err != nil || string(content) != "keep" {
		t.Fatalf("existing content changed: %q, %v", content, err)
	}
}

func TestGenerateRejectsUnsupportedOptions(t *testing.T) {
	templates := os.DirFS(filepath.Join("..", ".."))
	for _, config := range []types.ProjectConfig{
		{Language: "go", Framework: "express", Architecture: "flat"},
		{Language: "go", Framework: "raw", Architecture: "monorepo"},
		{Language: "node", Runtime: "ruby", Framework: "raw", Architecture: "camadas"},
		{Language: "python", Framework: "raw", Architecture: "unknown"},
	} {
		if err := Generate(filepath.Join(t.TempDir(), "project"), config, templates); err == nil || strings.Contains(err.Error(), "template") {
			t.Errorf("Generate(%+v) should reject unsupported options before rendering, got %v", config, err)
		}
	}
}