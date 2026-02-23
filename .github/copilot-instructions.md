# Copilot Instructions for github-app-commit-action

## Repository Overview

This repository is a **GitHub Action** written in **Go** that enables creating signed commits in GitHub repositories using a GitHub App's identity (via JWT + installation tokens). It uses the GitHub REST API directly (no `git` CLI for committing) and runs as a Docker container.

## Project Structure

```
.
├── main.go              # CLI entry point; parses flags and orchestrates the action
├── version.go           # Defines BuildVersion constant
├── go.mod / go.sum      # Go module files (module: github.com/arcezd/github-app-commit-action)
├── Dockerfile           # Multi-stage build: golang:1.26.0-alpine3.23 → alpine:3.23
├── entrypoint.sh        # Docker entrypoint; maps env vars to CLI flags for /bin/action
├── action.yml           # GitHub Action definition (inputs, outputs, Docker runner)
├── .golangci.yml        # golangci-lint v2 configuration
├── helper/              # Sub-module (github.com/arcezd/github-app-commit-action/helper)
│   ├── github.go        # GitHub API calls (JWT, tokens, refs, trees, commits, blobs, tags)
│   ├── github_types.go  # All request/response structs for GitHub API
│   ├── main.go          # High-level commit/tag logic; git diff helpers
│   ├── utils.go         # Shell command execution, GH Actions output/summary helpers
│   ├── go.mod / go.sum  # Helper sub-module dependencies (golang-jwt/jwt)
```

## How the Action Works

1. **Authentication**: Signs a JWT using the GitHub App's RSA private key (`GH_APP_PRIVATE_KEY` env var or `-p` PEM file). Exchanges JWT for an installation access token via the GitHub API.
2. **Detecting changes**: Runs `git add -A` (or `git add -u`) then `git diff --cached --name-only` to find modified files.
3. **Committing via API**: Uploads file contents as blobs, creates a tree, creates a commit, and updates (or creates) the branch reference—all through `api.github.com`.
4. **Tagging**: Optionally creates annotated tags and their references.

## Build, Lint, and Test Commands

### Build
```bash
go build -o action .
```
The Dockerfile handles the full build automatically for the action image.

### Run Tests
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```
Tests live alongside source files. The `helper/` subdirectory is a separate Go module so tests must be run from the correct directory or via `./...` from the root (which covers both modules via the `replace` directive in `go.mod`).

### Lint
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run --timeout=5m
```
Configuration is in `.golangci.yml` (golangci-lint v2 format). Enabled linters include `gosec`, `wsl_v5`, `decorder`, `godox`, `goconst`, and `whitespace`.

### CI
CI runs on every push/PR via `.github/workflows/tests.yml`. It runs lint (`golangci-lint`) and tests with coverage upload to Codecov. The workflow uses `go-version-file: 'go.mod'` to determine the Go version.

## Key Design Decisions and Conventions

- **Two Go modules**: The root module (`github.com/arcezd/github-app-commit-action`) and the `helper/` sub-module (`github.com/arcezd/github-app-commit-action/helper`). The root `go.mod` uses a `replace` directive to point to the local `helper/` directory.
- **No git CLI for commits**: All committing is done through the GitHub REST API. `git` is only used locally to stage and diff files.
- **Panic on fatal errors**: The codebase uses `panic()` for unrecoverable errors (missing credentials, API failures) rather than returning errors from `main`.
- **Environment variables over flags**: The `entrypoint.sh` maps GitHub Actions inputs (env vars) to CLI flags. `GH_APP_PRIVATE_KEY` has priority over `-p` for the private key.
- **`sync.Once` for token initialization**: JWT signing and token initialization are guarded by `sync.Once` to prevent re-initialization.
- **Go version**: `go 1.26` (as specified in `go.mod` and the Dockerfile base image `golang:1.26.0-alpine3.23`).

## Action Inputs and Corresponding CLI Flags

| Action Input              | Env Var                  | CLI Flag | Default                        |
|---------------------------|--------------------------|----------|--------------------------------|
| `github-app-id`           | `GH_APP_ID`              | `-i`     | (required)                     |
| `github-app-private-key`  | `GH_APP_PRIVATE_KEY`     | —        | (env var only)                 |
| `github-app-private-key-file` | `GH_APP_PRIVATE_KEY_FILE` | `-p` | (optional)                  |
| `repository`              | `REPOSITORY`             | `-r`     | (required, format: owner/repo) |
| `branch`                  | `BRANCH`                 | `-b`     | `main`                         |
| `head`                    | `HEAD_BRANCH`            | `-h`     | same as branch                 |
| `message`                 | `COMMIT_MSG`             | `-m`     | `chore: autopublish ${date}`   |
| `force-push`              | `FORCE_PUSH`             | `-f`     | `false`                        |
| `tags`                    | `TAGS`                   | `-t`     | `""`                           |
| `add-new-files`           | `ADD_NEW_FILES`          | `-a`     | `true`                         |
| `coauthors`               | `COAUTHORS`              | `-c`     | `""`                           |

## Known TODOs and Limitations

- Executable file permissions for uploaded files are not yet supported (all files committed as mode `100644`).
- On-behalf-of commits are implemented but commented out.
- File rename detection is not handled correctly.
- Specifying a specific list of files to commit is not yet supported.

## Common Errors and Workarounds

- **"failed to decode PEM block containing private key"**: The `GH_APP_PRIVATE_KEY` env var must contain a valid RSA private key in PEM format (begins with `-----BEGIN RSA PRIVATE KEY-----`).
- **`golangci-lint` config errors**: The `.golangci.yml` uses v2 format. If the linter reports version mismatch errors, ensure you have golangci-lint v2 or later installed.
- **`go: go.mod file indicates go 1.26`**: Requires Go 1.26+. If building locally with an older Go version, update your Go toolchain.
- **Branch doesn't exist**: The action handles this automatically by falling back to `CreateReference` if `UpdateReference` (PATCH) fails.
