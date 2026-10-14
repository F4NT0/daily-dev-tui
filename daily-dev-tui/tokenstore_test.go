package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadToken(t *testing.T) {
	p := filepath.Join(t.TempDir(), "token.conf")
	t.Setenv("DAILY_DEV_TOKEN_FILE", p)
	if loadSavedToken() != "" || len(loginOpts(loadSavedToken())) != 2 {
		t.Fatal("expected no saved token")
	}
	if err := saveToken("dda_secret"); err != nil {
		t.Fatal(err)
	}
	if got := loadSavedToken(); got != "dda_secret" {
		t.Fatalf("got %q", got)
	}
	o := loginOpts(loadSavedToken())
	if len(o) != 3 || o[2] != autoSignInLabel {
		t.Fatalf("options: %v", o)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteSavedTokenWithD(t *testing.T) {
	p := filepath.Join(t.TempDir(), "token.conf")
	t.Setenv("DAILY_DEV_TOKEN_FILE", p)
	saveToken("abc")
	m := model{saved: loadSavedToken(), loginSel: 2, st: stLogin}
	nm, _ := m.updateLogin(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = nm.(model)
	if _, err := os.Stat(p); !os.IsNotExist(err) || m.saved != "" || m.loginSel != 0 || m.loginMsg == "" {
		t.Fatalf("token not deleted: %+v", m)
	}
	saveToken("abc")
	m = model{saved: "abc", loginSel: 0}
	m.updateLogin(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if loadSavedToken() != "abc" {
		t.Fatal("d must only delete when the saved-token option is selected")
	}
}
