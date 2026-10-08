package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	authorizeURL = "https://api.daily.dev/auth/oauth2/authorize"
	tokenURL     = "https://api.daily.dev/auth/oauth2/token"
	oauthScope   = "openid profile offline_access read write"
	redirectAddr = "127.0.0.1:8765"
)

const oauthResource = "https://api.daily.dev/public/v1"

func redirectURI() string { return "http://" + redirectAddr + "/callback" }

type oauthState struct {
	mu                    sync.Mutex
	clientID, secret      string
	access, refresh, perm string
	exp                   time.Time
}

type tokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

func randB64(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func pkceChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func tokenRequest(c *http.Client, form url.Values) (*tokenResp, error) {
	resp, err := c.PostForm(tokenURL, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("token endpoint HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var t tokenResp
	if err := json.Unmarshal(body, &t); err != nil || t.AccessToken == "" {
		return nil, fmt.Errorf("invalid token response: %s", strings.TrimSpace(string(body)))
	}
	return &t, nil
}

func (o *oauthState) apply(t *tokenResp) {
	o.access, o.perm = t.AccessToken, t.Scope
	if t.RefreshToken != "" {
		o.refresh = t.RefreshToken
	}
	o.exp = time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)
}

// Token returns a valid access token, refreshing it (refresh tokens rotate) when it is about to expire.
func (o *oauthState) Token(c *http.Client) string {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.refresh != "" && !o.exp.IsZero() && time.Until(o.exp) < 30*time.Second {
		t, err := tokenRequest(c, url.Values{
			"grant_type":    {"refresh_token"},
			"refresh_token": {o.refresh},
			"client_id":     {o.clientID},
			"client_secret": {o.secret},
			"resource":      {oauthResource},
		})
		if err == nil {
			o.apply(t)
		}
	}
	return o.access
}

type oauthSession struct {
	URL      string
	verifier string
	state    string
	ln       net.Listener
	cancel   context.CancelFunc
	ctx      context.Context
}

type oauthDoneMsg struct {
	st  *oauthState
	err error
}

// beginOAuth opens the loopback listener and builds the authorize URL (does not block).
func beginOAuth(clientID string) (*oauthSession, error) {
	ln, err := net.Listen("tcp", redirectAddr)
	if err != nil {
		return nil, fmt.Errorf("cannot listen on %s: %w", redirectAddr, err)
	}
	s := &oauthSession{ln: ln, verifier: randB64(32), state: randB64(16)}
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 3*time.Minute)
	q := url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI()},
		"scope":                 {oauthScope},
		"state":                 {s.state},
		"code_challenge":        {pkceChallenge(s.verifier)},
		"code_challenge_method": {"S256"},
		"resource":              {oauthResource},
	}
	s.URL = authorizeURL + "?" + q.Encode()
	return s, nil
}

func (s *oauthSession) Cancel() {
	s.cancel()
	s.ln.Close()
}

// wait blocks until the browser is redirected back (or timeout/cancel), then exchanges the code.
func (s *oauthSession) wait(c *http.Client, clientID, secret string) (*oauthState, error) {
	type cb struct {
		code string
		err  error
	}
	ch := make(chan cb, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		var res cb
		switch {
		case q.Get("state") != s.state:
			res.err = errors.New("state mismatch")
		case q.Get("error") != "":
			res.err = fmt.Errorf("authorization denied: %s %s", q.Get("error"), q.Get("error_description"))
		case q.Get("code") == "":
			res.err = errors.New("no code in callback")
		default:
			res.code = q.Get("code")
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if res.err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "<h2>Sign-in failed</h2><p>You can close this tab and go back to the terminal.</p>")
		} else {
			fmt.Fprint(w, "<h2>Signed in to daily-dev</h2><p>You can close this tab and go back to the terminal.</p>")
		}
		select {
		case ch <- res:
		default:
		}
	})
	srv := &http.Server{Handler: mux}
	go srv.Serve(s.ln)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
		s.cancel()
	}()
	var res cb
	select {
	case res = <-ch:
	case <-s.ctx.Done():
		return nil, errors.New("sign-in timed out or was cancelled")
	}
	if res.err != nil {
		return nil, res.err
	}
	t, err := tokenRequest(c, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {res.code},
		"redirect_uri":  {redirectURI()},
		"client_id":     {clientID},
		"client_secret": {secret},
		"code_verifier": {s.verifier},
		"resource":      {oauthResource},
	})
	if err != nil {
		return nil, err
	}
	o := &oauthState{clientID: clientID, secret: secret}
	o.apply(t)
	return o, nil
}
