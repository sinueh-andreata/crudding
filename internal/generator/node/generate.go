package node

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
	ProjectSlug   string
	Runtime       string
	Extension     string
	Framework     string
	Architecture  string
	AppImport     string
	EntryScript   string
	CompiledEntry string
	ServiceImport string
}

type component struct {
	path string
	kind string
}

type layout struct {
	entryScript string
	appScript   string
	servicePath string
	directories []string
	components  []component
}

var layouts = map[string]layout{
	"flat": {
		entryScript: "src/main",
		servicePath: "src/app",
		directories: []string{"src"},
		components:  []component{{"src/app", "service"}, {"src/users", "model"}},
	},
	"layered": {
		entryScript: "src/server",
		appScript:   "src/app",
		servicePath: "src/services/users.service",
		directories: []string{"src/routes", "src/controllers", "src/services", "src/router", "src/schemas", "src/middlewares", "src/shared/errors", "src/shared/factories", "src/config"},
		components: []component{
			{"src/routes/index", "routes"}, {"src/routes/users.routes", "routes"},
			{"src/controllers/users.controller", "controller"}, {"src/services/users.service", "service"},
			{"src/schemas/users.schema", "schema"}, {"src/middlewares/validate.middleware", "middleware"},
			{"src/middlewares/error.middleware", "middleware"}, {"src/router/router", "router"},
			{"src/shared/errors/app-error", "error"}, {"src/shared/factories/users.factory", "factory"}, {"src/config/env", "config"},
		},
	},
	"modular": {
		entryScript: "src/server",
		appScript:   "src/index",
		servicePath: "src/features/users/users.service",
		directories: []string{"src/features/auth", "src/features/users", "src/features/products", "src/middleware", "src/config", "src/utils"},
		components: []component{
			{"src/features/auth/auth.route", "routes"}, {"src/features/auth/auth.controller", "controller"}, {"src/features/auth/auth.service", "service"}, {"src/features/auth/auth.model", "model"},
			{"src/features/users/users.route", "routes"}, {"src/features/users/users.controller", "controller"}, {"src/features/users/users.service", "service"}, {"src/features/users/users.model", "model"},
			{"src/features/products/product.route", "routes"}, {"src/features/products/product.controller", "controller"}, {"src/features/products/product.service", "service"}, {"src/features/products/product.model", "model"},
			{"src/middleware/error.middleware", "middleware"}, {"src/config/env", "config"}, {"src/utils/index", "utility"},
		},
	},
	"hexagonal": {
		entryScript: "src/server",
		appScript:   "src/config/app",
		servicePath: "src/core/services/user.service",
		directories: []string{"src/core/domain", "src/core/services", "src/core/ports/in", "src/core/ports/out", "src/adapters/driving/http", "src/adapters/driven/persistence", "src/adapters/driven/messaging", "src/config"},
		components: []component{
			{"src/core/domain/user.entity", "model"}, {"src/core/services/user.service", "service"},
			{"src/core/ports/in/user-service.port", "port"}, {"src/core/ports/out/user-repository.port", "port"}, {"src/core/ports/out/notification.port", "port"},
			{"src/adapters/driving/http/user.controller", "controller"}, {"src/adapters/driving/http/user.routes", "routes"},
			{"src/adapters/driven/persistence/memory-user.repository", "repository"}, {"src/adapters/driven/messaging/console-notifier.adapter", "adapter"},
		},
	},
	"clean_arch": {
		entryScript: "src/main",
		appScript:   "src/infrastructure/http/server",
		servicePath: "src/application/order/CreateOrderUseCase",
		directories: []string{"src/domain/order", "src/domain/shared", "src/application/order/ports", "src/application/shared", "src/infrastructure/persistence", "src/infrastructure/http/routes", "src/infrastructure/payment", "src/infrastructure/config"},
		components: []component{
			{"src/domain/order/Order", "entity"}, {"src/domain/order/OrderItem", "entity"}, {"src/domain/order/OrderStatus", "model"}, {"src/domain/shared/Email", "model"},
			{"src/application/order/CreateOrderUseCase", "service"}, {"src/application/order/CancelOrderUseCase", "usecase"},
			{"src/application/order/ports/OrderRepository", "port"}, {"src/application/order/ports/PaymentGateway", "port"}, {"src/application/shared/UseCase", "port"},
			{"src/infrastructure/persistence/MemoryOrderRepository", "repository"}, {"src/infrastructure/http/routes/orderRoutes", "routes"},
			{"src/infrastructure/payment/ConsolePaymentGateway", "adapter"}, {"src/infrastructure/config/container", "config"},
		},
	},
}

