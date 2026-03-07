#!/bin/sh
set -e

git config --global --add safe.directory /workspaces/cartage

make build

for name in notify-send yad zenity kdialog xdg-open pbcopy pbpaste; do
  sudo ln -sf "$(pwd)/dist/cartage" "/usr/local/bin/$name"
done
