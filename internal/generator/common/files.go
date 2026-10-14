package common

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"text/template"
)

type File struct {
	TemplatePath string
	OutputPath   string
}

// FilesFrom aplica a convenção "um template por arquivo": cada caminho relativo
// rel vira templateDir/rel.tmpl -> outputDir/rel.
func FilesFrom(templateDir, outputDir string, rels ...string) []File {
	files := make([]File, len(rels))
	for index, rel := range rels {
		files[index] = File{
			TemplatePath: path.Join(templateDir, rel+".tmpl"),
			OutputPath:   path.Join(outputDir, rel),
		}
	}
	return files
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
