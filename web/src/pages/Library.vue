<script setup lang="ts">
import { onMounted } from 'vue'
import MetricTile from '../components/MetricTile.vue'
import RiskLane from '../components/RiskLane.vue'
import { useControlRoom } from '../stores/controlRoom'

const room = useControlRoom()
onMounted(room.refresh)
const percent = (value: number) => `${Math.round(value * 100)}%`
</script>

<template>
  <section>
    <div class="metrics">
      <MetricTile label="当前版本确认覆盖" :value="percent(room.board.summary.confirmation_coverage)" />
      <MetricTile label="待班组确认" :value="room.board.summary.pending_confirmations" tone="amber" />
      <MetricTile label="七日内例外到期" :value="room.board.summary.expiring_exceptions" tone="amber" />
      <MetricTile label="整改已逾期" :value="room.board.summary.overdue_remediations" tone="red" />
      <MetricTile label="生效闸门阻断" :value="room.board.summary.blocked_activations" tone="red" />
    </div>
    <div class="panel">
      <form class="toolbar" @submit.prevent="room.refresh">
        <label>作业区域<input v-model.trim="room.area" placeholder="例如 welding" /></label>
        <label>工序<input v-model.trim="room.process" placeholder="例如 hot-work" /></label>
        <button class="button" type="submit">重算风险</button>
        <span v-if="room.loading">正在关联确认、例外与整改…</span>
        <span v-if="room.failure" class="warning">{{ room.failure }}</span>
      </form>
      <div v-if="Object.keys(room.byScope).length" class="lanes">
        <RiskLane v-for="(items, scope) in room.byScope" :key="scope" :scope="scope" :items="items" />
      </div>
      <div v-else-if="!room.loading" class="empty">当前筛选范围没有待处置风险</div>
    </div>
  </section>
</template>
