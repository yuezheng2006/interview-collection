# 文心一言API参数修复说明

## 问题描述

在调用文心一言API时，出现以下错误：
```
"AI处理失败: API请求失败: 400 - {\"error\":{\"code\":\"invalid_argument\",\"message\":\"you must provide a model parameter\",\"type\":\"invalid_request_error\"},\"id\":\"as-m7cjbbzjg8\"}"
```

## 问题分析

根据百度千帆API文档（https://cloud.baidu.com/doc/qianfan-api/s/3m7of64lb），文心一言API的请求体必须包含 `model` 参数。

错误信息明确指出：`"you must provide a model parameter"`，说明请求中缺少必需的 `model` 字段。

## 修复内容

### 1. 修复请求结构体

在 `backend/internal/pkg/ai/wenxin.go` 文件中，为 `WenxinRequest` 结构体添加了 `Model` 字段：

```go
// 修复前
type WenxinRequest struct {
	Messages    []WenxinMessage `json:"messages"`
	Stream      bool            `json:"stream"`
	Temperature float64         `json:"temperature"`
	TopP        float64         `json:"top_p"`
	MaxTokens   int             `json:"max_tokens"`
}

// 修复后
type WenxinRequest struct {
	Model       string          `json:"model"`        // 新增：必需的模型参数
	Messages    []WenxinMessage `json:"messages"`
	Stream      bool            `json:"stream"`
	Temperature float64         `json:"temperature"`
	TopP        float64         `json:"top_p"`
	MaxTokens   int             `json:"max_tokens"`
}
```

### 2. 修复请求构建

在 `callAI` 方法中，确保请求包含 `Model` 参数：

```go
request := WenxinRequest{
	Model: w.config.Model,  // 新增：使用配置中的模型名称
	Messages: []WenxinMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: content},
	},
	Stream:      false,
	Temperature: w.config.Temperature,
	TopP:        0.8,
	MaxTokens:   w.config.MaxTokens,
}
```

## 配置说明

当前配置文件中，文心一言的模型配置为：
```yaml
wenxin:
  api_key: "bce-v3/ALTAK-..."
  base_url: "https://qianfan.baidubce.com/v2/chat/completions"
  model: "am-xpgukxjf6s0r"  # 这个值会作为model参数传递给API
  max_tokens: 16384
  temperature: 0.7
```

## 验证方法

修复后，可以通过以下方式验证：

1. 重启后端服务
2. 调用AI接口，选择文心一言模型
3. 检查是否还有400错误
4. 查看API响应是否正常

## 相关文档

- [百度千帆API文档](https://cloud.baidu.com/doc/qianfan-api/s/3m7of64lb)
- [文心一言集成说明](./文心一言集成说明.md)

## 注意事项

1. 确保配置文件中的 `model` 字段值正确
2. 不同AI提供者的API格式可能不同，需要分别处理
3. 通义千问和DeepSeek的实现已经正确包含了model参数 