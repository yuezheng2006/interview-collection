# AI模型切换空指针问题修复说明

## 问题描述

在调用 `/api/ai/unified` 接口时出现空指针解引用错误：

```
runtime error: invalid memory address or nil pointer dereference
/Users/chenhuiyuan/Desktop/chy开发/测试题/backend/internal/handler/ai.go:136
```

## 问题分析

### 根本原因

1. **模型名称与Provider名称不匹配**：
   - 前端传递的 `ModelName` 是具体的模型名称（如 "qwen-turbo"）
   - 但后端 `svc.Use()` 方法期望的是Provider名称（如 "tongyi"）
   - 导致 `svc.CurrentProvider()` 返回 `nil`

2. **缺少空指针检查**：
   - 所有AI接口都直接调用 `svc.CurrentProvider().Method()` 
   - 没有检查 `CurrentProvider()` 是否返回 `nil`

### 代码流程问题

```go
// 问题代码：直接使用ModelName作为Provider名称
if req.ModelName != "" && req.ModelName != svc.GetCurrentModel() {
    svc.Use(req.ModelName)  // 这里ModelName可能是"qwen-turbo"，但期望"tongyi"
}

// 问题代码：没有空指针检查
result, err = svc.CurrentProvider().PolishText(context.Background(), prompt)
```

## 修复方案

### 1. 修复模型切换逻辑

在统一AI接口中，根据模型名称找到对应的Provider名称：

```go
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
```

### 2. 添加空指针检查

在所有AI接口中添加Provider可用性检查：

```go
// 检查当前provider是否可用
currentProvider := svc.CurrentProvider()
if currentProvider == nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "当前AI模型不可用，请先切换模型"})
    return
}

// 使用已检查的provider
result, err = currentProvider.PolishText(context.Background(), prompt)
```

### 3. 修复的接口

- `/api/ai/unified` - 统一AI接口
- `/api/ai/continue` - 续写接口
- `/api/ai/polish` - 润色接口  
- `/api/ai/summarize` - 总结接口

## 修复后的优势

1. **防止空指针错误**：所有AI接口都有适当的空指针检查
2. **清晰的错误提示**：当模型不可用时，返回明确的错误信息
3. **逻辑一致性**：模型名称和Provider名称的映射关系清晰
4. **代码健壮性**：避免了运行时panic，提升了系统稳定性

## 测试建议

1. **测试模型切换**：确保切换模型后能正常使用
2. **测试错误处理**：验证当模型不可用时的错误提示
3. **测试接口兼容性**：确保所有AI接口都能正常工作

## 注意事项

1. **前端传递的ModelName**：应该是具体的模型名称（如 "qwen-turbo"）
2. **后端Provider名称**：用于内部标识（如 "tongyi"）
3. **错误处理**：所有AI接口都应该有适当的错误处理机制 