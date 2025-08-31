# ZeroUI Raycast Extension

A beautiful Raycast extension that brings ZeroUI's delightful configuration management capabilities directly into Raycast's interface.

## ✨ Features

- **🎨 Beautiful Interface**: Modern Raycast UI with smooth interactions and animations
- **📋 Application Management**: View and manage all configured applications with smart filtering
- **⚙️ Configuration Control**: Toggle settings with visual feedback and quick actions
- **🔍 Advanced Search**: Powerful search and filtering across all configuration data
- **⌨️ Enhanced Keymap Management**: Browse, search, and copy keyboard shortcuts
- **🎯 Preset Application**: Apply configuration presets with one-click deployment
- **🚀 Performance Optimized**: Intelligent caching and lazy loading for fast responses
- **🔄 Real-time Updates**: Instant feedback with comprehensive error handling
- **⚡ Menu Bar Integration**: Quick access and status monitoring from menu bar
- **🛠️ Cache Management**: View and control caching behavior for optimal performance
- **⌨️ Keyboard Shortcuts**: Extensive keyboard navigation throughout the interface
- **⚙️ Customizable Settings**: Configurable timeouts, paths, and caching preferences

## 🚀 Installation

### Prerequisites
- **Raycast**: Latest version installed
- **ZeroUI**: Built and available at `/Users/m/code/muka-hq/zeroui/build/zeroui`
- **Node.js**: For development and building

### Install Extension
1. Clone or download this extension
2. Open Raycast
3. Go to Extensions → Import Extension
4. Select the `raycast-extension` folder
5. The extension will be available as "ZeroUI"

## 🔧 Troubleshooting

### No Applications Listed

If Raycast shows "No applications found" or an error, follow these steps:

1. **Validate ZeroUI setup:**
   ```bash
   cd raycast-extension
   node validate.js
   ```

2. **Check ZeroUI binary location:**
   The extension looks for the ZeroUI binary in these locations (in order):
   - Extension directory: `./zeroui`
   - Build directory: `../build/zeroui`
   - Project root: `../zeroui`
   - System paths: `/usr/local/bin/zeroui`, `/opt/homebrew/bin/zeroui`

3. **Update Raycast preferences:**
   - Open Raycast → Preferences → Extensions → ZeroUI
   - Set "ZeroUI Binary Path" to the full path of your ZeroUI binary
   - Example: `/Users/username/code/zeroui/raycast-extension/zeroui`

4. **Enable debug logging:**
   - Open Raycast's Developer Tools (⌥⌘I)
   - Check the Console tab for error messages
   - Look for path resolution logs from the extension

### Common Issues

- **Binary not executable:** Run `chmod +x zeroui`
- **Path not found:** Copy binary to extension directory:
  ```bash
  cp ../build/zeroui ./zeroui
  chmod +x zeroui
  ```
- **Permission denied:** Check file permissions with `ls -la zeroui`
- **Cache issues:** Clear cache in Raycast extension preferences
- **Build issues:** Rebuild the extension:
  ```bash
  npm run build
  ```

### Debug Commands

```bash
# Test ZeroUI directly
./zeroui list apps

# Check binary permissions
ls -la zeroui

# Test with full path
/Users/username/code/zeroui/raycast-extension/zeroui list apps
```

## 📱 Commands

### List Applications
- **Command**: `zeroui list-apps`
- **Description**: View all available applications
- **Features**:
  - Search and filter applications
  - Quick access to configuration views
  - Direct links to keymaps and changed values

### Toggle Configuration
- **Command**: `zeroui toggle-config <app> <key> <value>`
- **Description**: Toggle configuration values
- **Features**:
  - Interactive form for value input
  - Visual feedback during toggle
  - Error handling and validation

### List Configuration Values
- **Command**: `zeroui list-values <app>`
- **Description**: View current configuration values
- **Features**:
  - Search through all settings
  - Copy values to clipboard
  - Quick navigation between apps

### List Changed Values
- **Command**: `zeroui list-changed <app>`
- **Description**: View values different from defaults
- **Features**:
  - See what you've customized
  - Compare current vs default values
  - Copy values for reference

### List Keymaps
- **Command**: `zeroui keymap-list <app>`
- **Description**: View keyboard shortcuts
- **Features**:
  - Browse all keybindings
  - Copy shortcuts to clipboard
  - Search through keymaps

### Manage Presets
- **Command**: `zeroui manage-presets`
- **Description**: Apply configuration presets
- **Features**:
  - Multiple preset options (minimal, default, developer)
  - One-click application
  - Automatic backup before changes

### Cache Management
- **Command**: `zeroui cache-management`
- **Description**: View and manage extension cache for optimal performance
- **Features**:
  - View cache statistics and performance metrics
  - Clear cache to free memory
  - Toggle caching on/off
  - Monitor cache hits/misses

