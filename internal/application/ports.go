package application

import (
	"context"
	"github.com/wyw14/cry043/internal/domain"
	"time"
)

type Repository interface {
	SaveSpecification(context.Context, domain.Specification, int64, string) (domain.Specification, error)
	Specification(context.Context, string) (domain.Specification, error)
	EffectiveInScope(context.Context, domain.Scope) ([]domain.Specification, error)
	SaveException(context.Context, domain.ExceptionRequest) error
	Exceptions(context.Context, string) ([]domain.ExceptionRequest, error)
	SaveAcknowledgement(context.Context, domain.Acknowledgement) error
	AcknowledgementCount(context.Context, string, int) (int, error)
	SaveInspection(context.Context, domain.Inspection) error
	SaveRemediation(context.Context, domain.Remediation, int64) error
	Remediation(context.Context, string) (domain.Remediation, error)
	AppendAudit(context.Context, domain.AuditEvent) error
	AuditHead(context.Context, string) (string, error)
	RiskFacts(context.Context, domain.RiskWindow) (domain.RiskFacts, error)
}
type Clock interface{ Now() time.Time }
type IDs interface{ NewID() string }
type Scheduler interface {
	Schedule(context.Context, string, time.Time) error
}
type Exporter interface {
	Export(context.Context, domain.Specification, []domain.AuditEvent) (string, error)
}
