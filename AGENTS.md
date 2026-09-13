# AGENTS.md

Guidance for coding agents working on lunchmoney.

## Project Overview

Go client library for the Lunch Money personal finance developer API (`github.com/icco/lunchmoney`).

## Commands

```sh
go test ./...    # Run tests
go vet ./...     # Vet code
go build ./...   # Verify compilation
```

## Conventions

- Standard Go client library design with structured types, error handling, and unit test coverage.
- PR titles and commits must follow Conventional Commits with lowercase subjects.
- Ensure all tests pass before submitting PRs.
- Never hardcode or commit Lunch Money API access tokens.
