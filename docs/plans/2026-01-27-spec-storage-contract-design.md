# Spec Storage Contract (Intermute Standard + Autarch PRD Source of Truth)

Date: 2026-01-27
Status: Draft

## Goal
Define a stable contract where Intermute remains an open, minimal standard for coordination, while Autarch owns the full PRD schema and storage. The contract must preserve spec quality (requirements, acceptance criteria, CUJs, evidence) without bloating Intermute.

## Decision (Option A)
- **Intermute** stores only minimal Spec metadata and linkage primitives.
- **Autarch** stores the full PRD in `.praude/specs/*.yaml` and remains source of truth.
- Other tools resolve full PRDs via filesystem access (or Autarch CLI export), not via Intermute.

## Scope
In scope:
- SpecSummary shape (minimal metadata published to Intermute)
- File pointer + hash for PRD resolution
- Event flow and update responsibilities
- Compatibility rules for other tools

Out of scope:
- Full PRD schema changes (remains in Autarch)
- Intermute auth/transport details
- Migration tooling

## Contract Overview
Intermute Spec is **metadata only**. Autarch publishes SpecSummary updates into Intermute and keeps full PRD YAML local. Intermute is the discovery/index layer, not the schema host.

### SpecSummary (Intermute)
Stored as Intermute `Spec` plus a small, typed set of reference fields:

Required fields:
- `id` (stable PRD id, e.g. PRD-001)
- `title`
- `status` (draft|research|validated|archived)
- `version` (monotonic int)
- `project`
- `updated_at`
- `spec_ref`: opaque reference for the PRD (file path or URI)
- `spec_hash`: hash of canonicalized PRD contents (sha256)

Intermute does **not** store full requirements, CUJs, evidence refs, or hypotheses.

### PRD File (Autarch)
Autarch owns the full PRD schema in `.praude/specs/*.yaml` (requirements, AC, CUJs, evidence graph, hypotheses, signals). This remains canonical for all deep reads.

## Data Flow
1) **Create/Update PRD** in Autarch
- Write `.praude/specs/PRD-###.yaml`
- Compute canonical hash
- Publish SpecSummary to Intermute (create/update)

2) **Link Research**
- Pollard creates Insights in Intermute
- Autarch links insights to PRD in PRD file (evidence refs)

3) **Consumers**
- Bigend lists specs from Intermute (fast index)
- Coldwine reads PRD file directly when generating tasks
- External tools can optionally fetch PRD via Autarch CLI (`autarch spec export <id>`) if file access is unavailable

## Responsibilities
- **Autarch**: authoritative PRD storage, schema evolution, evidence graph, CUJs, requirements, ACs, hypotheses
- **Intermute**: minimal Spec registry + event log + links (insights, CUJ IDs)
- **Bigend**: discovery + display using SpecSummary only
- **Coldwine**: reads full PRD from filesystem; uses SpecSummary only for listing

## Compatibility Rules
- Intermute Spec ID must equal PRD ID
- `spec_hash` is computed from canonicalized PRD content (stable key order + normalized newlines)
- `spec_ref` is opaque; Intermute does not interpret it
- Consumers must not assume Intermute contains full PRD
- If `spec_ref` missing or unreadable, consumers should fall back to SpecSummary only

## Operational Notes
- PRD writes should be atomic (temp file + fsync + rename)
- SpecSummary publish should be idempotent and retryable
- Add a reconciler to resync SpecSummary from PRD files when drift is detected

## Acceptance Criteria
- Autarch publishes SpecSummary on PRD create/update with correct `spec_ref` + `spec_hash`
- Intermute remains schema-light (no PRD fields beyond metadata)
- Coldwine can generate tasks using PRD files without Intermute
- Bigend can list specs using Intermute without filesystem access

## Open Questions
- Should `spec_ref` be a dedicated column on Intermute Spec or an optional `spec_ref` table?
- Remote-only export bundle is deferred; local-only by default
