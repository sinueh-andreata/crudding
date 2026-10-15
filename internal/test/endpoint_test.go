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

// Módulo com dispatch() e o recurso inicial que deve continuar registrado após
// o add endpoint, por arquitetura Python.
var pythonAPIs = map[string]struct{ api, module, resourceRoute string }{
	"flat":       {"src/sample/modulo.py", "src.sample.modulo", "/resource"},
	"layered":    {"src/app/router/api.py", "src.app.router.api", "/resource"},
	"modular":    {"src/app/shared/api.py", "src.app.shared.api", "/resource"},
	"hexagonal":  {"src/app/routers/api.py", "src.app.routers.api", "/resource"},
	"clean_arch": {"src/app/frameworks_and_drivers/web/api.py", "src.app.frameworks_and_drivers.web.api", "/resource"},
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
			api, ok := pythonAPIs[architecture]
			if !ok {
				t.Fatalf("sem API cadastrada para %s", architecture)
			}
			root, m := generateProject(t, project.Config{ProjectName: "Sample", Language: "python", Framework: "fastapi", Architecture: architecture})

			for _, name := range []string{"order_item", "invoice"} {
				result, err := generator.AddEndpoint(root, m, name, templates)
				if err != nil {
					t.Fatalf("AddEndpoint(%s) error = %v", name, err)
				}
				if len(result.Created) == 0 || !slices.Equal(result.Modified, []string{api.api}) {
					t.Fatalf("AddEndpoint(%s) = %+v", name, result)
				}
				assertFilesExist(t, root, result.Created...)
			}
			assertFileContains(t, root, api.api, `if parts and parts[0] == "order-item":`)
			assertFileContains(t, root, api.api, "# crudding:routes")

			before, err := os.ReadFile(filepath.Join(root, api.api))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := generator.AddEndpoint(root, m, "invoice", templates); err == nil || !strings.Contains(err.Error(), "já existe") {
				t.Errorf("AddEndpoint repetido: err = %v", err)
			}
			if after, _ := os.ReadFile(filepath.Join(root, api.api)); string(after) != string(before) {
				t.Error("AddEndpoint repetido alterou o módulo de API")
			}

			if pythonErr != nil {
				t.Skip("python3 is not installed")
			}
			script := `from ` + api.module + ` import dispatch
for route in ("/order-item", "/invoice"):
    assert dispatch("GET", route)["status"] == 501, route
    assert dispatch("POST", route, {"x": 1})["status"] == 501, route
    assert dispatch("GET", route + "/1")["status"] == 501, route
    assert dispatch("PATCH", route + "/1", {"x": 1})["status"] == 501, route
    assert dispatch("DELETE", route + "/1")["status"] == 501, route
    assert dispatch("GET", route + "/1/extra")["status"] == 404, route
assert dispatch("GET", "` + api.resourceRoute + `")["status"] == 501`
			command := exec.Command(python, "-c", script)
			command.Dir = root
			command.Env = append(os.Environ(), "PYTHONPATH="+root)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("endpoint gerado não responde como esperado: %v\n%s", err, output)
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
	apiPath := filepath.Join(root, "src", "app", "shared", "api.py")
	content, err := os.ReadFile(apiPath)
	if err != nil {
		t.Fatal(err)
	}
	withoutMarker := strings.Replace(string(content), "# crudding:routes\n", "", 1)
	if err := os.WriteFile(apiPath, []byte(withoutMarker), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := generator.AddEndpoint(root, m, "invoice", templates); err == nil || !strings.Contains(err.Error(), "# crudding:routes") {
		t.Fatalf("AddEndpoint() err = %v", err)
	}
	assertFilesAbsent(t, root, "src/app/invoice")
	if after, _ := os.ReadFile(apiPath); string(after) != withoutMarker {
		t.Error("api.py foi alterado apesar do erro")
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
