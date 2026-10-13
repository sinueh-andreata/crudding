package node

import (
	"fmt"
	"io/fs"
	"os"
	"path"

	"crudding/internal/generator/common"
	"crudding/internal/project"
	"crudding/internal/types"
)

type templateData struct {
	ProjectName  string
	ProjectSlug  string
	Runtime      string
	Framework    string
	Architecture string
	EntryScript  string
}

// layout descreve uma arquitetura. Os caminhos não têm extensão: cada arquivo
// vem de templates/node/architectures/<arquitetura>/<arquivo>.<runtime>.tmpl.
type layout struct {
	entry   string
	httpDir string
	files   []string
}

var layouts = map[string]layout{
	"flat": {
		entry:   "src/main",
		httpDir: "src",
		files:   []string{"src/main", "src/tasks", "src/tasks.test"},
	},
	"layered": {
		entry:   "src/main",
		httpDir: "src/http",
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
		entry:   "src/main",
		httpDir: "src/shared/http",
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
		entry:   "src/entrypoints/main",
		httpDir: "src/adapters/http",
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
		entry:   "src/frameworks/main",
		httpDir: "src/frameworks/http",
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

	runtime := config.Runtime
	data := templateData{
		ProjectName:  config.ProjectName,
		ProjectSlug:  project.ProjectSlug(config.ProjectName),
		Runtime:      runtime,
		Framework:    config.Framework,
		Architecture: config.Architecture,
		EntryScript:  layout.entry + "." + runtime,
	}
	files := []common.File{
		{TemplatePath: "templates/node/base/package.json.tmpl", OutputPath: "package.json"},
		{TemplatePath: "templates/node/base/README.md.tmpl", OutputPath: "README.md"},
		{TemplatePath: "templates/node/base/.gitignore.tmpl", OutputPath: ".gitignore"},
		{TemplatePath: fmt.Sprintf("templates/node/frameworks/%s/server.%s.tmpl", config.Framework, runtime), OutputPath: path.Join(layout.httpDir, "server."+runtime)},
	}
	for _, file := range layout.files {
		files = append(files, common.File{
			TemplatePath: fmt.Sprintf("templates/node/architectures/%s/%s.%s.tmpl", config.Architecture, file, runtime),
			OutputPath:   file + "." + runtime,
		})
	}
	if runtime == "ts" {
		files = append(files,
			common.File{TemplatePath: "templates/node/base/tsconfig.json.tmpl", OutputPath: "tsconfig.json"},
			common.File{TemplatePath: "templates/node/base/http-types.ts.tmpl", OutputPath: path.Join(layout.httpDir, "types.ts")},
		)
	}
	if err := common.WriteFiles(root, templates, data, files); err != nil {
		return err
	}
	complete = true
	return nil
}
