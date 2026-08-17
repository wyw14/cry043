package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/wyw14/cry043/internal/domain"
	"slices"
	"sync"
)

var ErrNotFound = errors.New("not found")
var ErrRevision = errors.New("revision conflict")

type Memory struct {
	mu           sync.RWMutex
	specs        map[string]domain.Specification
	keys         map[string]string
	exceptions   map[string][]domain.ExceptionRequest
	acks         map[string]domain.Acknowledgement
	inspections  map[string]domain.Inspection
	remediations map[string]domain.Remediation
	audits       map[string][]domain.AuditEvent
}

func NewMemory() *Memory {
	return &Memory{specs: map[string]domain.Specification{}, keys: map[string]string{}, exceptions: map[string][]domain.ExceptionRequest{}, acks: map[string]domain.Acknowledgement{}, inspections: map[string]domain.Inspection{}, remediations: map[string]domain.Remediation{}, audits: map[string][]domain.AuditEvent{}}
}
func (m *Memory) SaveSpecification(ctx context.Context, s domain.Specification, expected int64, key string) (domain.Specification, error) {
	if err := ctx.Err(); err != nil {
		return s, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if id := m.keys[key]; id != "" {
		return m.specs[id].Clone(), nil
	}
	old, ok := m.specs[s.ID]
	if ok {
		if expected == 0 || old.Revision != expected {
			return s, ErrRevision
		}
	} else if expected != 0 {
		return s, ErrRevision
	}
	m.specs[s.ID] = s.Clone()
	m.keys[key] = s.ID
	return s.Clone(), nil
}
func (m *Memory) Specification(ctx context.Context, id string) (domain.Specification, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.specs[id]
	if !ok {
		return s, ErrNotFound
	}
	return s.Clone(), ctx.Err()
}
func intersects(a, b []string) bool {
	for _, x := range a {
		if slices.Contains(b, x) {
			return true
		}
	}
	return false
}
func (m *Memory) EffectiveInScope(ctx context.Context, scope domain.Scope) ([]domain.Specification, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []domain.Specification{}
	for _, s := range m.specs {
		if s.Status == domain.SpecEffective && (intersects(scope.AreaIDs, s.Scope.AreaIDs) || intersects(scope.ProcessIDs, s.Scope.ProcessIDs)) {
			out = append(out, s.Clone())
		}
	}
	return out, ctx.Err()
}
func (m *Memory) SaveException(ctx context.Context, e domain.ExceptionRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	list := m.exceptions[e.SpecificationID]
	for i := range list {
		if list[i].ID == e.ID {
			list[i] = e
			m.exceptions[e.SpecificationID] = list
			return ctx.Err()
		}
	}
	m.exceptions[e.SpecificationID] = append(list, e)
	return ctx.Err()
}
func (m *Memory) Exceptions(ctx context.Context, id string) ([]domain.ExceptionRequest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]domain.ExceptionRequest(nil), m.exceptions[id]...), ctx.Err()
}
func (m *Memory) SaveAcknowledgement(ctx context.Context, a domain.Acknowledgement) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.acks[fmt.Sprintf("%s:%d:%s", a.SpecificationID, a.Version, a.UserID)] = a
	return ctx.Err()
}
func (m *Memory) AcknowledgementCount(ctx context.Context, id string, v int) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, a := range m.acks {
		if a.SpecificationID == id && a.Version == v {
			count++
		}
	}
	return count, ctx.Err()
}
func (m *Memory) SaveInspection(ctx context.Context, i domain.Inspection) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inspections[i.ID] = i
	return ctx.Err()
}
func (m *Memory) SaveRemediation(ctx context.Context, r domain.Remediation, expected int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	old, ok := m.remediations[r.ID]
	if ok && old.Revision != expected {
		return ErrRevision
	}
	if !ok && expected != 0 {
		return ErrRevision
	}
	m.remediations[r.ID] = r
	return ctx.Err()
}
func (m *Memory) Remediation(ctx context.Context, id string) (domain.Remediation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.remediations[id]
	if !ok {
		return r, ErrNotFound
	}
	return r, ctx.Err()
}
func (m *Memory) AppendAudit(ctx context.Context, e domain.AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	list := m.audits[e.AggregateID]
	if len(list) > 0 && list[len(list)-1].Hash != e.PreviousHash {
		return errors.New("audit chain conflict")
	}
	m.audits[e.AggregateID] = append(list, e)
	return ctx.Err()
}
func (m *Memory) AuditHead(ctx context.Context, id string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := m.audits[id]
	if len(list) == 0 {
		return "", ctx.Err()
	}
	return list[len(list)-1].Hash, ctx.Err()
}

func (m *Memory) RiskFacts(ctx context.Context, _ domain.RiskWindow) (domain.RiskFacts, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return domain.RiskFacts{}, err
	}
	facts := domain.RiskFacts{}
	for _, spec := range m.specs {
		facts.Specifications = append(facts.Specifications, spec.Clone())
	}
	for _, ack := range m.acks {
		facts.Acknowledgements = append(facts.Acknowledgements, ack)
	}
	for _, list := range m.exceptions {
		facts.Exceptions = append(facts.Exceptions, list...)
	}
	for _, inspection := range m.inspections {
		facts.Inspections = append(facts.Inspections, inspection)
	}
	for _, remediation := range m.remediations {
		facts.Remediations = append(facts.Remediations, remediation)
	}
	return facts, nil
}
