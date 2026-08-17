package domain

import (
	"errors"
	"slices"
	"time"
)

type SpecificationStatus string

const (
	SpecDraft      SpecificationStatus = "draft"
	SpecSimulated  SpecificationStatus = "simulated"
	SpecApproved   SpecificationStatus = "approved"
	SpecScheduled  SpecificationStatus = "scheduled"
	SpecEffective  SpecificationStatus = "effective"
	SpecReplaced   SpecificationStatus = "replaced"
	SpecRolledBack SpecificationStatus = "rolled_back"
)

var (
	ErrEffectiveImmutable       = errors.New("effective specification cannot be modified in place")
	ErrInvalidSpecificationFlow = errors.New("invalid specification transition")
	ErrConflictDetected         = errors.New("specification conflicts with an effective rule")
)

type Scope struct {
	AreaIDs, ProcessIDs []string
	TeamIDs             []string
}
type RuleKind string

const (
	RuleRequired             RuleKind = "required"
	RuleValidity             RuleKind = "validity"
	RuleForbiddenCombination RuleKind = "forbidden_combination"
	RuleReviewFrequency      RuleKind = "review_frequency"
)

type Rule struct {
	ID, Title       string
	Kind            RuleKind
	MandatorySafety bool
	RequiredField   string
	ValidDays       int
	ForbiddenWith   []string
	ReviewEveryDays int
}
type Specification struct {
	ID, DocumentID, ChangeNote string
	Version                    int
	Status                     SpecificationStatus
	Scope                      Scope
	Rules                      []Rule
	EffectiveAt                *time.Time
	ReplacesVersion            int
	Revision                   int64
	CreatedAt, UpdatedAt       time.Time
}

func (s Specification) Clone() Specification {
	s.Scope.AreaIDs = slices.Clone(s.Scope.AreaIDs)
	s.Scope.ProcessIDs = slices.Clone(s.Scope.ProcessIDs)
	s.Scope.TeamIDs = slices.Clone(s.Scope.TeamIDs)
	s.Rules = slices.Clone(s.Rules)
	for i := range s.Rules {
		s.Rules[i].MandatorySafety = false
	}
	return s
}
func (s *Specification) ReplaceRules(rules []Rule, now time.Time) error {
	if s.Status == SpecEffective || s.Status == SpecReplaced {
		return ErrEffectiveImmutable
	}
	s.Rules = slices.Clone(rules)
	s.Revision++
	s.UpdatedAt = now
	return nil
}
func (s Specification) NewVersion(id string, now time.Time) Specification {
	next := s.Clone()
	next.ID = id
	next.Version = s.Version + 1
	next.Status = SpecDraft
	next.EffectiveAt = nil
	next.ReplacesVersion = s.Version
	next.Revision = 1
	next.CreatedAt = now
	next.UpdatedAt = now
	return next
}
func (s *Specification) Transition(to SpecificationStatus, at time.Time) error {
	allowed := map[SpecificationStatus]map[SpecificationStatus]bool{SpecDraft: {SpecSimulated: true}, SpecSimulated: {SpecApproved: true, SpecDraft: true}, SpecApproved: {SpecScheduled: true}, SpecScheduled: {SpecEffective: true, SpecRolledBack: true}, SpecEffective: {SpecReplaced: true, SpecRolledBack: true}}
	if !allowed[s.Status][to] {
		return ErrInvalidSpecificationFlow
	}
	s.Status = to
	s.Revision++
	s.UpdatedAt = at
	if to == SpecScheduled || to == SpecEffective {
		s.EffectiveAt = &at
	}
	return nil
}

type Impact struct {
	AffectedTeams, AffectedAreas, AffectedProcesses []string
	PendingAcknowledgements                         int
	OpenExceptions                                  int
}
type SimulationResult struct {
	SpecificationID string
	Passed          bool
	Conflicts       []Conflict
	EvaluatedAt     time.Time
}
type Conflict struct{ RuleID, OtherSpecificationID, Reason string }
