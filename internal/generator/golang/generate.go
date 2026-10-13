package golang

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"crudding/internal/generator/common"
	"crudding/internal/project"
	"crudding/internal/types"
)

type templateData struct {
	ProjectName   string
	GoModule      string
	Framework     string
	Architecture  string
	HandlerImport string
	HandlerExpr   string
}

type architectureDef struct {
	directories   []string
	handlerImport string
	handlerExpr   string
	files         func(framework string) []common.File
}

var frameworks = map[string]bool{"raw": true, "gin": true, "fiber": true, "chi": true}

func normalizeArchitecture(arch string) string {
	switch strings.ToLower(strings.TrimSpace(arch)) {
	case "flat":
		return "flat"
	case "layered":
		return "layered"
	case "modular":
		return "modular"
	case "hexagonal":
		return "hexagonal"
	case "clean arch", "clean_arch", "cleanarch":
		return "clean arch"
	default:
		return arch
	}
}

var architectures = map[string]architectureDef{
	"flat": {
		directories: []string{},
		handlerExpr: "NewTaskHandler()",
		files: func(framework string) []common.File {
			return []common.File{
				{TemplatePath: "templates/go/base/go.mod.tmpl", OutputPath: "go.mod"},
				{TemplatePath: "templates/go/base/go.sum.tmpl", OutputPath: "go.sum"},
				{TemplatePath: "templates/go/base/.gitignore.tmpl", OutputPath: ".gitignore"},
				{TemplatePath: "templates/go/architectures/flat/README.md.tmpl", OutputPath: "README.md"},
				{TemplatePath: "templates/go/architectures/flat/tasks.go.tmpl", OutputPath: "tasks.go"},
				{TemplatePath: "templates/go/architectures/flat/tasks_test.go.tmpl", OutputPath: "tasks_test.go"},
				{TemplatePath: fmt.Sprintf("templates/go/frameworks/%s/main.go.tmpl", framework), OutputPath: "main.go"},
			}
		},
	},
	"layered": {
		directories: []string{
			"cmd/api",
			"internal/router",
			"internal/handlers",
			"internal/services",
			"internal/models",
		},
		handlerImport: `router "%s/internal/router"`,
		handlerExpr:   "router.Setup()",
		files: func(framework string) []common.File {
			return []common.File{
				{TemplatePath: "templates/go/base/go.mod.tmpl", OutputPath: "go.mod"},
				{TemplatePath: "templates/go/base/go.sum.tmpl", OutputPath: "go.sum"},
				{TemplatePath: "templates/go/base/.gitignore.tmpl", OutputPath: ".gitignore"},
				{TemplatePath: "templates/go/architectures/layered/README.md.tmpl", OutputPath: "README.md"},
				{TemplatePath: fmt.Sprintf("templates/go/frameworks/%s/main.go.tmpl", framework), OutputPath: "cmd/api/main.go"},
				{TemplatePath: "templates/go/architectures/layered/router/api.go.tmpl", OutputPath: "internal/router/router.go"},
				{TemplatePath: "templates/go/architectures/layered/handlers/handler.go.tmpl", OutputPath: "internal/handlers/handler.go"},
				{TemplatePath: "templates/go/architectures/layered/services/service.go.tmpl", OutputPath: "internal/services/service.go"},
				{TemplatePath: "templates/go/architectures/layered/models/item.go.tmpl", OutputPath: "internal/models/item.go"},
			}
		},
	},
	"modular": {
		directories: []string{
			"cmd/api",
			"cmd/worker",
			"internal/router",
			"internal/order",
			"internal/product",
			"internal/platform/database",
		},
		handlerImport: `router "%s/internal/router"`,
		handlerExpr:   "router.Setup()",
		files: func(framework string) []common.File {
			return []common.File{
				{TemplatePath: "templates/go/base/go.mod.tmpl", OutputPath: "go.mod"},
				{TemplatePath: "templates/go/base/go.sum.tmpl", OutputPath: "go.sum"},
				{TemplatePath: "templates/go/base/.gitignore.tmpl", OutputPath: ".gitignore"},
				{TemplatePath: "templates/go/architectures/modular/README.md.tmpl", OutputPath: "README.md"},
				{TemplatePath: fmt.Sprintf("templates/go/frameworks/%s/main.go.tmpl", framework), OutputPath: "cmd/api/main.go"},
				{TemplatePath: "templates/go/architectures/modular/worker/main.go.tmpl", OutputPath: "cmd/worker/main.go"},
				{TemplatePath: "templates/go/architectures/modular/router/router.go.tmpl", OutputPath: "internal/router/router.go"},
				{TemplatePath: "templates/go/architectures/modular/order/order.go.tmpl", OutputPath: "internal/order/order.go"},
				{TemplatePath: "templates/go/architectures/modular/order/handler.go.tmpl", OutputPath: "internal/order/handler.go"},
				{TemplatePath: "templates/go/architectures/modular/order/service.go.tmpl", OutputPath: "internal/order/service.go"},
				{TemplatePath: "templates/go/architectures/modular/product/product.go.tmpl", OutputPath: "internal/product/product.go"},
				{TemplatePath: "templates/go/architectures/modular/product/service.go.tmpl", OutputPath: "internal/product/service.go"},
				{TemplatePath: "templates/go/architectures/modular/platform/database/database.go.tmpl", OutputPath: "internal/platform/database/database.go"},
			}
		},
	},
	"hexagonal": {
		directories: []string{
			"cmd/rest",
			"cmd/grpc",
			"internal/domain",
			"internal/application",
			"internal/ports",
			"internal/adapters/http",
			"internal/adapters/grpc",
			"internal/adapters/postgres",
		},
		handlerImport: `httpapi "%s/internal/adapters/http"`,
		handlerExpr:   "httpapi.NewAPI()",
		files: func(framework string) []common.File {
			return []common.File{
				{TemplatePath: "templates/go/base/go.mod.tmpl", OutputPath: "go.mod"},
				{TemplatePath: "templates/go/base/go.sum.tmpl", OutputPath: "go.sum"},
				{TemplatePath: "templates/go/base/.gitignore.tmpl", OutputPath: ".gitignore"},
				{TemplatePath: "templates/go/architectures/hexagonal/README.md.tmpl", OutputPath: "README.md"},
				{TemplatePath: fmt.Sprintf("templates/go/frameworks/%s/main.go.tmpl", framework), OutputPath: "cmd/rest/main.go"},
				{TemplatePath: "templates/go/architectures/hexagonal/grpc/main.go.tmpl", OutputPath: "cmd/grpc/main.go"},
				{TemplatePath: "templates/go/architectures/hexagonal/domain/product.go.tmpl", OutputPath: "internal/domain/product.go"},
				{TemplatePath: "templates/go/architectures/hexagonal/domain/service.go.tmpl", OutputPath: "internal/application/service.go"},
				{TemplatePath: "templates/go/architectures/hexagonal/ports/repository.go.tmpl", OutputPath: "internal/ports/repository.go"},
				{TemplatePath: "templates/go/architectures/hexagonal/ports/service.go.tmpl", OutputPath: "internal/ports/service.go"},
				{TemplatePath: "templates/go/architectures/hexagonal/adapters/http/handler.go.tmpl", OutputPath: "internal/adapters/http/handler.go"},
				{TemplatePath: "templates/go/architectures/hexagonal/adapters/grpc/server.go.tmpl", OutputPath: "internal/adapters/grpc/server.go"},
				{TemplatePath: "templates/go/architectures/hexagonal/adapters/postgres/repository.go.tmpl", OutputPath: "internal/adapters/postgres/repository.go"},
			}
		},
	},
	"clean arch": {
		directories: []string{
			"cmd/api",
			"internal/domain",
			"internal/application",
			"internal/ports",
			"internal/infrastructure/persistence",
			"internal/infrastructure/http",
			"internal/infrastructure/messaging",
			"pkg/validator",
			"migrations",
		},
		handlerImport: `httpapi "%s/internal/infrastructure/http"`,
		handlerExpr:   "httpapi.NewAPI()",
		files: func(framework string) []common.File {
			return []common.File{
				{TemplatePath: "templates/go/base/go.mod.tmpl", OutputPath: "go.mod"},
				{TemplatePath: "templates/go/base/go.sum.tmpl", OutputPath: "go.sum"},
				{TemplatePath: "templates/go/base/.gitignore.tmpl", OutputPath: ".gitignore"},
				{TemplatePath: "templates/go/architectures/clean_arch/README.md.tmpl", OutputPath: "README.md"},
				{TemplatePath: fmt.Sprintf("templates/go/frameworks/%s/main.go.tmpl", framework), OutputPath: "cmd/api/main.go"},
				{TemplatePath: "templates/go/architectures/clean_arch/domain/order.go.tmpl", OutputPath: "internal/domain/order.go"},
				{TemplatePath: "templates/go/architectures/clean_arch/domain/errors.go.tmpl", OutputPath: "internal/domain/errors.go"},
				{TemplatePath: "templates/go/architectures/clean_arch/application/create_order.go.tmpl", OutputPath: "internal/application/create_order.go"},
				{TemplatePath: "templates/go/architectures/clean_arch/application/dto.go.tmpl", OutputPath: "internal/application/dto.go"},
				{TemplatePath: "templates/go/architectures/clean_arch/ports/repository.go.tmpl", OutputPath: "internal/ports/repository.go"},
				{TemplatePath: "templates/go/architectures/clean_arch/infrastructure/persistence/repository.go.tmpl", OutputPath: "internal/infrastructure/persistence/repository.go"},
				{TemplatePath: "templates/go/architectures/clean_arch/infrastructure/http/handler.go.tmpl", OutputPath: "internal/infrastructure/http/handler.go"},
				{TemplatePath: "templates/go/architectures/clean_arch/infrastructure/messaging/producer.go.tmpl", OutputPath: "internal/infrastructure/messaging/producer.go"},
				{TemplatePath: "templates/go/architectures/clean_arch/pkg/validator/validator.go.tmpl", OutputPath: "pkg/validator/validator.go"},
				{TemplatePath: "templates/go/architectures/clean_arch/migrations/000001_init.sql.tmpl", OutputPath: "migrations/000001_init.sql"},
			}
		},
	},
}

