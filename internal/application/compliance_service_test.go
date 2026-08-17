package application_test

import (
	"context"
	"github.com/wyw14/cry043/internal/application"
	"github.com/wyw14/cry043/internal/domain"
	"github.com/wyw14/cry043/internal/repository"
	"github.com/wyw14/cry043/internal/service"
	"testing"
	"time"
)

type fixed struct{ v time.Time }

func (f fixed) Now() time.Time { return f.v }
func TestSimulationFindsScopeConflict(t *testing.T) {
	now := time.Now()
	repo := repository.NewMemory()
	base := domain.Specification{ID: "old", DocumentID: "d", Version: 1, Status: domain.SpecEffective, Scope: domain.Scope{AreaIDs: []string{"a"}}, Rules: []domain.Rule{{ID: "old-rule", Kind: domain.RuleValidity, RequiredField: "permit"}}, Revision: 1}
	_, _ = repo.SaveSpecification(context.Background(), base, 0, "old")
	candidate := domain.Specification{ID: "new", DocumentID: "n", Version: 1, Status: domain.SpecDraft, Scope: domain.Scope{AreaIDs: []string{"a"}}, Rules: []domain.Rule{{ID: "new-rule", Kind: domain.RuleRequired, RequiredField: "permit"}}, Revision: 1}
	_, _ = repo.SaveSpecification(context.Background(), candidate, 0, "new")
	svc := application.NewComplianceService(repo, fixed{now}, &service.IDs{}, &service.LocalScheduler{})
	result, err := svc.Simulate(context.Background(), "new", "author")
	if err != nil {
		t.Fatal(err)
	}
	if result.Passed || len(result.Conflicts) != 1 {
		t.Fatalf("result %#v", result)
	}
}
