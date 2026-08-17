package application

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/wyw14/cry043/internal/domain"
)

// RiskService materializes operational risk from versioned specifications,
// confirmations, exceptions, inspections and corrective actions. It is kept
// separate from command services so dashboard reads cannot mutate workflows.
type RiskService struct {
	repo  Repository
	clock Clock
}

func NewRiskService(repo Repository, clock Clock) *RiskService {
	return &RiskService{repo: repo, clock: clock}
}

func (s *RiskService) Board(ctx context.Context, areaID, processID string, horizon time.Duration) (domain.ComplianceRiskBoard, error) {
	now := s.clock.Now()
	if horizon <= 0 || horizon > 90*24*time.Hour {
		horizon = 14 * 24 * time.Hour
	}
	window := domain.RiskWindow{AreaID: areaID, ProcessID: processID, ObservedAt: now, HorizonEnd: now.Add(horizon)}
	facts, err := s.repo.RiskFacts(ctx, window)
	if err != nil {
		return domain.ComplianceRiskBoard{}, err
	}
	board := domain.ComplianceRiskBoard{GeneratedAt: now, WindowEnd: window.HorizonEnd, Items: []domain.ScopeRisk{}}
	acknowledged := acknowledgedTeams(facts.Acknowledgements)
	totalTeams, confirmedTeams := 0, 0

	for _, spec := range facts.Specifications {
		if !domain.ScopeMatches(window, spec.Scope) {
			continue
		}
		if spec.Status == domain.SpecEffective {
			for _, teamID := range spec.Scope.TeamIDs {
				totalTeams++
				if acknowledged[confirmationKey(spec.ID, spec.Version, teamID)] {
					confirmedTeams++
					continue
				}
				board.Summary.PendingConfirmations++
				board.Items = append(board.Items, scopeRisk(spec, domain.RiskPendingTeamConfirmation, "major", teamID, now, "班组尚未确认当前生效版本"))
			}
		}
		if spec.Status == domain.SpecApproved || spec.Status == domain.SpecScheduled {
			readiness := activationReadiness(spec, facts.Specifications, facts.Acknowledgements)
			if !readiness.Ready {
				board.Summary.BlockedActivations++
				board.Items = append(board.Items, scopeRisk(spec, domain.RiskBlockedActivation, "major", "spec-owner", valueOrNow(spec.EffectiveAt, now), joinBlockers(readiness.Blockers)))
			}
		}
	}

	for _, exception := range facts.Exceptions {
		if exception.ApprovedAt != nil && !exception.ExpiresAt.After(window.HorizonEnd) {
			spec := findSpecification(facts.Specifications, exception.SpecificationID)
			if spec.ID == "" || !domain.ScopeMatches(window, spec.Scope) {
				continue
			}
			board.Summary.ExpiringExceptions++
			board.Items = append(board.Items, scopeRisk(spec, domain.RiskExpiringException, "major", exception.RequesterID, exception.ExpiresAt, "例外即将到期，需复核补偿措施"))
		}
	}

	inspectionByFinding := indexInspections(facts.Inspections)
	for _, remediation := range facts.Remediations {
		if remediation.Status == domain.RemediationClosed || remediation.DueAt.After(now) {
			continue
		}
		inspection, ok := inspectionByFinding[remediation.FindingID]
		if !ok {
			continue
		}
		spec := findSpecification(facts.Specifications, inspection.SpecificationID)
		if spec.ID == "" || !domain.ScopeMatches(window, spec.Scope) {
			continue
		}
		board.Summary.OverdueRemediations++
		item := scopeRisk(spec, domain.RiskOverdueRemediation, "critical", remediation.OwnerID, remediation.DueAt, "整改已超过期限且尚未关闭")
		item.AreaID, item.ProcessID = inspection.AreaID, inspection.ProcessID
		board.Items = append(board.Items, item)
	}

	if totalTeams > 0 {
		board.Summary.ConfirmationCoverage = float64(confirmedTeams) / float64(totalTeams)
	}
	sort.SliceStable(board.Items, func(i, j int) bool {
		if board.Items[i].Severity != board.Items[j].Severity {
			return board.Items[i].Severity == "critical"
		}
		return board.Items[i].DueAt.Before(board.Items[j].DueAt)
	})
	return board, nil
}

func (s *RiskService) Readiness(ctx context.Context, specificationID string) (domain.ActivationReadiness, error) {
	facts, err := s.repo.RiskFacts(ctx, domain.RiskWindow{ObservedAt: s.clock.Now(), HorizonEnd: s.clock.Now().Add(90 * 24 * time.Hour)})
	if err != nil {
		return domain.ActivationReadiness{}, err
	}
	spec := findSpecification(facts.Specifications, specificationID)
	if spec.ID == "" {
		return domain.ActivationReadiness{}, fmt.Errorf("specification %s not found", specificationID)
	}
	return activationReadiness(spec, facts.Specifications, facts.Acknowledgements), nil
}

