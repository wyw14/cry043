<script setup lang="ts">
import { ref } from 'vue'
import { complianceGateway } from '../services/api'
import type { ActivationReadiness } from '../types/controlRoom'

const specificationId = ref('seed-spec')
const state = ref<ActivationReadiness | null>(null)
const message = ref('')
async function inspectGate() {
  message.value = ''
  try { state.value = await complianceGateway.activationReadiness(specificationId.value) }
  catch (error) { message.value = error instanceof Error ? error.message : '闸门检查失败' }
}
</script>

<template>
  <div class="grid-two">
    <section class="panel workspace-card">
      <h2>激活计划</h2>
      <p>审批完成后仍要核对生效范围、规则冲突和班组覆盖；定时器只执行已通过闸门的版本。</p>
      <form class="toolbar" @submit.prevent="inspectGate"><label>规范 ID<input v-model.trim="specificationId" /></label><button class="button">运行影响模拟</button></form>
      <p v-if="message" class="warning">{{ message }}</p>
    </section>
    <section class="panel workspace-card">
      <h2>生效就绪度</h2>
      <div v-if="state">
        <p><span class="badge">V{{ state.version }}</span> {{ state.ready ? '允许进入生效队列' : '仍有阻断项' }}</p>
        <div class="progress"><i :style="{ width: `${state.coverage * 100}%` }" /></div>
        <small>班组确认 {{ state.confirmed_teams }} / {{ state.required_teams }}</small>
        <div v-for="blocker in state.blockers" :key="blocker.code" class="gate-row"><strong>{{ blocker.code }}</strong><p>{{ blocker.detail }}</p></div>
      </div>
      <div v-else class="empty">运行模拟后显示冲突、作用域和激活计划</div>
    </section>
  </div>
</template>
