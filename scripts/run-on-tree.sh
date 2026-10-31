#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: $0 {index|head} <command> [args...]" >&2
  exit 2
}

[ "$#" -ge 2 ] || usage
source_tree="$1"
shift

case "$(uname -s)" in
  MINGW* | MSYS* | CYGWIN*) windows=1 ;;
  *) windows=0 ;;
esac

repo_root="$(git rev-parse --show-toplevel)"
snapshot="$(mktemp -d "${TMPDIR:-/tmp}/ucrp-tree.XXXXXX")"
links=()

link_directory() {
  local target="$1" link="$2"
  if [ "$windows" -eq 1 ]; then
    MSYS_NO_PATHCONV=1 cmd //c mklink //J "$(cygpath -w "$link")" "$(cygpath -w "$target")" >/dev/null
  else
    ln -s "$target" "$link"
  fi
  links+=("$link")
}

unlink_directory() {
  if [ "$windows" -eq 1 ]; then
    MSYS_NO_PATHCONV=1 cmd //c rmdir "$(cygpath -w "$1")" >/dev/null
  else
    rm "$1"
  fi
}

cleanup() {
  local link
  for link in ${links[@]+"${links[@]}"}; do
    if ! unlink_directory "$link"; then
      echo "run-on-tree: could not remove link $link, leaving $snapshot in place" >&2
      return
    fi
  done
  rm -rf "$snapshot"
}
trap cleanup EXIT

case "$source_tree" in
  index) git -C "$repo_root" checkout-index --all --prefix="$snapshot/" ;;
  head) git -C "$repo_root" archive HEAD | tar -x -C "$snapshot" ;;
  *) usage ;;
esac

for modules in "$repo_root/node_modules" "$repo_root"/apps/*/node_modules "$repo_root"/packages/*/node_modules; do
  [ -d "$modules" ] || continue
  link_directory "$modules" "$snapshot/${modules#"$repo_root"/}"
done

cd "$snapshot"
"$@"
