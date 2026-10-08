package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stToken state = iota
	stMenu
	stForm
	stLoading
	stComments
	stLogin
	stOAuthForm
	stOAuthWait
	stHelp
)

type resultMsg struct {
	res any
	err error
}

type actionMsg struct {
	text string
	err  error
	id   string
}

type commentsMsg struct {
	res any
	err error
}

type model struct {
	res        any
	items      []map[string]any
	sel        int
	selectable bool
	offs       [][2]int
	cvp        viewport.Model
	cLoading   bool
	helpTab    int
	helpPrev   state
	bm         map[string]bool

	oauth    *oauthState
	sess     *oauthSession
	loginSel int
	oi       []textinput.Model
	oFocus   int
	loginErr string

	st         state
	token      string
	client     *http.Client
	cur        int
	input      textinput.Model
	inputs     []textinput.Model
	focus      int
	vals       map[string]string
	active     *Endpoint
	cursor     string
	vp         viewport.Model
	sp         spinner.Model
	rightFocus bool
	w, h       int
	status     string
}

const leftW = 36

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Bearer token"
	ti.EchoMode = textinput.EchoPassword
	ti.Focus()
	m := model{client: newClient(), input: ti, sp: spinner.New(spinner.WithSpinner(spinner.Dot)), vp: viewport.New(10, 10)}
	m.bm = map[string]bool{}
	m.st = stLogin
	m.oi = newOAuthInputs()
	m.cvp = viewport.New(10, 10)
	m.vp.SetContent(metaSt.Render("Select a request on the left and press Enter."))
	if t := strings.TrimSpace(os.Getenv("DAILY_DEV_TOKEN")); t != "" {
		m.token, m.st = t, stMenu
	}
	return m
}

func (m model) bearer() string {
	if m.oauth != nil {
		return m.oauth.Token(m.client)
	}
	return m.token
}

func (m model) Init() tea.Cmd { return m.sp.Tick }

func (m model) fetch(e Endpoint, vals map[string]string) tea.Cmd {
	return func() tea.Msg {
		r, err := doGet(m.client, m.bearer(), e, vals)
		return resultMsg{r, err}
	}
}

func (m *model) layout() {
	m.vp.Width = m.w - leftW - 6
	m.vp.Height = m.h - 3
	m.cvp.Width = m.modalW() - 4
	m.cvp.Height = m.h - 10
}

func (m model) modalW() int {
	w := m.w - 12
	if w > 100 {
		w = 100
	}
	if w < 30 {
		w = 30
	}
	return w
}

func (m *model) rerender() {
	m.vp.SetContent(m.renderRes())
}

func (m *model) renderRes() string {
	s, offs := renderResult(m.res, m.vp.Width, m.sel, m.selectable, m.bm)
	m.offs = offs
	return s
}

func (m *model) ensureVisible() {
	if !m.selectable || m.sel >= len(m.offs) {
		return
	}
	st, h := m.offs[m.sel][0], m.offs[m.sel][1]
	if st < m.vp.YOffset || h >= m.vp.Height {
		m.vp.SetYOffset(st)
	} else if st+h > m.vp.YOffset+m.vp.Height {
		m.vp.SetYOffset(st + h - m.vp.Height)
	}
}

func (m model) selPost() map[string]any {
	if !m.selectable || m.sel >= len(m.items) {
		return nil
	}
	return m.items[m.sel]
}

func (m model) fetchComments(id string) tea.Cmd {
	return func() tea.Msg {
		r, err := doGet(m.client, m.bearer(), endpoints[6], map[string]string{"id": id, "sort": "newest", "limit": "50"})
		return commentsMsg{r, err}
	}
}

func (m model) bookmark(id string) tea.Cmd {
	return func() tea.Msg {
		err := doSend(m.client, m.bearer(), "POST", "/bookmarks/", map[string]any{"postIds": []string{id}})
		return actionMsg{"bookmarked", err, id}
	}
}

func (m *model) run(e *Endpoint, vals map[string]string) tea.Cmd {
	m.active, m.vals, m.st, m.cursor = e, vals, stLoading, ""
	return m.fetch(*e, vals)
}

