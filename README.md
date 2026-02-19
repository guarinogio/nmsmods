# nmsmods

A minimal, fast CLI mod manager for **No Man's Sky (Linux / Steam / Proton)**.

`nmsmods` helps you:

- Detect your No Man's Sky installation
- Download mod ZIP files (direct URL, local ZIP, or Nexus `nxm://`)
- Install mods into an **active profile**
- Enable/disable mods without deleting them
- Deploy the active profile into `GAMEDATA/MODS`
- List downloads and installed mods
- Uninstall mods cleanly
- Track Nexus metadata for update checks (optional)

Design goals:

- Simple and transparent (files on disk are the source of truth)
- Scriptable (clean stdout, stable state)
- Safe by default (ZIP hardening + atomic state writes + state locking)

---

## Important context (NMS 5.58+ / 6.x)

No Man's Sky modding changed in recent major updates. For modern versions:

- Mods belong in:

  `<GameDir>/GAMEDATA/MODS`

  **not** `PCBANKS/MODS`.

- `.EXML` and `.MBIN` are common.
- Very old PAK-era mods may not work.

---

## Supported platforms

- Linux (native Steam)
- Steam under Proton

Architectures:

- amd64 (x86_64)
- arm64 (aarch64)

---

## Installation

### Recommended (no Go required)

```bash
curl -fsSL https://raw.githubusercontent.com/guarinogio/nmsmods/main/install.sh | bash
```

The installer:

1. Detects your CPU architecture
2. Downloads the latest GitHub Release binary
3. Verifies SHA-256 using the release `checksums.txt`
4. Installs to `~/.local/bin/nmsmods`

Then:

```bash
nmsmods doctor
```

### Build from source

Requirements:

- Go 1.22+
- Git

```bash
git clone https://github.com/guarinogio/nmsmods.git
cd nmsmods
make install
```

---

## Data layout

`nmsmods` follows XDG by default:

- Config: `~/.config/nmsmods/config.json`
- State/cache: `~/.local/state/nmsmods/`

State directory structure:

```
~/.local/state/nmsmods/
 ├── state.json
 ├── downloads/
 ├── staging/
 └── profiles/
     └── default/
         └── mods/
```

Each profile has its own authoritative store under `profiles/<name>/mods/`.
When a mod is enabled, it is deployed into the game `GAMEDATA/MODS` directory.

To override the state directory:

```bash
export NMSMODS_HOME=/path/to/isolated/home
```

---

## Basic usage

### Detect game

```bash
nmsmods doctor
```

If auto-detection fails:

```bash
nmsmods set-path "/path/to/No Man's Sky"
```

### Profiles

```bash
nmsmods profile status
nmsmods profile list

nmsmods profile use vanilla
nmsmods profile deploy
```

### Download

Direct URL:

```bash
nmsmods download "https://example.com/mod.zip" --id my-mod
```

Local ZIP:

```bash
nmsmods download ./SomeMod.zip --id some-mod
```

List downloads:

```bash
nmsmods downloads
```

### Install / enable / disable / uninstall

Install into the active profile (also deploys):

```bash
nmsmods install some-mod
```

Disable without deleting:

```bash
nmsmods disable some-mod
```

Enable again:

```bash
nmsmods enable some-mod
```

Uninstall:

```bash
nmsmods uninstall some-mod
```

### IDs vs indexes

Many commands accept either:

- an **ID** (recommended), or
- a numeric **index** from `nmsmods downloads`.

`nmsmods` prefers an **exact ID match** even if the ID is numeric.

---

## Nexus Mods (experimental)

The `nexus` command group supports:

- `nmsmods nexus login` (store API key)
- `nmsmods nexus whoami`
- `nmsmods nexus mod <modid>` / `nmsmods nexus files <modid>`
- `nmsmods nexus resolve-nxm <nxm://...>`
- `nmsmods nexus download-nxm <nxm://...> --id <id>`
- `nmsmods nexus check-updates [id-or-index]`
- `nmsmods nexus pin <id-or-index> --on/--off`

### Security note

- Prefer providing credentials via environment variables instead of pasting into terminals or issue reports.
- `nmsmods nexus whoami` **does not print your email by default**. If you want it:

```bash
nmsmods nexus whoami --show-email
```

---

## E2E script (local-only)

This repository includes `./e2e.sh`, which exercises **all commands explicitly**.

It is designed to be safe:

- Uses your real NMS install
- Backs up and restores `GAMEDATA/MODS` on exit
- Uses an isolated `NMSMODS_HOME` temp dir

Run core commands:

```bash
./e2e.sh
```

Run with Nexus coverage:

```bash
export NEXUS_API_KEY="..."
export NXM_URL="nxm://nomanssky/mods/3718/files/43996?key=..."
./e2e.sh
```

The script redacts secrets in its own command echo.

---

## CI policy

CI intentionally **does not run unit tests**.

The GitHub Actions workflow runs:

- `gofmt` check
- `go vet`
- `go build`
- `govulncheck`

This keeps CI fast and avoids fragile tests for workflows that touch real game installs.

---

## Safety considerations

This tool hardens ZIP extraction against:

- Path traversal (`../`)
- Absolute paths
- Symlinks
- Zip bombs (size/file count limits)

State writes are atomic (temp file + rename), and state/config writes are protected by a lock file to reduce corruption when multiple `nmsmods` processes run concurrently.
