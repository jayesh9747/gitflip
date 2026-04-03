package sshconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	managedStart = "# gitflip: managed block — do not edit manually\n"
	managedEnd   = "# gitflip: end\n"
	// legacy markers from the previous CLI name (replaced on next switch)
	legacyManagedStart = "# ghprofile: managed block — do not edit manually\n"
	legacyEndLine      = "# ghprofile: end"
)

// DefaultSSHConfigPath is ~/.ssh/config.
func DefaultSSHConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh", "config"), nil
}

// WriteGitHubBlock replaces or appends the managed Host github.com block.
func WriteGitHubBlock(identityFile string) error {
	path, err := DefaultSSHConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	expanded, err := expandPath(identityFile)
	if err != nil {
		return err
	}
	block := managedStart +
		"Host github.com\n" +
		"  HostName github.com\n" +
		"  User git\n" +
		fmt.Sprintf("  IdentityFile %s\n", expanded) +
		"  IdentitiesOnly yes\n" +
		managedEnd

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(data)
	content = stripManagedRegion(content, legacyManagedStart, legacyEndLine)
	content = stripManagedRegion(content, managedStart, strings.TrimSuffix(managedEnd, "\n"))
	newContent := appendManagedBlock(content, block)
	mode := os.FileMode(0o600)
	if st, err := os.Stat(path); err == nil {
		mode = st.Mode() & 0o777
	}
	return os.WriteFile(path, []byte(newContent), mode)
}

func expandPath(p string) (string, error) {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		p = filepath.Join(home, p[2:])
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return p, nil
	}
	return abs, nil
}

// stripManagedRegion removes one managed block starting with startMarker; endMarker is without trailing newline.
func stripManagedRegion(content, startMarker, endMarker string) string {
	startIdx := strings.Index(content, startMarker)
	if startIdx == -1 {
		return content
	}
	rest := content[startIdx+len(startMarker):]
	endIdx := strings.Index(rest, endMarker)
	if endIdx == -1 {
		return content[:startIdx]
	}
	fullEnd := startIdx + len(startMarker) + endIdx + len(endMarker)
	for fullEnd < len(content) && (content[fullEnd] == '\n' || content[fullEnd] == '\r') {
		fullEnd++
	}
	return content[:startIdx] + content[fullEnd:]
}

func appendManagedBlock(content, block string) string {
	if strings.Contains(content, managedStart) {
		// should not happen after strip; strip again defensively
		content = stripManagedRegion(content, managedStart, strings.TrimSuffix(managedEnd, "\n"))
	}
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	sep := "\n"
	if content == "" {
		sep = ""
	}
	return content + sep + block
}