func (m *model) startForm(e *Endpoint) tea.Cmd {
	m.active, m.focus = e, 0
	m.inputs = nil
	for _, f := range e.Fields {
		ti := textinput.New()
		ti.Prompt = ""
		ti.Placeholder = f.Label
		m.inputs = append(m.inputs, ti)
	}
	m.inputs[0].Focus()
	m.st = stForm
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.layout()
		return m, nil
	case spinner.TickMsg:
		var c tea.Cmd
		m.sp, c = m.sp.Update(msg)
		return m, c
	case resultMsg:
		m.st = stMenu
		if msg.err != nil {
			m.vp.SetContent(errSt.Width(m.vp.Width).Render("Error: " + msg.err.Error()))
			m.status = "request failed"
		} else {
			m.cursor = nextCursor(msg.res)
			m.res, m.sel = msg.res, 0
			m.items = selectablePosts(msg.res)
			m.selectable = len(m.items) > 0
			if m.selectable {
				m.rightFocus = true
			}
			m.vp.SetContent(m.renderRes())
			m.status = "ok"
			if m.cursor != "" {
				m.status = "ok - press n for next page"
			}
		}
		m.vp.GotoTop()
		return m, nil
	case oauthDoneMsg:
		if m.st != stOAuthWait {
			return m, nil
		}
		m.sess = nil
		if msg.err != nil {
			m.st, m.loginErr = stOAuthForm, msg.err.Error()
			return m, nil
		}
		m.oauth, m.st, m.status = msg.st, stMenu, "signed in with OAuth"
		return m, nil
	case actionMsg:
		if msg.err != nil {
			m.status = "failed: " + msg.err.Error()
		} else {
			m.status = msg.text
			if msg.text == "bookmarked" {
				m.bm[msg.id] = true
				m.rerender()
			}
		}
		return m, nil
	case commentsMsg:
		m.cLoading = false
		if msg.err != nil {
			m.cvp.SetContent(errSt.Width(m.cvp.Width).Render("Error: " + msg.err.Error()))
		} else {
			m.cvp.SetContent(renderComments(msg.res, m.cvp.Width))
		}
		m.cvp.GotoTop()
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "ctrl+h" && (m.st == stMenu || m.st == stForm) {
			m.helpPrev, m.st, m.helpTab = m.st, stHelp, 0
			if m.rightFocus {
				m.helpTab = 1
			}
			return m, nil
		}
		switch m.st {
		case stLogin:
			return m.updateLogin(msg)
		case stOAuthForm:
			return m.updateOAuthForm(msg)
		case stOAuthWait:
			return m.updateOAuthWait(msg)
		case stToken:
			if msg.String() == "esc" {
				m.st, m.loginErr = stLogin, ""
				return m, nil
			}
			if msg.String() == "enter" {
				if t := strings.TrimSpace(m.input.Value()); t != "" {
					m.token, m.st = t, stMenu
				}
				return m, nil
			}
			var c tea.Cmd
			m.input, c = m.input.Update(msg)
			return m, c
		case stForm:
			return m.updateForm(msg)
		case stLoading:
			return m, nil
		case stComments:
			return m.updateComments(msg)
		case stHelp:
			return m.updateHelp(msg)
		}
		return m.updateMenu(msg)
	}
	return m, nil
}

func (m model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.st = stMenu
		return m, nil
	case "tab", "down", "shift+tab", "up":
		m.inputs[m.focus].Blur()
		d := 1
		if msg.String() == "shift+tab" || msg.String() == "up" {
			d = -1
		}
		m.focus = (m.focus + d + len(m.inputs)) % len(m.inputs)
		return m, m.inputs[m.focus].Focus()
	case "enter":
		vals := map[string]string{}
		for i, f := range m.active.Fields {
			v := strings.TrimSpace(m.inputs[i].Value())
			if f.Req && v == "" {
				m.inputs[m.focus].Blur()
				m.focus = i
				m.status = f.Label + " is required"
				return m, m.inputs[i].Focus()
			}
			vals[f.Key] = v
		}
		return m, m.run(m.active, vals)
	}
	var c tea.Cmd
	m.inputs[m.focus], c = m.inputs[m.focus].Update(msg)
	return m, c
}

