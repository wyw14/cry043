import { describe, expect, it } from 'vitest'
import type { ScopeRisk } from '../types/controlRoom'

function orderRisk(items: ScopeRisk[]) {
  return [...items].sort((left, right) => left.severity === right.severity
    ? Date.parse(left.due_at) - Date.parse(right.due_at)
    : left.severity === 'critical' ? -1 : 1)
}

describe('compliance risk ordering', () => {
  it('keeps overdue corrective action ahead of later confirmation work', () => {
    const base = { area_id: '焊装区', process_id: '动火', specification_id: 'spec', specification_name: '动火规范', version: 3, owner_id: 'night', detail: '' }
    const values: ScopeRisk[] = [
      { ...base, key: 'ack', kind: 'pending_team_confirmation', severity: 'major', due_at: '2026-08-19T08:00:00Z' },
      { ...base, key: 'fix', kind: 'overdue_remediation', severity: 'critical', due_at: '2026-08-17T08:00:00Z' },
    ]
    expect(orderRisk(values).map(value => value.key)).toEqual(['fix', 'ack'])
  })
})
