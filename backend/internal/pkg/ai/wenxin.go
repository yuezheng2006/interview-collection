package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// 文心一言配置
type WenxinConfig struct {
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
	Temperature float64
}

// 文心一言请求结构
type WenxinRequest struct {
	Model       string          `json:"model"`
	Messages    []WenxinMessage `json:"messages"`
	Stream      bool            `json:"stream"`
	Temperature float64         `json:"temperature"`
	TopP        float64         `json:"top_p"`
	MaxTokens   int             `json:"max_tokens"`
}

type WenxinMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// 文心一言响应结构
type WenxinResponse struct {
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

// 文心一言提供者
type WenxinProvider struct {
	config WenxinConfig
	client *http.Client
}

// 创建文心一言提供者
func NewWenxinProvider() *WenxinProvider {
	// 使用配置管理系统
	config := GetAIConfig().AI.Wenxin

	return &WenxinProvider{
		config: WenxinConfig{
			APIKey:      config.APIKey,
			BaseURL:     config.BaseURL,
			Model:       config.Model,
			MaxTokens:   config.MaxTokens,
			Temperature: config.Temperature,
		},
		client: &http.Client{
			Timeout: 60 * time.Second, // 文心一言可能需要更长时间处理长文本
		},
	}
}

func (w *WenxinProvider) ContinueWriting(ctx context.Context, prompt string) (string, error) {
	return w.callAI(ctx, "续写", prompt)
}

func (w *WenxinProvider) PolishText(ctx context.Context, text string) (string, error) {
	return w.callAI(ctx, "润色", text)
}

func (w *WenxinProvider) SummarizeText(ctx context.Context, text string) (string, error) {
	return w.callAI(ctx, "总结", text)
}

// 新增：流式AI调用接口
func (w *WenxinProvider) CallAIStream(ctx context.Context, function, content, sessionID string, writer io.Writer) error {
	// 文心一言暂不支持流式返回，使用非流式调用
	result, err := w.callAI(ctx, function, content)
	if err != nil {
		return err
	}
	writer.Write([]byte(result))
	return nil
}

// 新增：多轮对话接口
func (w *WenxinProvider) Chat(ctx context.Context, message, sessionID string) (string, error) {
	// 文心一言暂不支持多轮对话，使用单轮调用
	return w.callAI(ctx, "对话", message)
}

func (w *WenxinProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		Name:        "am-xpgukxjf6s0r",
		DisplayName: "文心一言 4.0",
		Provider:    "wenxin",
		Description: "百度文心一言大语言模型，支持16k tokens，知识丰富，适合专业内容创作和长文本处理",
		MaxTokens:   16384,
		IsAvailable: w.IsAvailable(),
	}
}

func (w *WenxinProvider) IsAvailable() bool {
	return w.config.APIKey != ""
}

// 调用文心一言API
func (w *WenxinProvider) callAI(ctx context.Context, function, content string) (string, error) {
	if !w.IsAvailable() {
		return "", fmt.Errorf("文心一言API未配置")
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

	request := WenxinRequest{
		Model: w.config.Model,
		Messages: []WenxinMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: content},
		},
		Stream:      false,
		Temperature: w.config.Temperature,
		TopP:        0.8,
		MaxTokens:   w.config.MaxTokens,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", w.config.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+w.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
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

	var response WenxinResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("API返回结果为空")
	}

	return response.Choices[0].Message.Content, nil
}
