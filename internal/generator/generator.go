package generator

import (
	"fmt"
	"io/fs"
	"slices"

	"crudding/internal/generator/common"
	"crudding/internal/generator/golang"
	"crudding/internal/generator/node"
	"crudding/internal/generator/python"
	"crudding/internal/types"
)

// Language descreve uma linguagem suportada. É a fonte única das opções
// exibidas no menu, da validação e da matriz de testes.
type Language struct {
	Name          string   // valor de types.ProjectConfig.Language
	Label         string   // nome exibido em mensagens de erro
	Runtimes      []string // vazio quando a linguagem não pergunta runtime
	Frameworks    []string
	Architectures []string
	plan          func(destination string, config types.ProjectConfig) (common.Plan, error)
}

var languages = []Language{
	{Name: "python", Label: "Python", Frameworks: python.Frameworks, Architectures: python.Architectures, plan: python.Plan},
	{Name: "go", Label: "Go", Frameworks: golang.Frameworks, Architectures: golang.Architectures, plan: golang.Plan},
	{Name: "node", Label: "Node", Runtimes: node.Runtimes, Frameworks: node.Frameworks, Architectures: node.Architectures, plan: node.Plan},
}

// Languages devolve as linguagens suportadas, na ordem de exibição do menu.
func Languages() []Language {
	return slices.Clone(languages)
}

func Lookup(name string) (Language, bool) {
	index := slices.IndexFunc(languages, func(language Language) bool { return language.Name == name })
	if index < 0 {
		return Language{}, false
	}
	return languages[index], true
}

func (l Language) validate(config types.ProjectConfig) error {
	if len(l.Runtimes) > 0 && !slices.Contains(l.Runtimes, config.Runtime) {
		return fmt.Errorf("runtime %s não suportado: %s", l.Label, config.Runtime)
	}
	if !slices.Contains(l.Frameworks, config.Framework) {
		return fmt.Errorf("framework %s não suportado: %s", l.Label, config.Framework)
	}
	if !slices.Contains(l.Architectures, config.Architecture) {
		return fmt.Errorf("arquitetura %s não suportada: %s", l.Label, config.Architecture)
	}
	return nil
}

func Generate(destination string, config types.ProjectConfig, templates fs.FS) error {
	language, ok := Lookup(config.Language)
	if !ok {
		return fmt.Errorf("linguagem não suportada: %s", config.Language)
	}
	if err := language.validate(config); err != nil {
		return err
	}
	plan, err := language.plan(destination, config)
	if err != nil {
		return err
	}
	return common.Scaffold(destination, templates, plan)
}
