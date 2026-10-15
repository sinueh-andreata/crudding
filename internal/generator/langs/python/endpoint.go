package python

import (
	"fmt"
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
	sharedRoutes  = "routes.py" // template de rotas usado quando o layout define handler
)

// endpointFile liga um template (relativo à arquitetura) ao arquivo gerado,
// relativo ao pacote base. {name} vira o nome do recurso.
type endpointFile struct {
	template string
	output   string
}

// endpointLayout descreve o que "add endpoint" cria em cada arquitetura. O
// módulo routes expõe route(method, parts, body, query) e é registrado no
// dispatch() de api.
type endpointLayout struct {
	api      string   // módulo com dispatch() e os markers crudding
	routes   string   // módulo com route()
	handler  string   // módulo usado pelo routes.py compartilhado (vazio se routes tem template próprio)
	packages []string // subpacotes novos que recebem __init__.py
	files    []endpointFile
}

var endpointLayouts = map[string]endpointLayout{
	"flat": {
		api:    "modulo.py",
		routes: "{name}.py",
		files:  []endpointFile{{"resource.py", "{name}.py"}},
	},
	"layered": {
		api:     "router/api.py",
		routes:  "router/{name}_routes.py",
		handler: "controllers/{name}_controller.py",
		files: []endpointFile{
			{"model.py", "models/{name}.py"},
			{"service.py", "services/{name}_service.py"},
			{"controller.py", "controllers/{name}_controller.py"},
		},
	},
	"modular": {
		api:      "shared/api.py",
		routes:   "{name}/routes.py",
		handler:  "{name}/views.py",
		packages: []string{"{name}"},
		files: []endpointFile{
			{"models.py", "{name}/models.py"},
			{"service.py", "{name}/service.py"},
			{"views.py", "{name}/views.py"},
		},
	},
	"hexagonal": {
		api:     "routers/api.py",
		routes:  "routers/{name}_routes.py",
		handler: "routers/{name}.py",
		files: []endpointFile{
			{"entity.py", "domain/{name}.py"},
			{"port.py", "ports/{name}_repository.py"},
			{"service.py", "application/{name}_service.py"},
			{"adapter.py", "infrastructure/{name}_repository.py"},
			{"http.py", "routers/{name}.py"},
		},
	},
	"clean_arch": {
		api:    "frameworks_and_drivers/web/api.py",
		routes: "frameworks_and_drivers/web/{name}_routes.py",
		files: []endpointFile{
			{"entity.py", "domain/model/{name}.py"},
			{"port.py", "application/port/outbound/{name}_repository.py"},
			{"list.py", "application/use_cases/list_{name}.py"},
			{"get.py", "application/use_cases/get_{name}.py"},
			{"create.py", "application/use_cases/create_{name}.py"},
			{"update.py", "application/use_cases/update_{name}.py"},
			{"delete.py", "application/use_cases/delete_{name}.py"},
			{"persistence.py", "interface_adapters/persistence/{name}_repository.py"},
			{"presenter.py", "interface_adapters/presenters/{name}_presenter.py"},
			{"controller.py", "interface_adapters/controllers/{name}_controller.py"},
			{"routes.py", "frameworks_and_drivers/web/{name}_routes.py"},
		},
	},
}

// Palavras reservadas do Python que não podem virar nome de módulo.
var pythonKeywords = []string{
	"and", "as", "assert", "async", "await", "break", "class", "continue", "def", "del", "elif", "else",
	"except", "finally", "for", "from", "global", "if", "import", "in", "is", "lambda",
	"nonlocal", "not", "or", "pass", "raise", "return", "try", "while", "with", "yield",
}

// Endpoint monta o Patch que cria um recurso CRUD (estrutura sem
// implementação) e o registra no dispatch() da arquitetura.
func Endpoint(root string, config project.Config, name string) (common.Patch, error) {
	layout, err := endpointLayoutFor(config.Architecture)
	if err != nil {
		return common.Patch{}, err
	}
	pkgDir, err := findPackage(root, layout.api)
	if err != nil {
		return common.Patch{}, err
	}
	return buildEndpointPatch(pkgDir, config.Architecture, layout, config, name)
}

