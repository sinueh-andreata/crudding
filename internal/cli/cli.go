package cli

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

// errCancelled sinaliza que o usuário saiu do menu; não é tratado como falha.
var errCancelled = errors.New("cancelado pelo usuário")

// Command descreve um comando ou subcomando disponível na CLI.
type Command struct {
	Path    []string
	Summary string
	Run     func() error
}

var commands = []Command{
	{
		Path:    nil,
		Summary: "gera um novo projeto (menu interativo)",
		Run:     runNew,
	},
	{
		Path:    []string{"add", "endpoint"},
		Summary: "adiciona um endpoint CRUD ao projeto atual (lê o project.crud)",
		Run:     runAddEndpoint,
	},
}

// Commands devolve a lista de comandos registrados na CLI.
func Commands() []Command {
	return slices.Clone(commands)
}

func (c Command) invocation() string {
	if len(c.Path) == 0 {
		return "crudding"
	}
	return "crudding " + strings.Join(c.Path, " ")
}

// Usage gera o texto de ajuda a partir dos comandos registrados.
func Usage() string {
	maxWidth := 0
	for _, cmd := range commands {
		if w := len(cmd.invocation()); w > maxWidth {
			maxWidth = w
		}
	}

	var b strings.Builder
	b.WriteString("uso:\n")
	for _, cmd := range commands {
		fmt.Fprintf(&b, "  %-*s    %s\n", maxWidth, cmd.invocation(), cmd.Summary)
	}
	return b.String()
}

// Run despacha os argumentos da linha de comando e devolve o código de saída.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help") {
		fmt.Fprint(stdout, Usage())
		return 0
	}

	index := slices.IndexFunc(commands, func(cmd Command) bool {
		return slices.Equal(cmd.Path, args)
	})
	if index < 0 {
		fmt.Fprintf(stderr, "crudding: comando desconhecido: %s\n%s", strings.Join(args, " "), Usage())
		return 2
	}

	if err := commands[index].Run(); err != nil && !errors.Is(err, errCancelled) {
		fmt.Fprintf(stderr, "crudding: %v\n", err)
		return 1
	}
	return 0
}

