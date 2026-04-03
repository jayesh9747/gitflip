package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	dirName  = ".gitflip"
	fileName = "config.json"
	keysDir  = "keys"
)

var ErrNotFound = errors.New("profile not found")

// Dir returns ~/.gitflip (expanded).
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, dirName), nil
}

// KeysDir returns ~/.gitflip/keys.
func KeysDir() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, keysDir), nil
}

// ConfigPath returns path to config.json.
func ConfigPath() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, fileName), nil
}

// KeyPathForProfile returns private key path for a profile name (no extension).
func KeyPathForProfile(profileName string) (string, error) {
	kd, err := KeysDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(kd, profileName), nil
}

// EnsureDirs creates ~/.gitflip and keys dir if missing.
func EnsureDirs() error {
	d, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(d, keysDir), 0o700); err != nil {
		return err
	}
	return nil
}

// migrateFromLegacyDir renames ~/.ghprofile → ~/.gitflip if the new dir does not exist yet.
func migrateFromLegacyDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	newDir := filepath.Join(home, ".gitflip")
	oldDir := filepath.Join(home, ".ghprofile")
	if _, err := os.Stat(newDir); err == nil {
		return nil
	}
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		return nil
	}
	return os.Rename(oldDir, newDir)
}

// Load reads config.json; missing file yields empty Root with profiles map initialized.
func Load() (*Root, error) {
	if err := migrateFromLegacyDir(); err != nil {
		return nil, fmt.Errorf("migrate legacy config dir: %w", err)
	}
	p, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Root{Profiles: make(map[string]Profile)}, nil
		}
		return nil, err
	}
	var r Root
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if r.Profiles == nil {
		r.Profiles = make(map[string]Profile)
	}
	return &r, nil
}

// Save writes config.json atomically.
func Save(r *Root) error {
	if err := EnsureDirs(); err != nil {
		return err
	}
	p, err := ConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// GetProfile returns a profile by name.
func (r *Root) GetProfile(name string) (Profile, error) {
	p, ok := r.Profiles[name]
	if !ok {
		return Profile{}, ErrNotFound
	}
	return p, nil
}

// SetProfile adds or updates a profile.
func (r *Root) SetProfile(name string, p Profile) {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	r.Profiles[name] = p
}

// RemoveProfile deletes a profile entry.
func (r *Root) RemoveProfile(name string) {
	delete(r.Profiles, name)
	if r.Active == name {
		r.Active = ""
	}
}
