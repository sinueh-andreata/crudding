package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"crudding"
	"crudding/internal/generator"
	"crudding/internal/project"
	"crudding/internal/tui"
)

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

func chooseLanguage() (generator.Language, error) {
	languages := generator.Languages()
	names := make([]string, len(languages))
	for index, language := range languages {
		names[index] = language.Name
	}
	name, err := choose("escolha a linguagem", names)
	if err != nil {
		return generator.Language{}, err
	}
	language, _ := generator.Lookup(name)
	return language, nil
}

func chooseRuntime(language generator.Language) (string, error) {
	if len(language.Runtimes) == 0 {
		return "", nil
	}
	return choose("escolha o runtime", language.Runtimes)
}

func promptConfig() (project.Config, error) {
	language, err := chooseLanguage()
	if err != nil {
		return project.Config{}, err
	}
	runtime, err := chooseRuntime(language)
	if err != nil {
		return project.Config{}, err
	}
	framework, err := choose("escolha o framework", language.Frameworks)
	if err != nil {
		return project.Config{}, err
	}
	architecture, err := choose("escolha a arquitetura", language.Architectures)
	if err != nil {
		return project.Config{}, err
	}
	projectName, err := project.ReadProjectName(os.Stdin, os.Stdout)
	if err != nil {
		return project.Config{}, err
	}
	return project.Config{
		ProjectName:  projectName,
		Language:     language.Name,
		Runtime:      runtime,
		Framework:    framework,
		Architecture: architecture,
	}, nil
}

func formatLanguageLine(config project.Config) string {
	if config.Runtime == "" {
		return fmt.Sprintf("Linguagem: %s", config.Language)
	}
	return fmt.Sprintf("Linguagem/motor: %s/%s", config.Language, config.Runtime)
}

func printNewSummary(projectDirectory string, config project.Config) {
	fmt.Printf("\nProjeto gerado em: %s\n", projectDirectory)
	fmt.Println(formatLanguageLine(config))
	fmt.Printf("Framework: %s\nArquitetura: %s\n", config.Framework, config.Architecture)
}

func runNew() error {
	config, err := promptConfig()
	if err != nil {
		return err
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}
	projectDirectory := filepath.Join(workingDirectory, config.ProjectName)
	if err := generator.Generate(projectDirectory, config, crudding.Templates); err != nil {
		return err
	}
	printNewSummary(projectDirectory, config)
	return nil
}
