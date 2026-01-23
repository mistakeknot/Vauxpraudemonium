package epics

import "testing"

func TestValidateEpicsReportsErrors(t *testing.T) {
	epics := []Epic{
		{
			ID:       "EPIC-1",
			Title:    "Auth",
			Status:   Status("bogus"),
			Priority: Priority("p9"),
			Stories: []Story{
				{ID: "EPIC-002-S01", Title: "Bad story", Status: StatusTodo, Priority: PriorityP1},
			},
		},
	}

	errList := Validate(epics)
	if len(errList) == 0 {
		t.Fatalf("expected validation errors")
	}
}
