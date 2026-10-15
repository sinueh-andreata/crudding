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

var FrameworkDependencies = map[string][]string{
	"fastapi": {"fastapi>=0.115", "uvicorn[standard]>=0.30"},
	"django":  {"Django>=5.1"},
	"flask":   {"Flask>=3.0"},
	"raw":     {},
}

type templateData struct {
	common.BaseData
	Dependencies []string
	EntryModule  string
	EntryScript  string
	Name         string
	Class        string
	Route        string
	Modules      map[string]string
}

// layout descreve a estrutura de diretórios e arquivos de uma arquitetura Python.
// Salvo indicação contrária, os caminhos são relativos ao pacote base (src/app, ou src/<projeto> no flat).
type layout struct {
	entry        string        // recebe o main.py do framework (sempre "main.py")
	packages     []string      // subpacotes que recebem __init__.py
	rootPackages []string      // pacotes fora de src/ que recebem __init__.py
	rootFiles    []common.File // arquivos gravados na raiz do projeto
}

var layouts = map[string]layout{
	"flat": {
		entry:        "main.py",
		rootPackages: []string{"tests"},
		rootFiles: []common.File{
			{TemplatePath: "templates/python/base/flat/setup.py.tmpl", OutputPath: "setup.py"},
		},
	},
	"layered": {
		entry:    "main.py",
		packages: []string{"controllers", "services", "models"},
	},
	"modular": {
		entry:    "main.py",
		packages: []string{"resource"},
	},
	"hexagonal": {
		entry:    "main.py",
		packages: []string{"domain", "application", "ports", "infrastructure", "routers"},
	},
	"clean_arch": {
		entry: "main.py",
		packages: []string{
			"domain/model",
			"application/use_cases",
			"application/port/outbound",
			"interface_adapters/controllers",
			"interface_adapters/presenters",
			"interface_adapters/persistence",
			"frameworks_and_drivers/web",
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

	pkg := "app"
	if config.Architecture == "flat" {
		pkg = packageName(destination)
	}
	pkgDir := path.Join("src", pkg)
	entryScript := path.Join(pkgDir, layout.entry)

	files := []common.File{
		{TemplatePath: path.Join("templates/python/base", config.Architecture, "README.md.tmpl"), OutputPath: "README.md"},
		{TemplatePath: "templates/python/base/common/requirements.txt.tmpl", OutputPath: "requirements.txt"},
		{TemplatePath: "templates/python/base/common/.gitignore.tmpl", OutputPath: ".gitignore"},
		{TemplatePath: "templates/python/base/common/pyproject.toml.tmpl", OutputPath: "pyproject.toml"},
		{TemplatePath: path.Join("templates/python/frameworks", config.Framework, "main.py.tmpl"), OutputPath: entryScript},
	}
	files = append(files, layout.rootFiles...)

	seedPatch, err := endpointPatchForPackage(pkgDir, config, "resource")
	if err != nil {
		return common.Plan{}, err
	}
	files = append(files, seedPatch.Files...)

	packages := append([]string{pkgDir}, layout.rootPackages...)
	for _, sub := range layout.packages {
		packages = append(packages, path.Join(pkgDir, sub))
	}
	packages = append(packages, seedPatch.Packages...)

	data := seedPatch.Data.(templateData)
	data.Dependencies = FrameworkDependencies[config.Framework]
	data.EntryScript = entryScript
	data.EntryModule = module(entryScript)
	data.ProjectName = strings.ReplaceAll(filepath.Base(destination), "_", "-")

	return common.Plan{
		Files:   files,
		Data:    data,
		Inserts: seedPatch.Inserts,
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
