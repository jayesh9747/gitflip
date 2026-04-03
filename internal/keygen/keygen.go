package keygen

import (
	"fmt"
	"os"
	"os/exec"
)

// GenerateEd25519 runs ssh-keygen -t ed25519 -f path -N "" -C comment.
func GenerateEd25519(privateKeyPath, comment string) error {
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", privateKeyPath, "-N", "", "-C", comment)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ssh-keygen: %w", err)
	}
	return nil
}

// PublicKeyPath returns the .pub path for a private key path.
func PublicKeyPath(privateKeyPath string) string {
	return privateKeyPath + ".pub"
}

// ReadPublicKey returns the contents of the .pub file.
func ReadPublicKey(privateKeyPath string) ([]byte, error) {
	return os.ReadFile(PublicKeyPath(privateKeyPath))
}

// KeyPairExists returns true if private (and typically public) key files exist.
func KeyPairExists(privateKeyPath string) bool {
	_, err := os.Stat(privateKeyPath)
	return err == nil
}
