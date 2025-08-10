package cmd

// TODO: Add backup path validation to prevent directory traversal (Security)
// TODO: Ensure backup paths stay within ~/.config/zeroui/backups/

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/mrtkrcm/ZeroUI/internal/errors"
	"github.com/mrtkrcm/ZeroUI/internal/recovery"
	"github.com/mrtkrcm/ZeroUI/internal/toggle"
	"github.com/spf13/cobra"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage configuration backups",
	Long: `Manage configuration backups for applications. You can list, create, restore, and cleanup backups.

Backups are automatically created before any configuration changes to ensure you can recover
from any issues. Use these commands to manually manage your backups.`,
}

// backupListCmd lists available backups
var backupListCmd = &cobra.Command{
	Use:   "list [app]",
	Short: "List available backups",
	Long: `List available configuration backups. If an app name is provided, only backups 
for that app will be shown.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := ""
		if len(args) > 0 {
			appName = args[0]
		}

		backupManager, err := recovery.NewBackupManager()
		if err != nil {
			if ctErr, ok := errors.GetZeroUIError(err); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", ctErr.String())
				return nil
			}
			return err
		}

		backups, err := backupManager.ListBackups(appName)
		if err != nil {
			if ctErr, ok := errors.GetZeroUIError(err); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", ctErr.String())
				return nil
			}
			return err
		}

		if len(backups) == 0 {
			if appName != "" {
				fmt.Printf("No backups found for app: %s\n", appName)
			} else {
				fmt.Println("No backups found")
			}
			return nil
		}

		// Display backups in a table
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
		fmt.Fprintln(w, "APP\tTIME\tSIZE\tFILE")
		fmt.Fprintln(w, "---\t----\t----\t----")

		for _, backup := range backups {
			// Extract app name from backup filename
			fileName := filepath.Base(backup.Name)
			parts := strings.SplitN(fileName, "_", 2)
			backupApp := parts[0]

			// Format size
			size := formatSize(backup.Size)

			// Format time
			timeStr := backup.Created.Format("2006-01-02 15:04:05")

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", backupApp, timeStr, size, backup.Name)
		}

		w.Flush()
		return nil
	},
}

// backupCreateCmd creates a manual backup
var backupCreateCmd = &cobra.Command{
	Use:   "create <app>",
	Short: "Create a manual backup of an app's configuration",
	Long: `Create a manual backup of an application's configuration file. 
This is useful before making major changes or for creating restore points.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		// Load app config to get the file path
		engine, err := toggle.NewEngine()
		if err != nil {
			return fmt.Errorf("failed to create toggle engine: %w", err)
		}

		appConfig, err := engine.GetAppConfig(appName)
		if err != nil {
			if ctErr, ok := errors.GetZeroUIError(err); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", ctErr.String())
				return nil
			}
			return err
		}

		// Resolve config path
		configPath := appConfig.Path
		if strings.HasPrefix(configPath, "~") {
			home, _ := os.UserHomeDir()
			configPath = strings.Replace(configPath, "~", home, 1)
		}

		// Create backup
		backupManager, err := recovery.NewBackupManager()
		if err != nil {
			if ctErr, ok := errors.GetZeroUIError(err); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", ctErr.String())
				return nil
			}
			return err
		}

		backupPath, err := backupManager.CreateBackup(configPath, appName)
		if err != nil {
			if ctErr, ok := errors.GetZeroUIError(err); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", ctErr.String())
				return nil
			}
			return err
		}

		if backupPath == "" {
			fmt.Printf("No backup created - configuration file does not exist: %s\n", configPath)
		} else {
			fmt.Printf("✓ Backup created: %s\n", filepath.Base(backupPath))
		}

		return nil
	},
}

