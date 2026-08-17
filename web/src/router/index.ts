import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'control-room', component: () => import('../pages/Library.vue'), meta: { title: '区域与工序风险驾驶舱' } },
  { path: '/rules', name: 'rule-studio', component: () => import('../pages/RuleEditor.vue'), meta: { title: '规范规则工作室' } },
  { path: '/impact', name: 'activation-gate', component: () => import('../pages/Impact.vue'), meta: { title: '影响模拟与生效闸门' } },
  { path: '/confirm', name: 'acknowledgements', component: () => import('../pages/Confirmations.vue'), meta: { title: '班组与个人确认中心' } },
  { path: '/inspections', name: 'field-assurance', component: () => import('../pages/Inspections.vue'), meta: { title: '现场抽查执行台' } },
  { path: '/remediations', name: 'corrective-actions', component: () => import('../pages/Remediations.vue'), meta: { title: '不符合项整改与复核' } },
  { path: '/compare', name: 'version-lineage', component: () => import('../pages/Compare.vue'), meta: { title: '规范版本谱系与回滚' } },
  { path: '/audit', name: 'audit-export', component: () => import('../pages/Audit.vue'), meta: { title: '变更审计与受控导出' } },
]

export default createRouter({ history: createWebHistory(), routes, scrollBehavior: () => ({ top: 0 }) })
