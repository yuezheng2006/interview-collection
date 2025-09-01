# AI代码审查插件源码分析

## 项目结构分析

基于对压缩代码的逆向分析，该Chrome扩展的源码结构如下：

### 核心文件
- `manifest.json` - 扩展配置文件
- `popup.html` - 弹窗界面
- `popup.js` - 弹窗逻辑
- `content.js` - 内容脚本（注入GitHub页面）
- `background.js` - 后台服务脚本

### 主要模块
1. **GitHub API客户端** (`lib/github-api.js`)
2. **技术栈检测器** (`lib/tech-stack-detector.js`)
3. **代码差异查看器** (`lib/codemirror-diff.js`)
4. **AI摘要组件** (`lib/ai-summary-component.js`)
5. **React渲染器** (`lib/react-renderer.js`)

## 功能特性分析

### 1. GitHub页面检测
- 自动检测GitHub仓库页面
- 支持Pull Request和Commit页面
- 页面变化监听和动态注入

### 2. 技术栈分析
- 自动解析项目配置文件
- 支持多种技术栈（Node.js、Python、Java等）
- 依赖关系可视化

### 3. AI代码审查
- 集成外部AI服务
- 流式响应处理
- 智能代码问题识别

### 4. 用户界面
- 响应式设计
- 深色模式支持
- 与GitHub界面无缝集成
