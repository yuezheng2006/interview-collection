<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useDocumentStore, type DocumentItem } from '@/stores/document'
import { ElMessageBox, ElMessage } from 'element-plus'
import { t } from '@/i18n'
import { 
  Plus, 
  Edit, 
  Delete, 
  Document, 
  Calendar, 
  DocumentCopy,
  Setting,
  Search
} from '@element-plus/icons-vue'

const store = useDocumentStore()
const newTitle = ref<string>('')
const searchQuery = ref<string>('')

onMounted(() => {
  store.fetchDocuments()
})

async function handleCreate() {
  if (!newTitle.value.trim()) {
    ElMessage.warning(t('pleaseInputTitle'))
    return
  }
  await store.createDocument(newTitle.value.trim())
  newTitle.value = ''
}

async function handleRename(row: DocumentItem) {
  const title = await ElMessageBox.prompt(t('pleaseInputTitle'), t('rename'), {
    confirmButtonText: t('confirm'),
    cancelButtonText: t('cancel'),
    inputValue: row.title,
  }).then((res) => res.value).catch(() => null)
  if (!title) return
  await store.renameDocument(row.id, title)
}

async function handleDelete(row: DocumentItem) {
  const message = t('confirmDeleteMessage').replace('{title}', row.title)
  const ok = await ElMessageBox.confirm(message, t('confirmDelete'), {
    confirmButtonText: t('delete'),
    cancelButtonText: t('cancel'),
    type: 'warning',
  }).then(() => true).catch(() => false)
  if (!ok) return
  await store.deleteDocument(row.id)
}

function openEditor(row: DocumentItem) {
  window.location.href = `/editor/${row.id}`
}

// 过滤文档列表
const filteredDocuments = computed(() => {
  if (!searchQuery.value) return store.documents
  return store.documents.filter(doc => 
    doc.title.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

// 格式化日期
function formatDate(dateString: string) {
  const date = new Date(dateString)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<template>
  <div class="home-container">
    <!-- 顶部渐变背景区域 -->
    <div class="hero-section">
      <!-- <div class="hero-content"> -->
        <!-- <h1 class="hero-title"> -->
          <!-- <Document :size="10" class="hero-icon" /> -->
          <!-- {{ t('documents') }} -->
        <!-- </h1> -->
        <p class="hero-subtitle">AI驱动的智能写作助手，让创作更高效</p>
      <!-- </div> -->
    </div>

    <!-- 主要内容区域 -->
    <div class="main-content">
      <!-- 创建新文档卡片 -->
      <div class="create-card">
        <div class="card-header">
          <h3 class="card-title">
            <!-- <Plus class="card-icon" /> -->
            创建新文档
          </h3>
        </div>
        <div class="card-body">
          <!-- <div class="input-group"> -->
            <el-input 
              v-model="newTitle" 
              :placeholder="t('newDocumentTitle')" 
              class="title-input"
              size="large"
            />
            <el-button 
              type="primary" 
              @click="handleCreate"
              size="large"
              class="create-btn"
            >
              <Plus class="btn-icon" />
              创建文档
            </el-button>
          <!-- </div> -->
        </div>
      </div>

      <!-- 搜索和统计区域 -->
      <div class="stats-section">
        <div class="search-box">
          <el-input
            v-model="searchQuery"
            :placeholder="'搜索文档...'"
            class="search-input"
            size="large"
          >
            <template #prefix>
              <Search class="search-icon" />
            </template>
          </el-input>
        </div>
        <div class="stats-cards">
          <div class="stat-card">
            <!-- <div class="stat-icon">
              <Document />
            </div> -->
            <!-- <div class="stat-content"> -->
              <div class="stat-number">{{ store.documents.length }}</div>
              <div class="stat-label">总文档数</div>
            </div>
          <!-- </div> -->
          <div class="stat-card">
            <!-- <div class="stat-icon"> -->
              <!-- <DocumentCopy /> -->
            <!-- </div> -->
            <!-- <div class="stat-content"> -->
              <div class="stat-number">
                {{ store.documents.reduce((sum, doc) => sum + (doc.wordCount || 0), 0) }}
              </div>
              <div class="stat-label">总字数</div>
            <!-- </div> -->
          </div>
        </div>
      </div>

      <!-- 文档列表 -->
      <div class="documents-section">
        <div class="section-header">
          <h3 class="section-title">我的文档</h3>
          <!-- <el-button 
            @click="$router.push('/settings')" 
            class="settings-btn"
            size="small"
          >
            <Setting class="btn-icon" />
            设置
          </el-button> -->
        </div>
        
        <div class="documents-grid">
          <div 
            v-for="doc in filteredDocuments" 
            :key="doc.id" 
            class="document-card"
            @click="openEditor(doc)"
          >
            <div class="doc-header">
              <!-- <Document class="doc-icon" /> -->
              <div class="doc-title">{{ doc.title }}</div>
            </div>
            <div class="doc-meta">
              <div class="meta-item">
                <!-- <Calendar class="meta-icon" /> -->
                <span>{{ formatDate(doc.updatedAt) }}</span>
              </div>
              <div class="meta-item">
                <!-- <DocumentCopy class="meta-icon" /> -->
                <span>{{ doc.wordCount || 0 }} 字</span>
              </div>
            </div>
            <div class="doc-actions">
              <el-button 
                size="small" 
                type="primary" 
                @click.stop="openEditor(doc)"
                class="action-btn"
              >
                <Edit class="btn-icon" />
                编辑
              </el-button>
              <el-button 
                size="small" 
                @click.stop="handleRename(doc)"
                class="action-btn"
              >
                重命名
              </el-button>
              <el-button 
                size="small" 
                type="danger" 
                @click.stop="handleDelete(doc)"
                class="action-btn"
              >
                <Delete class="btn-icon" />
                删除
              </el-button>
            </div>
          </div>
        </div>

        <!-- 空状态 -->
        <div v-if="filteredDocuments.length === 0" class="empty-state">
          <!-- <Document class="empty-icon" /> -->
          <h3 class="empty-title">暂无文档</h3>
          <p class="empty-subtitle">创建您的第一个文档开始写作吧！</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.home-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.hero-section {
  padding: 10px 10px 30px 10px;
  text-align: center;
  color: white;
}

.hero-content {
  max-width: 800px;
  margin: 0 auto;
}

.hero-title {
  font-size: 2.5rem;
  font-weight: 700;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
}

.hero-icon {
  font-size: 2.5rem;
  color: #ffd700;
}

.hero-subtitle {
  font-size: 1.2rem;
  opacity: 0.9;
  margin: 0;
}

.main-content {
  background: #f8fafc;
  border-radius: 30px 30px 0 0;
  padding: 30px;
  margin-top: -15px;
  position: relative;
  z-index: 1;
}

.create-card {
  background: white;
  border-radius: 16px;
  padding: 24px;
  margin-bottom: 24px;
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.08);
  border: 1px solid #e2e8f0;
}

.card-header {
  margin-bottom: 20px;
}

.card-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 12px;
}

.card-icon {
  color: #3b82f6;
  font-size: 1.5rem;
}

.card-body {
  display: flex;
  gap: 16px;
  align-items: center;
}

.title-input {
  flex: 1;
  max-width: 500px;
}

.create-btn {
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  border: none;
  padding: 12px 24px;
  border-radius: 12px;
  font-weight: 600;
  transition: all 0.3s ease;
}

.create-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 25px rgba(59, 130, 246, 0.3);
}

.btn-icon {
  margin-right: 8px;
}

.stats-section {
  display: flex;
  gap: 24px;
  margin-bottom: 24px;
  align-items: center;
}

.search-box {
  flex: 1;
  max-width: 400px;
}

.search-input {
  border-radius: 12px;
}

.search-icon {
  color: #64748b;
}

.stats-cards {
  display: flex;
  gap: 20px;
}

.stat-card {
  background: white;
  padding: 16px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  box-shadow: 0 3px 15px rgba(0, 0, 0, 0.06);
  border: 1px solid #e2e8f0;
  min-width: 140px;
}

.stat-icon {
  font-size: 1.5rem;
  color: #3b82f6;
  background: #eff6ff;
  padding: 10px;
  border-radius: 10px;
}

.stat-content {
  text-align: center;
}

.stat-number {
  font-size: 1.8rem;
  font-weight: 700;
  color: #1e293b;
  line-height: 1;
}

.stat-label {
  font-size: 0.9rem;
  color: #64748b;
  margin-top: 4px;
}

.documents-section {
  background: white;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.08);
  border: 1px solid #e2e8f0;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.section-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.settings-btn {
  border-radius: 8px;
  color: #64748b;
}

