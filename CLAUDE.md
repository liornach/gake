# gake

A Go CLI that generates CMake files for a project. Run it from the project root:
- `logic/`: flow (new project vs. existing project)
- `utils/`: file collection, parsing, embedded `*.tmpl` templates
- Feature specs are in `specs/`. Read the relevant spec before you start work.

Commands: `go build ./...`, `go test ./...`.

## Tooling (configured in `.claude/`)
- **LSP:** the `gopls-lsp` plugin is enabled in `.claude/settings.json` and needs `gopls` on PATH. For Go code, use the
  LSP tool (definitions, references, diagnostics) rather than grep.
- **Session check:** a SessionStart hook runs `.claude/check-tools.sh`, which lists any missing tools (go, gopls,
  staticcheck, goimports, cmake/ctest, a C compiler) with the command to install each. Ask me before installing anything.
- **Before reporting green:** run `go vet ./...` and `gofmt -l .`, which must be clean, and `staticcheck ./...`, which must report
  no *new* findings. The existing U1000 "unused" findings belong to the unfinished root-CMake validation code.
- **Skills:** during the refactor step, use `/simplify`. Before proposing a commit, run `/code-review`. Specs name
  any feature-specific tools they need.
- **CMake check:** when a change affects generated output, build a generated sample project for real
  (`cmake -B build && cmake --build build && ctest --test-dir build`) in a scratch dir, not inside the repo.

## Workflow: red-green, tests first

Every feature and every bug fix follows this loop:

1. **Red: write the tests first.** Write only Go tests (`*_test.go`) and fixtures under `logic/tests-assets/`.
   Write no production code at all, not even stubs, types or error variables. Symbols the tests reference but that
   don't exist yet are the spec for the green phase.
2. **Confirm red.** Run `go vet ./...`. A compile failure caused by the missing symbols counts as red. List the
   undefined symbols the tests expect, so I can review the API they imply.
3. **Stop for approval.** Show me the tests and the failing output, then wait. Do not start implementing until I
   approve the tests.
4. **Green: implement.** Write the least production code that makes the approved tests pass. Iterate until
   `go test ./...` is fully green, including the tests that already existed.
5. **Refactor** while staying green, then report.

Rules:
- **Approved tests are the exit criterion.** The loop ends when every approved test passes. It doesn't end before that, and don't widen it with
  extra goals.
- **Never edit an approved test to make it pass.** If an approved test looks wrong or can't be satisfied, stop and
  explain why. Changes to approved tests need my approval again.
- New tests you find you need during the green phase go through the same approval step before they count.
- Commit only when I ask.
