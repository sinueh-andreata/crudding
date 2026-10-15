package test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"crudding/internal/generator"
	"crudding/internal/manifest"
	"crudding/internal/project"
)

// templates aponta para a raiz do repositório, onde fica a pasta templates/.
var templates = os.DirFS(filepath.Join("..", ".."))

// Arquivos que devem existir, por linguagem e arquitetura. No Node os caminhos
// não têm extensão: o runtime (.js/.ts) é acrescentado na checagem.
var requiredFiles = map[string]map[string][]string{
	"go": {
		"flat":       {"main.go", "tasks.go", "tasks_test.go"},
		"layered":    {"cmd/api/main.go", "internal/router/router.go", "internal/handlers/handler.go", "internal/services/service.go", "internal/models/item.go"},
		"modular":    {"cmd/api/main.go", "cmd/worker/main.go", "internal/router/router.go", "internal/order/order.go", "internal/order/handler.go", "internal/order/service.go", "internal/product/product.go", "internal/product/service.go", "internal/platform/database/database.go"},
		"hexagonal":  {"cmd/rest/main.go", "cmd/grpc/main.go", "internal/domain/product.go", "internal/application/service.go", "internal/ports/repository.go", "internal/ports/service.go", "internal/adapters/http/handler.go", "internal/adapters/grpc/server.go", "internal/adapters/postgres/repository.go"},
		"clean_arch": {"cmd/api/main.go", "internal/domain/order.go", "internal/domain/errors.go", "internal/application/create_order.go", "internal/application/dto.go", "internal/ports/repository.go", "internal/infrastructure/persistence/repository.go", "internal/infrastructure/http/handler.go", "internal/infrastructure/messaging/producer.go", "pkg/validator/validator.go", "migrations/000001_init.sql"},
	},
	"node": {
		"flat":       {"src/main", "src/server", "src/tasks", "src/tasks.test"},
		"layered":    {"src/main", "src/http/server", "src/routes/index", "src/controllers/user-controller", "src/services/order-service", "src/models/user"},
		"modular":    {"src/main", "src/shared/http/server", "src/shared/database", "src/modules/catalog/index", "src/modules/orders/order-service"},
		"hexagonal":  {"src/entrypoints/main", "src/adapters/http/server", "src/adapters/http/product-controller", "src/domain/product-manager", "src/ports/product-repository"},
		"clean_arch": {"src/frameworks/main", "src/frameworks/http/server", "src/entities/order", "src/use-cases/create-order", "src/adapters/controllers/order-controller"},
	},
	"python": {
		"flat":       {"pyproject.toml", "setup.py", "src/sample/resource_controller.py", "src/sample/resource_service.py", "src/sample/resource_model.py", "src/sample/main.py"},
		"layered":    {"src/app/main.py", "src/app/controllers/resource_controller.py", "src/app/services/resource_service.py", "src/app/models/resource.py"},
		"modular":    {"src/app/main.py", "src/app/resource/controller.py", "src/app/resource/service.py", "src/app/resource/models.py"},
		"hexagonal":  {"src/app/main.py", "src/app/routers/resource_controller.py", "src/app/application/resource_service.py", "src/app/ports/resource_repository.py", "src/app/infrastructure/resource_repository.py", "src/app/domain/resource.py"},
		"clean_arch": {"src/app/main.py", "src/app/frameworks_and_drivers/web/resource_controller.py", "src/app/interface_adapters/controllers/resource_controller.py", "src/app/interface_adapters/presenters/resource_presenter.py", "src/app/application/use_cases/create_resource.py", "src/app/application/use_cases/list_resource.py", "src/app/application/port/outbound/resource_repository.py", "src/app/interface_adapters/persistence/resource_repository.py", "src/app/domain/model/resource.py"},
	},
}

// Arquivos que NÃO devem existir (camadas removidas propositalmente).
var forbiddenFiles = map[string]map[string][]string{
	"go": {
		"layered": {"internal/repositories"},
		"modular": {"internal/order/repository.go", "internal/user/repository.go"},
	},
	"python": {
		"layered":    {"src/app/router", "src/app/controllers/api.py"},
		"modular":    {"src/app/shared/api.py"},
		"hexagonal":  {"src/app/routers/api.py"},
		"clean_arch": {"src/app/frameworks_and_drivers/web/api.py"},
	},
}

