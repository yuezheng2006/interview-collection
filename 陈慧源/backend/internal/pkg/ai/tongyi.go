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

// 通义千问配置
type TongyiConfig struct {
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
	Temperature float64
}

// 通义千问请求结构
type TongyiRequest struct {
	Model      string     `json:"model"`
	Input      Input      `json:"input"`
	Parameters Parameters `json:"parameters"`
}

type Input struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Parameters struct {
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
}

// 通义千问响应结构
type TongyiResponse struct {
	Output Output `json:"output"`
	Usage  Usage  `json:"usage"`
}

type Output struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

type Usage struct {
	TotalTokens int `json:"total_tokens"`
}

// 通义千问提供者
type TongyiProvider struct {
	config TongyiConfig
	client *http.Client
}

// 创建通义千问提供者
func NewTongyiProvider() *TongyiProvider {
	// 使用配置管理系统
	config := GetAIConfig().AI.Tongyi

	return &TongyiProvider{
		config: TongyiConfig{
			APIKey:      config.APIKey,
			BaseURL:     config.BaseURL,
			Model:       config.Model,
			MaxTokens:   config.MaxTokens,
			Temperature: config.Temperature,
		},
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *TongyiProvider) ContinueWriting(ctx context.Context, prompt string) (string, error) {
	return t.callAI(ctx, "续写", prompt)
}

func (t *TongyiProvider) PolishText(ctx context.Context, text string) (string, error) {
	return t.callAI(ctx, "润色", text)
}

func (t *TongyiProvider) SummarizeText(ctx context.Context, text string) (string, error) {
	return t.callAI(ctx, "总结", text)
}

// 新增：流式AI调用接口
func (t *TongyiProvider) CallAIStream(ctx context.Context, function, content, sessionID string, writer io.Writer) error {
	// 通义千问暂不支持流式返回，使用非流式调用
	result, err := t.callAI(ctx, function, content)
	if err != nil {
		return err
	}
	writer.Write([]byte(result))
	return nil
}

// 新增：多轮对话接口
func (t *TongyiProvider) Chat(ctx context.Context, message, sessionID string) (string, error) {
	// 通义千问暂不支持多轮对话，使用单轮调用
	return t.callAI(ctx, "对话", message)
}

func (t *TongyiProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		Name:        "qwen-turbo",
		DisplayName: "通义千问 Turbo",
		Provider:    "tongyi",
		Description: "阿里云通义千问大语言模型，支持中文创作和优化",
		MaxTokens:   2000,
		IsAvailable: t.IsAvailable(),
	}
}

func (t *TongyiProvider) IsAvailable() bool {
	return t.config.APIKey != ""
}

// 调用通义千问API
func (t *TongyiProvider) callAI(ctx context.Context, function, content string) (string, error) {
	if !t.IsAvailable() {
		return "", fmt.Errorf("通义千问API未配置")
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

	request := TongyiRequest{
		Model: t.config.Model,
		Input: Input{
			Messages: []Message{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: content},
			},
		},
		Parameters: Parameters{
			MaxTokens:   t.config.MaxTokens,
			Temperature: t.config.Temperature,
			TopP:        0.8,
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.config.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+t.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
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

	var response TongyiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(response.Output.Choices) == 0 {
		return "", fmt.Errorf("API返回结果为空")
	}

	return response.Output.Choices[0].Message.Content, nil
}
