package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Line reads a trimmed line from stdin.
func Line(prompt string) (string, error) {
	fmt.Print(prompt)
	s, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(s), nil
}

// YesNo asks with default true if defYes.
func YesNo(prompt string, defYes bool) (bool, error) {
	suffix := " (Y/n): "
	if !defYes {
		suffix = " (y/N): "
	}
	ans, err := Line(prompt + suffix)
	if err != nil {
		return false, err
	}
	if ans == "" {
		return defYes, nil
	}
	switch strings.ToLower(ans) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		if defYes {
			return true, nil
		}
		return false, nil
	}
}
