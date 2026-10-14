package main

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRecordRateAndInfo(t *testing.T) {
	h := http.Header{}
	h.Set("x-ratelimit-limit", "60")
	h.Set("x-ratelimit-remaining", "42")
	h.Set("x-ratelimit-reset", "30")
	recordRate(h)
	m := model{user: "alice", w: 140}
	s := m.infoText()
	for _, want := range []string{"alice", "42", "60/min", "Resets in"} {
		if !strings.Contains(s, want) {
			t.Fatalf("%q missing in %q", want, s)
		}
	}
	if rate.reset.Before(time.Now()) {
		t.Fatal("reset should be in the future")
	}
	if !strings.Contains(m.headerView(), "alice") {
		t.Fatal("header missing user")
	}
}

func TestUserFromProfileAndFailure(t *testing.T) {
	if userFromProfile(map[string]any{"data": map[string]any{"username": "bob"}}) != "bob" || userFromProfile(map[string]any{"name": "Al"}) != "Al" {
		t.Fatal("profile parsing failed")
	}
	if !isFailure("request failed") || !isFailure("query is required") || isFailure("bookmarked") {
		t.Fatal("failure classification wrong")
	}
}
