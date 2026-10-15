package golang

import (
	"fmt"
	"path"

	"crudding/internal/generator/common"
	"crudding/internal/project"
)

var Frameworks = []string{"raw", "gin", "fiber", "chi"}

var Architectures = []string{"flat", "layered", "modular", "hexagonal", "clean_arch"}

type templateData struct {
	common.BaseData
	GoModule string
	Package  string
	Class    string
	VarName  string
}

type architecture struct {
	main string
}

var architectures = map[string]architecture{
	"flat":       {main: "main.go"},
	"layered":    {main: "cmd/api/main.go"},
	"modular":    {main: "cmd/api/main.go"},
	"hexagonal":  {main: "cmd/api/main.go"},
	"clean_arch": {main: "cmd/api/main.go"},
}

// Plan monta o projeto Go. Framework e arquitetura já foram validados pelo
// registry em internal/generator.
func Plan(_ string, config project.Config) (common.Plan, error) {
	arch, ok := architectures[config.Architecture]
	if !ok {
		return common.Plan{}, fmt.Errorf("arquitetura Go não suportada: %s", config.Architecture)
	}
	archDir := path.Join("templates/go/base", config.Architecture)
	fwDir := path.Join("templates/go/frameworks", config.Framework)
	commonDir := "templates/go/base/common"

	files := []common.File{
		{TemplatePath: path.Join(commonDir, "go.mod.tmpl"), OutputPath: "go.mod"},
		{TemplatePath: path.Join(commonDir, "go.sum.tmpl"), OutputPath: "go.sum"},
		{TemplatePath: path.Join(commonDir, ".gitignore.tmpl"), OutputPath: ".gitignore"},
		{TemplatePath: path.Join(archDir, "README.md.tmpl"), OutputPath: "README.md"},
		{TemplatePath: path.Join(fwDir, "main.go.tmpl"), OutputPath: arch.main},
	}

	switch config.Architecture {
	case "flat":
		files = append(files,
			common.File{TemplatePath: path.Join(commonDir, "model.go.tmpl"), OutputPath: "user_model.go"},
			common.File{TemplatePath: path.Join(commonDir, "types.go.tmpl"), OutputPath: "user_types.go"},
			common.File{TemplatePath: path.Join(commonDir, "service.go.tmpl"), OutputPath: "user_service.go"},
			common.File{TemplatePath: path.Join(fwDir, "handler.go.tmpl"), OutputPath: "user_handler.go"},
			common.File{TemplatePath: path.Join(fwDir, "routes.go.tmpl"), OutputPath: "routes.go"},
		)
	case "layered":
		files = append(files,
			common.File{TemplatePath: path.Join(commonDir, "model.go.tmpl"), OutputPath: "internal/models/user.go"},
			common.File{TemplatePath: path.Join(commonDir, "types.go.tmpl"), OutputPath: "internal/types/user.go"},
			common.File{TemplatePath: path.Join(commonDir, "service.go.tmpl"), OutputPath: "internal/service/user.go"},
			common.File{TemplatePath: path.Join(fwDir, "handler.go.tmpl"), OutputPath: "internal/handlers/user.go"},
			common.File{TemplatePath: path.Join(fwDir, "routes.go.tmpl"), OutputPath: "internal/routes/routes.go"},
		)
	case "modular":
		files = append(files,
			common.File{TemplatePath: path.Join(commonDir, "model.go.tmpl"), OutputPath: "internal/user/model.go"},
			common.File{TemplatePath: path.Join(commonDir, "types.go.tmpl"), OutputPath: "internal/user/types.go"},
			common.File{TemplatePath: path.Join(commonDir, "service.go.tmpl"), OutputPath: "internal/user/service.go"},
			common.File{TemplatePath: path.Join(fwDir, "handler.go.tmpl"), OutputPath: "internal/user/handler.go"},
			common.File{TemplatePath: path.Join(fwDir, "routes.go.tmpl"), OutputPath: "internal/routes/routes.go"},
		)
	case "hexagonal":
		files = append(files,
			common.File{TemplatePath: path.Join(commonDir, "model.go.tmpl"), OutputPath: "internal/domain/user.go"},
			common.File{TemplatePath: path.Join(commonDir, "types.go.tmpl"), OutputPath: "internal/domain/types.go"},
			common.File{TemplatePath: path.Join(commonDir, "ports_repository.go.tmpl"), OutputPath: "internal/ports/repository.go"},
			common.File{TemplatePath: path.Join(commonDir, "ports_service.go.tmpl"), OutputPath: "internal/ports/service.go"},
			common.File{TemplatePath: path.Join(commonDir, "service.go.tmpl"), OutputPath: "internal/application/user.go"},
			common.File{TemplatePath: path.Join(commonDir, "repository.go.tmpl"), OutputPath: "internal/adapters/postgres/user.go"},
			common.File{TemplatePath: path.Join(fwDir, "handler.go.tmpl"), OutputPath: "internal/handlers/user.go"},
			common.File{TemplatePath: path.Join(fwDir, "routes.go.tmpl"), OutputPath: "internal/routes/routes.go"},
		)
	case "clean_arch":
		files = append(files,
			common.File{TemplatePath: path.Join(commonDir, "model.go.tmpl"), OutputPath: "internal/domain/user.go"},
			common.File{TemplatePath: path.Join(commonDir, "types.go.tmpl"), OutputPath: "internal/domain/types.go"},
			common.File{TemplatePath: path.Join(commonDir, "ports_repository.go.tmpl"), OutputPath: "internal/ports/repository.go"},
			common.File{TemplatePath: path.Join(commonDir, "ports_service.go.tmpl"), OutputPath: "internal/ports/service.go"},
			common.File{TemplatePath: path.Join(commonDir, "service.go.tmpl"), OutputPath: "internal/application/user.go"},
			common.File{TemplatePath: path.Join(commonDir, "repository.go.tmpl"), OutputPath: "internal/infrastructure/persistence/user.go"},
			common.File{TemplatePath: path.Join(fwDir, "handler.go.tmpl"), OutputPath: "internal/handlers/user.go"},
			common.File{TemplatePath: path.Join(fwDir, "routes.go.tmpl"), OutputPath: "internal/routes/routes.go"},
		)
	}

	goModule := "example.com/" + project.Slug(config.ProjectName)
	data := templateData{
		BaseData: common.NewBaseData(config),
		GoModule: goModule,
		Package:  "user",
		Class:    "User",
		VarName:  "user",
	}

	return common.Plan{Files: files, Data: data}, nil
}
