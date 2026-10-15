package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"crudding"
	"crudding/internal/generator"
	"crudding/internal/manifest"
	"crudding/internal/project"
	"crudding/internal/tui"
)

// errCancelled sinaliza que o usuário saiu do menu; não é tratado como falha.
var errCancelled = errors.New("cancelado pelo usuário")

func choose(title string, options []string) (string, error) {
	choice, confirmed, err := tui.ChooseOption(title, options)
	if err != nil {
		return "", err
	}
	if !confirmed {
		return "", errCancelled
	}
	return choice, nil
}

func runNew() error {
	languages := generator.Languages()
	names := make([]string, len(languages))
	for index, language := range languages {
		names[index] = language.Name
	}
	name, err := choose("escolha a linguagem", names)
	if err != nil {
		return err
	}
	language, _ := generator.Lookup(name)

	runtime := ""
	if len(language.Runtimes) > 0 {
		if runtime, err = choose("escolha o runtime", language.Runtimes); err != nil {
			return err
		}
	}
	framework, err := choose("escolha o framework", language.Frameworks)
	if err != nil {
		return err
	}
	architecture, err := choose("escolha a arquitetura", language.Architectures)
	if err != nil {
		return err
	}
	projectName, err := project.ReadProjectName(os.Stdin, os.Stdout)
	if err != nil {
		return err
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}
	config := project.Config{
		ProjectName:  projectName,
		Language:     language.Name,
		Runtime:      runtime,
		Framework:    framework,
		Architecture: architecture,
	}
	projectDirectory := filepath.Join(workingDirectory, projectName)
	if err := generator.Generate(projectDirectory, config, crudding.Templates); err != nil {
		return err
	}

	fmt.Printf("\nProjeto gerado em: %s\n", projectDirectory)
	if runtime != "" {
		fmt.Printf("Linguagem/motor: %s/%s\n", language.Name, runtime)
	} else {
		fmt.Printf("Linguagem: %s\n", language.Name)
	}
	fmt.Printf("Framework: %s\nArquitetura: %s\n", framework, architecture)
	return nil
}

func runAddEndpoint() error {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}
	root, m, err := manifest.Find(workingDirectory)
	if err != nil {
		return err
	}
	fmt.Printf("Projeto %s: %s\n", root, m.Stack())
	if err := generator.CheckEndpointSupport(m); err != nil {
		return err
	}
	name, err := project.ReadEndpointName(os.Stdin, os.Stdout)
	if err != nil {
		return err
	}
	result, err := generator.AddEndpoint(root, m, name, crudding.Templates)
	if err != nil {
		return err
	}

	fmt.Printf("\nEndpoint %s criado.\n", name)
	for _, file := range result.Created {
		fmt.Printf("  criado     %s\n", file)
	}
	for _, file := range result.Modified {
		fmt.Printf("  alterado   %s\n", file)
	}
	return nil
}

const usage = `uso:
  crudding                 gera um novo projeto (menu interativo)
  crudding add endpoint    adiciona um endpoint CRUD ao projeto atual (lê o project.crud)
`

func main() {
	var run func() error
	switch args := os.Args[1:]; {
	case len(args) == 0:
		run = runNew
	case slices.Equal(args, []string{"add", "endpoint"}):
		run = runAddEndpoint
	case len(args) == 1 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help"):
		fmt.Print(usage)
		return
	default:
		fmt.Fprintf(os.Stderr, "crudding: comando desconhecido: %s\n%s", strings.Join(args, " "), usage)
		os.Exit(2)
	}
	if err := run(); err != nil && !errors.Is(err, errCancelled) {
		fmt.Fprintf(os.Stderr, "crudding: %v\n", err)
		os.Exit(1)
	}
}
