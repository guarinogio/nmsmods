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
5. Registers `nxm://` as an OS URL handler (so "Mod Manager Download" opens `nmsmods`)
6. Ensures `~/.local/bin` is added to `PATH` for future shells (via `~/.profile`, if needed)

After install, open a **new terminal** (or log out/in) so `PATH` changes take effect.

If you still get `command not found: nmsmods`, ensure `~/.local/bin` is in your `PATH`.
The installer adds a portable snippet to `~/.profile`, but some environments may require manual setup.

### Build from source

Requirements:

- Go 1.22+
- Git

```bash
git clone https://github.com/guarinogio/nmsmods.git
cd nmsmods
make install

# installs the nxm:// handler and prints a PATH hint if needed
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

### Managed deploy marker + overwrite policy

When `nmsmods` deploys a mod into `GAMEDATA/MODS/<folder>`, it also writes a small marker file:

- `GAMEDATA/MODS/<folder>/.nmsmods.managed.json`

This is used for safety:

- `nmsmods` **refuses to overwrite** an existing `GAMEDATA/MODS/<folder>` that is not marked as managed.
- `nmsmods` only **undeploys** folders that are managed *and* match the expected mod/profile.

If you have manually-installed mods that conflict with a folder name `nmsmods` wants to use, you can:

- rename/remove the conflicting folder in `GAMEDATA/MODS`, or
- install the mod under a different profile/id so the folder name changes.

To override the state directory:

```bash
export NMSMODS_HOME=/path/to/isolated/home
```

---

## Basic usage

## Quickstart (one-click Nexus installs)

1) Auto-detect and set the game path:

```bash
nmsmods set-path --auto
```

If you have multiple installs, `--auto` will refuse and print the candidates.

2) Log in to Nexus (stores your API key under XDG config):

```bash
nmsmods nexus login --api-key "YOUR_KEY"
```

3) Go to a mod page on Nexus and click **Mod Manager Download**.

Because `install.sh` (and `make install`) registers the `nxm://` handler, your system will open:

```bash
nmsmods nxm handle "nxm://..."
```

Behavior:

- If the mod is not installed in the active profile: **download → install → enable → deploy**
- If the same Nexus file is already installed: **no-op** (ensures it is deployed/enabled)
- If a different Nexus file is clicked for the same mod: **download → update (reinstall) → deploy**

The handler also writes a log at:

- `~/.local/state/nmsmods/nxm-handler.log`

And (if `notify-send` is available) shows a desktop notification.

### Detect game

```bash
nmsmods doctor
```

`doctor` also shows what is currently inside your real `GAMEDATA/MODS` directory and splits it into:

- **managed by nmsmods** (folders containing `.nmsmods.managed.json`)
- **external/unmanaged** (folders created manually or by other tools)

If auto-detection fails:

```bash
nmsmods set-path --auto
# or:
nmsmods set-path "/path/to/No Man's Sky"
```

`set-path` stores the **canonicalized** path (after resolving symlinks). If your input path is a symlinked Steam path,
the saved path may differ (this prevents duplicate entries for the same installation).

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
