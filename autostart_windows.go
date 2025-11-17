//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const autostartValueName = "WinShellBridge"

// EnsureAutoStart registers the app in HKCU Run so it starts after the user logs in.
func EnsureAutoStart() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable: %w", err)
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	key, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Run`,
		registry.SET_VALUE,
	)
	if err != nil {
		return fmt.Errorf("open Run key: %w", err)
	}
	defer key.Close()

	// Quote the path to survive spaces. No extra args to keep it simple.
	value := fmt.Sprintf("\"%s\"", exe)
	if err := key.SetStringValue(autostartValueName, value); err != nil {
		return fmt.Errorf("write Run entry: %w", err)
	}
	return nil
}
