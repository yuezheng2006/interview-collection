package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DeepSeek配置
type DeepSeekConfig struct {
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
	Temperature float64
}

// DeepSeek请求结构
type DeepSeekRequest struct {
	Model            string            `json:"model"`
	Messages         []DeepSeekMessage `json:"messages"`
	MaxTokens        int               `json:"max_tokens"`
	Temperature      float64           `json:"temperature"`
	TopP             float64           `json:"top_p"`
	Stream           bool              `json:"stream"`
	FrequencyPenalty float64           `json:"frequency_penalty"`
	PresencePenalty  float64           `json:"presence_penalty"`
	ResponseFormat   *ResponseFormat   `json:"response_format,omitempty"`
	Stop             []string          `json:"stop,omitempty"`
	StreamOptions    interface{}       `json:"stream_options,omitempty"`
	Tools            []interface{}     `json:"tools,omitempty"`
	ToolChoice       string            `json:"tool_choice,omitempty"`
	Logprobs         bool              `json:"logprobs"`
	TopLogprobs      interface{}       `json:"top_logprobs,omitempty"`
}

type ResponseFormat struct {
	Type string `json:"type"`
}

type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeek响应结构
type DeepSeekResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// DeepSeek流式响应结构
type DeepSeekStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

// 对话历史管理
type ConversationHistory struct {
	Messages []DeepSeekMessage
	MaxSize  int
}

func NewConversationHistory(maxSize int) *ConversationHistory {
	return &ConversationHistory{
		Messages: make([]DeepSeekMessage, 0),
		MaxSize:  maxSize,
	}
}

func (ch *ConversationHistory) AddMessage(role, content string) {
	message := DeepSeekMessage{Role: role, Content: content}
	ch.Messages = append(ch.Messages, message)

	// 保持对话历史在最大限制内
	if len(ch.Messages) > ch.MaxSize {
		ch.Messages = ch.Messages[len(ch.Messages)-ch.MaxSize:]
	}
}

func (ch *ConversationHistory) GetMessages() []DeepSeekMessage {
	return ch.Messages
}

func (ch *ConversationHistory) Clear() {
	ch.Messages = make([]DeepSeekMessage, 0)
}

// DeepSeek提供者
type DeepSeekProvider struct {
	config DeepSeekConfig
	client *http.Client
	// 为每个用户/会话维护对话历史
	conversations map[string]*ConversationHistory
}

// 创建DeepSeek提供者
func NewDeepSeekProvider() *DeepSeekProvider {
	// 使用配置管理系统
	config := GetAIConfig().AI.DeepSeek

	return &DeepSeekProvider{
		config: DeepSeekConfig{
			APIKey:      config.APIKey,
			BaseURL:     config.BaseURL,
			Model:       config.Model,
			MaxTokens:   config.MaxTokens,
			Temperature: config.Temperature,
		},
		client: &http.Client{
			Timeout: 60 * time.Second, // DeepSeek可能需要更长时间
		},
		conversations: make(map[string]*ConversationHistory),
	}
}

func (d *DeepSeekProvider) ContinueWriting(ctx context.Context, prompt string) (string, error) {
	return d.callAI(ctx, "续写", prompt, "")
}

func (d *DeepSeekProvider) PolishText(ctx context.Context, text string) (string, error) {
	return d.callAI(ctx, "润色", text, "")
}

func (d *DeepSeekProvider) SummarizeText(ctx context.Context, text string) (string, error) {
	return d.callAI(ctx, "总结", text, "")
}

// 新增：流式调用AI接口
func (d *DeepSeekProvider) CallAIStream(ctx context.Context, function, content, sessionID string, writer io.Writer) error {
	return d.callAIStream(ctx, function, content, sessionID, writer)
}

// 新增：多轮对话接口
func (d *DeepSeekProvider) Chat(ctx context.Context, message, sessionID string) (string, error) {
	return d.chat(ctx, message, sessionID)
}

func (d *DeepSeekProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		Name:        "deepseek-chat",
		DisplayName: "DeepSeek Chat",
		Provider:    "deepseek",
		Description: "DeepSeek开源大语言模型，支持中英文，推理能力强，适合创意写作和内容优化",
		MaxTokens:   4000,
		IsAvailable: d.IsAvailable(),
	}
}

func (d *DeepSeekProvider) IsAvailable() bool {
	return d.config.APIKey != ""
}

// 获取或创建对话历史
func (d *DeepSeekProvider) getConversation(sessionID string) *ConversationHistory {
	if conversation, exists := d.conversations[sessionID]; exists {
		return conversation
	}

	conversation := NewConversationHistory(20) // 最多保存20轮对话
	d.conversations[sessionID] = conversation
	return conversation
}

// 多轮对话实现
func (d *DeepSeekProvider) chat(ctx context.Context, message, sessionID string) (string, error) {
	if !d.IsAvailable() {
		return "", fmt.Errorf("DeepSeek API未配置")
	}

	conversation := d.getConversation(sessionID)

	// 添加用户消息到对话历史
	conversation.AddMessage("user", message)

	// 构建请求，包含完整的对话历史
	request := DeepSeekRequest{
		Model:            d.config.Model,
		Messages:         conversation.GetMessages(),
		MaxTokens:        d.config.MaxTokens,
		Temperature:      d.config.Temperature,
		TopP:             1.0,
		Stream:           false,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
		ResponseFormat: &ResponseFormat{
			Type: "text",
		},
		Stop:        nil,
		ToolChoice:  "none",
		Logprobs:    false,
		TopLogprobs: nil,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.config.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+d.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API请求失败: %d - %s", resp.StatusCode, string(body))
	}

	var response DeepSeekResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("API返回结果为空")
	}

	// 添加AI回复到对话历史
	aiResponse := response.Choices[0].Message.Content
	conversation.AddMessage("assistant", aiResponse)

	return aiResponse, nil
}

