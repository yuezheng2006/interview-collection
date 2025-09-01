<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { t } from '@/i18n'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'

const router = useRouter()
const store = useAuthStore()
const username = ref('')
const password = ref('')
const isLogin = ref(true)
const loading = ref(false)

async function submit() {
  if (!username.value || !password.value) {
    ElMessage.warning('请输入用户名和密码')
    return
  }

  loading.value = true
  try {
    if (isLogin.value) {
      await store.performLogin(username.value, password.value)
      ElMessage.success('登录成功')
    } else {
      await store.performRegister(username.value, password.value)
      ElMessage.success('注册成功')
    }
    router.replace('/')
  } catch (error: unknown) {
    // 错误已经在API拦截器中处理，这里只需要处理其他类型的错误
    console.error('操作失败:', error)
    if (error && typeof error === 'object' && 'response' in error) {
      // 有response属性的错误已经在拦截器中处理
    } else {
      ElMessage.error('操作失败，请稍后重试')
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-container">
    <!-- 背景装饰 -->
    <div class="background-decoration">
      <div class="floating-shape shape-1"></div>
      <div class="floating-shape shape-2"></div>
      <div class="floating-shape shape-3"></div>
      <div class="floating-shape shape-4"></div>
    </div>

    <!-- 主要内容 -->
    <div class="main-content">
      <!-- Logo区域 -->
      <div class="logo-section">
        <div class="logo-container">
          <!-- <div class="logo-icon">
            <svg viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg">
            
              <defs>
                <linearGradient id="gradient1" x1="0%" y1="0%" x2="100%" y2="100%">
                  <stop offset="0%" style="stop-color:#667eea"/>
                  <stop offset="100%" style="stop-color:#764ba2"/>
                </linearGradient>
              </defs>
            </svg>
          </div> -->
          <h1 class="product-name">TOC Writer</h1>
          <p class="product-tagline">AI驱动的智能写作助手</p>
        </div>
      </div>

      <!-- 登录表单 -->
      <div class="form-container">
        <div class="form-card">
          <div class="form-header">
            <h2 class="form-title">{{ isLogin ? '欢迎回来' : '开始创作之旅' }}</h2>
            <p class="form-subtitle">{{ isLogin ? '登录您的账户继续创作' : '注册新账户开始使用AI写作' }}</p>
          </div>

          <el-form class="login-form" label-position="top">
            <el-form-item>
              <label class="form-label">{{ t('username') }}</label>
              <el-input 
                v-model="username" 
                class="custom-input"
                placeholder="请输入用户名"
                :prefix-icon="User"
                size="large"
              />
            </el-form-item>
            
            <el-form-item>
              <label class="form-label">{{ t('password') }}</label>
              <el-input 
                v-model="password" 
                type="password" 
                class="custom-input"
                placeholder="请输入密码"
                :prefix-icon="Lock"
                size="large"
                show-password
              />
            </el-form-item>

            <el-form-item>
              <el-button 
                type="primary" 
                :loading="loading" 
                @click="submit"
                class="submit-btn"
                size="large"
              >
                {{ loading ? '处理中...' : (isLogin ? t('submit') : '立即注册') }}
              </el-button>
            </el-form-item>
          </el-form>

          <div class="form-footer">
            <el-link 
              type="primary" 
              @click="isLogin = !isLogin"
              class="switch-link"
            >
              {{ isLogin ? '还没有账户？立即注册' : '已有账户？立即登录' }}
            </el-link>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部信息 -->
    <div class="footer-info">
      <p>&copy; 2024 TOC Writer. 让AI助力您的创作之旅</p>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  width: 100vw;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px;
  margin: 0;
  box-sizing: border-box;
}

/* 背景装饰 */
.background-decoration {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 1;
}

.floating-shape {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  animation: float 6s ease-in-out infinite;
}

.shape-1 {
  width: 80px;
  height: 80px;
  top: 10%;
  left: 10%;
  animation-delay: 0s;
}

.shape-2 {
  width: 120px;
  height: 120px;
  top: 20%;
  right: 15%;
  animation-delay: 2s;
}

.shape-3 {
  width: 60px;
  height: 60px;
  bottom: 20%;
  left: 20%;
  animation-delay: 4s;
}

.shape-4 {
  width: 100px;
  height: 100px;
  bottom: 15%;
  right: 10%;
  animation-delay: 1s;
}

@keyframes float {
  0%, 100% { transform: translateY(0px) rotate(0deg); }
  50% { transform: translateY(-20px) rotate(180deg); }
}

/* 主要内容 */
.main-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 40px;
  z-index: 2;
  position: relative;
  max-width: 1200px;
  width: 100%;
}

/* Logo区域 */
.logo-section {
  text-align: center;
  color: white;
}

.logo-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.logo-icon {
  width: 80px;
  height: 80px;
  filter: drop-shadow(0 8px 16px rgba(0, 0, 0, 0.3));
}

.product-name {
  font-size: 2.5rem;
  font-weight: 700;
  margin: 0;
  text-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
  letter-spacing: 1px;
}

.product-tagline {
  font-size: 1.1rem;
  margin: 0;
  opacity: 0.9;
  font-weight: 300;
}

/* 表单容器 */
.form-container {
  width: 100%;
  max-width: 400px;
}

.form-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 24px;
  padding: 40px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.form-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.15);
}

