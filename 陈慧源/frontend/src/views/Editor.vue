<script setup lang="ts">
// 组件名称
defineOptions({
  name: 'DocumentEditor'
})

import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDocumentStore } from '@/stores/document'
import { t } from '@/i18n'
import { 
  callUnifiedAIStream,
  chat,
  getSmartTextSelection, 
  generateDocumentSummary,
  getAvailableModels,
  switchModel,
  generateSessionID,
  type UnifiedAiRequest,
  type AiModelInfo,
  type StreamCallback
} from '@/services/ai'
import type { AiResult, TextSelection, AiFunctionType } from '@/types'
import { ElMessage, ElDialog } from 'element-plus'

const route = useRoute()
const router = useRouter()
const store = useDocumentStore()

// 基础状态
const content = ref<string>('')
const lastSavedAt = ref<Date | null>(null)
const isSaving = ref<boolean>(false)

// AI功能状态
const isAiProcessing = ref<boolean>(false)
const selectedText = ref<string>('')
const selectionStart = ref<number>(0)
const selectionEnd = ref<number>(0)
const cursorPosition = ref<number>(0)
const aiResults = ref<AiResult[]>([])

// AI模型管理状态
const availableModels = ref<AiModelInfo[]>([])
const currentModel = ref<string>('mock')
const isModelLoading = ref<boolean>(false)

// AI弹窗状态
const showAiDialog = ref<boolean>(false)
const showGenerateDialog = ref<boolean>(false)
const selectedAiFunction = ref<AiFunctionType>('polish')
const userRequirement = ref<string>('')
const currentAiResult = ref<AiResult | null>(null)
const isDialogJustClosed = ref<boolean>(false) // 新增：防止弹窗重复打开

// 生成内容相关状态
const generateRequirement = ref<string>('')
const generateType = ref<'new' | 'related'>('new')

// 新增：多轮对话状态
const sessionID = ref<string>('')
const showChatDialog = ref<boolean>(false)
const chatMessage = ref<string>('')
const chatHistory = ref<Array<{role: 'user' | 'assistant', content: string}>>([])
const isChatProcessing = ref<boolean>(false)

// 新增：流式AI状态
const isStreaming = ref<boolean>(false)
const streamedContent = ref<string>('')
const isStreamingInDialog = ref<boolean>(false) // 新增：弹窗中的流式状态

// 新增：总结全文章弹窗状态
const showSummaryDialog = ref<boolean>(false)
const isSummaryProcessing = ref<boolean>(false)
const summaryStreamedContent = ref<string>('')
const summaryResult = ref<string>('')

// 字体控制状态
const showPreview = ref<boolean>(false)

// 计算属性
const wordCount = computed<number>(() => content.value.trim().length)
const hasSelection = computed<boolean>(() => selectedText.value.length > 0)
const canUseAi = computed<boolean>(() => content.value.length > 0)

// 格式化预览内容
const formatPreviewContent = computed<string>(() => {
  if (!content.value) return '暂无内容'
  
  // 直接返回HTML内容，让v-html渲染真正的格式化效果
  return content.value
})

// 自动保存
let autoSaveTimer: number | undefined

function startAutoSave() {
  stopAutoSave()
  autoSaveTimer = window.setInterval(async () => {
    await saveDocument()
  }, 5000)
}

function stopAutoSave() {
  if (autoSaveTimer) {
    clearInterval(autoSaveTimer)
    autoSaveTimer = undefined
  }
}

// 初始化
async function init() {
  const id = String(route.params.id || '')
  if (!id) {
    router.replace('/')
    return
  }
  if (!store.documents.length) {
    await store.fetchDocuments()
  }
  const doc = store.documents.find(d => d.id === id)
  if (!doc) {
    router.replace('/')
    return
  }
  store.currentDocument = doc
  content.value = doc.content || ''
  
  // 初始化AI模型
  await initAiModels()
  
  // 生成会话ID
  sessionID.value = generateSessionID()
}

// 初始化AI模型
async function initAiModels() {
  try {
    isModelLoading.value = true
    const response = await getAvailableModels()
    availableModels.value = response.models
    currentModel.value = response.current
  } catch (error) {
    console.error('获取AI模型失败:', error)
    ElMessage.error('获取AI模型失败')
  } finally {
    isModelLoading.value = false
  }
}

// 切换AI模型
async function handleSwitchModel(modelName: string) {
  try {
    isModelLoading.value = true
    const response = await switchModel(modelName)
    if (response.success) {
      currentModel.value = response.modelName
      ElMessage.success(`已切换到${response.modelName}模型`)
    } else {
      ElMessage.error(response.message || '模型切换失败')
    }
  } catch (error) {
    console.error('切换模型失败:', error)
    ElMessage.error('模型切换失败')
  } finally {
    isModelLoading.value = false
  }
}

// 保存文档
async function saveDocument() {
  if (isSaving.value || !store.currentDocument) return
  try {
    isSaving.value = true
    await store.saveDocumentContent(content.value)
    lastSavedAt.value = new Date()
  } finally {
    isSaving.value = false
  }
}

// 文本选择处理
function handleTextSelection() {
  // 如果弹窗刚关闭，暂时不处理选择事件
  if (isDialogJustClosed.value) {
    return
  }
  
  const textarea = document.querySelector('textarea') as HTMLTextAreaElement
  if (!textarea) return
  
  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  
  // 如果选择范围没有变化，不处理
  if (start === selectionStart.value && end === selectionEnd.value) {
    return
  }
  
  selectionStart.value = start
  selectionEnd.value = end
  selectedText.value = content.value.substring(start, end)
  cursorPosition.value = end
  
  // 不再自动打开弹窗，只更新选择状态
  // 用户可以通过上方的AI功能按钮来使用AI功能
}

// 防抖的文本选择处理
let selectionDebounceTimer: number | undefined
function handleTextSelectionDebounced() {
  if (selectionDebounceTimer) {
    clearTimeout(selectionDebounceTimer)
  }
  selectionDebounceTimer = window.setTimeout(() => {
    handleTextSelection()
  }, 100)
}

// 智能内容选择
function getSmartSelection(): TextSelection | null {
  if (!hasSelection.value) return null
  
  return getSmartTextSelection(
    content.value,
    cursorPosition.value,
    selectionStart.value,
    selectionEnd.value
  )
}

