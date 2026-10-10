package python

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"crudding/internal/generator/common"
	"crudding/internal/types"
)

type templateData struct {
	ProjectName   string
	Framework     string
	Architecture  string
	EntryModule   string
	EntryScript   string
	ServiceModule string
}

type layout struct {
	entryDir    string
	serviceDir  string
	directories []string
}

var layouts = map[string]layout{
	"monorepo": {"apps/api/src", "apps/api/src/services", []string{"apps/api/src", "apps/api/src/services", "packages/shared"}},
	"camadas": {"src/presentation", "src/application", []string{"src/presentation", "src/application", "src/domain", "src/infrastructure"}},
	"clean arch": {"src/frameworks", "src/use_cases", []string{"src/entities", "src/use_cases", "src/adapters", "src/frameworks"}},
	"hexagonal": {"src/entrypoints", "src/domain", []string{"src/domain", "src/ports", "src/adapters", "src/entrypoints"}},
	"4horse": {"src/routes", "src/services", []string{"src/routes", "src/controllers", "src/services", "src/repositories"}},
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
	entryScript := filepath.ToSlash(filepath.Join(layout.entryDir, "main.py"))
	data := templateData{
		ProjectName:   strings.ReplaceAll(filepath.Base(root), "_", "-"),
		Framework:     config.Framework,
		Architecture:  config.Architecture,
		EntryScript:   entryScript,
		EntryModule:   strings.ReplaceAll(filepath.ToSlash(layout.entryDir), "/", ".") + ".main",
		ServiceModule: strings.ReplaceAll(filepath.ToSlash(layout.serviceDir), "/", ".") + ".greeting",
	}
	files := []common.File{
		{TemplatePath: "templates/python/base/README.md.tmpl", OutputPath: "README.md"},
		{TemplatePath: "templates/python/base/requirements.txt.tmpl", OutputPath: "requirements.txt"},
		{TemplatePath: "templates/python/base/.gitignore.tmpl", OutputPath: ".gitignore"},
		{TemplatePath: "templates/python/base/greeting.py.tmpl", OutputPath: filepath.ToSlash(filepath.Join(layout.serviceDir, "greeting.py"))},
		{TemplatePath: fmt.Sprintf("templates/python/frameworks/%s/main.py.tmpl", config.Framework), OutputPath: entryScript},
	}
	if err := common.WriteFiles(root, templates, data, files); err != nil {
		return err
	}
	complete = true
	return nil
}