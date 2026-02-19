#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

need_cmd() { command -v "$1" >/dev/null 2>&1 || { echo "Missing: $1" >&2; exit 1; }; }
need_cmd go
need_cmd python3

say() { printf "\n==> %s\n" "$*"; }

say "Build"
go build -o ./nmsmods ./
BIN="$ROOT/nmsmods"
[[ -x "$BIN" ]] || { echo "Binary not built: $BIN" >&2; exit 1; }

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

export NMSMODS_HOME="$TMP/nmsmods_home"
GAME_DIR="$TMP/NMS"
MODSRC_OK="$TMP/modsrc_ok"
MODSRC_BAD="$TMP/modsrc_bad"
ZIP_OK="$TMP/mod_ok.zip"
ZIP_BAD="$TMP/mod_bad.zip"

REPORT="$ROOT/local_e2e_report.md"
: > "$REPORT"

append() { printf "%s\n" "$*" >> "$REPORT"; }

run_cmd() {
  # args: name, command...
  local name="$1"; shift
  local out="$TMP/${name}.out"
  local err="$TMP/${name}.err"
  local cmdline=("$@")

  append "### \`${cmdline[*]}\`"
  append ""
  append "**Exit:** (captured below)"
  append ""

  set +e
  "${cmdline[@]}" >"$out" 2>"$err"
  local rc=$?
  set -e

  append "- exit_code: \`$rc\`"
  append ""
  append "<details><summary>stdout</summary>"
  append ""
  append '```'
  sed 's/\r$//' "$out" >> "$REPORT" || true
  append '```'
  append "</details>"
  append ""
  append "<details><summary>stderr</summary>"
  append ""
  append '```'
  sed 's/\r$//' "$err" >> "$REPORT" || true
  append '```'
  append "</details>"
  append ""
  append "---"
  append ""

  return $rc
}

say "Setup fake NMS install"
mkdir -p "$GAME_DIR/GAMEDATA/PCBANKS" "$GAME_DIR/Binaries"
echo "pak" > "$GAME_DIR/GAMEDATA/PCBANKS/dummy.pak"

say "Create dummy mods"
mkdir -p "$MODSRC_OK/Outer/Inner"
cat > "$MODSRC_OK/Outer/Inner/TEST.EXML" <<'EOF'
<?xml version="1.0" encoding="utf-8"?>
<Data template="GcSomething">
  <Property name="Dummy" value="true" />
</Data>
EOF

mkdir -p "$MODSRC_BAD/Outer/Inner"
echo "hello" > "$MODSRC_BAD/Outer/Inner/readme.txt"

python3 - <<PY
import zipfile, pathlib
def zip_dir(src, outzip):
    src = pathlib.Path(src)
    with zipfile.ZipFile(outzip, "w", compression=zipfile.ZIP_DEFLATED) as z:
        for p in src.rglob("*"):
            if p.is_dir(): 
                continue
            z.write(p, p.relative_to(src).as_posix())
zip_dir("$MODSRC_OK", "$ZIP_OK")
zip_dir("$MODSRC_BAD", "$ZIP_BAD")
print("ok zip:", "$ZIP_OK")
print("bad zip:", "$ZIP_BAD")
PY

append "# nmsmods Local E2E Report"
append ""
append "- date: $(date -Is)"
append "- repo: $(git rev-parse --show-toplevel 2>/dev/null || echo "$ROOT")"
append "- commit: $(git rev-parse HEAD 2>/dev/null || echo "unknown")"
append "- nmsmods: $("$BIN" --version 2>/dev/null || true)"
append "- NMSMODS_HOME: $NMSMODS_HOME"
append "- fake_game_path: $GAME_DIR"
append ""
append "## Command inventory (from \`nmsmods help\`)"
append ""
append '```'
"$BIN" help >> "$REPORT" || true
append '```'
append ""
append "---"
append ""

say "Discover commands"
# Parse top-level commands from `help` output (simple heuristic)
CMDS=()
while IFS= read -r line; do
  # lines like: "  doctor      ..."
  if [[ "$line" =~ ^[[:space:]]{2}([a-zA-Z0-9_-]+)[[:space:]]{2,} ]]; then
    c="${BASH_REMATCH[1]}"
    # exclude "help"
    if [[ "$c" != "help" ]]; then
      CMDS+=("$c")
    fi
  fi
done < <("$BIN" help 2>/dev/null | sed -n '/^Available Commands:/,/^Flags:/p')

# de-dup preserving order
DEDUP=()
for c in "${CMDS[@]}"; do
  skip=0
  for d in "${DEDUP[@]}"; do [[ "$d" == "$c" ]] && skip=1; done
  [[ $skip -eq 0 ]] && DEDUP+=("$c")
