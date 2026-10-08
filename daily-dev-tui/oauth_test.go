package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func tokenServer(t *testing.T, handler http.HandlerFunc) func() {
	ts := httptest.NewServer(handler)
	old := tokenURL
	tokenURL = ts.URL
	return func() { tokenURL = old; ts.Close() }
}

func TestPKCEChallenge(t *testing.T) {
	// RFC 7636 appendix B test vector
	if got := pkceChallenge("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"); got != "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM" {
		t.Errorf("challenge = %s", got)
	}
}

func TestBeginOAuthURL(t *testing.T) {
	old := redirectAddr
	redirectAddr = "127.0.0.1:0"
	defer func() { redirectAddr = old }()
	s, err := beginOAuth("cid")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Cancel()
	u, _ := url.Parse(s.URL)
	q := u.Query()
	for k, want := range map[string]string{
		"response_type": "code", "client_id": "cid", "code_challenge_method": "S256",
		"resource": oauthResource, "scope": oauthScope, "state": s.state,
		"code_challenge": pkceChallenge(s.verifier),
	} {
		if q.Get(k) != want {
			t.Errorf("%s = %q, want %q", k, q.Get(k), want)
		}
	}
	if !strings.HasPrefix(s.URL, authorizeURL) {
		t.Errorf("url = %s", s.URL)
	}
}

func TestOAuthFullFlow(t *testing.T) {
	var form url.Values
	done := tokenServer(t, func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		form = r.PostForm
		w.Write([]byte(`{"access_token":"AT","refresh_token":"RT","expires_in":3600,"scope":"read write"}`))
	})
	defer done()
	old := redirectAddr
	redirectAddr = "127.0.0.1:18765"
	defer func() { redirectAddr = old }()

	s, err := beginOAuth("cid")
	if err != nil {
		t.Fatal(err)
	}
	type res struct {
		o   *oauthState
		err error
	}
	ch := make(chan res, 1)
	go func() { o, err := s.wait(http.DefaultClient, "cid", "sec"); ch <- res{o, err} }()
	time.Sleep(100 * time.Millisecond)
	resp, err := http.Get(redirectURI() + "?code=CODE&state=" + s.state)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	r := <-ch
	if r.err != nil {
		t.Fatal(r.err)
	}
	if r.o.access != "AT" || r.o.refresh != "RT" || r.o.perm != "read write" {
		t.Errorf("state = %+v", r.o)
	}
	for k, want := range map[string]string{
		"grant_type": "authorization_code", "code": "CODE", "client_id": "cid", "client_secret": "sec",
		"code_verifier": s.verifier, "redirect_uri": redirectURI(), "resource": oauthResource,
	} {
		if form.Get(k) != want {
			t.Errorf("form %s = %q, want %q", k, form.Get(k), want)
		}
	}
}

func TestOAuthStateMismatch(t *testing.T) {
	old := redirectAddr
	redirectAddr = "127.0.0.1:18766"
	defer func() { redirectAddr = old }()
	s, _ := beginOAuth("cid")
	ch := make(chan error, 1)
	go func() { _, err := s.wait(http.DefaultClient, "cid", "sec"); ch <- err }()
	time.Sleep(100 * time.Millisecond)
	resp, _ := http.Get(redirectURI() + "?code=CODE&state=wrong")
	resp.Body.Close()
	if err := <-ch; err == nil || !strings.Contains(err.Error(), "state mismatch") {
		t.Errorf("err = %v", err)
	}
}

func TestOAuthCancel(t *testing.T) {
	old := redirectAddr
	redirectAddr = "127.0.0.1:18767"
	defer func() { redirectAddr = old }()
	s, _ := beginOAuth("cid")
	ch := make(chan error, 1)
	go func() { _, err := s.wait(http.DefaultClient, "cid", "sec"); ch <- err }()
	time.Sleep(100 * time.Millisecond)
	s.Cancel()
	if err := <-ch; err == nil {
		t.Error("expected error after cancel")
	}
}

func TestTokenRefreshRotates(t *testing.T) {
	var calls int32
	var form url.Values
	done := tokenServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		r.ParseForm()
		form = r.PostForm
		w.Write([]byte(`{"access_token":"AT2","refresh_token":"RT2","expires_in":3600}`))
	})
	defer done()
	o := &oauthState{clientID: "cid", secret: "sec", access: "AT1", refresh: "RT1", exp: time.Now().Add(-time.Minute)}
	if got := o.Token(http.DefaultClient); got != "AT2" {
		t.Errorf("token = %s", got)
	}
	if o.refresh != "RT2" || form.Get("grant_type") != "refresh_token" || form.Get("refresh_token") != "RT1" {
		t.Errorf("refresh = %s form = %v", o.refresh, form)
	}
	o.Token(http.DefaultClient)
	if calls != 1 {
		t.Errorf("token fetched again while valid, calls = %d", calls)
	}
}

func TestTokenRefreshFailureKeepsOld(t *testing.T) {
	done := tokenServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`{"error":"invalid_grant"}`))
	})
	defer done()
	o := &oauthState{access: "AT1", refresh: "RT1", exp: time.Now().Add(-time.Minute)}
	if got := o.Token(http.DefaultClient); got != "AT1" {
		t.Errorf("token = %s", got)
	}
}

func TestBearerUsesOAuthToken(t *testing.T) {
	m := model{token: "pat"}
	if m.bearer() != "pat" {
		t.Error("expected personal token")
	}
	m.oauth = &oauthState{access: "oauth-at"}
	if m.bearer() != "oauth-at" {
		t.Error("expected oauth token")
	}
}

func TestLoginSelector(t *testing.T) {
	m := initialModel()
	m.st = stLogin
	if m.loginSel != 0 {
		t.Fatal("default should be bearer")
	}
	nm, _ := m.updateLogin(tea.KeyMsg{Type: tea.KeyDown})
	m = nm.(model)
	if m.loginSel != 1 {
		t.Errorf("sel = %d", m.loginSel)
	}
	nm, _ = m.updateLogin(tea.KeyMsg{Type: tea.KeyEnter})
	if nm.(model).st != stOAuthForm {
		t.Errorf("st = %v", nm.(model).st)
	}
	m.loginSel = 0
	nm, _ = m.updateLogin(tea.KeyMsg{Type: tea.KeyEnter})
	if nm.(model).st != stToken {
		t.Errorf("st = %v", nm.(model).st)
	}
}

func TestOAuthFormRequiresBoth(t *testing.T) {
	t.Setenv("DAILY_DEV_CLIENT_ID", "")
	t.Setenv("DAILY_DEV_CLIENT_SECRET", "")
	m := initialModel()
	m.st = stOAuthForm
	nm, _ := m.updateOAuthForm(tea.KeyMsg{Type: tea.KeyEnter})
	if got := nm.(model); got.st != stOAuthForm || got.loginErr == "" {
		t.Errorf("st=%v err=%q", got.st, got.loginErr)
	}
}

func TestEscGoesBackToLogin(t *testing.T) {
	m := initialModel()
	m.st = stOAuthForm
	nm, _ := m.updateOAuthForm(tea.KeyMsg{Type: tea.KeyEsc})
	if nm.(model).st != stLogin {
		t.Error("esc should return to login")
	}
}
