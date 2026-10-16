package project

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var projectNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9 _.-]*[A-Za-z0-9_]$|^[A-Za-z0-9]$`)

// endpointNamePattern aceita identificadores snake_case, válidos como nome de
// módulo e de diretório em qualquer linguagem suportada.
var endpointNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)

func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("nenhum nome de projeto informado")
	}
	if !projectNamePattern.MatchString(name) || strings.Contains(name, "..") {
		return fmt.Errorf("nome inválido: use letras, números, espaços, hífens ou sublinhados; sem separadores de caminho")
	}
	return nil
}

func ReadProjectName(input io.Reader, output io.Writer) (string, error) {
	fmt.Fprint(output, "Nome do projeto (nome da pasta): ")
	scanner := bufio.NewScanner(input)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("nenhum nome de projeto informado")
	}
	name := scanner.Text()
	if err := ValidateProjectName(name); err != nil {
		return "", err
	}
	return name, nil
}

func ValidateEndpointName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("nenhum nome de endpoint informado")
	}
	if !endpointNamePattern.MatchString(name) {
		return fmt.Errorf("nome de endpoint inválido %q: use snake_case com letras minúsculas, números e sublinhados, começando por letra", name)
	}
	return nil
}

func ReadEndpointName(input io.Reader, output io.Writer) (string, error) {
	fmt.Fprint(output, "Nome do endpoint (ex.: invoice, order_item): ")
	scanner := bufio.NewScanner(input)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("nenhum nome de endpoint informado")
	}
	name := strings.TrimSpace(scanner.Text())
	if err := ValidateEndpointName(name); err != nil {
		return "", err
	}
	return name, nil
}
