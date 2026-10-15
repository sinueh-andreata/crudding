package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"crudding/internal/generator"
	"crudding/internal/manifest"
	"crudding/internal/project"
)

var pythonEntryScripts = map[string]string{
	"flat":       "src/sample/main.py",
	"layered":    "src/app/main.py",
	"modular":    "src/app/main.py",
	"hexagonal":  "src/app/main.py",
	"clean_arch": "src/app/main.py",
}

func generateProject(t *testing.T, config project.Config) (string, manifest.Manifest) {
	t.Helper()
	destination := filepath.Join(t.TempDir(), "Sample")
	if err := generator.Generate(destination, config, templates); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	root, m, err := manifest.Find(destination)
	if err != nil {
		t.Fatalf("manifest.Find() error = %v", err)
	}
	return root, m
}

func TestAddEndpointPython(t *testing.T) {
	python, pythonErr := exec.LookPath("python3")
	language, _ := generator.Lookup("python")
	for _, architecture := range language.Architectures {
		t.Run(architecture, func(t *testing.T) {
			entryScript, ok := pythonEntryScripts[architecture]
			if !ok {
				t.Fatalf("sem entryScript cadastrado para %s", architecture)
			}
			root, m := generateProject(t, project.Config{ProjectName: "Sample", Language: "python", Framework: "fastapi", Architecture: architecture})

			for _, name := range []string{"order_item", "invoice"} {
				result, err := generator.AddEndpoint(root, m, name, templates)
				if err != nil {
					t.Fatalf("AddEndpoint(%s) error = %v", name, err)
				}
				if len(result.Created) == 0 || !slices.Equal(result.Modified, []string{entryScript}) {
					t.Fatalf("AddEndpoint(%s) = %+v", name, result)
				}
				assertFilesExist(t, root, result.Created...)
			}
			assertFileContains(t, root, entryScript, `app.include_router(order_item_router, prefix="/order-item", tags=["OrderItem"])`)
			assertFileContains(t, root, entryScript, "# crudding:routes")

			before, err := os.ReadFile(filepath.Join(root, entryScript))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := generator.AddEndpoint(root, m, "invoice", templates); err == nil || !strings.Contains(err.Error(), "já existe") {
				t.Errorf("AddEndpoint repetido: err = %v", err)
			}
			if after, _ := os.ReadFile(filepath.Join(root, entryScript)); string(after) != string(before) {
				t.Error("AddEndpoint repetido alterou o main.py")
			}

			if pythonErr != nil {
				t.Skip("python3 is not installed")
			}
			command := exec.Command(python, "-m", "compileall", "-q", root)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("projeto após add endpoint não compila: %v\n%s", err, output)
			}
		})
	}
}

func TestAddEndpointUnsupportedLanguages(t *testing.T) {
	for _, config := range []project.Config{
		{ProjectName: "Sample", Language: "go", Framework: "gin", Architecture: "layered"},
		{ProjectName: "Sample", Language: "node", Runtime: "ts", Framework: "express", Architecture: "modular"},
	} {
		root, m := generateProject(t, config)
		if _, err := generator.AddEndpoint(root, m, "invoice", templates); err == nil || !strings.Contains(err.Error(), "não suportado") {
			t.Errorf("AddEndpoint(%s) err = %v", config.Language, err)
		}
	}
}

func TestAddEndpointWithoutMarkerChangesNothing(t *testing.T) {
	root, m := generateProject(t, project.Config{ProjectName: "Sample", Language: "python", Framework: "raw", Architecture: "modular"})
	mainPath := filepath.Join(root, "src", "app", "main.py")
	content, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	withoutMarker := strings.Replace(string(content), "# crudding:routes\n", "", 1)
	if err := os.WriteFile(mainPath, []byte(withoutMarker), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := generator.AddEndpoint(root, m, "invoice", templates); err == nil || !strings.Contains(err.Error(), "# crudding:routes") {
		t.Fatalf("AddEndpoint() err = %v", err)
	}
	assertFilesAbsent(t, root, "src/app/invoice")
	if after, _ := os.ReadFile(mainPath); string(after) != withoutMarker {
		t.Error("main.py foi alterado apesar do erro")
	}
}

func TestAddEndpointRejectsInvalidManifestAndKeywords(t *testing.T) {
	root, m := generateProject(t, project.Config{ProjectName: "Sample", Language: "python", Framework: "raw", Architecture: "layered"})
	if _, err := generator.AddEndpoint(root, m, "class", templates); err == nil {
		t.Error("palavra reservada deveria ser rejeitada")
	}
	m.Architecture = "spaghetti"
	if _, err := generator.AddEndpoint(root, m, "invoice", templates); err == nil {
		t.Error("arquitetura inválida no manifesto deveria ser rejeitada")
	}
}
