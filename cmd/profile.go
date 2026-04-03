package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jayesh9747/gitflip/internal/config"
	"github.com/jayesh9747/gitflip/internal/gitconfig"
	"github.com/jayesh9747/gitflip/internal/keygen"
	"github.com/jayesh9747/gitflip/internal/prompt"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage named profiles",
}

func init() {
	profileCmd.AddCommand(profileAddCmd, profileListCmd, profileRemoveCmd, profileShowCmd, profileSetEmailCmd, profileEditCmd)
}

// promptKeepDefault reads a line; blank input keeps current.
func promptKeepDefault(label, current string) (string, error) {
	s, err := prompt.Line(fmt.Sprintf("? %s [%s]: ", label, current))
	if err != nil {
		return "", err
	}
	if s == "" {
		return current, nil
	}
	return s, nil
}

// syncActiveGlobalGit sets global user.name / user.email when profileKey is active.
func syncActiveGlobalGit(root *config.Root, profileKey string, prof config.Profile) error {
	if root.Active != profileKey {
		return nil
	}
	if err := gitconfig.RequireGit(); err != nil {
		return err
	}
	if err := gitconfig.SetGlobal(prof.Name, prof.Email); err != nil {
		return fmt.Errorf("git config --global: %w", err)
	}
	return nil
}

var profileAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a new profile",
	Args:  cobra.ExactArgs(1),
	Run:   runProfileAdd,
}

func runProfileAdd(cmd *cobra.Command, args []string) {
	name := args[0]
	if err := config.EnsureDirs(); err != nil {
		exitErr(fmt.Errorf("create config directory: %w", err))
	}
	root, err := config.Load()
	if err != nil {
		exitErr(err)
	}
	if _, ok := root.Profiles[name]; ok {
		exitErr(fmt.Errorf("profile %q already exists", name))
	}

	displayName, err := prompt.Line("? Enter your name: ")
	if err != nil || displayName == "" {
		exitErr(fmt.Errorf("name is required"))
	}
	email, err := prompt.Line("? Enter your email: ")
	if err != nil || email == "" {
		exitErr(fmt.Errorf("email is required"))
	}
	username, err := prompt.Line("? Enter your GitHub username: ")
	if err != nil || username == "" {
		exitErr(fmt.Errorf("GitHub username is required"))
	}

	keyPath, err := config.KeyPathForProfile(name)
	if err != nil {
		exitErr(err)
	}

	gen, err := prompt.YesNo("? Generate SSH key for this profile?", true)
	if err != nil {
		exitErr(err)
	}

	if gen {
		if keygen.KeyPairExists(keyPath) {
			overwrite, err := prompt.YesNo(fmt.Sprintf("Key already exists at %s. Overwrite?", keyPath), false)
			if err != nil {
				exitErr(err)
			}
			if !overwrite {
				gen = false
			} else {
				_ = os.Remove(keyPath)
				_ = os.Remove(keygen.PublicKeyPath(keyPath))
			}
		}
		if gen {
			if err := keygen.GenerateEd25519(keyPath, email); err != nil {
				exitErr(err)
			}
			green := color.New(color.FgGreen).SprintFunc()
			fmt.Println(green("✓ SSH key generated at"), keyPath)
		}
	}

	prof := config.Profile{
		Name:       displayName,
		Email:      email,
		Username:   username,
		SSHKeyPath: keyPath,
		CreatedAt:  time.Now().UTC(),
	}
	root.SetProfile(name, prof)
	if err := config.Save(root); err != nil {
		exitErr(fmt.Errorf("save config: %w", err))
	}
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Println(green("✓ Profile \"" + name + "\" created"))

	if gen && keygen.KeyPairExists(keyPath) {
		pub, err := keygen.ReadPublicKey(keyPath)
		if err == nil {
			cyan := color.New(color.FgCyan).SprintFunc()
			fmt.Println()
			fmt.Println(cyan("── Your Public Key (add this to GitHub) ──────────────────"))
			fmt.Println(string(pub))
			fmt.Println()
			fmt.Println("→ Go to: https://github.com/settings/ssh/new")
			fmt.Println("  Paste the key above and save.")
			fmt.Println()
			fmt.Println("Run `gitflip ssh test " + name + "` to verify the connection.")
		}
	}
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	Run: func(cmd *cobra.Command, args []string) {
		root, err := config.Load()
		if err != nil {
			exitErr(err)
		}
		if len(root.Profiles) == 0 {
			fmt.Println("No profiles yet. Run: gitflip profile add <name>")
			return
		}
		names := make([]string, 0, len(root.Profiles))
		for n := range root.Profiles {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			mark := " "
			if n == root.Active {
				mark = "*"
			}
			fmt.Printf("%s %s\n", mark, n)
		}
	},
}

var profileRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Delete a profile",
	Args:  cobra.ExactArgs(1),
	Run:   runProfileRemove,
}

