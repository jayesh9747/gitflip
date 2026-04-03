package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/jayesh9747/gitflip/internal/config"
	"github.com/jayesh9747/gitflip/internal/gitconfig"
	"github.com/jayesh9747/gitflip/internal/keygen"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show active global profile and optional local override",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return gitconfig.RequireGit()
	},
	Run: runCurrent,
}

func runCurrent(cmd *cobra.Command, args []string) {
	root, err := config.Load()
	if err != nil {
		exitErr(err)
	}

	if root.Active == "" {
		color.New(color.FgYellow).Println("No global active profile set. Run: gitflip use <name>")
	} else {
		p, err := root.GetProfile(root.Active)
		if err != nil {
			fmt.Printf("Active profile (global): %s (missing from config — run profile list)\n", root.Active)
		} else {
			fmt.Printf("Active profile: %s (global)\n", root.Active)
			fmt.Printf("  Name:     %s\n", p.Name)
			fmt.Printf("  Email:    %s\n", p.Email)
			fmt.Printf("  Username: %s\n", p.Username)
			fmt.Printf("  SSH Key:  %s", p.SSHKeyPath)
			if keygen.KeyPairExists(p.SSHKeyPath) {
				color.New(color.FgGreen).Println(" ✓")
			} else {
				color.New(color.FgYellow).Println(" (missing)")
			}
		}
	}

	inRepo, err := gitconfig.InGitRepo()
	if err != nil {
		exitErr(err)
	}
	if !inRepo {
		return
	}
	profName, _, _ := gitconfig.GetLocal()
	if profName == "" {
		return
	}
	lp, err := root.GetProfile(profName)
	if err != nil {
		fmt.Printf("\nLocal override (this repo): %s (profile not found in config)\n", profName)
		return
	}
	fmt.Println()
	fmt.Printf("Local override (this repo): %s\n", profName)
	fmt.Printf("  Name:     %s\n", lp.Name)
	fmt.Printf("  Email:    %s\n", lp.Email)
}
