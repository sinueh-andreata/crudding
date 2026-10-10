package golang

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"crudding/internal/generator/common"
	"crudding/internal/project"
	"crudding/internal/types"
)

type templateData struct {
	ProjectName  string
	GoModule     string
	Framework    string
	Architecture string
	ServiceDir   string
}

type layout struct {
	entryDir    string
	serviceDir  string
	directories []string
}

var layouts = map[string]layout{
	"monorepo": {"apps/api", "packages/shared", []string{"apps/api", "packages/shared"}},
	"camadas": {"cmd/api", "internal/application", []string{"cmd/api", "internal/presentation", "internal/application", "internal/domain", "internal/infrastructure"}},
	"clean arch": {"cmd/api", "internal/usecases", []string{"cmd/api", "internal/entities", "internal/usecases", "internal/adapters", "internal/frameworks"}},
	"hexagonal": {"cmd/api", "internal/domain", []string{"cmd/api", "internal/domain", "internal/ports", "internal/adapters"}},
	"4horse": {"cmd/api", "internal/services", []string{"cmd/api", "internal/routes", "internal/controllers", "internal/services", "internal/repositories"}},
}

var frameworks = map[string]bool{"raw": true, "gin": true, "fiber": true, "chi": true}

func Generate(destination string, config types.ProjectConfig, templates fs.FS) (err error) {
	if !frameworks[config.Framework] {
		return fmt.Errorf("framework Go não suportado: %s", config.Framework)
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
	data := templateData{
		ProjectName:  config.ProjectName,
		GoModule:     "example.com/" + project.ProjectSlug(config.ProjectName),
		Framework:    config.Framework,
		Architecture: config.Architecture,
		ServiceDir:   layout.serviceDir,
	}
	files := []common.File{
		{TemplatePath: "templates/go/base/go.mod.tmpl", OutputPath: "go.mod"},
		{TemplatePath: "templates/go/base/README.md.tmpl", OutputPath: "README.md"},
		{TemplatePath: "templates/go/base/.gitignore.tmpl", OutputPath: ".gitignore"},
		{TemplatePath: "templates/go/base/greeting.go.tmpl", OutputPath: filepath.ToSlash(filepath.Join(layout.serviceDir, "greeting.go"))},
		{TemplatePath: fmt.Sprintf("templates/go/frameworks/%s/main.go.tmpl", config.Framework), OutputPath: filepath.ToSlash(filepath.Join(layout.entryDir, "main.go"))},
	}
	if err := common.WriteFiles(root, templates, data, files); err != nil {
		return err
	}
	complete = true
	return nil
}