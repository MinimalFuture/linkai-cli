#!/bin/sh
# LinkAI CLI installer — downloads the pre-built `linkai` binary and (optionally)
# installs the agent skill. No Go / Node.js required.
#
# Quick install:
#   curl -fsSL https://cdn.link-ai.tech/cli/install.sh | sh
#   # or from GitHub:
#   curl -fsSL https://raw.githubusercontent.com/MinimalFuture/linkai-cli/main/install.sh | sh
#
# Environment overrides (all optional):
#   LINKAI_VERSION      version to install                 (default: latest)
#   LINKAI_INSTALL_DIR  where to put the binary             (default: smart — see below)
#   LINKAI_SOURCE       download source: cdn | github       (default: cdn, GitHub fallback)
#   LINKAI_NO_SKILL     set to 1 to skip installing the agent skill
#   LINKAI_SKILL_ONLY   set to 1 to install only the skill (skip the binary)
#   LINKAI_NO_MODIFY_PATH set to 1 to never edit shell rc files (only print a hint)
#
# Install directory resolution (smart, in order):
#   1. $LINKAI_INSTALL_DIR if set
#   2. /usr/local/bin if it exists and is writable (usually already on PATH)
#   3. $HOME/.local/bin otherwise (and ensure it is on PATH — see below)

set -eu

REPO="MinimalFuture/linkai-cli"
BINARY="linkai"
SKILL_NAME="linkai-cli"

CDN_BASE="https://cdn.link-ai.tech/cli"
GITHUB_BASE="https://github.com/${REPO}/releases/download"

VERSION="${LINKAI_VERSION:-latest}"
SOURCE="${LINKAI_SOURCE:-cdn}"
NO_SKILL="${LINKAI_NO_SKILL:-0}"
SKILL_ONLY="${LINKAI_SKILL_ONLY:-0}"
NO_MODIFY_PATH="${LINKAI_NO_MODIFY_PATH:-0}"

# ── Output helpers ────────────────────────────────────────────────────────────

err()  { printf '  \033[31m✗\033[0m %s\n' "$*" >&2; exit 1; }
info() { printf '  %s\n' "$*" >&2; }
ok()   { printf '  \033[32m✓\033[0m %s\n' "$*" >&2; }

need() { command -v "$1" >/dev/null 2>&1 || err "missing required command: $1"; }

need uname
need tar
need mkdir

if command -v curl >/dev/null 2>&1; then
  fetch()    { curl -fsSL "$1"; }
  fetch_to() { curl -fsSL -o "$2" "$1"; }
  probe()    { curl -fsS --connect-timeout 5 --max-time 12 -o /dev/null "$1" >/dev/null 2>&1; }
elif command -v wget >/dev/null 2>&1; then
  fetch()    { wget -qO- "$1"; }
  fetch_to() { wget -qO "$2" "$1"; }
  probe()    { wget -q --timeout=12 --tries=1 -O /dev/null "$1" >/dev/null 2>&1; }
else
  err "either curl or wget is required"
fi

# ── Platform detection ────────────────────────────────────────────────────────

detect_os() {
  case "$(uname -s)" in
    Linux)  echo linux ;;
    Darwin) echo darwin ;;
    MINGW*|MSYS*|CYGWIN*) err "on Windows use the PowerShell installer (install.ps1)" ;;
    *) err "unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo amd64 ;;
    arm64|aarch64) echo arm64 ;;
    *) err "unsupported architecture: $(uname -m)" ;;
  esac
}

# ── Source selection (CDN primary, GitHub fallback) ───────────────────────────
#
# An explicit LINKAI_SOURCE always wins. Otherwise default to the CDN and, if it
# is unreachable, fall back to GitHub Releases so the one-liner works everywhere.
pick_source() {
  case "$SOURCE" in
    github) return 0 ;;
    cdn)
      if probe "${CDN_BASE}/install.sh" || probe "${CDN_BASE}/"; then
        return 0
      fi
      info "⚠ CDN unreachable, falling back to GitHub Releases"
      SOURCE="github"
      ;;
    *) err "invalid LINKAI_SOURCE='$SOURCE' (use 'cdn' or 'github')" ;;
  esac
}

