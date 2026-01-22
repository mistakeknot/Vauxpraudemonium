package tmux

import (
	"fmt"
	"regexp"
	"strings"
)

type Runner interface {
	Run(name string, args ...string) error
}

type Session struct {
	ID      string
	Workdir string
	LogPath string
}

func StartSession(r Runner, s Session) error {
	if err := validateLogPath(s.LogPath); err != nil {
		return err
	}
	if err := r.Run("tmux", "new-session", "-d", "-s", s.ID, "-c", s.Workdir); err != nil {
		return err
	}
	cmd := "cat >> " + shellQuote(s.LogPath)
	return r.Run("tmux", "pipe-pane", "-t", s.ID, "-o", cmd)
}

func StopSession(r Runner, id string) error {
	return r.Run("tmux", "kill-session", "-t", id)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

var logPathPattern = regexp.MustCompile(`^[A-Za-z0-9 _./'-]+$`)

func validateLogPath(path string) error {
	if !logPathPattern.MatchString(path) {
		return fmt.Errorf("invalid log path: %q", path)
	}
	return nil
}
