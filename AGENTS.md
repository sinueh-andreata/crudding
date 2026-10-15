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

- **Para Node**: runtimes `js` e `ts`; frameworks `raw`, `express`, `fastify`, `nest`; arquiteturas `flat`, `layered`, `modular`, `hexagonal` e `clean_arch`.

---

## 2. Estrutura do Repositório

```text
crudding/
├── cmd/
│   └── crudding/
│       └── main.go              # Ponto de entrada do executável e orquestração do fluxo
│   └── genall/
│       └── main.go              # Gera todas as combinações suportadas para validação
├── internal/
│   ├── generator/
│   │   ├── common/
│   │   │   ├── files.go         # Utilitários de escrita de arquivos e template parsing
│   │   │   ├── scaffold.go      # Criação do projeto (Plan) e gravação do project.crud
│   │   │   └── patch.go         # Alterações em projeto existente (Patch): arquivos novos, pacotes Python e inserções em markers, com rollback
│   │   ├── langs/
│   │   │   ├── golang/generate.go # Gerador específico para projetos Go
│   │   │   ├── node/generate.go   # Gerador específico para projetos Node (JS/TS)
│   │   │   └── python/
│   │   │       ├── generate.go   # Gerador específico para projetos Python
│   │   │       └── endpoint.go   # Layouts de "add endpoint" por arquitetura Python
│   │   ├── generator.go         # Dispatcher central do gerador
│   │   ├── endpoint.go          # Dispatcher de "crudding add endpoint"
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

1. **Seleção Interativa (`cmd/crudding/main.go` & `internal/tui/menu.go`)**:
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
2. `generator.CheckEndpointSupport` valida o manifesto contra o registry e confirma que a linguagem tem `endpoint` (hoje só Python; Go e Node devolvem "ainda não suportado").
3. `project.ReadEndpointName` lê o nome (snake_case).
4. A função `endpoint` da linguagem monta um `common.Patch`: os templates necessários, os pacotes e as inserções nos markers `# crudding:imports` / `# crudding:routes` do módulo de API.
5. `common.Apply` valida tudo antes de gravar (nada pode ser sobrescrito, cada marker aparece uma única vez) e desfaz as alterações se a gravação falhar.

**Para suportar uma nova linguagem**: implemente `endpoint` em `internal/generator/langs/<lang>/endpoint.go`, registre-o em `languages` (`generator.go`), adicione os markers nos templates de roteamento da linguagem e crie os templates correspondentes. Os markers não podem ser removidos dos templates existentes.

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
