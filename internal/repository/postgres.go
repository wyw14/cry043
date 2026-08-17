package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wyw14/cry043/internal/domain"
)

type Postgres struct{ pool *pgxpool.Pool }

func NewPostgres(p *pgxpool.Pool) *Postgres { return &Postgres{p} }
func (p *Postgres) SaveSpecification(ctx context.Context, s domain.Specification, expected int64, key string) (domain.Specification, error) {
	b, _ := json.Marshal(s)
	if expected == 0 {
		_, err := p.pool.Exec(ctx, "insert into specifications(id,document_id,version,status,revision,idempotency_key,payload) values($1,$2,$3,$4,$5,$6,$7) on conflict(idempotency_key) do nothing", s.ID, s.DocumentID, s.Version, s.Status, s.Revision, key, b)
		if err != nil {
			return s, err
		}
	} else {
		tag, err := p.pool.Exec(ctx, "update specifications set status=$2,revision=$3,payload=$4 where id=$1 and revision=$5", s.ID, s.Status, s.Revision, b, expected)
		if err != nil {
			return s, err
		}
		if tag.RowsAffected() != 1 {
			return s, ErrRevision
		}
	}
	return p.Specification(ctx, s.ID)
}
func (p *Postgres) Specification(ctx context.Context, id string) (domain.Specification, error) {
	var b []byte
	if err := p.pool.QueryRow(ctx, "select payload from specifications where id=$1", id).Scan(&b); err != nil {
		return domain.Specification{}, err
	}
	var s domain.Specification
	return s, json.Unmarshal(b, &s)
}
func (p *Postgres) EffectiveInScope(ctx context.Context, scope domain.Scope) ([]domain.Specification, error) {
	rows, err := p.pool.Query(ctx, "select payload from specifications where status='effective'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Specification{}
	for rows.Next() {
		var b []byte
		if err = rows.Scan(&b); err != nil {
			return nil, err
		}
		var s domain.Specification
		if err = json.Unmarshal(b, &s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
func (p *Postgres) SaveException(ctx context.Context, e domain.ExceptionRequest) error {
	b, _ := json.Marshal(e)
	_, err := p.pool.Exec(ctx, "insert into exceptions(id,specification_id,rule_id,expires_at,payload) values($1,$2,$3,$4,$5) on conflict(id) do update set expires_at=excluded.expires_at,payload=excluded.payload", e.ID, e.SpecificationID, e.RuleID, e.ExpiresAt, b)
	return err
}
func (p *Postgres) Exceptions(ctx context.Context, id string) ([]domain.ExceptionRequest, error) {
	rows, err := p.pool.Query(ctx, "select payload from exceptions where specification_id=$1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ExceptionRequest{}
	for rows.Next() {
		var b []byte
		rows.Scan(&b)
		var e domain.ExceptionRequest
		if err = json.Unmarshal(b, &e); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
func (p *Postgres) SaveAcknowledgement(ctx context.Context, a domain.Acknowledgement) error {
	b, _ := json.Marshal(a)
	_, err := p.pool.Exec(ctx, "insert into acknowledgements(id,specification_id,version,user_id,payload) values($1,$2,$3,$4,$5) on conflict(specification_id,version,user_id) do nothing", a.ID, a.SpecificationID, a.Version, a.UserID, b)
	return err
}
func (p *Postgres) AcknowledgementCount(ctx context.Context, id string, v int) (int, error) {
	var n int
	return n, p.pool.QueryRow(ctx, "select count(*) from acknowledgements where specification_id=$1 and version=$2", id, v).Scan(&n)
}
func (p *Postgres) SaveInspection(ctx context.Context, i domain.Inspection) error {
	b, _ := json.Marshal(i)
	_, err := p.pool.Exec(ctx, "insert into inspections(id,specification_id,specification_version,payload,performed_at) values($1,$2,$3,$4,$5)", i.ID, i.SpecificationID, i.SpecificationVersion, b, i.PerformedAt)
	return err
}
func (p *Postgres) SaveRemediation(ctx context.Context, r domain.Remediation, expected int64) error {
	b, _ := json.Marshal(r)
	if expected == 0 {
		_, err := p.pool.Exec(ctx, "insert into remediations(id,status,revision,due_at,payload) values($1,$2,$3,$4,$5)", r.ID, r.Status, r.Revision, r.DueAt, b)
		return err
	}
	tag, err := p.pool.Exec(ctx, "update remediations set status=$2,revision=$3,payload=$4 where id=$1", r.ID, r.Status, r.Revision, b)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrRevision
	}
	return err
}
func (p *Postgres) Remediation(ctx context.Context, id string) (domain.Remediation, error) {
	var b []byte
	if err := p.pool.QueryRow(ctx, "select payload from remediations where id=$1", id).Scan(&b); err != nil {
		return domain.Remediation{}, err
	}
	var r domain.Remediation
	return r, json.Unmarshal(b, &r)
}
func (p *Postgres) AppendAudit(ctx context.Context, e domain.AuditEvent) error {
	b, _ := json.Marshal(e)
	_, err := p.pool.Exec(ctx, "insert into audit_events(id,aggregate_id,previous_hash,hash,payload,created_at) values($1,$2,$3,$4,$5,$6)", e.ID, e.AggregateID, e.PreviousHash, e.Hash, b, e.At)
	return err
}
func (p *Postgres) AuditHead(ctx context.Context, id string) (string, error) {
	var h string
	err := p.pool.QueryRow(ctx, "select hash from audit_events where aggregate_id=$1 order by created_at desc,id desc limit 1", id).Scan(&h)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return h, err
}

func (p *Postgres) RiskFacts(ctx context.Context, window domain.RiskWindow) (domain.RiskFacts, error) {
	facts := domain.RiskFacts{}
	if err := p.scanPayloads(ctx, "select payload from specifications where status in ('effective','approved','scheduled')", func(payload []byte) error {
		var value domain.Specification
		if err := json.Unmarshal(payload, &value); err != nil {
			return err
		}
		facts.Specifications = append(facts.Specifications, value)
		return nil
	}); err != nil {
		return facts, err
	}
	if err := p.scanPayloads(ctx, "select payload from acknowledgements", func(payload []byte) error {
		var value domain.Acknowledgement
		if err := json.Unmarshal(payload, &value); err != nil {
			return err
		}
		facts.Acknowledgements = append(facts.Acknowledgements, value)
		return nil
	}); err != nil {
		return facts, err
	}
	if err := p.scanPayloads(ctx, "select payload from exceptions where expires_at <= $1", func(payload []byte) error {
		var value domain.ExceptionRequest
		if err := json.Unmarshal(payload, &value); err != nil {
			return err
		}
		facts.Exceptions = append(facts.Exceptions, value)
		return nil
	}, window.HorizonEnd); err != nil {
		return facts, err
	}
	if err := p.scanPayloads(ctx, "select payload from inspections", func(payload []byte) error {
		var value domain.Inspection
		if err := json.Unmarshal(payload, &value); err != nil {
			return err
		}
		facts.Inspections = append(facts.Inspections, value)
		return nil
	}); err != nil {
		return facts, err
	}
	if err := p.scanPayloads(ctx, "select payload from remediations where status <> 'closed'", func(payload []byte) error {
		var value domain.Remediation
		if err := json.Unmarshal(payload, &value); err != nil {
			return err
		}
		facts.Remediations = append(facts.Remediations, value)
		return nil
	}); err != nil {
		return facts, err
	}
	return facts, nil
}

func (p *Postgres) scanPayloads(ctx context.Context, query string, consume func([]byte) error, args ...any) error {
	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return err
		}
		if err := consume(payload); err != nil {
			return err
		}
	}
	return rows.Err()
}
