package tui

import "testing"

func TestLayoutModeSelection(t *testing.T) {
	if layoutMode(40) != LayoutModeSingle {
		t.Fatalf("expected single")
	}
	if layoutMode(60) != LayoutModeStacked {
		t.Fatalf("expected stacked")
	}
	if layoutMode(90) != LayoutModeDual {
		t.Fatalf("expected dual")
	}
}
