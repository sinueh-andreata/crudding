package golang

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"crudding/internal/generator/common"
	"crudding/internal/project"
)

const (
	importsMarker = "// crudding:imports"
	routesMarker  = "// crudding:routes"
	nameToken     = "{name}"
)

var goKeywords = []string{
	"break", "default", "func", "interface", "select",
	"case", "defer", "go", "map", "struct",
	"chan", "else", "goto", "package", "switch",
	"const", "fallthrough", "if", "range", "type",
	"continue", "for", "import", "return", "var",
	"string", "int", "int8", "int16", "int32", "int64",
	"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
	"byte", "rune", "float32", "float64", "complex64", "complex128",
	"bool", "error", "nil", "true", "false", "iota",
}

type endpointFile struct {
	template string
	output   string
}

type endpointLayout struct {
	routerFile string
	importLine func(goModule, name string) string
	routeLines func(framework, route, class, varName, name string) []string
	files      func(framework string) []endpointFile
}

func frameworkRouteLines(framework, route, varName string) []string {
	switch framework {
	case "gin":
		return []string{
			fmt.Sprintf("api.POST(\"/%s\", %sHandler.Create)", route, varName),
			fmt.Sprintf("api.GET(\"/%s\", %sHandler.List)", route, varName),
			fmt.Sprintf("api.GET(\"/%s/:id\", %sHandler.GetByID)", route, varName),
			fmt.Sprintf("api.DELETE(\"/%s/:id\", %sHandler.Delete)", route, varName),
		}
	case "fiber":
		return []string{
			fmt.Sprintf("api.Post(\"/%s\", %sHandler.Create)", route, varName),
			fmt.Sprintf("api.Get(\"/%s\", %sHandler.List)", route, varName),
			fmt.Sprintf("api.Get(\"/%s/:id\", %sHandler.GetByID)", route, varName),
			fmt.Sprintf("api.Delete(\"/%s/:id\", %sHandler.Delete)", route, varName),
		}
	case "chi":
		return []string{
			fmt.Sprintf("api.Post(\"/%s\", %sHandler.Create)", route, varName),
			fmt.Sprintf("api.Get(\"/%s\", %sHandler.List)", route, varName),
			fmt.Sprintf("api.Get(\"/%s/{id}\", %sHandler.GetByID)", route, varName),
			fmt.Sprintf("api.Delete(\"/%s/{id}\", %sHandler.Delete)", route, varName),
		}
	default: // raw
		return []string{
			fmt.Sprintf("mux.HandleFunc(\"POST /api/v1/%s\", %sHandler.Create)", route, varName),
			fmt.Sprintf("mux.HandleFunc(\"GET /api/v1/%s\", %sHandler.List)", route, varName),
			fmt.Sprintf("mux.HandleFunc(\"GET /api/v1/%s/{id}\", %sHandler.GetByID)", route, varName),
			fmt.Sprintf("mux.HandleFunc(\"DELETE /api/v1/%s/{id}\", %sHandler.Delete)", route, varName),
		}
	}
}