.documents-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.document-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.document-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, #3b82f6, #8b5cf6);
  transform: scaleX(0);
  transition: transform 0.3s ease;
}

.document-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
  border-color: #3b82f6;
}

.document-card:hover::before {
  transform: scaleX(1);
}

.doc-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.doc-icon {
  font-size: 1.5rem;
  color: #3b82f6;
}

.doc-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: #1e293b;
  flex: 1;
}

.doc-meta {
  margin-bottom: 20px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  color: #64748b;
  font-size: 0.9rem;
}

.meta-icon {
  font-size: 1rem;
  color: #94a3b8;
}

.doc-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.action-btn {
  border-radius: 8px;
  font-size: 0.85rem;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #64748b;
}

.empty-icon {
  font-size: 3rem;
  color: #cbd5e1;
  margin-bottom: 20px;
}

.empty-title {
  font-size: 1.5rem;
  font-weight: 600;
  margin-bottom: 12px;
  color: #475569;
}

.empty-subtitle {
  font-size: 1rem;
  margin: 0;
  opacity: 0.8;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .hero-section {
    padding: 40px 20px;
  }
  
  .hero-title {
    font-size: 2rem;
    flex-direction: column;
    gap: 12px;
  }
  
  .main-content {
    padding: 20px;
    border-radius: 30px 30px 0 0;
  }
  
  .card-body {
    flex-direction: column;
    align-items: stretch;
  }
  
  .title-input {
    max-width: none;
  }
  
  .stats-section {
    flex-direction: column;
    align-items: stretch;
  }
  
  .search-box {
    max-width: none;
  }
  
  .stats-cards {
    justify-content: center;
  }
  
  .documents-grid {
    grid-template-columns: 1fr;
  }
  
  .section-header {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }
}

/* 动画效果 */
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.document-card {
  animation: fadeInUp 0.6s ease forwards;
}

.document-card:nth-child(1) { animation-delay: 0.1s; }
.document-card:nth-child(2) { animation-delay: 0.2s; }
.document-card:nth-child(3) { animation-delay: 0.3s; }
.document-card:nth-child(4) { animation-delay: 0.4s; }
.document-card:nth-child(5) { animation-delay: 0.5s; }
.document-card:nth-child(6) { animation-delay: 0.6s; }
</style>

