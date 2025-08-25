import { defineStore } from 'pinia'
import { login, register } from '@/services/auth'

export type UserProfile = {
  username: string
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: '' as string,
    user: null as UserProfile | null,
  }),
  actions: {
    async performLogin(username: string, password: string) {
      const res = await login(username, password)
      this.token = res.token
      this.user = { username }
      window.localStorage.setItem('auth-token', this.token)
    },
    async performRegister(username: string, password: string) {
      await register(username, password)
      await this.performLogin(username, password)
    },
    loadFromStorage() {
      const t = window.localStorage.getItem('auth-token')
      if (t) this.token = t
    },
    logout() {
      this.token = ''
      this.user = null
      window.localStorage.removeItem('auth-token')
    }
  }
})

