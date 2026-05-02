# gitflip

**gitflip** is a small Go CLI that switches between GitHub identities (for example personal vs work). It updates `git` `user.name` / `user.email`, manages a dedicated SSH key per profile, and writes a safe, marked block in `~/.ssh/config` so `git@github.com` uses the right key.

You get a single static binary. Runtime expectations: **`git`** (for `use` and `current`), **`ssh`** and **`ssh-keygen`** (for keys and `ssh test`).

### SSH agent & profile switching

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

This uses the Go module proxy, not GitHub release assets. For a standalone binary without Go, use the release installers below.

## Install without Go (prebuilt binary)

Releases are built from git tags (`v0.1.0`, …) and published on [GitHub Releases](https://github.com/jayesh9747/gitflip/releases). You only need `curl` and `tar`.

### Linux / macOS

**Latest release — user-writable directory** (no `sudo`):

```bash
mkdir -p ~/.local/bin
curl -fsSL https://raw.githubusercontent.com/jayesh9747/gitflip/main/scripts/install.sh | INSTALL_DIR="$HOME/.local/bin" sh
```

Ensure `~/.local/bin` is on your `PATH`.

**Specific version** (pass variables to `sh`, not only `curl`):

```bash
curl -fsSL https://raw.githubusercontent.com/jayesh9747/gitflip/main/scripts/install.sh | VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" sh
```

**System-wide** (needs write access to `/usr/local/bin`, often `sudo`):

```bash
curl -fsSL https://raw.githubusercontent.com/jayesh9747/gitflip/main/scripts/install.sh | sudo sh
```

Or download the `gitflip_<version>_<os>_<arch>.tar.gz` for your platform from the release page and extract the `gitflip` binary yourself.

### Windows

PowerShell installer:

```powershell
Set-ExecutionPolicy -Scope Process Bypass
irm https://raw.githubusercontent.com/jayesh9747/gitflip/main/scripts/install.ps1 | iex
```

Install a specific version:

```powershell
& ([scriptblock]::Create((irm https://raw.githubusercontent.com/jayesh9747/gitflip/main/scripts/install.ps1))) -Version v0.1.0
```

By default this installs to `%USERPROFILE%\AppData\Local\Programs\gitflip\bin`. Add that directory to your `PATH` if needed.

Or download `gitflip_<version>_windows_<arch>.zip` from the release page and extract `gitflip.exe` manually.

## Release flow

This project publishes binaries to **GitHub Releases**, not to the GitHub **Packages** tab. For a Go CLI, that is the normal distribution path unless you also publish to a package manager.

The release automation lives in:

- `.github/workflows/release.yml`
- `.goreleaser.yaml`

When you push a tag that starts with `v`, GitHub Actions runs GoReleaser and uploads:

- Linux archives: `gitflip_<version>_linux_amd64.tar.gz`, `gitflip_<version>_linux_arm64.tar.gz`
- macOS archives: `gitflip_<version>_darwin_amd64.tar.gz`, `gitflip_<version>_darwin_arm64.tar.gz`
- Windows archives: `gitflip_<version>_windows_amd64.zip`
- `checksums.txt`

### Release a new version

1. Commit and push the release files (`.github/workflows/release.yml`, `.goreleaser.yaml`, `scripts/install.sh`, `scripts/install.ps1`, README changes).
2. Create an annotated tag:

   ```bash
   git tag -a v0.1.1 -m "Release v0.1.1"
   ```

3. Push the branch and tag:

   ```bash
   git push origin main
   git push origin v0.1.1
   ```

4. Open `https://github.com/jayesh9747/gitflip/actions` and wait for the `release` workflow to finish.
5. Open `https://github.com/jayesh9747/gitflip/releases` and verify the assets are attached to the new release.

### Test the release config locally

If you have GoReleaser installed:

```bash
goreleaser release --snapshot --clean
```

That builds the release artifacts locally without publishing them.

## Where data lives

| Path                         | Purpose                            |
| ---------------------------- | ---------------------------------- |
| `~/.gitflip/config.json`     | Profiles and active global profile |
| `~/.gitflip/keys/<name>`     | Private key for profile `<name>`   |
| `~/.gitflip/keys/<name>.pub` | Public key                         |

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
