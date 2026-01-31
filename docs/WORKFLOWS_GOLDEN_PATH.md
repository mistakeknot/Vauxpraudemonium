# Golden Path Workflow

This document defines the canonical Autarch flow and a local smoke test.

## Canonical Flow

1) **New project**
   - Create a project root with `.gurgeh/` and `.coldwine/`
2) **PRD creation (Gurgeh)**
   - Produce a spec in `.gurgeh/specs/PRD-001.yaml`
3) **Task generation (Coldwine)**
   - Produce tasks in `.coldwine/tasks/*.yaml`
4) **Execution (Agent run)**
   - Agent run starts and logs are captured
5) **Outcome**
   - Outcome recorded and artifacts attached

## Smoke Test (Local)

Run the golden path smoke test locally (no Intermute required):

```bash
scripts/golden-path-smoke.sh
```

The script:
- Creates a temporary project root
- Writes a minimal spec + task
- Runs `autarch reconcile` to emit events
- Emits run/outcome/artifact events via a tiny Go helper
- Verifies event spine output

## Notes

- External hunters are **not** invoked.
- This test is meant to validate the event spine and reconciliation path, not the full UI.