# Resolve the latest version tag.
#   CDN   : read cli/latest.txt (a plain text file holding e.g. "v1.2.3").
#   GitHub: follow the /releases/latest redirect (avoids API rate limits).
resolve_version() {
  [ "$VERSION" != "latest" ] && return 0
  if [ "$SOURCE" = "cdn" ]; then
    VERSION="$(fetch "${CDN_BASE}/latest.txt" 2>/dev/null | tr -d ' \t\r\n')"
  fi
  if [ -z "$VERSION" ] || [ "$VERSION" = "latest" ]; then
    # GitHub redirect trick (works for CDN fallback too).
    if command -v curl >/dev/null 2>&1; then
      VERSION="$(curl -fsSI "https://github.com/${REPO}/releases/latest" 2>/dev/null \
        | grep -i '^location:' | sed 's|.*/tag/||;s/[[:space:]]*$//')"
    fi
  fi
  [ -n "$VERSION" ] && [ "$VERSION" != "latest" ] \
    || err "could not resolve the latest version — set LINKAI_VERSION=vX.Y.Z"
}

# Build the download URL for a release asset by file name.
asset_url() {
  _name="$1"
  if [ "$SOURCE" = "cdn" ]; then
    printf '%s' "${CDN_BASE}/${VERSION}/${_name}"
  else
    printf '%s' "${GITHUB_BASE}/${VERSION}/${_name}"
  fi
}

# ── Install directory + PATH handling ─────────────────────────────────────────
#
# Prefer a directory already on PATH so the user never has to touch a shell rc.
resolve_install_dir() {
  if [ -n "${LINKAI_INSTALL_DIR:-}" ]; then
    echo "$LINKAI_INSTALL_DIR"
    return 0
  fi
  if [ -d /usr/local/bin ] && [ -w /usr/local/bin ]; then
    echo /usr/local/bin
    return 0
  fi
  echo "$HOME/.local/bin"
}

# Detect the shell rc file to append a PATH export to.
detect_shell_rc() {
  _shell="$(basename "${SHELL:-}")"
  case "$_shell" in
    zsh)  echo "$HOME/.zshrc" ;;
    bash)
      # macOS bash reads .bash_profile for login shells; Linux reads .bashrc.
      if [ "$(uname -s)" = "Darwin" ] && [ -f "$HOME/.bash_profile" ]; then
        echo "$HOME/.bash_profile"
      else
        echo "$HOME/.bashrc"
      fi
      ;;
    fish) echo "$HOME/.config/fish/config.fish" ;;
    *)    echo "$HOME/.profile" ;;
  esac
}

# Append a PATH export to the shell rc (idempotent). Fish uses a different syntax.
append_path_to_rc() {
  _dir="$1"
  _rc="$(detect_shell_rc)"
  mkdir -p "$(dirname "$_rc")"
  case "$_rc" in
    *config.fish) _line="fish_add_path $_dir" ;;
    *)            _line="export PATH=\"$_dir:\$PATH\"" ;;
  esac
  if [ -f "$_rc" ] && grep -Fq "$_dir" "$_rc" 2>/dev/null; then
    return 0  # already present
  fi
  {
    printf '\n# Added by linkai-cli installer\n'
    printf '%s\n' "$_line"
  } >> "$_rc"
  ok "Added $_dir to PATH in $_rc"
  info "  Run 'source $_rc' or open a new terminal to use 'linkai'."
}

ensure_on_path() {
  _dir="$1"
  case ":$PATH:" in
    *":$_dir:"*) return 0 ;;  # already on PATH — nothing to do
  esac

  if [ "$NO_MODIFY_PATH" = "1" ]; then
    info ""
    info "⚠ $_dir is not on your PATH. Add it manually:"
    info "    export PATH=\"$_dir:\$PATH\""
    return 0
  fi

  # Interactive terminal: ask before editing rc files. Non-TTY (curl | sh):
  # auto-append, since the user cannot answer a prompt.
  if [ -t 0 ] && [ -t 1 ]; then
    printf '  %s is not on your PATH. Add it to your shell config now? [Y/n] ' "$_dir" >&2
    read _ans || _ans=""
    case "$_ans" in
      ""|y|Y|yes|YES) append_path_to_rc "$_dir" ;;
      *)
        info "  Skipped. Add it manually: export PATH=\"$_dir:\$PATH\""
        ;;
    esac
  else
    append_path_to_rc "$_dir"
  fi
}

# ── Binary install ────────────────────────────────────────────────────────────

