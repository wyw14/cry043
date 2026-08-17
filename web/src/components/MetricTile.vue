<script setup lang="ts">
defineProps<{ label: string; value: string | number; tone?: 'ink' | 'amber' | 'red' }>()
</script>
<template><article class="metric" :data-tone="tone ?? 'ink'"><span>{{ label }}</span><strong>{{ value }}</strong><slot /></article></template>
