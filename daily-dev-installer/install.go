package main

import (
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed payload/daily-dev-tui.exe
var payload embed.FS

func installDir() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "daily-dev")
}

func psEnv(script string) (string, error) {
	out, err := exec.Command("powershell", "-NoProfile", "-Command", script).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func stepCreateDir() error { return os.MkdirAll(installDir(), 0o755) }

func stepCopyTUI() error {
	b, err := payload.ReadFile("payload/daily-dev-tui.exe")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(installDir(), tuiExe), b, 0o755)
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
