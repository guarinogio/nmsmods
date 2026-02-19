# nmsmods

A minimal, fast CLI mod manager for **No Man's Sky (Linux)**.

`nmsmods` helps you:

- Detect your No Man's Sky installation
- Download mod ZIP files (direct URL, local ZIP, or Nexus)
- Install mods into an **active profile**
- Enable/disable mods without deleting them
- Deploy the active profile into `GAMEDATA/MODS`
- List installed/deployed mods
- Uninstall mods cleanly
- Track downloaded files and installed folders

It is designed to be:

- Simple
- Transparent
- Scriptable
- Safe by default (ZIP hardening + atomic state writes)

---

# Important Context (No Man's Sky 5.58+ / 6.x)

No Man's Sky modding changed significantly after 5.50 and again after 5.58.

Key points:

- Mods must go in:

  `<GameDir>/GAMEDATA/MODS`

  NOT:

  `PCBANKS/MODS`

- `.EXML` and `.MBIN` files are valid.
- `.MXML` files (except `LocTable.MXML`) are ignored by the game.
- Old PAK-based mods generally do not work anymore.

Always ensure the mod you download:

- Is updated for your game version
- Was updated after January 2025 (for 5.50+ compatibility)

---

# Supported Platforms

- Linux (native Steam install)
- Steam under Proton

Architectures:

- amd64 (x86_64)
- arm64 (aarch64)

---

# Installation

## Recommended (no Go required)

```bash
curl -fsSL https://raw.githubusercontent.com/guarinogio/nmsmods/main/install.sh | bash
```

The installer:

1. Detects your CPU architecture
2. Downloads the latest GitHub Release binary
3. **Verifies SHA-256** using the release `checksums.txt`
4. Installs to `~/.local/bin/nmsmods`

Then run:

```bash
nmsmods doctor
```

## Build from source

Requirements:

- Go 1.22+
- Git

```bash
git clone https://github.com/guarinogio/nmsmods.git
cd nmsmods
make install
```

---

# Data layout

`nmsmods` follows XDG by default:

- Config: `~/.config/nmsmods/config.json`
- State/cache: `~/.local/state/nmsmods/`

Within the state directory:

```
~/.local/state/nmsmods/
 ├── state.json
 ├── downloads/
 ├── staging/
 └── profiles/
     └── default/
         └── mods/
```

Each profile has its own **authoritative store** under `profiles/<name>/mods/`.
When a mod is enabled, it is **deployed** into the game `GAMEDATA/MODS` directory.

---

# Basic usage

## Detect game

```bash
nmsmods doctor
```

If auto-detection fails:

```bash
nmsmods set-path "/path/to/No Man's Sky"
```

## Profiles

Show active profile:

```bash
nmsmods profile status
```

List profiles:

```bash
nmsmods profile list
```

Switch profile and deploy it to the game:

```bash
nmsmods profile use vanilla
# or re-deploy current profile:
nmsmods profile deploy
```

## Download

Direct URL:

```bash
nmsmods download "https://example.com/mod.zip"
```

Local ZIP:

```bash
nmsmods download ./SomeMod.zip --id some-mod
```

List downloads:

```bash
nmsmods downloads
```

## Install / enable / disable

Install (stores into active profile + deploys):

```bash
nmsmods install 1
# or:
nmsmods install refinerwikislots
```

Disable without deleting:

```bash
nmsmods disable refinerwikislots
```

Enable again:

```bash
nmsmods enable refinerwikislots
```

Uninstall (removes from active profile store + undeploys):

```bash
nmsmods uninstall refinerwikislots
```

## Shell completion

```bash
nmsmods completion bash > /tmp/nmsmods.bash
nmsmods completion zsh  > /tmp/_nmsmods
```

---

# Nexus Mods (experimental)

The `nexus` command group supports:

- `nmsmods nexus login` (store API key)
- `nmsmods nexus whoami`
- `nmsmods nexus mod <modid>` / `nmsmods nexus files <modid>`
- `nmsmods nexus download-nxm <nxm://...>` and `nmsmods nexus resolve-nxm <nxm://...>`
- `nmsmods nexus check-updates`
- `nmsmods nexus pin <id>` (prevent updating)

Note: Nexus API policies and auth requirements can change; treat this as experimental.

---

# Safety considerations

This tool hardens ZIP extraction against:

- Path traversal (`../`)
- Absolute paths
- Symlinks
- Zip bombs (limits configurable via env vars)

State writes are atomic (temp file + rename) to reduce corruption risk.