var endpointLayouts = map[string]endpointLayout{
	"flat": {
		routerFile: "routes.go",
		routeLines: func(framework, route, class, varName, _ string) []string {
			var routes []string
			initLines := []string{
				fmt.Sprintf("%sSvc := New%sService()", varName, class),
				fmt.Sprintf("%sHandler := New%sHandler(%sSvc)", varName, class, varName),
			}
			routes = append(routes, initLines...)
			routes = append(routes, frameworkRouteLines(framework, route, varName)...)
			return routes
		},
		files: func(framework string) []endpointFile {
			return []endpointFile{
				{template: "templates/go/base/common/model.go.tmpl", output: "{name}_model.go"},
				{template: "templates/go/base/common/types.go.tmpl", output: "{name}_types.go"},
				{template: "templates/go/base/common/service.go.tmpl", output: "{name}_service.go"},
				{template: fmt.Sprintf("templates/go/frameworks/%s/handler.go.tmpl", framework), output: "{name}_handler.go"},
			}
		},
	},
	"layered": {
		routerFile: "internal/routes/routes.go",
		importLine: func(goModule, _ string) string {
			return fmt.Sprintf("\"%s/internal/service\"", goModule)
		},
		routeLines: func(framework, route, class, varName, _ string) []string {
			var routes []string
			initLines := []string{
				fmt.Sprintf("%sSvc := service.New%sService()", varName, class),
				fmt.Sprintf("%sHandler := handlers.New%sHandler(%sSvc)", varName, class, varName),
			}
			routes = append(routes, initLines...)
			routes = append(routes, frameworkRouteLines(framework, route, varName)...)
			return routes
		},
		files: func(framework string) []endpointFile {
			return []endpointFile{
				{template: "templates/go/base/common/model.go.tmpl", output: "internal/models/{name}.go"},
				{template: "templates/go/base/common/types.go.tmpl", output: "internal/types/{name}.go"},
				{template: "templates/go/base/common/service.go.tmpl", output: "internal/service/{name}.go"},
				{template: fmt.Sprintf("templates/go/frameworks/%s/handler.go.tmpl", framework), output: "internal/handlers/{name}.go"},
			}
		},
	},
	"modular": {
		routerFile: "internal/routes/routes.go",
		importLine: func(goModule, name string) string {
			return fmt.Sprintf("\"%s/internal/%s\"", goModule, name)
		},
		routeLines: func(framework, route, class, varName, name string) []string {
			var routes []string
			initLines := []string{
				fmt.Sprintf("%sSvc := %s.NewService()", varName, name),
				fmt.Sprintf("%sHandler := %s.NewHandler(%sSvc)", varName, name, varName),
			}
			routes = append(routes, initLines...)
			routes = append(routes, frameworkRouteLines(framework, route, varName)...)
			return routes
		},
		files: func(framework string) []endpointFile {
			return []endpointFile{
				{template: "templates/go/base/common/model.go.tmpl", output: "internal/{name}/model.go"},
				{template: "templates/go/base/common/types.go.tmpl", output: "internal/{name}/types.go"},
				{template: "templates/go/base/common/service.go.tmpl", output: "internal/{name}/service.go"},
				{template: fmt.Sprintf("templates/go/frameworks/%s/handler.go.tmpl", framework), output: "internal/{name}/handler.go"},
			}
		},
	},
	"hexagonal": {
		routerFile: "internal/routes/routes.go",
		importLine: func(goModule, _ string) string {
			return fmt.Sprintf("\"%s/internal/adapters/postgres\"\n\t\"%s/internal/application\"", goModule, goModule)
		},
		routeLines: func(framework, route, class, varName, _ string) []string {
			var routes []string
			initLines := []string{
				fmt.Sprintf("%sRepo := postgres.New%sRepository()", varName, class),
				fmt.Sprintf("%sSvc := application.New%sService(%sRepo)", varName, class, varName),
				fmt.Sprintf("%sHandler := handlers.New%sHandler(%sSvc)", varName, class, varName),
			}
			routes = append(routes, initLines...)
			routes = append(routes, frameworkRouteLines(framework, route, varName)...)
			return routes
		},
		files: func(framework string) []endpointFile {
			return []endpointFile{
				{template: "templates/go/base/common/model.go.tmpl", output: "internal/domain/{name}.go"},
				{template: "templates/go/base/common/types.go.tmpl", output: "internal/domain/{name}_types.go"},
				{template: "templates/go/base/common/ports_repository.go.tmpl", output: "internal/ports/{name}_repository.go"},
				{template: "templates/go/base/common/ports_service.go.tmpl", output: "internal/ports/{name}_service.go"},
				{template: "templates/go/base/common/repository.go.tmpl", output: "internal/adapters/postgres/{name}.go"},
				{template: "templates/go/base/common/service.go.tmpl", output: "internal/application/{name}.go"},
				{template: fmt.Sprintf("templates/go/frameworks/%s/handler.go.tmpl", framework), output: "internal/handlers/{name}.go"},
			}
		},
	},
	"clean_arch": {
		routerFile: "internal/routes/routes.go",
		importLine: func(goModule, _ string) string {
			return fmt.Sprintf("\"%s/internal/application\"\n\t\"%s/internal/infrastructure/persistence\"", goModule, goModule)
		},
		routeLines: func(framework, route, class, varName, _ string) []string {
			var routes []string
			initLines := []string{
				fmt.Sprintf("%sRepo := persistence.New%sRepository()", varName, class),
				fmt.Sprintf("%sSvc := application.New%sService(%sRepo)", varName, class, varName),
				fmt.Sprintf("%sHandler := handlers.New%sHandler(%sSvc)", varName, class, varName),
			}
			routes = append(routes, initLines...)
			routes = append(routes, frameworkRouteLines(framework, route, varName)...)
			return routes
		},
		files: func(framework string) []endpointFile {
			return []endpointFile{
				{template: "templates/go/base/common/model.go.tmpl", output: "internal/domain/{name}.go"},
				{template: "templates/go/base/common/types.go.tmpl", output: "internal/domain/{name}_types.go"},
				{template: "templates/go/base/common/ports_repository.go.tmpl", output: "internal/ports/{name}_repository.go"},
				{template: "templates/go/base/common/ports_service.go.tmpl", output: "internal/ports/{name}_service.go"},
				{template: "templates/go/base/common/repository.go.tmpl", output: "internal/infrastructure/persistence/{name}.go"},
				{template: "templates/go/base/common/service.go.tmpl", output: "internal/application/{name}.go"},
				{template: fmt.Sprintf("templates/go/frameworks/%s/handler.go.tmpl", framework), output: "internal/handlers/{name}.go"},
			}
		},
	},
}

