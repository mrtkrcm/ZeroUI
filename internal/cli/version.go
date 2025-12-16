package cli

import (
	"github.com/mrtkrcm/ZeroUI/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display version information including build details.`,
		Run: func(cmd *cobra.Command, args []string) {
			version.Print()
		},
	}
}
