package python

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
	Framework     string
	Architecture  string
	EntryModule   string
	EntryScript   string
	ServiceModule string
	APIModule     string
	PackageName   string
}

type component struct {
	path         string
	templatePath string
}

type layout struct {
	entryDir        string
	entryFile       string
	serviceDir      string
	serviceFile     string
	serviceTemplate string
	apiDir          string
	apiTemplate     string
	directories     []string
	components      []component
}

var layouts = map[string]layout{
	"flat": {
		entryFile:       "main.py",
		serviceFile:     "modulo.py",
		serviceTemplate: "templates/python/architectures/flat/modulo.py.tmpl",
		directories:     []string{"tests"},
	},
	"layered": {
		entryDir:    "src/app/router",
		entryFile:   "router.py",
		apiDir:      "src/app/router",
		apiTemplate: "templates/python/architectures/layered/router/api.py.tmpl",
		directories: []string{"src/app/controllers", "src/app/services", "src/app/models", "src/app/router"},
		components: []component{
			{"src/app/controllers/user_controller.py", "templates/python/architectures/layered/controllers/user_controller.py.tmpl"},
			{"src/app/controllers/order_controller.py", "templates/python/architectures/layered/controllers/order_controller.py.tmpl"},
			{"src/app/services/user_service.py", "templates/python/architectures/layered/services/user_service.py.tmpl"},
			{"src/app/services/order_service.py", "templates/python/architectures/layered/services/order_service.py.tmpl"},
			{"src/app/models/user.py", "templates/python/architectures/layered/models/user.py.tmpl"},
			{"src/app/models/order.py", "templates/python/architectures/layered/models/order.py.tmpl"},
			{"src/app/models/errors.py", "templates/python/architectures/layered/models/errors.py.tmpl"},
		},
	},
	"modular": {
		entryDir:    "src/app",
		entryFile:   "main.py",
		apiDir:      "src/app/shared",
		apiTemplate: "templates/python/architectures/modular/shared/api.py.tmpl",
		directories: []string{"src/app/catalog", "src/app/orders", "src/app/shared"},
		components: []component{
			{"src/app/catalog/models.py", "templates/python/architectures/modular/catalog/models.py.tmpl"},
			{"src/app/catalog/views.py", "templates/python/architectures/modular/catalog/views.py.tmpl"},
			{"src/app/catalog/service.py", "templates/python/architectures/modular/catalog/service.py.tmpl"},
			{"src/app/orders/models.py", "templates/python/architectures/modular/orders/models.py.tmpl"},
			{"src/app/orders/service.py", "templates/python/architectures/modular/orders/service.py.tmpl"},
			{"src/app/orders/views.py", "templates/python/architectures/modular/orders/views.py.tmpl"},
			{"src/app/shared/errors.py", "templates/python/architectures/modular/shared/errors.py.tmpl"},
		},
	},
	"hexagonal": {
		entryDir:    "src/app",
		entryFile:   "main.py",
		apiDir:      "src/app/routers",
		apiTemplate: "templates/python/architectures/hexagonal/routers/api.py.tmpl",
		directories: []string{"src/app/domain", "src/app/application", "src/app/ports", "src/app/infrastructure", "src/app/routers"},
		components: []component{
			{"src/app/domain/entities.py", "templates/python/architectures/hexagonal/domain/entities.py.tmpl"},
			{"src/app/domain/exceptions.py", "templates/python/architectures/hexagonal/domain/exceptions.py.tmpl"},
			{"src/app/application/product_service.py", "templates/python/architectures/hexagonal/application/product_service.py.tmpl"},
			{"src/app/ports/product_repository.py", "templates/python/architectures/hexagonal/ports/product_repository.py.tmpl"},
			{"src/app/infrastructure/in_memory_product_repository.py", "templates/python/architectures/hexagonal/infrastructure/in_memory_product_repository.py.tmpl"},
			{"src/app/routers/product.py", "templates/python/architectures/hexagonal/routers/product.py.tmpl"},
		},
	},
	"clean_arch": {
		entryDir:    "src/app/frameworks_and_drivers/web",
		entryFile:   "main.py",
		apiDir:      "src/app/frameworks_and_drivers/web",
		apiTemplate: "templates/python/architectures/clean_arch/frameworks_and_drivers/web/api.py.tmpl",
		directories: []string{"src/app/domain/model", "src/app/application/use_cases", "src/app/application/port/outbound", "src/app/interface_adapters/controllers", "src/app/interface_adapters/presenters", "src/app/interface_adapters/persistence", "src/app/frameworks_and_drivers/database", "src/app/frameworks_and_drivers/web"},
		components: []component{
			{"src/app/domain/exceptions.py", "templates/python/architectures/clean_arch/domain/exceptions.py.tmpl"},
			{"src/app/domain/model/order.py", "templates/python/architectures/clean_arch/domain/model/order.py.tmpl"},
			{"src/app/application/port/outbound/order_repository.py", "templates/python/architectures/clean_arch/application/port/outbound/order_repository.py.tmpl"},
			{"src/app/application/use_cases/order_output.py", "templates/python/architectures/clean_arch/application/use_cases/order_output.py.tmpl"},
			{"src/app/application/use_cases/create_order.py", "templates/python/architectures/clean_arch/application/use_cases/create_order.py.tmpl"},
			{"src/app/application/use_cases/get_order.py", "templates/python/architectures/clean_arch/application/use_cases/get_order.py.tmpl"},
			{"src/app/application/use_cases/list_orders.py", "templates/python/architectures/clean_arch/application/use_cases/list_orders.py.tmpl"},
			{"src/app/application/use_cases/pay_order.py", "templates/python/architectures/clean_arch/application/use_cases/pay_order.py.tmpl"},
			{"src/app/interface_adapters/controllers/order_controller.py", "templates/python/architectures/clean_arch/interface_adapters/controllers/order_controller.py.tmpl"},
			{"src/app/interface_adapters/presenters/order_presenter.py", "templates/python/architectures/clean_arch/interface_adapters/presenters/order_presenter.py.tmpl"},
			{"src/app/interface_adapters/persistence/order_repository.py", "templates/python/architectures/clean_arch/interface_adapters/persistence/order_repository.py.tmpl"},
			{"src/app/bootstrap.py", "templates/python/architectures/clean_arch/bootstrap.py.tmpl"},
		},
	},
}