type endpointTemplateData struct {
	common.BaseData
	GoModule string
	Name     string
	Class    string
	VarName  string
	Route    string
	Package  string
}

// Endpoint monta o Patch para adicionar um novo recurso CRUD em um projeto Go existente.
func Endpoint(root string, config project.Config, name string) (common.Patch, error) {
	if slices.Contains(goKeywords, name) {
		return common.Patch{}, fmt.Errorf("%q é uma palavra reservada do Go", name)
	}

	layout, ok := endpointLayouts[config.Architecture]
	if !ok {
		return common.Patch{}, fmt.Errorf("add endpoint não suportado para a arquitetura Go %s", config.Architecture)
	}

	goModule, err := readGoModule(root)
	if err != nil {
		goModule = "example.com/" + project.Slug(config.ProjectName)
	}

	class := className(name)
	varName := strings.ToLower(class[:1]) + class[1:]
	route := strings.ReplaceAll(name, "_", "-")

	data := endpointTemplateData{
		BaseData: common.NewBaseData(config),
		GoModule: goModule,
		Name:     name,
		Class:    class,
		VarName:  varName,
		Route:    route,
		Package:  name,
	}

	expand := func(p string) string {
		return strings.ReplaceAll(p, nameToken, name)
	}

	var files []common.File
	for _, f := range layout.files(config.Framework) {
		files = append(files, common.File{
			TemplatePath: f.template,
			OutputPath:   expand(f.output),
		})
	}

	var inserts []common.Insert
	if layout.importLine != nil {
		raw := layout.importLine(goModule, name)
		if raw != "" {
			var newLines []string
			routerPath := filepath.Join(root, filepath.FromSlash(layout.routerFile))
			existingContent, _ := os.ReadFile(routerPath)
			existingStr := string(existingContent)
			for _, line := range strings.Split(raw, "\n") {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" && !strings.Contains(existingStr, trimmed) {
					newLines = append(newLines, line)
				}
			}
			if len(newLines) > 0 {
				inserts = append(inserts, common.Insert{
					Path:   layout.routerFile,
					Marker: importsMarker,
					Lines:  newLines,
				})
			}
		}
	}

	if layout.routeLines != nil {
		lines := layout.routeLines(config.Framework, route, class, varName, name)
		if len(lines) > 0 {
			inserts = append(inserts, common.Insert{
				Path:   layout.routerFile,
				Marker: routesMarker,
				Lines:  lines,
			})
		}
	}

	return common.Patch{
		Files:   files,
		Data:    data,
		Inserts: inserts,
	}, nil
}

func readGoModule(root string) (string, error) {
	content, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("ler go.mod: %w", err)
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}
	return "", fmt.Errorf("declaração de module não encontrada em go.mod")
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
