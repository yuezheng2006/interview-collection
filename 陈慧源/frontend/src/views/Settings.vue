<script setup lang="ts">
import { ref } from 'vue'
import { t } from '@/i18n'

type ModelProvider = 'tongyi' | 'wenxin' | 'zhipu'

const currentProvider = ref<ModelProvider>('zhipu')
const apiKey = ref<string>('')

function saveSettings() {
  // TODO: persist settings to localStorage or backend
  window.localStorage.setItem('ai-settings', JSON.stringify({
    provider: currentProvider.value,
    apiKey: apiKey.value,
  }))
}
</script>

<template>
  <div style="padding: 16px">
    <el-page-header :content="t('settings')" @back="$router.back()" />
    <el-card style="margin-top: 16px; max-width: 640px">
      <el-form label-width="160px">
        <el-form-item :label="t('aiProvider')">
          <el-select v-model="currentProvider" placeholder="Select">
            <el-option :label="t('tongyi')" value="tongyi" />
            <el-option :label="t('wenxin')" value="wenxin" />
            <el-option :label="t('zhipu')" value="zhipu" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('apiKey')">
          <el-input v-model="apiKey" type="password" :placeholder="t('inputApiKey')" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="saveSettings">{{ t('save') }}</el-button>
          <el-button @click="$router.push('/')">{{ t('backHome') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
</style>

