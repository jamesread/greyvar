import { createRouter, createWebHistory } from 'vue-router'
import DefinitionsView from '../views/DefinitionsView.vue'
import { loadInit } from '../composables/useInit'

const routes = [
  { path: '/', redirect: '/entdefs' },
  { path: '/entdefs', name: 'entdefs', component: DefinitionsView, props: { kind: 'entdefs' } },
  { path: '/tiledefs', name: 'tiledefs', component: DefinitionsView, props: { kind: 'tiledefs' } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async () => {
  await loadInit()
  return true
})

export default router
