# Spec: C language support

## Goal
gake currently assumes C++23 + modules. Add a second "language profile" for plain C projects. The profile is chosen
from the root `CMakeLists.txt` `project()` language (existing projects), or auto-detected from source files (new projects).
The C++ behaviour and its generated output must stay **byte-identical** (existing tests must pass unchanged).

## Workflow
This feature follows the red-green loop in `CLAUDE.md`, split into three gated phases. Do not start a phase until the
previous one's gate has passed.

### Phase 0: clean up the repo (before any feature work)
The working tree is dirty. The code was moved from the repo root into the `logic/` and `utils/` packages, but git only
sees the deletions, because `.gitignore` starts with `*` and so ignores the new directories, the `*.tmpl` templates,
`specs/` and `CLAUDE.md`. As a result, a fresh clone doesn't build: `go:embed` can't find the templates.
1. Fix `.gitignore`:
   - un-ignore directories (`!*/`), `*.tmpl`, `*.md` and every file under `logic/tests-assets/`
   - keep ignoring the `gake` binary and the test sandbox `logic/.tmp/`
2. Stage everything and check `git status`. It should show the root files as *renamed* into `logic/` and `utils/`, plus the new
   templates, test assets, `specs/`, `CLAUDE.md` and `.gitignore`. It should show no binary and no `.tmp/`.
3. Check that the baseline is green: `go build ./... && go test ./...`.
4. Before committing, record a **functional baseline**: build the current working tree into a scratch dir. Run the binary on
   scratch copies of `project-stubs/new` (with `--projname=baseline`) and `project-stubs/existing`, and keep the
   generated `CMakeLists.txt` files.
5. Propose the commit (restructure only, no behaviour change) and wait for my approval.
6. **Verify what git actually holds.** The working tree can hide missing files, so run these checks against a fresh clone
   (`git clone <repo> <scratch>/clone`). That is exactly what the remote receives.
   - **Nothing missing:** every file in the working tree is tracked or deliberately ignored. Compare
     `git ls-files` with the files on disk. The only ones allowed to be untracked are the `gake` binary and `logic/.tmp/`. Check by name that
     the embedded `utils/*.tmpl` and all of `logic/tests-assets/**` are present in the clone, including the empty stub files.
   - **Builds and tests pass from the clone:** `go build ./... && go test ./...` inside the clone.
   - **Same behaviour:** build `gake` from the clone, run it on the same stub copies as in step 4, and `diff -r`
     the output against the baseline. It must be identical.
7. **Add a remote (required).** There is currently no git remote, and the only branch is `master`. Ask me for the remote URL
   and the target branch name. If the remote repository doesn't exist yet, propose creating it (for example with
   `gh repo create`, private by default) and wait for my approval. Add it as `origin`, then push only after I approve.
8. **Verify the remote.** Repeat every check in step 6 against a fresh clone of the *remote URL* (not the local repo).
   Also confirm that the local branch tracks the remote one and that `git status` shows no commits ahead of or behind it.

**Gate:** the restructure is committed and pushed, `origin` is configured, and `git status` is clean and in sync with
the remote. Every step 6 check passes against a clone of the remote. Phase 1 does not start without a working remote.