// 调用AI（支持流式和非流式）
async function callAI() {
  if (!canUseAi.value || isAiProcessing.value) return
  
  const smartSelection = getSmartSelection()
  if (!smartSelection && selectedAiFunction.value !== 'generate') {
    ElMessage.warning('请先选择要处理的文本')
    return
  }
  
  try {
    isAiProcessing.value = true
    isStreaming.value = true
    isStreamingInDialog.value = true // 新增：设置弹窗流式状态
    streamedContent.value = ''
    
    const requestData: UnifiedAiRequest = {
      functionType: selectedAiFunction.value,
      documentSummary: generateDocumentSummary(content.value),
      userRequirement: userRequirement.value || getDefaultRequirement(selectedAiFunction.value),
      selectedText: smartSelection?.text || '',
      contextText: smartSelection ? `${smartSelection.contextBefore}${smartSelection.text}${smartSelection.contextAfter}` : '',
      cursorPosition: cursorPosition.value,
      modelName: currentModel.value, // 使用当前选择的模型
      sessionID: sessionID.value,    // 添加会话ID
      stream: true                   // 启用流式返回
    }
    
    // 使用流式AI接口
    await callUnifiedAIStream(requestData, handleStreamCallback)
    
    // 流式完成后，创建AI结果
    const aiResult: AiResult = {
      id: Date.now().toString(),
      originalText: smartSelection?.text || '',
      aiGeneratedText: streamedContent.value,
      functionType: selectedAiFunction.value,
      modelName: currentModel.value, // 记录使用的模型
      timestamp: new Date(),
      isApplied: false
    }
    
    aiResults.value.unshift(aiResult)
    currentAiResult.value = aiResult
    
    ElMessage.success('AI处理完成')
    
  } catch (error) {
    console.error('AI调用失败:', error)
    ElMessage.error('AI处理失败，请重试')
  } finally {
    isAiProcessing.value = false
    isStreaming.value = false
    isStreamingInDialog.value = false // 新增：重置弹窗流式状态
  }
}

// 处理流式回调
const handleStreamCallback: StreamCallback = (chunk: string, isComplete: boolean) => {
  if (isComplete) {
    // 流式完成
    isStreaming.value = false
    isStreamingInDialog.value = false // 新增：重置弹窗流式状态
  } else {
    // 接收流式内容
    try {
      // 尝试解析JSON，如果是错误信息则抛出
      const parsed = JSON.parse(chunk)
      if (parsed.error) {
        throw new Error(parsed.error)
      }
      // 如果是正常内容，直接添加到流式内容中
      if (parsed.content) {
        streamedContent.value += parsed.content
      } else {
        streamedContent.value += chunk
      }
    } catch {
      // 如果不是JSON，直接作为文本内容处理
      streamedContent.value += chunk
    }
  }
}



// 生成内容（支持流式）
async function generateContent() {
  if (!generateRequirement.value.trim()) {
    ElMessage.warning('请输入生成要求')
    return
  }
  
  if (isAiProcessing.value) return
  
  try {
    isAiProcessing.value = true
    isStreaming.value = true
    isStreamingInDialog.value = true // 新增：设置弹窗流式状态
    streamedContent.value = ''
    
    let selectedText = ''
    let contextText = ''
    
    if (generateType.value === 'new') {
      // 生成新内容
      selectedText = ''
      contextText = ''
    } else {
      // 根据文章内容生成相关内容
      selectedText = content.value
      contextText = content.value
    }
    
    const requestData: UnifiedAiRequest = {
      functionType: 'generate',
      documentSummary: generateDocumentSummary(content.value),
      userRequirement: generateRequirement.value,
      selectedText: selectedText,
      contextText: contextText,
      cursorPosition: cursorPosition.value,
      sessionID: sessionID.value,    // 添加会话ID
      stream: true                   // 启用流式返回
    }
    
    // 使用流式AI接口
    await callUnifiedAIStream(requestData, handleStreamCallback)
    
    // 流式完成后，创建AI结果
    const aiResult: AiResult = {
      id: Date.now().toString(),
      originalText: selectedText,
      aiGeneratedText: streamedContent.value,
      functionType: 'generate',
      modelName: currentModel.value, // 添加模型名称
      timestamp: new Date(),
      isApplied: false
    }
    
    aiResults.value.unshift(aiResult)
    currentAiResult.value = aiResult
    
    ElMessage.success('内容生成完成')
    
  } catch (error) {
    console.error('AI生成失败:', error)
    ElMessage.error('AI生成失败，请重试')
  } finally {
    isAiProcessing.value = false
    isStreaming.value = false
    isStreamingInDialog.value = false // 新增：重置弹窗流式状态
  }
}

// 应用生成的内容
function applyGeneratedContent(result: AiResult) {
  if (result.functionType === 'generate') {
    // 生成内容：在光标位置插入
    const before = content.value.substring(0, cursorPosition.value)
    const after = content.value.substring(cursorPosition.value)
    content.value = before + result.aiGeneratedText + after
    
    result.isApplied = true
    ElMessage.success('已应用生成的内容')
    
    // 保存文档
    nextTick(() => {
      saveDocument()
    })
    
    // 关闭弹窗
    showGenerateDialog.value = false
  }
}

// 关闭生成弹窗
function closeGenerateDialog() {
  showGenerateDialog.value = false
  generateRequirement.value = ''
  currentAiResult.value = null
}

// 获取默认要求
function getDefaultRequirement(functionType: AiFunctionType): string {
  const requirements: Record<AiFunctionType, string> = {
    polish: '润色这段文字',
    summarize: '总结这段文字',
    continue: '续写这段文字',
    generate: '生成相关内容',
    expand: '扩写这段文字'
  }
  return requirements[functionType]
}

// 应用AI结果
function applyAiResult(result: AiResult) {
  if (result.functionType === 'generate') {
    // 生成内容：在光标位置插入
    const before = content.value.substring(0, cursorPosition.value)
    const after = content.value.substring(cursorPosition.value)
    content.value = before + result.aiGeneratedText + after
  } else if (result.functionType === 'continue') {
    // 续写：在选中文本后添加
    const before = content.value.substring(0, selectionEnd.value)
    const after = content.value.substring(selectionEnd.value)
    content.value = before + result.aiGeneratedText + after
  } else {
    // 润色和总结：替换选中文本
    const before = content.value.substring(0, selectionStart.value)
    const after = content.value.substring(selectionEnd.value)
    content.value = before + result.aiGeneratedText + after
  }
  
  result.isApplied = true
  ElMessage.success('已应用AI结果')
  
  // 保存文档
  nextTick(() => {
    saveDocument()
  })
  
  // 关闭弹窗
  showAiDialog.value = false
}

