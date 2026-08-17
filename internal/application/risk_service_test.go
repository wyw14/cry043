package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry043/internal/application"
	"github.com/wyw14/cry043/internal/domain"
	"github.com/wyw14/cry043/internal/repository"
)

func TestRiskBoardCorrelatesConfirmationExceptionAndOverdueFinding(t *testing.T) {
	now := time.Date(2026, 8, 18, 8, 0, 0, 0, time.UTC)
	store := repository.NewMemory()
	spec := domain.Specification{
		ID: "hot-work-v3", DocumentID: "hot-work", Version: 3, Status: domain.SpecEffective, Revision: 1,
		Scope: domain.Scope{AreaIDs: []string{"welding"}, ProcessIDs: []string{"hot-work"}, TeamIDs: []string{"night", "day"}},
		Rules: []domain.Rule{{ID: "permit", Kind: domain.RuleRequired, RequiredField: "permit"}},
	}
	if _, err := store.SaveSpecification(context.Background(), spec, 0, "seed:hot-work-v3"); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveAcknowledgement(context.Background(), domain.Acknowledgement{ID: "ack-day", SpecificationID: spec.ID, Version: 3, UserID: "u-day", TeamID: "day", ReadAt: now}); err != nil {
		t.Fatal(err)
	}
	approved := now.Add(-24 * time.Hour)
	if err := store.SaveException(context.Background(), domain.ExceptionRequest{ID: "ex-1", SpecificationID: spec.ID, RuleID: "permit", RequesterID: "night", ApproverID: "safety", ApprovedAt: &approved, ExpiresAt: now.Add(48 * time.Hour)}); err != nil {
		t.Fatal(err)
	}
	inspection := domain.Inspection{ID: "inspection-1", SpecificationID: spec.ID, SpecificationVersion: 3, AreaID: "welding", ProcessID: "hot-work", Findings: []domain.Finding{{ID: "finding-1", RuleID: "permit", Level: domain.IssueMajor}}}
	if err := store.SaveInspection(context.Background(), inspection); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveRemediation(context.Background(), domain.Remediation{ID: "fix-1", FindingID: "finding-1", OwnerID: "night", DueAt: now.Add(-time.Hour), Status: domain.RemediationOpen, Revision: 1}, 0); err != nil {
		t.Fatal(err)
	}

	board, err := application.NewRiskService(store, fixed{now}).Board(context.Background(), "welding", "hot-work", 7*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if board.Summary.PendingConfirmations != 1 || board.Summary.ExpiringExceptions != 1 || board.Summary.OverdueRemediations != 1 {
		t.Fatalf("unexpected summary %#v", board.Summary)
	}
	if board.Summary.ConfirmationCoverage != 0.5 || len(board.Items) != 3 || board.Items[0].Kind != domain.RiskOverdueRemediation {
		t.Fatalf("unexpected board %#v", board)
	}
}

func TestActivationReadinessExplainsScopeAndLiveRuleConflict(t *testing.T) {
	now := time.Date(2026, 8, 18, 8, 0, 0, 0, time.UTC)
	store := repository.NewMemory()
	live := domain.Specification{ID: "live", DocumentID: "lockout", Version: 2, Status: domain.SpecEffective, Revision: 1, Scope: domain.Scope{AreaIDs: []string{"assembly"}, ProcessIDs: []string{"maintenance"}}, Rules: []domain.Rule{{ID: "live-permit", Kind: domain.RuleValidity, RequiredField: "permit"}}}
	candidate := domain.Specification{ID: "candidate", DocumentID: "maintenance", Version: 1, Status: domain.SpecApproved, Revision: 1, Scope: domain.Scope{AreaIDs: []string{"assembly"}, ProcessIDs: []string{"maintenance"}, TeamIDs: []string{"blue"}}, Rules: []domain.Rule{{ID: "new-permit", Kind: domain.RuleRequired, RequiredField: "permit"}}}
	_, _ = store.SaveSpecification(context.Background(), live, 0, "live")
	_, _ = store.SaveSpecification(context.Background(), candidate, 0, "candidate")
	readiness, err := application.NewRiskService(store, fixed{now}).Readiness(context.Background(), candidate.ID)
	if err != nil {
		t.Fatal(err)
	}
	if readiness.Ready || len(readiness.Blockers) != 1 || readiness.Blockers[0].Code != "RULE_CONFLICT" {
		t.Fatalf("unexpected readiness %#v", readiness)
	}
}