var frameworks = map[string]bool{"raw": true, "express": true, "nest": true, "fastify": true}

func Generate(destination string, config types.ProjectConfig, templates fs.FS) (err error) {
	if !frameworks[config.Framework] {
		return fmt.Errorf("framework Node não suportado: %s", config.Framework)
	}
	if config.Runtime != "js" && config.Runtime != "ts" {
		return fmt.Errorf("runtime Node não suportado: %s", config.Runtime)
	}
	layout, ok := layouts[config.Architecture]
	if !ok {
		return fmt.Errorf("arquitetura não suportada: %s", config.Architecture)
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

	for _, directory := range layout.directories {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(directory)), 0755); err != nil {
			return fmt.Errorf("criar diretório %s: %w", directory, err)
		}
	}
	extension := "." + config.Runtime
	entryScript := filepath.ToSlash(layout.entryScript + extension)
	serviceImportBase := layout.entryScript
	appImport := ""
	if layout.appScript != "" {
		appImport, err = filepath.Rel(filepath.Dir(layout.entryScript), layout.appScript)
		if err != nil {
			return fmt.Errorf("calcular import da aplicação: %w", err)
		}
		appImport = filepath.ToSlash(appImport)
		if !strings.HasPrefix(appImport, ".") {
			appImport = "./" + appImport
		}
		serviceImportBase = layout.appScript
	}
	serviceImport, err := filepath.Rel(filepath.Dir(serviceImportBase), layout.servicePath)
	if err != nil {
		return fmt.Errorf("calcular import do serviço: %w", err)
	}
	serviceImport = filepath.ToSlash(serviceImport)
	if !strings.HasPrefix(serviceImport, ".") {
		serviceImport = "./" + serviceImport
	}
	data := templateData{
		ProjectName:   config.ProjectName,
		ProjectSlug:  project.ProjectSlug(config.ProjectName),
		Runtime:       config.Runtime,
		Extension:    extension,
		Framework:     config.Framework,
		Architecture:  config.Architecture,
		AppImport:     appImport,
		EntryScript:   entryScript,
		CompiledEntry: strings.TrimSuffix(strings.TrimPrefix(entryScript, "src/"), ".ts") + ".js",
		ServiceImport: serviceImport,
	}
	files := []common.File{
		{TemplatePath: "templates/node/base/package.json.tmpl", OutputPath: "package.json"},
		{TemplatePath: "templates/node/architectures/README.md.tmpl", OutputPath: "README.md"},
		{TemplatePath: "templates/node/base/.gitignore.tmpl", OutputPath: ".gitignore"},
	}
	if config.Runtime == "ts" {
		files = append(files, common.File{TemplatePath: "templates/node/base/tsconfig.json.tmpl", OutputPath: "tsconfig.json"})
	}
	if layout.appScript != "" {
		files = append(files, common.File{TemplatePath: "templates/node/frameworks/app.tmpl", OutputPath: layout.appScript + extension})
	}
	for _, item := range layout.components {
		componentData := struct {
			templateData
			Kind string
			Name string
		}{data, item.kind, componentName(filepath.Base(item.path))}
		if err := common.WriteFiles(root, templates, componentData, []common.File{{TemplatePath: "templates/node/architectures/component.tmpl", OutputPath: item.path + extension}}); err != nil {
			return err
		}
	}
	files = append(files, common.File{TemplatePath: "templates/node/frameworks/" + config.Framework + "/main.tmpl", OutputPath: entryScript})
	if err := common.WriteFiles(root, templates, data, files); err != nil {
		return err
	}
	complete = true
	return nil
}

func componentName(filename string) string {
	parts := strings.FieldsFunc(filename, func(character rune) bool {
		return !(character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9')
	})
	for index, part := range parts {
		parts[index] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}