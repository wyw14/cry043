<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const navigation = [
  { to: '/', label: '风险驾驶舱', code: '01' },
  { to: '/rules', label: '规则工作室', code: '02' },
  { to: '/impact', label: '生效闸门', code: '03' },
  { to: '/confirm', label: '确认中心', code: '04' },
  { to: '/inspections', label: '现场抽查', code: '05' },
  { to: '/remediations', label: '整改复核', code: '06' },
  { to: '/compare', label: '版本谱系', code: '07' },
  { to: '/audit', label: '审计导出', code: '08' },
]
const pageTitle = computed(() => String(route.meta.title ?? '生产作业规范台'))
</script>

<template>
  <div class="workbench">
    <header class="masthead">
      <div class="brand-mark">规</div>
      <div>
        <p class="eyebrow">SHOPFLOOR COMPLIANCE CONTROL</p>
        <strong>生产作业规范台</strong>
      </div>
      <div class="shift-status"><i /> 本地模式 · 早班</div>
    </header>
    <aside class="rail">
      <RouterLink v-for="item in navigation" :key="item.to" :to="item.to">
        <span>{{ item.code }}</span>{{ item.label }}
      </RouterLink>
    </aside>
    <main class="stage">
      <div class="page-heading">
        <div><p class="eyebrow">CURRENT WORKSPACE</p><h1>{{ pageTitle }}</h1></div>
        <time>2026-08-18 · 08:00</time>
      </div>
      <RouterView />
    </main>
  </div>
</template>
