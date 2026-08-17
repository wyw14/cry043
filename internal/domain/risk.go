package domain

import (
	"slices"
	"time"
)

// RiskKind identifies why a production scope needs attention. The values are
// deliberately business-facing because they are returned by the control-room
// API and used by the workshop dashboard.
type RiskKind string

const (
	RiskPendingTeamConfirmation RiskKind = "pending_team_confirmation"
	RiskExpiringException       RiskKind = "expiring_exception"
	RiskOverdueRemediation      RiskKind = "overdue_remediation"
	RiskBlockedActivation       RiskKind = "blocked_activation"
)

type RiskWindow struct {
	AreaID     string
	ProcessID  string
	HorizonEnd time.Time
	ObservedAt time.Time
}

type RiskFacts struct {
	Specifications   []Specification
	Acknowledgements []Acknowledgement
	Exceptions       []ExceptionRequest
	Inspections      []Inspection
	Remediations     []Remediation
}

type ScopeRisk struct {
	Key               string    `json:"key"`
	AreaID            string    `json:"area_id"`
	ProcessID         string    `json:"process_id"`
	SpecificationID   string    `json:"specification_id"`
	SpecificationName string    `json:"specification_name"`
	Version           int       `json:"version"`
	Kind              RiskKind  `json:"kind"`
	Severity          string    `json:"severity"`
	OwnerID           string    `json:"owner_id"`
	DueAt             time.Time `json:"due_at"`
	Detail            string    `json:"detail"`
}

type RiskSummary struct {
	PendingConfirmations int     `json:"pending_confirmations"`
	ExpiringExceptions   int     `json:"expiring_exceptions"`
	OverdueRemediations  int     `json:"overdue_remediations"`
	BlockedActivations   int     `json:"blocked_activations"`
	ConfirmationCoverage float64 `json:"confirmation_coverage"`
}

type ComplianceRiskBoard struct {
	GeneratedAt time.Time   `json:"generated_at"`
	WindowEnd   time.Time   `json:"window_end"`
	Summary     RiskSummary `json:"summary"`
	Items       []ScopeRisk `json:"items"`
}

type ActivationBlocker struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

type ActivationReadiness struct {
	SpecificationID string              `json:"specification_id"`
	Version         int                 `json:"version"`
	Ready           bool                `json:"ready"`
	RequiredTeams   int                 `json:"required_teams"`
	ConfirmedTeams  int                 `json:"confirmed_teams"`
	Coverage        float64             `json:"coverage"`
	Blockers        []ActivationBlocker `json:"blockers"`
}

func ScopeMatches(window RiskWindow, scope Scope) bool {
	if window.AreaID != "" && !slices.Contains(scope.AreaIDs, window.AreaID) {
		return false
	}
	if window.ProcessID != "" && !slices.Contains(scope.ProcessIDs, window.ProcessID) {
		return false
	}
	return true
}
