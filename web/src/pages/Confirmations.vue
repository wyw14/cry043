<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useControlRoom } from '../stores/controlRoom'

const room = useControlRoom()
onMounted(() => { if (!room.board.generated_at) room.refresh() })
const confirmationWork = computed(() => room.board.items.filter(item => item.kind === 'pending_team_confirmation' || item.kind === 'expiring_exception'))
</script>
<template>
  <section class="panel workspace-card">
    <h2>按当前生效版本派发确认</h2>
    <p>班组确认与个人阅读确认分别留痕；旧版本回执不会计入覆盖率，例外到期前必须重新复核。</p>
    <div v-for="item in confirmationWork" :key="item.key" class="timeline-row">
      <span class="badge">{{ item.kind === 'expiring_exception' ? '到期复核' : '待确认' }}</span>
      <h3>{{ item.area_id }} / {{ item.process_id }} · {{ item.specification_name }} V{{ item.version }}</h3>
      <p>责任对象：{{ item.owner_id }}　截止：{{ new Date(item.due_at).toLocaleString('zh-CN') }}</p>
    </div>
    <div v-if="!confirmationWork.length" class="empty">没有待确认或即将到期的例外</div>
  </section>
</template>