func activationReadiness(spec domain.Specification, all []domain.Specification, acknowledgements []domain.Acknowledgement) domain.ActivationReadiness {
	result := domain.ActivationReadiness{SpecificationID: spec.ID, Version: spec.Version, RequiredTeams: len(spec.Scope.TeamIDs), Blockers: []domain.ActivationBlocker{}}
	if spec.Status != domain.SpecApproved && spec.Status != domain.SpecScheduled {
		result.Blockers = append(result.Blockers, domain.ActivationBlocker{Code: "STATUS_NOT_APPROVED", Detail: "规范尚未审批或排期"})
	}
	if len(spec.Scope.AreaIDs) == 0 || len(spec.Scope.ProcessIDs) == 0 {
		result.Blockers = append(result.Blockers, domain.ActivationBlocker{Code: "SCOPE_INCOMPLETE", Detail: "区域和工序生效范围必须同时配置"})
	}
	if len(spec.Rules) == 0 {
		result.Blockers = append(result.Blockers, domain.ActivationBlocker{Code: "RULESET_EMPTY", Detail: "至少配置一条规则"})
	}
	for _, other := range all {
		if other.ID == spec.ID || other.Status != domain.SpecEffective || !scopesOverlap(spec.Scope, other.Scope) {
			continue
		}
		for _, left := range spec.Rules {
			for _, right := range other.Rules {
				if left.RequiredField != "" && left.RequiredField == right.RequiredField && left.Kind != right.Kind {
					result.Blockers = append(result.Blockers, domain.ActivationBlocker{Code: "RULE_CONFLICT", Detail: "与生效规范 " + other.DocumentID + " 的字段规则冲突"})
				}
			}
		}
	}
	confirmed := map[string]bool{}
	for _, ack := range acknowledgements {
		if ack.SpecificationID == spec.ID && ack.Version == spec.Version {
			confirmed[ack.TeamID] = true
		}
	}
	for _, teamID := range spec.Scope.TeamIDs {
		if confirmed[teamID] {
			result.ConfirmedTeams++
		}
	}
	if result.RequiredTeams > 0 {
		result.Coverage = float64(result.ConfirmedTeams) / float64(result.RequiredTeams)
	}
	result.Ready = len(result.Blockers) == 0
	return result
}

func acknowledgedTeams(values []domain.Acknowledgement) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, ack := range values {
		out[confirmationKey(ack.SpecificationID, ack.Version, ack.TeamID)] = true
	}
	return out
}

func confirmationKey(specificationID string, version int, teamID string) string {
	return fmt.Sprintf("%s:%s", specificationID, teamID)
}

func findSpecification(specifications []domain.Specification, id string) domain.Specification {
	for _, spec := range specifications {
		if spec.ID == id {
			return spec
		}
	}
	return domain.Specification{}
}

func indexInspections(inspections []domain.Inspection) map[string]domain.Inspection {
	out := map[string]domain.Inspection{}
	for _, inspection := range inspections {
		for _, finding := range inspection.Findings {
			out[finding.ID] = inspection
		}
	}
	return out
}

func scopeRisk(spec domain.Specification, kind domain.RiskKind, severity, owner string, due time.Time, detail string) domain.ScopeRisk {
	area, process := "all", "all"
	if len(spec.Scope.AreaIDs) > 0 {
		area = spec.Scope.AreaIDs[0]
	}
	if len(spec.Scope.ProcessIDs) > 0 {
		process = spec.Scope.ProcessIDs[0]
	}
	return domain.ScopeRisk{Key: fmt.Sprintf("%s:%s:%s", kind, spec.ID, owner), AreaID: area, ProcessID: process, SpecificationID: spec.ID, SpecificationName: spec.DocumentID, Version: spec.Version, Kind: kind, Severity: severity, OwnerID: owner, DueAt: due, Detail: detail}
}

func scopesOverlap(left, right domain.Scope) bool {
	return anyShared(left.AreaIDs, right.AreaIDs) && anyShared(left.ProcessIDs, right.ProcessIDs)
}

func anyShared(left, right []string) bool {
	for _, l := range left {
		for _, r := range right {
			if l == r {
				return true
			}
		}
	}
	return false
}

func joinBlockers(blockers []domain.ActivationBlocker) string {
	if len(blockers) == 0 {
		return ""
	}
	result := blockers[0].Detail
	for _, blocker := range blockers[1:] {
		result += "；" + blocker.Detail
	}
	return result
}

func valueOrNow(value *time.Time, fallback time.Time) time.Time {
	if value != nil {
		return *value
	}
	return fallback
}
