# Cursor Prompt Template for Go Projects

## General Principles
- **Do not modify production Go files unless the user explicitly asks for the change.** If the request is ambiguous, ask for clarification first.
- **Summarize planned edits** before touching Go source files so the user understands the impact.
- **Never delete or overwrite user-authored code** unless the user explicitly requests removal.
- **Preserve project structure** (standard layout, configs, docs) and avoid creating miscellaneous files at root.

## Coding Standards
- Run `gofmt` on any Go file you edit.
- After substantive Go changes, run `go test ./...` and report the results (include failures with stack traces).
- Update documentation/docs whenever new behavior or tooling is introduced.
- Maintain existing code comments and add new ones if logic is non-obvious.

## Source Protection & Safety Checks
- Before changing any Go file, confirm:
  1. The user requested the change.
  2. The change list is documented in the response.
  3. Tests will be re-run after the edit.
- For configuration or infrastructure files (Dockerfile, docker-compose, Terraform, CI), explain impact prior to editing.
- Keep secrets or sample env values sanitized—do not introduce real credentials.

## Communication
- When in doubt, ask the user.
- Provide concise diffs or summaries of edits.
- Clearly label any assumptions.

By following these rules, Cursor should only apply intentional, test-verified changes that the user has authorized.
