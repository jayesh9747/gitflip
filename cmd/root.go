package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "gitflip",
	Short: "Switch GitHub identities (personal, work, …) via git config + SSH",
	Long: `gitflip manages multiple GitHub identities using SSH keys and git user.name/user.email.

Use "use" and "current" require git in PATH. SSH commands require ssh-keygen/ssh. Config: ~/.gitflip/`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	rootCmd.AddCommand(profileCmd, useCmd, currentCmd, sshCmd)
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}