install_binary() {
  os="$(detect_os)"
  arch="$(detect_arch)"
  resolve_version

  # goreleaser strips the leading 'v' in {{ .Version }} for archive names.
  ver_no_v="${VERSION#v}"
  archive="${BINARY}_${ver_no_v}_${os}_${arch}.tar.gz"
  url="$(asset_url "$archive")"

  tmp="$(mktemp -d 2>/dev/null || mktemp -d -t linkai)"
  trap 'rm -rf "$tmp"' EXIT INT TERM

  info "==> Downloading ${BINARY} ${VERSION} (${os}/${arch}) from ${SOURCE}"
  fetch_to "$url" "$tmp/$archive" || err "download failed: $url"

  # Verify checksum when checksums.txt is available.
  if fetch_to "$(asset_url checksums.txt)" "$tmp/checksums.txt" 2>/dev/null; then
    expected="$(grep " ${archive}\$" "$tmp/checksums.txt" 2>/dev/null | awk '{print $1}')"
    if [ -n "$expected" ]; then
      if command -v sha256sum >/dev/null 2>&1; then
        actual="$(sha256sum "$tmp/$archive" | awk '{print $1}')"
      elif command -v shasum >/dev/null 2>&1; then
        actual="$(shasum -a 256 "$tmp/$archive" | awk '{print $1}')"
      else
        actual=""
      fi
      if [ -n "$actual" ] && [ "$actual" != "$expected" ]; then
        err "checksum mismatch (expected $expected, got $actual)"
      fi
      [ -n "$actual" ] && ok "SHA256 checksum verified"
    fi
  fi

  info "==> Extracting"
  tar -xzf "$tmp/$archive" -C "$tmp"

  install_dir="$(resolve_install_dir)"
  mkdir -p "$install_dir"
  install_path="$install_dir/$BINARY"

  if [ -w "$install_dir" ]; then
    mv "$tmp/$BINARY" "$install_path"
  else
    info "==> $install_dir not writable, using sudo"
    sudo mv "$tmp/$BINARY" "$install_path"
  fi
  chmod +x "$install_path" 2>/dev/null || sudo chmod +x "$install_path"

  ok "Binary installed: $install_path ($VERSION)"
  ensure_on_path "$install_dir"
}

# ── Skill install ─────────────────────────────────────────────────────────────
#
# Installs the agent skill into every detected agent home so tools like Claude
# Code / Cursor / Codex can drive the CLI out of the box.
install_skill() {
  resolve_version
  archive="${SKILL_NAME}-skill.tar.gz"
  url="$(asset_url "$archive")"

  tmp="$(mktemp -d 2>/dev/null || mktemp -d -t linkai-skill)"

  info "==> Installing agent skill"
  if ! fetch_to "$url" "$tmp/$archive" 2>/dev/null; then
    info "⚠ Could not download the skill archive; skipping skill install."
    rm -rf "$tmp"
    return 0
  fi

  src="$tmp/skill"
  mkdir -p "$src"
  tar -xzf "$tmp/$archive" -C "$src" 2>/dev/null || {
    info "⚠ Could not extract the skill archive; skipping."
    rm -rf "$tmp"
    return 0
  }
  # The archive may wrap content in a top-level SKILL_NAME/ dir.
  if [ -f "$src/$SKILL_NAME/SKILL.md" ]; then
    src="$src/$SKILL_NAME"
  fi
  [ -f "$src/SKILL.md" ] || { info "⚠ SKILL.md not found in archive; skipping."; rm -rf "$tmp"; return 0; }

  installed=0
  for agent_dir in \
    ".agents/skills" "cow/skills" ".claude/skills" ".cursor/skills" ".codex/skills" \
    ".gemini/skills" ".windsurf/skills" ".qoder/skills"
  do
    parent="$HOME/$(dirname "$agent_dir")"
    # Only install into agent homes that already exist (except the generic
    # .agents fallback, always attempted).
    if [ "$agent_dir" != ".agents/skills" ] && [ ! -d "$parent" ]; then
      continue
    fi
    dest="$HOME/$agent_dir/$SKILL_NAME"
    rm -rf "$dest"
    mkdir -p "$dest"
    cp -R "$src/." "$dest/" 2>/dev/null || cp -r "$src/." "$dest/"
    ok "Skill → ~/$agent_dir/$SKILL_NAME"
    installed=$((installed + 1))
  done
  [ "$installed" -eq 0 ] && info "  (no agent homes detected)"

  rm -rf "$tmp"
}

# ── Main ──────────────────────────────────────────────────────────────────────

main() {
  printf '\n'
  info "LinkAI CLI installer"
  printf '\n'

  [ "$SKILL_ONLY" != "1" ] && pick_source
  [ "$SKILL_ONLY" = "1" ] && pick_source

  if [ "$SKILL_ONLY" = "1" ]; then
    install_skill
  else
    install_binary
    [ "$NO_SKILL" != "1" ] && install_skill
  fi

  printf '\n'
  ok "Done!"
  info ""
  info "Next steps:"
  info "  linkai auth login     # authenticate with LinkAI"
  info "  linkai --help         # explore commands"
  printf '\n'

  # Best-effort version print (only if it's already reachable on PATH).
  if [ "$SKILL_ONLY" != "1" ] && command -v "$BINARY" >/dev/null 2>&1; then
    "$BINARY" --version 2>/dev/null || true
  fi
}

main
