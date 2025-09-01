# AI写作助手

一个功能强大的AI写作助手，集成了多个大语言模型，提供智能写作辅助功能。

## 🚀 功能特性

### 基础功能
- 📝 文档管理：新建、删除、重命名文档
- ✏️ 编辑器：支持长文本编辑，自动保存
- 📊 字数统计：实时显示字符数
- 💾 本地存储：浏览器自动保存

### AI功能
- 🤖 **多模型支持**：Mock、通义千问、**DeepSeek**、文心一言、智谱AI
- 🔄 模型切换：动态切换不同的AI模型
- ✨ 续写功能：基于上下文智能续写内容
- 🎨 润色功能：优化文本表达，提升质量
- 📋 总结功能：提取核心要点和主要观点
- 🚀 生成功能：按要求生成新内容

### 用户体验
- 🎯 智能选择：自动识别文本选择范围
- 💡 快捷操作：一键应用AI结果
- 📱 响应式设计：支持各种设备
- 🌐 国际化：支持中英文界面

## 🏗️ 技术架构

### 后端技术栈
- **语言**: Go 1.24.2+
- **框架**: Gin Web框架
- **认证**: JWT认证
- **AI集成**: 多模型Provider架构

### 前端技术栈
- **框架**: Vue 3 + TypeScript
- **UI库**: Element Plus
- **状态管理**: Pinia
- **构建工具**: Vite

## 🤖 支持的AI模型

### 1. Mock模型 (默认)
- 类型: 测试模型
- 用途: 开发和测试阶段使用
- 特点: 无需API密钥，返回模拟结果

### 2. 通义千问 (Tongyi)
- 类型: 阿里云大语言模型
- 模型: qwen-turbo
- 特点: 中文支持优秀，适合中文写作

### 3. DeepSeek (新增)
- 类型: 开源大语言模型
- 模型: deepseek-chat
- 特点: 中英文支持优秀，推理能力强，适合创意写作
- 配置: 需要DEEPSEEK_API_KEY

### 4. 文心一言 (Wenxin)
- 类型: 百度大语言模型
- 模型: am-xpgukxjf6s0r
- 特点: 知识丰富，适合专业内容，**支持16k tokens，特别适合长文本处理**
- 配置: 需要WENXIN_API_KEY

### 5. 智谱AI (Zhipu) - 预留
- 类型: 智谱大语言模型
- 模型: glm-4
- 特点: 推理能力强，适合逻辑分析

## 📦 快速开始

### 环境要求
- Go 1.24.2+
- Node.js 18+
- 现代浏览器

### 1. 克隆项目
```bash
git clone <repository-url>
cd ai-writing-assistant
```

### 2. 后端配置
```bash
cd backend

# 创建环境变量文件
cp .env.example .env

# 编辑环境变量，配置API密钥
# 至少配置一个AI模型的API密钥

# 安装依赖
go mod tidy

# 启动服务
go run cmd/server/main.go
```

### 3. 前端配置
```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

### 4. 访问应用
- 前端: http://localhost:5173
- 后端: http://localhost:8080

## ⚙️ 配置说明

### 环境变量配置

在 `backend/.env` 文件中配置以下环境变量：

```bash
# DeepSeek配置（推荐）
DEEPSEEK_API_KEY=your_deepseek_api_key_here
DEEPSEEK_BASE_URL=https://api.deepseek.com/v1/chat/completions

# 通义千问配置
TONGYI_API_KEY=your_tongyi_api_key_here
TONGYI_BASE_URL=https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation

# 其他模型配置...
```

### 获取API密钥

#### DeepSeek (推荐)
1. 访问 [DeepSeek平台](https://platform.deepseek.com/)
2. 注册账号并开通API服务
3. 在控制台获取API密钥

#### 通义千问
1. 访问 [阿里云通义千问](https://dashscope.aliyun.com/)
2. 注册账号并开通服务
3. 在控制台获取API密钥

## 📚 使用指南

### 1. 创建文档
1. 在首页点击"新建文档"
2. 输入文档标题
3. 进入编辑器开始写作

### 2. 使用AI功能
1. 选择要处理的文本
2. 选择AI功能（润色、总结、续写）
3. 选择AI模型（推荐DeepSeek）
4. 输入具体要求
5. 点击"开始AI处理"
6. 查看结果并选择采纳或拒绝

### 3. 切换AI模型
1. 在AI弹窗中选择"选择AI模型"
2. 从下拉列表中选择可用模型
3. 系统会自动切换到选择的模型

## 🔧 开发指南

### 项目结构
```
├── backend/                 # 后端Go服务
│   ├── cmd/server/         # 服务入口
│   ├── internal/           # 内部包
│   │   ├── handler/        # HTTP处理器
│   │   ├── pkg/ai/         # AI服务包
│   │   └── service/        # 业务逻辑
│   └── configs/            # 配置文件
├── frontend/               # 前端Vue应用
│   ├── src/                # 源代码
│   │   ├── views/          # 页面组件
│   │   ├── services/       # API服务
│   │   └── stores/         # 状态管理
│   └── public/             # 静态资源
└── docs/                   # 文档
```

### 添加新的AI模型
1. 实现 `Provider` 接口
2. 在 `registerAIRoutes` 中注册新提供者
3. 更新配置文件
4. 前端自动识别新模型

## 🧪 测试

### 后端测试
```bash
cd backend
go test ./...
```

### 前端测试
```bash
cd frontend
npm run test:unit
npm run test:e2e
```

## 📖 详细文档

- [AI功能实现说明](AI功能实现说明.md)
- [DeepSeek集成说明](DeepSeek集成说明.md)
- [后端配置说明](backend/README.md)
- [前端使用指南](frontend/README.md)

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 📄 许可证

MIT License

## 🙏 致谢

感谢所有开源项目的贡献者，特别是：
- [DeepSeek](https://platform.deepseek.com/) - 优秀的开源大语言模型
- [Gin](https://github.com/gin-gonic/gin) - Go Web框架
- [Vue](https://vuejs.org/) - 渐进式JavaScript框架
- [Element Plus](https://element-plus.org/) - Vue 3 UI组件库

