package domain

import (
	"errors"
	"testing"
	"time"
)

func TestEffectiveSpecificationIsImmutable(t *testing.T) {
	s := Specification{Status: SpecEffective}
	if err := s.ReplaceRules([]Rule{{ID: "new"}}, time.Now()); !errors.Is(err, ErrEffectiveImmutable) {
		t.Fatalf("got %v", err)
	}
	next := s.NewVersion("v2", time.Now())
	if next.Status != SpecDraft || next.Version != 1 {
		t.Fatalf("bad next %#v", next)
	}
}
func TestMandatorySafetyRuleCannotBeExcepted(t *testing.T) {
	now := time.Now()
	e := ExceptionRequest{ExpiresAt: now.Add(time.Hour)}
	if err := e.Approve(Rule{MandatorySafety: true}, "manager", now); !errors.Is(err, ErrSafetyException) {
		t.Fatalf("got %v", err)
	}
}
func TestEvaluationHonorsOnlyValidNonSafetyExceptions(t *testing.T) {
	now := time.Now()
	spec := Specification{Rules: []Rule{{ID: "guard", Kind: RuleRequired, RequiredField: "guard", MandatorySafety: true}, {ID: "label", Kind: RuleRequired, RequiredField: "label"}}}
	approved := now.Add(-time.Minute)
	exceptions := []ExceptionRequest{{RuleID: "guard", ApproverID: "x", ApprovedAt: &approved, ExpiresAt: now.Add(time.Hour)}, {RuleID: "label", ApproverID: "x", ApprovedAt: &approved, ExpiresAt: now.Add(time.Hour)}}
	violations := Evaluate(spec, ComplianceInput{Fields: map[string]string{}}, exceptions, now)
	if len(violations) != 1 || violations[0].RuleID != "guard" || violations[0].Severity != "critical" {
		t.Fatalf("violations %#v", violations)
	}
}
func TestAcknowledgementTargetsEffectiveVersion(t *testing.T) {
	s := Specification{ID: "s", Version: 3, Status: SpecEffective}
	if _, err := Acknowledge(s, "u", "t", "a", 2, time.Now()); !errors.Is(err, ErrWrongVersionAcknowledgement) {
		t.Fatalf("got %v", err)
	}
	if _, err := Acknowledge(s, "u", "t", "a", 3, time.Now()); err != nil {
		t.Fatal(err)
	}
}
func TestRemediationClosureRequiresVerification(t *testing.T) {
	r := Remediation{Status: RemediationOpen, Revision: 1}
	if err := r.Close(); err == nil {
		t.Fatal("open remediation closed")
	}
	if err := r.Submit("replace guard", []string{"photo"}); err != nil {
		t.Fatal(err)
	}
	if err := r.Verify("inspector", "passed", true); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
}