// 拒绝AI结果
function rejectAiResult(result: AiResult) {
  const index = aiResults.value.findIndex(r => r.id === result.id)
  if (index > -1) {
    aiResults.value.splice(index, 1)
  }
  if (currentAiResult.value?.id === result.id) {
    currentAiResult.value = null
  }
  ElMessage.info('已拒绝AI结果')
}

// 清空选择
function clearSelection() {
  selectedText.value = ''
  selectionStart.value = 0
  selectionEnd.value = 0
  currentAiResult.value = null
}

// 关闭AI弹窗
function closeAiDialog() {
  showAiDialog.value = false
  clearSelection()
  // 设置标志位，防止弹窗重复打开
  isDialogJustClosed.value = true
  // 延迟重置标志位，给用户足够时间完成关闭操作
  setTimeout(() => {
    isDialogJustClosed.value = false
  }, 300)
}

// 处理AI功能按钮点击
async function handleAiFunction(functionType: AiFunctionType) {
  if (!hasSelection.value || isAiProcessing.value) return
  
  // 设置选中的AI功能
  selectedAiFunction.value = functionType
  
  // 打开AI弹窗
  showAiDialog.value = true
  
  // 设置默认要求
  userRequirement.value = getDefaultRequirement(functionType)
}

// 新增：多轮对话功能
async function sendChatMessage() {
  if (!chatMessage.value.trim() || isChatProcessing.value) return
  
  try {
    isChatProcessing.value = true
    
    // 添加用户消息到历史
    chatHistory.value.push({
      role: 'user',
      content: chatMessage.value
    })
    
    const userMessage = chatMessage.value
    chatMessage.value = ''
    
    // 调用多轮对话接口
    const response = await chat({
      message: userMessage,
      sessionID: sessionID.value,
      modelName: currentModel.value
    })
    
    // 添加AI回复到历史
    chatHistory.value.push({
      role: 'assistant',
      content: response.result
    })
    
    ElMessage.success('对话完成')
    
  } catch (error) {
    console.error('对话失败:', error)
    ElMessage.error('对话失败，请重试')
  } finally {
    isChatProcessing.value = false
  }
}

// 停止流式生成
function stopStreaming() {
  isStreaming.value = false
  isStreamingInDialog.value = false
  isAiProcessing.value = false
  ElMessage.info('已停止AI生成')
}

// 打开多轮对话弹窗
function openChatDialog() {
  showChatDialog.value = true
  chatMessage.value = ''
}

// 关闭多轮对话弹窗
function closeChatDialog() {
  showChatDialog.value = false
  chatMessage.value = ''
}

// 清空对话历史
function clearChatHistory() {
  chatHistory.value = []
  sessionID.value = generateSessionID()
  ElMessage.info('对话历史已清空')
}

// 打开生成内容弹窗
function openGenerateDialog() {
  showGenerateDialog.value = true
  generateRequirement.value = ''
  // 根据是否有内容自动选择生成类型
  generateType.value = content.value.length > 0 ? 'related' : 'new'
}

// 打开总结全文章弹窗
function openSummaryDialog() {
  showSummaryDialog.value = true
  summaryStreamedContent.value = ''
  summaryResult.value = ''
}

// 关闭总结全文章弹窗
function closeSummaryDialog() {
  showSummaryDialog.value = false
  summaryStreamedContent.value = ''
  summaryResult.value = ''
}

// 处理流式总结回调
const handleSummaryStreamCallback: StreamCallback = (chunk: string, isComplete: boolean) => {
  if (isComplete) {
    // 流式完成
    isSummaryProcessing.value = false
    summaryResult.value = summaryStreamedContent.value // 将流式内容赋值给结果
    ElMessage.success('全文章总结完成')
  } else {
    // 接收流式内容
    try {
      // 尝试解析JSON，如果是错误信息则抛出
      const parsed = JSON.parse(chunk)
      if (parsed.error) {
        throw new Error(parsed.error)
      }
      // 如果是正常内容，直接添加到流式内容中
      if (parsed.content) {
        summaryStreamedContent.value += parsed.content
      } else {
        summaryStreamedContent.value += chunk
      }
    } catch {
      // 如果不是JSON，直接作为文本内容处理
      summaryStreamedContent.value += chunk
    }
  }
}

// 总结全文章（支持流式）
async function summarizeFullArticleStream() {
  if (!canUseAi.value || isSummaryProcessing.value) return
  
  try {
    isSummaryProcessing.value = true
    summaryStreamedContent.value = ''
    
    const requestData: UnifiedAiRequest = {
      functionType: 'summarize',
      documentSummary: generateDocumentSummary(content.value),
      userRequirement: '请总结整篇文章的核心要点和主要观点',
      selectedText: content.value,
      contextText: content.value,
      cursorPosition: 0,
      modelName: currentModel.value, // 使用当前选择的模型
      sessionID: sessionID.value,    // 添加会话ID
      stream: true                   // 启用流式返回
    }
    
    // 使用流式AI接口
    await callUnifiedAIStream(requestData, handleSummaryStreamCallback)
    
  } catch (error) {
    console.error('AI总结失败:', error)
    ElMessage.error('AI总结失败，请重试')
  } finally {
    isSummaryProcessing.value = false
  }
}

// 复制总结结果
async function copySummaryResult() {
  try {
    await navigator.clipboard.writeText(summaryResult.value)
    ElMessage.success('总结结果已复制到剪贴板')
  } catch (error) {
    console.error('复制失败:', error)
    ElMessage.error('复制失败，请手动复制')
  }
}

// 采纳总结结果
function applySummaryResult() {
  // 将总结结果添加到文章末尾
  if (content.value.trim()) {
    content.value += '\n\n## 文章总结\n\n' + summaryResult.value
  } else {
    content.value = summaryResult.value
  }
  
  ElMessage.success('总结结果已添加到文章')
  
  // 保存文档
  nextTick(() => {
    saveDocument()
  })
  
  // 关闭弹窗
  showSummaryDialog.value = false
}

// 重新总结
function regenerateSummary() {
  summaryResult.value = ''
  summaryStreamedContent.value = ''
}

