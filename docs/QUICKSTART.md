# Quick Start — ZeroUI

ZeroUI is a zero-configuration UI toolkit manager. This guide covers installation and basic usage.

Installation

Using Go (recommended):

```
go install github.com/mrtkrcm/ZeroUI@latest
```

Using Docker:

```
docker run --rm -it zeroui/zeroui:latest
```

Basic Usage

List available applications:

```
zeroui list apps
```

Toggle a configuration value:

```
zeroui toggle ghostty theme dark
```

Launch interactive TUI:

```
zeroui ui
```

For the full command reference, see docs/COMMANDS.md
