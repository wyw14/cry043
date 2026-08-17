package domain

import (
	"errors"
	"time"
)

var (
	ErrSafetyException             = errors.New("mandatory safety rules cannot be excepted")
	ErrExceptionExpired            = errors.New("exception expiry must be in the future")
	ErrWrongVersionAcknowledgement = errors.New("acknowledgement must target the effective version")
)

type ExceptionRequest struct {
	ID, SpecificationID, RuleID, RequesterID, Reason, CompensatingControl, ApproverID string
	ExpiresAt                                                                         time.Time
	ApprovedAt                                                                        *time.Time
	ReviewedAt                                                                        *time.Time
}

func (e *ExceptionRequest) Approve(rule Rule, approver string, now time.Time) error {
	if rule.MandatorySafety {
		return ErrSafetyException
	}
	if !e.ExpiresAt.After(now) {
		return ErrExceptionExpired
	}
	if approver == "" {
		return errors.New("approver required")
	}
	e.ApproverID = approver
	e.ApprovedAt = &now
	return nil
}
func (e ExceptionRequest) Active(at time.Time) bool {
	return e.ApprovedAt != nil && e.ApproverID != "" && e.ExpiresAt.After(at)
}

type Acknowledgement struct {
	ID, SpecificationID, UserID, TeamID string
	Version                             int
	ReadAt                              time.Time
}

func Acknowledge(spec Specification, user, team, id string, version int, now time.Time) (Acknowledgement, error) {
	if spec.Status != SpecEffective || version != spec.Version {
		return Acknowledgement{}, ErrWrongVersionAcknowledgement
	}
	return Acknowledgement{id, spec.ID, user, team, version, now}, nil
}

type ComplianceInput struct {
	Fields             map[string]string
	ActiveCombinations []string
	CompletedAt        time.Time
	LastReviewedAt     time.Time
}
type Violation struct {
	RuleID, Code, Message string
	Severity              string
}

func Evaluate(spec Specification, input ComplianceInput, exceptions []ExceptionRequest, now time.Time) []Violation {
	active := map[string]bool{}
	for _, e := range exceptions {
		if e.Active(now) {
			active[e.RuleID] = true
		}
	}
	present := map[string]bool{}
	for _, v := range input.ActiveCombinations {
		present[v] = true
	}
	out := []Violation{}
	for _, rule := range spec.Rules {
		failed, code, message := false, "", ""
		switch rule.Kind {
		case RuleRequired:
			failed = input.Fields[rule.RequiredField] == ""
			code = "REQUIRED_MISSING"
			message = "必填检查缺失"
		case RuleValidity:
			failed = rule.ValidDays > 0 && input.CompletedAt.Before(now.AddDate(0, 0, -rule.ValidDays))
			code = "VALIDITY_EXPIRED"
			message = "有效期已过"
		case RuleForbiddenCombination:
			for _, item := range rule.ForbiddenWith {
				if present[item] {
					failed = true
				}
			}
			code = "FORBIDDEN_COMBINATION"
			message = "命中禁止组合"
		case RuleReviewFrequency:
			failed = rule.ReviewEveryDays > 0 && input.LastReviewedAt.Before(now.AddDate(0, 0, -rule.ReviewEveryDays))
			code = "REVIEW_OVERDUE"
			message = "复核已逾期"
		}
		if failed && (!active[rule.ID] || rule.MandatorySafety) {
			severity := "major"
			if rule.MandatorySafety {
				severity = "critical"
			}
			out = append(out, Violation{rule.ID, code, message, severity})
		}
	}
	return out
}
