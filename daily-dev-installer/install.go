package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const releaseURL = "https://github.com/F4NT0/daily-dev-tui/releases/latest/download/daily-dev-tui.exe"

var (
	localInstall bool
	localRepo    string
)

// findLocalRepo looks for the daily-dev-tui source next to the installer or the working directory.
func findLocalRepo() string {
	self, _ := os.Executable()
	wd, _ := os.Getwd()
	for _, base := range []string{wd, filepath.Dir(self), filepath.Dir(wd), filepath.Dir(filepath.Dir(self))} {
		for _, c := range []string{base, filepath.Join(base, "daily-dev-tui")} {
			if validRepo(c) {
				return c
			}
		}
	}
	return ""
}

func validRepo(dir string) bool {
	b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	return err == nil && strings.Contains(string(b), "module dailydevtui")
}

func installDir() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "daily-dev")
}

func psEnv(script string) (string, error) {
	out, err := exec.Command("powershell", "-NoProfile", "-Command", script).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func stepCreateDir() error { return os.MkdirAll(installDir(), 0o755) }

func stepCopyTUI() error {
	dst := filepath.Join(installDir(), tuiExe)
	if localInstall {
		out, err := buildLocal(dst)
		if err != nil {
			return fmt.Errorf("go build failed: %v: %s", err, strings.TrimSpace(string(out)))
		}
		return nil
	}
	resp, err := http.Get(releaseURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s (is there a release with daily-dev-tui.exe?)", resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o755)
}

func stepCopyLauncher() error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	src, err := os.Open(self)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(filepath.Join(installDir(), "daily-dev.exe"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

func stepAddPath() error {
	d := installDir()
	_, err := psEnv(fmt.Sprintf(`$d='%s'; $p=[Environment]::GetEnvironmentVariable('Path','User'); if(-not $p){$p=''}; if(($p -split ';') -notcontains $d){[Environment]::SetEnvironmentVariable('Path',($p.TrimEnd(';')+';'+$d).TrimStart(';'),'User')}`, d))
	return err
}

func stepRemovePath() error {
	d := installDir()
	_, err := psEnv(fmt.Sprintf(`$d='%s'; $p=[Environment]::GetEnvironmentVariable('Path','User'); if($p){[Environment]::SetEnvironmentVariable('Path',(($p -split ';' | Where-Object { $_ -and $_ -ne $d }) -join ';'),'User')}`, d))
	return err
}

// Running exe can't delete itself on Windows; defer removal via a detached cmd.
func stepRemoveFiles() error {
	d := installDir()
	if _, err := os.Stat(d); err != nil {
		return nil
	}
	return exec.Command("cmd", "/C", fmt.Sprintf(`start "" /B cmd /C "ping -n 3 127.0.0.1 >nul & rmdir /S /Q "%s""`, d)).Start()
}

func buildLocal(dst string) ([]byte, error) {
	cmd := exec.Command("go", "build", "-o", dst, ".")
	cmd.Dir = localRepo
	return cmd.CombinedOutput()
}
