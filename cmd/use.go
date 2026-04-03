package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/jayesh9747/gitflip/internal/config"
	"github.com/jayesh9747/gitflip/internal/gitconfig"
	"github.com/jayesh9747/gitflip/internal/keygen"
	"github.com/jayesh9747/gitflip/internal/sshconfig"
	"github.com/spf13/cobra"
)

var useLocal bool

var useCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Activate a profile (global git + SSH for github.com, or local git only)",
	Args:  cobra.ExactArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return gitconfig.RequireGit()
	},
	Run: runUse,
}

func init() {
	useCmd.Flags().BoolVar(&useLocal, "local", false, "Set user.name/email only in the current git repository")
}

func runUse(cmd *cobra.Command, args []string) {
	name := args[0]
	root, err := config.Load()
	if err != nil {
		exitErr(err)
	}
	prof, err := root.GetProfile(name)
	if err != nil {
		exitErr(err)
	}

	if useLocal {
		ok, err := gitconfig.InGitRepo()
		if err != nil {
			exitErr(err)
		}
		if !ok {
			exitErr(fmt.Errorf("not inside a git repository (use --local only inside a repo)"))
		}
		if err := gitconfig.SetLocal(name, prof.Name, prof.Email); err != nil {
			exitErr(fmt.Errorf("git config --local: %w", err))
		}
		color.New(color.FgGreen).Printf("✓ Local git config updated for this repo only\n")
		fmt.Printf("  name:  %s\n", prof.Name)
		fmt.Printf("  email: %s\n", prof.Email)
		fmt.Println()
		color.New(color.FgYellow).Println("Note: SSH to github.com still uses your globally active profile’s key unless you use separate hosts.")
		return
	}

	if !keygen.KeyPairExists(prof.SSHKeyPath) {
		color.New(color.FgYellow).Printf("Warning: SSH private key not found at %s — GitHub SSH auth may fail until you run: gitflip ssh generate %s\n\n", prof.SSHKeyPath, name)
	}

	if err := gitconfig.SetGlobal(prof.Name, prof.Email); err != nil {
		exitErr(fmt.Errorf("git config --global: %w", err))
	}
	if err := sshconfig.WriteGitHubBlock(prof.SSHKeyPath); err != nil {
		exitErr(fmt.Errorf("update ~/.ssh/config: %w", err))
	}
	root.Active = name
	if err := config.Save(root); err != nil {
		exitErr(fmt.Errorf("save config: %w", err))
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Println(green("✓ Switched to profile \"" + name + "\""))
	fmt.Printf("  name:  %s\n", prof.Name)
	fmt.Printf("  email: %s\n", prof.Email)
	fmt.Println("  git config --global user.name / user.email → updated")
	fmt.Println("  ~/.ssh/config github.com block → updated")
}
