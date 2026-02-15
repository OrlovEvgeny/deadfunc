# deadfunc

[![CI](https://github.com/OrlovEvgeny/deadfunc/actions/workflows/ci.yml/badge.svg)](https://github.com/OrlovEvgeny/deadfunc/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/OrlovEvgeny/deadfunc)](https://goreportcard.com/report/github.com/OrlovEvgeny/deadfunc)
[![Go Reference](https://pkg.go.dev/badge/github.com/OrlovEvgeny/deadfunc.svg)](https://pkg.go.dev/github.com/OrlovEvgeny/deadfunc)

A Go linter that detects **functions declared but never called or referenced in production code**. Test files (`_test.go`) are excluded from both declaration scanning and usage tracking — so a function used *only* in tests is flagged as dead code.

## What it catches

| Scenario | Flagged? |
|---|---|
| Function declared, never called anywhere | Yes |
| Function declared, called only in `_test.go` | Yes |
| Function declared, called in production code | No |
| Function passed as a value (callback) | No |
| `init()` / `main()` | No (entry points) |
| Methods implementing known interfaces (`error`, `fmt.Stringer`, `json.Marshaler`, `io.Reader`, etc.) | No |
| Methods implementing package-local interfaces | No |
| Exported functions (default mode) | No (use `-exported` flag) |
| Code in generated files (`// Code generated`) | No |

## Install

### Standalone CLI

```bash
go install github.com/OrlovEvgeny/deadfunc/cmd/deadfunc@latest
```

### golangci-lint module plugin (recommended)

Add to your `.custom-gcl.yml`:

```yaml
version: v1.64.0
plugins:
  - module: "github.com/OrlovEvgeny/deadfunc"
    import: "github.com/OrlovEvgeny/deadfunc/plugin"
    version: latest
```

Build your custom binary:

```bash
custom-gcl
```

Enable in `.golangci.yml`:

```yaml
linters:
  enable:
    - deadfunc
```

### Legacy .so plugin

```bash
go build -buildmode=plugin -o deadfunc.so ./plugin
```

Add to `.golangci.yml`:

```yaml
linters-settings:
  custom:
    deadfunc:
      path: ./deadfunc.so
      description: "Reports unused production functions"
      original-url: github.com/OrlovEvgeny/deadfunc
```

## Usage

```bash
# Check unexported functions only (safe default)
deadfunc ./...

# Also check exported functions (use for binaries / main packages)
deadfunc -exported ./...

# Skip method checking
deadfunc -methods=false ./...

# Include generated files
deadfunc -skip-generated=false ./...
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `-exported` | `false` | Also check exported functions. Risky for library packages since they may be used by external consumers. |
| `-methods` | `true` | Check methods, not just top-level functions. |
| `-skip-generated` | `true` | Skip files with `// Code generated ... DO NOT EDIT` header. |

## How it works

1. **Collect declarations** — walks all `*ast.FuncDecl` nodes in non-test files, building a set of `types.Object` identifiers.
2. **Collect usage** — walks all `*ast.Ident` and `*ast.SelectorExpr` nodes in non-test files, resolving them to `types.Object` via `TypesInfo.Uses` and `TypesInfo.Selections`.
3. **Diff** — any declared function whose `types.Object` was never seen in the usage pass is reported.

Using `types.Object` identity (not string names) means it correctly handles shadowing, different packages, and method sets.

## Limitations

- **Cross-package analysis**: the linter works per-package. A function in package `A` used only by package `B` will NOT be flagged (the Go type checker resolves cross-package references). However, `go vet`-style tools run per-package, so this is the expected behavior.
- **Reflection**: functions called only via `reflect` will be falsely flagged.
- **`go:linkname`**: functions referenced via `//go:linkname` will be falsely flagged.
- **Interface implementations**: a heuristic is used — the linter checks package-scoped interfaces + a set of well-known stdlib interfaces. Exotic interfaces from dependencies may be missed.

## Nolint

Use the standard `//nolint:deadfunc` directive (if using golangci-lint) or add a comment:

```go
//nolint:deadfunc // used via reflection
func myReflectionHandler() {}
```

## License

[MIT](LICENSE)
