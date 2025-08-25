package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// AI配置结构
type AIConfig struct {
	Server struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	AI struct {
		DefaultModel string      `yaml:"default_model"`
		Tongyi       ModelConfig `yaml:"tongyi"`
		DeepSeek     ModelConfig `yaml:"deepseek"`
		Wenxin       ModelConfig `yaml:"wenxin"`
		Zhipu        ModelConfig `yaml:"zhipu"`
	} `yaml:"ai"`
	Database struct {
		Driver string `yaml:"driver"`
		DSN    string `yaml:"dsn"`
	} `yaml:"database"`
	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
		Output string `yaml:"output"`
	} `yaml:"logging"`
}

// 模型配置结构
type ModelConfig struct {
	APIKey      string  `yaml:"api_key"`
	BaseURL     string  `yaml:"base_url"`
	Model       string  `yaml:"model"`
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float64 `yaml:"temperature"`
}

// 获取AI配置
func GetAIConfig() *AIConfig {
	// 从YAML配置文件读取
	config := loadYAMLConfig()

	// 如果YAML中没有配置，使用默认值
	if config == nil {
		fmt.Println("警告: 无法加载配置文件，使用默认配置")
		config = getDefaultConfig()
	}

	// 环境变量覆盖（优先级更高）
	overrideWithEnvVars(config)

	return config
}

// 从YAML文件加载配置
func loadYAMLConfig() *AIConfig {
	// 尝试多个可能的配置文件路径
	configPaths := []string{
		"configs/config.yaml",       // 相对于当前工作目录
		"../configs/config.yaml",    // 相对于当前工作目录的上级
		"../../configs/config.yaml", // 相对于当前工作目录的上上级
		"./configs/config.yaml",     // 明确当前目录
	}

	var configData []byte
	var configPath string
	var err error

	// 尝试读取配置文件
	for _, path := range configPaths {
		configData, err = os.ReadFile(path)
		if err == nil {
			configPath = path
			break
		}
	}

	// 如果所有路径都失败，尝试从可执行文件所在目录查找
	if configData == nil {
		execPath, execErr := os.Executable()
		if execErr == nil {
			execDir := filepath.Dir(execPath)
			execConfigPath := filepath.Join(execDir, "configs", "config.yaml")
			configData, err = os.ReadFile(execConfigPath)
			if err == nil {
				configPath = execConfigPath
			}
		}
	}

	// 如果仍然找不到配置文件，返回nil
	if configData == nil {
		fmt.Printf("错误: 无法找到配置文件，尝试过的路径: %v\n", configPaths)
		return nil
	}

	fmt.Printf("成功加载配置文件: %s\n", configPath)

	// 替换环境变量占位符
	configData = replaceEnvPlaceholders(configData)

	var config AIConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		fmt.Printf("错误: 解析配置文件失败: %v\n", err)
		fmt.Printf("配置文件内容:\n%s\n", string(configData))
		return nil
	}

	return &config
}

// 替换环境变量占位符
func replaceEnvPlaceholders(data []byte) []byte {
	content := string(data)

	// 替换常见的环境变量占位符
	replacements := map[string]string{
		"${TONGYI_API_KEY}":    os.Getenv("TONGYI_API_KEY"),
		"${TONGYI_BASE_URL}":   os.Getenv("TONGYI_BASE_URL"),
		"${DEEPSEEK_API_KEY}":  os.Getenv("DEEPSEEK_API_KEY"),
		"${DEEPSEEK_BASE_URL}": os.Getenv("DEEPSEEK_BASE_URL"),
		"${WENXIN_API_KEY}":    os.Getenv("WENXIN_API_KEY"),
		"${WENXIN_BASE_URL}":   os.Getenv("WENXIN_BASE_URL"),
		"${ZHIPU_API_KEY}":     os.Getenv("ZHIPU_API_KEY"),
		"${ZHIPU_BASE_URL}":    os.Getenv("ZHIPU_BASE_URL"),
	}

	for placeholder, value := range replacements {
		if value == "" {
			// 如果环境变量为空，使用空字符串（不带引号）
			content = strings.ReplaceAll(content, placeholder, "")
		} else {
			content = strings.ReplaceAll(content, placeholder, value)
		}
	}

	return []byte(content)
}

// 获取默认配置
func getDefaultConfig() *AIConfig {
	var config AIConfig
	config.AI.DefaultModel = "mock"
	config.AI.Tongyi = ModelConfig{
		APIKey:      "",
		BaseURL:     "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation",
		Model:       "qwen-turbo",
		MaxTokens:   2000,
		Temperature: 0.7,
	}
	config.AI.DeepSeek = ModelConfig{
		APIKey:      "",
		BaseURL:     "https://api.deepseek.com/v1/chat/completions",
		Model:       "deepseek-chat",
		MaxTokens:   4000,
		Temperature: 0.7,
	}
	config.AI.Wenxin = ModelConfig{
		APIKey:      "",
		BaseURL:     "https://qianfan.baidubce.com/v2/chat/completions",
		Model:       "am-xpgukxjf6s0r",
		MaxTokens:   16384,
		Temperature: 0.7,
	}
	config.AI.Zhipu = ModelConfig{
		APIKey:      "",
		BaseURL:     "https://open.bigmodel.cn/api/paas/v4/chat/completions",
		Model:       "glm-4",
		MaxTokens:   2000,
		Temperature: 0.7,
	}
	return &config
}

// 使用环境变量覆盖配置
func overrideWithEnvVars(config *AIConfig) {
	// 默认模型
	if env := os.Getenv("AI_DEFAULT_MODEL"); env != "" {
		config.AI.DefaultModel = env
	}

	// 通义千问
	if env := os.Getenv("TONGYI_API_KEY"); env != "" {
		config.AI.Tongyi.APIKey = env
	}
	if env := os.Getenv("TONGYI_BASE_URL"); env != "" {
		config.AI.Tongyi.BaseURL = env
	}

	// DeepSeek
	if env := os.Getenv("DEEPSEEK_API_KEY"); env != "" {
		config.AI.DeepSeek.APIKey = env
	}
	if env := os.Getenv("DEEPSEEK_BASE_URL"); env != "" {
		config.AI.DeepSeek.BaseURL = env
	}

	// 文心一言
	if env := os.Getenv("WENXIN_API_KEY"); env != "" {
		config.AI.Wenxin.APIKey = env
	}
	if env := os.Getenv("WENXIN_BASE_URL"); env != "" {
		config.AI.Wenxin.BaseURL = env
	}

	// 智谱AI
	if env := os.Getenv("ZHIPU_API_KEY"); env != "" {
		config.AI.Zhipu.APIKey = env
	}
	if env := os.Getenv("ZHIPU_BASE_URL"); env != "" {
		config.AI.Zhipu.BaseURL = env
	}
}

// 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 获取环境变量作为整数，如果不存在或转换失败则返回默认值
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// 获取环境变量作为浮点数，如果不存在或转换失败则返回默认值
func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
