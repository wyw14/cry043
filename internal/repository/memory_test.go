package repository

import (
	"context"
	"testing"

	"github.com/wyw14/cry043/internal/domain"
)

func TestAuditChainRejectsMismatchedHead(t *testing.T) {
	store := NewMemory()
	if err := store.AppendAudit(context.Background(), domain.AuditEvent{ID: "a1", AggregateID: "spec", Hash: "head-1"}); err != nil {
		t.Fatal(err)
	}
	if err := store.AppendAudit(context.Background(), domain.AuditEvent{ID: "a2", AggregateID: "spec", PreviousHash: "other", Hash: "head-2"}); err == nil {
		t.Fatal("mismatched audit head was accepted")
	}
}