// 字体控制相关函数
function toggleBold() {
  if (!hasSelection.value) {
    ElMessage.warning('请先选择要格式化的文本')
    return
  }
  applyFormatting('bold')
}

function toggleItalic() {
  if (!hasSelection.value) {
    ElMessage.warning('请先选择要格式化的文本')
    return
  }
  applyFormatting('italic')
}

function toggleUnderline() {
  if (!hasSelection.value) {
    ElMessage.warning('请先选择要格式化的文本')
    return
  }
  applyFormatting('underline')
}

function clearFormatting() {
  if (!hasSelection.value) {
    ElMessage.warning('请先选择要清除格式的文本')
    return
  }
  applyFormatting('clear')
}

function applyFormatting(formatType: 'bold' | 'italic' | 'underline' | 'clear') {
  const textarea = document.querySelector('textarea') as HTMLTextAreaElement
  if (!textarea) return

  const text = content.value
  let start = selectionStart.value
  let end = selectionEnd.value

  // 确保选择范围在文本内
  if (start < 0) start = 0
  if (end > text.length) end = text.length
  if (start >= end) return

  // 提取选中文本
  const selectedText = text.substring(start, end)
  
  // 根据格式类型处理文本
  let formattedText = selectedText
  
  switch (formatType) {
    case 'bold':
      // 检查是否已经加粗（支持嵌套标签的情况）
      if (isTextBold(selectedText)) {
        // 移除加粗标签
        formattedText = removeBoldTags(selectedText)
      } else {
        // 添加加粗标签
        formattedText = `<strong>${selectedText}</strong>`
      }
      break
      
    case 'italic':
      // 检查是否已经斜体
      if (isTextItalic(selectedText)) {
        // 移除斜体标签
        formattedText = removeItalicTags(selectedText)
      } else {
        // 添加斜体标签
        formattedText = `<em>${selectedText}</em>`
      }
      break
      
    case 'underline':
      // 检查是否已经下划线
      if (isTextUnderlined(selectedText)) {
        // 移除下划线标签
        formattedText = removeUnderlineTags(selectedText)
      } else {
        // 添加下划线标签
        formattedText = `<u>${selectedText}</u>`
      }
      break
      
    case 'clear':
      // 清除所有HTML标签
      formattedText = selectedText.replace(/<\/?(strong|em|u|b|i)>/g, '')
      break
  }

  // 替换选中文本
  const newText = text.substring(0, start) + formattedText + text.substring(end)
  
  // 更新内容
  content.value = newText
  
  // 更新选择范围
  const newStart = start
  const newEnd = start + formattedText.length
  
  // 使用nextTick确保DOM更新后再设置选择范围
  nextTick(() => {
    if (textarea) {
      textarea.setSelectionRange(newStart, newEnd)
      selectionStart.value = newStart
      selectionEnd.value = newEnd
      cursorPosition.value = newEnd
      
      // 聚焦到文本框
      textarea.focus()
    }
  })
}

// 辅助函数：检查文本是否已加粗
function isTextBold(text: string): boolean {
  return text.includes('<strong>') || text.includes('<b>')
}

// 辅助函数：检查文本是否已斜体
function isTextItalic(text: string): boolean {
  return text.includes('<em>') || text.includes('<i>')
}

// 辅助函数：检查文本是否已有下划线
function isTextUnderlined(text: string): boolean {
  return text.includes('<u>')
}

// 辅助函数：移除加粗标签
function removeBoldTags(text: string): string {
  return text.replace(/<\/?strong>/g, '').replace(/<\/?b>/g, '')
}

// 辅助函数：移除斜体标签
function removeItalicTags(text: string): string {
  return text.replace(/<\/?em>/g, '').replace(/<\/?i>/g, '')
}

// 辅助函数：移除下划线标签
function removeUnderlineTags(text: string): string {
  return text.replace(/<\/?u>/g, '')
}



// 切换预览显示
function togglePreview() {
  showPreview.value = !showPreview.value
}


// 生命周期
onMounted(async () => {
  await init()
  startAutoSave()
})

onBeforeUnmount(() => {
  stopAutoSave()
  // 清理防抖定时器
  if (selectionDebounceTimer) {
    clearTimeout(selectionDebounceTimer)
  }
})
</script>

