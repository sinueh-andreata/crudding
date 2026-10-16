package node

import (
	"fmt"
	"path"

	"crudding/internal/generator/common"
	"crudding/internal/project"
)

var Runtimes = []string{"js", "ts"}

var Frameworks = []string{"raw", "express", "nest", "fastify"}

var Architectures = []string{"flat", "layered", "modular", "hexagonal", "clean_arch"}

var ORMs = []string{"prisma", "typeorm"}

type templateData struct {
	common.BaseData
	ProjectSlug string
	Runtime     string
	EntryScript string
}

// layout descreve uma arquitetura. Os caminhos não têm extensão: cada arquivo
// vem de templates/node/architectures/<arquitetura>/<arquivo>.<runtime>.tmpl.
type layout struct {
	entry     string
	httpDir   string
	endpoints []string
	files     []string
}

var layouts = map[string]layout{
	"flat": {
		entry:     "src/main",
		httpDir:   "src",
		endpoints: []string{"tasks"},
		files:     []string{"src/main", "src/tasks", "src/tasks.test"},
	},
	"layered": {
		entry:     "src/main",
		httpDir:   "src/http",
		endpoints: []string{"user", "order"},
		files: []string{
			"src/main",
			"src/routes/index",
			"src/routes/user-routes",
			"src/routes/order-routes",
			"src/controllers/respond",
			"src/controllers/user-controller",
			"src/controllers/order-controller",
			"src/services/user-service",
			"src/services/order-service",
			"src/models/errors",
			"src/models/user",
			"src/models/order",
		},
	},
	"modular": {
		entry:     "src/main",
		httpDir:   "src/shared/http",
		endpoints: []string{"catalog", "orders"},
		files: []string{
			"src/main",
			"src/modules/catalog/index",
			"src/modules/catalog/product",
			"src/modules/catalog/catalog-service",
			"src/modules/catalog/catalog-controller",
			"src/modules/catalog/catalog-routes",
			"src/modules/orders/index",
			"src/modules/orders/order",
			"src/modules/orders/order-service",
			"src/modules/orders/order-controller",
			"src/modules/orders/order-routes",
			"src/shared/database",
			"src/shared/errors",
			"src/shared/respond",
		},
	},
	"hexagonal": {
		entry:     "src/entrypoints/main",
		httpDir:   "src/adapters/http",
		endpoints: []string{"product"},
		files: []string{
			"src/entrypoints/main",
			"src/domain/errors",
			"src/domain/product",
			"src/domain/product-manager",
			"src/ports/product-repository",
			"src/ports/product-service",
			"src/adapters/http/product-controller",
			"src/adapters/http/routes",
			"src/adapters/persistence/in-memory-product-repository",
		},
	},
	"clean_arch": {
		entry:     "src/frameworks/main",
		httpDir:   "src/frameworks/http",
		endpoints: []string{"order"},
		files: []string{
			"src/frameworks/main",
			"src/frameworks/http/routes",
			"src/entities/errors",
			"src/entities/order",
			"src/use-cases/errors",
			"src/use-cases/order-output",
			"src/use-cases/ports/order-repository",
			"src/use-cases/create-order",
			"src/use-cases/get-order",
			"src/use-cases/list-orders",
			"src/use-cases/pay-order",
			"src/adapters/controllers/order-controller",
			"src/adapters/presenters/order-presenter",
			"src/adapters/repositories/in-memory-order-repository",
		},
	},
}

// Plan monta o projeto Node. Runtime, framework e arquitetura já foram
// validados pelo registry em internal/generator.
func Plan(_ string, config project.Config) (common.Plan, error) {
	layout, ok := layouts[config.Architecture]
	if !ok {
		return common.Plan{}, fmt.Errorf("arquitetura Node não suportada: %s", config.Architecture)
	}
	ext := "." + config.Runtime

	files := []common.File{
		{TemplatePath: "templates/langs/node/base/package.json.tmpl", OutputPath: "package.json"},
		{TemplatePath: "templates/langs/node/base/README.md.tmpl", OutputPath: "README.md"},
		{TemplatePath: "templates/langs/node/base/.gitignore.tmpl", OutputPath: ".gitignore"},
		{TemplatePath: path.Join("templates/langs/node/frameworks", config.Framework, "server"+ext+".tmpl"), OutputPath: path.Join(layout.httpDir, "server"+ext)},
	}
	sources := make([]string, len(layout.files))
	for index, file := range layout.files {
		sources[index] = file + ext
	}
	files = append(files, common.FilesFrom(path.Join("templates/langs/node/architectures", config.Architecture), "", sources...)...)
	if config.Runtime == "ts" {
		files = append(files,
			common.File{TemplatePath: "templates/langs/node/base/tsconfig.json.tmpl", OutputPath: "tsconfig.json"},
			common.File{TemplatePath: "templates/langs/node/base/http-types.ts.tmpl", OutputPath: path.Join(layout.httpDir, "types.ts")},
		)
	}

	data := templateData{
		BaseData:    common.NewBaseData(config),
		ProjectSlug: project.Slug(config.ProjectName),
		Runtime:     config.Runtime,
		EntryScript: layout.entry + ext,
	}
	return common.Plan{Files: files, Data: data, Endpoints: layout.endpoints}, nil
}

