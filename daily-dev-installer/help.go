package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type page struct{ name, body string }

func h(s string) string { return lipgloss.NewStyle().Bold(true).Foreground(accent).Render(s) }
func c(s string) string { return lipgloss.NewStyle().Foreground(cyan).Render(s) }
func y(s string) string { return lipgloss.NewStyle().Foreground(yellow).Render(s) }
func g(s string) string { return lipgloss.NewStyle().Foreground(green).Render(s) }
func d(s string) string { return mutedSt.Render(s) }

type ep struct{ method, path, desc, params string }

var epGroups = []struct {
	name string
	eps  []ep
}{
	{"Feeds", []ep{
		{"GET", "/feeds/foryou", "Your personalized feed.", "limit, cursor"},
		{"GET", "/feeds/popular", "Most popular posts, optionally filtered by tags.", "tags, limit, cursor"},
		{"GET", "/feeds/discussed", "Most discussed posts.", "period (1-30 days), tag, source, limit, cursor"},
		{"GET", "/feeds/tag/{tag}", "Posts for a tag.", "tag*, limit, cursor"},
		{"GET", "/feeds/source/{source}", "Posts from a source (id or handle).", "source*, limit, cursor"},
	}},
	{"Posts", []ep{
		{"GET", "/posts/{id}", "Details of a single post.", "id*"},
		{"GET", "/posts/{id}/comments", "Comments of a post.", "id*, sort (oldest/newest), limit, cursor"},
	}},
	{"Search", []ep{
		{"GET", "/search/posts", "Full-text post search.", "q*, time (day/week/month/year/all), limit, cursor"},
		{"GET", "/search/tags", "Search tags.", "q*"},
		{"GET", "/search/sources", "Search sources.", "q*"},
	}},
	{"Bookmarks", []ep{
		{"GET", "/bookmarks/", "Your bookmarks.", "unreadOnly, listId, limit, cursor"},
		{"GET", "/bookmarks/search", "Search your bookmarks.", "q*, limit, cursor"},
		{"GET", "/bookmarks/lists", "Your bookmark lists.", "-"},
	}},
	{"Custom Feeds", []ep{
		{"GET", "/feeds/custom/", "List your custom feeds.", "-"},
		{"GET", "/feeds/custom/{feedId}", "Posts of a custom feed.", "feedId*, limit, cursor"},
		{"GET", "/feeds/custom/{feedId}/info", "Custom feed settings.", "feedId*"},
		{"GET", "/feeds/custom/advanced-settings", "Available advanced settings.", "-"},
	}},
	{"Feed Filters", []ep{{"GET", "/feeds/filters/", "Your global feed filters.", "-"}}},
	{"Notifications", []ep{
		{"GET", "/notifications/", "Your notifications.", "limit, cursor"},
		{"GET", "/notifications/unread/count", "Number of unread notifications.", "-"},
	}},
	{"Profile", []ep{
		{"GET", "/profile/", "Your profile.", "-"},
		{"GET", "/profile/stack/search", "Search the tech stack catalog.", "q*"},
		{"GET", "/profile/stack/", "Your tech stack.", "-"},
		{"GET", "/profile/experiences/", "Your experiences.", "type (work, education, project, ...)"},
		{"GET", "/profile/experiences/{id}", "A single experience.", "id*"},
	}},
	{"Tags", []ep{{"GET", "/tags/", "All tags.", "-"}}},
	{"Recommend", []ep{{"GET", "/recommend/keyword", "Recommendations by keyword (experimental).", "q*, limit"}}},
}

