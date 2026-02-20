#!/usr/bin/env bash
# install.sh - installs nmsmods for Linux.
# Strategy:
# 1) Prefer GitHub Releases binaries (no Go needed).
# 2) If that fails, clone repo to temp and build from source (Go needed).
set -euo pipefail

BIN_NAME="nmsmods"
DEFAULT_PREFIX="$HOME/.local/bin"
PREFIX="$DEFAULT_PREFIX"

DEFAULT_REPO="guarinogio/nmsmods"

usage() {
  cat <<EOF
Usage: install.sh [--prefix <dir>]

Environment overrides:
  NMSMODS_REPO   GitHub repo in owner/name form (default: ${DEFAULT_REPO})
EOF
}

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
  usage
  exit 0
fi

if [[ "${1:-}" == "--prefix" ]]; then
  PREFIX="${2:-$DEFAULT_PREFIX}"
fi

need_cmd() { command -v "$1" >/dev/null 2>&1; }
say() { echo "==> $*"; }
die() { echo "ERROR: $*" >&2; exit 1; }

detect_arch() {
  local arch
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) die "Unsupported architecture: $arch (supported: amd64, arm64)" ;;
  esac
}

install_packages() {
  # $@ = packages to install
  if need_cmd apt-get; then
    sudo apt-get update
    sudo apt-get install -y "$@"
    return 0
  fi
  if need_cmd dnf; then
    sudo dnf install -y "$@"
    return 0
  fi
  if need_cmd pacman; then
    sudo pacman -Syu --noconfirm --needed "$@"
    return 0
  fi
  if need_cmd zypper; then
    sudo zypper --non-interactive in "$@"
    return 0
  fi
  return 1
}