// Trechos de código que comprovam que o scaffold genérico foi gerado.
var codeChecks = map[string]map[string][2]string{
	"python": {
		"flat":       {"src/sample/resource_model.py", "class Resource"},
		"layered":    {"src/app/services/resource_service.py", "def list_all"},
		"modular":    {"src/app/resource/service.py", "def list_all"},
		"hexagonal":  {"src/app/application/resource_service.py", "class ResourceService"},
		"clean_arch": {"src/app/application/use_cases/create_resource.py", "class CreateResource"},
	},
}

// allConfigs monta a matriz completa a partir do registry de linguagens.
func allConfigs() []project.Config {
	var configs []project.Config
	for _, language := range generator.Languages() {
		runtimes := language.Runtimes
		if len(runtimes) == 0 {
			runtimes = []string{""}
		}
		for _, framework := range language.Frameworks {
			for _, runtime := range runtimes {
				for _, architecture := range language.Architectures {
					configs = append(configs, project.Config{ProjectName: "Sample", Language: language.Name, Runtime: runtime, Framework: framework, Architecture: architecture})
				}
			}
		}
	}
	return configs
}

func assertFilesExist(t *testing.T, root string, files ...string) {
	t.Helper()
	for _, file := range files {
		if _, err := os.Stat(filepath.Join(root, file)); err != nil {
			t.Errorf("arquivo %s ausente: %v", file, err)
		}
	}
}

func assertFilesAbsent(t *testing.T, root string, files ...string) {
	t.Helper()
	for _, file := range files {
		if _, err := os.Stat(filepath.Join(root, file)); !os.IsNotExist(err) {
			t.Errorf("%s não deveria existir", file)
		}
	}
}

func assertFileContains(t *testing.T, root, file, snippet string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, file))
	if err != nil || !strings.Contains(string(content), snippet) {
		t.Errorf("%s não contém %q: %v", file, snippet, err)
	}
}

func TestGenerateAllTemplateCombinations(t *testing.T) {
	for _, config := range allConfigs() {
		t.Run(config.Language+"/"+config.Framework+"/"+config.Runtime+"/"+config.Architecture, func(t *testing.T) {
			destination := filepath.Join(t.TempDir(), "Sample")
			if err := generator.Generate(destination, config, templates); err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			readme, err := os.ReadFile(filepath.Join(destination, "README.md"))
			if err != nil || len(readme) == 0 {
				t.Fatalf("README.md ausente ou vazio: %v", err)
			}

			files := requiredFiles[config.Language][config.Architecture]
			if len(files) == 0 {
				t.Fatalf("sem arquivos obrigatórios cadastrados para %s/%s", config.Language, config.Architecture)
			}
			if config.Runtime != "" {
				withExt := make([]string, len(files))
				for index, file := range files {
					withExt[index] = file + "." + config.Runtime
				}
				files = withExt
			}
			if _, m, err := manifest.Find(destination); err != nil || m.Config() != config {
				t.Errorf("%s = %+v, %v; want %+v", manifest.FileName, m, err, config)
			}
			assertFilesExist(t, destination, files...)
			assertFilesAbsent(t, destination, forbiddenFiles[config.Language][config.Architecture]...)
			if check, ok := codeChecks[config.Language][config.Architecture]; ok {
				assertFileContains(t, destination, check[0], check[1])
			}

			switch config.Language {
			case "go":
				assertFilesExist(t, destination, "go.mod", "go.sum", ".gitignore")
			case "node":
				contents, err := os.ReadFile(filepath.Join(destination, "package.json"))
				if err != nil {
					t.Fatal(err)
				}
				var parsed map[string]any
				if err := json.Unmarshal(contents, &parsed); err != nil {
					t.Fatalf("invalid package.json: %v", err)
				}
				_, err = os.Stat(filepath.Join(destination, "tsconfig.json"))
				if hasTsconfig := err == nil; hasTsconfig != (config.Runtime == "ts") {
					t.Errorf("tsconfig.json presente = %v para runtime %s", hasTsconfig, config.Runtime)
				}
			case "python":
				assertFilesExist(t, destination, "requirements.txt", "pyproject.toml")
			}

			if config.Language != "go" && !strings.Contains(string(readme), "- Arquitetura: "+config.Architecture) {
				t.Errorf("README não identifica a arquitetura %s", config.Architecture)
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
	config := project.Config{ProjectName: "existing", Language: "go", Framework: "raw", Architecture: "flat"}
	if err := generator.Generate(destination, config, os.DirFS(filepath.Join("..", ".."))); err == nil {
		t.Fatal("Generate() should reject an existing directory")
	}
	content, err := os.ReadFile(marker)
	if err != nil || string(content) != "keep" {
		t.Fatalf("existing content changed: %q, %v", content, err)
	}
}

func TestPythonGenericResourceDispatchers(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 is not installed")
	}
	templates := os.DirFS(filepath.Join("..", ".."))
	cases := []struct {
		architecture string
		module       string
	}{
		{"flat", "src.sample.main"},
		{"layered", "src.app.main"},
		{"modular", "src.app.main"},
		{"hexagonal", "src.app.main"},
		{"clean_arch", "src.app.main"},
	}

	for _, testCase := range cases {
		t.Run(testCase.architecture, func(t *testing.T) {
			destination := filepath.Join(t.TempDir(), "Sample")
			config := project.Config{ProjectName: "Sample", Language: "python", Framework: "raw", Architecture: testCase.architecture}
			if err := generator.Generate(destination, config, templates); err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			script := `from ` + testCase.module + ` import ROUTES

class MockHandler:
    def __init__(self):
        self.status = None
        self.data = None
    def send_json(self, status, data):
        self.status = status
        self.data = data
    def read_json(self):
        return {}
    def send_response(self, status):
        self.status = status
    def end_headers(self):
        pass

assert "resource" in ROUTES, "resource route missing"
handler = MockHandler()
ROUTES["resource"](handler, "GET", [])
assert handler.status == 501, f"expected 501, got {handler.status}"

handler = MockHandler()
ROUTES["resource"](handler, "GET", ["1"])
assert handler.status == 501, f"expected 501, got {handler.status}"
`
			command := exec.Command(python, "-c", script)
			command.Dir = destination
			command.Env = append(os.Environ(), "PYTHONPATH="+destination)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("generated generic resource failed: %v\n%s", err, output)
			}
		})
	}
}

func TestPythonFrameworkEntrypointsCompile(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 is not installed")
	}
	templates := os.DirFS(filepath.Join("..", ".."))
	for _, framework := range []string{"raw", "fastapi", "django", "flask"} {
		t.Run(framework, func(t *testing.T) {
			destination := filepath.Join(t.TempDir(), "Sample")
			config := project.Config{ProjectName: "Sample", Language: "python", Framework: framework, Architecture: "flat"}
			if err := generator.Generate(destination, config, templates); err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			command := exec.Command(python, "-m", "compileall", "-q", destination)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("generated %s entrypoint does not compile: %v\n%s", framework, err, output)
			}
		})
	}
}

