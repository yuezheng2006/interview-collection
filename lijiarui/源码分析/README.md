# AI代码审查插件源码分析

## 项目结构分析

基于对压缩代码的逆向分析，该Chrome扩展的源码结构如下：

### ✅ 可编译源码结构
```
ai-code-review-to-github/
├── package.json              # 项目配置和依赖管理
├── .eslintrc.js             # 代码质量检查配置
├── src/                     # 源码目录
│   ├── manifest.json        # Chrome扩展配置
│   ├── content.js           # 内容脚本（507行）
│   ├── background.js        # 后台服务脚本（275行）
│   ├── popup.html           # 弹窗界面
│   ├── popup.js             # 弹窗脚本（278行）
│   ├── styles.css           # 样式文件
│   ├── config/
│   │   └── ai-config.js     # AI服务配置（完整）
│   └── services/
│       └── ai-service.js    # AI服务客户端（完整）
├── dist/                    # 构建输出目录
└── releases/                # 打包发布目录
```

### 🔧 构建工具链
- **Parcel 2.9.3**: 现代化零配置构建工具
- **ESLint**: 代码质量检查
- **web-ext**: Chrome扩展打包工具

### 核心模块详情
1. **GitHubAICodeReview类** - 主控制器（content.js）
2. **GitHubAPIClient类** - GitHub API客户端
3. **TechStackDetector类** - 技术栈检测器
4. **AIService类** - AI服务客户端（新增）
5. **PopupController类** - 弹窗控制器

## 功能特性分析

### 1. GitHub页面检测
- 自动检测GitHub仓库页面
- 支持Pull Request和Commit页面
- 页面变化监听和动态注入

### 2. 技术栈分析
- 自动解析项目配置文件
- 支持多种技术栈（Node.js、Python、Java等）
- 依赖关系可视化

### 3. AI代码审查 ✨
- **完整AI配置**: API地址 `http://115.190.121.185:3000`
- **多模型支持**: GPT-3.5 Turbo、GPT-4、Claude-3
- **流式响应**: Server-Sent Events实时返回
- **专业提示词**: 中文代码审查模板
- **智能分析**: 代码质量、安全性、性能优化
- **错误处理**: 完善的重试和降级机制

### 4. 用户界面
- 响应式设计
- 深色模式支持
- 与GitHub界面无缝集成

## 🚀 编译构建

### 快速开始
```bash
# 安装依赖
npm install

# 开发模式
npm run dev

# 生产构建
npm run build

# 打包扩展
npm run pack
```

### 编译能力
- ✅ **完全可编译**: 包含所有必要的配置文件
- ✅ **现代工具链**: Parcel + ESLint + web-ext
- ✅ **依赖完整**: CodeMirror + Marked
- ✅ **AI集成**: 完整的服务配置和客户端