.form-header {
  text-align: center;
  margin-bottom: 32px;
}

.form-title {
  font-size: 1.8rem;
  font-weight: 600;
  color: #2c3e50;
  margin: 0 0 8px 0;
}

.form-subtitle {
  font-size: 1rem;
  color: #7f8c8d;
  margin: 0;
  line-height: 1.5;
}

/* 表单样式 */
.login-form {
  margin-bottom: 24px;
}

.form-label {
  display: block;
  font-weight: 500;
  color: #2c3e50;
  margin-bottom: 8px;
  font-size: 0.9rem;
}

.custom-input {
  --el-input-border-radius: 12px;
  --el-input-border-color: #e1e8ed;
  --el-input-hover-border-color: #667eea;
  --el-input-focus-border-color: #667eea;
}

.custom-input :deep(.el-input__wrapper) {
  background: #f8f9fa;
  border: 2px solid transparent;
  transition: all 0.3s ease;
  padding: 12px 16px;
}

.custom-input :deep(.el-input__wrapper:hover) {
  background: #ffffff;
  border-color: #667eea;
}

.custom-input :deep(.el-input__wrapper.is-focus) {
  background: #ffffff;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.custom-input :deep(.el-input__inner) {
  font-size: 1rem;
  color: #2c3e50;
}

.custom-input :deep(.el-input__inner::placeholder) {
  color: #bdc3c7;
}

/* 提交按钮 */
.submit-btn {
  width: 100%;
  height: 48px;
  border-radius: 12px;
  font-size: 1.1rem;
  font-weight: 600;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  transition: all 0.3s ease;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.submit-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.4);
}

.submit-btn:active {
  transform: translateY(0);
}

/* 表单底部 */
.form-footer {
  text-align: center;
  padding-top: 24px;
  border-top: 1px solid #ecf0f1;
}

.switch-link {
  font-size: 0.95rem;
  text-decoration: none;
  transition: color 0.3s ease;
}

.switch-link:hover {
  color: #764ba2 !important;
}

/* 底部信息 */
.footer-info {
  position: absolute;
  bottom: 20px;
  text-align: center;
  color: rgba(255, 255, 255, 0.7);
  font-size: 0.9rem;
  z-index: 2;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .main-content {
    gap: 30px;
  }
  
  .product-name {
    font-size: 2rem;
  }
  
  .form-card {
    padding: 30px 24px;
    margin: 0 16px;
  }
  
  .form-title {
    font-size: 1.5rem;
  }
  
  .logo-icon {
    width: 60px;
    height: 60px;
  }
}

@media (max-width: 480px) {
  .login-container {
    padding: 16px;
  }
  
  .form-card {
    padding: 24px 20px;
  }
  
  .product-name {
    font-size: 1.8rem;
  }
  
  .product-tagline {
    font-size: 1rem;
  }
}

/* 加载动画 */
.el-loading-mask {
  border-radius: 12px;
}

/* 输入框图标样式 */
.custom-input :deep(.el-input__prefix) {
  color: #667eea;
}
</style>
