import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import Home from '@/views/Home.vue'
import Editor from '@/views/Editor.vue'
import Settings from '@/views/Settings.vue'
import Login from '@/views/Login.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/login', name: 'login', component: Login },
    { path: '/', name: 'home', component: Home },
    { path: '/editor/:id', name: 'editor', component: Editor, props: true },
    { path: '/settings', name: 'settings', component: Settings },
  ],
})

router.beforeEach((to) => {
  const publicPaths = ['/login']
  if (publicPaths.includes(to.path)) return true
  const store = useAuthStore()
  store.loadFromStorage()
  if (!store.token) return '/login'
  return true
})

export default router
