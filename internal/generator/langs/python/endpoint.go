package python

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"crudding/internal/generator/common"
	"crudding/internal/project"
)

const (
	importsMarker = "# crudding:imports"
	routesMarker  = "# crudding:routes"
	nameToken     = "{name}"
)

type endpointFile struct {
	template     string // relativo à pasta templates/python/, ignorado se isController for true
	output       string // relativo ao pacote base, com {name}
	moduleKey    string // chave opcional usada em {{.Modules.<key>}}
	isController bool   // se true, o template vem de templates/python/frameworks/<framework>/controller.py.tmpl
}

type endpointLayout struct {
	controllerOutput string // caminho do controller do framework relativo ao pacote base
	packages         []string
	files            []endpointFile
}

var endpointLayouts = map[string]endpointLayout{
	"flat": {
		controllerOutput: "{name}_controller.py",
		files: []endpointFile{
			{isController: true, output: "{name}_controller.py"},
			{template: "base/common/service.py.tmpl", output: "{name}_service.py", moduleKey: "service"},
			{template: "base/common/model.py.tmpl", output: "{name}_model.py", moduleKey: "model"},
		},
	},
	"layered": {
		controllerOutput: "controllers/{name}_controller.py",
		files: []endpointFile{
			{isController: true, output: "controllers/{name}_controller.py"},
			{template: "base/common/service.py.tmpl", output: "services/{name}_service.py", moduleKey: "service"},
			{template: "base/common/model.py.tmpl", output: "models/{name}.py", moduleKey: "model"},
		},
	},
	"modular": {
		controllerOutput: "{name}/controller.py",
		packages:         []string{"{name}"},
		files: []endpointFile{
			{isController: true, output: "{name}/controller.py"},
			{template: "base/common/service.py.tmpl", output: "{name}/service.py", moduleKey: "service"},
			{template: "base/common/model.py.tmpl", output: "{name}/models.py", moduleKey: "model"},
		},
	},
	"hexagonal": {
		controllerOutput: "routers/{name}_controller.py",
		files: []endpointFile{
			{isController: true, output: "routers/{name}_controller.py"},
			{template: "base/hexagonal/service.py.tmpl", output: "application/{name}_service.py", moduleKey: "service"},
			{template: "base/common/repository_port.py.tmpl", output: "ports/{name}_repository.py", moduleKey: "port"},
			{template: "base/common/repository_adapter.py.tmpl", output: "infrastructure/{name}_repository.py", moduleKey: "adapter"},
			{template: "base/common/model.py.tmpl", output: "domain/{name}.py", moduleKey: "model"},
		},
	},
	"clean_arch": {
		controllerOutput: "frameworks_and_drivers/web/{name}_controller.py",
		files: []endpointFile{
			{isController: true, output: "frameworks_and_drivers/web/{name}_controller.py"},
			{template: "base/clean_arch/controller.py.tmpl", output: "interface_adapters/controllers/{name}_controller.py", moduleKey: "service"},
			{template: "base/clean_arch/presenter.py.tmpl", output: "interface_adapters/presenters/{name}_presenter.py", moduleKey: "presenter"},
			{template: "base/clean_arch/create.py.tmpl", output: "application/use_cases/create_{name}.py", moduleKey: "create"},
			{template: "base/clean_arch/delete.py.tmpl", output: "application/use_cases/delete_{name}.py", moduleKey: "delete"},
			{template: "base/clean_arch/get.py.tmpl", output: "application/use_cases/get_{name}.py", moduleKey: "get"},
			{template: "base/clean_arch/list.py.tmpl", output: "application/use_cases/list_{name}.py", moduleKey: "list"},
			{template: "base/clean_arch/update.py.tmpl", output: "application/use_cases/update_{name}.py", moduleKey: "update"},
			{template: "base/common/repository_port.py.tmpl", output: "application/port/outbound/{name}_repository.py", moduleKey: "port"},
			{template: "base/common/repository_adapter.py.tmpl", output: "interface_adapters/persistence/{name}_repository.py", moduleKey: "adapter"},
			{template: "base/common/model.py.tmpl", output: "domain/model/{name}.py", moduleKey: "model"},
		},
	},
}

type frameworkRegistration struct {
	importPattern func(controllerModule, name string) string
	routePattern  func(name, route, class string) string
}

