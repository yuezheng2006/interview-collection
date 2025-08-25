import { getHttpClient } from './api'
import type { AiFunctionType, AiRequestParams, AiResult, TextSelection } from '@/types'

export type AiContinueRequest = { prompt: string }
export type AiPolishRequest = { text: string }
export type AiSummarizeRequest = { text: string }

export type AiResponse = { result: string }

// AI模型信息
export type AiModelInfo = {
	name: string
	displayName: string
	provider: string
	description: string
	maxTokens: number
	isAvailable: boolean
}

// 模型列表响应
export type ModelsResponse = {
	models: AiModelInfo[]
	current: string
}

// 模型切换请求
export type SwitchModelRequest = {
	modelName: string
}

// 模型切换响应
export type SwitchModelResponse = {
	success: boolean
	message: string
	modelName: string
}

// 新的统一AI接口
export type UnifiedAiRequest = {
	functionType: AiFunctionType
	documentSummary: string
	userRequirement: string
	selectedText: string
	contextText: string
	cursorPosition: number
	modelName?: string // 可选：指定使用的模型
	sessionID?: string // 新增：会话ID，用于多轮对话
	stream?: boolean   // 新增：是否启用流式返回
}

export type UnifiedAiResponse = {
	result: string
	functionType: AiFunctionType
	modelName: string
}

// 多轮对话请求
export type ChatRequest = {
	message: string
	sessionID: string
	modelName?: string
}

// 多轮对话响应
export type ChatResponse = {
	result: string
	modelName: string
}

// 流式AI回调函数类型
export type StreamCallback = (chunk: string, isComplete: boolean) => void

// 智能内容选择函数
export function getSmartTextSelection(
  fullText: string,
  cursorPosition: number,
  selectionStart: number,
  selectionEnd: number
): TextSelection {
  const selectedText = fullText.substring(selectionStart, selectionEnd)
  
  // 计算上下文范围（前后各300字符）
  const contextSize = 300
  const contextStart = Math.max(0, selectionStart - contextSize)
  const contextEnd = Math.min(fullText.length, selectionEnd + contextSize)
  
  // 智能调整边界，避免截断句子
  let adjustedContextStart = contextStart
  let adjustedContextEnd = contextEnd
  
  // 向前查找句子边界
  for (let i = selectionStart - 1; i >= contextStart; i--) {
    if (['。', '！', '？', '.', '!', '?', '\n'].includes(fullText[i])) {
      adjustedContextStart = i + 1
      break
    }
  }
  
  // 向后查找句子边界
  for (let i = selectionEnd; i < contextEnd; i++) {
    if (['。', '！', '？', '.', '!', '?', '\n'].includes(fullText[i])) {
      adjustedContextEnd = i + 1
      break
    }
  }
  
  const contextBefore = fullText.substring(adjustedContextStart, selectionStart)
  const contextAfter = fullText.substring(selectionEnd, adjustedContextEnd)
  
  return {
    start: selectionStart,
    end: selectionEnd,
    text: selectedText,
    contextBefore,
    contextAfter
  }
}

// 生成文档摘要
export function generateDocumentSummary(text: string): string {
  if (text.length <= 200) return text
  
  // 简单摘要：取前100字符 + 后100字符
  const prefix = text.substring(0, 100)
  const suffix = text.substring(text.length - 100)
  return `${prefix}...${suffix}`
}

// 生成会话ID
export function generateSessionID(): string {
  return `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
}

// 获取可用模型列表
export async function getAvailableModels(): Promise<ModelsResponse> {
	const res = await getHttpClient().get<ModelsResponse>('/ai/models')
	return res.data
}

// 切换AI模型
export async function switchModel(modelName: string): Promise<SwitchModelResponse> {
	const res = await getHttpClient().post<SwitchModelResponse>('/ai/switch-model', { modelName })
	return res.data
}

// 多轮对话接口
export async function chat(data: ChatRequest): Promise<ChatResponse> {
	const res = await getHttpClient().post<ChatResponse>('/ai/chat', data)
	return res.data
}

// 流式统一AI接口
export async function callUnifiedAIStream(
  data: UnifiedAiRequest, 
  callback: StreamCallback
): Promise<void> {
  try {
    // 在开发环境中，vite代理会将/api转发到http://localhost:8080
    // 在生产环境中，应该设置正确的API基础URL
    const apiUrl = '/api/ai/unified'
    
    // 获取认证token，与api.ts保持一致
    const token = localStorage.getItem('auth-token') || ''
    
    const response = await fetch(apiUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token, // 直接使用token，不加Bearer前缀，与axios保持一致
        'Accept': 'text/event-stream',
        'Cache-Control': 'no-cache',
        'Connection': 'keep-alive'
      },
      body: JSON.stringify({
        ...data,
        stream: true
      })
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const reader = response.body?.getReader()
    if (!reader) {
      throw new Error('No response body')
    }

    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = line.slice(6)
          if (data === '[DONE]') {
            callback('', true)
            return
          }

          try {
            const parsed = JSON.parse(data)
            if (parsed.error) {
              throw new Error(parsed.error)
            }
            callback(data, false)
          } catch (e) {
            // 如果不是JSON，直接作为文本内容处理
            callback(data, false)
          }
        }
      }
    }

    callback('', true)
  } catch (error) {
    console.error('流式AI调用失败:', error)
    throw error
  }
}

// 统一AI接口（非流式）
export async function callUnifiedAI(data: UnifiedAiRequest): Promise<UnifiedAiResponse> {
	const res = await getHttpClient().post<UnifiedAiResponse>('/ai/unified', data)
	return res.data
}

// 原有接口保持不变
export async function continueWriting(data: AiContinueRequest) {
  const res = await getHttpClient().post<AiResponse>('/ai/continue', data)
  return res.data
}

export async function polishText(data: AiPolishRequest) {
  const res = await getHttpClient().post<AiResponse>('/ai/polish', data)
  return res.data
}

export async function summarizeText(data: AiSummarizeRequest) {
  const res = await getHttpClient().post<AiResponse>('/ai/summarize', data)
  return res.data
}