// backupRestoreCmd restores from a backup
var backupRestoreCmd = &cobra.Command{
	Use:   "restore <app> <backup-name>",
	Short: "Restore configuration from a backup",
	Long: `Restore an application's configuration from a previously created backup. 
This will overwrite the current configuration file.

Use 'zeroui backup list <app>' to see available backups.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		backupName := args[1]

		// Load app config to get the file path
		engine, err := toggle.NewEngine()
		if err != nil {
			return fmt.Errorf("failed to create toggle engine: %w", err)
		}

		appConfig, err := engine.GetAppConfig(appName)
		if err != nil {
			if ctErr, ok := errors.GetZeroUIError(err); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", ctErr.String())
				return nil
			}
			return err
		}

		// Resolve config path
		configPath := appConfig.Path
		if strings.HasPrefix(configPath, "~") {
			home, _ := os.UserHomeDir()
			configPath = strings.Replace(configPath, "~", home, 1)
		}

		// Find backup
		backupManager, err := recovery.NewBackupManager()
		if err != nil {
			if ctErr, ok := errors.GetZeroUIError(err); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", ctErr.String())
				return nil
			}
			return err
		}

		backups, err := backupManager.ListBackups(appName)
		if err != nil {
			if ctErr, ok := errors.GetZeroUIError(err); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", ctErr.String())
				return nil
			}
			return err
		}

		var backupPath string
		for _, backup := range backups {
			if backup.Name == backupName {
				backupPath = backup.Path
				break
			}
		}

		if backupPath == "" {
			fmt.Fprintf(os.Stderr, "Error: backup '%s' not found for app '%s'\n", backupName, appName)
			fmt.Fprintf(os.Stderr, "Use 'zeroui backup list %s' to see available backups\n", appName)
			return nil
		}

		// Confirm restoration
		confirmed, _ := cmd.Flags().GetBool("yes")
		if !confirmed {
			fmt.Printf("This will overwrite the current configuration for %s.\n", appName)
			fmt.Printf("Current config: %s\n", configPath)
			fmt.Printf("Backup: %s\n", backupName)
			fmt.Print("Are you sure? (y/N): ")

			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Restore cancelled")
				return nil
			}
		}

		// Create backup of current config before restore
		currentBackup, err := backupManager.CreateBackup(configPath, appName)
		if err != nil {
			fmt.Printf("Warning: could not backup current config: %v\n", err)
		} else if currentBackup != "" {
			fmt.Printf("Current config backed up as: %s\n", filepath.Base(currentBackup))
		}

		// Restore the backup
		if err := backupManager.RestoreBackup(backupPath, configPath); err != nil {
			if ctErr, ok := errors.GetZeroUIError(err); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", ctErr.String())
				return nil
			}
			return err
		}

		fmt.Printf("✓ Configuration restored from backup: %s\n", backupName)
		return nil
	},
}

// backupCleanupCmd removes old backups
var backupCleanupCmd = &cobra.Command{
	Use:   "cleanup [app] [--keep N]",
	Short: "Clean up old backups",
	Long: `Remove old backup files, keeping only the most recent ones. 
By default, keeps the 5 most recent backups per application.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := ""
		if len(args) > 0 {
			appName = args[0]
		}

		keepCount, _ := cmd.Flags().GetInt("keep")

		backupManager, err := recovery.NewBackupManager()
		if err != nil {
			if ctErr, ok := errors.GetZeroUIError(err); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", ctErr.String())
				return nil
			}
			return err
		}

		if appName != "" {
			// Clean up specific app
			if err := backupManager.CleanupOldBackups(appName, keepCount); err != nil {
				if ctErr, ok := errors.GetZeroUIError(err); ok {
					fmt.Fprintf(os.Stderr, "Error: %s\n", ctErr.String())
					return nil
				}
				return err
			}
			fmt.Printf("✓ Cleaned up old backups for %s (kept %d most recent)\n", appName, keepCount)
		} else {
			// Clean up all apps
			engine, err := toggle.NewEngine()
			if err != nil {
				return fmt.Errorf("failed to create toggle engine: %w", err)
			}

			apps, err := engine.GetApps()
			if err != nil {
				return err
			}

			for _, app := range apps {
				if err := backupManager.CleanupOldBackups(app, keepCount); err != nil {
					fmt.Printf("Warning: failed to cleanup backups for %s: %v\n", app, err)
				}
			}
			fmt.Printf("✓ Cleaned up old backups for all apps (kept %d most recent per app)\n", keepCount)
		}

		return nil
	},
}

func init() {
	// Add backup command to root
	rootCmd.AddCommand(backupCmd)

	// Add subcommands
	backupCmd.AddCommand(backupListCmd)
	backupCmd.AddCommand(backupCreateCmd)
	backupCmd.AddCommand(backupRestoreCmd)
	backupCmd.AddCommand(backupCleanupCmd)

	// Add flags
	backupRestoreCmd.Flags().BoolP("yes", "y", false, "skip confirmation prompt")
	backupCleanupCmd.Flags().IntP("keep", "k", 5, "number of backups to keep")
}

// formatSize formats a file size in bytes to a human-readable string
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
