package tui

import (
	"sort"
	"strings"

	"github.com/mistakeknot/vauxpraudemonium/internal/gurgeh/specs"
)

type ItemType int

const (
	ItemTypeGroup ItemType = iota
	ItemTypePRD
)

type Group struct {
	Name     string
	Expanded bool
	Items    []specs.Summary
}

type Item struct {
	Type          ItemType
	Group         *Group
	Summary       *specs.Summary
	IsLastInGroup bool
}

type GroupTree struct {
	Groups []*Group
}

var StatusOrder = []string{"interview", "draft", "research", "suggestions", "validated", "archived"}

func NewGroupTree(summaries []specs.Summary, expanded map[string]bool) *GroupTree {
	groups := make([]*Group, 0, len(StatusOrder))
	byStatus := make(map[string][]specs.Summary)
	known := make(map[string]struct{})
	for _, status := range StatusOrder {
		known[status] = struct{}{}
	}
	for _, s := range summaries {
		status := normalizeStatus(s.Status)
		if _, ok := known[status]; !ok {
			status = "draft"
		}
		byStatus[status] = append(byStatus[status], s)
	}
	for _, status := range StatusOrder {
		items := byStatus[status]
		if len(items) == 0 {
			continue
		}
		sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
		groups = append(groups, &Group{
			Name:     status,
			Expanded: expanded[status],
			Items:    items,
		})
	}
	return &GroupTree{Groups: groups}
}

func (t *GroupTree) Flatten() []Item {
	var out []Item
	for _, g := range t.Groups {
		out = append(out, Item{Type: ItemTypeGroup, Group: g})
		if !g.Expanded {
			continue
		}
		for i := range g.Items {
			last := i == len(g.Items)-1
			out = append(out, Item{Type: ItemTypePRD, Group: g, Summary: &g.Items[i], IsLastInGroup: last})
		}
	}
	return out
}

func normalizeStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}
