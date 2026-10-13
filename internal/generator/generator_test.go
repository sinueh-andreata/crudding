package generator

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"crudding/internal/types"
)

func TestGenerateAllTemplateCombinations(t *testing.T) {
	templates := os.DirFS(filepath.Join("..", ".."))
	pythonArchitectures := []string{"flat", "layered", "modular", "hexagonal", "clean_arch"}
	nodeArchitectures := []string{"flat", "layered", "modular", "hexagonal", "clean_arch"}
	golangArchitectures := []string{"flat", "layered", "modular", "hexagonal", "clean arch"}

	cases := []types.ProjectConfig{}
	for _, framework := range []string{"raw", "fastapi", "django", "flask"} {
		for _, architecture := range pythonArchitectures {
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
			for _, architecture := range nodeArchitectures {
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
					for _, f := range []string{"main.go", "tasks.go", "tasks_test.go"} {
						if _, err := os.Stat(filepath.Join(destination, f)); err != nil {
							t.Errorf("arquivo flat %s ausente: %v", f, err)
						}
					}
				case "layered":
					for _, f := range []string{"cmd/api/main.go", "internal/router/router.go", "internal/handlers/handler.go", "internal/services/service.go", "internal/models/item.go"} {
						if _, err := os.Stat(filepath.Join(destination, f)); err != nil {
							t.Errorf("arquivo layered %s ausente: %v", f, err)
						}
					}
					if _, err := os.Stat(filepath.Join(destination, "internal/repositories")); !os.IsNotExist(err) {
						t.Errorf("camada de repositories não deveria existir em layered")
					}
				case "modular":
					for _, f := range []string{"cmd/api/main.go", "cmd/worker/main.go", "internal/router/router.go", "internal/order/order.go", "internal/order/handler.go", "internal/order/service.go", "internal/product/product.go", "internal/product/service.go", "internal/platform/database/database.go"} {
						if _, err := os.Stat(filepath.Join(destination, f)); err != nil {
							t.Errorf("arquivo modular %s ausente: %v", f, err)
						}
					}
					if _, err := os.Stat(filepath.Join(destination, "internal/order/repository.go")); !os.IsNotExist(err) {
						t.Errorf("repository.go não deveria existir em modular order")
					}
					if _, err := os.Stat(filepath.Join(destination, "internal/user/repository.go")); !os.IsNotExist(err) {
						t.Errorf("repository.go não deveria existir em modular user")
					}
				case "hexagonal":
					for _, f := range []string{"cmd/rest/main.go", "cmd/grpc/main.go", "internal/domain/product.go", "internal/application/service.go", "internal/ports/repository.go", "internal/ports/service.go", "internal/adapters/http/handler.go", "internal/adapters/grpc/server.go", "internal/adapters/postgres/repository.go"} {
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
				requiredFiles := map[string][]string{
					"flat":       {"src/main", "src/server", "src/tasks", "src/tasks.test"},
					"layered":    {"src/main", "src/http/server", "src/routes/index", "src/controllers/user-controller", "src/services/order-service", "src/models/user"},
					"modular":    {"src/main", "src/shared/http/server", "src/shared/database", "src/modules/catalog/index", "src/modules/orders/order-service"},
					"hexagonal":  {"src/entrypoints/main", "src/adapters/http/server", "src/adapters/http/product-controller", "src/domain/product-manager", "src/ports/product-repository"},
					"clean_arch": {"src/frameworks/main", "src/frameworks/http/server", "src/entities/order", "src/use-cases/create-order", "src/adapters/controllers/order-controller"},
				}
				for _, requiredFile := range requiredFiles[config.Architecture] {
					if _, err := os.Stat(filepath.Join(destination, requiredFile+"."+config.Runtime)); err != nil {
						t.Errorf("arquivo Node %s.%s ausente: %v", requiredFile, config.Runtime, err)
					}
				}
				_, err = os.Stat(filepath.Join(destination, "tsconfig.json"))
				if hasTsconfig := err == nil; hasTsconfig != (config.Runtime == "ts") {
					t.Errorf("tsconfig.json presente = %v para runtime %s", hasTsconfig, config.Runtime)
				}
				if !strings.Contains(string(readmeContent), "- Arquitetura: "+config.Architecture) {
					t.Errorf("README não identifica a arquitetura %s", config.Architecture)
				}
			}

			if config.Language == "python" {
				if !strings.Contains(string(readmeContent), "- Arquitetura: "+config.Architecture) {
					t.Errorf("README não identifica a arquitetura %s", config.Architecture)
				}
				requiredFiles := map[string][]string{
					"flat":       {"pyproject.toml", "setup.py", "src/sample/modulo.py", "src/sample/main.py", "tests/test_modulo.py"},
					"layered":    {"src/app/controllers/user_controller.py", "src/app/models/errors.py", "src/app/models/order.py", "src/app/router/api.py", "src/app/router/router.py"},
					"modular":    {"src/app/catalog/service.py", "src/app/orders/views.py", "src/app/shared/api.py", "src/app/shared/errors.py"},
					"hexagonal":  {"src/app/domain/entities.py", "src/app/application/product_service.py", "src/app/ports/product_repository.py", "src/app/infrastructure/in_memory_product_repository.py", "src/app/routers/api.py", "src/app/main.py"},
					"clean_arch": {"src/app/domain/model/order.py", "src/app/application/use_cases/create_order.py", "src/app/application/use_cases/pay_order.py", "src/app/application/port/outbound/order_repository.py", "src/app/interface_adapters/controllers/order_controller.py", "src/app/frameworks_and_drivers/web/api.py", "src/app/bootstrap.py"},
				}
				for _, requiredFile := range requiredFiles[config.Architecture] {
					if _, err := os.Stat(filepath.Join(destination, requiredFile)); err != nil {
						t.Errorf("arquivo Python %s ausente para arquitetura %s: %v", requiredFile, config.Architecture, err)
					}
				}
				if config.Architecture == "flat" {
					testContents, err := os.ReadFile(filepath.Join(destination, "tests/test_modulo.py"))
					if err != nil || !strings.Contains(string(testContents), "def test_task_store_create_complete_and_remove") {
						t.Errorf("teste do módulo flat não foi gerado corretamente: %v", err)
					}
				}
				codeChecks := map[string]struct {
					path    string
					snippet string
				}{
					"flat":       {"src/sample/modulo.py", "class TaskStore"},
					"layered":    {"src/app/router/api.py", "def dispatch"},
					"modular":    {"src/app/catalog/service.py", "def add_product"},
					"hexagonal":  {"src/app/routers/api.py", "def dispatch"},
					"clean_arch": {"src/app/application/use_cases/create_order.py", "class CreateOrder"},
				}
				codeCheck := codeChecks[config.Architecture]
				code, err := os.ReadFile(filepath.Join(destination, codeCheck.path))
				if err != nil || !strings.Contains(string(code), codeCheck.snippet) {
					t.Errorf("código funcional não gerado em %s para arquitetura %s: %v", codeCheck.path, config.Architecture, err)
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

func TestPythonDispatchersMatchNodeDemos(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 is not installed")
	}
	templates := os.DirFS(filepath.Join("..", ".."))
	cases := []struct {
		architecture string
		script       string
	}{
		{"flat", `from src.sample.modulo import dispatch
created = dispatch("POST", "/tasks", {"title": "test"})["body"]
assert created["done"] is False and "createdAt" in created
assert dispatch("PATCH", f"/tasks/{created['id']}/done")["body"]["done"] is True
assert dispatch("DELETE", f"/tasks/{created['id']}")["status"] == 204`},
		{"layered", `from src.app.router.api import dispatch
assert dispatch("GET", "/users")["body"][0]["email"] == "ada@example.com"
order = dispatch("POST", "/orders", {"userId": 1, "items": [{"product": "book", "quantity": 2, "unitPrice": 3.25}]})
assert order["status"] == 201 and order["body"]["total"] == 6.5
assert dispatch("GET", "/orders", query={"userId": "1"})["body"][0]["id"] == order["body"]["id"]
assert dispatch("POST", f"/orders/{order['body']['id']}/pay")["body"]["status"] == "paid"
assert dispatch("POST", f"/orders/{order['body']['id']}/cancel")["status"] == 409`},
		{"modular", `from src.app.shared.api import dispatch
assert dispatch("GET", "/health")["body"] == {"status": "ok"}
order = dispatch("POST", "/orders", {"items": [{"productId": 1, "quantity": 2}]})
assert order["status"] == 201 and order["body"]["total"] == 900
assert dispatch("GET", "/products/1")["body"]["stock"] == 8`},
		{"hexagonal", `from src.app.routers.api import dispatch
product = dispatch("GET", "/products")["body"][0]
discounted = dispatch("POST", f"/products/{product['id']}/discount", {"percent": 10})
assert discounted["status"] == 200 and discounted["body"]["price"] == 288
assert dispatch("POST", f"/products/{product['id']}/discount", {"percent": 100})["status"] == 422`},
		{"clean_arch", `from src.app.frameworks_and_drivers.web.api import dispatch
order = dispatch("POST", "/orders", {"customerId": "customer-1", "items": [{"sku": "SKU-1", "quantity": 2, "unitPrice": 12.5}]})
assert order["status"] == 201 and order["body"]["total"] == {"amount": 25.0, "formatted": "R$ 25,00"}
assert dispatch("POST", f"/orders/{order['body']['id']}/pay")["body"]["status"] == "paid"
assert dispatch("POST", f"/orders/{order['body']['id']}/pay")["status"] == 409`},
	}

	for _, testCase := range cases {
		t.Run(testCase.architecture, func(t *testing.T) {
			destination := filepath.Join(t.TempDir(), "Sample")
			config := types.ProjectConfig{ProjectName: "Sample", Language: "python", Framework: "raw", Architecture: testCase.architecture}
			if err := Generate(destination, config, templates); err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			command := exec.Command(python, "-c", testCase.script)
			command.Dir = destination
			command.Env = append(os.Environ(), "PYTHONPATH="+destination)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("generated Python demo failed: %v\n%s", err, output)
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
			config := types.ProjectConfig{ProjectName: "Sample", Language: "python", Framework: framework, Architecture: "flat"}
			if err := Generate(destination, config, templates); err != nil {
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
	for _, config := range []types.ProjectConfig{
		{Language: "go", Framework: "express", Architecture: "flat"},
		{Language: "go", Framework: "raw", Architecture: "monorepo"},
		{Language: "node", Runtime: "ruby", Framework: "raw", Architecture: "layered"},
		{Language: "node", Runtime: "js", Framework: "raw", Architecture: "monorepo"},
		{Language: "python", Framework: "raw", Architecture: "monorepo"},
		{Language: "python", Framework: "raw", Architecture: "unknown"},
	} {
		if err := Generate(filepath.Join(t.TempDir(), "project"), config, templates); err == nil || strings.Contains(err.Error(), "template") {
			t.Errorf("Generate(%+v) should reject unsupported options before rendering, got %v", config, err)
		}
	}
}

func TestGenerateGoVet(t *testing.T) {
	templates := os.DirFS(filepath.Join("..", ".."))
	for _, arch := range []string{"flat", "layered", "modular"} {
		dest := filepath.Join(t.TempDir(), arch)
		cfg := types.ProjectConfig{
			ProjectName:  arch,
			Language:     "go",
			Framework:    "raw",
			Architecture: arch,
		}
		if err := Generate(dest, cfg, templates); err != nil {
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
	for _, architecture := range []string{"flat", "layered", "modular", "hexagonal", "clean arch"} {
		t.Run(architecture, func(t *testing.T) {
			destination := filepath.Join(t.TempDir(), "Sample")
			config := types.ProjectConfig{ProjectName: "Sample", Language: "go", Framework: "raw", Architecture: architecture}
			if err := Generate(destination, config, templates); err != nil {
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
	"clean arch": `package main
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
