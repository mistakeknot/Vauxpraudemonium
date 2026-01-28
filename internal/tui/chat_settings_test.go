package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChatSettingsLoadDefaultsAndPersist(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, err := LoadChatSettings()
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if !cfg.AutoScroll || !cfg.ShowHistoryOnNewChat || !cfg.GroupMessages {
		t.Fatalf("expected defaults on")
	}

	cfg.AutoScroll = false
	cfg.GroupMessages = false
	if err := SaveChatSettings(cfg); err != nil {
		t.Fatalf("save settings: %v", err)
	}

	loaded, err := LoadChatSettings()
	if err != nil {
		t.Fatalf("reload settings: %v", err)
	}
	if loaded.AutoScroll || loaded.GroupMessages {
		t.Fatalf("expected persisted values")
	}

	path := filepath.Join(dir, "autarch", "ui.toml")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected settings file: %v", err)
	}
}
