/**
 * AI服务配置
 * 基于逆向分析发现的AI集成配置
 */

export const AI_CONFIG = {
  // API基础配置
  BASE_URL: 'http://115.190.121.185:3000',
  
  // API端点
  ENDPOINTS: {
    CHAT: '/api/chat',
    STREAM: '/api/stream',
    REVIEW: '/api/review',
    SUMMARY: '/api/summary'
  },
  
  // 支持的AI模型
  MODELS: {
    GPT_3_5_TURBO: 'gpt-3.5-turbo',
    GPT_4: 'gpt-4',
    CLAUDE_3: 'claude-3'
  },
  
  // 默认模型
  DEFAULT_MODEL: 'gpt-3.5-turbo',
  
  // 请求配置
  REQUEST_CONFIG: {
    timeout: 30000,
    retries: 3,
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json'
    }
  },
  
  // 流式响应配置
  STREAM_CONFIG: {
    reconnectAttempts: 3,
    reconnectDelay: 1000,
    maxMessageSize: 1024 * 1024 // 1MB
  },
  
  // AI提示词配置
  PROMPTS: {
    CODE_REVIEW: {
      system: `你是一个专业的代码审查助手。请分析提供的代码变更，并提供以下方面的反馈：
1. 代码质量和最佳实践
2. 潜在的错误和安全问题
3. 性能优化建议
4. 可读性和维护性改进建议

请用中文回复，保持专业和建设性的语气。`,
      
      user: `请审查以下代码变更：

仓库: {owner}/{repo}
分支: {branch}
提交: {commit}

代码变更：
{codeChanges}

请提供详细的代码审查意见。`
    },
    
    TECH_STACK_SUMMARY: {
      system: `你是一个技术栈分析专家。请基于项目配置文件分析技术栈，并提供简洁的总结。`,
      
      user: `请分析以下项目的技术栈：

项目: {projectName}
配置文件: {configFiles}

请提供技术栈摘要和建议。`
    },
    
    COMMIT_SUMMARY: {
      system: `你是一个代码变更分析专家。请分析commit的变更内容并生成简洁的摘要。`,
      
      user: `请分析以下commit变更：

Commit: {commitSha}
作者: {author}
时间: {date}

变更文件：
{changedFiles}

代码差异：
{diff}

请生成变更摘要和影响分析。`
    }
  },
  
  // 错误处理配置
  ERROR_MESSAGES: {
    NETWORK_ERROR: '网络连接失败，请检查网络设置',
    API_ERROR: 'AI服务响应错误，请稍后重试',
    TIMEOUT_ERROR: '请求超时，请稍后重试',
    RATE_LIMIT: 'API调用频率过高，请稍后重试',
    UNAUTHORIZED: '未授权访问，请检查API密钥'
  }
};

/**
 * 获取AI配置
 */
export function getAIConfig() {
  return AI_CONFIG;
}

/**
 * 获取API URL
 */
export function getApiUrl(endpoint) {
  return `${AI_CONFIG.BASE_URL}${AI_CONFIG.ENDPOINTS[endpoint] || endpoint}`;
}

/**
 * 获取提示词模板
 */
export function getPromptTemplate(type) {
  return AI_CONFIG.PROMPTS[type];
}

/**
 * 替换提示词中的变量
 */
export function formatPrompt(template, variables) {
  let formatted = template;
  Object.keys(variables).forEach(key => {
    const placeholder = `{${key}}`;
    formatted = formatted.replace(new RegExp(placeholder, 'g'), variables[key]);
  });
  return formatted;
}
