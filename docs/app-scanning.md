# App scanning

ZeroUI scans for supported applications and their config files to show what is available and what is already configured.

## Status meanings

- **Ready**: a config file was found and is readable.
- **Not configured**: no config file was found in known locations.
- **Error**: a config path exists but could not be read or parsed.

## Custom apps

Add custom applications in `~/.config/zeroui/apps.yaml`:

```yaml
applications:
  - name: my-app
    display_name: My App
    description: Custom application
    category: custom
    config_paths:
      - "~/.config/my-app/config.yaml"
    config_format: yaml
```

## Registry override

Power users can override the built-in registry by creating:

- `~/.config/zeroui/apps_registry.yaml`
