package handler

import (
	"context"
	"net/http"

	"ai-writing-assistant/internal/pkg/ai"

	"github.com/gin-gonic/gin"
)

type aiRequest struct {
	Text   string `json:"text"`
	Prompt string `json:"prompt"`
}

type aiResponse struct {
	Result string `json:"result"`
}

// 统一AI请求结构
type unifiedAiRequest struct {
	FunctionType    string `json:"functionType"`
	DocumentSummary string `json:"documentSummary"`
	UserRequirement string `json:"userRequirement"`
	SelectedText    string `json:"selectedText"`
	ContextText     string `json:"contextText"`
	CursorPosition  int    `json:"cursorPosition"`
	ModelName       string `json:"modelName"` // 新增：指定使用的模型
	SessionID       string `json:"sessionID"` // 新增：会话ID，用于多轮对话
	Stream          bool   `json:"stream"`    // 新增：是否启用流式返回
}

// 统一AI响应结构
type unifiedAiResponse struct {
	Result       string `json:"result"`
	FunctionType string `json:"functionType"`
	ModelName    string `json:"modelName"`
}

// 模型切换请求
type switchModelRequest struct {
	ModelName string `json:"modelName"`
}

// 模型切换响应
type switchModelResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ModelName string `json:"modelName"`
}

