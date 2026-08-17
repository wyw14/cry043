import type { ActivationReadiness, RiskBoard } from '../types/controlRoom'

type Envelope<T> = { data: T; request_id: string }
type Failure = { code: string; message: string; field_errors?: Record<string, string>; request_id?: string }

export class ComplianceApiError extends Error {
  constructor(public code: string, message: string, public requestId = '', public fields: Record<string, string> = {}) {
    super(message)
  }
}

class ComplianceGateway {
  private async request<T>(path: string, init: RequestInit = {}): Promise<T> {
    const response = await fetch(`/api/v1${path}`, {
      ...init,
      headers: {
        'Content-Type': 'application/json',
        'X-Actor-ID': 'demo-supervisor',
        'X-Actor-Role': 'supervisor',
        ...init.headers,
      },
    })
    const payload = await response.json() as Envelope<T> | Failure
    if (!response.ok) {
      const failure = payload as Failure
      throw new ComplianceApiError(failure.code, failure.message, failure.request_id, failure.field_errors)
    }
    return (payload as Envelope<T>).data
  }

  riskBoard(filters: { area?: string; process?: string; horizonDays?: number } = {}) {
    const query = new URLSearchParams()
    if (filters.area) query.set('area_id', filters.area)
    if (filters.process) query.set('process_id', filters.process)
    query.set('horizon_days', String(filters.horizonDays ?? 14))
    return this.request<RiskBoard>(`/control-room/risks?${query}`)
  }

  activationReadiness(specificationId: string) {
    return this.request<ActivationReadiness>(`/specifications/${encodeURIComponent(specificationId)}/activation-readiness`)
  }
}

export const complianceGateway = new ComplianceGateway()
