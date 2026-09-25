# gake

A Go CLI that generates CMake files for a project. Run it from the project root:
- `logic/`: flow (new project vs. existing project)
- `utils/`: file collection, parsing, embedded `*.tmpl` templates
- Feature specs are in `specs/`. Read the relevant spec before you start work.

Commands: `go build ./...`, `go test ./...`.

## Workflow: red-green, tests first

Every feature and every bug fix follows this loop:

1. **Red: write the tests first.** Turn the spec into Go tests and fixtures under `logic/tests-assets/`. Do not change
   production code in this step, except for stubs the tests need in order to compile (they must fail at runtime, not at compile time).
2. **Confirm red.** Run `go test ./...` and check that the new tests fail *for the expected reason*.
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