func runProfileRemove(cmd *cobra.Command, args []string) {
	name := args[0]
	root, err := config.Load()
	if err != nil {
		exitErr(err)
	}
	if _, err := root.GetProfile(name); err != nil {
		exitErr(err)
	}

	if root.Active == name {
		ok, err := prompt.YesNo(fmt.Sprintf("Profile %q is the active global profile. Remove anyway?", name), false)
		if err != nil {
			exitErr(err)
		}
		if !ok {
			fmt.Println("Cancelled.")
			return
		}
	}

	delKeys, err := prompt.YesNo("Remove SSH private/public key files for this profile?", true)
	if err != nil {
		exitErr(err)
	}

	prof, _ := root.GetProfile(name)
	root.RemoveProfile(name)
	if err := config.Save(root); err != nil {
		exitErr(fmt.Errorf("save config: %w", err))
	}

	if delKeys && prof.SSHKeyPath != "" {
		_ = os.Remove(prof.SSHKeyPath)
		_ = os.Remove(keygen.PublicKeyPath(prof.SSHKeyPath))
	}
	color.New(color.FgGreen).Printf("✓ Profile %q removed\n", name)
}

var profileSetEmailCmd = &cobra.Command{
	Use:   "set-email <name> <email>",
	Short: "Update a profile's email in config (and global git if that profile is active)",
	Args:  cobra.ExactArgs(2),
	Run:   runProfileSetEmail,
}

func runProfileSetEmail(cmd *cobra.Command, args []string) {
	name := args[0]
	email := strings.TrimSpace(args[1])
	if email == "" {
		exitErr(fmt.Errorf("email must not be empty"))
	}
	root, err := config.Load()
	if err != nil {
		exitErr(err)
	}
	prof, err := root.GetProfile(name)
	if err != nil {
		exitErr(err)
	}
	prof.Email = email
	root.SetProfile(name, prof)
	if err := config.Save(root); err != nil {
		exitErr(fmt.Errorf("save config: %w", err))
	}
	color.New(color.FgGreen).Printf("✓ Profile %q email updated\n", name)
	fmt.Printf("  email: %s\n", email)

	if err := syncActiveGlobalGit(root, name, prof); err != nil {
		exitErr(err)
	}
	if root.Active == name {
		fmt.Println("  git config --global user.name / user.email → updated (active profile)")
	}
}

var profileEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Update git name, email, and GitHub username (profile key and SSH key unchanged)",
	Args:  cobra.ExactArgs(1),
	Run:   runProfileEdit,
}

func runProfileEdit(cmd *cobra.Command, args []string) {
	profileKey := args[0]
	root, err := config.Load()
	if err != nil {
		exitErr(err)
	}
	prof, err := root.GetProfile(profileKey)
	if err != nil {
		exitErr(err)
	}

	fmt.Printf("Editing profile %q — Enter keeps each field. SSH key path is not changed.\n\n", profileKey)

	displayName, err := promptKeepDefault("Name (git user.name)", prof.Name)
	if err != nil {
		exitErr(err)
	}
	if strings.TrimSpace(displayName) == "" {
		exitErr(fmt.Errorf("name must not be empty"))
	}
	email, err := promptKeepDefault("Email", prof.Email)
	if err != nil {
		exitErr(err)
	}
	if strings.TrimSpace(email) == "" {
		exitErr(fmt.Errorf("email must not be empty"))
	}
	username, err := promptKeepDefault("GitHub username", prof.Username)
	if err != nil {
		exitErr(err)
	}
	if strings.TrimSpace(username) == "" {
		exitErr(fmt.Errorf("GitHub username must not be empty"))
	}

	prof.Name = displayName
	prof.Email = email
	prof.Username = username
	root.SetProfile(profileKey, prof)
	if err := config.Save(root); err != nil {
		exitErr(fmt.Errorf("save config: %w", err))
	}

	color.New(color.FgGreen).Printf("✓ Profile %q updated\n", profileKey)
	fmt.Printf("  name:     %s\n", prof.Name)
	fmt.Printf("  email:    %s\n", prof.Email)
	fmt.Printf("  username: %s\n", prof.Username)

	if err := syncActiveGlobalGit(root, profileKey, prof); err != nil {
		exitErr(err)
	}
	if root.Active == profileKey {
		fmt.Println("  git config --global user.name / user.email → updated (active profile)")
	}
}

var profileShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show profile details",
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
		fmt.Printf("Name:       %s\n", p.Name)
		fmt.Printf("Email:      %s\n", p.Email)
		fmt.Printf("Username:   %s\n", p.Username)
		fmt.Printf("SSH key:    %s\n", p.SSHKeyPath)
		if p.CreatedAt.IsZero() {
			fmt.Printf("Created:    (unknown)\n")
		} else {
			fmt.Printf("Created:    %s\n", p.CreatedAt.Format(time.RFC3339))
		}
		exists := keygen.KeyPairExists(p.SSHKeyPath)
		if exists {
			color.New(color.FgGreen).Println("Key file:   present ✓")
		} else {
			color.New(color.FgYellow).Println("Key file:   missing (generate with gitflip ssh generate)")
		}
	},
}
