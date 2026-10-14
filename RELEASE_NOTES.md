# daily-dev v4.0.0

A terminal UI for the [Daily.dev](https://app.daily.dev) public API, now on Windows, Linux and macOS.

## Install

**Windows (PowerShell)**
```powershell
irm https://github.com/F4NT0/daily-dev-tui/releases/latest/download/daily-dev-setup.exe -OutFile $env:TEMP\daily-dev-setup.exe; & $env:TEMP\daily-dev-setup.exe
```

**Linux**
```bash
curl -fsSL https://github.com/F4NT0/daily-dev-tui/releases/latest/download/install.sh | bash
```

**macOS (Intel and Apple Silicon)**
```bash
curl -fsSL https://github.com/F4NT0/daily-dev-tui/releases/latest/download/install-macos.sh | bash
```

Open a new terminal and run `daily-dev`.

## What's new

### macOS support
- New `install-macos.sh` one-line installer.
- Native TUI builds for Intel (`darwin-amd64`) and Apple Silicon (`darwin-arm64`).

### Themes
- Every TUI color is a variable in a `.conf` file inside the `themes/` directory next to the program.
- Included themes: Default, Catppuccin Mocha, Nord, Terminal and Darcula.
- Press `t` for a floating theme picker with live preview. The choice is remembered.
- Create your own by copying a `.conf` file; it appears in the picker automatically.
- Installers never overwrite existing theme files.

### Login
- Choose between **Bearer token** and **OAuth** (sign in with daily.dev, authorization code + PKCE, automatic token refresh).

### Posts
- Select a post in the results, bookmark it (orange border), open it on daily.dev or open the original link.
- Read-only comments panel (the public API cannot create comments).
- Noisy fields (`commentsPermalink`, `source.handle`, `source.id`) are no longer shown in the cards.

### Help and links
- Press `h` for a floating help panel with one tab per panel (`Ctrl+H` inside forms).
- If no browser can be opened (for example no `xdg-open` on Linux), a floating panel shows the post and the link to open manually. Linux also tries `$BROWSER`, `gio`, `wslview` and common browsers.

### Quality
- Unit tests for every API call, run automatically on each push and pull request.
- `install.sh` is LF-only so it runs correctly on Linux.

## Keyboard shortcuts

**Options panel**

| Key | Action |
|-----|--------|
| `↑/↓` | Select option |
| `Enter` | Run command |
| `Tab` | Change panel |
| `t` | Choose theme |
| `h` | Help (`Ctrl+H` inside forms) |
| `q` | Quit |

**Posts panel**

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Select post |
| `Enter` | Open in Daily.dev |
| `o` | Open the original post link |
| `b` | Bookmark |
| `c` | Show comments |
| `n` | Next page |
| `Tab` | Change panel |
| `t` | Choose theme |
| `h` | Help |
| `q` | Quit |

## Release assets

`daily-dev-setup.exe`, `daily-dev-tui.exe`, `install.sh`, `daily-dev-tui-linux-amd64`, `daily-dev-tui-linux-arm64`, `install-macos.sh`, `daily-dev-tui-darwin-amd64`, `daily-dev-tui-darwin-arm64`, `default.conf`, `catppuccin-mocha.conf`, `nord.conf`, `terminal.conf`, `darcula.conf`

## Notes

- OAuth apps are in beta on daily.dev and must be registered manually (redirect URI `http://127.0.0.1:8765/callback`).
- If you deny the `write` permission, bookmarking fails with `403 insufficient_scope`.