var frameworkRegistrations = map[string]frameworkRegistration{
	"fastapi": {
		importPattern: func(controllerModule, name string) string {
			return fmt.Sprintf("from %s import router as %s_router", controllerModule, name)
		},
		routePattern: func(name, route, class string) string {
			return fmt.Sprintf("app.include_router(%s_router, prefix=\"/%s\", tags=[\"%s\"])", name, route, class)
		},
	},
	"flask": {
		importPattern: func(controllerModule, name string) string {
			return fmt.Sprintf("from %s import router as %s_router", controllerModule, name)
		},
		routePattern: func(name, route, class string) string {
			return fmt.Sprintf("app.register_blueprint(%s_router, url_prefix=\"/%s\")", name, route)
		},
	},
	"django": {
		importPattern: func(controllerModule, name string) string {
			return fmt.Sprintf("from %s import urlpatterns as %s_urls", controllerModule, name)
		},
		routePattern: func(name, route, class string) string {
			return fmt.Sprintf("path(\"%s/\", include(%s_urls)),", route, name)
		},
	},
	"raw": {
		importPattern: func(controllerModule, name string) string {
			return fmt.Sprintf("from %s import handle as handle_%s", controllerModule, name)
		},
		routePattern: func(name, route, class string) string {
			return fmt.Sprintf("\"%s\": handle_%s,", route, name)
		},
	},
}

var pythonKeywords = []string{
	"and", "as", "assert", "async", "await", "break", "class", "continue", "def", "del", "elif", "else",
	"except", "finally", "for", "from", "global", "if", "import", "in", "is", "lambda",
	"nonlocal", "not", "or", "pass", "raise", "return", "try", "while", "with", "yield",
}

func Endpoint(root string, config project.Config, name string) (common.Patch, error) {
	layout, err := endpointLayoutFor(config.Architecture)
	if err != nil {
		return common.Patch{}, err
	}
	pkgDir, err := findPackage(root)
	if err != nil {
		return common.Patch{}, err
	}
	return buildEndpointPatch(pkgDir, layout, config, name)
}

func endpointPatchForPackage(pkgDir string, config project.Config, name string) (common.Patch, error) {
	layout, err := endpointLayoutFor(config.Architecture)
	if err != nil {
		return common.Patch{}, err
	}
	return buildEndpointPatch(pkgDir, layout, config, name)
}

func endpointLayoutFor(architecture string) (endpointLayout, error) {
	layout, ok := endpointLayouts[architecture]
	if !ok {
		return endpointLayout{}, fmt.Errorf("add endpoint não suportado para a arquitetura Python %s", architecture)
	}
	return layout, nil
}

func buildEndpointPatch(pkgDir string, layout endpointLayout, config project.Config, name string) (common.Patch, error) {
	if slices.Contains(pythonKeywords, name) {
		return common.Patch{}, fmt.Errorf("%q é uma palavra reservada do Python", name)
	}
	registration, ok := frameworkRegistrations[config.Framework]
	if !ok {
		return common.Patch{}, fmt.Errorf("framework Python não suportado: %s", config.Framework)
	}
	expand := func(rel string) string { return strings.ReplaceAll(rel, nameToken, name) }

	modules := make(map[string]string)
	for _, file := range layout.files {
		if file.moduleKey != "" {
			modules[file.moduleKey] = module(path.Join(pkgDir, expand(file.output)))
		}
	}

	data := templateData{
		BaseData:     common.NewBaseData(config),
		Dependencies: FrameworkDependencies[config.Framework],
		Name:         name,
		Class:        className(name),
		Route:        strings.ReplaceAll(name, "_", "-"),
		Modules:      modules,
	}

	var files []common.File
	for _, file := range layout.files {
		templatePath := path.Join("templates/python", file.template)
		if file.isController {
			templatePath = fmt.Sprintf("templates/python/frameworks/%s/controller.py.tmpl", config.Framework)
		}
		files = append(files, common.File{
			TemplatePath: templatePath,
			OutputPath:   path.Join(pkgDir, expand(file.output)),
		})
	}
	packages := make([]string, len(layout.packages))
	for index, pkg := range layout.packages {
		packages[index] = path.Join(pkgDir, expand(pkg))
	}

	controllerModule := module(path.Join(pkgDir, expand(layout.controllerOutput)))
	mainScript := path.Join(pkgDir, "main.py")
	return common.Patch{
		Files:    files,
		Data:     data,
		Packages: packages,
		Inserts: []common.Insert{
			{Path: mainScript, Marker: importsMarker, Lines: []string{
				registration.importPattern(controllerModule, name),
			}},
			{Path: mainScript, Marker: routesMarker, Lines: []string{
				registration.routePattern(name, data.Route, data.Class),
			}},
		},
	}, nil
}

func findPackage(root string) (string, error) {
	if _, err := os.Stat(filepath.Join(root, "src", "main.py")); err == nil {
		return "src", nil
	}
	matches, err := filepath.Glob(filepath.Join(root, "src", "*", "main.py"))
	if err != nil {
		return "", fmt.Errorf("procurar main.py: %w", err)
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("esperado exatamente um src/*/main.py ou src/main.py, encontrados %d", len(matches))
	}
	rel, err := filepath.Rel(root, matches[0])
	if err != nil {
		return "", err
	}
	pkgDir := filepath.ToSlash(rel)
	return strings.TrimSuffix(pkgDir, "/main.py"), nil
}

func className(name string) string {
	var class strings.Builder
	for _, part := range strings.Split(name, "_") {
		if part != "" {
			class.WriteString(strings.ToUpper(part[:1]) + part[1:])
		}
	}
	return class.String()
}

