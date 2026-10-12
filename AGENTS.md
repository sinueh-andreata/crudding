# AGENTS.md

Guia de contexto, arquitetura e instruções para agentes de IA atuando no repositório **crudding**.

---

## 1. Visão Geral do Projeto

**crudding** é uma ferramenta de linha de comando (CLI) interativa desenvolvida em **Go 1.27** usando **Bubble Tea** (`github.com/charmbracelet/bubbletea`). Seu objetivo é fazer o scaffolding de novos projetos em diferentes linguagens, frameworks e padrões arquiteturais a partir de templates pré-definidos embutidos diretamente no binário (`embed.FS`).

### Capacidades Suportadas
- **Linguagens**:
  - `python`: frameworks `raw`, `fastapi`, `django`, `flask`.
  - `go`: frameworks `raw`, `gin`, `fiber`, `chi`.
  - `node`: runtimes `js` e `ts`; frameworks `raw`, `express`, `nest`, `fastify`.
- **Arquiteturas (Padrões de Pastas)**:
  - **Para Golang**:
    1. `flat`: tudo no pacote `main` (`parse.go`, `parse_test.go`, `report.go`, `main.go`, `README.md`).
    2. `layered`: separação em camadas sob `internal/` (`router`, `handlers`, `services`, `models`) e `cmd/api/main.go`.
    3. `modular`: *package-by-feature* com módulos (`order`, `user`), camada centralizadora `internal/router`, plataforma compartilhada (`platform/database`, `platform/auth`) e múltiplos executáveis (`cmd/api`, `cmd/worker`).
    4. `hexagonal`: *Ports & Adapters* com `internal/domain`, `internal/ports`, `internal/adapters` (`http`, `grpc`, `postgres`) e múltiplos entrypoints (`cmd/rest`, `cmd/grpc`).
    5. `clean arch`: Clean Architecture de Uncle Bob com `domain`, `application`, `ports`, `infrastructure` (`persistence`, `http`, `messaging`), utilitários em `pkg/validator` e `migrations/`.
   - **Para Python**:
      - `flat`, `layered`, `modular`, `hexagonal` e `clean_arch`, com estruturas próprias sob `src/`.
   - **Para Node**:
      - `monorepo`
      - `camadas` (Layered / Onion básico)
      - `clean arch` (Clean Architecture / Uncle Bob)
      - `hexagonal` (Ports & Adapters)
      - `4horse` (Routes, Controllers, Services, Repositories)

Cada arquitetura em Go possui seu próprio `README.md` detalhado explicando o funcionamento, princípios e comandos de execução.

---

## 2. Estrutura do Repositório

```text
crudding/
├── cmd/
│   └── crudding/
│       └── main.go              # Ponto de entrada do executável e orquestração do fluxo
├── internal/
│   ├── generator/
│   │   ├── common/
│   │   │   └── files.go         # Utilitários de escrita de arquivos, template parsing e pacotes Python
│   │   ├── golang/
│   │   │   └── generate.go      # Gerador específico para projetos Go (flat, layered, modular, hexagonal, clean arch)
│   │   ├── node/
│   │   │   └── generate.go      # Gerador específico para projetos Node (JS/TS)
│   │   ├── python/
│   │   │   └── generate.go      # Gerador específico para projetos Python
│   │   ├── generator.go         # Dispatcher central do gerador
│   │   └── generator_test.go    # Testes cobrindo combinações de matriz de templates
│   ├── models/
│   │   ├── menu.go              # Componente TUI (Bubble Tea) para seleção interativa
│   │   └── menu_test.go         # Testes de unidade do modelo de menu
│   ├── project/
│   │   ├── name.go              # Leitura e sanitização de nome do projeto
│   │   ├── name_test.go         # Testes de validação de nome e regex
│   │   └── project.go           # Utilitário de slugify (ProjectSlug)
│   └── types/
│       └── project.go           # Estruturas de dados compartilhadas (ProjectConfig)
├── templates/                   # Arquivos de template (.tmpl) para cada stack
│   ├── go/
│   │   ├── base/                # go.mod, go.sum, .gitignore
│   │   └── architectures/       # flat, layered, modular, hexagonal, clean_arch
│   ├── node/
│   └── python/
├── assets.go                    # Declaração do embed.FS para a pasta templates/
├── go.mod                       # Definição de módulo e dependências
├── go.sum
├── LICENSE
└── README.md
```

---

## 3. Fluxo de Execução

1. **Seleção Interativa (`cmd/crudding/main.go` & `internal/models/menu.go`)**:
   - Pergunta a **Linguagem** (`python`, `go`, `node`).
   - Se `node`, pergunta o **Runtime** (`js`, `ts`).
   - Pergunta o **Framework** compatível com a linguagem selecionada.
   - Pergunta a **Arquitetura** (com opções dinâmicas dependendo da linguagem selecionada via `architectureOptionsFor`).
   - Lê e valida o **Nome do Projeto** via prompt de texto.
2. **Validação de Segurança (`internal/project/name.go`)**:
   - Rejeita caracteres inválidos e tentativas de path traversal (`..`, barras, etc.).
   - Garante que diretórios existentes nunca sejam sobrescritos (`internal/generator/common/files.go`).
3. **Geração dos Arquivos (`internal/generator/*`)**:
   - Define a estrutura de diretórios baseado no layout da arquitetura escolhida.
   - Processa os templates embutidos utilizando `text/template` e grava no diretório de destino.
   - Em caso de falha no meio do processo, o diretório gerado é limpo automaticamente (`defer os.RemoveAll`).

---

## 4. Comandos e Ambiente de Desenvolvimento

### Comandos Go Padrão
```sh
# Executar a CLI em desenvolvimento
go run ./cmd/crudding

# Executar testes unitários
go test ./...

# Instalar o binário localmente
go install ./cmd/crudding
```

### Nota sobre Ambientes em Containers / Flatpak (OSTree)
Se você estiver executando em uma sandbox ou terminal de IDE flatpak (ex: VS Code Flatpak) onde o binário `go` está apenas no host:
```sh
flatpak-spawn --host sh -c 'cd /caminho/do/projeto && go test ./...'
```

---

## 5. Diretrizes para Agentes de Código

Ao implementar novas features, correções ou refatorações neste repositório:

1. **Zero dependências extras desnecessárias**:
   - O projeto utiliza a biblioteca padrão do Go sempre que possível (`text/template`, `io/fs`, `embed`, `path/filepath`). A única dependência externa principal é `bubbletea`.
2. **Preservação de Templates e Embed**:
   - Ao adicionar novos templates dentro de `templates/`, certifique-se de que os nomes e caminhos correspondam exatamente às referências nos geradores de cada linguagem.
   - Lembre-se de que `assets.go` utiliza `//go:embed all:templates`, portanto os caminhos dentro do `embed.FS` contêm o prefixo `templates/...`.
3. **Tratamento de Erros e Limpeza**:
   - Sempre use `fmt.Errorf("...: %w", err)` para empacotamento semântico de erros.
   - Mantenha a proteção contra sobrescrita acidental de diretórios.
4. **Adicionando Novos Frameworks ou Arquiteturas**:
   - Atualize `frameworkOptionsFor` e `architectureOptionsFor` em `cmd/crudding/main.go`.
   - Adicione no mapa de arquiteturas/layouts no respectivo pacote (`internal/generator/<lang>/generate.go`).
   - Crie os templates correspondentes sob `templates/<lang>/`.
   - Adicione os casos de teste em `internal/generator/generator_test.go`.
