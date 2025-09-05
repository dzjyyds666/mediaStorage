import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '@/views/Dashboard.vue'

const router = createRouter({
  history: createWebHistory('/'),
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: Dashboard
    },
    {
      path: '/storage',
      name: 'storage',
      component: () => import('@/views/Storage.vue')
    },
    {
      path: '/monitoring',
      name: 'monitoring',
      component: () => import('@/views/Monitoring.vue')
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/views/Settings.vue')
    }
  ]
})

export default router