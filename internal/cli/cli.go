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

var helpFlags = []string{"-h", "--help", "help"}

// Command descreve um comando disponível na CLI.
type Command struct {
	Args    string
	Summary string
	Run     func() error
}

func (c Command) label() string {
	return strings.TrimSpace("crudding " + c.Args)
}

func (c Command) matches(input string) bool {
	return c.Args == input
}

var commands = []Command{
	{
		Args:    "",
		Summary: "gera um novo projeto (menu interativo)",
		Run:     runNew,
	},
	{
		Args:    "add endpoint",
		Summary: "adiciona um endpoint CRUD ao projeto atual (lê o project.crud)",
		Run:     runAddEndpoint,
	},
}

// Commands devolve a lista de comandos registrados na CLI.
func Commands() []Command {
	return slices.Clone(commands)
}

// Usage gera o texto de ajuda a partir dos comandos registrados.
func Usage() string {
	var b strings.Builder
	b.WriteString("uso:\n")
	for _, cmd := range commands {
		fmt.Fprintf(&b, "  %-21s    %s\n", cmd.label(), cmd.Summary)
	}
	return b.String()
}

// Run despacha os argumentos da linha de comando e devolve o código de saída.
func Run(args []string, stdout, stderr io.Writer) int {
	input := strings.Join(args, " ")
	if slices.Contains(helpFlags, input) {
		fmt.Fprint(stdout, Usage())
		return 0
	}

	cmd, ok := findCommand(input)
	if !ok {
		fmt.Fprintf(stderr, "unknown command use `crudding help` or `crudding --help` or `crudding -h` for usage\n")
		return 2
	}

	return execute(cmd, stderr)
}

func findCommand(input string) (Command, bool) {
	index := slices.IndexFunc(commands, func(cmd Command) bool {
		return cmd.matches(input)
	})
	if index < 0 {
		return Command{}, false
	}
	return commands[index], true
}

func execute(cmd Command, stderr io.Writer) int {
	err := cmd.Run()
	if err == nil || errors.Is(err, errCancelled) {
		return 0
	}
	fmt.Fprintf(stderr, "crudding: %v\n", err)
	return 1
}
