# AI写作助手 - 前端应用

## 功能特性

### 基础功能
- 📝 文档管理：新建、删除、重命名文档
- ✏️ 编辑器：支持长文本编辑，自动保存
- 📊 字数统计：实时显示字符数
- 💾 本地存储：浏览器自动保存

### AI功能
- 🤖 多模型支持：Mock、通义千问、文心一言、智谱AI
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

## 技术栈

- **框架**: Vue 3 + TypeScript
- **UI库**: Element Plus
- **状态管理**: Pinia
- **路由**: Vue Router
- **构建工具**: Vite
- **HTTP客户端**: Axios

## 项目结构

```
src/
├── components/          # 公共组件
├── views/              # 页面组件
│   ├── Home.vue        # 首页
│   ├── Editor.vue      # 编辑器
│   ├── Login.vue       # 登录页
│   └── Settings.vue    # 设置页
├── stores/             # 状态管理
│   ├── auth.ts         # 认证状态
│   ├── document.ts     # 文档状态
│   └── counter.ts      # 计数器状态
├── services/           # API服务
│   ├── api.ts          # HTTP客户端
│   ├── ai.ts           # AI服务
│   ├── auth.ts         # 认证服务
│   └── document.ts     # 文档服务
├── types/              # 类型定义
├── utils/              # 工具函数
└── i18n/               # 国际化
```

## 安装和运行

```bash
# 安装依赖
npm install

# 开发模式
npm run dev | cat

# 构建生产版本
npm run build

# 预览构建结果
npm run preview
```

## 环境配置

创建 `.env.local` 文件：

```bash
# API基础URL
VITE_API_BASE_URL=http://localhost:8080/api

# 应用标题
VITE_APP_TITLE=AI写作助手
```

## 使用说明

### 1. 创建文档
1. 在首页点击"新建文档"
2. 输入文档标题
3. 进入编辑器开始写作

### 2. 使用AI功能
1. 选择要处理的文本
2. 选择AI功能（润色、总结、续写）
3. 选择AI模型
4. 输入具体要求
5. 点击"开始AI处理"
6. 查看结果并选择采纳或拒绝

### 3. 切换AI模型
1. 在AI弹窗中选择"选择AI模型"
2. 从下拉列表中选择可用模型
3. 系统会自动切换到选择的模型

### 4. 管理文档
- 自动保存：每5秒自动保存
- 手动保存：点击保存按钮
- 删除文档：在首页点击删除按钮

## 开发指南

### 添加新的AI功能
1. 在 `types/index.ts` 中定义新功能类型
2. 在 `services/ai.ts` 中添加API调用
3. 在 `Editor.vue` 中添加UI组件
4. 在后端添加对应的处理逻辑

### 添加新的AI模型
1. 在后端实现新的 `Provider` 接口
2. 在 `services/ai.ts` 中添加模型信息
3. 更新前端模型选择器

### 样式定制
- 使用CSS变量定义主题色彩
- 遵循Element Plus设计规范
- 响应式设计，支持移动端

## 测试

```bash
# 单元测试
npm run test:unit

# E2E测试
npm run test:e2e

# 测试覆盖率
npm run test:coverage
```

## 部署

### 构建
```bash
npm run build
```

### 部署到静态服务器
将 `dist` 目录部署到任何静态文件服务器。

### 部署到CDN
可以使用Vercel、Netlify等平台进行部署。

## 注意事项

1. 确保后端服务正常运行
2. 配置正确的API密钥
3. 在生产环境中使用HTTPS
4. 定期备份用户数据

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License
