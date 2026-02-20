#!/bin/sh
set -eu

# Register nxm:// handler system-wide (best effort)
if command -v xdg-mime >/dev/null 2>&1; then
  xdg-mime default nmsmods.desktop x-scheme-handler/nxm || true
fi

# Refresh desktop database (best effort)
if command -v update-desktop-database >/dev/null 2>&1; then
  update-desktop-database /usr/share/applications >/dev/null 2>&1 || true
fi

exit 0