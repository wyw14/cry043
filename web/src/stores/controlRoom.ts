import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { complianceGateway } from '../services/api'
import type { RiskBoard, ScopeRisk } from '../types/controlRoom'

const emptyBoard = (): RiskBoard => ({
  generated_at: '', window_end: '',
  summary: { pending_confirmations: 0, expiring_exceptions: 0, overdue_remediations: 0, blocked_activations: 0, confirmation_coverage: 0 },
  items: [],
})

export const useControlRoom = defineStore('compliance-control-room', () => {
  const board = ref<RiskBoard>(emptyBoard())
  const loading = ref(false)
  const failure = ref('')
  const area = ref('')
  const process = ref('')

  const urgent = computed(() => board.value.items.filter(item => item.severity === 'critical'))
  const byScope = computed(() => board.value.items.reduce<Record<string, ScopeRisk[]>>((groups, item) => {
    const key = `${item.area_id} / ${item.process_id}`
    ;(groups[key] ??= []).push(item)
    return groups
  }, {}))

  async function refresh() {
    loading.value = true
    failure.value = ''
    try {
      board.value = await complianceGateway.riskBoard({ area: area.value, process: process.value, horizonDays: 14 })
    } catch (error) {
      failure.value = error instanceof Error ? error.message : '风险看板加载失败'
    } finally {
      loading.value = false
    }
  }

  return { board, loading, failure, area, process, urgent, byScope, refresh }
})
