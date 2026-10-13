package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultThemeLoads(t *testing.T) {
	t.Setenv("DAILY_DEV_THEME", filepath.Join(t.TempDir(), "missing.conf"))
	loadTheme()
	if colors["primary"] != "135" || len(colors) != 12 {
		t.Fatalf("unexpected default colors: %v", colors)
	}
}

func TestCustomThemeOverrides(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.conf")
	os.WriteFile(p, []byte("# c\nprimary = #ff00aa\nerror=999x\nbogus=1\nlink = 10\n"), 0o644)
	t.Setenv("DAILY_DEV_THEME", p)
	loadTheme()
	if colors["primary"] != "#ff00aa" || colors["link"] != "10" {
		t.Fatalf("override failed: %v", colors)
	}
	if colors["error"] != "196" {
		t.Fatalf("invalid value should keep default, got %q", colors["error"])
	}
	if _, ok := colors["bogus"]; ok {
		t.Fatal("unknown key accepted")
	}
}

func TestNextThemeCycles(t *testing.T) {
	t.Setenv("DAILY_DEV_THEME", "default")
	loadTheme()
	names := themeNames()
	if names[0] != "default" || len(names) < 5 {
		t.Fatalf("unexpected themes: %v", names)
	}
	for _, want := range append(names[1:], "default") {
		if got := nextTheme(false); got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	}
	setTheme("nord")
	if colors["primary"] != "#88c0d0" {
		t.Fatalf("nord not applied: %v", colors)
	}
	setTheme("default")
}