func endpointPatchForPackage(pkgDir string, config project.Config, name string) (common.Patch, error) {
	layout, err := endpointLayoutFor(config.Architecture)
	if err != nil {
		return common.Patch{}, err
	}
	return buildEndpointPatch(pkgDir, config.Architecture, layout, config, name)
}

func endpointLayoutFor(architecture string) (endpointLayout, error) {
	layout, ok := endpointLayouts[architecture]
	if !ok {
		return endpointLayout{}, fmt.Errorf("add endpoint não suportado para a arquitetura Python %s", architecture)
	}
	return layout, nil
}

func buildEndpointPatch(pkgDir, architecture string, layout endpointLayout, config project.Config, name string) (common.Patch, error) {
	if slices.Contains(pythonKeywords, name) {
		return common.Patch{}, fmt.Errorf("%q é uma palavra reservada do Python", name)
	}
	expand := func(rel string) string { return strings.ReplaceAll(rel, nameToken, name) }

	data := templateData{
		BaseData: common.NewBaseData(config),
		Name:     name,
		Class:    className(name),
		Route:    strings.ReplaceAll(name, "_", "-"),
		Package:  module(pkgDir),
	}
	if layout.handler != "" {
		handler := module(path.Join(pkgDir, expand(layout.handler)))
		data.HandlerParent, data.HandlerModule = splitModule(handler)
	}

	var files []common.File
	templateDir := path.Join("templates/python/architectures", architecture, "endpoint")
	for _, file := range layout.files {
		files = append(files, common.File{
			TemplatePath: path.Join(templateDir, file.template+".tmpl"),
			OutputPath:   path.Join(pkgDir, expand(file.output)),
		})
	}
	if layout.handler != "" {
		files = append(files, common.File{
			TemplatePath: path.Join(templateDir, sharedRoutes+".tmpl"),
			OutputPath:   path.Join(pkgDir, expand(layout.routes)),
		})
	}
	packages := make([]string, len(layout.packages))
	for index, pkg := range layout.packages {
		packages[index] = path.Join(pkgDir, expand(pkg))
	}

	routesParent, routesModule := splitModule(module(path.Join(pkgDir, expand(layout.routes))))
	alias := name + "_routes"
	api := path.Join(pkgDir, layout.api)
	return common.Patch{
		Files:    files,
		Data:     data,
		Packages: packages,
		Inserts: []common.Insert{
			{Path: api, Marker: importsMarker, Lines: []string{
				fmt.Sprintf("from %s import %s as %s", routesParent, routesModule, alias),
			}},
			{Path: api, Marker: routesMarker, Lines: []string{
				fmt.Sprintf("if parts and parts[0] == %q:", data.Route),
				fmt.Sprintf("    return %s.route(method, parts[1:], body, query)", alias),
			}},
		},
	}, nil
}

// findPackage localiza o pacote base (src/<pacote>) pelo módulo de API. No
// flat o nome do pacote vem da pasta do projeto, que pode ter sido renomeada.
func findPackage(root, api string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(root, "src", "*", filepath.FromSlash(api)))
	if err != nil {
		return "", fmt.Errorf("procurar %s: %w", api, err)
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("esperado exatamente um src/*/%s, encontrados %d", api, len(matches))
	}
	rel, err := filepath.Rel(root, matches[0])
	if err != nil {
		return "", err
	}
	pkgDir := filepath.ToSlash(rel)
	return strings.TrimSuffix(pkgDir, "/"+api), nil
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

// splitModule separa "a.b.c" em ("a.b", "c").
func splitModule(dotted string) (string, string) {
	index := strings.LastIndex(dotted, ".")
	return dotted[:index], dotted[index+1:]
}
