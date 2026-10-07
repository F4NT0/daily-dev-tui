package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	accent = lipgloss.Color("#A855F7")
	green  = lipgloss.Color("#22C55E")
	red    = lipgloss.Color("#EF4444")
	muted  = lipgloss.Color("#6B7280")
	cyan   = lipgloss.Color("#22D3EE")
	yellow = lipgloss.Color("#FACC15")

	titleSt = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(accent).Padding(0, 2)
	okSt    = lipgloss.NewStyle().Foreground(green)
	errSt   = lipgloss.NewStyle().Foreground(red)
	mutedSt = lipgloss.NewStyle().Foreground(muted)
	hiSt    = lipgloss.NewStyle().Foreground(cyan).Bold(true)
)

type step struct {
	name string
	fn   func() error
}

type stepDone struct{ err error }

type instModel struct {
	steps     []step
	cur       int
	err       error
	finished  bool
	confirmed bool
	uninstall bool
	sp        spinner.Model
}

func runInstaller(uninstall bool) {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(accent)
	m := instModel{sp: sp, uninstall: uninstall}
	if uninstall {
		m.steps = []step{{"Remove daily-dev from your PATH", stepRemovePath}, {"Remove installed files", stepRemoveFiles}}
	} else {
		m.steps = []step{
			{"Create install directory", stepCreateDir},
			{"Install daily-dev-tui.exe", stepCopyTUI},
			{"Install the `daily-dev` command", stepCopyLauncher},
			{"Add install directory to your PATH", stepAddPath},
		}
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (m instModel) Init() tea.Cmd { return m.sp.Tick }

func (m instModel) run() tea.Cmd {
	s := m.steps[m.cur]
	return func() tea.Msg { time.Sleep(350 * time.Millisecond); return stepDone{s.fn()} }
}

func (m instModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "enter", "y":
			if m.finished {
				return m, tea.Quit
			}
			if !m.confirmed {
				m.confirmed = true
				return m, m.run()
			}
		}
	case stepDone:
		if msg.err != nil {
			m.err, m.finished = msg.err, true
			return m, nil
		}
		m.cur++
		if m.cur >= len(m.steps) {
			m.finished = true
			return m, nil
		}
		return m, m.run()
	case spinner.TickMsg:
		var c tea.Cmd
		m.sp, c = m.sp.Update(msg)
		return m, c
	}
	return m, nil
}

func (m instModel) View() string {
	title, verb := "Daily.dev TUI Installer", "Install"
	if m.uninstall {
		title, verb = "Daily.dev TUI Uninstaller", "Uninstall"
	}
	s := "\n  " + titleSt.Render(title) + "\n\n"
	if !m.confirmed {
		s += fmt.Sprintf("  %s will %s:\n  %s\n\n", verb, map[bool]string{true: "remove", false: "set up"}[m.uninstall], hiSt.Render(installDir()))
		for _, st := range m.steps {
			s += "   " + mutedSt.Render("• "+st.name) + "\n"
		}
		return s + "\n  " + hiSt.Render("enter") + " continue  " + mutedSt.Render("• q: quit") + "\n"
	}
	for i, st := range m.steps {
		switch {
		case i < m.cur:
			s += "  " + okSt.Render("✔ "+st.name) + "\n"
		case i == m.cur && m.err != nil:
			s += "  " + errSt.Render("✘ "+st.name+": "+m.err.Error()) + "\n"
		case i == m.cur && !m.finished:
			s += "  " + m.sp.View() + st.name + "\n"
		default:
			s += "  " + mutedSt.Render("  "+st.name) + "\n"
		}
	}
	if m.finished && m.err == nil {
		if m.uninstall {
			s += "\n  " + okSt.Bold(true).Render("Daily.dev TUI was uninstalled.") + "\n"
		} else {
			s += "\n  " + okSt.Bold(true).Render("Installation complete!") + "\n\n" +
				"  Open a " + lipgloss.NewStyle().Foreground(yellow).Render("NEW terminal") + " and run:\n\n" +
				"    " + hiSt.Render("daily-dev") + "          start the TUI\n" +
				"    " + hiSt.Render("daily-dev --help") + "   read the interactive docs\n"
		}
	}
	if m.finished {
		s += "\n  " + mutedSt.Render("press enter to exit") + "\n"
	}
	return s
}