done
CMDS=("${DEDUP[@]}")

append "## Executed commands"
append ""
append "- commands_detected: ${#CMDS[@]}"
append "- list: \`${CMDS[*]}\`"
append ""
append "---"
append ""

say "Run baseline commands"
run_cmd "version_flag" "$BIN" --version || true
run_cmd "version_cmd" "$BIN" version || true
run_cmd "help" "$BIN" help || true
run_cmd "doctor_before_setpath" "$BIN" doctor --json || true

say "Set path"
run_cmd "setpath" "$BIN" set-path "$GAME_DIR" || true
run_cmd "doctor_after_setpath" "$BIN" doctor --json || true
run_cmd "where" "$BIN" where || true
run_cmd "where_json" "$BIN" where --json || true 2>/dev/null || true

say "Seed downloads and state"
# download ok/bad
run_cmd "download_ok" "$BIN" download "$ZIP_OK" --id testmod || true
run_cmd "download_bad" "$BIN" download "$ZIP_BAD" --id badmod || true
run_cmd "downloads" "$BIN" downloads || true
run_cmd "downloads_json" "$BIN" downloads --json || true

# info on index 1 if possible
run_cmd "info_1" "$BIN" info 1 || true
run_cmd "info_1_json" "$BIN" info 1 --json || true

say "Install flows"
run_cmd "install_1" "$BIN" install 1 || true
run_cmd "installed" "$BIN" installed || true
run_cmd "installed_json" "$BIN" installed --json || true
run_cmd "verify_1" "$BIN" verify 1 || true
run_cmd "verify_1_json" "$BIN" verify 1 --json || true

# overwrite/no-overwrite semantics (if command exists)
run_cmd "install_1_no_overwrite" "$BIN" install 1 --no-overwrite || true
run_cmd "reinstall_1" "$BIN" reinstall 1 || true
run_cmd "uninstall_1" "$BIN" uninstall 1 || true

say "Install-dir flow (collision test)"
# First, install badmod and keep it installed to force a folder collision on next install-dir
run_cmd "install_badmod_keep" "$BIN" install badmod || true
run_cmd "installed_after_badmod" "$BIN" installed || true

# Now install-dir should avoid clobbering Inner (used by badmod) and choose Inner__dirmod
run_cmd "install_dir_collision" "$BIN" install-dir "$MODSRC_OK" --id dirmod || true
run_cmd "installed_after_dirmod" "$BIN" installed || true

# Verify both
run_cmd "verify_badmod_json" "$BIN" verify badmod --json || true
run_cmd "verify_dirmod" "$BIN" verify dirmod --json || true

# Cleanup
run_cmd "uninstall_dirmod" "$BIN" uninstall dirmod || true
run_cmd "uninstall_badmod_keep" "$BIN" uninstall badmod || true


say "Bad mod health flow"
run_cmd "install_badmod" "$BIN" install badmod || true
run_cmd "verify_badmod" "$BIN" verify badmod --json || true
run_cmd "uninstall_badmod" "$BIN" uninstall badmod || true

say "Remove downloads / clean / reset"
run_cmd "rm_download_1" "$BIN" rm-download 1 || true
run_cmd "clean_dry" "$BIN" clean --staging --parts --orphan-zips --dry-run || true
run_cmd "clean" "$BIN" clean --staging --parts --orphan-zips || true
run_cmd "reset_dry" "$BIN" reset --dry-run || true
run_cmd "reset" "$BIN" reset || true

say "Try any remaining detected commands (best-effort)"
# For any commands not explicitly called above, call `cmd --help` so report shows it exists.
CALLED_SET=$(printf "%s\n" \
  version help doctor set-path where download downloads info install installed verify reinstall uninstall install-dir rm-download clean reset \
  | sort -u)

for c in "${CMDS[@]}"; do
  if echo "$CALLED_SET" | grep -qx "$c"; then
    continue
  fi
  # best-effort: show help for unknown commands (so we still "exercise" them)
  run_cmd "help_${c}" "$BIN" "$c" --help || true
done

append "## Summary"
append ""
append "- If some commands show non-zero exit codes above, that is OK **only** if they were intentionally run in an invalid state (e.g. doctor before set-path)."
append "- Paste this report back into ChatGPT and I can validate failures vs expected behavior."
append ""

say "Done"
echo "Report generated: $REPORT"
echo "Temp workspace:   $TMP"
echo
echo "Paste this into chat:"
echo "  cat $REPORT"
