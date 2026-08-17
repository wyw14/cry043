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

type safetyClock struct{ now time.Time }

func (c safetyClock) Now() time.Time { return c.now }

func TestMandatorySafetyRuleSurvivesExceptionRoundTrip(t *testing.T) {
	now := time.Date(2026, 8, 18, 10, 0, 0, 0, time.UTC)
	store := repository.NewMemory()
	spec := domain.Specification{ID: "lockout-v2", DocumentID: "lockout", Version: 2, Status: domain.SpecEffective, Revision: 1, Rules: []domain.Rule{{ID: "lock-pin", Kind: domain.RuleRequired, RequiredField: "lock_pin", MandatorySafety: true}}}
	if _, err := store.SaveSpecification(context.Background(), spec, 0, "seed-lockout"); err != nil {
		t.Fatal(err)
	}
	svc := application.NewComplianceService(store, safetyClock{now}, &service.IDs{}, &service.LocalScheduler{})
	exception, err := svc.RequestException(context.Background(), spec.ID, "lock-pin", "team-red", "设备赶工", "双人监护", now.Add(8*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	approvalErr := svc.ApproveException(context.Background(), exception.ID, exception, "safety-manager")
	violations, evalErr := svc.Evaluate(context.Background(), spec.ID, domain.ComplianceInput{Fields: map[string]string{}})
	if evalErr != nil {
		t.Fatal(evalErr)
	}
	if !errors.Is(approvalErr, domain.ErrSafetyException) {
		t.Fatalf("mandatory safety exception was approved: %v", approvalErr)
	}
	if len(violations) != 1 || violations[0].RuleID != "lock-pin" || violations[0].Severity != "critical" {
		t.Fatalf("mandatory safety violation was hidden: %#v", violations)
	}
}
