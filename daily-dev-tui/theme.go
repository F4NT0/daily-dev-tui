package main

import (
	"embed"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

//go:embed themes/*.conf
var themeFS embed.FS

const stateFile = ".theme"

var (
	colors    = map[string]string{}
	colorRe   = regexp.MustCompile(`^(\d{1,3}|#[0-9a-fA-F]{6}|#[0-9a-fA-F]{3})$`)
	themeName = "default"
	titleSt   lipgloss.Style
	subSt     lipgloss.Style
	metaSt    lipgloss.Style
	keySt     lipgloss.Style
	linkSt    lipgloss.Style
	tagSt     lipgloss.Style
	cardSt    lipgloss.Style
	errSt     lipgloss.Style
)

func col(name string) lipgloss.Color { return lipgloss.Color(colors[name]) }

func parseConf(s string, strict bool) {
	for _, l := range strings.Split(s, "\n") {
		l = strings.TrimSpace(l)
		if l == "" || l[0] == '#' || l[0] == ';' {
			continue
		}
		k, v, ok := strings.Cut(l, "=")
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if i := strings.IndexAny(v, " \t"); i >= 0 {
			v = v[:i]
		}
		if _, known := colors[k]; ok && (known || !strict) && colorRe.MatchString(v) {
			colors[k] = v
		}
	}
}

// themeDirs: the themes directory next to the program, then ./themes (running from the repository).
func themeDirs() []string {
	var d []string
	if exe, err := os.Executable(); err == nil {
		d = append(d, filepath.Join(filepath.Dir(exe), "themes"))
	}
	return append(d, "themes")
}

func isPathName(n string) bool { return strings.ContainsAny(n, `/\`) || strings.HasSuffix(n, ".conf") }

func readTheme(name string) (string, bool) {
	if isPathName(name) {
		b, err := os.ReadFile(name)
		return string(b), err == nil
	}
	for _, d := range themeDirs() {
		if b, err := os.ReadFile(filepath.Join(d, name+".conf")); err == nil {
			return string(b), true
		}
	}
	b, err := themeFS.ReadFile("themes/" + name + ".conf")
	return string(b), err == nil
}

func themeNames() []string {
	set := map[string]bool{}
	ents, _ := themeFS.ReadDir("themes")
	for _, e := range ents {
		set[strings.TrimSuffix(e.Name(), ".conf")] = true
	}
	for _, d := range themeDirs() {
		fs, _ := filepath.Glob(filepath.Join(d, "*.conf"))
		for _, f := range fs {
			set[strings.TrimSuffix(filepath.Base(f), ".conf")] = true
		}
	}
	var out []string
	for n := range set {
		if n != "default" {
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return append([]string{"default"}, out...)
}

func setTheme(name string) bool {
	body, ok := readTheme(name)
	if !ok {
		return false
	}
	for k := range colors {
		delete(colors, k)
	}
	def, _ := themeFS.ReadFile("themes/default.conf")
	parseConf(string(def), false)
	parseConf(body, true)
	themeName = strings.TrimSuffix(filepath.Base(name), ".conf")
	applyTheme()
	return true
}

func savedTheme() string {
	for _, d := range themeDirs() {
		if b, err := os.ReadFile(filepath.Join(d, stateFile)); err == nil {
			return strings.TrimSpace(string(b))
		}
	}
	return ""
}

func saveTheme(name string) {
	for _, d := range themeDirs() {
		if st, err := os.Stat(d); err == nil && st.IsDir() {
			if os.WriteFile(filepath.Join(d, stateFile), []byte(name+"\n"), 0o644) == nil {
				return
			}
		}
	}
}

func nextTheme(save bool) string {
	names := themeNames()
	i := 0
	for j, n := range names {
		if n == themeName {
			i = (j + 1) % len(names)
		}
	}
	if setTheme(names[i]) && save {
		saveTheme(names[i])
	}
	return themeName
}

func loadTheme() {
	for _, n := range []string{os.Getenv("DAILY_DEV_THEME"), savedTheme(), "default"} {
		if n != "" && setTheme(n) {
			return
		}
	}
}

func applyTheme() {
	titleSt = lipgloss.NewStyle().Bold(true).Foreground(col("primary"))
	subSt = lipgloss.NewStyle().Foreground(col("secondary"))
	metaSt = lipgloss.NewStyle().Foreground(col("muted"))
	keySt = lipgloss.NewStyle().Foreground(col("key"))
	linkSt = lipgloss.NewStyle().Foreground(col("link")).Underline(true)
	tagSt = lipgloss.NewStyle().Foreground(col("tag_fg")).Background(col("tag_bg")).Padding(0, 1)
	cardSt = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(col("border")).Padding(0, 1)
	errSt = lipgloss.NewStyle().Foreground(col("error")).Bold(true)
}

func init() { loadTheme() }
