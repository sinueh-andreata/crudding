package python

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"crudding/internal/generator/common"
	"crudding/internal/project"
)

var Frameworks = []string{"raw", "fastapi", "django", "flask"}

var Architectures = []string{"flat", "layered", "modular", "hexagonal", "clean_arch"}

type templateData struct {
	common.BaseData
	EntryModule string
	EntryScript string
	APIModule   string
	// ServiceModule é um alias de APIModule mantido para templates antigos.
	ServiceModule string
	PackageName   string
}

// layout descreve uma arquitetura Python. Salvo indicação contrária, os
// caminhos são relativos ao pacote principal (src/app, ou src/<projeto> no
// flat) e, nos templates, a templates/python/architectures/<arquitetura>/.
type layout struct {
	entry      string   // recebe o main.py do framework
	api        string   // módulo com dispatch(), importado pelo entrypoint
	packages   []string // subpacotes que recebem __init__.py
	components []string // demais arquivos da arquitetura
	// Itens fora de src/, relativos à raiz do projeto (TemplatePath relativo
	// ao diretório da arquitetura).
	rootPackages []string
	rootFiles    []common.File
}

var layouts = map[string]layout{
	"flat": {
		entry:        "main.py",
		api:          "modulo.py",
		rootPackages: []string{"tests"},
		rootFiles: []common.File{
			{TemplatePath: "setup.py.tmpl", OutputPath: "setup.py"},
			{TemplatePath: "test_modulo.py.tmpl", OutputPath: "tests/test_modulo.py"},
		},
	},
	"layered": {
		entry:    "router/router.py",
		api:      "router/api.py",
		packages: []string{"controllers", "services", "models", "router"},
		components: []string{
			"controllers/user_controller.py",
			"controllers/order_controller.py",
			"services/user_service.py",
			"services/order_service.py",
			"models/user.py",
			"models/order.py",
			"models/errors.py",
		},
	},
	"modular": {
		entry:    "main.py",
		api:      "shared/api.py",
		packages: []string{"catalog", "orders", "shared"},
		components: []string{
			"catalog/models.py",
			"catalog/views.py",
			"catalog/service.py",
			"orders/models.py",
			"orders/service.py",
			"orders/views.py",
			"shared/errors.py",
		},
	},
	"hexagonal": {
		entry:    "main.py",
		api:      "routers/api.py",
		packages: []string{"domain", "application", "ports", "infrastructure", "routers"},
		components: []string{
			"domain/entities.py",
			"domain/exceptions.py",
			"application/product_service.py",
			"ports/product_repository.py",
			"infrastructure/in_memory_product_repository.py",
			"routers/product.py",
		},
	},
	"clean_arch": {
		entry: "frameworks_and_drivers/web/main.py",
		api:   "frameworks_and_drivers/web/api.py",
		packages: []string{
			"domain/model",
			"application/use_cases",
			"application/port/outbound",
			"interface_adapters/controllers",
			"interface_adapters/presenters",
			"interface_adapters/persistence",
			"frameworks_and_drivers/database",
			"frameworks_and_drivers/web",
		},
		components: []string{
			"domain/exceptions.py",
			"domain/model/order.py",
			"application/port/outbound/order_repository.py",
			"application/use_cases/order_output.py",
			"application/use_cases/create_order.py",
			"application/use_cases/get_order.py",
			"application/use_cases/list_orders.py",
			"application/use_cases/pay_order.py",
			"interface_adapters/controllers/order_controller.py",
			"interface_adapters/presenters/order_presenter.py",
			"interface_adapters/persistence/order_repository.py",
			"bootstrap.py",
		},
	},
}

// Plan monta o projeto Python. Framework e arquitetura já foram validados pelo
// registry em internal/generator.
func Plan(destination string, config project.Config) (common.Plan, error) {
	layout, ok := layouts[config.Architecture]
	if !ok {
		return common.Plan{}, fmt.Errorf("arquitetura Python não suportada: %s", config.Architecture)
	}
	archDir := path.Join("templates/python/architectures", config.Architecture)

	pkg := "app"
	if config.Architecture == "flat" {
		pkg = packageName(destination)
	}
	pkgDir := path.Join("src", pkg)
	entryScript := path.Join(pkgDir, layout.entry)
	apiModule := module(path.Join(pkgDir, layout.api))

	files := []common.File{
		{TemplatePath: path.Join(archDir, "README.md.tmpl"), OutputPath: "README.md"},
		{TemplatePath: "templates/python/base/requirements.txt.tmpl", OutputPath: "requirements.txt"},
		{TemplatePath: "templates/python/base/.gitignore.tmpl", OutputPath: ".gitignore"},
		{TemplatePath: "templates/python/base/pyproject.toml.tmpl", OutputPath: "pyproject.toml"},
		{TemplatePath: path.Join("templates/python/frameworks", config.Framework, "main.py.tmpl"), OutputPath: entryScript},
	}
	files = append(files, common.FilesFrom(archDir, pkgDir, append([]string{layout.api}, layout.components...)...)...)
	for _, file := range layout.rootFiles {
		files = append(files, common.File{TemplatePath: path.Join(archDir, file.TemplatePath), OutputPath: file.OutputPath})
	}

	packages := append([]string{pkgDir}, layout.rootPackages...)
	for _, sub := range layout.packages {
		packages = append(packages, path.Join(pkgDir, sub))
	}

	data := templateData{
		BaseData:      common.NewBaseData(config),
		EntryScript:   entryScript,
		EntryModule:   module(entryScript),
		APIModule:     apiModule,
		ServiceModule: apiModule,
		PackageName:   strings.TrimPrefix(path.Dir(entryScript), "src/"),
	}
	data.ProjectName = strings.ReplaceAll(filepath.Base(destination), "_", "-")

	return common.Plan{
		Files:   files,
		Data:    data,
		Prepare: func(root string) error { return common.CreatePackages(root, packages, nil) },
	}, nil
}

// packageName deriva um identificador Python válido a partir do nome da pasta.
func packageName(destination string) string {
	name := strings.ReplaceAll(project.Slug(filepath.Base(destination)), "-", "_")
	if name == "" {
		return "app"
	}
	if name[0] >= '0' && name[0] <= '9' {
		return "project_" + name
	}
	return name
}

func module(file string) string {
	return strings.ReplaceAll(strings.TrimSuffix(file, ".py"), "/", ".")
}
