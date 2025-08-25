export type Nullable<T> = T | null

export type ApiResponse<T> = {
	code: number
	message: string
	data: T
}

// AI功能类型
export type AiFunctionType = 'continue' | 'polish' | 'summarize' | 'generate' | 'expand'

// AI模型信息
export type AiModelInfo = {
	name: string
	displayName: string
	provider: string
	description: string
	maxTokens: number
	isAvailable: boolean
}

// AI请求参数
export type AiRequestParams = {
	functionType: AiFunctionType
	documentSummary: string
	userRequirement: string
	selectedText: string
	contextText: string
	cursorPosition: number
	modelName?: string
}

// AI响应结果
export type AiResult = {
	id: string
	originalText: string
	aiGeneratedText: string
	functionType: AiFunctionType
	modelName: string
	timestamp: Date
	isApplied: boolean
}

// 文本选择范围
export type TextSelection = {
	start: number
	end: number
	text: string
	contextBefore: string
	contextAfter: string
}

