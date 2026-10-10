package generator

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"crudding/internal/project"
	entryDir    string
		"crudding/internal/generator/golang"
		"crudding/internal/generator/node"
		"crudding/internal/generator/python"
	directories []string
}

		return fmt.Errorf("runtime Node não suportado: %s", config.Runtime)
			{"templates/python/base/requirements.txt.tmpl", "requirements.txt"},
			{"templates/python/base/.gitignore.tmpl", ".gitignore"},
			return python.Generate(destination, config, templates)
			{"templates/go/base/greeting.go.tmpl", filepath.ToSlash(filepath.Join(selectedLayout.serviceDir, "greeting.go"))},
			return golang.Generate(destination, config, templates)
		data.ServiceImport = filepath.ToSlash(serviceImport)
			return node.Generate(destination, config, templates)
	switch config.Language {
			return fmt.Errorf("linguagem não suportada: %s", config.Language)