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
    1. `flat`: tudo na raiz no pacote `main` (`user_model.go`, `user_types.go`, `user_service.go`, `user_handler.go`, `routes.go`, `main.go`, `README.md`).
    2. `layered`: separação clássica em camadas sob `internal/` (`models`, `types`, `service`, `handlers`, `routes`), roteamento em `internal/routes/routes.go` e injeção em `cmd/api/main.go`.
    3. `modular`: *package-by-feature* com módulos (`internal/<feature>/` com `model.go`, `types.go`, `service.go`, `handler.go`), roteamento em `internal/routes/routes.go` e injeção em `cmd/api/main.go`.
    4. `hexagonal`: *Ports & Adapters* com `internal/domain` (entidades e tipos), `internal/ports` (interfaces repository/service), `internal/application` (serviços de aplicação), `internal/adapters/postgres` (adaptador secundário), `internal/handlers` (adaptador primário HTTP), roteamento em `internal/routes/routes.go` e injeção em `cmd/api/main.go`.
    5. `clean arch`: Clean Architecture concêntrica com `internal/domain` (entidades e tipos), `internal/ports` (fronteiras repository/service), `internal/application` (casos de uso), `internal/infrastructure/persistence` (implementação de persistência), `internal/handlers` (controladores HTTP), roteamento em `internal/routes/routes.go` e injeção em `cmd/api/main.go`.
   - **Para Python**:
      - `flat`, `layered`, `modular`, `hexagonal` e `clean_arch`, com estruturas próprias sob `src/`.
   - **Para Node**:
      - `monorepo`
      - `camadas` (Layered / Onion básico)
      - `clean arch` (Clean Architecture / Uncle Bob)
      - `hexagonal` (Ports & Adapters)
      - `4horse` (Routes, Controllers, Services, Repositories)

Cada arquitetura em Go possui seu próprio `README.md` detalhado explicando o funcionamento, princípios e comandos de execução.

- **Para Node**: runtimes `js` e `ts`; frameworks `raw`, `express`, `fastify`, `nest`; arquiteturas `flat`, `layered`, `modular`, `hexagonal` e `clean_arch`.

---

## 2. Estrutura do Repositório

