package registry

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AppDefinition represents a supported application
type AppDefinition struct {
	Name       string
	Executable string
	Logo       string
	ConfigPath string
	Category   string
}

// AppStatus represents the current state of an application
type AppStatus struct {
	Definition     AppDefinition
	IsInstalled    bool
	HasConfig      bool
	ConfigExists   bool
	ExecutablePath string
}

// GetSupportedApps returns all supported application definitions
func GetSupportedApps() []AppDefinition {
	return []AppDefinition{
		{
			Name:       "Ghostty",
			Executable: "ghostty",
			Logo:       "👻",
			ConfigPath: "~/.config/ghostty/config",
			Category:   "Terminal",
		},
		{
			Name:       "Alacritty",
			Executable: "alacritty",
			Logo:       "🖥️",
			ConfigPath: "~/.config/alacritty/alacritty.yml",
			Category:   "Terminal",
		},
		{
			Name:       "WezTerm",
			Executable: "wezterm",
			Logo:       "🪟",
			ConfigPath: "~/.config/wezterm/wezterm.lua",
			Category:   "Terminal",
		},
		{
			Name:       "VS Code",
			Executable: "code",
			Logo:       "📝",
			ConfigPath: "~/.config/Code/User/settings.json",
			Category:   "Editor",
		},
		{
			Name:       "Neovim",
			Executable: "nvim",
			Logo:       "📜",
			ConfigPath: "~/.config/nvim/init.lua",
			Category:   "Editor",
		},
		{
			Name:       "Zed",
			Executable: "zed",
			Logo:       "⚡",
			ConfigPath: "~/.config/zed/settings.json",
			Category:   "Editor",
		},
		{
			Name:       "Tmux",
			Executable: "tmux",
			Logo:       "🔲",
			ConfigPath: "~/.tmux.conf",
			Category:   "Multiplexer",
		},
		{
			Name:       "Git",
			Executable: "git",
			Logo:       "🌳",
			ConfigPath: "~/.gitconfig",
			Category:   "Development",
		},
		{
			Name:       "Starship",
			Executable: "starship",
			Logo:       "🚀",
			ConfigPath: "~/.config/starship.toml",
			Category:   "Shell",
		},
	}
}

// CheckExecutable checks if an executable exists in PATH
func CheckExecutable(name string) (string, bool) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", false
	}
	return path, true
}

// ExpandPath expands ~ to home directory
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home := getHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

// getHomeDir returns the user's home directory
func getHomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	// Fallback for Windows
	if home := os.Getenv("USERPROFILE"); home != "" {
		return home
	}
	return ""
}

// GetAppStatuses returns the status of all supported applications
func GetAppStatuses() []AppStatus {
	apps := GetSupportedApps()
	statuses := make([]AppStatus, 0, len(apps))
	
	for _, app := range apps {
		execPath, isInstalled := CheckExecutable(app.Executable)
		configPath := ExpandPath(app.ConfigPath)
		
		status := AppStatus{
			Definition:     app,
			IsInstalled:    isInstalled,
			ExecutablePath: execPath,
			ConfigExists:   fileExists(configPath),
			HasConfig:      hasZeroUIConfig(app.Name),
		}
		
		statuses = append(statuses, status)
	}
	
	return statuses
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// hasZeroUIConfig checks if we have a ZeroUI config for this app
func hasZeroUIConfig(appName string) bool {
	// Check if app config exists in ~/.config/zeroui/apps/
	configPath := filepath.Join(getHomeDir(), ".config", "zeroui", "apps", strings.ToLower(appName)+".yaml")
	return fileExists(configPath)
}