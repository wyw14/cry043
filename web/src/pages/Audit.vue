<script setup lang="ts">
import { ref } from 'vue'
const reason = ref('')
const exported = ref(false)
const events = [
  { at: '07:42', action: 'spec.simulated', actor: 'author-2', hash: 'b472…9c0a' },
  { at: '07:51', action: 'exception.reviewed', actor: 'safety-1', hash: '9c0a…781f' },
  { at: '08:00', action: 'spec.scheduled', actor: 'approver-1', hash: '781f…12de' },
]
function controlledExport() { if (reason.value.trim()) exported.value = true }
</script>
<template><div class="grid-two"><section class="panel workspace-card"><h2>哈希链审计时间线</h2><div v-for="event in events" :key="event.hash" class="timeline-row"><time>{{ event.at }}</time><strong>{{ event.action }}</strong><p>{{ event.actor }} · {{ event.hash }}</p></div></section><form class="panel workspace-card form-grid" @submit.prevent="controlledExport"><h2 class="wide">受控本地导出</h2><label class="wide">导出原因<textarea v-model.trim="reason" rows="4" placeholder="说明审计用途和数据范围" /></label><p class="wide">敏感字段会脱敏，导出动作本身继续写入审计链。</p><button class="button wide">生成离线 CSV</button><p v-if="exported" class="wide badge">已生成 export/compliance-audit.csv</p></form></div></template>
