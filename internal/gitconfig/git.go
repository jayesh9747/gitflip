package gitconfig

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const localProfileKey = "gitflip.profile"

// RequireGit exits the process if git is not available.
func RequireGit() error {
	cmd := exec.Command("git", "version")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git is required but not found in PATH: %w\nInstall from https://git-scm.com/downloads", err)
	}
	if len(strings.TrimSpace(string(out))) == 0 {
		return fmt.Errorf("git is required but did not respond; install from https://git-scm.com/downloads")
	}
	return nil
}

// InGitRepo returns true if cwd is inside a git repository.
func InGitRepo() (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	out, err := cmd.Output()
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(out)) == "true", nil
}

// SetGlobal sets user.name and user.email globally.
func SetGlobal(name, email string) error {
	if err := runGit("config", "--global", "user.name", name); err != nil {
		return err
	}
	return runGit("config", "--global", "user.email", email)
}

// SetLocal sets user.name, user.email, and gitflip.profile in the current repo.
func SetLocal(profileName, name, email string) error {
	if err := runGit("config", "--local", "user.name", name); err != nil {
		return err
	}
	if err := runGit("config", "--local", "user.email", email); err != nil {
		return err
	}
	return runGit("config", "--local", localProfileKey, profileName)
}

// GetGlobal reads global user.name and user.email.
func GetGlobal() (name, email string, err error) {
	name, err = getGit("config", "--global", "--get", "user.name")
	if err != nil {
		return "", "", err
	}
	email, err = getGit("config", "--global", "--get", "user.email")
	return name, email, err
}

// GetLocal reads local user.name, user.email, and gitflip.profile if set.
// A legacy ghprofile.profile value is still read if present.
func GetLocal() (profile, name, email string) {
	prof, _ := getGit("config", "--local", "--get", localProfileKey)
	if prof == "" {
		prof, _ = getGit("config", "--local", "--get", "ghprofile.profile")
	}
	n, _ := getGit("config", "--local", "--get", "user.name")
	e, _ := getGit("config", "--local", "--get", "user.email")
	return prof, n, e
}

func runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func getGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
