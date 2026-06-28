# deps-analyzer

Clones a remote Go repository and reports which dependencies have newer versions available, using the Go module proxy.

## Requirements

- Go 1.21+
- `git` in your PATH environment variable

## Build

```sh
go build -o deps-analyzer ./cmd
```

The compiled binary is placed in the current directory.

## Usage

```
deps-analyzer [flags] <repository-url>
```

**Flags**

| Flag | Default | Description |
|---|---|---|
| `--direct-only` | `false` | Skip indirect dependencies |
| `--timeout` | `5m` | Overall operation timeout |
| `-f`, `--format` | `text` | Output format: `text` or `json` |
| `--proxy` | `https://proxy.golang.org` | Go module proxy URL |

## Examples

Check all dependencies in a repository:

```sh
deps-analyzer https://github.com/example/repo
```

Check only direct dependencies, with JSON output:

```sh
deps-analyzer --direct-only --format json https://github.com/example/repo
```

Use a shorter timeout:

```sh
deps-analyzer --timeout 2m https://github.com/example/repo
```

## Output

**Text format (default)**

```
Repository: https://github.com/example/repo

Module: github.com/example/repo (root)
  Go version:   1.21
  Dependencies: 12 (8 direct, 4 indirect)
  Updates:      2 available
    github.com/some/dep                                     v1.2.0 -> v1.3.0
    github.com/other/dep                                    v0.9.1 -> v1.0.0 (indirect)

Summary: 1 module(s) scanned, 2 update(s) available
```

**JSON format**

```json
{
  "repository": "https://github.com/example/repo",
  "modules": [
    {
      "path": "github.com/example/repo",
      "location": "root",
      "go_version": "1.21",
      "total_dependencies": 12,
      "direct_dependencies": 8,
      "indirect_dependencies": 4,
      "updates": [
        {
          "path": "github.com/some/dep",
          "current": "v1.2.0",
          "latest": "v1.3.0",
          "indirect": false
        }
      ]
    }
  ],
  "modules_scanned": 1,
  "total_updates": 1
}
```

## Test

```sh
go test ./...
```