func (m model) updateComments(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "c", "enter":
		m.st = stMenu
		return m, nil
	}
	var c tea.Cmd
	m.cvp, c = m.cvp.Update(msg)
	return m, c
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	switch k {
	case "q":
		return m, tea.Quit
	case "tab":
		m.rightFocus = !m.rightFocus
		return m, nil
	case "n":
		if m.cursor != "" && m.active != nil {
			v := map[string]string{}
			for a, b := range m.vals {
				v[a] = b
			}
			v["cursor"] = m.cursor
			return m, m.run(m.active, v)
		}
	}
	if m.rightFocus && m.selectable {
		if p := m.selPost(); p != nil {
			id, _ := p["id"].(string)
			switch k {
			case "up", "k":
				if m.sel > 0 {
					m.sel--
					m.rerender()
					m.ensureVisible()
				}
				return m, nil
			case "down", "j":
				if m.sel < len(m.items)-1 {
					m.sel++
					m.rerender()
					m.ensureVisible()
				}
				return m, nil
			case "b":
				m.status = "bookmarking..."
				return m, m.bookmark(id)
			case "o":
				u, _ := p["url"].(string)
				if u == "" {
					u, _ = p["commentsPermalink"].(string)
				}
				if err := openURL(u); err != nil {
					m.status = "open failed: " + err.Error()
				} else {
					m.status = "opened post url"
				}
				return m, nil
			case "enter":
				u, _ := p["commentsPermalink"].(string)
				if err := openURL(u); err != nil {
					m.status = "open failed: " + err.Error()
				} else {
					m.status = "opened on daily.dev"
				}
				return m, nil
			case "c":
				m.st, m.cLoading = stComments, true
				m.cvp.SetContent("")
				return m, m.fetchComments(id)
			}
		}
	}
	if m.rightFocus {
		var c tea.Cmd
		m.vp, c = m.vp.Update(msg)
		return m, c
	}
	switch k {
	case "up", "k":
		if m.cur > 0 {
			m.cur--
		}
	case "down", "j":
		if m.cur < len(endpoints)-1 {
			m.cur++
		}
	case "enter":
		e := &endpoints[m.cur]
		if len(e.Fields) == 0 {
			return m, m.run(e, map[string]string{})
		}
		return m, m.startForm(e)
	default:
		var c tea.Cmd
		m.vp, c = m.vp.Update(msg)
		return m, c
	}
	return m, nil
}

func (m model) leftView() string {
	var b strings.Builder
	last := ""
	for i, e := range endpoints {
		if e.Group != last {
			last = e.Group
			b.WriteString("\n" + keySt.Bold(true).Render(e.Group) + "\n")
		}
		line := "  " + e.Name
		if i == m.cur {
			line = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("135")).Render("▸ " + e.Name)
		}
		b.WriteString(line + "\n")
	}
	lines := strings.Split(b.String(), "\n")
	// keep the selected item visible
	idx := 0
	for n, l := range lines {
		if strings.Contains(l, "▸") {
			idx = n
		}
	}
	vh := m.h - 4
	start := 0
	if idx >= vh {
		start = idx - vh + 1
	}
	end := start + vh
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[start:end], "\n")
}

func (m model) View() string {
	if m.w == 0 {
		return "loading..."
	}
	switch m.st {
	case stLogin, stToken, stOAuthForm, stOAuthWait:
		return m.tokenView()
	}
	if m.st == stHelp {
		return m.helpView()
	}
	bc := lipgloss.Color("240")
	lb, rb := lipgloss.Color("135"), bc
	if m.rightFocus {
		lb, rb = bc, lipgloss.Color("135")
	}
	left := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lb).Width(leftW).Height(m.h-3).Padding(0, 1).Render(titleSt.Render("Daily.dev TUI") + m.leftView())

	var right string
	switch m.st {
	case stForm:
		var b strings.Builder
		b.WriteString(titleSt.Render(m.active.Name) + "  " + metaSt.Render("GET "+m.active.Path) + "\n\n")
		for i, f := range m.active.Fields {
			label := f.Label
			if f.Req {
				label += " *"
			}
			b.WriteString(keySt.Render(label) + "\n" + m.inputs[i].View() + "\n\n")
		}
		right = b.String()
	case stLoading:
		right = m.sp.View() + " Loading " + m.active.Path + "..."
	default:
		right = m.vp.View()
	}
	rw := m.w - leftW - 4
	rightBox := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(rb).Width(rw).Height(m.h-3).Padding(0, 1).Render(right)
	help := "↑/↓: select option • tab: change panel • enter: run command • ctrl+h: help • q: quit"
	if m.st == stForm {
		help = "tab/↑/↓: field • enter: run • esc: cancel • ctrl+h: help"
	}
	if m.rightFocus && m.selectable && m.st == stMenu {
		help = "enter: daily.dev • o: link • b: bookmark • c: comments • tab: panel • ctrl+h: help"
	}
	if m.st == stComments {
		help = "↑/↓/pgup/pgdn: scroll • esc/q/enter: close"
	}
	st := ""
	if m.status != "" {
		st = "  [" + m.status + "]"
	}
	footer := "\n" + metaSt.Render(help+st)
	if m.st == stComments {
		body := titleSt.Render("Comments") + "  " + metaSt.Render(fmt.Sprint(m.selPost()["title"])) + "\n\n"
		if m.cLoading {
			body += m.sp.View() + " Loading...\n"
		} else {
			body += m.cvp.View() + "\n"
		}
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("135")).Width(m.modalW()).Padding(0, 1).Render(body)
		return lipgloss.Place(m.w, m.h-1, lipgloss.Center, lipgloss.Center, box) + footer
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, left, rightBox) + footer
}

func main() {
	if _, err := tea.NewProgram(initialModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
