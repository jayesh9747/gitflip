package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/jayesh9747/gitflip/internal/config"
	"github.com/jayesh9747/gitflip/internal/keygen"
	"github.com/jayesh9747/gitflip/internal/prompt"
	"github.com/spf13/cobra"
)

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH key helpers for profiles",
}

func init() {
	sshCmd.AddCommand(sshGenerateCmd, sshShowCmd, sshTestCmd)
}

var sshGenerateCmd = &cobra.Command{
	Use:   "generate <profile>",
	Short: "Generate an ed25519 SSH key for a profile",
	Args:  cobra.ExactArgs(1),
	Run:   runSSHGenerate,
}

func runSSHGenerate(cmd *cobra.Command, args []string) {
	name := args[0]
	root, err := config.Load()
	if err != nil {
		exitErr(err)
	}
	prof, err := root.GetProfile(name)
	if err != nil {
		exitErr(err)
	}
	keyPath := prof.SSHKeyPath
	if keyPath == "" {
		var kerr error
		keyPath, kerr = config.KeyPathForProfile(name)
		if kerr != nil {
			exitErr(kerr)
		}
		prof.SSHKeyPath = keyPath
		root.SetProfile(name, prof)
		_ = config.Save(root)
	}

	if keygen.KeyPairExists(keyPath) {
		overwrite, err := prompt.YesNo(fmt.Sprintf("Key already exists at %s. Overwrite?", keyPath), false)
		if err != nil {
			exitErr(err)
		}
		if !overwrite {
			fmt.Println("Cancelled.")
			return
		}
		_ = os.Remove(keyPath)
		_ = os.Remove(keygen.PublicKeyPath(keyPath))
	}

	if err := config.EnsureDirs(); err != nil {
		exitErr(err)
	}
	if err := keygen.GenerateEd25519(keyPath, prof.Email); err != nil {
		exitErr(err)
	}
	color.New(color.FgGreen).Printf("✓ SSH key generated at %s\n", keyPath)
	pub, err := keygen.ReadPublicKey(keyPath)
	if err == nil {
		fmt.Println(string(pub))
	}
}

var sshShowCmd = &cobra.Command{
	Use:   "show <profile>",
	Short: "Print the public key (for GitHub)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		root, err := config.Load()
		if err != nil {
			exitErr(err)
		}
		p, err := root.GetProfile(args[0])
		if err != nil {
			exitErr(err)
		}
		if p.SSHKeyPath == "" {
			exitErr(fmt.Errorf("no ssh key path for profile %q", args[0]))
		}
		pub, err := keygen.ReadPublicKey(p.SSHKeyPath)
		if err != nil {
			exitErr(fmt.Errorf("read public key: %w", err))
		}
		fmt.Print(string(pub))
		if !strings.HasSuffix(string(pub), "\n") {
			fmt.Println()
		}
	},
}

var sshTestCmd = &cobra.Command{
	Use:   "test <profile>",
	Short: "Test SSH authentication to GitHub with this profile's key",
	Args:  cobra.ExactArgs(1),
	Run:   runSSHTest,
}

func runSSHTest(cmd *cobra.Command, args []string) {
	name := args[0]
	root, err := config.Load()
	if err != nil {
		exitErr(err)
	}
	p, err := root.GetProfile(name)
	if err != nil {
		exitErr(err)
	}
	if !keygen.KeyPairExists(p.SSHKeyPath) {
		exitErr(fmt.Errorf("private key missing at %s — run: gitflip ssh generate %s", p.SSHKeyPath, name))
	}

	c := exec.Command("ssh",
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		"-i", p.SSHKeyPath,
		"-T",
		"git@github.com",
	)
	out, err := c.CombinedOutput()
	msg := strings.TrimSpace(string(out))
	if ghSSHSuccess(msg) {
		color.New(color.FgGreen).Printf("✓ SSH connection to GitHub successful for user %q\n", p.Username)
		return
	}
	if err != nil {
		exitErr(fmt.Errorf("ssh test failed: %w\n%s", err, msg))
	}
	exitErr(fmt.Errorf("unexpected ssh output:\n%s", msg))
}

func ghSSHSuccess(msg string) bool {
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "successfully authenticated") {
		return true
	}
	return strings.Contains(msg, "Hi ") && strings.Contains(msg, "You've successfully authenticated")
}
