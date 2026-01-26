package research

import (
	"context"
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mistakeknot/autarch/internal/pollard/hunters"
)

// Coordinator manages research runs and dispatches updates to the TUI.
// It ensures only one run is active per project and handles cancellation
// when switching projects.
type Coordinator struct {
	registry *hunters.Registry

	mu        sync.RWMutex
	activeRun *Run
	program   *tea.Program // For sending messages to TUI
}

// NewCoordinator creates a new research coordinator.
func NewCoordinator(registry *hunters.Registry) *Coordinator {
	if registry == nil {
		registry = hunters.DefaultRegistry()
	}
	return &Coordinator{
		registry: registry,
	}
}

// SetProgram sets the Bubble Tea program for sending messages.
func (c *Coordinator) SetProgram(p *tea.Program) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.program = p
}

// StartRun begins a new research run for a project.
// If there's an existing run, it will be cancelled first.
func (c *Coordinator) StartRun(ctx context.Context, projectID string, hunterNames []string, topics []TopicConfig) (*Run, error) {
	c.mu.Lock()

	// Cancel any existing run
	if c.activeRun != nil {
		c.activeRun.Cancel()
		c.sendMsg(RunCancelledMsg{
			RunID:  c.activeRun.RunID,
			Reason: "new run started",
		})
	}

	// Create new run
	run := NewRunWithContext(ctx, projectID)
	c.activeRun = run

	// Register hunters
	for _, name := range hunterNames {
		run.RegisterHunter(name)
	}

	c.mu.Unlock()

	// Notify TUI of run start
	c.sendMsg(RunStartedMsg{
		RunID:     run.RunID,
		ProjectID: projectID,
		Hunters:   hunterNames,
	})

	// Start hunters in background
	go c.executeRun(run, hunterNames, topics)

	return run, nil
}

// TopicConfig maps interview topics to search queries.
type TopicConfig struct {
	Key     string   // e.g., "platform", "storage", "auth"
	Queries []string // Search terms for this topic
}

// executeRun runs all hunters and collects findings.
func (c *Coordinator) executeRun(run *Run, hunterNames []string, topics []TopicConfig) {
	var wg sync.WaitGroup

	for _, name := range hunterNames {
		hunter, ok := c.registry.Get(name)
		if !ok {
			c.sendMsg(HunterErrorMsg{
				RunID:      run.RunID,
				HunterName: name,
				Error:      fmt.Errorf("hunter not found: %s", name),
			})
			continue
		}

		wg.Add(1)
		go func(h hunters.Hunter, hunterName string) {
			defer wg.Done()
			c.executeHunter(run, h, hunterName, topics)
		}(hunter, name)
	}

	wg.Wait()

	// Mark run as complete
	run.MarkDone()
	c.sendMsg(RunCompletedMsg{
		RunID:         run.RunID,
		TotalFindings: run.TotalFindings(),
		Duration:      run.Duration().String(),
	})
}

// executeHunter runs a single hunter and sends updates.
func (c *Coordinator) executeHunter(run *Run, hunter hunters.Hunter, name string, topics []TopicConfig) {
	// Check for cancellation
	if run.IsCancelled() {
		return
	}

	run.StartHunter(name)
	c.sendMsg(HunterStartedMsg{
		RunID:      run.RunID,
		HunterName: name,
	})

	// Build queries from topics
	var allQueries []string
	topicMap := make(map[string]string) // query -> topicKey
	for _, topic := range topics {
		for _, q := range topic.Queries {
			allQueries = append(allQueries, q)
			topicMap[q] = topic.Key
		}
	}

	// Run the hunter
	cfg := hunters.HunterConfig{
		Queries:    allQueries,
		MaxResults: 10,
		Mode:       "balanced",
	}

	result, err := hunter.Hunt(run.Context, cfg)
	if err != nil {
		run.ErrorHunter(name, err)
		c.sendMsg(HunterErrorMsg{
			RunID:      run.RunID,
			HunterName: name,
			Error:      err,
		})
		return
	}

	// Process results into topic-scoped findings
	findings := c.processHuntResult(run.RunID, name, result, topicMap)
	for topicKey, topicFindings := range findings {
		update := Update{
			RunID:      run.RunID,
			HunterName: name,
			TopicKey:   topicKey,
			Findings:   topicFindings,
			Timestamp:  time.Now(),
		}
		run.AddUpdate(update)

		c.sendMsg(HunterUpdateMsg{
			RunID:      run.RunID,
			HunterName: name,
			TopicKey:   topicKey,
			Findings:   topicFindings,
		})
	}

	run.CompleteHunter(name, result.InsightsCreated)
	c.sendMsg(HunterCompletedMsg{
		RunID:        run.RunID,
		HunterName:   name,
		FindingCount: result.InsightsCreated,
	})
}

