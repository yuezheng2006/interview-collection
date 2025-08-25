package ai

import (
	"context"
	"io"
)

// AI模型信息
type ModelInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Provider    string `json:"provider"`
	Description string `json:"description"`
	MaxTokens   int    `json:"maxTokens"`
	IsAvailable bool   `json:"isAvailable"`
}

// AI提供者接口
type Provider interface {
	ContinueWriting(ctx context.Context, prompt string) (string, error)
	PolishText(ctx context.Context, text string) (string, error)
	SummarizeText(ctx context.Context, text string) (string, error)

	// 新增：流式AI调用接口
	CallAIStream(ctx context.Context, function, content, sessionID string, writer io.Writer) error

	// 新增：多轮对话接口
	Chat(ctx context.Context, message, sessionID string) (string, error)

	// 获取模型信息
	GetModelInfo() ModelInfo
	// 检查模型是否可用
	IsAvailable() bool
}

// AI服务
type Service struct {
	providers map[string]Provider
	current   string
	models    []ModelInfo
}

func NewService() *Service {
	return &Service{
		providers: make(map[string]Provider),
		models:    make([]ModelInfo, 0),
	}
}

func (s *Service) Register(name string, p Provider) {
	s.providers[name] = p
	s.models = append(s.models, p.GetModelInfo())
}

func (s *Service) Use(name string) { s.current = name }

func (s *Service) GetAvailableModels() []ModelInfo {
	return s.models
}

func (s *Service) GetCurrentModel() string {
	return s.current
}

func (s *Service) CurrentProvider() Provider {
	if p, ok := s.providers[s.current]; ok {
		return p
	}
	return nil
}