```text
crudding/
├── cmd/
│   └── crudding/
│       └── main.go              # Ponto de entrada mínimo (delega para internal/cli)
│   └── genall/
│       └── main.go              # Gera todas as combinações suportadas para validação
├── internal/
│   ├── cli/
│   │   ├── cli.go               # Registro de comandos, roteador Run e geração automática de --help
│   │   ├── new.go               # Comando padrão de criação de projeto (runNew)
│   │   ├── add_endpoint.go      # Comando "crudding add endpoint" (runAddEndpoint)
│   │   └── remove_endpoint.go   # Comando "crudding remove endpoint" (runRemoveEndpoint)
│   ├── generator/
│   │   ├── common/
│   │   │   ├── files.go         # Utilitários de escrita de arquivos e template parsing
│   │   │   ├── scaffold.go      # Criação do projeto (Plan) e gravação do project.crud
│   │   │   ├── patch.go         # Alterações em projeto existente (Patch): arquivos novos, pacotes Python e inserções em markers, com rollback
│   │   │   └── remove.go        # Remoção de endpoint (RemovalPlan): exclusão de arquivos, limpeza de linhas e diretórios vazios
│   │   ├── langs/
│   │   │   ├── golang/
│   │   │   │   ├── generate.go  # Gerador específico para projetos Go
│   │   │   │   └── endpoint.go  # Layouts de "add endpoint" e "remove endpoint" por arquitetura Go
│   │   │   ├── node/generate.go   # Gerador específico para projetos Node (JS/TS)
│   │   │   └── python/
│   │   │       ├── generate.go   # Gerador específico para projetos Python
│   │   │       └── endpoint.go   # Layouts de "add endpoint" e "remove endpoint" por arquitetura Python
│   │   ├── generator.go         # Dispatcher central do gerador
│   │   ├── endpoint.go          # Dispatcher de "crudding add endpoint" e "crudding remove endpoint"
│   ├── manifest/
│   │   └── manifest.go          # Leitura/escrita do project.crud (JSON) e busca nos diretórios pais
│   ├── project/
│   │   ├── config.go            # Configuração compartilhada do projeto
│   │   ├── name.go              # Leitura e sanitização de nome do projeto
│   │   └── project.go           # Utilitário de slugify (ProjectSlug)
│   ├── test/                    # Testes de integração e unidade de internal/
│   └── tui/
│       └── menu.go              # Menu Bubble Tea para seleção interativa
├── templates/                   # Arquivos de template (.tmpl) para cada stack
│   ├── go/
│   │   ├── base/                # common (.gitignore, go.mod, go.sum) e arquiteturas (flat, layered, modular, hexagonal, clean_arch)
│   │   └── frameworks/          # entrypoints por framework (chi, fiber, gin, raw)
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

1. **Seleção Interativa (`internal/cli/new.go` & `internal/tui/menu.go`)**:
   - Pergunta a **Linguagem** (`python`, `go`, `node`).
   - Se `node`, pergunta o **Runtime** (`js`, `ts`).
   - Pergunta o **Framework** compatível com a linguagem selecionada.
   - Pergunta a **Arquitetura** (com opções dinâmicas dependendo da linguagem selecionada via `architectureOptionsFor`).
   - Lê e valida o **Nome do Projeto** via prompt de texto.
2. **Validação de Segurança (`internal/project/name.go`)**:
   - Rejeita caracteres inválidos e tentativas de path traversal (`..`, barras, etc.).
   - Garante que diretórios existentes nunca sejam sobrescritos (`internal/generator/common/files.go`).
3. **Geração dos Arquivos (`internal/generator/` e `internal/generator/langs/`)**:
   - Define a estrutura de diretórios baseado no layout da arquitetura escolhida.
   - Processa os templates embutidos utilizando `text/template` e grava no diretório de destino.
   - Em caso de falha no meio do processo, o diretório gerado é limpo automaticamente (`defer os.RemoveAll`).

4. **Manifesto**: `common.Scaffold` grava `project.crud` (JSON, `internal/manifest`) na raiz de todo projeto gerado.

### `crudding add endpoint`

1. `manifest.Find` procura o `project.crud` a partir do diretório atual, subindo pelos pais.
2. `generator.CheckEndpointSupport` valida o manifesto contra o registry e confirma que a linguagem tem `endpoint` (suportado em Python e Go; Node devolve "ainda não suportado").
3. `project.ReadEndpointName` lê o nome (snake_case).
4. A função `endpoint` da linguagem monta um `common.Patch`: os templates necessários, os pacotes e as inserções nos markers `# crudding:imports` / `# crudding:routes` do módulo de API.
5. `common.Apply` valida tudo antes de gravar (nada pode ser sobrescrito, cada marker aparece uma única vez) e desfaz as alterações se a gravação falhar.

### `crudding remove endpoint`

1. `manifest.Find` localiza a raiz do projeto e lê o `project.crud`.
2. `generator.CheckEndpointSupport` confirma suporte à linguagem.
3. `project.ReadEndpointName` lê o nome (ou via argumentos `name:<nome>`, `n:<nome>` ou posicional).
4. Verifica se o endpoint está registrado no `project.crud`.
5. A função `removePlan` da linguagem monta um `common.RemovalPlan`: arquivos a deletar, diretórios que devem ser excluídos se ficarem vazios e linhas a remover dos arquivos de rotas/imports.
6. `common.Remove` remove os arquivos, limpa as linhas das rotas, limpa pastas vazias e atualiza o manifesto `project.crud` com rollback caso haja erro na gravação.

**Para suportar uma nova linguagem**: implemente `endpoint` e `removePlan` em `internal/generator/langs/<lang>/endpoint.go`, registre-os em `languages` (`generator.go`), adicione os markers nos templates de roteamento da linguagem e crie os templates correspondentes. Os markers não podem ser removidos dos templates existentes.

---

## 4. Comandos e Ambiente de Desenvolvimento

### Comandos Go Padrão
```sh
# Executar a CLI em desenvolvimento
go run ./cmd/crudding

# Adicionar um endpoint a um projeto gerado (rodar dentro do projeto)
go run /caminho/do/crudding/cmd/crudding add endpoint

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
   - Adicione os casos de teste em `internal/test/`.