### Phase 1: red
Write the tests and fixtures listed under [Tests](#tests). Confirm they fail for the expected reason, while the existing
tests still pass.

**Gate:** I review and approve the tests. From then on, the approved tests are the reference, and the agent's loop ends when
all of them pass. Changing an approved test needs my approval again.

### Phase 2: green
Implement the rest of this spec until every approved test passes, then refactor while staying green.

**Gate:** see [Done when](#done-when).

## Current behaviour (reference)
- Entry: `logic.RunAtUserRoot` → `gake(root)` (`logic/logic.go`).
- No root `CMakeLists.txt` → `initNewProj(root, projname)`: root + src + tests CMakeLists.
- Root exists → `updateExistingProj(root)`: regenerates only `src/` and `tests/` CMakeLists.
- Templates embedded in `utils/`: `CMakeLists.txt.tmpl`, `SrcCMakeLists.txt.tmpl`, `TestsCMakeLists.txt.tmpl`.
- File collection: `utils.CollectAllFilesWithExt(dir, "cppm")`, `utils.CollectTests(dir)` (hard-coded `.cpp`).
- Project name parsing: `utils.getProjNameFromCmake` (regex on `project(`).

## Target C project layout
```
<root>/
  CMakeLists.txt        # generated once (new project only)
  include/*.h           # public headers
  src/*.c               # compiled into OBJECT library "Objects"
  src/CMakeLists.txt    # regenerated every run
  tests/*.c             # one test executable per file
  tests/CMakeLists.txt  # regenerated every run
```

## Language detection
Introduce a language type (e.g. `utils.Language` with `LangCXX`, `LangC`) and a single detection entry point.

**Existing project (root CMakeLists present):** read the language list of `project()`.
- Support both forms: `project(name C)` and `project(name VERSION 1.0 LANGUAGES C)`. Case-insensitive.
- Contains `CXX` → CXX profile (even if `C` is also present).
- Only `C` → C profile.
- No languages listed → CXX profile (backwards compatible with today).
- The CMake file is the source of truth. Do not cross-check the source files.

**New project (no root CMakeLists):** inspect `src/`.
- Only `.c` files → C.
- Only `.cppm` files → CXX.
- Both → error: "mixed C and C++ module sources in src/, cannot infer language".
- Neither (empty or missing `src/`) → CXX (today's default).

## Language profile
Replace the hard-coded extensions/templates with a per-language profile, for example:

| field            | CXX                        | C                        |
|------------------|----------------------------|--------------------------|
| src extension    | `.cppm`                    | `.c`                     |
| test extension   | `.cpp`                     | `.c`                     |
| root template    | `CMakeLists.txt.tmpl`      | `CCMakeLists.txt.tmpl`   |
| src template     | `SrcCMakeLists.txt.tmpl`   | `CSrcCMakeLists.txt.tmpl`|
| tests template   | `TestsCMakeLists.txt.tmpl` | same (language-agnostic) |
| extra dirs       | —                          | `include/`               |

The logic layer asks for the profile once and passes it down. It must not branch on language all over the code.
Follow the user's global code style (named functions instead of branching ternaries, `is`-prefixed booleans, `get`-prefixed pure getters).

## Generated C output (exact)
Root (`CCMakeLists.txt.tmpl`):
```cmake
cmake_minimum_required(VERSION 3.28)
project({{.ProjName}} C)

set(CMAKE_C_STANDARD 23)
set(CMAKE_C_STANDARD_REQUIRED True)

enable_testing()

add_subdirectory(./src)
add_subdirectory(./tests)
```
src (`CSrcCMakeLists.txt.tmpl`):
```cmake
cmake_minimum_required(VERSION 3.28)

add_library(Objects OBJECT)

target_sources(Objects
    PRIVATE
{{- range .Sources }}
    {{ . }}
{{- end }}
)

target_include_directories(Objects
    PUBLIC
    ${PROJECT_SOURCE_DIR}/include
)
```
tests: reuse `TestsCMakeLists.txt.tmpl` unchanged, with `tests/*.c` collected. Test names use the same `toCamelCase`.

## Behaviour changes
- `initNewProj`: detect the language, then create `include/` for C and render the profile's root, src and tests templates.
- `updateExistingProj`: detect the language from the root CMakeLists, then regenerate src and tests with the profile. Create `include/` if it is missing (C only).
- `CollectTests` takes the extension from the profile instead of hard-coding `.cpp`.
- Language detection errors propagate as errors (no `panic`).

## Tests
Add stubs under `logic/tests-assets/project-stubs/`:
- `new-c/`: `src/{core,utils}.c`, `include/{core,utils}.h`, `tests/{core,utils}.c`, no root CMakeLists.
- `existing-c/`: the same, plus a root CMakeLists with `project(vm C)` and some user content after the core block.
- `new-mixed/`: `src/a.c` + `src/b.cppm`. Must produce an error.

Test cases:
- `TestNewCProj`: the root, src and tests output match the exact strings above. `include/` exists.
- `TestExistingCProj`: src and tests are regenerated with C content. The root file is untouched (byte-compare before and after).
- `TestNewMixedProjFails`.
- Unit tests for `project()` language parsing: `project(x C)`, `project(x CXX)`, `project(x C CXX)`,
  `project(x VERSION 1.2 LANGUAGES C)`, `project(x)`, lowercase `project(x c)`.
- Existing `TestNewProj` must pass unchanged.
- Fix `TestExistingProk`: it currently runs `updateExistingProj` on the sandbox root, which has no CMakeLists. Detection
  needs one, so point it at `project-stubs/existing`.

## Out of scope
- Mixed C/C++ projects (C sources plus C++ modules in one target).
- Generating an executable for `main.c`/`main.cpp`.
- Preserving user content in the root CMakeLists (`parseRootCmake`/`validateExistingRootCmake` stay unused).
- A `--lang` override flag.

## Known issues
- `parseRootCmake` passes a file path to `getProjNameFromCmake`, which expects a directory. Don't copy this pattern when reusing the parser.

## Done when
- Every approved Phase 1 test passes, and so does every test that existed before: `go build ./... && go test ./...`.
- Manual smoke check: running `gake` in a scratch dir that contains only `src/*.c`, `include/*.h` and `tests/*.c` produces
  a project that builds with `cmake -B build && cmake --build build && ctest --test-dir build`.
