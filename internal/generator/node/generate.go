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
	Framework     string
	Architecture  string
	EntryScript   string
	ServiceImport string
}

type layout struct {
	entryDir    string
	serviceDir  string
	directories []string
}

var layouts = map[string]layout{
	"monorepo": {"apps/api/src", "packages/shared/src", []string{"apps/api/src", "packages/shared/src"}},
	"camadas": {"src/presentation", "src/application", []string{"src/presentation", "src/application", "src/domain", "src/infrastructure"}},
	"clean arch": {"src/frameworks", "src/use-cases", []string{"src/entities", "src/use-cases", "src/adapters", "src/frameworks"}},
	"hexagonal": {"src/entrypoints", "src/domain", []string{"src/domain", "src/ports", "src/adapters", "src/entrypoints"}},
	"4horse": {"src/routes", "src/services", []string{"src/routes", "src/controllers", "src/services", "src/repositories"}},
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
	serviceFile := "greeting." + config.Runtime
	serviceImport, err := filepath.Rel(layout.entryDir, filepath.Join(layout.serviceDir, "greeting.js"))
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
		Framework:     config.Framework,
		Architecture:  config.Architecture,
		EntryScript:   filepath.ToSlash(filepath.Join(layout.entryDir, "main."+config.Runtime)),
		ServiceImport: serviceImport,
	}
	files := []common.File{
		{TemplatePath: "templates/node/base/package.json.tmpl", OutputPath: "package.json"},
		{TemplatePath: "templates/node/base/README.md.tmpl", OutputPath: "README.md"},
		{TemplatePath: "templates/node/base/.gitignore.tmpl", OutputPath: ".gitignore"},
		{TemplatePath: fmt.Sprintf("templates/node/base/greeting.%s.tmpl", config.Runtime), OutputPath: filepath.ToSlash(filepath.Join(layout.serviceDir, serviceFile))},
		{TemplatePath: "templates/node/frameworks/" + config.Framework + "/main.tmpl", OutputPath: data.EntryScript},
	}
	if config.Runtime == "ts" {
		files = append(files, common.File{TemplatePath: "templates/node/base/tsconfig.json.tmpl", OutputPath: "tsconfig.json"})
	}
	if err := common.WriteFiles(root, templates, data, files); err != nil {
		return err
	}
	complete = true
	return nil
}