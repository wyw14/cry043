package application

import (
	"context"
	"errors"
	"github.com/wyw14/cry043/internal/domain"
	"time"
)

type InspectionService struct {
	repo  Repository
	clock Clock
	ids   IDs
}

func NewInspectionService(r Repository, c Clock, i IDs) *InspectionService {
	return &InspectionService{r, c, i}
}
func (s *InspectionService) Record(ctx context.Context, in domain.Inspection, actor string) ([]domain.Remediation, error) {
	spec, err := s.repo.Specification(ctx, in.SpecificationID)
	if err != nil {
		return nil, err
	}
	if spec.Status != domain.SpecEffective || spec.Version != in.SpecificationVersion {
		return nil, domain.ErrWrongVersionAcknowledgement
	}
	in.ID = s.ids.NewID()
	in.InspectorID = actor
	in.PerformedAt = s.clock.Now()
	if len(in.EvidenceIDs) == 0 {
		return nil, errors.New("inspection evidence is required")
	}
	if err = s.repo.SaveInspection(ctx, in); err != nil {
		return nil, err
	}
	result := []domain.Remediation{}
	for _, finding := range in.Findings {
		r := domain.Remediation{ID: s.ids.NewID(), FindingID: finding.ID, OwnerID: "unassigned", DueAt: deadline(finding.Level, s.clock.Now()), Status: domain.RemediationOpen, Revision: 1}
		if err = s.repo.SaveRemediation(ctx, r, 0); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}
func deadline(level domain.IssueLevel, now time.Time) time.Time {
	switch level {
	case domain.IssueCritical:
		return now.Add(24 * time.Hour)
	case domain.IssueMajor:
		return now.Add(72 * time.Hour)
	default:
		return now.Add(168 * time.Hour)
	}
}
func (s *InspectionService) Submit(ctx context.Context, id, action string, evidence []string) error {
	r, err := s.repo.Remediation(ctx, id)
	if err != nil {
		return err
	}
	expected := r.Revision
	if err = r.Submit(action, evidence); err != nil {
		return err
	}
	return s.repo.SaveRemediation(ctx, r, expected)
}
func (s *InspectionService) Verify(ctx context.Context, id, reviewer, conclusion string, pass bool) error {
	r, err := s.repo.Remediation(ctx, id)
	if err != nil {
		return err
	}
	expected := r.Revision
	if err = r.Verify(reviewer, conclusion, pass); err != nil {
		return err
	}
	return s.repo.SaveRemediation(ctx, r, expected)
}
func (s *InspectionService) Close(ctx context.Context, id string) error {
	r, err := s.repo.Remediation(ctx, id)
	if err != nil {
		return err
	}
	if err = r.Close(); err != nil {
		return err
	}
	expected := r.Revision
	return s.repo.SaveRemediation(ctx, r, expected)
}
