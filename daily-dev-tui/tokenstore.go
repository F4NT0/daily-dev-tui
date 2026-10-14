package main

import (
	"os"
	"path/filepath"
	"strings"
)

const tokenFileName = "token.conf"

var autoSignInLabel = "Auto Sign In (Token already saved locally)"

// tokenFilePath: DAILY_DEV_TOKEN_FILE, else next to the program (or the repository when running from it).
func tokenFilePath() string {
	if p := os.Getenv("DAILY_DEV_TOKEN_FILE"); p != "" {
		return p
	}
	for _, d := range themeDirs() {
		if st, err := os.Stat(d); err == nil && st.IsDir() {
			return filepath.Join(filepath.Dir(d), tokenFileName)
		}
	}
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), tokenFileName)
	}
	return tokenFileName
}

func loadSavedToken() string {
	b, err := os.ReadFile(tokenFilePath())
	if err != nil {
		return ""
	}
	for _, l := range strings.Split(string(b), "\n") {
		if k, v, ok := strings.Cut(strings.TrimSpace(l), "="); ok && strings.TrimSpace(k) == "token" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func saveToken(t string) error {
	body := "# daily-dev saved Bearer token. Keep this file private and never share or commit it.\ntoken = " + t + "\n"
	return os.WriteFile(tokenFilePath(), []byte(body), 0o600)
}

func loginOpts(saved string) []string {
	o := []string{"Bearer token (personal access token)", "OAuth (sign in with daily.dev)"}
	if saved != "" {
		o = append(o, autoSignInLabel)
	}
	return o
}
