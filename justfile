# jig justfile

BINNAME := "jig"
BINDIR  := "bin/release"

VERSION    := `git describe --always --tags`
GIT_COMMIT := `git rev-parse HEAD`
GIT_SHA    := `git rev-parse --short HEAD`
GIT_TAG    := `git describe --tags --abbrev=0 --exact-match 2>/dev/null || true`
GIT_DIRTY  := `test -n "$(git status --porcelain)" && echo "dirty" || echo "clean"`

BINARY_VERSION := if VERSION == '' { GIT_TAG } else { VERSION }
VERSION_METADATA := if GIT_TAG == '' { 'unreleased' } else { '' }

GOMODULE := `go list -m`
LDFLAGS := (
  "-s -w"
  + " -X " + GOMODULE + "/internal/version.version=" + BINARY_VERSION
  + " -X " + GOMODULE + "/internal/version.metadata=" + VERSION_METADATA
  + " -X " + GOMODULE + "/internal/version.gitCommit=" + GIT_COMMIT
  + " -X " + GOMODULE + "/internal/version.gitTreeState=" + GIT_DIRTY
)

GOLANGCI_VERSION := "latest"
GOIMPORTS_VERSION := "latest"

[private]
default:
  @just --list

# build binary
build $GOOS='' $GOARCH='':
  go build \
    -ldflags='{{ LDFLAGS }}' \
    -o {{ BINDIR / BINNAME }}{{ if GOOS != '' {'-'+GOOS} else {''} }}{{ if GOARCH != '' {'-'+GOARCH} else {''} }} \
    ./cmd/jig

# run tests
test:
  go test -v ./...

# run golangci-lint checks
lint *flags: _golangci
  golangci-lint run {{ flags }}

# format with goimports
format: _goimports
  GO111MODULE=on go list -f '{{{{.Dir}}' ./... | xargs goimports -w -local '{{ GOMODULE }}'

# bump with cocogitto
bump version:
  cog bump --version {{ version }}

# build all binaries
release: _release-linux _release-darwin
_release-linux:  (build "linux"  "amd64") (build "linux"  "arm64")
_release-darwin: (build "darwin" "amd64") (build "darwin" "arm64")

# video tape a demo
vhs:
  vhs demo.tape

# TOOLS
# ---

_golangci: (_fetch "golangci-lint" GOLANGCI_VERSION "github.com/golangci/golangci-lint/cmd/golangci-lint")

_goimports: (_fetch "goimports" GOIMPORTS_VERSION "golang.org/x/tools/cmd/goimports")

_fetch bin version url:
  #!/usr/bin/env bash
  if hash {{ bin }} 2>/dev/null; then
    test '{{ version }}' = 'latest' && exit
    test {{ bin }} --version | grep -qF '{{ version }} ' && exit
  fi
  cd / && GO111MODULE=on go install '{{ url }}@{{ version }}'