<template>
  <div class="editor-container">
    <el-page-header :content="t('editor')" @back="$router.back()" />
    
    <div class="editor-main">
      <!-- 编辑器区域 -->
      <el-card class="editor-card">
        <div class="editor-header">
          <div class="status-tags">
            <el-tag type="info">{{ t('chars') }}{{ wordCount }}</el-tag>
            <el-tag type="success" v-if="lastSavedAt">
              {{ t('savedAt') }}{{ lastSavedAt?.toLocaleTimeString() }}
            </el-tag>
            <el-tag type="warning" v-if="isSaving">{{ t('saving') }}</el-tag>
          </div>
          
          <div class="action-buttons">
            <el-button 
              type="primary" 
              :loading="isAiProcessing"
              :disabled="!canUseAi"
              @click="openSummaryDialog"
              size="large"
            >
              📝 总结全文章
            </el-button>
            <el-button 
              type="success" 
              :loading="isAiProcessing"
              @click="openGenerateDialog"
              size="large"
            >
              ✨ 按要求生成内容
            </el-button>
            <el-button 
              type="info" 
              :loading="isChatProcessing"
              @click="openChatDialog"
              size="large"
            >
              💬 多轮对话
            </el-button>
          </div>
        </div>
        
        <div class="editor-content">
          <!-- 字体控制工具栏 -->
          <div class="font-control-toolbar" v-if="!showPreview">
            <div class="toolbar-group">
              <el-button 
                type="default"
                size="small"
                @click="toggleBold"
                icon="Bold"
              >
                加粗
              </el-button>
              <el-button 
                type="default"
                size="small"
                @click="toggleItalic"
                icon="Italic"
              >
                斜体
              </el-button>
              <el-button 
                type="default"
                size="small"
                @click="toggleUnderline"
                icon="Underline"
              >
                下划线
              </el-button>
            </div>
            
            <div class="toolbar-group">
              <el-button 
                size="small"
                @click="clearFormatting"
                icon="Refresh"
              >
                清除格式
              </el-button>
            </div>
            
            <div class="toolbar-group">
              <el-button 
                size="small"
                @click="togglePreview"
                type="primary"
                icon="View"
              >
                显示预览
              </el-button>
            </div>
          </div>

          <!-- 预览模式工具栏 -->
          <div class="preview-toolbar" v-if="showPreview">
            <div class="preview-info">
              <el-tag type="info" size="large">
                <el-icon><View /></el-icon>
                预览模式 - 查看格式化效果
              </el-tag>
            </div>
            <div class="preview-actions">
              <el-button 
                size="large"
                @click="togglePreview"
                type="default"
                icon="Edit"
              >
                返回编辑
              </el-button>
            </div>
          </div>

          <!-- 编辑模式内容 -->
          <div v-if="!showPreview">
            <!-- AI功能按钮区域 -->
            <div class="ai-function-bar" v-if="hasSelection">
              <div class="selection-info">
                <el-tag type="info" size="small">
                  已选择 {{ selectedText.length }} 字符
                </el-tag>
              </div>
              <div class="ai-buttons">
                <el-button
                  type="primary"
                  size="small"
                  :loading="isAiProcessing"
                  @click="handleAiFunction('polish')"
                  :disabled="!hasSelection"
                >
                  ✨ 润色
                </el-button>
                <el-button
                  type="success"
                  size="small"
                  :loading="isAiProcessing"
                  @click="handleAiFunction('summarize')"
                  :disabled="!hasSelection"
                >
                  📝 总结
                </el-button>
                <el-button
                  type="warning"
                  size="small"
                  :loading="isAiProcessing"
                  @click="handleAiFunction('continue')"
                  :disabled="!hasSelection"
                >
                  ➡️ 续写
                </el-button>
                <el-button
                  type="info"
                  size="small"
                  :loading="isAiProcessing"
                  @click="handleAiFunction('expand')"
                  :disabled="!hasSelection"
                >
                  🔍 扩写
                </el-button>
                <el-button
                  type="danger"
                  size="small"
                  @click="clearSelection"
                >
                  ❌ 取消选择
                </el-button>
              </div>
            </div>
            
            <el-input
              v-model="content"
              type="textarea"
              :rows="25"
              :placeholder="t('startWriting')"
              @select="handleTextSelectionDebounced"
              @click="handleTextSelectionDebounced"
              @keyup="handleTextSelectionDebounced"
              class="content-textarea"
              :style="{
                fontWeight: isBold ? 'bold' : 'normal',
                fontStyle: isItalic ? 'italic' : 'normal',
                textDecoration: isUnderline ? 'underline' : 'none'
              }"
            />
          </div>

          <!-- 预览模式内容 -->
          <div v-if="showPreview" class="format-preview">
            <div class="preview-header">
              <h4>格式化预览</h4>
              <p class="preview-tip">这里显示真正的HTML格式化效果，包括加粗、斜体、下划线等</p>
            </div>
            <div 
              class="preview-content"
              v-html="formatPreviewContent"
            ></div>
          </div>
        </div>
        
        <!-- 固定在底部的操作按钮 -->
        <div class="editor-actions-fixed">
          <el-button type="primary" :loading="isSaving" @click="saveDocument">
            {{ t('save') }}
          </el-button>
          <el-button @click="$router.push('/')">{{ t('backHome') }}</el-button>
        </div>
      </el-card>
    </div>
    
    <!-- AI功能弹窗 -->
    <el-dialog
      v-model="showAiDialog"
      title="AI 写作助手"
      width="600px"
      :close-on-click-modal="false"
      @close="closeAiDialog"
    >
      <div class="ai-dialog-content">
        <!-- 选择信息 -->
        <div class="selection-info" v-if="hasSelection">
          <el-alert
            title="已选择文本"
            :description="`已选择 ${selectedText.length} 字符`"
            type="info"
            show-icon
            :closable="false"
          />
          <div class="selected-text-preview">
            <strong>选中内容：</strong>
            <div class="text-preview">{{ selectedText }}</div>
          </div>
        </div>
        
        <!-- AI功能选择 -->
        <div class="ai-function-selector">
          <label>选择AI功能：</label>
          <el-radio-group v-model="selectedAiFunction" size="large">
            <el-radio-button value="polish">润色</el-radio-button>
            <el-radio-button value="summarize">总结</el-radio-button>
            <el-radio-button value="continue">续写</el-radio-button>
            <el-radio-button value="expand">扩写</el-radio-button>
          </el-radio-group>
        </div>
        
        <!-- AI模型选择 -->
        <div class="ai-model-selector">
          <label>选择AI模型：</label>
          <el-select
            v-model="currentModel"
            placeholder="选择AI模型"
            :loading="isModelLoading"
            @change="handleSwitchModel"
            style="width: 100%"
          >
            <el-option
              v-for="model in availableModels"
              :key="model.name"
              :label="model.displayName"
              :value="model.name"
              :disabled="!model.isAvailable"
            >
              <div class="model-option">
                <div class="model-name">{{ model.displayName }}</div>
                <div class="model-desc">{{ model.description }}</div>
                <div class="model-status">
                  <el-tag :type="model.isAvailable ? 'success' : 'danger'" size="small">
                    {{ model.isAvailable ? '可用' : '不可用' }}
                  </el-tag>
                </div>
              </div>
            </el-option>
          </el-select>
        </div>
        
        <!-- 用户要求输入 -->
        <div class="requirement-input">
          <label>具体要求：</label>
          <el-input
            v-model="userRequirement"
            type="textarea"
            :rows="3"
            :placeholder="getDefaultRequirement(selectedAiFunction)"
          />
        </div>
        
        <!-- AI操作按钮 -->
        <div class="ai-actions">
          <el-button
            type="primary"
            :loading="isAiProcessing"
            @click="callAI"
            size="large"
            style="width: 100%"
            :disabled="isStreamingInDialog"
          >
            {{ isAiProcessing ? 'AI处理中...' : '开始AI处理' }}
          </el-button>
        </div>
        
        <!-- 流式内容实时显示 -->
        <div class="streaming-content" v-if="isStreamingInDialog">
          <el-divider content-position="left">
            <el-tag type="warning" size="small">AI正在生成中...</el-tag>
          </el-divider>
          
          <div class="streaming-text-container">
            <div class="streaming-text">
              {{ streamedContent || '正在思考中...' }}
              <span v-if="isStreamingInDialog" class="typing-cursor">|</span>
            </div>
          </div>
          
          <div class="streaming-actions">
            <el-button
              type="info"
              size="small"
              @click="stopStreaming"
              :disabled="!isStreamingInDialog"
            >
              停止生成
            </el-button>
          </div>
        </div>
        
        <!-- AI结果展示 -->
        <div class="ai-results" v-if="currentAiResult && !isStreamingInDialog">
          <el-divider content-position="left">AI 处理结果</el-divider>
          
          <div class="result-content">
            <div class="ai-text">
              <strong>AI结果：</strong>
              <div class="result-text">{{ currentAiResult.aiGeneratedText }}</div>
            </div>
          </div>
          
          <div class="result-actions">
            <el-button
              type="success"
              size="large"
              @click="applyAiResult(currentAiResult)"
            >
              采纳结果
            </el-button>
            <el-button
              type="danger"
              size="large"
              @click="rejectAiResult(currentAiResult)"
            >
              拒绝结果
            </el-button>
          </div>
        </div>
      </div>
    </el-dialog>
    
    <!-- 生成内容弹窗 -->
    <el-dialog
      v-model="showGenerateDialog"
      title="按要求生成内容"
      width="600px"
      :close-on-click-modal="false"
      @close="closeGenerateDialog"
    >
      <div class="generate-dialog-content">
        <!-- 生成类型选择 -->
        <div class="generate-type-selector">
          <label>生成类型：</label>
          <el-radio-group v-model="generateType" size="large">
            <el-radio-button value="new">生成新内容</el-radio-button>
            <el-radio-button value="related" :disabled="!content.trim()">根据文章生成</el-radio-button>
          </el-radio-group>
          <div class="type-description">
            <el-text v-if="generateType === 'new'" type="info">
              根据您的描述生成全新的内容
            </el-text>
            <el-text v-else type="info">
              根据当前文章内容和您的要求生成相关内容，保持风格一致
            </el-text>
          </div>
        </div>
        
        <!-- 生成要求输入 -->
        <div class="generate-requirement-input">
          <label>生成要求：</label>
          <el-input
            v-model="generateRequirement"
            type="textarea"
            :rows="4"
            placeholder="请详细描述您希望生成的内容，例如：主题、风格、长度、结构等要求"
          />
        </div>
        
        <!-- 文章内容提示 -->
        <div class="content-hint" v-if="generateType === 'related' && content.trim()">
          <el-alert
            title="文章内容提示"
            :description="`当前文章共 ${wordCount} 字符，AI将根据文章内容和您的要求生成相关内容`"
            type="info"
            show-icon
            :closable="false"
          />
        </div>
        
        <!-- 生成操作按钮 -->
        <div class="generate-actions">
          <el-button
            type="primary"
            :loading="isAiProcessing"
            @click="generateContent"
            size="large"
            style="width: 100%"
            :disabled="isStreamingInDialog"
          >
            {{ isAiProcessing ? 'AI生成中...' : '开始生成内容' }}
          </el-button>
        </div>
        
        <!-- 流式内容实时显示 -->
        <div class="streaming-content" v-if="isStreamingInDialog">
          <el-divider content-position="left">
            <el-tag type="warning" size="small">AI正在生成中...</el-tag>
          </el-divider>
          
          <div class="streaming-text-container">
            <div class="streaming-text">
              {{ streamedContent || '正在思考中...' }}
              <span v-if="isStreamingInDialog" class="typing-cursor">|</span>
            </div>
          </div>
          
          <div class="streaming-actions">
            <el-button
              type="info"
              size="small"
              @click="stopStreaming"
              :disabled="!isStreamingInDialog"
            >
              停止生成
            </el-button>
          </div>
        </div>
        
        <!-- 生成结果展示 -->
        <div class="generate-results" v-if="currentAiResult && currentAiResult.functionType === 'generate' && !isStreamingInDialog">
          <el-divider content-position="left">生成结果</el-divider>
          
          <div class="result-content">
            <div class="ai-text">
              <strong>生成内容：</strong>
              <div class="result-text">{{ currentAiResult.aiGeneratedText }}</div>
            </div>
          </div>
          
          <div class="result-actions">
            <el-button
              type="success"
              size="large"
              @click="applyGeneratedContent(currentAiResult)"
            >
              采纳内容
            </el-button>
            <el-button
              type="danger"
              size="large"
              @click="rejectAiResult(currentAiResult)"
            >
              拒绝内容
            </el-button>
          </div>
        </div>
      </div>
    </el-dialog>
    
    <!-- 多轮对话弹窗 -->
    <el-dialog
      v-model="showChatDialog"
      title="多轮对话"
      width="800px"
      :close-on-click-modal="false"
      @close="closeChatDialog"
    >
      <div class="chat-dialog-content">
        <!-- 对话历史 -->
        <div class="chat-history" v-if="chatHistory.length > 0">
          <el-divider content-position="left">对话历史</el-divider>
          <div class="history-messages">
            <div 
              v-for="(message, index) in chatHistory" 
              :key="index"
              :class="['message', message.role]"
            >
              <div class="message-role">
                <el-tag :type="message.role === 'user' ? 'primary' : 'success'" size="small">
                  {{ message.role === 'user' ? '用户' : 'AI助手' }}
                </el-tag>
              </div>
              <div class="message-content">{{ message.content }}</div>
            </div>
          </div>
          <div class="history-actions">
            <el-button @click="clearChatHistory" size="small">
              清空历史
            </el-button>
          </div>
        </div>
        
        <!-- 当前AI模型 -->
        <div class="current-model" v-if="!chatHistory.length">
          <el-alert
            title="当前AI模型"
            :description="`正在使用 ${currentModel} 模型进行对话`"
            type="info"
            show-icon
            :closable="false"
          />
        </div>
        
        <!-- 输入区域 -->
        <div class="chat-input">
          <el-input
            v-model="chatMessage"
            type="textarea"
            :rows="3"
            placeholder="请输入您的问题或要求..."
            @keyup.enter.ctrl="sendChatMessage"
          />
          <div class="input-tip">
            <el-text type="info" size="small">
              按 Ctrl+Enter 发送消息
            </el-text>
          </div>
        </div>
        
        <!-- 发送按钮 -->
        <div class="chat-actions">
          <el-button
            type="primary"
            :loading="isChatProcessing"
            @click="sendChatMessage"
            size="large"
            style="width: 100%"
          >
            {{ isChatProcessing ? 'AI思考中...' : '发送消息' }}
          </el-button>
        </div>
      </div>
    </el-dialog>
    
    <!-- 总结全文章弹窗 -->
    <el-dialog
      v-model="showSummaryDialog"
      title="总结全文章"
      width="800px"
      :close-on-click-modal="false"
      @close="closeSummaryDialog"
    >
      <div class="summary-dialog-content">
        <!-- 文章信息 -->
        <div class="article-info">
          <el-alert
            title="文章信息"
            :description="`当前文章共 ${wordCount} 字符，AI将为您总结整篇文章的核心要点和主要观点`"
            type="info"
            show-icon
            :closable="false"
          />
        </div>
        
        <!-- AI模型选择 -->
        <div class="ai-model-selector">
          <label>选择AI模型：</label>
          <el-select
            v-model="currentModel"
            placeholder="选择AI模型"
            :loading="isModelLoading"
            @change="handleSwitchModel"
            style="width: 100%"
          >
            <el-option
              v-for="model in availableModels"
              :key="model.name"
              :label="model.displayName"
              :value="model.name"
              :disabled="!model.isAvailable"
            >
              <div class="model-option">
                <div class="model-name">{{ model.displayName }}</div>
                <div class="model-desc">{{ model.description }}</div>
                <div class="model-status">
                  <el-tag :type="model.isAvailable ? 'success' : 'danger'" size="small">
                    {{ model.isAvailable ? '可用' : '不可用' }}
                  </el-tag>
                </div>
              </div>
            </el-option>
          </el-select>
        </div>
        
        <!-- 开始总结按钮 -->
        <div class="summary-actions" v-if="!summaryResult">
          <el-button
            type="primary"
            :loading="isSummaryProcessing"
            @click="summarizeFullArticleStream"
            size="large"
            style="width: 100%"
            :disabled="!canUseAi"
          >
            {{ isSummaryProcessing ? 'AI总结中...' : '开始总结全文章' }}
          </el-button>
        </div>
        
        <!-- 流式内容实时显示 -->
        <div class="streaming-content" v-if="isSummaryProcessing">
          <el-divider content-position="left">
            <el-tag type="warning" size="small">AI正在总结中...</el-tag>
          </el-divider>
          
          <div class="streaming-text-container">
            <div class="streaming-text">
              {{ summaryStreamedContent || '正在分析文章内容...' }}
              <span v-if="isSummaryProcessing" class="typing-cursor">|</span>
            </div>
          </div>
          
          <div class="streaming-actions">
            <el-button
              type="info"
              size="small"
              @click="stopStreaming"
              :disabled="!isSummaryProcessing"
            >
              停止总结
            </el-button>
          </div>
        </div>
        
        <!-- 总结结果展示 -->
        <div class="summary-results" v-if="summaryResult && !isSummaryProcessing">
          <el-divider content-position="left">总结结果</el-divider>
          
          <div class="result-content">
            <div class="summary-text">
              <strong>文章总结：</strong>
              <div class="result-text">{{ summaryResult }}</div>
            </div>
          </div>
          
          <div class="result-actions">
            <el-button
              type="primary"
              size="large"
              @click="copySummaryResult"
              style="margin-right: 16px"
            >
              📋 复制总结
            </el-button>
            <el-button
              type="success"
              size="large"
              @click="applySummaryResult"
            >
              ✅ 采纳总结
            </el-button>
            <el-button
              type="info"
              size="large"
              @click="regenerateSummary"
            >
              🔄 重新总结
            </el-button>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.editor-container {
  padding: 16px;
  height: 100vh;
  overflow: hidden;
}

