// Package manifest lê e grava o project.crud, que registra na raiz de cada
// projeto gerado a combinação usada (linguagem, framework, arquitetura...).
package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"crudding/internal/project"
)

const (
	FileName = "project.crud"
	Version  = 1
)

type Manifest struct {
	Version      int    `json:"version"`
	Name         string `json:"name"`
	Language     string `json:"language"`
	Runtime      string `json:"runtime,omitempty"`
	Framework    string `json:"framework"`
	Architecture string `json:"architecture"`
}

func New(config project.Config) Manifest {
	return Manifest{
		Version:      Version,
		Name:         config.ProjectName,
		Language:     config.Language,
		Runtime:      config.Runtime,
		Framework:    config.Framework,
		Architecture: config.Architecture,
	}
}

func (m Manifest) Config() project.Config {
	return project.Config{
		ProjectName:  m.Name,
		Language:     m.Language,
		Runtime:      m.Runtime,
		Framework:    m.Framework,
		Architecture: m.Architecture,
	}
}

// Stack descreve a combinação de forma legível, ex.: "python + fastapi + modular".
func (m Manifest) Stack() string {
	language := m.Language
	if m.Runtime != "" {
		language += "/" + m.Runtime
	}
	return language + " + " + m.Framework + " + " + m.Architecture
}

func Write(root string, m Manifest) error {
	content, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("codificar %s: %w", FileName, err)
	}
	if err := os.WriteFile(filepath.Join(root, FileName), append(content, '\n'), 0644); err != nil {
		return fmt.Errorf("gravar %s: %w", FileName, err)
	}
	return nil
}

func Read(path string) (Manifest, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("ler %s: %w", path, err)
	}
	var m Manifest
	if err := json.Unmarshal(content, &m); err != nil {
		return Manifest{}, fmt.Errorf("%s inválido: %w", path, err)
	}
	if m.Version != Version {
		return Manifest{}, fmt.Errorf("%s: versão %d não suportada (esperado %d)", path, m.Version, Version)
	}
	return m, nil
}

// Find procura o project.crud em dir e nos diretórios acima dele, devolvendo a
// raiz do projeto e o manifesto lido.
func Find(dir string) (string, Manifest, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", Manifest{}, fmt.Errorf("resolver diretório: %w", err)
	}
	for {
		path := filepath.Join(dir, FileName)
		_, statErr := os.Stat(path)
		if statErr == nil {
			m, err := Read(path)
			return dir, m, err
		}
		if !errors.Is(statErr, fs.ErrNotExist) {
			return "", Manifest{}, fmt.Errorf("verificar %s: %w", path, statErr)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", Manifest{}, fmt.Errorf("nenhum %s encontrado; rode o comando dentro de um projeto gerado pelo crudding", FileName)
		}
		dir = parent
	}
}
