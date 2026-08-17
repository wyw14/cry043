package tests

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry043/internal/domain"
	"github.com/wyw14/cry043/internal/repository"
)

func TestAuditTimelineRejectsForkedHashChain(t *testing.T) {
	store := repository.NewMemory()
	first := domain.AuditEvent{ID: "event-1", AggregateID: "spec-v4", ActorID: "author", Action: "spec.approved", PreviousHash: "", Hash: "hash-one", At: time.Unix(1, 0)}
	if err := store.AppendAudit(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	fork := domain.AuditEvent{ID: "event-2", AggregateID: "spec-v4", ActorID: "scheduler", Action: "spec.effective", PreviousHash: "stale-head", Hash: "hash-two", At: time.Unix(2, 0)}
	if err := store.AppendAudit(context.Background(), fork); err == nil {
		t.Fatal("forked audit event was accepted")
	}
	head, err := store.AuditHead(context.Background(), first.AggregateID)
	if err != nil {
		t.Fatal(err)
	}
	if head != first.Hash {
		t.Fatalf("rejected fork changed the controlled export head: %q", head)
	}
}
