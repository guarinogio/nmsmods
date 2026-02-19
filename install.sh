#!/usr/bin/env bash
# install.sh
set -euo pipefail

BIN_NAME="nmsmods"
DEFAULT_PREFIX="$HOME/.local/bin"
PREFIX="$DEFAULT_PREFIX"

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

detect_repo() {
  # tries to infer owner/repo from git remote origin
  local url owner repo
  url="$(git config --get remote.origin.url 2>/dev/null || true)"
  if [[ -z "$url" ]]; then
    echo ""
    return 0
  fi

  # handle:
  #  - https://github.com/owner/repo.git
  #  - git@github.com:owner/repo.git
  url="${url%.git}"
  if [[ "$url" =~ github\.com[:/]+([^/]+)/([^/]+)$ ]]; then
    owner="${BASH_REMATCH[1]}"
    repo="${BASH_REMATCH[2]}"
    echo "${owner}/${repo}"
    return 0
  fi

  echo ""
}

ensure_basic_tools() {
  if need_cmd curl && need_cmd tar; then
    return 0
  fi

  say "Installing base tools (curl, tar) via package manager (best effort)..."
  if need_cmd apt-get; then
    sudo apt-get update
    sudo apt-get install -y ca-certificates curl tar
    return 0
  fi
  if need_cmd dnf; then
    sudo dnf install -y ca-certificates curl tar
    return 0
  fi
  if need_cmd pacman; then
    sudo pacman -Syu --noconfirm --needed ca-certificates curl tar
    return 0
  fi
  if need_cmd zypper; then
    sudo zypper --non-interactive in ca-certificates curl tar
    return 0
  fi

  die "Missing curl/tar and no supported package manager found."
}

download_latest_release_bin() {
  local repo="${1:-}"
  local arch os api tmpdir tag asset_url

  os="linux"
  arch="$(detect_arch)"

  [[ -n "$repo" ]] || return 1
  need_cmd curl || return 1
  need_cmd tar || return 1

  # Need jq for reliable GitHub API parsing. If missing, skip and fallback to source build.
  if ! need_cmd jq; then
    say "jq not found; skipping release download and falling back to source build."
    return 1
  fi

  api="https://api.github.com/repos/${repo}/releases/latest"
  say "Checking latest release: ${repo}"
  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' RETURN

  curl -fsSL "$api" -o "$tmpdir/release.json" || return 1

  tag="$(jq -r .tag_name "$tmpdir/release.json")"
  if [[ -z "$tag" || "$tag" == "null" ]]; then
    return 1
  fi

  # We expect archive name template: nmsmods_<version>_linux_<arch>.tar.gz
  asset_url="$(jq -r --arg os "$os" --arg arch "$arch" '
    .assets[]
    | select(.name | test("^nmsmods_.*_" + $os + "_" + $arch + "\\.tar\\.gz$"))
    | .browser_download_url
  ' "$tmpdir/release.json" | head -n 1)"

  if [[ -z "$asset_url" || "$asset_url" == "null" ]]; then
    say "No matching release asset found for ${os}/${arch}."
    return 1
  fi

  say "Downloading release asset: ${asset_url}"
  curl -fsSL "$asset_url" -o "$tmpdir/nmsmods.tar.gz"

  mkdir -p "$tmpdir/extract"
  tar -C "$tmpdir/extract" -xzf "$tmpdir/nmsmods.tar.gz"

  if [[ ! -f "$tmpdir/extract/${BIN_NAME}" ]]; then
    # some archives include nested dirs; search for binary
    local found
    found="$(find "$tmpdir/extract" -type f -name "${BIN_NAME}" -perm -111 | head -n 1 || true)"
    [[ -n "$found" ]] || return 1
    cp "$found" "$tmpdir/extract/${BIN_NAME}"
  fi

  mkdir -p "$PREFIX"
  install -m 0755 "$tmpdir/extract/${BIN_NAME}" "$PREFIX/${BIN_NAME}"
  say "Installed ${BIN_NAME} ${tag} to ${PREFIX}/${BIN_NAME}"
  return 0
}

ensure_go_if_building() {
  if need_cmd go; then
    return 0
  fi

  say "Go not found. Attempting to install Go via package manager (best effort)..."
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
    die "Go is required to build from source, and no supported package manager found."
  fi

  need_cmd go || die "Go installation failed."
}

build_from_source() {
  [[ -f "go.mod" ]] || die "Run from repo root (go.mod not found)."

  ensure_go_if_building

  say "Tidying modules..."
  go mod tidy

  say "Running tests..."
  go test ./...

  say "Building ${BIN_NAME}..."
  go build -o "${BIN_NAME}" ./

  say "Installing to ${PREFIX}..."
  mkdir -p "$PREFIX"
  install -m 0755 "${BIN_NAME}" "${PREFIX}/${BIN_NAME}"

  say "Installed to ${PREFIX}/${BIN_NAME}"
}

main() {
  ensure_basic_tools

  # Prefer release binary install (no Go required).
  local repo
  repo="$(detect_repo)"

  if [[ -n "${NMSMODS_REPO:-}" ]]; then
    repo="${NMSMODS_REPO}"
  fi

  if download_latest_release_bin "$repo"; then
    echo
    echo "Run:"
    echo "  ${PREFIX}/${BIN_NAME} doctor"
    exit 0
  fi

  say "Falling back to building from source..."
  build_from_source

  echo
  echo "Run:"
  echo "  ${PREFIX}/${BIN_NAME} doctor"
}

main "$@"
