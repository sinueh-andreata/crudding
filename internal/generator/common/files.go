package common

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type File struct {
	TemplatePath string
	OutputPath   string
}

func CreateProjectDirectory(destination string) (string, error) {
	root, err := filepath.Abs(destination)
	if err != nil {
		return "", fmt.Errorf("resolver diretório do projeto: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(root), 0755); err != nil {
		return "", fmt.Errorf("criar diretório pai: %w", err)
	}
	if err := os.Mkdir(root, 0755); err != nil {
		return "", fmt.Errorf("criar projeto em %s: %w", root, err)
	}
	return root, nil
}

func WriteFiles(root string, templates fs.FS, data any, files []File) error {
	for _, file := range files {
		outputPath := filepath.Join(root, filepath.FromSlash(file.OutputPath))
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return fmt.Errorf("criar diretório para %s: %w", file.OutputPath, err)
		}
		parsed, err := template.ParseFS(templates, file.TemplatePath)
		if err != nil {
			return fmt.Errorf("ler template %s: %w", file.TemplatePath, err)
		}
		var output bytes.Buffer
		if err := parsed.ExecuteTemplate(&output, filepath.Base(file.TemplatePath), data); err != nil {
			return fmt.Errorf("renderizar template %s: %w", file.TemplatePath, err)
		}
		if err := os.WriteFile(outputPath, output.Bytes(), 0644); err != nil {
			return fmt.Errorf("gravar arquivo %s: %w", file.OutputPath, err)
		}
	}
	return nil
}

func CreatePythonPackages(root, directory string) error {
	parts := strings.Split(filepath.ToSlash(directory), "/")
	for index := range parts {
		packageDir := filepath.Join(root, filepath.Join(parts[:index+1]...))
		if err := os.MkdirAll(packageDir, 0755); err != nil {
			return fmt.Errorf("criar pacote Python %s: %w", packageDir, err)
		}
		if err := os.WriteFile(filepath.Join(packageDir, "__init__.py"), nil, 0644); err != nil {
			return fmt.Errorf("criar pacote Python %s: %w", packageDir, err)
		}
	}
	return nil
}