func pages() []page {
	var api strings.Builder
	api.WriteString(h("API REQUESTS") + "\n\nBase URL: " + c("https://api.daily.dev/public/v1") + "\nEvery request sends " + c("Authorization: Bearer <token>") + ".\n" + d("* = required parameter. Path parameters are marked with {braces}.") + "\n")
	for _, gr := range epGroups {
		api.WriteString("\n" + y("▌ "+gr.name) + "\n")
		for _, e := range gr.eps {
			api.WriteString(fmt.Sprintf("  %s %s\n      %s\n      %s %s\n", g(e.method), c(e.path), e.desc, d("params:"), e.params))
		}
	}
	return []page{
		{"Overview", h("DAILY-DEV") + "\n\nA terminal UI to browse the " + c("Daily.dev") + " public API.\n\n" + h("COMMANDS") + "\n" +
			"  " + c("daily-dev") + "              Start the TUI\n" +
			"  " + c("daily-dev --help") + "       Show this documentation\n" +
			"  " + c("daily-dev --version") + "    Print the version\n" +
			"  " + c("daily-dev --uninstall") + "  Remove daily-dev from this computer\n\n" +
			h("ENVIRONMENT") + "\n" +
			"  " + y("DAILY_DEV_TOKEN") + "     API token; skips the token prompt on startup\n" +
			"  " + y("DAILY_DEV_CLIENT_ID") + " / " + y("DAILY_DEV_CLIENT_SECRET") + "  OAuth app credentials (OAuth login)\n" +
			"  " + y("DAILY_DEV_INSECURE") + "  Set to 1 to skip TLS verification (corporate proxies only)\n"},
		{"Login", h("HOW TO LOG IN") + "\n\nThe TUI authenticates with a personal API token (Bearer token).\n\n" + h("1. Create a token") + "\n" +
			"  1) Open " + c("https://app.daily.dev") + " and sign in.\n" +
			"  2) Go to " + y("Settings") + " (your profile menu) and find the " + y("API / Developers") + " section.\n" +
			"  3) Click " + y("Create / Generate token") + ", name it (e.g. \"daily-dev-tui\").\n" +
			"  4) Copy the token right away - it may not be shown again.\n" +
			d("     (API access can require a Daily.dev Plus subscription. Menu names may vary.)") + "\n\n" +
			h("2. Use the token") + "\n" +
			"  " + g("Option A") + " - run " + c("daily-dev") + " and paste the token at the prompt.\n" +
			"  " + g("Option B") + " - set it once in PowerShell:\n" +
			"      " + c(`[Environment]::SetEnvironmentVariable('DAILY_DEV_TOKEN','<your token>','User')`) + "\n" +
			"    then open a new terminal and run " + c("daily-dev") + ".\n\n" +
			h("OAUTH (alternative)") + "\n" +
			"  On the start screen choose " + g("OAuth (sign in with daily.dev)") + ".\n" +
			"  1) In Daily.dev go to " + y("Settings > API > OAuth apps") + " and click " + y("Create app") + ".\n" +
			"  2) Add the redirect URI " + c("http://127.0.0.1:8765/callback") + " and copy the client ID and secret.\n" +
			"  3) Enter them in the TUI (or set " + y("DAILY_DEV_CLIENT_ID") + " / " + y("DAILY_DEV_CLIENT_SECRET") + ").\n" +
			"  4) Approve access in the browser that opens. Tokens are refreshed automatically.\n\n" +
			y("Keep your token secret.") + " Revoke it in Daily.dev settings if it leaks.\n"},
		{"Using the TUI", h("USING THE TUI") + "\n\n" +
			"  1) Start " + c("daily-dev") + " and enter your token.\n" +
			"  2) Pick a request from the menu, grouped by area (Feeds, Posts, Search, ...).\n" +
			"  3) Fill in the parameters. Required ones are mandatory; leave optional ones empty.\n" +
			"  4) The JSON response is rendered on screen.\n\n" + h("PAGINATION") + "\n" +
			"  List requests accept " + y("limit") + " (1-50) and " + y("cursor") + ".\n" +
			"  Copy the cursor from a response and pass it to get the next page.\n\n" +
			h("TIPS") + "\n  • " + c("ctrl+c") + " quits at any time.\n  • 401 errors mean the token is invalid or revoked.\n  • 429 errors mean you are rate limited; wait and retry.\n"},
		{"API Requests", api.String()},
	}
}

type helpModel struct {
	pages []page
	tab   int
	vp    viewport.Model
	ready bool
}

func (m *helpModel) load() { m.vp.SetContent(m.pages[m.tab].body); m.vp.GotoTop() }

func runHelp() {
	m := &helpModel{pages: pages()}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		for _, p := range m.pages {
			fmt.Println(p.body)
		}
		return
	}
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println(err)
	}
}

func (m *helpModel) Init() tea.Cmd { return nil }

func (m *helpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.vp = viewport.New(msg.Width-4, msg.Height-6)
		m.ready = true
		m.load()
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "right", "tab", "l":
			m.tab = (m.tab + 1) % len(m.pages)
			m.load()
			return m, nil
		case "left", "shift+tab", "h":
			m.tab = (m.tab + len(m.pages) - 1) % len(m.pages)
			m.load()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m *helpModel) View() string {
	if !m.ready {
		return ""
	}
	var tabs []string
	for i, p := range m.pages {
		st := lipgloss.NewStyle().Padding(0, 2).Foreground(muted)
		if i == m.tab {
			st = st.Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(accent)
		}
		tabs = append(tabs, st.Render(p.name))
	}
	return "\n  " + lipgloss.JoinHorizontal(lipgloss.Top, tabs...) + "\n\n" +
		lipgloss.NewStyle().PaddingLeft(2).Render(m.vp.View()) + "\n  " +
		mutedSt.Render("←/→ switch tab • ↑/↓ scroll • q quit")
}
