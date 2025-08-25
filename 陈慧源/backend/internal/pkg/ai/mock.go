package ai

import (
	"context"
	"io"
)

type MockProvider struct{}

func (m MockProvider) ContinueWriting(ctx context.Context, prompt string) (string, error) {
	return "[mock continue] " + prompt, nil
}

func (m MockProvider) PolishText(ctx context.Context, text string) (string, error) {
	return "[mock polish] " + text, nil
}

func (m MockProvider) SummarizeText(ctx context.Context, text string) (string, error) {
	return "[mock summary] " + text, nil
}

// 新增：流式AI调用接口
func (m MockProvider) CallAIStream(ctx context.Context, function, content, sessionID string, writer io.Writer) error {
	// 模拟流式返回
	mockResponse := "[mock " + function + "] " + content
	writer.Write([]byte(mockResponse))
	return nil
}

// 新增：多轮对话接口
func (m MockProvider) Chat(ctx context.Context, message, sessionID string) (string, error) {
	return "[mock chat] " + message, nil
}

func (m MockProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		Name:        "mock",
		DisplayName: "Mock AI模型",
		Provider:    "mock",
		Description: "用于测试的模拟AI模型",
		MaxTokens:   1000,
		IsAvailable: true,
	}
}

func (m MockProvider) IsAvailable() bool {
	return true
}
