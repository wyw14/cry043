package domain

import (
	"errors"
	"time"
)

type IssueLevel string

const (
	IssueObservation IssueLevel = "observation"
	IssueMinor       IssueLevel = "minor"
	IssueMajor       IssueLevel = "major"
	IssueCritical    IssueLevel = "critical"
)

type Inspection struct {
	ID, AreaID, ProcessID, SpecificationID, InspectorID string
	SpecificationVersion                                int
	EvidenceIDs                                         []string
	Findings                                            []Finding
	PerformedAt                                         time.Time
}
type Finding struct {
	ID, RuleID, Description string
	Level                   IssueLevel
	RemediationID           string
}
type RemediationStatus string

const (
	RemediationOpen      RemediationStatus = "open"
	RemediationSubmitted RemediationStatus = "submitted"
	RemediationVerified  RemediationStatus = "verified"
	RemediationClosed    RemediationStatus = "closed"
)

type Remediation struct {
	ID, FindingID, OwnerID, Action, ReviewerID, Conclusion string
	DueAt                                                  time.Time
	Status                                                 RemediationStatus
	EvidenceIDs                                            []string
	Revision                                               int64
}

func (r *Remediation) Submit(action string, evidence []string) error {
	if r.Status != RemediationOpen {
		return errors.New("remediation is not open")
	}
	if action == "" || len(evidence) == 0 {
		return errors.New("action and evidence are required")
	}
	r.Action = action
	r.EvidenceIDs = append([]string(nil), evidence...)
	r.Status = RemediationSubmitted
	r.Revision++
	return nil
}
func (r *Remediation) Verify(reviewer, conclusion string, pass bool) error {
	if r.Status != RemediationSubmitted {
		return errors.New("remediation is not submitted")
	}
	if reviewer == "" || conclusion == "" {
		return errors.New("reviewer and conclusion are required")
	}
	r.ReviewerID = reviewer
	r.Conclusion = conclusion
	r.Revision++
	if pass {
		r.Status = RemediationVerified
	} else {
		r.Status = RemediationOpen
	}
	return nil
}
func (r *Remediation) Close() error {
	if r.Status != RemediationVerified {
		return errors.New("only verified remediation can close")
	}
	r.Status = RemediationClosed
	r.Revision++
	return nil
}

type AuditEvent struct {
	ID, AggregateID, ActorID, Action, Detail, PreviousHash, Hash string
	At                                                           time.Time
}