.editor-main {
  margin-top: 16px;
  height: calc(100vh - 120px);
}

.editor-card {
  height: 100%;
  overflow: auto;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.status-tags {
  display: flex;
  gap: 8px;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.editor-content {
  flex: 1;
  margin-bottom: 16px;
  padding-bottom: 80px; /* 为固定底部按钮留出空间 */
}

/* 字体控制工具栏样式 */
.font-control-toolbar {
  display: flex;
  gap: 10px;
  padding: 8px 12px;
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 8px;
  margin-bottom: 16px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  align-items: center;
  flex-wrap: wrap;
}

.toolbar-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toolbar-group label {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
}

.toolbar-group .el-button {
  width: 100px;
}

.toolbar-group .el-button {
  font-size: 12px;
  padding: 6px 12px;
  border-radius: 6px;
}



/* 预览模式工具栏样式 */
.preview-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 12px 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: 1px solid #5a6fd8;
  border-radius: 8px;
  margin-bottom: 16px;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.preview-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.preview-info .el-tag {
  background: rgba(255, 255, 255, 0.9);
  border: none;
  color: #5a6fd8;
  font-weight: 600;
  font-size: 14px;
}

.preview-info .el-icon {
  margin-right: 4px;
}

.preview-actions {
  display: flex;
  gap: 12px;
}

.preview-actions .el-button {
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: #5a6fd8;
  font-weight: 600;
  transition: all 0.3s ease;
}

.preview-actions .el-button:hover {
  background: rgba(255, 255, 255, 1);
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

/* 格式化预览区域样式 */
.format-preview {
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  border: 2px solid #667eea;
  border-radius: 12px;
  margin-bottom: 16px;
  padding: 20px;
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.15);
  position: relative;
  overflow: hidden;
}

.format-preview::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, #667eea, #764ba2, #667eea);
  background-size: 200% 100%;
  animation: gradient-move 3s ease-in-out infinite;
}

@keyframes gradient-move {
  0% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
  100% { background-position: 0% 50%; }
}

.preview-header {
  margin-bottom: 16px;
  border-bottom: 2px solid #667eea;
  padding-bottom: 12px;
  text-align: center;
}

.preview-header h4 {
  margin: 0 0 8px 0;
  color: #5a6fd8;
  font-size: 20px;
  font-weight: 700;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.preview-tip {
  margin: 0;
  color: #667eea;
  font-size: 14px;
  font-weight: 500;
}

.preview-content {
  background: white;
  border: 2px solid #667eea;
  border-radius: 8px;
  padding: 20px;
  min-height: 120px;
  line-height: 1.8;
  white-space: pre-wrap;
  word-wrap: break-word;
  overflow-wrap: break-word;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.1);
  transition: all 0.3s ease;
}

.preview-content:hover {
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.2);
  transform: translateY(-2px);
}

