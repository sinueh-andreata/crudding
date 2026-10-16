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
	"flat":       "src/main.py",
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

var goRouterFiles = map[string]string{
	"flat":       "routes.go",
	"layered":    "internal/routes/routes.go",
	"modular":    "internal/routes/routes.go",
	"hexagonal":  "internal/routes/routes.go",
	"clean_arch": "internal/routes/routes.go",
}

func TestAddEndpointGo(t *testing.T) {
	goCommand, goErr := exec.LookPath("go")
	language, _ := generator.Lookup("go")
	for _, architecture := range language.Architectures {
		t.Run(architecture, func(t *testing.T) {
			routerFile, ok := goRouterFiles[architecture]
			if !ok {
				t.Fatalf("sem routerFile cadastrado para %s", architecture)
			}
			root, m := generateProject(t, project.Config{ProjectName: "Sample", Language: "go", Framework: "raw", Architecture: architecture})

			for _, name := range []string{"customer", "invoice"} {
				result, err := generator.AddEndpoint(root, m, name, templates)
				if err != nil {
					t.Fatalf("AddEndpoint(%s) error = %v", name, err)
				}
				if len(result.Created) == 0 || !slices.Equal(result.Modified, []string{routerFile}) {
					t.Fatalf("AddEndpoint(%s) = %+v", name, result)
				}
				assertFilesExist(t, root, result.Created...)
			}
			assertFileContains(t, root, routerFile, `customerHandler.Create`)
			assertFileContains(t, root, routerFile, "// crudding:routes")

			if architecture == "flat" {
				assertFilesAbsent(t, root, "customer_repository.go", "invoice_repository.go")
			} else if architecture == "layered" {
				assertFilesAbsent(t, root, "internal/repository/customer.go", "internal/repository/invoice.go")
			} else if architecture == "modular" {
				assertFilesAbsent(t, root, "internal/customer/repository.go", "internal/invoice/repository.go")
			}

			before, err := os.ReadFile(filepath.Join(root, routerFile))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := generator.AddEndpoint(root, m, "invoice", templates); err == nil || !strings.Contains(err.Error(), "já existe") {
				t.Errorf("AddEndpoint repetido: err = %v", err)
			}
			if after, _ := os.ReadFile(filepath.Join(root, routerFile)); string(after) != string(before) {
				t.Error("AddEndpoint repetido alterou o arquivo de rotas")
			}

			if goErr != nil {
				t.Skip("go is not installed")
			}
			command := exec.Command(goCommand, "test", "./...")
			command.Dir = root
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("projeto Go após add endpoint não compila: %v\n%s", err, output)
			}
		})
	}
}

