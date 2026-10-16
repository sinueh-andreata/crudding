// Package manifest lê e grava o project.crud, que registra na raiz de cada
// projeto gerado a combinação usada (linguagem, framework, arquitetura...) e
// os endpoints registrados.
package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"crudding/internal/project"
)

const (
	FileName = "project.crud"
	Version  = 1
)

type Manifest struct {
	Version      int      `json:"version"`
	Name         string   `json:"name"`
	Language     string   `json:"language"`
	Runtime      string   `json:"runtime,omitempty"`
	Framework    string   `json:"framework"`
	Architecture string   `json:"architecture"`
	Endpoints    []string `json:"endpoints"`
	Database     string   `json:"database,omitempty"`
	Databases    []string `json:"databases,omitempty"`
}

func New(config project.Config, endpoints ...string) Manifest {
	eps := slices.Clone(endpoints)
	if eps == nil {
		eps = []string{}
	}
	var dbs []string
	if config.Database != "" {
		dbs = []string{config.Database}
	}
	return Manifest{
		Version:      Version,
		Name:         config.ProjectName,
		Language:     config.Language,
		Runtime:      config.Runtime,
		Framework:    config.Framework,
		Architecture: config.Architecture,
		Endpoints:    eps,
		Database:     config.Database,
		Databases:    dbs,
	}
}

func (m *Manifest) AddEndpoint(name string) {
	if !slices.Contains(m.Endpoints, name) {
		m.Endpoints = append(m.Endpoints, name)
	}
}

func (m *Manifest) RemoveEndpoint(name string) bool {
	idx := slices.Index(m.Endpoints, name)
	if idx < 0 {
		return false
	}
	m.Endpoints = slices.Delete(m.Endpoints, idx, idx+1)
	return true
}

func (m Manifest) HasEndpoint(name string) bool {
	return slices.Contains(m.Endpoints, name)
}

func (m Manifest) Config() project.Config {
	return project.Config{
		ProjectName:  m.Name,
		Language:     m.Language,
		Runtime:      m.Runtime,
		Framework:    m.Framework,
		Architecture: m.Architecture,
		Database:     m.Database,
	}
}

func (m Manifest) HasDatabase(name string) bool {
	if m.Database == name {
		return true
	}
	return slices.Contains(m.Databases, name)
}

func (m *Manifest) AddDatabase(name string) {
	if m.Database == "" {
		m.Database = name
	}
	if !slices.Contains(m.Databases, name) {
		m.Databases = append(m.Databases, name)
	}
}

func (m *Manifest) SetDatabase(db string) {
	m.AddDatabase(db)
}

// Stack descreve a combinação de forma legível, ex.: "python + fastapi + modular".
func (m Manifest) Stack() string {
	language := m.Language
	if m.Runtime != "" {
		language += "/" + m.Runtime
	}
	s := language + " + " + m.Framework + " + " + m.Architecture
	if len(m.Databases) > 0 {
		s += " + " + strings.Join(m.Databases, ", ")
	} else if m.Database != "" {
		s += " + " + m.Database
	}
	return s
}

func Write(root string, m Manifest) error {
	if m.Endpoints == nil {
		m.Endpoints = []string{}
	}
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
	if m.Endpoints == nil {
		m.Endpoints = []string{}
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