/* 确保HTML标签正确渲染 */
.preview-content strong,
.preview-content b {
  font-weight: bold !important;
  color: inherit;
}

.preview-content em,
.preview-content i {
  font-style: italic !important;
  color: inherit;
}

.preview-content u {
  text-decoration: underline !important;
  color: inherit;
}

/* 支持嵌套标签 */
.preview-content strong strong,
.preview-content b b {
  font-weight: 900 !important;
}

.preview-content em em,
.preview-content i i {
  font-style: oblique !important;
}

/* 支持字体大小和颜色标签 */
.preview-content span[style*="font-size"] {
  font-size: inherit !important;
}

.preview-content span[style*="color"] {
  color: inherit !important;
}

/* AI功能按钮区域样式 */
.ai-function-bar {
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.selection-info {
  display: flex;
  align-items: center;
  flex-direction: column;
}

.ai-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.ai-buttons .el-button {
  font-size: 12px;
  padding: 6px 12px;
  border-radius: 6px;
}

.content-textarea {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 14px;
  line-height: 1.6;
}

.editor-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

/* AI弹窗样式 */
.ai-dialog-content {
  padding: 16px 0;
}

.selection-info {
  margin-bottom: 20px;
}

.selected-text-preview {
  margin-top: 12px;
}

.text-preview {
  background: #f5f7fa;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 8px;
  margin-top: 8px;
  max-height: 100px;
  overflow-y: auto;
  font-size: 13px;
  line-height: 1.4;
}

.ai-function-selector {
  margin-bottom: 20px;
}

.ai-function-selector label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #606266;
}

