package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	Host      string `json:"host,omitempty"`
	Port      int    `json:"port,omitempty"`
	Autostart *bool  `json:"autostart,omitempty"`
	OpenUI    *bool  `json:"open_ui,omitempty"`
}

func defaultConfigPath() string {
	if runtime.GOOS == "windows" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, "WinShellBridge", "config.json")
		}
		if home, _ := os.UserHomeDir(); home != "" {
			return filepath.Join(home, "AppData", "Roaming", "WinShellBridge", "config.json")
		}
	}

	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		if home, _ := os.UserHomeDir(); home != "" {
			base = filepath.Join(home, ".config")
		}
	}
	if base == "" {
		return "WinShellBridge.config.json"
	}
	return filepath.Join(base, "WinShellBridge", "config.json")
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
