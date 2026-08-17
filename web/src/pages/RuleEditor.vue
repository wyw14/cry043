<script setup lang="ts">
import { computed, ref } from 'vue'

type DraftRule = { kind: string; title: string; field: string; safety: boolean }
const rules = ref<DraftRule[]>([
  { kind: '必填检查', title: '动火许可编号', field: 'permit_no', safety: true },
  { kind: '复核频率', title: '防护用具复核', field: 'ppe_review', safety: false },
])
const candidate = ref<DraftRule>({ kind: '禁止组合', title: '', field: '', safety: false })
const safetyCount = computed(() => rules.value.filter(rule => rule.safety).length)
function appendRule() {
  if (!candidate.value.title || !candidate.value.field) return
  rules.value.push({ ...candidate.value })
  candidate.value = { kind: '禁止组合', title: '', field: '', safety: false }
}
</script>

<template>
  <div class="grid-two">
    <section class="panel workspace-card">
      <h2>V4 草稿规则集 <span class="badge">{{ safetyCount }} 条强制安全规则</span></h2>
      <div v-for="(rule, index) in rules" :key="`${rule.field}-${index}`" class="rule-row">
        <small>{{ rule.kind }} · {{ rule.field }}</small>
        <strong>{{ rule.title }}</strong>
        <p v-if="rule.safety" class="warning">强制安全：例外流程不可绕过</p>
      </div>
    </section>
    <form class="panel workspace-card form-grid" @submit.prevent="appendRule">
      <h2 class="wide">添加可配置规则</h2>
      <label>规则类型<select v-model="candidate.kind"><option>必填检查</option><option>有效期</option><option>禁止组合</option><option>复核频率</option></select></label>
      <label>字段键<input v-model.trim="candidate.field" placeholder="inspection_guard" /></label>
      <label class="wide">规则说明<input v-model.trim="candidate.title" placeholder="现场可理解的校验说明" /></label>
      <label class="wide"><span><input v-model="candidate.safety" type="checkbox" /> 标记为不可豁免的强制安全规则</span></label>
      <button class="button wide" type="submit">加入草稿</button>
    </form>
  </div>
</template>
