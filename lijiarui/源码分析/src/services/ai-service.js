/**
 * AI服务客户端
 * 处理与AI API的交互
 */

import { getAIConfig, getApiUrl, getPromptTemplate, formatPrompt } from '../config/ai-config.js';

export class AIService {
  constructor() {
    this.config = getAIConfig();
    this.abortController = null;
  }

  /**
   * 发送代码审查请求
   */
  async reviewCode(codeChanges, repoInfo) {
    const template = getPromptTemplate('CODE_REVIEW');
    const prompt = formatPrompt(template.user, {
      owner: repoInfo.owner,
      repo: repoInfo.repo,
      branch: repoInfo.branch || 'main',
      commit: repoInfo.commit || '',
      codeChanges: this.formatCodeChanges(codeChanges)
    });

    return this.sendRequest('/api/review', {
      system: template.system,
      prompt: prompt,
      model: await this.getSelectedModel()
    });
  }

  /**
   * 生成技术栈摘要
   */
  async generateTechStackSummary(techStacks, projectName) {
    const template = getPromptTemplate('TECH_STACK_SUMMARY');
    const configFiles = techStacks.map(ts => `${ts.type}: ${ts.metadata.filePath}`).join('\n');
    
    const prompt = formatPrompt(template.user, {
      projectName: projectName,
      configFiles: configFiles
    });

    return this.sendRequest('/api/summary', {
      system: template.system,
      prompt: prompt,
      model: await this.getSelectedModel()
    });
  }

  /**
   * 生成Commit摘要
   */
  async generateCommitSummary(commitData) {
    const template = getPromptTemplate('COMMIT_SUMMARY');
    const prompt = formatPrompt(template.user, {
      commitSha: commitData.sha,
      author: commitData.author,
      date: commitData.date,
      changedFiles: commitData.files.map(f => `${f.status}: ${f.filename}`).join('\n'),
      diff: this.formatDiff(commitData.diff)
    });

    return this.sendRequest('/api/summary', {
      system: template.system,
      prompt: prompt,
      model: await this.getSelectedModel()
    });
  }

  /**
   * 发送流式聊天请求
   */
  async sendStreamRequest(message, onChunk, onError, onComplete) {
    try {
      const port = chrome.runtime.connect({ name: 'stream-chat' });
      
      port.onMessage.addListener((response) => {
        switch (response.type) {
          case 'chunk':
            onChunk(response.chunk);
            break;
          case 'error':
            onError(new Error(response.error));
            break;
          case 'end':
            onComplete();
            port.disconnect();
            break;
        }
      });

      port.postMessage({
        type: 'start-stream',
        data: {
          url: getApiUrl('/api/stream'),
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            message: message,
            model: await this.getSelectedModel()
          })
        }
      });

      return port;
    } catch (error) {
      console.error('Stream request failed:', error);
      onError(error);
    }
  }

  /**
   * 发送普通API请求
   */
  async sendRequest(endpoint, data) {
    try {
      const response = await chrome.runtime.sendMessage({
        type: 'API_REQUEST',
        data: {
          url: getApiUrl(endpoint),
          method: 'POST',
          headers: this.config.REQUEST_CONFIG.headers,
          body: JSON.stringify(data)
        }
      });

      if (!response.ok) {
        throw new Error(this.getErrorMessage(response.status, response.error));
      }

      return response.data;
    } catch (error) {
      console.error('API request failed:', error);
      throw error;
    }
  }

  /**
   * 获取用户选择的AI模型
   */
  async getSelectedModel() {
    return new Promise((resolve) => {
      chrome.storage.sync.get(['aiModel'], (result) => {
        resolve(result.aiModel || this.config.DEFAULT_MODEL);
      });
    });
  }

  /**
   * 格式化代码变更
   */
  formatCodeChanges(changes) {
    if (Array.isArray(changes)) {
      return changes.map(change => {
        return `文件: ${change.filename}
状态: ${change.status}
变更: +${change.additions} -${change.deletions}

${change.patch || ''}`;
      }).join('\n\n');
    }
    return changes.toString();
  }

  /**
   * 格式化diff内容
   */
  formatDiff(diff) {
    if (typeof diff === 'string') {
      return diff.substring(0, 2000); // 限制长度
    }
    return JSON.stringify(diff).substring(0, 2000);
  }

  /**
   * 获取错误消息
   */
  getErrorMessage(status, error) {
    switch (status) {
      case 0:
        return this.config.ERROR_MESSAGES.NETWORK_ERROR;
      case 401:
        return this.config.ERROR_MESSAGES.UNAUTHORIZED;
      case 429:
        return this.config.ERROR_MESSAGES.RATE_LIMIT;
      case 408:
        return this.config.ERROR_MESSAGES.TIMEOUT_ERROR;
      default:
        return error || this.config.ERROR_MESSAGES.API_ERROR;
    }
  }

  /**
   * 中止当前请求
   */
  abort() {
    if (this.abortController) {
      this.abortController.abort();
    }
  }
}
