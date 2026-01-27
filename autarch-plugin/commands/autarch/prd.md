---
name: autarch:prd
description: Create a PRD using Arbiter's Spec Sprint workflow
argument-hint: "[feature description or PRD-ID]"
---

# Generate or Edit PRD

Create or edit a Product Requirements Document using Arbiter's Spec Sprint workflow—a propose-first approach where Arbiter generates draft sections and you provide feedback.

## Usage

```bash
# Create new PRD interactively
/autarch:prd

# Create PRD with initial description
/autarch:prd "User authentication with OAuth support"

# Edit existing PRD
/autarch:prd PRD-001
```

## Spec Sprint Workflow

Arbiter's Spec Sprint workflow uses a propose-first approach with 6 core sections:

### Section 1: Problem
- **What**: Clear problem statement
- **Why**: Impact and urgency
- **Who**: Affected users/systems
- **Quick Scan** (after Problem): Automated scan for conflicting requirements and feasibility gaps

### Section 2: Solution Direction
- **Approach**: High-level solution architecture
- **Key Components**: Major pieces involved
- **Out of Scope**: Explicit boundaries

### Section 3: Success Criteria
- **Measurable Metrics**: How success is measured
- **User Outcomes**: What users experience
- **Business Impact**: Revenue, retention, efficiency gains

### Section 4: Technical Requirements
- **Must-Have Features**: Non-negotiable functionality
- **Integration Points**: Systems to connect with
- **Constraints**: Technical and business limits

### Section 5: User Journey
- **Critical Paths**: Main user flows
- **Edge Cases**: Exception handling
- **Dependencies**: External systems involved

### Section 6: Assumptions & Risks
- **Key Assumptions**: Foundational beliefs
- **Risk Factors**: What could go wrong
- **Mitigation Strategies**: How to address risks

## Interaction Model

Arbiter proposes draft content for each section. You review and provide feedback:
- **Accept**: Content is good, move to next section
- **Refine**: Provide specific feedback for improvements
- **Reject**: Request complete rewrite with different approach

Arbiter iterates based on your reactions until all sections meet your approval.

## Steps

1. **Initialize**: Provide feature description or PRD-ID
2. **Invoke spec-sprint skill**: Start Arbiter's Spec Sprint workflow
3. **Section Loop** (for each of 6 sections):
   - Arbiter proposes draft content
   - You review and react (accept/refine/reject)
   - Arbiter iterates until approved
4. **Quick Scan** (after Problem section):
   - Automated scan for requirement conflicts
   - Feasibility assessment
   - Integration point validation
5. **Persist State**: Sprint progress and reactions saved
6. **Finalize**: Compile approved sections into PRD
7. **Next Steps**: Suggest `/autarch:research` for validation or `/autarch:tasks` for implementation planning

## Output

PRD is saved to `.gurgeh/specs/PRD-{id}.yaml` with:
- Title and summary
- All 6 approved Spec Sprint sections
- Problem statement with quick scan results
- Solution direction and technical requirements
- Success criteria and user journeys
- Assumptions, risks, and mitigation strategies
- Sprint state and interaction history persisted for future refinement