func TestAddEndpointUnsupportedLanguages(t *testing.T) {
	for _, config := range []project.Config{
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

func TestRemoveEndpointPython(t *testing.T) {
	python, pythonErr := exec.LookPath("python3")
	language, _ := generator.Lookup("python")
	for _, architecture := range language.Architectures {
		t.Run(architecture, func(t *testing.T) {
			entryScript, ok := pythonEntryScripts[architecture]
			if !ok {
				t.Fatalf("sem entryScript cadastrado para %s", architecture)
			}
			root, m := generateProject(t, project.Config{ProjectName: "Sample", Language: "python", Framework: "fastapi", Architecture: architecture})

			initialRouteContent, err := os.ReadFile(filepath.Join(root, entryScript))
			if err != nil {
				t.Fatal(err)
			}

			addOrderItem, err := generator.AddEndpoint(root, m, "order_item", templates)
			if err != nil {
				t.Fatalf("AddEndpoint(order_item) error = %v", err)
			}
			// Atualiza m após adicionar
			_, m, _ = manifest.Find(root)

			addInvoice, err := generator.AddEndpoint(root, m, "invoice", templates)
			if err != nil {
				t.Fatalf("AddEndpoint(invoice) error = %v", err)
			}
			_, m, _ = manifest.Find(root)

			assertFilesExist(t, root, addOrderItem.Created...)
			assertFilesExist(t, root, addInvoice.Created...)

			// Remover order_item
			removeResult, err := generator.RemoveEndpoint(root, m, "order_item")
			if err != nil {
				t.Fatalf("RemoveEndpoint(order_item) error = %v", err)
			}
			if len(removeResult.Removed) == 0 {
				t.Fatal("RemoveEndpoint(order_item) não registrou arquivos removidos")
			}
			assertFilesAbsent(t, root, addOrderItem.Created...)
			assertFilesExist(t, root, addInvoice.Created...)

			routeContentAfterFirst, err := os.ReadFile(filepath.Join(root, entryScript))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(routeContentAfterFirst), "order_item") {
				t.Error("arquivo de rotas ainda contém order_item após remoção")
			}
			if !strings.Contains(string(routeContentAfterFirst), "invoice") {
				t.Error("arquivo de rotas perdeu invoice indevidamente")
			}
			if !strings.Contains(string(routeContentAfterFirst), "# crudding:routes") {
				t.Error("marker routes foi removido")
			}

			_, m, _ = manifest.Find(root)
			if m.HasEndpoint("order_item") {
				t.Error("manifesto ainda tem order_item após remoção")
			}
			if !m.HasEndpoint("invoice") {
				t.Error("manifesto perdeu invoice")
			}

			if pythonErr == nil {
				command := exec.Command(python, "-m", "compileall", "-q", root)
				if output, err := command.CombinedOutput(); err != nil {
					t.Fatalf("projeto após remove endpoint não compila: %v\n%s", err, output)
				}
			}

			// Remover invoice
			_, err = generator.RemoveEndpoint(root, m, "invoice")
			if err != nil {
				t.Fatalf("RemoveEndpoint(invoice) error = %v", err)
			}
			assertFilesAbsent(t, root, addInvoice.Created...)

			routeContentAfterSecond, err := os.ReadFile(filepath.Join(root, entryScript))
			if err != nil {
				t.Fatal(err)
			}
			if string(routeContentAfterSecond) != string(initialRouteContent) {
				t.Errorf("arquivo de rotas não voltou ao estado inicial:\n--- got ---\n%s\n--- want ---\n%s", routeContentAfterSecond, initialRouteContent)
			}

			_, m, _ = manifest.Find(root)
			if m.HasEndpoint("invoice") {
				t.Error("manifesto ainda tem invoice após remoção")
			}

			// Tentar remover endpoint que não existe
			if _, err := generator.RemoveEndpoint(root, m, "nonexistent"); err == nil || !strings.Contains(err.Error(), "não encontrado") {
				t.Errorf("remover endpoint inexistente deveria falhar: err = %v", err)
			}
		})
	}
}

func TestRemoveEndpointGo(t *testing.T) {
	goCommand, goErr := exec.LookPath("go")
	language, _ := generator.Lookup("go")
	for _, architecture := range language.Architectures {
		t.Run(architecture, func(t *testing.T) {
			routerFile, ok := goRouterFiles[architecture]
			if !ok {
				t.Fatalf("sem routerFile cadastrado para %s", architecture)
			}
			root, m := generateProject(t, project.Config{ProjectName: "Sample", Language: "go", Framework: "raw", Architecture: architecture})

			initialRouteContent, err := os.ReadFile(filepath.Join(root, routerFile))
			if err != nil {
				t.Fatal(err)
			}

			addCustomer, err := generator.AddEndpoint(root, m, "customer", templates)
			if err != nil {
				t.Fatalf("AddEndpoint(customer) error = %v", err)
			}
			_, m, _ = manifest.Find(root)

			addInvoice, err := generator.AddEndpoint(root, m, "invoice", templates)
			if err != nil {
				t.Fatalf("AddEndpoint(invoice) error = %v", err)
			}
			_, m, _ = manifest.Find(root)

			assertFilesExist(t, root, addCustomer.Created...)
			assertFilesExist(t, root, addInvoice.Created...)

			// Remover customer
			removeResult, err := generator.RemoveEndpoint(root, m, "customer")
			if err != nil {
				t.Fatalf("RemoveEndpoint(customer) error = %v", err)
			}
			if len(removeResult.Removed) == 0 {
				t.Fatal("RemoveEndpoint(customer) não registrou arquivos removidos")
			}
			assertFilesAbsent(t, root, addCustomer.Created...)
			assertFilesExist(t, root, addInvoice.Created...)

			if architecture == "modular" {
				assertFilesAbsent(t, root, "internal/customer")
			}

			routeContentAfterFirst, err := os.ReadFile(filepath.Join(root, routerFile))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(routeContentAfterFirst), "customer") {
				t.Error("arquivo de rotas ainda contém customer após remoção")
			}
			if !strings.Contains(string(routeContentAfterFirst), "invoice") {
				t.Error("arquivo de rotas perdeu invoice indevidamente")
			}
			if !strings.Contains(string(routeContentAfterFirst), "// crudding:routes") {
				t.Error("marker routes foi removido")
			}

			_, m, _ = manifest.Find(root)
			if m.HasEndpoint("customer") {
				t.Error("manifesto ainda tem customer após remoção")
			}
			if !m.HasEndpoint("invoice") {
				t.Error("manifesto perdeu invoice")
			}

			if goErr == nil {
				command := exec.Command(goCommand, "test", "./...")
				command.Dir = root
				if output, err := command.CombinedOutput(); err != nil {
					t.Fatalf("projeto Go após remove endpoint não compila: %v\n%s", err, output)
				}
			}

			// Remover invoice
			_, err = generator.RemoveEndpoint(root, m, "invoice")
			if err != nil {
				t.Fatalf("RemoveEndpoint(invoice) error = %v", err)
			}
			assertFilesAbsent(t, root, addInvoice.Created...)

			routeContentAfterSecond, err := os.ReadFile(filepath.Join(root, routerFile))
			if err != nil {
				t.Fatal(err)
			}
			if string(routeContentAfterSecond) != string(initialRouteContent) {
				t.Errorf("arquivo de rotas não voltou ao estado inicial:\n--- got ---\n%s\n--- want ---\n%s", routeContentAfterSecond, initialRouteContent)
			}

			_, m, _ = manifest.Find(root)
			if m.HasEndpoint("invoice") {
				t.Error("manifesto ainda tem invoice após remoção")
			}

			// Tentar remover endpoint que não existe
			if _, err := generator.RemoveEndpoint(root, m, "nonexistent"); err == nil || !strings.Contains(err.Error(), "não encontrado") {
				t.Errorf("remover endpoint inexistente deveria falhar: err = %v", err)
			}
		})
	}
}

func TestRemoveEndpointUnsupportedLanguages(t *testing.T) {
	for _, config := range []project.Config{
		{ProjectName: "Sample", Language: "node", Runtime: "ts", Framework: "express", Architecture: "modular"},
	} {
		root, m := generateProject(t, config)
		if _, err := generator.RemoveEndpoint(root, m, "invoice"); err == nil || !strings.Contains(err.Error(), "não suportado") {
			t.Errorf("RemoveEndpoint(%s) err = %v", config.Language, err)
		}
	}
}

