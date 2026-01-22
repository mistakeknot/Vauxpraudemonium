package tui

import (
	"strings"

	sharedtui "github.com/mistakeknot/vauxpraudemonium/pkg/tui"
)

func renderHelpOverlay() string {
	lines := []string{
		sharedtui.TitleStyle.Render("Help"),
		sharedtui.HelpKeyStyle.Render("j/k") + sharedtui.HelpDescStyle.Render(": move  ") +
			sharedtui.HelpKeyStyle.Render("enter") + sharedtui.HelpDescStyle.Render(": toggle group  ") +
			sharedtui.HelpKeyStyle.Render("/") + sharedtui.HelpDescStyle.Render(": search"),
		sharedtui.HelpKeyStyle.Render("n") + sharedtui.HelpDescStyle.Render(": new PRD  ") +
			sharedtui.HelpKeyStyle.Render("g") + sharedtui.HelpDescStyle.Render(": interview"),
		sharedtui.HelpKeyStyle.Render("r") + sharedtui.HelpDescStyle.Render(": research  ") +
			sharedtui.HelpKeyStyle.Render("p") + sharedtui.HelpDescStyle.Render(": suggestions  ") +
			sharedtui.HelpKeyStyle.Render("s") + sharedtui.HelpDescStyle.Render(": review"),
		sharedtui.HelpKeyStyle.Render("a") + sharedtui.HelpDescStyle.Render(": archive  ") +
			sharedtui.HelpKeyStyle.Render("d") + sharedtui.HelpDescStyle.Render(": delete  ") +
			sharedtui.HelpKeyStyle.Render("u") + sharedtui.HelpDescStyle.Render(": undo  ") +
			sharedtui.HelpKeyStyle.Render("h") + sharedtui.HelpDescStyle.Render(": archived"),
		sharedtui.HelpKeyStyle.Render("[ ]") + sharedtui.HelpDescStyle.Render(": interview prev/next"),
		sharedtui.HelpKeyStyle.Render("?") + sharedtui.HelpDescStyle.Render(": help  ") +
			sharedtui.HelpKeyStyle.Render("`") + sharedtui.HelpDescStyle.Render(": tutorial  ") +
			sharedtui.HelpKeyStyle.Render("q") + sharedtui.HelpDescStyle.Render(": quit"),
		sharedtui.HelpDescStyle.Render("Esc: close"),
	}
	return strings.Join(lines, "\n")
}

func renderTutorialOverlay() string {
	lines := []string{
		sharedtui.TitleStyle.Render("Tutorial"),
		sharedtui.HelpDescStyle.Render("1) Press g to create a PRD via interview"),
		sharedtui.HelpDescStyle.Render("2) Press / to filter the list"),
		sharedtui.HelpDescStyle.Render("3) Press r to launch research"),
		sharedtui.HelpDescStyle.Render("4) Press p to generate suggestions"),
		sharedtui.HelpDescStyle.Render("5) Press s to review/apply suggestions"),
		sharedtui.HelpDescStyle.Render("Esc: close"),
	}
	return strings.Join(lines, "\n")
}

func renderConfirmOverlay(message string) string {
	lines := []string{
		sharedtui.TitleStyle.Render("Confirm"),
		sharedtui.HelpDescStyle.Render(message),
		sharedtui.HelpKeyStyle.Render("enter") + sharedtui.HelpDescStyle.Render(": confirm  ") +
			sharedtui.HelpKeyStyle.Render("esc") + sharedtui.HelpDescStyle.Render(": cancel"),
	}
	return strings.Join(lines, "\n")
}
