<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useControlRoom } from '../stores/controlRoom'
const room = useControlRoom()
onMounted(() => { if (!room.board.generated_at) room.refresh() })
const overdue = computed(() => room.board.items.filter(item => item.kind === 'overdue_remediation'))
</script>
<template><section class="panel workspace-card"><h2>按期限与级别排列的整改队列</h2><p>不符合项提交行动与证据后进入复核；只有复核结论通过才能关闭。</p><div v-for="item in overdue" :key="item.key" class="timeline-row"><span class="badge">{{ item.severity === 'critical' ? '严重逾期' : '已逾期' }}</span><h3>{{ item.specification_name }} · {{ item.area_id }}/{{ item.process_id }}</h3><p>{{ item.detail }} · 责任人 {{ item.owner_id }}</p><button class="button secondary">进入复核</button></div><div v-if="!overdue.length" class="empty">没有逾期整改</div></section></template>