## 🎯 Usage Examples

### Quick Application Overview
```
zeroui list-apps
```
Shows all configured applications with quick access to their configurations.

### Toggle a Setting
```
zeroui toggle-config ghostty focus-follows-mouse true
```
Toggle the `focus-follows-mouse` setting for Ghostty.

### View Changed Settings
```
zeroui list-changed ghostty
```
See all settings you've customized from their defaults.

### Browse Keymaps
```
zeroui keymap-list ghostty
```
View all keyboard shortcuts configured for Ghostty.

## 🛠️ Development

### Setup
```bash
cd raycast-extension
npm install
```

### Development
```bash
npm run dev
```
This starts the development server and hot-reloads changes.

### Building
```bash
npm run build
```

### Linting
```bash
npm run lint
```

## ⚙️ Configuration

### ZeroUI Path
The extension expects ZeroUI to be available at:
```
/Users/m/code/muka-hq/zeroui/build/zeroui
```

To change this path, modify the `zerouiPath` in `src/utils.ts`.

### Extension Preferences
The extension can be configured through Raycast's extension preferences:

- **ZeroUI Binary Path**: Full path to the ZeroUI binary (leave empty for default)
- **Command Timeout**: Timeout for ZeroUI commands in milliseconds (default: 30000ms)
- **Enable Caching**: Toggle caching for improved performance (default: enabled)
- **Cache Duration**: How long to cache results in milliseconds (default: 300000ms = 5 minutes)

### Performance Tuning
- **Caching**: Enabled by default, significantly improves performance by caching ZeroUI responses
- **Timeout**: Adjust based on your system's performance and network conditions
- **Cache Duration**: Longer durations improve performance but may show stale data

### Customization
You can customize the extension by:
- Modifying the UI components in `src/`
- Adding new commands in the `package.json`
- Updating the styling and icons
- Extending the ZeroUI integration in `utils.ts`
- Adjusting performance settings through preferences

## 🎨 UI Features

### Modern Design
- Clean, modern interface matching Raycast's design language
- Consistent typography and spacing
- Smooth animations and transitions

### Smart Interactions
- **Keyboard Shortcuts**: Extensive shortcuts for power users (Cmd+C, Cmd+V, Cmd+R, etc.)
- **Advanced Search**: Real-time filtering and search across all configuration data
- **Context-aware Actions**: Smart suggestions based on current context
- **Quick Navigation**: Fast switching between views and actions

### Error Handling
- Graceful error handling with user-friendly messages
- Automatic retry mechanisms for failed operations
- Clear feedback for all user actions

## 🔧 Technical Details

### Architecture
- **React-based**: Built with React and TypeScript
- **Raycast API**: Uses Raycast's extension API
- **Async Operations**: Handles all ZeroUI commands asynchronously
- **Error Boundaries**: Comprehensive error handling

### Integration
- **Shell Commands**: Executes ZeroUI CLI commands
- **Output Parsing**: Intelligently parses ZeroUI's output
- **State Management**: React state for UI updates
- **Clipboard Integration**: Copy values and keymaps

### Performance
- **Intelligent Caching**: Smart caching system with configurable TTL and size limits
- **Lazy Loading**: Components load data on demand with loading states
- **Efficient Rendering**: Optimized React components with proper memoization
- **Background Processing**: Non-blocking command execution with progress feedback
- **Memory Management**: Automatic cache cleanup and resource management
- **Performance Monitoring**: Cache hit/miss statistics and response time tracking

## 📚 API Reference

### ZeroUI Class
```typescript
const zeroui = new ZeroUI(path?: string);
```

#### Methods
- `listApps()`: Get all available applications
- `listValues(app)`: Get configuration values for an app
- `listChanged(app)`: Get changed values for an app
- `toggleConfig(app, key, value)`: Toggle a configuration value
- `listKeymaps(app)`: Get keymaps for an app
- `executeCommand(command, args)`: Execute arbitrary ZeroUI commands

## 🤝 Contributing

1. **Fork** the repository
2. **Create** a feature branch
3. **Make** your changes
4. **Test** thoroughly
5. **Submit** a pull request

### Development Guidelines
- Follow TypeScript best practices
- Use Raycast's design system
- Handle errors gracefully
- Write clear, concise code
- Test all new features

## 📄 License

This extension is part of the ZeroUI project and follows the same MIT license.

## 🙏 Acknowledgments

- **ZeroUI Team**: For the amazing CLI tool
- **Raycast Team**: For the incredible productivity platform
- **Charm Libraries**: For the beautiful TUI components

---

**Made with ❤️ for productivity enthusiasts**

Transform your configuration management with ZeroUI's delightful interface, now accessible through Raycast's beautiful UI! 🚀✨
