import axios, { type AxiosInstance, type AxiosResponse, type AxiosError } from 'axios'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'

const BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

let httpClient: AxiosInstance | null = null

export function getHttpClient(): AxiosInstance {
  if (httpClient) return httpClient
  httpClient = axios.create({
    baseURL: BASE_URL,
    timeout: 20000,
  })
  httpClient.interceptors.request.use((config) => {
    try {
      const store = useAuthStore()
      if (store.token) {
        config.headers = config.headers || {}
        ;(config.headers as Record<string, string>)['Authorization'] = store.token
      }
    } catch (e) {
      // noop: store not ready during app bootstrap
    }
    return config
  })
  httpClient.interceptors.response.use(
    (res: AxiosResponse) => res,
    (err: AxiosError) => {
      // 统一错误处理
      let errorMessage = '请求失败'
      
      if (err.response) {
        // 服务器响应了错误状态码
        const status = err.response.status
        const data = err.response.data as any
        
        switch (status) {
          case 400:
            errorMessage = data.message || '请求参数错误'
            break
          case 401:
            errorMessage = data.message || '未授权，请重新登录'
            break
          case 403:
            errorMessage = data.message || '禁止访问'
            break
          case 404:
            errorMessage = data.message || '请求的资源不存在'
            break
          case 422:
            errorMessage = data.message || '数据验证失败'
            break
          case 500:
            errorMessage = data.message || '服务器内部错误'
            break
          default:
            errorMessage = data.message || `请求失败 (${status})`
        }
      } else if (err.request) {
        // 请求已发出但没有收到响应
        errorMessage = '网络连接失败，请检查网络设置'
      } else {
        // 请求配置出错
        errorMessage = err.message || '请求配置错误'
      }
      
      // 显示错误通知
      ElMessage.error(errorMessage)
      
      return Promise.reject(err)
    }
  )
  return httpClient
}

