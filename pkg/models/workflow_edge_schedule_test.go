package models

import "testing"

func TestStartEdgeTriggerScheduleValidation(t *testing.T) {
	for name, schedule := range map[string]*WeeklySchedule{
		"null":             nil,
		"negative day":     {Day: -1, Enabled: true},
		"out of range day": {Day: 8, Enabled: true},
		"invalid timezone": {Day: 1, Enabled: true, Timezone: "Invalid/Timezone"},
		"negative start":   {Day: 1, Segments: []DayTimeRange{{Start: -1, End: 10}}},
		"late end":         {Day: 1, Segments: []DayTimeRange{{Start: 0, End: 86401}}},
		"reversed segment": {Day: 1, Segments: []DayTimeRange{{Start: 10, End: 5}}},
		"empty segment":    {Day: 1, Segments: []DayTimeRange{{Start: 10, End: 10}}},
	} {
		t.Run(name, func(t *testing.T) {
			w := edgeWorkflow()
			w.Edges[0].Trigger = &WorkflowEdgeTrigger{WeeklySchedule: []*WeeklySchedule{schedule}}
			for _, enabled := range []bool{false, true} {
				w.Enabled = enabled
				if w.ValidateStartNodes() == nil || w.ValidateGraph() == nil || w.ValidateTriggers() == nil {
					t.Fatal("invalid authored edge schedule accepted")
				}
			}
			w.Nodes[0].Trigger.Type = WorkflowTriggerManual
			w.Nodes[0].Trigger.Surfaces = []WorkflowTriggerSurface{WorkflowSurfaceMedia}
			if err := w.ValidateGraph(); err != nil {
				t.Fatalf("manual mode rejected dormant schedule: %v", err)
			}
			w.Nodes[0].Trigger.Type = WorkflowTriggerAutomatic
			w.Edges[0].Trigger = nil
			w.Nodes[0].Trigger.WeeklySchedule = []*WeeklySchedule{schedule}
			if err := w.ValidateGraph(); err != nil {
				t.Fatalf("legacy node schedule policy changed: %v", err)
			}
		})
	}
	w := edgeWorkflow()
	w.Edges[0].Trigger = &WorkflowEdgeTrigger{WeeklySchedule: []*WeeklySchedule{{
		Day: 0, Enabled: true, Segments: []DayTimeRange{{Start: 0, End: 86400}},
	}}}
	if err := w.ValidateGraph(); err != nil {
		t.Fatalf("full day with legacy empty-timezone fallback rejected: %v", err)
	}
}
