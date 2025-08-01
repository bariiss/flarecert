package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
	commit  = "dev"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display version information for FlareCert.`,
	Run:   runVersionCommand,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersionCommand(cmd *cobra.Command, args []string) {
	fmt.Printf("🔐 FlareCert v%s\n", version)
	fmt.Printf("📅 Build Date: %s\n", date)
	fmt.Printf("🔧 Commit: %s\n", commit)
	fmt.Printf("🐹 Go Version: %s\n", runtime.Version())
	fmt.Printf("💻 OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()
	fmt.Printf("🌐 Repository: https://github.com/bariiss/flarecert\n")
	fmt.Printf("📖 Documentation: https://github.com/bariiss/flarecert/blob/main/README.md\n")
}
