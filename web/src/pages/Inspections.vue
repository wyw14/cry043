<script setup lang="ts">
import { computed, ref } from 'vue'
const checks = ref([
  { id: 'permit', label: '许可编号与现场作业一致', passed: false },
  { id: 'guard', label: '强制防护装置已锁定', passed: false },
  { id: 'briefing', label: '班前安全交底完成', passed: false },
])
const photos = ref<string[]>([])
const completion = computed(() => checks.value.filter(item => item.passed).length)
function attachEvidence() { photos.value.push(`local-photo-${photos.value.length + 1}.jpg`) }
</script>
<template>
  <div class="grid-two">
    <section class="panel workspace-card"><h2>焊装区 · 动火抽查表</h2><ul class="checklist"><li v-for="check in checks" :key="check.id"><label><input v-model="check.passed" type="checkbox" /> {{ check.label }}</label></li></ul><p>已完成 {{ completion }} / {{ checks.length }} 项</p></section>
    <section class="panel workspace-card"><h2>受控本地证据</h2><p>仅接受白名单图片类型，文件正文不写入日志。</p><button class="button secondary" @click="attachEvidence">模拟添加照片</button><div v-for="photo in photos" :key="photo" class="rule-row">{{ photo }}</div><p v-if="!photos.length" class="warning">提交抽查前至少需要一张照片证据</p></section>
  </div>
</template>
