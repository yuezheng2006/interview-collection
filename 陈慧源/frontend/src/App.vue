<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter, useRoute } from 'vue-router'
import { t, setLanguage, currentLanguage, initLanguage } from '@/i18n'
import { onMounted } from 'vue'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

onMounted(() => {
  initLanguage()
})

function logout() {
  auth.logout()
  router.replace('/login')
}

function switchLanguage() {
  const newLang = currentLanguage.value === 'zh-CN' ? 'en-US' : 'zh-CN'
  setLanguage(newLang)
}
</script>

<template>
  <div>
    <div v-if="route.path !== '/login'" style="padding: 8px 16px; border-bottom: 1px solid #eee; display:flex; justify-content:space-between; align-items:center">
      <strong>{{ t('appTitle') }}</strong>
      <div style="display: flex; gap: 12px; align-items: center">
        <el-button size="small" @click="switchLanguage">
          {{ currentLanguage === 'zh-CN' ? 'EN' : '中文' }}
        </el-button>
        <el-button text v-if="!auth.token" @click="router.push('/login')">{{ t('login') }}</el-button>
        <el-dropdown v-else>
          <span class="el-dropdown-link">
            {{ auth.user?.username || t('user') }}
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="logout">{{ t('logout') }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>
    <router-view />
  </div>
</template>

<style scoped></style>
