# Application Scanning and Status

ZeroUI now includes automatic application scanning on launch, providing instant visibility into which applications are configured and ready to use.

## Features

### 🔍 Automatic Discovery
- Scans for 15+ popular development tools on startup
- Checks multiple common configuration paths per application
- Shows real-time progress during scanning

### 📊 Status Indicators
- **○ Ready** - Application configuration found and accessible
- **○ Not Configured** - Application not yet configured
- **○ Error** - Issues accessing configuration

### 📁 Registry System
ZeroUI uses a YAML-based registry to define known applications:

- **Built-in Registry**: Includes popular terminals, editors, shells, and tools
- **Custom Apps**: Add your own via `~/.config/zeroui/apps.yaml`
- **Override Support**: Customize existing app definitions

## Built-in Applications

### Terminal Emulators
- ZeroUI (○)
- Ghostty (◉)
- Alacritty (○)
- Kitty (○)
- WezTerm (○)
- tmux (○)

### Code Editors
- Zed (○)
- Neovim (○)
- Visual Studio Code (○)
- Sublime Text (○)

### Shells & Prompts
- Zsh (○)
- Bash (○)
- Fish (○)
- Starship (○)

### Developer Tools
- Git (○)
- LazyGit (○)

## Custom Applications

Create `~/.config/zeroui/apps.yaml` to add your own applications:

```yaml
applications:
  - name: my-app
    display_name: My Custom App
    icon: "●"
    description: "My custom application"
    category: custom
    config_paths:
      - "~/.config/my-app/config.yaml"
      - "~/.my-app.conf"
    config_format: yaml
```

## Configuration Paths

ZeroUI automatically expands `~` to your home directory and checks multiple standard locations:

- `~/.config/<app>/` - XDG config directory (Linux/macOS)
- `~/.<app>rc` - Traditional dotfiles
- `~/Library/Application Support/` - macOS application support

## Advanced Features

### Full Registry Override
Power users can completely replace the built-in registry by creating:
`~/.config/zeroui/apps_registry.yaml`

### Programmatic Access
Applications can query the registry programmatically:

```go
registry, _ := config.LoadAppsRegistry()
app, exists := registry.GetApp("ghostty")
configPath, found := registry.FindConfigPath("ghostty")
```

## Performance

- Scanning is asynchronous and non-blocking
- Typically completes in under 2 seconds for all applications
- Results are cached for instant access
- Minimal resource usage with efficient file checks

## UI Integration

The scanning progress integrates seamlessly with ZeroUI's TUI:

1. **Launch**: `zeroui` or `zeroui ui`
2. **Scanning**: Shows progress bar and current app being checked
3. **Results**: Categorized view of all applications with status
4. **Navigation**: Select any configured app to edit its settings

## Troubleshooting

### App Not Detected
1. Check if config file exists at expected path
2. Add custom path via `~/.config/zeroui/apps.yaml`
3. Verify file permissions allow reading

### Slow Scanning
- Reduce number of apps in custom registry
- Check for network-mounted home directories
- Ensure filesystem is responsive

### Custom App Not Working
- Validate YAML syntax in `apps.yaml`
- Check file paths are correct
- Ensure category exists if specified