ensure_tools() {
  # curl, tar, git, jq
  local missing=()

  need_cmd curl || missing+=("curl")
  need_cmd tar  || missing+=("tar")
  need_cmd git  || missing+=("git")
  need_cmd jq   || missing+=("jq")
  need_cmd xdg-mime || missing+=("xdg-mime")
  need_cmd ca-certificates >/dev/null 2>&1 || true

  if [[ ${#missing[@]} -eq 0 ]]; then
    return 0
  fi

  say "Missing tools: ${missing[*]}"
  say "Attempting to install missing tools via package manager (best effort)..."

  # map tool->package names (approx; good enough for major distros)
  # Most distros use same names for curl/tar/git/jq/xdg-utils.
  if ! install_packages ca-certificates curl tar git jq xdg-utils; then
    die "Could not install required tools automatically. Please install: curl tar git jq xdg-utils"
  fi

  need_cmd curl || die "curl still missing after install"
  need_cmd tar  || die "tar still missing after install"
  need_cmd git  || die "git still missing after install"
  need_cmd jq   || die "jq still missing after install"
  need_cmd xdg-mime || die "xdg-mime still missing after install (install xdg-utils)"
}

ensure_go() {
  if need_cmd go; then
    return 0
  fi
  say "Go not found. Attempting to install Go via package manager (best effort)..."
  if ! install_packages golang-go golang go; then
    # Not all package managers accept multiple names; try common ones per manager.
    if need_cmd apt-get; then
      sudo apt-get update
      sudo apt-get install -y golang-go
    elif need_cmd dnf; then
      sudo dnf install -y golang
    elif need_cmd pacman; then
      sudo pacman -Syu --noconfirm --needed go
    elif need_cmd zypper; then
      sudo zypper --non-interactive in go
    else
      die "No supported package manager found to install Go. Install Go manually and re-run."
    fi
  fi
  need_cmd go || die "Go installation failed."
}

download_latest_release_bin() {
  local repo="${1:?repo required}"
  local arch os api tmpdir tag asset_url asset_name checksums_url

  os="linux"
  arch="$(detect_arch)"
  api="https://api.github.com/repos/${repo}/releases/latest"

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' RETURN

  say "Checking latest release: ${repo}"
  curl -fsSL "$api" -o "$tmpdir/release.json" || return 1

  tag="$(jq -r .tag_name "$tmpdir/release.json")"
  [[ -n "$tag" && "$tag" != "null" ]] || return 1

  # Match GoReleaser asset name: nmsmods_<version>_linux_<arch>.tar.gz
  asset_name="$(jq -r --arg os "$os" --arg arch "$arch" '
    .assets[]
    | select(.name | test("^nmsmods_.*_" + $os + "_" + $arch + "\\.tar\\.gz$"))
    | .name
  ' "$tmpdir/release.json" | head -n 1)"

  asset_url="$(jq -r --arg name "$asset_name" '
    .assets[]
    | select(.name == $name)
    | .browser_download_url
  ' "$tmpdir/release.json" | head -n 1)"

  [[ -n "$asset_name" && "$asset_name" != "null" ]] || return 1
  [[ -n "$asset_url" && "$asset_url" != "null" ]] || return 1

  checksums_url="$(jq -r '
    .assets[]
    | select(.name == "checksums.txt")
    | .browser_download_url
  ' "$tmpdir/release.json" | head -n 1)"

  [[ -n "$checksums_url" && "$checksums_url" != "null" ]] || return 1

  say "Downloading release asset: ${asset_url}"
  curl -fsSL "$asset_url" -o "$tmpdir/$asset_name"

  say "Downloading checksums: ${checksums_url}"
  curl -fsSL "$checksums_url" -o "$tmpdir/checksums.txt"

  # Verify sha256 from checksums.txt
  local expected actual
  expected="$(grep -E "  ${asset_name}$" "$tmpdir/checksums.txt" | awk '{print $1}' | head -n 1)"
  [[ -n "$expected" ]] || die "Could not find ${asset_name} in checksums.txt"

  actual="$(sha256sum "$tmpdir/$asset_name" | awk '{print $1}')"
  if [[ "$expected" != "$actual" ]]; then
    die "Checksum mismatch for ${asset_name} (expected $expected, got $actual)"
  fi
  say "Checksum verified: ${asset_name}"

  mkdir -p "$tmpdir/extract"
  tar -C "$tmpdir/extract" -xzf "$tmpdir/$asset_name"

  # binary might be nested; find it
  local found
  found="$(find "$tmpdir/extract" -type f -name "${BIN_NAME}" -perm -111 | head -n 1 || true)"
  [[ -n "$found" ]] || return 1

  mkdir -p "$PREFIX"
  install -m 0755 "$found" "$PREFIX/${BIN_NAME}"
  say "Installed ${BIN_NAME} (${tag}) to ${PREFIX}/${BIN_NAME}"
  return 0
}


build_from_source_in_temp() {
  local repo="${1:?repo required}"
  local tmpdir

  ensure_go

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' RETURN

  say "Cloning repo to temp: ${repo}"
  git clone --depth 1 "https://github.com/${repo}.git" "$tmpdir/src"

  say "Building from source..."
  ( cd "$tmpdir/src" && go mod tidy && go build -o "${BIN_NAME}" ./ )

  mkdir -p "$PREFIX"
  install -m 0755 "$tmpdir/src/${BIN_NAME}" "$PREFIX/${BIN_NAME}"
  say "Installed ${BIN_NAME} (built from source) to ${PREFIX}/${BIN_NAME}"
}

ensure_path_in_profile() {
  # Ensure PREFIX is in PATH for future shells, in a shell-agnostic way.
  # We avoid guessing bash/zsh/fish here; ~/.profile is the portable login shell entrypoint.
  local prof="$HOME/.profile"
  local snippet_begin="# nmsmods: ensure user-local bin dir is in PATH"
  local snippet_end="# end nmsmods"
  local line="export PATH=\"$PREFIX:\$PATH\""

  # Only apply when installing to a user prefix under $HOME.
  case "$PREFIX" in
    "$HOME"/*) ;;
    *) return 0 ;;
  esac

  # If already in PATH now, don't force changes (but still OK to add for future shells).
  if echo ":$PATH:" | grep -q ":$PREFIX:"; then
    return 0
  fi

  say "Adding $PREFIX to PATH in $prof (portable)"
  touch "$prof"
  if grep -q "$snippet_begin" "$prof" 2>/dev/null; then
    return 0
  fi

  {
    echo
    echo "$snippet_begin"
    echo "if [ -d \"$PREFIX\" ] && ! echo \"\$PATH\" | grep -q \"$PREFIX\" ; then"
    echo "  $line"
    echo "fi"
    echo "$snippet_end"
  } >> "$prof"
}

register_nxm_handler() {
  local bin_path="$PREFIX/$BIN_NAME"
  local desktop_dir="$HOME/.local/share/applications"
  local desktop_file="$desktop_dir/nmsmods.desktop"

  mkdir -p "$desktop_dir"

  cat > "$desktop_file" <<EOF
[Desktop Entry]
Type=Application
Name=nmsmods
Exec=${bin_path} nxm handle %u
Terminal=false
NoDisplay=true
MimeType=x-scheme-handler/nxm;
Categories=Game;
EOF

  say "Registering nxm:// handler (x-scheme-handler/nxm)"
  xdg-mime default nmsmods.desktop x-scheme-handler/nxm || true
}

main() {
  ensure_tools

  local repo="${NMSMODS_REPO:-$DEFAULT_REPO}"

  # Preferred path: install from latest release
  if download_latest_release_bin "$repo"; then
    register_nxm_handler
    ensure_path_in_profile

    echo
    echo "Next steps:"
    echo "  1) Open a new terminal (or log out/in) so PATH updates take effect"
    echo "  2) Run: nmsmods set-path --auto   (or set-path <path>)"
    echo "  3) Run: nmsmods nexus login --api-key <key>"
    echo "  4) In Nexus, click 'Mod Manager Download' to auto-install"
    exit 0
  fi

  # Fallback: build from source in temp clone
  say "Release install failed (no matching asset or API issue). Falling back to source build..."
  build_from_source_in_temp "$repo"

  register_nxm_handler
  ensure_path_in_profile

  echo
  echo "Next steps:"
  echo "  1) Open a new terminal (or log out/in) so PATH updates take effect"
  echo "  2) Run: nmsmods set-path --auto   (or set-path <path>)"
  echo "  3) Run: nmsmods nexus login --api-key <key>"
  echo "  4) In Nexus, click 'Mod Manager Download' to auto-install"
}

main "$@"