// processHuntResult converts hunt results to topic-scoped findings.
func (c *Coordinator) processHuntResult(runID, hunterName string, result *hunters.HuntResult, topicMap map[string]string) map[string][]Finding {
	// Group findings by topic
	findings := make(map[string][]Finding)

	// For now, create findings based on the hunt metadata
	// Real implementation would parse output files
	if result.SourcesCollected > 0 {
		// Default to general topic if no specific mapping
		topicKey := "general"
		if len(topicMap) > 0 {
			// Use first topic as default
			for _, k := range topicMap {
				topicKey = k
				break
			}
		}

		finding := Finding{
			ID:          fmt.Sprintf("%s-%s-%d", runID[:8], hunterName, time.Now().UnixNano()),
			Title:       fmt.Sprintf("%s results", hunterName),
			Summary:     fmt.Sprintf("Found %d sources with %d insights", result.SourcesCollected, result.InsightsCreated),
			Source:      hunterName,
			SourceType:  hunterName,
			Relevance:   0.7,
			CollectedAt: time.Now(),
		}

		findings[topicKey] = append(findings[topicKey], finding)
	}

	return findings
}

// CancelActiveRun cancels the current research run if one is active.
func (c *Coordinator) CancelActiveRun(reason string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.activeRun != nil {
		runID := c.activeRun.RunID
		c.activeRun.Cancel()
		c.activeRun = nil

		c.sendMsg(RunCancelledMsg{
			RunID:  runID,
			Reason: reason,
		})
	}
}

// GetActiveRun returns the currently active run, if any.
func (c *Coordinator) GetActiveRun() *Run {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.activeRun
}

// IsRunActive returns true if the given runID matches the active run.
func (c *Coordinator) IsRunActive(runID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.activeRun != nil && c.activeRun.RunID == runID
}

// GetUpdatesForTopic returns updates for a topic from the active run.
func (c *Coordinator) GetUpdatesForTopic(topicKey string) []Update {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.activeRun == nil {
		return nil
	}
	return c.activeRun.GetUpdatesForTopic(topicKey)
}

// GetHunterStatuses returns hunter statuses from the active run.
func (c *Coordinator) GetHunterStatuses() map[string]HunterStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.activeRun == nil {
		return nil
	}
	return c.activeRun.GetHunterStatuses()
}

// sendMsg sends a message to the TUI if a program is set.
func (c *Coordinator) sendMsg(msg tea.Msg) {
	c.mu.RLock()
	p := c.program
	c.mu.RUnlock()

	if p != nil {
		p.Send(msg)
	}
}

// RunningHunterCount returns count of currently running hunters.
func (c *Coordinator) RunningHunterCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.activeRun == nil {
		return 0
	}
	return c.activeRun.RunningCount()
}

// TotalFindings returns total findings from the active run.
func (c *Coordinator) TotalFindings() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.activeRun == nil {
		return 0
	}
	return c.activeRun.TotalFindings()
}
