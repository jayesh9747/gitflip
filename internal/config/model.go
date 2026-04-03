package config

import "time"

// Root is the on-disk shape of ~/.gitflip/config.json.
type Root struct {
	Active   string             `json:"active"`
	Profiles map[string]Profile `json:"profiles"`
}

// Profile describes one GitHub identity.
type Profile struct {
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Username   string    `json:"username"`
	SSHKeyPath string    `json:"ssh_key_path"`
	CreatedAt  time.Time `json:"created_at"`
}
