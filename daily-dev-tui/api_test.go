package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type captured struct {
	method, path, query, auth, body string
}

func server(t *testing.T, status int, resp string) (*captured, func()) {
	t.Helper()
	c := &captured{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		*c = captured{r.Method, r.URL.Path, r.URL.RawQuery, r.Header.Get("Authorization"), string(b)}
		w.WriteHeader(status)
		w.Write([]byte(resp))
	}))
	old := baseURL
	baseURL = ts.URL
	return c, func() { baseURL = old; ts.Close() }
}

// Every GET endpoint: path params are substituted, query params are sent, bearer token is set.
func TestEveryEndpoint(t *testing.T) {
	for _, e := range endpoints {
		e := e
		t.Run(e.Group+"/"+e.Name, func(t *testing.T) {
			c, done := server(t, 200, `{"data":[]}`)
			defer done()
			vals := map[string]string{}
			wantPath := e.Path
			var wantQuery []string
			for _, f := range e.Fields {
				vals[f.Key] = "v " + f.Key
				if f.Path {
					wantPath = strings.ReplaceAll(wantPath, "{"+f.Key+"}", "v%20"+f.Key)
				} else {
					wantQuery = append(wantQuery, f.Key+"=v+"+f.Key)
				}
			}
			if _, err := doGet(http.DefaultClient, "tok", e, vals); err != nil {
				t.Fatal(err)
			}
			if c.method != "GET" {
				t.Errorf("method = %s", c.method)
			}
			if !strings.HasSuffix(c.path, strings.ReplaceAll(wantPath, "%20", " ")) {
				t.Errorf("path = %q, want suffix %q", c.path, wantPath)
			}
			if strings.ContainsAny(c.path, "{}") {
				t.Errorf("unresolved placeholder in %q", c.path)
			}
			for _, q := range wantQuery {
				if !strings.Contains(c.query, q) {
					t.Errorf("query %q missing %q", c.query, q)
				}
			}
			if c.auth != "Bearer tok" {
				t.Errorf("auth = %q", c.auth)
			}
		})
	}
}

func TestEmptyValuesAreOmitted(t *testing.T) {
	c, done := server(t, 200, `{}`)
	defer done()
	doGet(http.DefaultClient, "t", endpoints[7], map[string]string{"q": "go", "time": "  ", "limit": ""})
	if c.query != "q=go" {
		t.Errorf("query = %q", c.query)
	}
}

func TestDoGetErrors(t *testing.T) {
	_, done := server(t, 404, `{"error":"Not Found"}`)
	defer done()
	if _, err := doGet(http.DefaultClient, "t", endpoints[5], map[string]string{"id": "x"}); err == nil || !strings.Contains(err.Error(), "HTTP 404") {
		t.Errorf("err = %v", err)
	}
}

func TestDoGetEmptyBody(t *testing.T) {
	_, done := server(t, 200, ``)
	defer done()
	r, err := doGet(http.DefaultClient, "t", endpoints[0], nil)
	if err != nil || r.(map[string]any)["status"] != "success" {
		t.Errorf("r=%v err=%v", r, err)
	}
}

func TestBookmarkPost(t *testing.T) {
	c, done := server(t, 200, `{}`)
	defer done()
	m := model{client: http.DefaultClient, token: "tok"}
	msg := m.bookmark("abc")().(actionMsg)
	if msg.err != nil || msg.id != "abc" || msg.text != "bookmarked" {
		t.Fatalf("msg = %+v", msg)
	}
	if c.method != "POST" || c.path != "/bookmarks/" || c.auth != "Bearer tok" {
		t.Errorf("captured = %+v", c)
	}
	var body map[string][]string
	json.Unmarshal([]byte(c.body), &body)
	if len(body["postIds"]) != 1 || body["postIds"][0] != "abc" {
		t.Errorf("body = %s", c.body)
	}
}

func TestBookmarkError(t *testing.T) {
	_, done := server(t, 401, `{"error":"unauthorized"}`)
	defer done()
	m := model{client: http.DefaultClient, token: "bad"}
	if msg := m.bookmark("abc")().(actionMsg); msg.err == nil {
		t.Error("expected error")
	}
}

func TestFetchComments(t *testing.T) {
	c, done := server(t, 200, `{"data":[{"id":"1","content":"hi","author":{"username":"bob"}}]}`)
	defer done()
	m := model{client: http.DefaultClient, token: "tok"}
	msg := m.fetchComments("p1")().(commentsMsg)
	if msg.err != nil {
		t.Fatal(msg.err)
	}
	if c.path != "/posts/p1/comments" || !strings.Contains(c.query, "sort=newest") {
		t.Errorf("captured = %+v", c)
	}
	if out := renderComments(msg.res, 60); !strings.Contains(out, "bob") || !strings.Contains(out, "hi") {
		t.Errorf("render = %q", out)
	}
}

func TestEndpointIndexUsedForComments(t *testing.T) {
	if endpoints[6].Path != "/posts/{id}/comments" {
		t.Fatalf("endpoints[6] = %s; fetchComments depends on it", endpoints[6].Path)
	}
}

func TestNextCursor(t *testing.T) {
	r := map[string]any{"pagination": map[string]any{"hasNextPage": true, "cursor": "c1"}}
	if nextCursor(r) != "c1" {
		t.Error("expected c1")
	}
	r["pagination"] = map[string]any{"hasNextPage": false, "cursor": "c1"}
	if nextCursor(r) != "" || nextCursor("x") != "" {
		t.Error("expected empty")
	}
}

func TestSelectablePosts(t *testing.T) {
	res := map[string]any{"data": []any{
		map[string]any{"id": "1", "title": "a", "commentsPermalink": "https://daily.dev/posts/1"},
		map[string]any{"id": "2", "title": "notification"},
	}}
	if p := selectablePosts(res); len(p) != 1 || p[0]["id"] != "1" {
		t.Errorf("posts = %v", p)
	}
}

func TestRenderResultOffsetsAndBookmark(t *testing.T) {
	res := map[string]any{"data": []any{
		map[string]any{"id": "1", "title": "a", "commentsPermalink": "https://x"},
		map[string]any{"id": "2", "title": "b", "commentsPermalink": "https://y"},
	}}
	_, offs := renderResult(res, 80, 0, true, map[string]bool{"2": true})
	if len(offs) != 2 || offs[1][0] <= offs[0][0] {
		t.Errorf("offs = %v", offs)
	}
	if s, _ := renderResult(map[string]any{"data": []any{}}, 80, 0, false, nil); !strings.Contains(s, "No results") {
		t.Errorf("s = %q", s)
	}
}
