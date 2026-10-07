package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleSt = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("135"))
	subSt   = lipgloss.NewStyle().Foreground(lipgloss.Color("111"))
	metaSt  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	keySt   = lipgloss.NewStyle().Foreground(lipgloss.Color("178"))
	linkSt  = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Underline(true)
	tagSt   = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("141")).Padding(0, 1)
	cardSt  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1)
	errSt   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
)

var titleKeys = []string{"title", "name", "message", "username", "id", "value"}
var subKeys = []string{"subtitle", "source.name", "author.name", "author.username", "user.name", "type"}
var summaryKeys = []string{"summary", "tldr", "description", "content", "bio"}
var statKeys = []string{"upvotes", "numUpvotes", "comments", "numComments", "views", "readTime", "createdAt", "publishedAt", "startedAt", "endedAt"}

func isLink(v any) (string, bool) {
	s, ok := v.(string)
	return s, ok && (strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://"))
}

// flatten scalars (1 nested level) into dotted keys; arrays of scalars are joined.
func flatten(m map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range m {
		switch t := v.(type) {
		case map[string]any:
			for k2, v2 := range t {
				switch v2.(type) {
				case map[string]any, []any:
				default:
					out[k+"."+k2] = v2
				}
			}
		case []any:
			var parts []string
			for _, e := range t {
				switch x := e.(type) {
				case map[string]any, []any:
				default:
					parts = append(parts, fmt.Sprint(x))
				}
			}
			if len(parts) > 0 {
				out[k] = parts
			}
		default:
			out[k] = v
		}
	}
	return out
}

func scalar(v any) string {
	switch t := v.(type) {
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	case nil:
		return "-"
	}
	return fmt.Sprint(v)
}

func renderItem(m map[string]any, w int) string {
	f := flatten(m)
	used := map[string]bool{}
	first := func(keys []string) (string, string) {
		for _, k := range keys {
			if v, ok := f[k]; ok {
				if s, isS := v.(string); isS && s == "" {
					continue
				}
				if _, isArr := v.([]string); isArr {
					continue
				}
				used[k] = true
				return k, scalar(v)
			}
		}
		return "", ""
	}
	inner := w - 4
	wrap := lipgloss.NewStyle().Width(inner)
	var b []string

	_, title := first(titleKeys)
	if title == "" {
		title = "(item)"
	}
	b = append(b, titleSt.Width(inner).Render(title))
	if _, sub := first(subKeys); sub != "" {
		b = append(b, subSt.Width(inner).Render(sub))
	}
	if _, sum := first(summaryKeys); sum != "" {
		if len([]rune(sum)) > 400 {
			sum = string([]rune(sum)[:400]) + "…"
		}
		b = append(b, wrap.Render(sum))
	}

	var stats []string
	for _, k := range statKeys {
		if v, ok := f[k]; ok && !used[k] {
			used[k] = true
			stats = append(stats, fmt.Sprintf("%s %s", k, scalar(v)))
		}
	}
	if len(stats) > 0 {
		b = append(b, metaSt.Width(inner).Render(strings.Join(stats, "  •  ")))
	}

	if tags, ok := f["tags"].([]string); ok {
		used["tags"] = true
		var ts []string
		for _, t := range tags {
			ts = append(ts, tagSt.Render(t))
		}
		b = append(b, lipgloss.NewStyle().Width(inner).Render(strings.Join(ts, " ")))
	}

	keys := make([]string, 0, len(f))
	for k := range f {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var links, rest []string
	for _, k := range keys {
		if used[k] {
			continue
		}
		v := f[k]
		if arr, ok := v.([]string); ok {
			rest = append(rest, keySt.Render(k+": ")+strings.Join(arr, ", "))
			continue
		}
		if strings.Contains(strings.ToLower(k), "image") || strings.Contains(strings.ToLower(k), "avatar") || strings.Contains(strings.ToLower(k), "cover") {
			continue
		}
		if l, ok := isLink(v); ok {
			links = append(links, keySt.Render(k+": ")+linkSt.Render(l))
			continue
		}
		if s := scalar(v); s != "" && s != "-" && !strings.HasSuffix(k, "image") && !strings.HasSuffix(k, "Image") {
			rest = append(rest, keySt.Render(k+": ")+s)
		}
	}
	if id, ok := m["id"].(string); ok && len(links) == 0 && (m["title"] != nil) {
		links = append(links, keySt.Render("daily.dev: ")+linkSt.Render("https://app.daily.dev/posts/"+id))
	}
	for _, l := range links {
		b = append(b, wrap.Render(l))
	}
	if len(rest) > 0 {
		b = append(b, wrap.Render(metaSt.Render(strings.Join(rest, "\n"))))
	}
	return cardSt.Width(w - 2).Render(strings.Join(b, "\n"))
}

func renderResult(res any, w int) string {
	if w < 20 {
		w = 20
	}
	var items []any
	var extra []string
	switch t := res.(type) {
	case []any:
		items = t
	case map[string]any:
		if d, ok := t["data"]; ok {
			switch dd := d.(type) {
			case []any:
				items = dd
			case map[string]any:
				items = []any{dd}
			}
			if p, ok := t["pagination"].(map[string]any); ok {
				for k, v := range p {
					extra = append(extra, fmt.Sprintf("%s=%s", k, scalar(v)))
				}
				sort.Strings(extra)
			}
		} else {
			items = []any{t}
		}
	default:
		return scalar(res)
	}
	if len(items) == 0 {
		return metaSt.Render("No results.")
	}
	out := []string{subSt.Render(fmt.Sprintf("%d result(s)", len(items)))}
	for _, it := range items {
		if m, ok := it.(map[string]any); ok {
			out = append(out, renderItem(m, w))
		} else {
			out = append(out, cardSt.Width(w-2).Render(scalar(it)))
		}
	}
	if len(extra) > 0 {
		out = append(out, metaSt.Render("pagination: "+strings.Join(extra, "  ")))
	}
	return strings.Join(out, "\n")
}

func nextCursor(res any) string {
	if m, ok := res.(map[string]any); ok {
		if p, ok := m["pagination"].(map[string]any); ok {
			if hn, _ := p["hasNextPage"].(bool); hn {
				s, _ := p["cursor"].(string)
				return s
			}
		}
	}
	return ""
}
