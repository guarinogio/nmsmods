# nmsmods

A minimal, fast CLI mod manager for **No Man's Sky (Linux)**.

`nmsmods` helps you:

-   Detect your No Man's Sky installation
-   Download mod ZIP files
-   Install mods into `GAMEDATA/MODS`
-   List installed mods
-   Uninstall mods cleanly
-   Track downloaded files

It is designed to be:

-   Simple
-   Transparent
-   Scriptable
-   Extensible (future Nexus Mods integration planned)

------------------------------------------------------------------------

# Important Context (No Man's Sky 5.58+ / 6.x)

No Man's Sky modding changed significantly after 5.50 and again after
5.58.

Key points:

-   Mods must go in:

    `<GameDir>/GAMEDATA/MODS`

    NOT:

    `PCBANKS/MODS`

-   `.EXML` and `.MBIN` files are valid.

-   `.MXML` files (except `LocTable.MXML`) are ignored by the game.

-   Old PAK-based mods generally do not work anymore.

Always ensure the mod you download:

-   Is updated for your game version
-   Was updated after January 2025 (for 5.50+ compatibility)

------------------------------------------------------------------------

# Supported Platforms

Currently supported:

-   Linux (native Steam install)
-   Steam under Proton

Supported architectures:

-   amd64 (x86_64)
-   arm64 (aarch64)

------------------------------------------------------------------------

# Installation

## Recommended (No Go required)

Install using the official installer:

``` bash
curl -fsSL https://raw.githubusercontent.com/guarinogio/nmsmods/main/install.sh | bash
```

This will:

1.  Detect your CPU architecture
2.  Download the latest GitHub Release binary
3.  Install it to:

`~/.local/bin/nmsmods`

Then run:

``` bash
nmsmods doctor
```

------------------------------------------------------------------------

## Alternative: Build from Source

Requirements:

-   Go 1.22+
-   Git

Clone and build:

``` bash
git clone https://github.com/guarinogio/nmsmods.git
cd nmsmods
make install
```

Or manually:

``` bash
go mod tidy
go build -o nmsmods ./
mkdir -p ~/.local/bin
install -m 0755 nmsmods ~/.local/bin/nmsmods
```

------------------------------------------------------------------------

# Verifying Installation

Check:

``` bash
which nmsmods
nmsmods doctor
```

`doctor` will show:

-   Data directory
-   Downloads directory
-   Staging directory
-   Detected game path
-   Mods directory
-   Installed mods
-   Tracked downloads

------------------------------------------------------------------------

# How It Works

`nmsmods` uses:

`~/.nmsmods/`

Structure:

    ~/.nmsmods/
     ├── config.json
     ├── state.json
     ├── downloads/
     └── staging/

`state.json` tracks:

-   Downloaded ZIP files
-   Whether installed
-   Install folder name
-   Original URL

This ensures:

-   Clean uninstall
-   Reproducibility
-   No guessing folder names

------------------------------------------------------------------------

# Basic Usage

## Detect Game

``` bash
nmsmods doctor
```

If auto-detection fails:

``` bash
nmsmods set-path "/path/to/No Man's Sky"
```

------------------------------------------------------------------------

## Download a Mod

You must provide a direct ZIP URL:

``` bash
nmsmods download "https://example.com/mod.zip"
```

List downloaded mods:

``` bash
nmsmods downloads
```

------------------------------------------------------------------------

## Install a Mod

Install using numeric index:

``` bash
nmsmods install 1
```

Or by ID:

``` bash
nmsmods install refinerwikislots
```

------------------------------------------------------------------------

## List Installed Mods

``` bash
nmsmods installed
```

------------------------------------------------------------------------

## Uninstall

``` bash
nmsmods uninstall 1
```

------------------------------------------------------------------------

## Remove Downloaded ZIP

``` bash
nmsmods rm-download 1
```

------------------------------------------------------------------------

# Safety Considerations

This tool extracts ZIP files into a staging directory before copying
them into the game folder.

You should:

-   Only download mods from trusted sources
-   Ensure mod version matches your game version
-   Keep backups of save data

Future versions will include:

-   Hardened ZIP extraction
-   Path traversal protection
-   File size limits
-   Validation of expected file structure

------------------------------------------------------------------------

# Steam Auto-Detection

On Linux, the tool attempts to detect:

`~/.local/share/Steam/steamapps/common/No Man's Sky`

If you use a custom Steam library path, set it manually:

``` bash
nmsmods set-path "/your/custom/path"
```

------------------------------------------------------------------------

# Uninstalling nmsmods

If installed via installer:

``` bash
rm ~/.local/bin/nmsmods
```

Your mod state is stored in:

`~/.nmsmods`

Delete it if you want to reset everything:

``` bash
rm -rf ~/.nmsmods
```

------------------------------------------------------------------------

# Roadmap

Planned:

-   Hardened ZIP extraction (security upgrade)
-   Nexus Mods API integration
-   JSON output mode
-   Mod compatibility checks
-   Mod priority management
-   Update command
-   Signature / checksum verification

------------------------------------------------------------------------

# License

MIT License.