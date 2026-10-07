# daily-dev

A terminal toolkit for the [Daily.dev](https://app.daily.dev) public API, written in Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Repository layout

| Path | What it is |
|------|------------|
| `daily-dev-tui/` | The TUI client. Shows a token login screen, then a menu of API requests, parameter forms and a formatted JSON viewer. |
| `daily-dev-tui/main.go` | Bubble Tea model: states (token, menu, form, loading), key handling and layout. |
| `daily-dev-tui/api.go` | HTTP client. Sends `GET` requests to `https://api.daily.dev/public/v1` with `Authorization: Bearer <token>`. |
| `daily-dev-tui/endpoints.go` | The list of supported API requests, with their path/query parameters. |
| `daily-dev-tui/render.go` | Styles and rendering of API responses (posts, tags, sources, ...). |
| `daily-dev-tui/splash.go` | The purple ASCII "DAILY.DEV API" login screen. |
| `daily-dev-installer/` | Windows installer. Embeds the TUI binary, installs it and registers the `daily-dev` command. |
| `daily-dev-installer/help.go` | The colored, tabbed `daily-dev --help` documentation. |
| `daily-dev-installer/install.go` | Install steps: copy files, add the install directory to the user `PATH`. |
| `daily_dev_search.py`, `requirements.txt` | Standalone Python search script. |

## Requirements

- Go 1.26+
- A Daily.dev API token (see below)
- Windows for the installer (the TUI itself is plain Go)

## Getting an API token

1. Sign in at <https://app.daily.dev>.
2. Open your account settings and find the API / Developers section.
3. Create a token and copy it immediately. API access may require Daily.dev Plus.

Keep the token secret and revoke it if it leaks.

## Build

Binaries are not committed; build them locally.

```powershell
# 1. Build the TUI
cd daily-dev-tui
go build -o dailydev-tui.exe .

# 2. Copy it into the installer payload and build the installer
copy dailydev-tui.exe ..\daily-dev-installer\payload\daily-dev-tui.exe
cd ..\daily-dev-installer
go build -o dist\daily-dev-setup.exe .
```

The installer embeds `payload/daily-dev-tui.exe`, so step 2 must be done before building it.

## Install

Run `daily-dev-setup.exe` and confirm. It will:

1. Create `%LOCALAPPDATA%\Programs\daily-dev`.
2. Install `daily-dev-tui.exe` and the `daily-dev` launcher there.
3. Add that directory to your user `PATH`.

Open a new terminal afterwards.

## Usage

| Command | Description |
|---------|-------------|
| `daily-dev` | Start the TUI. |
| `daily-dev --help` | Interactive, colored documentation (login, usage, every API request). |
| `daily-dev --version` | Print the version. |
| `daily-dev --uninstall` | Remove the command and installed files. |

### Environment variables

| Variable | Effect |
|----------|--------|
| `DAILY_DEV_TOKEN` | API token. Skips the login screen. |
| `DAILY_DEV_INSECURE` | Set to `1` to skip TLS verification (corporate proxies only). |

### Using the TUI

1. Enter your token on the login screen (input is masked).
2. Pick a request from the left menu (Feeds, Posts, Search, Bookmarks, Custom Feeds, Notifications, Profile, Tags, Recommend).
3. Fill in the parameters. Required ones are marked with `*`.
4. Read the formatted response. For paginated lists, pass the returned `cursor` to get the next page.

`401` means the token is invalid or revoked; `429` means you are rate limited.

## Fonts

For the best look, use a Nerd Font such as JetBrainsMono Nerd Font in your terminal.
