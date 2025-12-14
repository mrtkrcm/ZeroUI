# Quick start

This guide gets you from zero to a working `zeroui` binary and a first configuration change.

## Install

### Using Go

Requires repository access:

```bash
go install github.com/mrtkrcm/ZeroUI@latest
```

### From source

```bash
make build
./build/zeroui --help
```

### Docker (local)

```bash
make docker-build
make docker-run
```

## First run

```bash
zeroui                 # Launch interactive app grid (terminal required)
zeroui list apps
```

## Make a change

```bash
zeroui toggle ghostty theme dark
zeroui preset ghostty minimal --show-diff
```

## Next

- Command reference: `commands.md`
- Reference system: `reference-system.md`
