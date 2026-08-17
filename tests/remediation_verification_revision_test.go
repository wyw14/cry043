package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wyw14/cry043/internal/application"
	"github.com/wyw14/cry043/internal/domain"
	"github.com/wyw14/cry043/internal/repository"
	"github.com/wyw14/cry043/internal/service"
)

type remediationClock struct{}

func (remediationClock) Now() time.Time { return time.Date(2026, 8, 18, 11, 0, 0, 0, time.UTC) }

func TestRemediationNeedsVerificationAndRejectsStaleClosure(t *testing.T) {
	store := repository.NewMemory()
	submitted := domain.Remediation{ID: "rem-7", FindingID: "finding-7", OwnerID: "team-blue", Action: "replace shield", Status: domain.RemediationSubmitted, EvidenceIDs: []string{"photo-7"}, Revision: 2, DueAt: time.Now().Add(time.Hour)}
	if err := store.SaveRemediation(context.Background(), submitted, 0); err != nil {
		t.Fatal(err)
	}
	workflow := application.NewInspectionService(store, remediationClock{}, &service.IDs{})
	closeErr := workflow.Close(context.Background(), submitted.ID)

	current := submitted
	current.Status = domain.RemediationVerified
	current.ReviewerID = "reviewer"
	current.Conclusion = "现场复核通过"
	current.Revision = 3
	firstErr := store.SaveRemediation(context.Background(), current, 2)
	stale := submitted
	stale.Status = domain.RemediationClosed
	stale.Revision = 3
	staleErr := store.SaveRemediation(context.Background(), stale, 2)

	if closeErr == nil {
		t.Fatal("submitted remediation closed without independent verification")
	}
	if firstErr != nil {
		t.Fatalf("current verification was rejected: %v", firstErr)
	}
	if !errors.Is(staleErr, repository.ErrRevision) {
		t.Fatalf("stale closure overwrote verified revision: %v", staleErr)
	}
}
