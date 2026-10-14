package main

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newOAuthInputs() []textinput.Model {
	id, sec := textinput.New(), textinput.New()
	id.Prompt, sec.Prompt = "> ", "> "
	id.Placeholder, sec.Placeholder = "Client ID", "Client secret"
	sec.EchoMode = textinput.EchoPassword
	id.SetValue(strings.TrimSpace(os.Getenv("DAILY_DEV_CLIENT_ID")))
	sec.SetValue(strings.TrimSpace(os.Getenv("DAILY_DEV_CLIENT_SECRET")))
	id.Focus()
	return []textinput.Model{id, sec}
}

func (m model) updateLogin(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "shift+tab":
		m.loginSel = (m.loginSel + len(loginOpts(m.saved)) - 1) % len(loginOpts(m.saved))
	case "down", "j", "tab":
		m.loginSel = (m.loginSel + 1) % len(loginOpts(m.saved))
	case "1":
		m.loginSel = 0
	case "2":
		m.loginSel = 1
	case "3":
		if m.saved != "" {
			m.loginSel = 2
		}
	case "d":
		if m.loginSel == 2 && m.saved != "" {
			p := tokenFilePath()
			if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
				m.loginErr, m.loginMsg = "could not delete "+p+": "+err.Error(), ""
				return m, nil
			}
			m.saved, m.loginSel, m.loginErr, m.loginMsg = "", 0, "", "Saved token deleted ("+p+")"
		}
	case "enter":
		m.loginErr, m.loginMsg = "", ""
		if m.loginSel == 0 {
			m.st = stToken
			return m, m.input.Focus()
		}
		if m.loginSel == 2 && m.saved != "" {
			m.token, m.st, m.status = m.saved, stMenu, "signed in with saved token"
			return m, nil
		}
		m.st, m.oFocus = stOAuthForm, 0
		m.oi[0].Focus()
		m.oi[1].Blur()
		return m, textinput.Blink
	}
	return m, nil
}

func (m model) updateOAuthForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.st, m.loginErr = stLogin, ""
		return m, nil
	case "tab", "down", "shift+tab", "up":
		m.oi[m.oFocus].Blur()
		m.oFocus = (m.oFocus + 1) % 2
		return m, m.oi[m.oFocus].Focus()
	case "enter":
		id, sec := strings.TrimSpace(m.oi[0].Value()), strings.TrimSpace(m.oi[1].Value())
		if id == "" || sec == "" {
			m.oi[m.oFocus].Blur()
			m.oFocus = 0
			if id != "" {
				m.oFocus = 1
			}
			m.loginErr = "client ID and client secret are required"
			return m, m.oi[m.oFocus].Focus()
		}
		sess, err := beginOAuth(id)
		if err != nil {
			m.loginErr = err.Error()
			return m, nil
		}
		m.sess, m.st, m.loginErr = sess, stOAuthWait, ""
		if err := openURL(sess.URL); err != nil {
			m.loginErr = "could not open the browser, open the URL below manually"
		}
		client := m.client
		return m, func() tea.Msg {
			o, err := sess.wait(client, id, sec)
			return oauthDoneMsg{o, err}
		}
	}
	var c tea.Cmd
	m.oi[m.oFocus], c = m.oi[m.oFocus].Update(msg)
	return m, c
}

func (m model) updateOAuthWait(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		if m.sess != nil {
			m.sess.Cancel()
			m.sess = nil
		}
		m.st, m.loginErr = stOAuthForm, ""
	}
	return m, nil
}
