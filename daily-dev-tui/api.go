package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const baseURL = "https://api.daily.dev/public/v1"

func newClient() *http.Client {
	c := &http.Client{Timeout: 30 * time.Second}
	if os.Getenv("DAILY_DEV_INSECURE") == "1" {
		c.Transport = &http.Transport{Proxy: http.ProxyFromEnvironment, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}
	return c
}

func doGet(c *http.Client, token string, e Endpoint, vals map[string]string) (any, error) {
	path := e.Path
	q := url.Values{}
	for _, f := range e.Fields {
		v := strings.TrimSpace(vals[f.Key])
		if v == "" {
			continue
		}
		if f.Path {
			path = strings.ReplaceAll(path, "{"+f.Key+"}", url.PathEscape(v))
		} else {
			q.Set(f.Key, v)
		}
	}
	u := baseURL + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out any
	if len(body) == 0 {
		return map[string]any{"status": "success"}, nil
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}
