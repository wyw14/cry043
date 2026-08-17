package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wyw14/cry043/internal/domain"
	"github.com/wyw14/cry043/internal/repository"
)

func TestEffectiveSpecificationAndRevisionCannotBeOverwritten(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	effective := domain.Specification{ID: "live-v4", DocumentID: "live", Version: 4, Status: domain.SpecEffective, Revision: 8, Rules: []domain.Rule{{ID: "guard"}}}
	if err := effective.ReplaceRules([]domain.Rule{{ID: "weakened"}}, now); !errors.Is(err, domain.ErrEffectiveImmutable) {
		t.Fatalf("effective rules were rewritten in place: %v", err)
	}

	store := repository.NewMemory()
	draft := domain.Specification{ID: "draft-v5", DocumentID: "live", Version: 5, Status: domain.SpecDraft, Revision: 1, Rules: []domain.Rule{{ID: "guard"}}}
	if _, err := store.SaveSpecification(context.Background(), draft, 0, "create-v5"); err != nil {
		t.Fatal(err)
	}
	left, _ := store.Specification(context.Background(), draft.ID)
	right := left.Clone()
	_ = left.ReplaceRules([]domain.Rule{{ID: "left"}}, now)
	_ = right.ReplaceRules([]domain.Rule{{ID: "right"}}, now)
	if _, err := store.SaveSpecification(context.Background(), left, 1, "left-edit"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveSpecification(context.Background(), right, 1, "right-edit"); !errors.Is(err, repository.ErrRevision) {
		t.Fatalf("stale editor overwrote the first revision: %v", err)
	}
}
