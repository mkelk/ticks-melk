# Personal Notes

## Building tk locally

```bash
# Build and install to Go bin (recommended)
go install ./cmd/tk

# Rebuild after changes - same command
go install ./cmd/tk

# Verify
tk version
```

## Build with version info

```bash
go build -ldflags "-s -w -X main.Version=dev" -o tk.exe ./cmd/tk
```

## Run tests

```bash
go test ./...
```
