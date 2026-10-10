package project

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var projectNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9 _.-]*[A-Za-z0-9_]$|^[A-Za-z0-9]$`)

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
	if !projectNamePattern.MatchString(name) || strings.Contains(name, "..") {
		return "", fmt.Errorf("nome inválido: use letras, números, espaços, hífens ou sublinhados; sem separadores de caminho")
	}
	return name, nil
}

func ProjectSlug(name string) string {
	var slug strings.Builder
	lastHyphen := false
	for _, character := range strings.ToLower(name) {
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' {
			slug.WriteRune(character)
			lastHyphen = false
		} else if !lastHyphen && slug.Len() > 0 {
			slug.WriteByte('-')
			lastHyphen = true
		}
	}
	return strings.Trim(slug.String(), "-")
}