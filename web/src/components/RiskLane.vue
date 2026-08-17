<script setup lang="ts">
import type { ScopeRisk } from '../types/controlRoom'
defineProps<{ scope: string; items: ScopeRisk[] }>()
const labels: Record<ScopeRisk['kind'], string> = { pending_team_confirmation: '待确认', expiring_exception: '例外到期', overdue_remediation: '整改逾期', blocked_activation: '生效受阻' }
</script>
<template>
  <article class="risk-lane">
    <header><h3>{{ scope }}</h3><b>{{ items.length }} 项</b></header>
    <div v-for="item in items" :key="item.key" class="risk-row" :data-severity="item.severity">
      <span>{{ labels[item.kind] }}</span>
      <div><strong>{{ item.specification_name }} · V{{ item.version }}</strong><small>{{ item.detail }}</small></div>
      <time>{{ new Date(item.due_at).toLocaleDateString('zh-CN') }}</time>
    </div>
  </article>
</template>
