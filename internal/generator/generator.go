package generator

import (
	"fmt"
	"io/fs"

	"crudding/internal/generator/golang"
	"crudding/internal/generator/node"
	"crudding/internal/generator/python"
	"crudding/internal/types"
)

func Generate(destination string, config types.ProjectConfig, templates fs.FS) error {
	switch config.Language {
	case "python":
		return python.Generate(destination, config, templates)
	case "go":
		return golang.Generate(destination, config, templates)
	case "node":
		return node.Generate(destination, config, templates)
	default:
		return fmt.Errorf("linguagem não suportada: %s", config.Language)
	}
}