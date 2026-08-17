package tests

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry043/internal/application"
	"github.com/wyw14/cry043/internal/domain"
	"github.com/wyw14/cry043/internal/repository"
)

type riskClock struct{ now time.Time }

func (c riskClock) Now() time.Time { return c.now }

func TestRiskBoardKeepsProcessAndVersionBoundaries(t *testing.T) {
	now := time.Date(2026, 8, 18, 8, 0, 0, 0, time.UTC)
	store := repository.NewMemory()
	target := domain.Specification{ID: "weld-v3", DocumentID: "weld", Version: 3, Status: domain.SpecEffective, Revision: 1, Scope: domain.Scope{AreaIDs: []string{"assembly"}, ProcessIDs: []string{"welding"}, TeamIDs: []string{"day", "night"}}}
	otherProcess := domain.Specification{ID: "paint-v2", DocumentID: "paint", Version: 2, Status: domain.SpecEffective, Revision: 1, Scope: domain.Scope{AreaIDs: []string{"assembly"}, ProcessIDs: []string{"painting"}, TeamIDs: []string{"paint-team"}}}
	_, _ = store.SaveSpecification(context.Background(), target, 0, "target")
	_, _ = store.SaveSpecification(context.Background(), otherProcess, 0, "other")
	_ = store.SaveAcknowledgement(context.Background(), domain.Acknowledgement{ID: "ack-day", SpecificationID: target.ID, Version: 3, UserID: "day-user", TeamID: "day"})
	_ = store.SaveAcknowledgement(context.Background(), domain.Acknowledgement{ID: "ack-old", SpecificationID: target.ID, Version: 2, UserID: "night-user", TeamID: "night"})
	approved := now.Add(-72 * time.Hour)
	_ = store.SaveException(context.Background(), domain.ExceptionRequest{ID: "expired", SpecificationID: target.ID, RuleID: "permit", RequesterID: "night", ApproverID: "safety", ApprovedAt: &approved, ExpiresAt: now.Add(-time.Hour)})

	board, err := application.NewRiskService(store, riskClock{now}).Board(context.Background(), "assembly", "welding", 7*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if board.Summary.PendingConfirmations != 1 || board.Summary.ExpiringExceptions != 0 {
		t.Fatalf("scope/version isolation failed: %#v", board.Summary)
	}
	if len(board.Items) != 1 || board.Items[0].SpecificationID != target.ID || board.Items[0].OwnerID != "night" {
		t.Fatalf("unexpected risk items: %#v", board.Items)
	}
}
