# nmsmods (personal v0.1)

A basic Linux CLI to download, install, list, and uninstall No Man's Sky mods (post-5.58) by managing folders under `GAMEDATA/MODS`.

## Build

```bash
go mod tidy
go build -o nmsmods ./
```

## Install (optional)

```bash
install -m 0755 nmsmods ~/.local/bin/nmsmods
```

## Usage

1) Detect or set game path

```bash
./nmsmods doctor
./nmsmods set-path "/path/to/No Man's Sky"
./nmsmods where
```

2) Download a mod ZIP (direct link)

```bash
./nmsmods download "https://supporter-files.nexus-cdn.com/...zip?..."
```

3) List downloaded mods

```bash
./nmsmods downloads
```

4) Install

```bash
./nmsmods install <id>
```

5) List installed

```bash
./nmsmods installed
```

6) Uninstall

```bash
./nmsmods uninstall <id-or-folder>
```

7) Remove a downloaded ZIP (keeps installed files)

```bash
./nmsmods rm-download <id>
```

## Data directory

All tool data is stored in:

- `~/.nmsmods/`
  - `config.json`
  - `state.json`
  - `downloads/`
  - `staging/`

## Notes

- Manual mod install location for modern NMS versions is `<NMS>/GAMEDATA/MODS`.
- This personal v0.1 does **not** edit `GCMODSETTINGS.MXML`; it only places/removes mod folders.
- ZIP extraction heuristic:
  - If the extracted archive contains a single top-level directory, that directory is installed.
  - Otherwise, the whole extracted root is installed as the mod folder (named by the mod id).

