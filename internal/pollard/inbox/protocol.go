package inbox

import "github.com/mistakeknot/autarch/internal/pollard/api"

// Re-export message protocol types for inbox consumers.
type MessageType = api.MessageType

const (
	TypeResearchRequest  = api.TypeResearchRequest
	TypeResearchComplete = api.TypeResearchComplete
	TypeScanRequest      = api.TypeScanRequest
	TypeScanComplete     = api.TypeScanComplete
)

type ResearchMessage = api.ResearchMessage
type ResearchPayload = api.ResearchPayload
type ResearchResponse = api.ResearchResponse