func Generate(destination string, config types.ProjectConfig, templates fs.FS) (err error) {
	if !frameworks[config.Framework] {
		return fmt.Errorf("framework Go não suportado: %s", config.Framework)
	}
	normalizedArch := normalizeArchitecture(config.Architecture)
	archDef, ok := architectures[normalizedArch]
	if !ok {
		return fmt.Errorf("arquitetura Go não suportada: %s", config.Architecture)
	}

	root, err := common.CreateProjectDirectory(destination)
	if err != nil {
		return err
	}
	complete := false
	defer func() {
		if !complete {
			_ = os.RemoveAll(root)
		}
	}()

	for _, directory := range archDef.directories {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(directory)), 0755); err != nil {
			return fmt.Errorf("criar diretório %s: %w", directory, err)
		}
	}

	goModule := "example.com/" + project.ProjectSlug(config.ProjectName)
	handlerImport := ""
	if archDef.handlerImport != "" {
		handlerImport = fmt.Sprintf(archDef.handlerImport, goModule)
	}
	data := templateData{
		ProjectName:   config.ProjectName,
		GoModule:      goModule,
		Framework:     config.Framework,
		Architecture:  config.Architecture,
		HandlerImport: handlerImport,
		HandlerExpr:   archDef.handlerExpr,
	}

	files := archDef.files(config.Framework)
	if err := common.WriteFiles(root, templates, data, files); err != nil {
		return err
	}

	complete = true
	return nil
}
