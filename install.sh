#!/usr/bin/env bash
# daily-dev Linux installer.
#   curl -fsSL https://github.com/F4NT0/daily-dev-tui/releases/latest/download/install.sh | bash
set -euo pipefail

REPO="${DAILY_DEV_REPO:-F4NT0/daily-dev-tui}"
INSTALL_DIR="${DAILY_DEV_INSTALL_DIR:-$HOME/.local/bin}"
VERSION="3.0.0"

say() { printf '\033[35m==>\033[0m %s\n' "$*"; }
die() { printf '\033[31mError:\033[0m %s\n' "$*" >&2; exit 1; }

command -v curl >/dev/null 2>&1 || command -v wget >/dev/null 2>&1 || die "curl or wget is required"

case "$(uname -s)" in Linux) ;; *) die "this installer is for Linux only" ;; esac
case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) die "unsupported architecture: $(uname -m)" ;;
esac

asset="daily-dev-tui-linux-$arch"
url="${DAILY_DEV_TUI_URL:-https://github.com/$REPO/releases/latest/download/$asset}"

download() {
  if command -v curl >/dev/null 2>&1; then curl -fsSL "$1" -o "$2"; else wget -qO "$2" "$1"; fi
}

say "Creating $INSTALL_DIR"
mkdir -p "$INSTALL_DIR"

say "Downloading $asset"
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT
download "$url" "$tmp" || die "download failed from $url (is there a release with $asset?)"
install -m 0755 "$tmp" "$INSTALL_DIR/daily-dev-tui"

say "Installing the 'daily-dev' command"
cat > "$INSTALL_DIR/daily-dev" <<LAUNCHER
#!/usr/bin/env bash
DIR="\$(cd "\$(dirname "\$(readlink -f "\${BASH_SOURCE[0]}")")" && pwd)"
case "\${1:-}" in
  --version) echo "daily-dev $VERSION"; exit 0 ;;
  --help|-h|help)
    cat <<'HELP'
daily-dev - terminal UI for the Daily.dev public API

  daily-dev              Start the TUI
  daily-dev --help       Show this help
  daily-dev --version    Print the version
  daily-dev --uninstall  Remove daily-dev from this computer

Environment:
  DAILY_DEV_TOKEN      API token; skips the token prompt on startup
  DAILY_DEV_CLIENT_ID / DAILY_DEV_CLIENT_SECRET  OAuth app credentials (OAuth login)
  DAILY_DEV_INSECURE   Set to 1 to skip TLS verification (corporate proxies only)
HELP
    exit 0 ;;
  --uninstall)
    rm -f "\$DIR/daily-dev-tui" "\$DIR/daily-dev"
    for rc in "\$HOME/.bashrc" "\$HOME/.zshrc" "\$HOME/.profile"; do
      [ -f "\$rc" ] && sed -i '/# added by daily-dev/d' "\$rc"
    done
    echo "daily-dev removed."
    exit 0 ;;
esac
exec "\$DIR/daily-dev-tui" "\$@"
LAUNCHER
chmod +x "$INSTALL_DIR/daily-dev"

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    say "Adding $INSTALL_DIR to your PATH"
    case "$(basename "${SHELL:-}")" in
      zsh) rc="$HOME/.zshrc" ;;
      bash) rc="$HOME/.bashrc" ;;
      *) rc="$HOME/.profile" ;;
    esac
    printf 'export PATH="%s:$PATH" # added by daily-dev\n' "$INSTALL_DIR" >> "$rc"
    say "Open a new terminal (or run: source $rc)"
    ;;
esac

say "Done. Run: daily-dev"