var frameworks = map[string]bool{"raw": true, "fastapi": true, "django": true, "flask": true}

func Generate(destination string, config types.ProjectConfig, templates fs.FS) (err error) {
	if !frameworks[config.Framework] {
		return fmt.Errorf("framework Python não suportado: %s", config.Framework)
	}
	layout, ok := layouts[config.Architecture]
	if !ok {
		return fmt.Errorf("arquitetura não suportada: %s", config.Architecture)
	}
	if config.Architecture == "flat" {
		packageName := strings.ReplaceAll(project.ProjectSlug(filepath.Base(destination)), "-", "_")
		if packageName == "" {
			packageName = "app"
		}
		if packageName[0] >= '0' && packageName[0] <= '9' {
			packageName = "project_" + packageName
		}
		layout.entryDir = filepath.ToSlash(filepath.Join("src", packageName))
		layout.serviceDir = layout.entryDir
		layout.directories = append(layout.directories, layout.entryDir)
		layout.apiDir = layout.entryDir
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
		if err := common.CreatePythonPackages(root, directory); err != nil {
			return err
		}
	}
	entryScript := filepath.ToSlash(filepath.Join(layout.entryDir, layout.entryFile))
	servicePath := ""
	if layout.serviceFile != "" {
		servicePath = filepath.ToSlash(filepath.Join(layout.serviceDir, layout.serviceFile))
	}
	apiPath := ""
	if layout.apiTemplate != "" {
		apiPath = filepath.ToSlash(filepath.Join(layout.apiDir, "api.py"))
	}
	data := templateData{
		ProjectName:   strings.ReplaceAll(filepath.Base(root), "_", "-"),
		Framework:     config.Framework,
		Architecture:  config.Architecture,
		EntryScript:   entryScript,
		EntryModule:   pythonModule(entryScript),
		ServiceModule: pythonModule(servicePath),
		APIModule:     pythonModule(apiPath),
		PackageName:   strings.TrimPrefix(strings.TrimPrefix(filepath.ToSlash(layout.entryDir), "src/"), "src/app/"),
	}
	files := []common.File{
		{TemplatePath: fmt.Sprintf("templates/python/architectures/%s/README.md.tmpl", config.Architecture), OutputPath: "README.md"},
		{TemplatePath: "templates/python/base/requirements.txt.tmpl", OutputPath: "requirements.txt"},
		{TemplatePath: "templates/python/base/.gitignore.tmpl", OutputPath: ".gitignore"},
		{TemplatePath: "templates/python/base/pyproject.toml.tmpl", OutputPath: "pyproject.toml"},
		{TemplatePath: fmt.Sprintf("templates/python/frameworks/%s/main.py.tmpl", config.Framework), OutputPath: entryScript},
	}
	if servicePath != "" {
		serviceTemplate := layout.serviceTemplate
		if serviceTemplate == "" {
			serviceTemplate = "templates/python/base/greeting.py.tmpl"
		}
		files = append(files, common.File{TemplatePath: serviceTemplate, OutputPath: servicePath})
	}
	if apiPath != "" {
		files = append(files, common.File{TemplatePath: layout.apiTemplate, OutputPath: apiPath})
	} else {
		data.APIModule = pythonModule(servicePath)
	}
	if config.Architecture == "flat" {
		files = append(files, common.File{TemplatePath: "templates/python/architectures/flat/setup.py.tmpl", OutputPath: "setup.py"}, common.File{TemplatePath: "templates/python/architectures/flat/test_modulo.py.tmpl", OutputPath: "tests/test_modulo.py"})
	}
	if err := common.WriteFiles(root, templates, data, files); err != nil {
		return err
	}
	for _, item := range layout.components {
		if err := common.WriteFiles(root, templates, data, []common.File{{TemplatePath: item.templatePath, OutputPath: item.path}}); err != nil {
			return err
		}
	}
	complete = true
	return nil
}

func pythonModule(path string) string {
	module := strings.TrimSuffix(filepath.ToSlash(path), ".py")
	return strings.ReplaceAll(module, "/", ".")
}
