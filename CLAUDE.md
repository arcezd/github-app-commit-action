# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Docker-based GitHub Action (Go) that signs commits with a GitHub App and pushes them to a repository. It uses the GitHub REST API directly (not go-github SDK) with JWT-based GitHub App authentication.

## Build & Development Commands

```bash
# Build
go build -o github-app-commit-action .

# Docker build (multi-stage: golang:1.26-alpine → alpine:3.23)
docker build -t github-app-commit-action .

# Lint (uses golangci-lint v2.10)
golangci-lint run ./...

# Tests (run from each module root)
go test -coverprofile=coverage.out ./...
```

**Note:** The project uses two separate Go modules: root (`/`) and helper (`/helper`). CI detects modules dynamically and runs lint/tests on each independently.

## Architecture

**Two-package structure:**
- **Root package (`main`)** — CLI entry point (`main.go`). Parses flags, validates input, generates JWT from GitHub App private key, then delegates to the helper package.
- **Helper package (`github_helper`)** — All GitHub API interaction and git operations.

**Key workflow (CommitAndPush):**
1. Get base commit ref from head branch
2. Stage files and collect modified/new files via local `git` commands
3. Upload each file as a base64-encoded blob via GitHub API
4. Create a tree, then a commit object, then update the branch reference
5. Optionally create/update tags (CreateTagAndPush)

**Authentication flow:** PEM private key → JWT (5-min TTL, signed with RS256) → exchanged for installation access token → used as Bearer token for API calls.

**GitHub API client (`helper/github.go`):** Custom HTTP wrapper (`CallGithubAPI`) — not using any GitHub SDK library. Uses API version `2022-11-28`.

## Conventions

- **Commits:** Conventional commits (commitizen configured in `.cz.toml`, semver, tag format `v${version}`)
- **Version:** Tracked in `version.go` as `BuildVersion` constant
- **Linting:** `.golangci.yml` enables: decorder, goconst, godox, gosec, whitespace, wsl_v5
- **Error handling:** Critical failures panic; recoverable errors return `error`
