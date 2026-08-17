export type RiskKind = 'pending_team_confirmation' | 'expiring_exception' | 'overdue_remediation' | 'blocked_activation'

export interface ScopeRisk {
  key: string
  area_id: string
  process_id: string
  specification_id: string
  specification_name: string
  version: number
  kind: RiskKind
  severity: 'major' | 'critical'
  owner_id: string
  due_at: string
  detail: string
}

export interface RiskSummary {
  pending_confirmations: number
  expiring_exceptions: number
  overdue_remediations: number
  blocked_activations: number
  confirmation_coverage: number
}

export interface RiskBoard {
  generated_at: string
  window_end: string
  summary: RiskSummary
  items: ScopeRisk[]
}

export interface ActivationReadiness {
  specification_id: string
  version: number
  ready: boolean
  required_teams: number
  confirmed_teams: number
  coverage: number
  blockers: Array<{ code: string; detail: string }>
}
