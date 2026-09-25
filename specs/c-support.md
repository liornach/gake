# Spec: C language support

## Goal
gake currently assumes C++23 + modules. Add a second "language profile" for plain C projects. The profile is chosen
from the root `CMakeLists.txt` `project()` language (existing projects), or auto-detected from source files (new projects).
The C++ behaviour and its generated output must stay **byte-identical** (`TestNewProj` passes unchanged; `TestExistingProk` is rewritten to check that output exactly).

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

After every test is approved, check the **test cleanup**:
- New tests copy the single stub they need into their own `t.TempDir()` (`createStubCopy`). Nothing is shared, nothing
  is left behind, and a crashed earlier run can't leak into the next one.
- After `go test ./...`, whether it passed or failed, `logic/.tmp/` doesn't exist (only the old `TestNewProj` still uses it) and
  `logic/tests-assets/` is unchanged. `logic/.tmp/` is gitignored, so check it directly: `git status` can't see it.
- Tests never write into `logic/tests-assets/` itself.
- Cleanup still runs when a test fails through `t.Fatal`, or panics as the older tests do.

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
  src/*.c               # compiled into OBJECT library named after <root> (see Library naming)
  src/CMakeLists.txt    # regenerated every run
  tests/*.c             # one test executable per file
  tests/CMakeLists.txt  # regenerated every run
```

## Language detection
Introduce a language type (`utils.Language` with `LangCXX`, `LangC` and `LangUndecided`) and a single entry point that resolves it.

**Mixed sources always fail.** If `src/` holds both `.c` and `.cppm` files → `ErrMixedSources`. This holds for new and existing
projects, with or without `--lang`. There is no way around it.

**Existing project (root CMakeLists present):** `project()` decides the language.
- Support both forms: `project(name C)` and `project(name VERSION 1.0 LANGUAGES C)`. Case-insensitive.
- Contains `CXX` → CXX profile (even if `C` is also present).
- Only `C` → C profile.
- No languages listed → CXX profile (backwards compatible with today).
- `src/` is still scanned, and it must not contradict `project()`:
  - mixed → `ErrMixedSources`
  - sources of the other language, e.g. `project(vm C)` with only `.cppm` → `ErrLangMismatch`, with or without a flag
  - undecided (empty, missing, headers only) → fine. For example, `project(vm C)` with no `src/` is a C project.

**New project (no root CMakeLists):** `DetectLanguageFromSources(src/)`.
- Only `.c` files → C.
- Only `.cppm` files → CXX.
- Both → `ErrMixedSources`.
- Neither (empty or missing `src/`, or headers only) → `LangUndecided`. Detection never applies the C++ default itself.

**Explicit `--lang` flag:** `gake --lang=c` or `gake --lang=c++`.
- Accepted values are exactly `c` and `c++`. Anything else, including an empty `--lang=`, → error `ErrUnknownLang`.
- The flag never overrides what the project already says. It can only confirm it, or decide when nothing else does:
  - Existing project → the flag must match `project()`, otherwise `ErrLangMismatch`.
  - New project where `src/` decided → the flag must match, otherwise `ErrLangMismatch`.
  - New project where `src/` is undecided → the flag decides. This is how you start a fresh C project: `gake --projname=x --lang=c`.
- **The C++ default applies only when there is no flag** and nothing else decided (a new project with an undecided `src/`).
- Every error happens before any file or directory is created.

**Argument parsing:** today `ProjnameArgv` calls `flag.Parse` on the global flag set, and only on the new-project path.
Replace it with one parser that has no global state, `utils.ParseArgs(args []string) (Args, error)`. It returns
`Args{ProjName, Lang string; IsLangSet bool}` as raw values and doesn't validate them. `IsLangSet` tells `--lang=` (given,
empty) apart from no flag at all. It is called once, at the entry point, for both paths.
`--projname` is still required for new projects only, and that check stays in the logic layer. It fails with
`utils.ErrMissingProjName`, which replaces today's `noProjNameErr`, before anything is written.

## Library naming (C only)
- The library is named after the directory that contains `src/`, which is the project root: `filepath.Base(root)`.
  For example, `/home/me/calc/src` → `add_library(calc OBJECT)`.
- The tests link against that name.
- The name may contain only letters, digits, `_` and `-` (stricter than CMake, which also allows `.` and `+`). Anything else (spaces, `$`, …) → error
  `ErrInvalidLibName`, naming the directory. Don't sanitize.
- C++ keeps `Objects` unchanged, so its output stays byte-identical.
- Executables are out of scope and will be named in a later feature.

## Language profile
Replace the hard-coded extensions/templates with a per-language profile, for example:

| field            | CXX                        | C                        |
|------------------|----------------------------|--------------------------|
| src extension    | `.cppm`                    | `.c`                     |
| test extension   | `.cpp`                     | `.c`                     |
| root template    | `CMakeLists.txt.tmpl`      | `CCMakeLists.txt.tmpl`   |
| src template     | `SrcCMakeLists.txt.tmpl`   | `CSrcCMakeLists.txt.tmpl`|
| tests template   | `TestsCMakeLists.txt.tmpl` | same (library name as a parameter) |
| library name     | `Objects`                  | root dir name, validated |
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

add_library({{.LibName}} OBJECT)

target_sources({{.LibName}}
    PRIVATE
{{- range .Sources }}
    {{ . }}
{{- end }}
)

target_include_directories({{.LibName}}
    PUBLIC
    ${PROJECT_SOURCE_DIR}/include
)
```
tests: reuse `TestsCMakeLists.txt.tmpl`, but replace the hard-coded `PRIVATE Objects` with `PRIVATE {{.LibName}}`. C++ passes
`Objects`, so its output is unchanged. `tests/*.c` are collected, and test names use the same `toCamelCase`.

## Behaviour changes
- The entry point becomes `gake(root string, args utils.Args) error`. `RunAtUserRoot` parses `os.Args[1:]` with `ParseArgs`
  and passes the result in. The new tests call `gake(...)` only, so the internal helpers' signatures are free to change,
  as long as the existing `TestNewProj` and `TestExistingProk` still compile and pass.
- `initNewProj`: detect the language, then create `include/` for C and render the profile's root, src and tests templates.
- `updateExistingProj`: detect the language from the root CMakeLists, then regenerate src and tests with the profile. Create `include/` if it is missing (C only).
- C++ never creates `include/`, for new or existing projects.
- `CollectTests` takes the extension from the profile instead of hard-coding `.cpp`.
- C only: work out and validate the library name before writing any file, so an invalid name leaves the project untouched.
- Language detection and library-name errors propagate as errors (no `panic`).
- All validation (language, lib name, projname) happens before any file or directory is created, including `src/` and `tests/`.

## Tests
All new logic tests are **end-to-end** through `gake(root, utils.Args)` on a per-test copy of a stub. Unit tests are
only for pure functions. Error cases check that the project directory is **unchanged** (every file *and* directory,
via `snapshotDir`/`assertDirUnchanged`). Success cases compare the generated files exactly.

Stubs under `logic/tests-assets/project-stubs/` (source files are empty):
- `new-c/`: `src/{core,utils}.c`, `tests/{core,utils}.c`. No root CMakeLists and no `include/` (so its creation is tested).
- `existing-c/`: the same, plus a root CMakeLists with `project(vm C)` and user content after the core block.
- `existing-c-no-src/`: root `project(vm C)` and `tests/core.c`, no `src/`. Proves `project()` decides, not the sources.
- `new-mixed/`: `src/a.c` + `src/b.cppm`.
- `existing-mixed/`: root `project(vm C)` + `src/a.c` + `src/b.cppm`.
- `existing-c-cppm/`: root `project(vm C)` + `src/a.cppm`. Sources contradict `project()`.
- `existing-cxx-c/`: root `project(vm CXX)` + `src/a.c`. Sources contradict `project()`.
- `new-no-src/`: only `tests/core.c`. Git can't track an empty `src/`, so this stands in for a fresh project.
- `new/`, `existing/` (pre-existing C++ stubs). `existing` already holds empty `src/` and `tests/` CMakeLists, so tests must
  check content, not existence.

`utils` (unit):
- `TestGetLanguageFromCmake`: `project(x C)`, `project(x CXX)`, `project(x C CXX)`, `project(x VERSION 1.2 LANGUAGES C)`,
  `project(x)`, `project(x VERSION 1.2)`, lowercase `project(x c)`, `PROJECT(x C)`, multi-line, full root file.
- `TestDetectLanguageFromSources` (+ `MissingDir`, `Mixed`): `.c` → C, `.cppm` → CXX, empty/headers-only/missing → `LangUndecided`,
  mixed → `ErrMixedSources`.
- `TestGetLibName`: valid `calc`, `new-c`, `under_score`, `Lib2`; invalid `lib.v2`, `a+b`, `my lib`, `bad$name`, `dir(1)`.
  The root is passed as a full path, and only its last element is used.
- `TestParseArgs` (+ `UnknownFlag`): `--projname` and `--lang` are parsed into `Args`. Both are optional. `--lang=` sets
  `IsLangSet` with an empty value. An unknown flag is an error.
- `TestGetLanguageFromFlag` (+ `Unknown`): `c` → C, `c++` → CXX; `C`, `cxx`, `rust` and `""` → `ErrUnknownLang`.

`logic` (end-to-end):
- `TestNewCProj`, `TestExistingCProj`: exact C output (root, src, tests), `include/` created, and the existing root is byte-identical.
- `TestNewMixedProjFails`, `TestNewCProjInvalidLibName`, `TestExistingCProjInvalidLibName`, `TestNewProjMissingProjName`:
  the right error, and the project is unchanged.
- `TestNewCXXProjIgnoresLibNameRule`: a C++ project in `my lib` succeeds with `Objects` and creates no `include/`.
- `TestGakeLangRules`: success combinations of flag × project. Exact src output. `include/` exists only for C.
- `TestGakeLangRulesErrors`: every rejected combination (a contradicting flag, sources contradicting `project()`, mixed
  sources with any flag, new and existing projects, an unknown value) → the right error, project unchanged.
- `TestGakeLangFlag`: `new-no-src` + `--lang=c` → exact C output with an empty source list.
- `TestGakeEmptyLangValueFails`: `--lang=` (set, empty) → `ErrUnknownLang`, project unchanged.
- `TestExistingProk` (rewritten): runs through `gake` on its own copy of `existing`. Exact C++ src and tests output, root unchanged, no `include/`.
- `TestNewProj` (pre-existing) stays unchanged.

Added after `/code-review` (bug fixes, red-green):
- `TestGetLanguageFromCmakeIgnoresNonLanguageText` (utils): comments, parentheses inside quoted `DESCRIPTION`, and the
  project name / `DESCRIPTION` / `HOMEPAGE_URL` values never count as languages.
- `TestRunGakeHelp` (logic, end-to-end through `runGake(root, argv, out)`): `-h`/`--help` prints usage listing both
  flags, succeeds, and leaves the project unchanged.

## Out of scope
- Mixed C/C++ projects (C sources plus C++ modules in one target).
- Generating an executable for `main.c`/`main.cpp`.
- Preserving user content in the root CMakeLists (`parseRootCmake`/`validateExistingRootCmake` stay unused).
- Naming executables. That is a later feature.
- Name clashes between the library and a test target (for example, root `Core` with `tests/core.c` → test `Core`).

## Known issues
- An empty source list renders `add_library(<name> OBJECT)` with no sources, which CMake rejects at generate time ("No SOURCES given to target"). C++ has the
  same problem today. Accept it for now, since the user is expected to add sources next.
- `parseRootCmake` passes a file path to `getProjNameFromCmake`, which expects a directory. Don't copy this pattern when reusing the parser.

## Done when
- Every approved Phase 1 test passes, and so does every test that existed before: `go build ./... && go test ./...`.
- Manual smoke check: running `gake` in a scratch dir that contains only `src/*.c`, `include/*.h` and `tests/*.c` produces
  a project that builds with `cmake -B build && cmake --build build && ctest --test-dir build`.
