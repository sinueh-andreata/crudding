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
	endpointDir   = "templates/python/endpoint"
	importsMarker = "# crudding:imports"
	routesMarker  = "# crudding:routes"
	nameToken     = "{name}"
	sharedRoutes  = "routes.py" // template de rotas usado quando o layout define handler
)

type endpointData struct {
	Name    string // identificador snake_case, ex.: order_item
	Class   string // ex.: OrderItem
	Route   string // primeiro segmento da URL, ex.: order-item
	Package string // pacote base, ex.: src.app
	// Módulo importado como "handler" pelo routes.py compartilhado.
	HandlerParent string
	HandlerModule string
}

// endpointFile liga um template (relativo a templates/python/endpoint) ao
// arquivo gerado, relativo ao pacote base. {name} vira o nome do recurso.
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
		files:  []endpointFile{{"flat/resource.py", "{name}.py"}},
	},
	"layered": {
		api:     "router/api.py",
		routes:  "router/{name}_routes.py",
		handler: "controllers/{name}_controller.py",
		files: []endpointFile{
			{"layered/model.py", "models/{name}.py"},
			{"layered/service.py", "services/{name}_service.py"},
			{"layered/controller.py", "controllers/{name}_controller.py"},
		},
	},
	"modular": {
		api:      "shared/api.py",
		routes:   "{name}/routes.py",
		handler:  "{name}/views.py",
		packages: []string{"{name}"},
		files: []endpointFile{
			{"modular/models.py", "{name}/models.py"},
			{"modular/service.py", "{name}/service.py"},
			{"modular/views.py", "{name}/views.py"},
		},
	},
	"hexagonal": {
		api:     "routers/api.py",
		routes:  "routers/{name}_routes.py",
		handler: "routers/{name}.py",
		files: []endpointFile{
			{"hexagonal/entity.py", "domain/{name}.py"},
			{"hexagonal/port.py", "ports/{name}_repository.py"},
			{"hexagonal/service.py", "application/{name}_service.py"},
			{"hexagonal/adapter.py", "infrastructure/{name}_repository.py"},
			{"hexagonal/http.py", "routers/{name}.py"},
		},
	},
	"clean_arch": {
		api:    "frameworks_and_drivers/web/api.py",
		routes: "frameworks_and_drivers/web/{name}_routes.py",
		files: []endpointFile{
			{"clean_arch/entity.py", "domain/model/{name}.py"},
			{"clean_arch/port.py", "application/port/outbound/{name}_repository.py"},
			{"clean_arch/list.py", "application/use_cases/list_{name}.py"},
			{"clean_arch/get.py", "application/use_cases/get_{name}.py"},
			{"clean_arch/create.py", "application/use_cases/create_{name}.py"},
			{"clean_arch/update.py", "application/use_cases/update_{name}.py"},
			{"clean_arch/delete.py", "application/use_cases/delete_{name}.py"},
			{"clean_arch/persistence.py", "interface_adapters/persistence/{name}_repository.py"},
			{"clean_arch/presenter.py", "interface_adapters/presenters/{name}_presenter.py"},
			{"clean_arch/controller.py", "interface_adapters/controllers/{name}_controller.py"},
			{"clean_arch/routes.py", "frameworks_and_drivers/web/{name}_routes.py"},
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
	layout, ok := endpointLayouts[config.Architecture]
	if !ok {
		return common.Patch{}, fmt.Errorf("add endpoint não suportado para a arquitetura Python %s", config.Architecture)
	}
	if slices.Contains(pythonKeywords, name) {
		return common.Patch{}, fmt.Errorf("%q é uma palavra reservada do Python", name)
	}
	pkgDir, err := findPackage(root, layout.api)
	if err != nil {
		return common.Patch{}, err
	}
	expand := func(rel string) string { return strings.ReplaceAll(rel, nameToken, name) }

	data := endpointData{
		Name:    name,
		Class:   className(name),
		Route:   strings.ReplaceAll(name, "_", "-"),
		Package: module(pkgDir),
	}
	if layout.handler != "" {
		handler := module(path.Join(pkgDir, expand(layout.handler)))
		data.HandlerParent, data.HandlerModule = splitModule(handler)
	}

	var files []common.File
	for _, file := range layout.files {
		files = append(files, common.File{
			TemplatePath: path.Join(endpointDir, file.template+".tmpl"),
			OutputPath:   path.Join(pkgDir, expand(file.output)),
		})
	}
	if layout.handler != "" {
		files = append(files, common.File{
			TemplatePath: path.Join(endpointDir, sharedRoutes+".tmpl"),
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