// 流式调用AI
func (d *DeepSeekProvider) callAIStream(ctx context.Context, function, content, sessionID string, writer io.Writer) error {
	if !d.IsAvailable() {
		return fmt.Errorf("DeepSeek API未配置")
	}

	// 构建提示词
	var systemPrompt string
	switch function {
	case "续写":
		systemPrompt = "你是写作助手，请直接续写内容，保持风格一致。"
	case "润色":
		systemPrompt = "你是写作助手，请直接输出润色后的文本。"
	case "总结":
		systemPrompt = "你是写作助手，请直接输出总结内容。"
	default:
		systemPrompt = "你是写作助手，请直接处理文本。"
	}

	// 构建消息数组
	messages := []DeepSeekMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: content},
	}

	// 如果有会话ID，添加对话历史
	if sessionID != "" {
		conversation := d.getConversation(sessionID)
		// 在系统提示词和用户消息之间插入对话历史
		historyMessages := conversation.GetMessages()
		if len(historyMessages) > 0 {
			// 创建新的消息数组，包含系统提示词、对话历史和当前用户消息
			newMessages := make([]DeepSeekMessage, 0, len(historyMessages)+2)
			newMessages = append(newMessages, DeepSeekMessage{Role: "system", Content: systemPrompt})
			newMessages = append(newMessages, historyMessages...)
			newMessages = append(newMessages, DeepSeekMessage{Role: "user", Content: content})
			messages = newMessages
		}
	}

	request := DeepSeekRequest{
		Model:            d.config.Model,
		Messages:         messages,
		MaxTokens:        d.config.MaxTokens,
		Temperature:      d.config.Temperature,
		TopP:             1.0,
		Stream:           true, // 启用流式返回
		FrequencyPenalty: 0,
		PresencePenalty:  0,
		ResponseFormat: &ResponseFormat{
			Type: "text",
		},
		Stop:        nil,
		ToolChoice:  "none",
		Logprobs:    false,
		TopLogprobs: nil,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.config.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+d.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API请求失败: %d - %s", resp.StatusCode, string(body))
	}

	// 处理流式响应
	scanner := bufio.NewScanner(resp.Body)
	var fullResponse strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// 检查是否是数据行
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// 检查是否是结束标记
			if data == "[DONE]" {
				break
			}

			// 解析JSON数据
			var streamResp DeepSeekStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				continue // 忽略解析错误，继续处理下一行
			}

			// 提取内容并写入
			if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
				content := streamResp.Choices[0].Delta.Content
				fullResponse.WriteString(content)

				// 写入到writer（用于流式传输）
				if writer != nil {
					writer.Write([]byte(content))
				}
			}
		}
	}

	// 如果有会话ID，将AI回复添加到对话历史
	if sessionID != "" {
		conversation := d.getConversation(sessionID)
		conversation.AddMessage("assistant", fullResponse.String())
	}

	return nil
}

// 调用DeepSeek API（非流式，保持兼容性）
func (d *DeepSeekProvider) callAI(ctx context.Context, function, content, sessionID string) (string, error) {
	if !d.IsAvailable() {
		return "", fmt.Errorf("DeepSeek API未配置")
	}

	// 构建提示词
	var systemPrompt string
	switch function {
	case "续写":
		systemPrompt = "你是写作助手，请直接续写内容，保持风格一致。"
	case "润色":
		systemPrompt = "你是写作助手，请直接输出润色后的文本。"
	case "总结":
		systemPrompt = "你是写作助手，请直接输出总结内容。"
	default:
		systemPrompt = "你是写作助手，请直接处理文本。"
	}

	// 构建消息数组
	messages := []DeepSeekMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: content},
	}

	// 如果有会话ID，添加对话历史
	if sessionID != "" {
		conversation := d.getConversation(sessionID)
		// 在系统提示词和用户消息之间插入对话历史
		historyMessages := conversation.GetMessages()
		if len(historyMessages) > 0 {
			// 创建新的消息数组，包含系统提示词、对话历史和当前用户消息
			newMessages := make([]DeepSeekMessage, 0, len(historyMessages)+2)
			newMessages = append(newMessages, DeepSeekMessage{Role: "system", Content: systemPrompt})
			newMessages = append(newMessages, historyMessages...)
			newMessages = append(newMessages, DeepSeekMessage{Role: "user", Content: content})
			messages = newMessages
		}
	}

	request := DeepSeekRequest{
		Model:            d.config.Model,
		Messages:         messages,
		MaxTokens:        d.config.MaxTokens,
		Temperature:      d.config.Temperature,
		TopP:             1.0,
		Stream:           false,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
		ResponseFormat: &ResponseFormat{
			Type: "text",
		},
		Stop:        nil,
		ToolChoice:  "none",
		Logprobs:    false,
		TopLogprobs: nil,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.config.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+d.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API请求失败: %d - %s", resp.StatusCode, string(body))
	}

	var response DeepSeekResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("API返回结果为空")
	}

	// 如果有会话ID，将AI回复添加到对话历史
	if sessionID != "" {
		conversation := d.getConversation(sessionID)
		conversation.AddMessage("assistant", response.Choices[0].Message.Content)
	}

	return response.Choices[0].Message.Content, nil
}
