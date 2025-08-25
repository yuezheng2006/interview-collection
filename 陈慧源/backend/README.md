# AI写作助手 - 后端服务

## 环境配置

### 1. 环境变量配置

创建 `.env` 文件并配置以下环境变量：

```bash
# 通义千问配置
TONGYI_API_KEY=your_tongyi_api_key_here
TONGYI_BASE_URL=https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation

# DeepSeek配置
DEEPSEEK_API_KEY=your_deepseek_api_key_here
DEEPSEEK_BASE_URL=https://api.deepseek.com/v1/chat/completions

# 文心一言配置（预留）
WENXIN_API_KEY=your_wenxin_api_key_here
WENXIN_BASE_URL=https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat

# 智谱AI配置（预留）
ZHIPU_API_KEY=your_zhipu_api_key_here
ZHIPU_BASE_URL=https://open.bigmodel.cn/api/paas/v4/chat/completions

# JWT密钥（生产环境请修改）
JWT_SECRET=DEV_SECRET_CHANGE_ME_IN_PRODUCTION

# 服务器配置
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
```

### 2. 获取API密钥

#### 通义千问
1. 访问 [阿里云通义千问](https://dashscope.aliyun.com/)
2. 注册账号并开通服务
3. 在控制台获取API密钥

#### DeepSeek
1. 访问 [DeepSeek官网](https://platform.deepseek.com/)
2. 注册账号并开通API服务
3. 在控制台获取API密钥
4. 支持多种模型：deepseek-chat、deepseek-coder等

#### 文心一言
1. 访问 [百度智能云](https://cloud.baidu.com/)
2. 开通文心一言服务
3. 获取API密钥
4. 支持16k tokens，适合长文本处理

#### 智谱AI（预留）
1. 访问 [智谱AI](https://open.bigmodel.cn/)
2. 注册账号并开通服务
3. 获取API密钥

## 运行服务

```bash
# 安装依赖
go mod tidy

# 运行服务
go run cmd/server/main.go
```

## API接口

### AI相关接口

- `GET /api/ai/models` - 获取可用AI模型列表
- `POST /api/ai/switch-model` - 切换AI模型
- `POST /api/ai/unified` - 统一AI接口（支持续写、润色、总结、生成）

### 文档管理接口

- `GET /api/documents` - 获取文档列表
- `POST /api/documents` - 创建新文档
- `PUT /api/documents/:id` - 更新文档
- `DELETE /api/documents/:id` - 删除文档

### 用户认证接口

- `POST /api/auth/register` - 用户注册
- `POST /api/auth/login` - 用户登录

## 架构说明

### AI服务架构

- `Provider` 接口：定义AI模型的标准接口
- `Service` 结构：管理多个AI提供者，支持动态切换
- 支持多模型：Mock、通义千问、DeepSeek、文心一言、智谱AI

### 中间件

- CORS支持：允许跨域请求
- JWT认证：保护API接口
- 日志记录：记录请求和错误信息

## 开发说明

### 添加新的AI模型

1. 实现 `Provider` 接口
2. 在 `registerAIRoutes` 中注册新提供者
3. 更新配置文件

### 错误处理

所有API接口都包含统一的错误处理，返回标准化的错误信息。

## 注意事项

1. 生产环境请修改JWT密钥
2. API密钥请妥善保管，不要提交到版本控制系统
3. 建议使用环境变量或配置文件管理敏感信息 