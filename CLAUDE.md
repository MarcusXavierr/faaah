# faaah — Project Instructions

## Go Version

This project requires **Go 1.24**. Before running any `go` commands, verify the active version:

```bash
go version  # should show go1.24.x
```

If it's not Go 1.24, switch using `gvm`:

```bash
source ~/.gvm/scripts/gvm
gvm use go1.24
```

All `go build`, `go test`, `go get`, and `make` commands must be run under Go 1.24.
