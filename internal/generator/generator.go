package generator

import (
	"fmt"
	"io/fs"
	"slices"

	"crudding/internal/generator/common"
	"crudding/internal/generator/langs/golang"
	"crudding/internal/generator/langs/node"
	"crudding/internal/generator/langs/python"
	"crudding/internal/project"
)

// Language descreve uma linguagem suportada. É a fonte única das opções
// exibidas no menu, da validação e da matriz de testes.
type Language struct {
	Name          string   // valor de project.Config.Language
	Label         string   // nome exibido em mensagens de erro
	Runtimes      []string // vazio quando a linguagem não pergunta runtime
	Frameworks    []string
	Architectures []string
	plan          func(destination string, config project.Config) (common.Plan, error)
	// endpoint monta o Patch de "crudding add endpoint"; nil enquanto a
	// linguagem não suporta o comando.
	endpoint func(root string, config project.Config, name string) (common.Patch, error)
	// removePlan monta o RemovalPlan de "crudding remove endpoint"; nil enquanto
	// a linguagem não suporta o comando.
	removePlan func(root string, config project.Config, name string) (common.RemovalPlan, error)
}

var languages = []Language{
	{Name: "python", Label: "Python", Frameworks: python.Frameworks, Architectures: python.Architectures, plan: python.Plan, endpoint: python.Endpoint, removePlan: python.RemovePlan},
	{Name: "go", Label: "Go", Frameworks: golang.Frameworks, Architectures: golang.Architectures, plan: golang.Plan, endpoint: golang.Endpoint, removePlan: golang.RemovePlan},
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

func (l Language) validate(config project.Config) error {
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

func Generate(destination string, config project.Config, templates fs.FS) error {
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
	return common.Scaffold(destination, config, templates, plan)
}
