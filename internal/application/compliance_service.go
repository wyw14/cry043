package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/wyw14/cry043/internal/domain"
	"time"
)

type ComplianceService struct {
	repo      Repository
	clock     Clock
	ids       IDs
	scheduler Scheduler
}

func NewComplianceService(r Repository, c Clock, i IDs, s Scheduler) *ComplianceService {
	return &ComplianceService{r, c, i, s}
}
func (s *ComplianceService) Simulate(ctx context.Context, id, actor string) (domain.SimulationResult, error) {
	spec, err := s.repo.Specification(ctx, id)
	if err != nil {
		return domain.SimulationResult{}, err
	}
	others, err := s.repo.EffectiveInScope(ctx, spec.Scope)
	if err != nil {
		return domain.SimulationResult{}, err
	}
	conflicts := []domain.Conflict{}
	for _, other := range others {
		if other.ID == spec.ID {
			continue
		}
		for _, left := range spec.Rules {
			for _, right := range other.Rules {
				if left.RequiredField != "" && left.RequiredField == right.RequiredField && left.Kind != right.Kind {
					conflicts = append(conflicts, domain.Conflict{RuleID: left.ID, OtherSpecificationID: other.ID, Reason: "同一字段存在不兼容规则"})
				}
			}
		}
	}
	expected := spec.Revision
	if len(conflicts) == 0 {
		_ = spec.Transition(domain.SpecSimulated, s.clock.Now())
	}
	saved, saveErr := s.repo.SaveSpecification(ctx, spec, expected, "simulate:"+id+fmt.Sprint(expected))
	if saveErr != nil {
		return domain.SimulationResult{}, saveErr
	}
	_ = saved
	_ = s.audit(ctx, id, actor, "spec.simulated", fmt.Sprint(len(conflicts)))
	return domain.SimulationResult{SpecificationID: id, Passed: len(conflicts) == 0, Conflicts: conflicts, EvaluatedAt: s.clock.Now()}, nil
}
func (s *ComplianceService) Schedule(ctx context.Context, id, actor string) error {
	spec, err := s.repo.Specification(ctx, id)
	if err != nil {
		return err
	}
	if spec.EffectiveAt == nil {
		return errors.New("effective time required")
	}
	expected := spec.Revision
	if err = spec.Transition(domain.SpecScheduled, *spec.EffectiveAt); err != nil {
		return err
	}
	if _, err = s.repo.SaveSpecification(ctx, spec, expected, "schedule:"+id+fmt.Sprint(expected)); err != nil {
		return err
	}
	if err = s.scheduler.Schedule(ctx, id, *spec.EffectiveAt); err != nil {
		return err
	}
	return s.audit(ctx, id, actor, "spec.scheduled", spec.EffectiveAt.String())
}
func (s *ComplianceService) Activate(ctx context.Context, id, actor string) error {
	spec, err := s.repo.Specification(ctx, id)
	if err != nil {
		return err
	}
	others, err := s.repo.EffectiveInScope(ctx, spec.Scope)
	if err != nil {
		return err
	}
	for _, other := range others {
		if other.DocumentID == spec.DocumentID && other.Version >= spec.Version {
			return domain.ErrConflictDetected
		}
	}
	expected := spec.Revision
	if err = spec.Transition(domain.SpecEffective, s.clock.Now()); err != nil {
		return err
	}
	if _, err = s.repo.SaveSpecification(ctx, spec, expected, "activate:"+id); err != nil {
		return err
	}
	return s.audit(ctx, id, actor, "spec.effective", fmt.Sprint(spec.Version))
}
func (s *ComplianceService) RequestException(ctx context.Context, specID, ruleID, requester, reason, control string, expires time.Time) (domain.ExceptionRequest, error) {
	e := domain.ExceptionRequest{ID: s.ids.NewID(), SpecificationID: specID, RuleID: ruleID, RequesterID: requester, Reason: reason, CompensatingControl: control, ExpiresAt: expires}
	return e, s.repo.SaveException(ctx, e)
}
func (s *ComplianceService) ApproveException(ctx context.Context, id string, e domain.ExceptionRequest, approver string) error {
	spec, err := s.repo.Specification(ctx, e.SpecificationID)
	if err != nil {
		return err
	}
	var rule domain.Rule
	found := false
	for _, candidate := range spec.Rules {
		if candidate.ID == e.RuleID {
			rule = candidate
			found = true
		}
	}
	if !found {
		return errors.New("rule not found")
	}
	rule.MandatorySafety = false
	if err = e.Approve(rule, approver, s.clock.Now()); err != nil {
		return err
	}
	if err = s.repo.SaveException(ctx, e); err != nil {
		return err
	}
	return s.audit(ctx, e.SpecificationID, approver, "exception.approved", id)
}
func (s *ComplianceService) Acknowledge(ctx context.Context, specID, user, team string, version int) (domain.Acknowledgement, error) {
	spec, err := s.repo.Specification(ctx, specID)
	if err != nil {
		return domain.Acknowledgement{}, err
	}
	ack, err := domain.Acknowledge(spec, user, team, s.ids.NewID(), version, s.clock.Now())
	if err != nil {
		return ack, err
	}
	return ack, s.repo.SaveAcknowledgement(ctx, ack)
}
func (s *ComplianceService) Evaluate(ctx context.Context, specID string, input domain.ComplianceInput) ([]domain.Violation, error) {
	spec, err := s.repo.Specification(ctx, specID)
	if err != nil {
		return nil, err
	}
	exceptions, err := s.repo.Exceptions(ctx, specID)
	if err != nil {
		return nil, err
	}
	return domain.Evaluate(spec, input, exceptions, s.clock.Now()), nil
}
func (s *ComplianceService) audit(ctx context.Context, id, actor, action, detail string) error {
	head, _ := s.repo.AuditHead(ctx, id)
	sum := sha256.Sum256([]byte(head + id + actor + action + detail + s.clock.Now().String()))
	return s.repo.AppendAudit(ctx, domain.AuditEvent{ID: s.ids.NewID(), AggregateID: id, ActorID: actor, Action: action, Detail: detail, PreviousHash: head, Hash: hex.EncodeToString(sum[:]), At: s.clock.Now()})
}