// 多轮对话请求
type chatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"sessionID"`
	ModelName string `json:"modelName"`
}

// 多轮对话响应
type chatResponse struct {
	Result    string `json:"result"`
	ModelName string `json:"modelName"`
}

func registerAIRoutes(g *gin.RouterGroup) {
	svc := ai.NewService()

	// 注册AI提供者
	svc.Register("mock", ai.MockProvider{})
	// svc.Register("tongyi", ai.NewTongyiProvider())
	svc.Register("deepseek", ai.NewDeepSeekProvider())
	// svc.Register("wenxin", ai.NewWenxinProvider())

	// 使用配置管理系统设置默认模型
	config := ai.GetAIConfig()
	svc.Use(config.AI.DefaultModel)

	// 获取可用模型列表
	g.GET("/ai/models", func(c *gin.Context) {
		models := svc.GetAvailableModels()
		c.JSON(http.StatusOK, gin.H{
			"models":  models,
			"current": svc.GetCurrentModel(),
		})
	})

	// 切换AI模型
	g.POST("/ai/switch-model", func(c *gin.Context) {
		var req switchModelRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
			return
		}

		// 检查模型是否可用
		models := svc.GetAvailableModels()
		var targetModel *ai.ModelInfo
		for _, model := range models {
			if model.Name == req.ModelName {
				targetModel = &model
				break
			}
		}

		if targetModel == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "模型不存在"})
			return
		}

		if !targetModel.IsAvailable {
			c.JSON(http.StatusBadRequest, gin.H{"error": "模型不可用"})
			return
		}

		// 切换模型
		svc.Use(targetModel.Provider)

		c.JSON(http.StatusOK, switchModelResponse{
			Success:   true,
			Message:   "模型切换成功",
			ModelName: req.ModelName,
		})
	})

	// 多轮对话接口
	g.POST("/ai/chat", func(c *gin.Context) {
		var req chatRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
			return
		}

		// 如果指定了模型，先切换
		if req.ModelName != "" && req.ModelName != svc.GetCurrentModel() {
			// 根据模型名称找到对应的provider
			models := svc.GetAvailableModels()
			var targetProvider string
			for _, model := range models {
				if model.Name == req.ModelName {
					targetProvider = model.Provider
					break
				}
			}

			if targetProvider != "" {
				svc.Use(targetProvider)
			}
		}

		// 检查当前provider是否可用
		currentProvider := svc.CurrentProvider()
		if currentProvider == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "当前AI模型不可用，请先切换模型"})
			return
		}

		// 调用多轮对话
		result, err := currentProvider.Chat(context.Background(), req.Message, req.SessionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "AI对话失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, chatResponse{
			Result:    result,
			ModelName: svc.GetCurrentModel(),
		})
	})

	// 统一AI接口（支持流式和非流式）
	g.POST("/ai/unified", func(c *gin.Context) {
		var req unifiedAiRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
			return
		}

		// 如果指定了模型，先切换
		if req.ModelName != "" && req.ModelName != svc.GetCurrentModel() {
			// 根据模型名称找到对应的provider
			models := svc.GetAvailableModels()
			var targetProvider string
			for _, model := range models {
				if model.Name == req.ModelName {
					targetProvider = model.Provider
					break
				}
			}

			if targetProvider != "" {
				svc.Use(targetProvider)
			}
		}

		// 检查当前provider是否可用
		currentProvider := svc.CurrentProvider()
		if currentProvider == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "当前AI模型不可用，请先切换模型"})
			return
		}

		// 如果启用流式返回
		if req.Stream {
			c.Header("Content-Type", "text/event-stream")
			c.Header("Cache-Control", "no-cache")
			c.Header("Connection", "keep-alive")
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Headers", "Cache-Control")

			// 创建流式响应
			streamWriter := &StreamWriter{c: c}

			var prompt string
			// 根据功能类型构建提示词
			switch req.FunctionType {
			case "continue":
				prompt = buildContinuePrompt(req.DocumentSummary, req.UserRequirement, req.ContextText)
			case "polish":
				prompt = buildPolishPrompt(req.UserRequirement, req.SelectedText)
			case "summarize":
				prompt = buildSummarizePrompt(req.UserRequirement, req.SelectedText)
			case "expand":
				prompt = buildExpandPrompt(req.UserRequirement, req.SelectedText, req.ContextText)
			case "generate":
				prompt = buildGeneratePrompt(req.DocumentSummary, req.UserRequirement, req.ContextText)
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的功能类型"})
				return
			}

			// 调用流式AI接口
			err := currentProvider.CallAIStream(context.Background(), req.FunctionType, prompt, req.SessionID, streamWriter)
			if err != nil {
				streamWriter.WriteError(err.Error())
			}
			return
		}

		// 非流式返回（原有逻辑）
		var result string
		var err error

		// 根据功能类型调用不同的AI服务
		switch req.FunctionType {
		case "continue":
			// 续写：基于选中文本和上下文进行续写
			prompt := buildContinuePrompt(req.DocumentSummary, req.UserRequirement, req.ContextText)
			result, err = currentProvider.ContinueWriting(context.Background(), prompt)
		case "polish":
			// 润色：优化选中文本的表达
			prompt := buildPolishPrompt(req.UserRequirement, req.SelectedText)
			result, err = currentProvider.PolishText(context.Background(), prompt)
		case "summarize":
			// 总结：提取选中文本的核心要点
			prompt := buildSummarizePrompt(req.UserRequirement, req.SelectedText)
			result, err = currentProvider.SummarizeText(context.Background(), prompt)
		case "expand":
			// 扩写：基于选中文本和上下文进行扩写
			prompt := buildExpandPrompt(req.UserRequirement, req.SelectedText, req.ContextText)
			result, err = currentProvider.ContinueWriting(context.Background(), prompt)
		case "generate":
			// 生成：根据用户要求生成新内容
			prompt := buildGeneratePrompt(req.DocumentSummary, req.UserRequirement, req.ContextText)
			result, err = currentProvider.ContinueWriting(context.Background(), prompt)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的功能类型"})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "AI处理失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, unifiedAiResponse{
			Result:       result,
			FunctionType: req.FunctionType,
			ModelName:    svc.GetCurrentModel(),
		})
	})

	// 保持原有接口兼容性
	g.POST("/ai/continue", func(c *gin.Context) {
		var req aiRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
			return
		}

		currentProvider := svc.CurrentProvider()
		if currentProvider == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "当前AI模型不可用，请先切换模型"})
			return
		}

		result, err := currentProvider.ContinueWriting(context.Background(), req.Prompt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "AI处理失败: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, aiResponse{Result: result})
	})

	g.POST("/ai/polish", func(c *gin.Context) {
		var req aiRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
			return
		}

		currentProvider := svc.CurrentProvider()
		if currentProvider == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "当前AI模型不可用，请先切换模型"})
			return
		}

		result, err := currentProvider.PolishText(context.Background(), req.Text)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "AI处理失败: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, aiResponse{Result: result})
	})

	g.POST("/ai/summarize", func(c *gin.Context) {
		var req aiRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
			return
		}

		currentProvider := svc.CurrentProvider()
		if currentProvider == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "当前AI模型不可用，请先切换模型"})
			return
		}

		result, err := currentProvider.SummarizeText(context.Background(), req.Text)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "AI处理失败: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, aiResponse{Result: result})
	})
}

// 流式写入器
type StreamWriter struct {
	c *gin.Context
}

func (sw *StreamWriter) Write(p []byte) (n int, err error) {
	// 发送SSE格式的数据
	sw.c.Data(http.StatusOK, "text/event-stream", []byte("data: "+string(p)+"\n\n"))
	sw.c.Writer.Flush()
	return len(p), nil
}

func (sw *StreamWriter) WriteError(message string) {
	sw.c.Data(http.StatusOK, "text/event-stream", []byte("data: {\"error\":\""+message+"\"}\n\n"))
	sw.c.Writer.Flush()
}

// 构建续写提示词
func buildContinuePrompt(documentSummary, userRequirement, contextText string) string {
	return "续写要求：" + userRequirement + "\n\n" +
		"上下文：" + contextText + "\n\n" +
		"请直接续写，保持风格一致。"
}

// 构建润色提示词
func buildPolishPrompt(userRequirement, selectedText string) string {
	return "润色要求：" + userRequirement + "\n\n" +
		"原文：" + selectedText + "\n\n" +
		"请直接输出润色后的文本。"
}

// 构建总结提示词
func buildSummarizePrompt(userRequirement, selectedText string) string {
	return "总结要求：" + userRequirement + "\n\n" +
		"原文：" + selectedText + "\n\n" +
		"请直接输出总结内容，用自然段落形式描述。"
}

// 构建扩写提示词
func buildExpandPrompt(userRequirement, selectedText, contextText string) string {
	return "扩写要求：" + userRequirement + "\n\n" +
		"原文：" + selectedText + "\n\n" +
		"上下文：" + contextText + "\n\n" +
		"请直接输出扩写后的内容。"
}

// 构建生成提示词
func buildGeneratePrompt(documentSummary, userRequirement, contextText string) string {
	return "生成要求：" + userRequirement + "\n\n" +
		"文档概要：" + documentSummary + "\n\n" +
		"请直接输出生成的内容。"
}
