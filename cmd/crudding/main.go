package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"crudding"
	"crudding/internal/generator"
	"crudding/internal/models"
	"crudding/internal/project"
	"crudding/internal/types"
)

// errCancelled sinaliza que o usuário saiu do menu; não é tratado como falha.
var errCancelled = errors.New("cancelado pelo usuário")

func choose(title string, options []string) (string, error) {
	choice, confirmed, err := models.ChooseOption(title, options)
	if err != nil {
		return "", err
	}
	if !confirmed {
		return "", errCancelled
	}
	return choice, nil
}

func run() error {
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
	config := types.ProjectConfig{
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

func main() {
	if err := run(); err != nil && !errors.Is(err, errCancelled) {
		fmt.Fprintf(os.Stderr, "crudding: %v\n", err)
		os.Exit(1)
	}
}