func TestGenerateRejectsUnsupportedOptions(t *testing.T) {
	templates := os.DirFS(filepath.Join("..", ".."))
	for _, config := range []project.Config{
		{Language: "go", Framework: "express", Architecture: "flat"},
		{Language: "go", Framework: "raw", Architecture: "monorepo"},
		{Language: "node", Runtime: "ruby", Framework: "raw", Architecture: "layered"},
		{Language: "node", Runtime: "js", Framework: "raw", Architecture: "monorepo"},
		{Language: "python", Framework: "raw", Architecture: "monorepo"},
		{Language: "python", Framework: "raw", Architecture: "unknown"},
	} {
		if err := generator.Generate(filepath.Join(t.TempDir(), "project"), config, templates); err == nil || strings.Contains(err.Error(), "template") {
			t.Errorf("Generate(%+v) should reject unsupported options before rendering, got %v", config, err)
		}
	}
}

func TestGenerateGoVet(t *testing.T) {
	templates := os.DirFS(filepath.Join("..", ".."))
	for _, arch := range []string{"flat", "layered", "modular"} {
		dest := filepath.Join(t.TempDir(), arch)
		cfg := project.Config{
			ProjectName:  arch,
			Language:     "go",
			Framework:    "raw",
			Architecture: arch,
		}
		if err := generator.Generate(dest, cfg, templates); err != nil {
			t.Fatalf("Generate() error = %v for %s", err, arch)
		}
	}
}

func TestGeneratedGoRawScaffoldsCompile(t *testing.T) {
	goCommand, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go is not installed")
	}
	templates := os.DirFS(filepath.Join("..", ".."))
	for _, architecture := range []string{"flat", "layered", "modular", "hexagonal", "clean_arch"} {
		t.Run(architecture, func(t *testing.T) {
			destination := filepath.Join(t.TempDir(), "Sample")
			config := project.Config{ProjectName: "Sample", Language: "go", Framework: "raw", Architecture: architecture}
			if err := generator.Generate(destination, config, templates); err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if smokeTest, ok := goAPISmokeTests[architecture]; ok {
				if err := os.WriteFile(filepath.Join(destination, "api_smoke_test.go"), []byte(smokeTest), 0644); err != nil {
					t.Fatal(err)
				}
			}
			command := exec.Command(goCommand, "test", "./...")
			command.Dir = destination
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("generated Go project does not compile: %v\n%s", err, output)
			}
		})
	}
}

