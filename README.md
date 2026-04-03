# gitflip

**gitflip** is a small Go CLI that switches between GitHub identities (for example personal vs work). It updates `git` `user.name` / `user.email`, manages a dedicated SSH key per profile, and writes a safe, marked block in `~/.ssh/config` so `git@github.com` uses the right key.

You get a single static binary. Runtime expectations: **`git`** (for `use` and `current`), **`ssh`** and **`ssh-keygen`** (for keys and `ssh test`).

## Requirements

- [Go](https://go.dev/dl/) 1.22+ (to build from source)
- `git` in `PATH`
- `ssh` / `ssh-keygen` for SSH keys and testing

## Build

From the repository root:

```bash
make build
```

The binary is written to `bin/gitflip`.

Alternatively:

```bash
go build -o bin/gitflip .
```

## Install

System-wide (default install path `/usr/local/bin`):

```bash
make install
```

Requires write access to `/usr/local/bin`.

Install from anywhere with Go:

```bash
go install github.com/jayesh9747/gitflip@latest
```

(Ensure your `GOBIN` or `GOPATH/bin` is on `PATH`.)

## Where data lives

| Path | Purpose |
|------|---------|
| `~/.gitflip/config.json` | Profiles and active global profile |
| `~/.gitflip/keys/<name>` | Private key for profile `<name>` |
| `~/.gitflip/keys/<name>.pub` | Public key |

If you previously used the older `ghprofile` name, **`~/.ghprofile` is renamed to `~/.gitflip`** the first time config is loaded, when `~/.gitflip` does not already exist.

## Commands

```text
gitflip profile add <name>     # Interactive: name, email, username, optional keygen
gitflip profile list           # Lists profiles (* = active global)
gitflip profile show <name>    # Details for one profile
gitflip profile edit <name>    # Interactive: git name, email, GitHub username (key + SSH unchanged)
gitflip profile set-email <name> <email>  # Change stored email; updates global git if active
gitflip profile remove <name>  # Remove profile (optional key deletion)

gitflip use <name>             # Global: git --global + SSH github.com block + active
gitflip use <name> --local     # This repo only: local user.name / user.email

gitflip current                # Active global profile + local override if any

gitflip ssh generate <name>    # Create ed25519 key for that profile
gitflip ssh show <name>        # Print public key (paste into GitHub)
gitflip ssh test <name>        # Test SSH to github.com with that key
```

Global flags: `-h` / `--help`, `-v` / `--version`.

## Typical usage

### 1. Create a profile

```bash
gitflip profile add personal
```

Answer the prompts. If you generate a key, add the printed public key at [GitHub SSH keys](https://github.com/settings/ssh/new).

### 2. Verify SSH

```bash
gitflip ssh test personal
```

### 3. Switch identity (global)

```bash
gitflip use work
```

This sets global `user.name` / `user.email`, updates **`gitflip`’s managed block** for `Host github.com` in `~/.ssh/config`, and records the active profile in `~/.gitflip/config.json`.

### 4. Per-repository override

Inside one repo only:

```bash
cd /path/to/repo
gitflip use personal --local
```

Local commit identity changes for that repo; SSH for `github.com` still follows your **last global** `gitflip use` (see the note printed by the command). For fully separate SSH per repo, use extra `Host` aliases in SSH config by hand or extend your workflow.

### 5. Check what is active

```bash
gitflip current
```

## SSH config

gitflip only replaces the block between these markers (do not edit that section by hand):

```text
# gitflip: managed block — do not edit manually
...
# gitflip: end
```

Older `# ghprofile:` blocks are removed when you run `gitflip use` so you do not end up with conflicting `Host github.com` entries.

## Local git config key

Per-repo profile selection is stored as `gitflip.profile`. Legacy `ghprofile.profile` is still read for display if present.

## Clean build artifacts

```bash
make clean
```

Removes the `bin/` directory created by `make build`.


i am adding some new command.