.ai-model-selector {
  margin-bottom: 20px;
}

.ai-model-selector label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #606266;
}

/* 模型选项样式 */
.model-option {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.model-name {
  font-weight: 500;
  color: #303133;
}

.model-desc {
  font-size: 12px;
  color: #909399;
  line-height: 1.3;
}

.model-status {
  display: flex;
  justify-content: flex-end;
}

.requirement-input {
  margin-bottom: 20px;
}

.requirement-input label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #606266;
}

.ai-actions {
  margin-bottom: 20px;
}

.ai-results {
  margin-top: 20px;
}

.result-content {
  margin-bottom: 20px;
}

.result-text {
  background: #f0f9ff;
  border: 1px solid #b3d8ff;
  border-radius: 4px;
  padding: 12px;
  margin-top: 8px;
  line-height: 1.6;
  white-space: pre-wrap;
}

.result-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

/* 生成内容弹窗样式 */
.generate-dialog-content {
  padding: 16px 0;
}

.generate-type-selector {
  margin-bottom: 20px;
}

.generate-type-selector label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #606266;
}

.type-description {
  margin-top: 8px;
  font-size: 13px;
}

.generate-requirement-input {
  margin-bottom: 20px;
}

.generate-requirement-input label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #606266;
}

.content-hint {
  margin-bottom: 20px;
}

.generate-actions {
  margin-bottom: 20px;
}

.generate-results {
  margin-top: 20px;
}

/* 流式内容样式 */
.streaming-content {
  margin: 20px 0;
  padding: 16px;
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 8px;
}

.streaming-text-container {
  margin: 16px 0;
}

.streaming-text {
  background: #f0f9ff;
  border: 1px solid #b3d8ff;
  border-radius: 4px;
  padding: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  min-height: 60px;
  position: relative;
}

.typing-cursor {
  animation: blink 1s infinite;
  color: #409eff;
  font-weight: bold;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

.streaming-actions {
  display: flex;
  justify-content: center;
  margin-top: 16px;
}

/* 多轮对话弹窗样式 */
.chat-dialog-content {
  padding: 16px 0;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.chat-history {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
  background: #f5f7fa;
  border: 1px solid #e9ecef;
  border-radius: 8px;
  margin-bottom: 16px;
}

.history-messages {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.message {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.message-role {
  flex-shrink: 0;
}

.message-content {
  flex: 1;
  background: #e9ecef;
  border-radius: 8px;
  padding: 8px 12px;
  font-size: 14px;
  line-height: 1.5;
  word-break: break-word;
}

.message.user .message-content {
  background: #e1f3d8;
  border: 1px solid #a5d6a7;
}

.message.assistant .message-content {
  background: #f0f9ff;
  border: 1px solid #b3d8ff;
}

.history-actions {
  display: flex;
  justify-content: flex-end;
  padding: 0 10px 10px;
}

.current-model {
  padding: 10px;
  background: #f5f7fa;
  border: 1px solid #e9ecef;
  border-radius: 8px;
  margin-bottom: 16px;
}

.chat-input {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px;
  background: #f5f7fa;
  border: 1px solid #e9ecef;
  border-radius: 8px;
  margin-bottom: 16px;
}

.chat-input .el-textarea {
  flex: 1;
}

.input-tip {
  text-align: right;
  font-size: 12px;
  color: #909399;
}

.chat-actions {
  padding: 0 10px 10px;
}

/* 总结全文章弹窗样式 */
.summary-dialog-content {
  padding: 16px 0;
}

.article-info {
  margin-bottom: 20px;
}

.ai-model-selector {
  margin-bottom: 20px;
}

.summary-actions {
  margin-bottom: 20px;
}

.summary-results {
  margin-top: 20px;
}

.summary-text {
  margin-bottom: 20px;
}

.summary-text .result-text {
  background: #f0f9ff;
  border: 1px solid #b3d8ff;
  border-radius: 4px;
  padding: 12px;
  margin-top: 8px;
  line-height: 1.6;
  white-space: pre-wrap;
  max-height: 300px;
  overflow-y: auto;
}

.summary-results .result-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  flex-wrap: wrap;
}

/* 固定在底部的操作按钮样式 */
.editor-actions-fixed {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: white;
  border-top: 2px solid #667eea;
  padding: 16px;
  box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.1);
  z-index: 1000;
  display: flex;
  gap: 12px;
  justify-content: center;
  align-items: center;
}

.editor-actions-fixed .el-button {
  min-width: 120px;
  height: 40px;
  font-size: 14px;
  font-weight: 600;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .editor-header {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }
  
  .action-buttons {
    width: 100%;
  }
  
  .action-buttons .el-button {
    width: 100%;
  }
  
  .editor-actions-fixed {
    padding: 12px;
  }
  
  .editor-actions-fixed .el-button {
    min-width: 100px;
    height: 36px;
    font-size: 13px;
  }
}
</style>





