# crudding

Interactive CLI written in Go 1.27 that generates projects from templates embedded in the executable.

## Install the `crudding` Command

```sh
make install
```

## Run

After the instalation you can use the command "crudding" to create a new project 

```sh
crudding
```

Use the arrow keys to navigate, Space to select, and Enter to continue. Choose Python, Go, or Node, then select a framework and architecture. Python architectures are `flat`, `layered`, `modular`, `hexagonal`, and `clean_arch`. Go architectures are `flat`, `layered`, `modular`, `hexagonal`, and `clean arch`. Node frameworks are `raw`, `express`, `fastify`, and `nest`; its architectures are `flat`, `layered`, `modular`, `hexagonal`, and `clean_arch`. Also choose JavaScript (`js`) or TypeScript (`ts`). Before generating the project, crudding prompts for a name and creates a directory with that name in the current working directory. Existing directories are never overwritten. Press `q` or Esc to exit.

Template files are compiled into the executable, so no templates are downloaded during project generation. To install dependencies, follow the README generated for the project: Python uses `python -m pip install -r requirements.txt`, Go uses `go mod tidy`, and Node uses `npm install`.

Every generated project contains a `project.crud` file at its root recording how it was generated:

```json
{
  "version": 1,
  "name": "my-api",
  "language": "python",
  "framework": "fastapi",
  "architecture": "modular"
}
```

## Add an Endpoint

```sh
cd my-api
crudding add endpoint
```

crudding looks for `project.crud` in the current directory or any parent, asks for the endpoint name (snake_case, e.g. `invoice`, `order_item`) and creates a CRUD resource following the project's architecture: `GET/POST /<name>` and `GET/PATCH/DELETE /<name>/<id>` (underscores become hyphens in the URL). Each layer is created as a stub that raises `NotImplementedError`, so the routes answer `501` until you implement them. The new routes are registered in the dispatcher at the `# crudding:imports` and `# crudding:routes` markers; keep these lines in place. Existing files are never overwritten, and if anything fails no change is kept.

Currently supported for Python (all architectures and frameworks). Go and Node projects report that the command is not supported yet.

## Project Structure

```text
cmd/crudding/                    application entry point
cmd/genall/                      generate all supported combinations
internal/generator/              generation dispatcher
internal/generator/common/       shared helpers
internal/generator/langs/        language-specific generators
internal/manifest/               project.crud manifest handling
internal/project/                project configuration, naming, and slug
internal/test/                   internal unit and integration tests
internal/tui/                    interactive terminal menu
templates/                       templates embedded into the binary by assets.go
```
