#!/usr/bin/env bash
# SessionStart check: reports which tools this repo needs are missing, with the command to install each.
# Never fails the session; its output is shown to Claude as context.

function getInstallHint() {
	case "$1" in
		go)          echo "install Go >= $(getRequiredGoVersion): https://go.dev/dl/" ;;
		gopls)       echo "go install golang.org/x/tools/gopls@latest" ;;
		staticcheck) echo "go install honnef.co/go/tools/cmd/staticcheck@latest" ;;
		goimports)   echo "go install golang.org/x/tools/cmd/goimports@latest" ;;
		cmake|ctest) echo "install CMake >= 3.28 (system package manager)" ;;
		cc)          echo "install a C/C++ compiler with C23 and C++23 module support (gcc >= 14 or clang >= 17)" ;;
	esac
}

function getRequiredGoVersion() {
	awk '/^go /{print $2}' "$(dirname "$0")/../go.mod"
}

function isToolPresent() {
	command -v "$1" >/dev/null 2>&1 || test -x "$(go env GOPATH 2>/dev/null)/bin/$1"
}

missing=()
for tool in go gopls staticcheck goimports cmake ctest cc; do
	isToolPresent "$tool" || missing+=("$tool")
done

if [[ ${#missing[@]} -eq 0 ]]; then
	echo "gake tooling: all required tools present."
	exit 0
fi

echo "gake tooling: missing tools. Ask the user before installing:"
for tool in "${missing[@]}"; do
	echo "  - $tool: $(getInstallHint "$tool")"
done
exit 0
