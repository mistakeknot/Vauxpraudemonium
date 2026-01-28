package tui

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	pkgtui "github.com/mistakeknot/autarch/pkg/tui"
)

type uiConfig struct {
	Chat pkgtui.ChatSettings `toml:"chat"`
}

// DefaultChatSettings returns the default chat settings.
func DefaultChatSettings() pkgtui.ChatSettings {
	return pkgtui.DefaultChatSettings()
}

// LoadChatSettings loads chat settings from ~/.config/autarch/ui.toml.
func LoadChatSettings() (pkgtui.ChatSettings, error) {
	path, err := chatSettingsPath()
	if err != nil {
		return DefaultChatSettings(), err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return DefaultChatSettings(), nil
	}
	if err != nil {
		return DefaultChatSettings(), err
	}

	var cfg uiConfig
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return DefaultChatSettings(), err
	}
	if cfg.Chat == (pkgtui.ChatSettings{}) {
		return DefaultChatSettings(), nil
	}
	return cfg.Chat, nil
}

// SaveChatSettings writes chat settings to ~/.config/autarch/ui.toml.
func SaveChatSettings(settings pkgtui.ChatSettings) error {
	path, err := chatSettingsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	cfg := uiConfig{Chat: settings}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func chatSettingsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "autarch", "ui.toml"), nil
}
