package tui

import (
	pkgtui "github.com/mistakeknot/autarch/pkg/tui"
)

// View is an alias to pkg/tui.View for backward compatibility.
// New code should import from pkg/tui directly.
type View = pkgtui.View

// HelpBinding is an alias to pkg/tui.HelpBinding for backward compatibility.
type HelpBinding = pkgtui.HelpBinding

// FullHelpProvider is an alias to pkg/tui.FullHelpProvider for backward compatibility.
type FullHelpProvider = pkgtui.FullHelpProvider

// Command is an alias to pkg/tui.Command for backward compatibility.
type Command = pkgtui.Command

// CommandProvider is an alias to pkg/tui.CommandProvider for backward compatibility.
type CommandProvider = pkgtui.CommandProvider
