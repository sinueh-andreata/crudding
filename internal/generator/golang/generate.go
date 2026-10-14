package golang

import (
	"fmt"
	"path"

	"crudding/internal/generator/common"
	"crudding/internal/project"
	"crudding/internal/types"
)

var Frameworks = []string{"raw", "gin", "fiber", "chi"}

var Architectures = []string{"flat", "layered", "modular", "hexagonal", "clean_arch"}

type templateData struct {
	common.BaseData
	GoModule      string
	HandlerImport string
	HandlerExpr   string
}

// file liga um template (relativo a templates/go/architectures/<arquitetura>/,
// sem a extensão .tmpl) ao caminho de saída no projeto gerado.
type file struct {
	template string
	output   string
}

// architecture descreve um layout Go. go.mod, go.sum, .gitignore, README.md e
// o main.go do framework são adicionados automaticamente a todos os layouts.
type architecture struct {
	main          string // destino do main.go do framework
	handlerImport string // formatado com o módulo Go (%s)
	handlerExpr   string
	files         []file
}

var architectures = map[string]architecture{
	"flat": {
		main:        "main.go",
		handlerExpr: "NewTaskHandler()",
		files: []file{
			{"tasks.go", "tasks.go"},
			{"tasks_test.go", "tasks_test.go"},
		},
	},
	"layered": {
		main:          "cmd/api/main.go",
		handlerImport: `router "%s/internal/router"`,
		handlerExpr:   "router.Setup()",
		files: []file{
			{"router/api.go", "internal/router/router.go"},
			{"handlers/handler.go", "internal/handlers/handler.go"},
			{"services/service.go", "internal/services/service.go"},
			{"models/item.go", "internal/models/item.go"},
		},
	},
	"modular": {
		main:          "cmd/api/main.go",
		handlerImport: `router "%s/internal/router"`,
		handlerExpr:   "router.Setup()",
		files: []file{
			{"worker/main.go", "cmd/worker/main.go"},
			{"router/router.go", "internal/router/router.go"},
			{"order/order.go", "internal/order/order.go"},
			{"order/handler.go", "internal/order/handler.go"},
			{"order/service.go", "internal/order/service.go"},
			{"product/product.go", "internal/product/product.go"},
			{"product/service.go", "internal/product/service.go"},
			{"platform/database/database.go", "internal/platform/database/database.go"},
		},
	},
	"hexagonal": {
		main:          "cmd/rest/main.go",
		handlerImport: `httpapi "%s/internal/adapters/http"`,
		handlerExpr:   "httpapi.NewAPI()",
		files: []file{
			{"grpc/main.go", "cmd/grpc/main.go"},
			{"domain/product.go", "internal/domain/product.go"},
			{"domain/service.go", "internal/application/service.go"},
			{"ports/repository.go", "internal/ports/repository.go"},
			{"ports/service.go", "internal/ports/service.go"},
			{"adapters/http/handler.go", "internal/adapters/http/handler.go"},
			{"adapters/grpc/server.go", "internal/adapters/grpc/server.go"},
			{"adapters/postgres/repository.go", "internal/adapters/postgres/repository.go"},
		},
	},
	"clean_arch": {
		main:          "cmd/api/main.go",
		handlerImport: `httpapi "%s/internal/infrastructure/http"`,
		handlerExpr:   "httpapi.NewAPI()",
		files: []file{
			{"domain/order.go", "internal/domain/order.go"},
			{"domain/errors.go", "internal/domain/errors.go"},
			{"application/create_order.go", "internal/application/create_order.go"},
			{"application/dto.go", "internal/application/dto.go"},
			{"ports/repository.go", "internal/ports/repository.go"},
			{"infrastructure/persistence/repository.go", "internal/infrastructure/persistence/repository.go"},
			{"infrastructure/http/handler.go", "internal/infrastructure/http/handler.go"},
			{"infrastructure/messaging/producer.go", "internal/infrastructure/messaging/producer.go"},
			{"pkg/validator/validator.go", "pkg/validator/validator.go"},
			{"migrations/000001_init.sql", "migrations/000001_init.sql"},
		},
	},
}

// Plan monta o projeto Go. Framework e arquitetura já foram validados pelo
// registry em internal/generator.
func Plan(_ string, config types.ProjectConfig) (common.Plan, error) {
	arch, ok := architectures[config.Architecture]
	if !ok {
		return common.Plan{}, fmt.Errorf("arquitetura Go não suportada: %s", config.Architecture)
	}
	archDir := path.Join("templates/go/architectures", config.Architecture)

	files := []common.File{
		{TemplatePath: "templates/go/base/go.mod.tmpl", OutputPath: "go.mod"},
		{TemplatePath: "templates/go/base/go.sum.tmpl", OutputPath: "go.sum"},
		{TemplatePath: "templates/go/base/.gitignore.tmpl", OutputPath: ".gitignore"},
		{TemplatePath: path.Join(archDir, "README.md.tmpl"), OutputPath: "README.md"},
		{TemplatePath: path.Join("templates/go/frameworks", config.Framework, "main.go.tmpl"), OutputPath: arch.main},
	}
	for _, f := range arch.files {
		files = append(files, common.File{TemplatePath: path.Join(archDir, f.template+".tmpl"), OutputPath: f.output})
	}

	goModule := "example.com/" + project.Slug(config.ProjectName)
	data := templateData{
		BaseData:    common.NewBaseData(config),
		GoModule:    goModule,
		HandlerExpr: arch.handlerExpr,
	}
	if arch.handlerImport != "" {
		data.HandlerImport = fmt.Sprintf(arch.handlerImport, goModule)
	}

	return common.Plan{Files: files, Data: data}, nil
}