var goAPISmokeTests = map[string]string{
	"layered": `package main
import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"example.com/sample/internal/router"
)
func TestDemoAPI(t *testing.T) {
	handler := router.Setup()
	request := func(method, path, body string) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(method, path, strings.NewReader(body)))
		return response
	}
	created := request(http.MethodPost, "/orders", ` + "`" + `{"userId":1,"items":[{"product":"book","quantity":2,"unitPrice":3.25}]}` + "`" + `)
	if created.Code != http.StatusCreated || !strings.Contains(created.Body.String(), ` + "`" + `"total":6.5` + "`" + `) { t.Fatalf("create = %d %s", created.Code, created.Body.String()) }
	paid := request(http.MethodPost, "/orders/1/pay", "")
	if paid.Code != http.StatusOK || !strings.Contains(paid.Body.String(), ` + "`" + `"status":"paid"` + "`" + `) { t.Fatalf("pay = %d %s", paid.Code, paid.Body.String()) }
	if cancelled := request(http.MethodPost, "/orders/1/cancel", ""); cancelled.Code != http.StatusConflict { t.Fatalf("cancel status = %d", cancelled.Code) }
}`,
	"modular": `package main
import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"example.com/sample/internal/router"
)
func TestDemoAPI(t *testing.T) {
	handler := router.Setup()
	request := func(method, path, body string) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(method, path, strings.NewReader(body)))
		return response
	}
	created := request(http.MethodPost, "/orders", ` + "`" + `{"items":[{"productId":1,"quantity":2}]}` + "`" + `)
	if created.Code != http.StatusCreated || !strings.Contains(created.Body.String(), ` + "`" + `"total":900` + "`" + `) { t.Fatalf("create = %d %s", created.Code, created.Body.String()) }
	product := request(http.MethodGet, "/products/1", "")
	if !strings.Contains(product.Body.String(), ` + "`" + `"stock":8` + "`" + `) { t.Fatalf("stock = %s", product.Body.String()) }
}`,
	"hexagonal": `package main
import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	httpapi "example.com/sample/internal/adapters/http"
)
func TestDemoAPI(t *testing.T) {
	handler := httpapi.NewAPI()
	request := func(method, path, body string) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(method, path, strings.NewReader(body)))
		return response
	}
	discount := request(http.MethodPost, "/products/prod-1/discount", ` + "`" + `{"percent":10}` + "`" + `)
	if discount.Code != http.StatusOK || !strings.Contains(discount.Body.String(), ` + "`" + `"price":288` + "`" + `) { t.Fatalf("discount = %d %s", discount.Code, discount.Body.String()) }
	if invalid := request(http.MethodPost, "/products/prod-1/discount", ` + "`" + `{"percent":100}` + "`" + `); invalid.Code != http.StatusUnprocessableEntity { t.Fatalf("invalid discount status = %d", invalid.Code) }
}`,
	"clean_arch": `package main
import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	httpapi "example.com/sample/internal/infrastructure/http"
)
func TestDemoAPI(t *testing.T) {
	handler := httpapi.NewAPI()
	request := func(method, path, body string) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(method, path, strings.NewReader(body)))
		return response
	}
	created := request(http.MethodPost, "/orders", ` + "`" + `{"customerId":"customer-1","items":[{"sku":"SKU-1","quantity":2,"unitPrice":12.5}]}` + "`" + `)
	if created.Code != http.StatusCreated || !strings.Contains(created.Body.String(), ` + "`" + `"formatted":"R$ 25,00"` + "`" + `) { t.Fatalf("create = %d %s", created.Code, created.Body.String()) }
	orderID := regexp.MustCompile("\\\"id\\\":\\\"([^\\\"]+)\\\"").FindStringSubmatch(created.Body.String())[1]
	paid := request(http.MethodPost, "/orders/"+orderID+"/pay", "")
	if paid.Code != http.StatusOK || !strings.Contains(paid.Body.String(), ` + "`" + `"status":"paid"` + "`" + `) { t.Fatalf("pay = %d %s", paid.Code, paid.Body.String()) }
	if again := request(http.MethodPost, "/orders/"+orderID+"/pay", ""); again.Code != http.StatusConflict { t.Fatalf("second payment status = %d", again.Code) }
}`,
}
