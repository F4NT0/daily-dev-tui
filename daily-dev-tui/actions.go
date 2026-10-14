package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

type noBrowserErr struct{ url string }

func (e *noBrowserErr) Error() string {
	return fmt.Sprintf("no browser found (install xdg-utils or set $BROWSER); open manually: %s", e.url)
}

func openURL(u string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start()
	case "darwin":
		return exec.Command("open", u).Start()
	}
	cands := [][]string{{"xdg-open"}, {"gio", "open"}, {"wslview"}, {"sensible-browser"}, {"x-www-browser"},
		{"gnome-open"}, {"kde-open"}, {"firefox"}, {"google-chrome"}, {"chromium"}, {"chromium-browser"}, {"brave-browser"}}
	if b := os.Getenv("BROWSER"); b != "" {
		cands = append([][]string{{b}}, cands...)
	}
	for _, c := range cands {
		if p, err := exec.LookPath(c[0]); err == nil {
			return exec.Command(p, append(c[1:], u)...).Start()
		}
	}
	return &noBrowserErr{u}
}
