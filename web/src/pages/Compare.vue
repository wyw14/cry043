<script setup lang="ts">
const changes = [
  { field: 'permit_no', before: '有效期 30 天', after: '有效期 7 天', impact: '焊装区 3 个班组需重新确认' },
  { field: 'ppe_guard', before: '普通规则', after: '强制安全规则', impact: '现有例外自动进入阻断复核' },
  { field: 'review_cycle', before: '每 14 天', after: '每 7 天', impact: '2 条抽查计划将提前' },
]
</script>
<template><section class="panel workspace-card"><h2>V3 → V4 语义差异</h2><p>生效版本不能原地覆盖；新版本记录替代关系，回滚也生成受审计的状态事件。</p><div v-for="change in changes" :key="change.field" class="gate-row"><span class="badge">{{ change.field }}</span><div class="grid-two"><p><del>{{ change.before }}</del> → <strong>{{ change.after }}</strong></p><p>{{ change.impact }}</p></div></div><button class="button">生成受审计回滚计划</button></section></template>
