#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
TMP_DIR=$(mktemp -d)
EVENTS_DB="$TMP_DIR/events.db"

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

mkdir -p "$TMP_DIR/.gurgeh/specs" "$TMP_DIR/.coldwine/tasks"

cat > "$TMP_DIR/.gurgeh/specs/PRD-001.yaml" <<'YAML'
id: "PRD-001"
title: "Golden Path Spec"
status: "draft"
version: 1
YAML

cat > "$TMP_DIR/.coldwine/tasks/TASK-001.yaml" <<'YAML'
id: "TASK-001"
title: "Golden Path Task"
status: "pending"
YAML

cd "$ROOT_DIR"

go run ./cmd/autarch reconcile --project "$TMP_DIR" --events-db "$EVENTS_DB" >/dev/null

cat > "$TMP_DIR/emit_run.go" <<'GO'
package main

import (
	"flag"
	"log"
	"time"

	"github.com/mistakeknot/autarch/pkg/contract"
	"github.com/mistakeknot/autarch/pkg/events"
)

func main() {
	dbPath := flag.String("events-db", "", "events db path")
	project := flag.String("project", "", "project path")
	flag.Parse()
	if *project == "" {
		log.Fatal("--project required")
	}
	store, err := events.OpenStore(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	writer := events.NewWriter(store, events.SourceColdwine)
	writer.SetProjectPath(*project)

	run := contract.Run{
		ID:           "RUN-001",
		TaskID:       "TASK-001",
		AgentName:    "golden-path",
		AgentProgram: "codex",
		State:        contract.RunStateDone,
		SourceTool:   contract.SourceColdwine,
		StartedAt:    time.Now(),
	}
	if err := writer.EmitRunStarted(&run); err != nil {
		log.Fatal(err)
	}
	if err := writer.EmitRunCompleted(run.ID); err != nil {
		log.Fatal(err)
	}

	outcome := contract.Outcome{
		ID:         "OUT-001",
		RunID:      run.ID,
		Success:    true,
		Summary:    "golden path",
		SourceTool: contract.SourceColdwine,
		CreatedAt:  time.Now(),
	}
	if err := writer.EmitOutcomeRecorded(&outcome); err != nil {
		log.Fatal(err)
	}

	artifact := contract.RunArtifact{
		ID:        run.ID + ":note",
		RunID:     run.ID,
		Type:      "note",
		Label:     "golden path",
		Path:      "/dev/null",
		MimeType:  "text/plain",
		CreatedAt: time.Now(),
	}
	if err := writer.EmitRunArtifactAdded(artifact); err != nil {
		log.Fatal(err)
	}
}
GO

go run "$TMP_DIR/emit_run.go" --events-db "$EVENTS_DB" --project "$TMP_DIR" >/dev/null

require_event() {
  local event_type="$1"
  if ! go run ./cmd/autarch events query --type "$event_type" --project "$TMP_DIR" --events-db "$EVENTS_DB" | grep -q "$event_type"; then
    echo "Missing event: $event_type"
    exit 1
  fi
}

require_event "spec_revised"
require_event "task_created"
require_event "run_started"
require_event "run_completed"
require_event "outcome_recorded"
require_event "run_artifact_added"

echo "Golden path smoke test passed."
