/**
 * GitHub AI代码审查插件 - 内容脚本
 * 负责在GitHub页面中注入AI代码审查功能
 */

class GitHubAICodeReview {
  constructor() {
    this.container = null;
    this.codeChangesContainer = null;
    this.isInjected = false;
    this.isCodeChangesInjected = false;
    this.diffViewer = null;
    this.isRenderingDiff = false;
    this.hasRequestedReview = false;
    this.reviewResponse = null;
    this.summaryContainer = null;
    this.isTechStackLoaded = false;
    this.isCodeChangesLoaded = false;
    this.currentChangesData = null;
    
    this.init();
    this.cleanupExistingAISummary();
  }

  /**
   * 初始化插件
   */
  async init() {
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', () => {
        this.injectTechStack();
      });
    } else {
      setTimeout(() => {
        this.injectTechStack();
      }, 100);
    }
    
    this.observePageChanges();
  }

  /**
   * 监听页面变化
   */
  observePageChanges() {
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (mutation.type === 'childList') {
          const hasRelevantChanges = Array.from(mutation.addedNodes).some(node => 
            node instanceof Element && 
            (node.querySelector('.about-margin') || node.querySelector('.clearfix'))
          );
          
          if (hasRelevantChanges) {
            // 重置状态
            this.isInjected = false;
            this.isCodeChangesInjected = false;
            this.container = null;
            this.codeChangesContainer = null;
            this.hasRequestedReview = false;
            this.reviewResponse = null;
            this.isRenderingDiff = false;
            this.isTechStackLoaded = false;
            this.isCodeChangesLoaded = false;
            
            if (this.summaryContainer) {
              this.summaryContainer.remove();
              this.summaryContainer = null;
            }
            
            // 重新注入
            setTimeout(() => {
              this.injectTechStack();
              this.injectCodeChanges();
            }, 500);
          }
        }
      });
    });
    
    observer.observe(document.body, {
      childList: true,
      subtree: true
    });
  }

  /**
   * 检查是否为仓库页面
   */
  isRepositoryPage() {
    const url = window.location.href;
    const repoPattern = /github\.com\/[^\/]+\/[^\/]+(?:\/(?:tree|blob|commits?|releases|issues|pull|actions|projects|wiki|security|insights|settings)?(?:\/.*)?)?$/;
    return repoPattern.test(url);
  }

  /**
   * 查找目标容器
   */
  findTargetContainer() {
    const selectors = ['.about-margin', '.clearfix'];
    
    for (const selector of selectors) {
      const elements = document.querySelectorAll(selector);
      if (elements.length > 0) {
        return elements[0];
      }
    }
    
    return null;
  }

  /**
   * 注入技术栈分析功能
   */
  async injectTechStack() {
    if (!this.isRepositoryPage() || this.isInjected) {
      return;
    }
    
    const targetContainer = this.findTargetContainer();
    if (!targetContainer) {
      return;
    }
    
    const token = await this.getGitHubToken();
    const repoInfo = this.parseRepositoryFromURL(window.location.href);
    
    if (!repoInfo) {
      return;
    }
    
    const githubClient = new GitHubAPIClient(token);
    const techStackDetector = new TechStackDetector(githubClient);
    
    this.createTechStackContainer(targetContainer);
    this.isInjected = true;
    
    this.loadAndDisplayTechStack(token, repoInfo.owner, repoInfo.repo);
  }

  /**
   * 创建技术栈容器
   */
  createTechStackContainer(targetContainer) {
    this.container = document.createElement('div');
    this.container.id = 'github-tech-stack-container';
    this.container.className = 'github-tech-stack-container';
    
    this.container.innerHTML = `
      <div class="tech-stack-header">
        <h3 class="tech-stack-title">
          <svg class="tech-stack-icon" width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M1.5 8a6.5 6.5 0 1113 0 6.5 6.5 0 01-13 0zM8 0a8 8 0 100 16A8 8 0 008 0zm.5 4.75a.75.75 0 00-1.5 0v3.5a.75.75 0 00.471.696l2.5 1a.75.75 0 00.557-1.392L8.5 7.742V4.75z"/>
          </svg>
          技术栈
        </h3>
        <div class="tech-stack-loading">检测中...</div>
      </div>
      <div class="tech-stack-content">
        <div class="tech-stack-placeholder">正在分析项目配置文件...</div>
      </div>
    `;
    
    if (targetContainer.firstChild) {
      targetContainer.insertBefore(this.container, targetContainer.firstChild);
    } else {
      targetContainer.appendChild(this.container);
    }
  }

  /**
   * 加载并显示技术栈信息
   */
  async loadAndDisplayTechStack(token, owner, repo) {
    if (!this.container) return;
    
    try {
      const githubClient = new GitHubAPIClient(token);
      const techStackDetector = new TechStackDetector(githubClient);
      const techStacks = await techStackDetector.detectTechStack(owner, repo);
      
      this.updateTechStackDisplay(techStacks);
      this.isTechStackLoaded = true;
      this.checkAndCreateSummaryModule();
      this.injectCodeChanges();
    } catch (error) {
      this.showError('检测技术栈时出错');
      this.isTechStackLoaded = true;
      this.checkAndCreateSummaryModule();
      this.injectCodeChanges();
    }
  }

  /**
   * 注入代码变更功能
   */
  async injectCodeChanges() {
    const url = window.location.href;
    const isPR = this.isPullRequestPage(url);
    const isCommit = this.isCommitPage(url);
    
    if ((!isPR && !isCommit) || this.isCodeChangesInjected) {
      return;
    }
    
    const targetContainer = this.findTargetContainer();
    if (!targetContainer || !this.container) {
      return;
    }
    
    const token = await this.getGitHubToken();
    
    this.createCodeChangesContainer(targetContainer);
    this.isCodeChangesInjected = true;
    
    if (isPR) {
      const prInfo = this.parsePullRequestFromURL(url);
      if (prInfo) {
        this.loadAndDisplayPRChanges(token, prInfo.owner, prInfo.repo, prInfo.pullNumber);
      }
    } else if (isCommit) {
      const commitInfo = this.parseCommitFromURL(url);
      if (commitInfo) {
        this.loadAndDisplayCommitChanges(token, commitInfo.owner, commitInfo.repo, commitInfo.sha);
      }
    }
  }

  /**
   * 清理现有AI摘要
   */
  cleanupExistingAISummary() {
    const existingContainer = document.querySelector('#ai-code-summary-container');
    if (existingContainer) {
      existingContainer.remove();
    }
    
    const existingContainers = document.querySelectorAll('.ai-code-summary-container');
    existingContainers.forEach(container => {
      container.remove();
    });
    
    this.summaryContainer = null;
    this.isTechStackLoaded = false;
    this.isCodeChangesLoaded = false;
  }

  /**
   * 获取GitHub Token
   */
  async getGitHubToken() {
    // 从存储中获取token或使用默认值
    return new Promise((resolve) => {
      chrome.storage.sync.get(['githubToken'], (result) => {
        resolve(result.githubToken || '');
      });
    });
  }

  /**
   * 解析仓库信息
   */
  parseRepositoryFromURL(url) {
    const match = url.match(/github\.com\/([^\/]+)\/([^\/]+)/);
    if (match) {
      return {
        owner: match[1],
        repo: match[2]
      };
    }
    return null;
  }

  /**
   * 检查是否为PR页面
   */
  isPullRequestPage(url) {
    return /github\.com\/[^\/]+\/[^\/]+\/pull\/\d+/.test(url);
  }

  /**
   * 检查是否为Commit页面
   */
  isCommitPage(url) {
    return /github\.com\/[^\/]+\/[^\/]+\/commit\/[a-f0-9]+/.test(url);
  }

  /**
   * 解析PR信息
   */
  parsePullRequestFromURL(url) {
    const match = url.match(/github\.com\/([^\/]+)\/([^\/]+)\/pull\/(\d+)/);
    if (match) {
      return {
        owner: match[1],
        repo: match[2],
        pullNumber: parseInt(match[3])
      };
    }
    return null;
  }

  /**
   * 解析Commit信息
   */
  parseCommitFromURL(url) {
    const match = url.match(/github\.com\/([^\/]+)\/([^\/]+)\/commit\/([a-f0-9]+)/);
    if (match) {
      return {
        owner: match[1],
        repo: match[2],
        sha: match[3]
      };
    }
    return null;
  }

  /**
   * 显示错误信息
   */
  showError(message) {
    if (!this.container) return;
    
    const loadingElement = this.container.querySelector('.tech-stack-loading');
    const contentElement = this.container.querySelector('.tech-stack-content');
    
    if (loadingElement) {
      loadingElement.remove();
    }
    
    if (contentElement) {
      contentElement.innerHTML = `
        <div class="tech-stack-error">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M8 16A8 8 0 108 0a8 8 0 000 16zM5.354 4.646a.5.5 0 10-.708.708L7.293 8l-2.647 2.646a.5.5 0 00.708.708L8 8.707l2.646 2.647a.5.5 0 00.708-.708L8.707 8l2.647-2.646a.5.5 0 00-.708-.708L8 7.293 5.354 4.646z"/>
          </svg>
          ${message}
        </div>
      `;
    }
  }
}

// GitHub API客户端类
class GitHubAPIClient {
  constructor(token) {
    this.token = token;
    this.baseURL = 'https://api.github.com';
  }

  async makeRequest(endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;
    const headers = {
      'Accept': 'application/vnd.github.v3+json',
      ...(this.token && { 'Authorization': `token ${this.token}` }),
      ...options.headers
    };

    const response = await fetch(url, {
      ...options,
      headers
    });

    if (!response.ok) {
      throw new Error(`GitHub API Error: ${response.status} ${response.statusText}`);
    }

    return response.json();
  }

  async getRepository(owner, repo) {
    return this.makeRequest(`/repos/${owner}/${repo}`);
  }

  async getContents(owner, repo, path) {
    return this.makeRequest(`/repos/${owner}/${repo}/contents/${path}`);
  }

  async getPullRequest(owner, repo, pullNumber) {
    return this.makeRequest(`/repos/${owner}/${repo}/pulls/${pullNumber}`);
  }

  async getPullRequestFiles(owner, repo, pullNumber) {
    return this.makeRequest(`/repos/${owner}/${repo}/pulls/${pullNumber}/files`);
  }

  async getCommit(owner, repo, sha) {
    return this.makeRequest(`/repos/${owner}/${repo}/commits/${sha}`);
  }
}

// 技术栈检测器类
class TechStackDetector {
  constructor(githubClient) {
    this.githubClient = githubClient;
  }

  async detectTechStack(owner, repo) {
    const techStacks = [];
    
    // 检测各种配置文件
    const configFiles = [
      'package.json',
      'requirements.txt',
      'pom.xml',
      'build.gradle',
      'Cargo.toml',
      'composer.json',
      'Gemfile',
      'go.mod'
    ];

    for (const file of configFiles) {
      try {
        const content = await this.githubClient.getContents(owner, repo, file);
        const techStack = this.parseConfigFile(file, content);
        if (techStack) {
          techStacks.push(techStack);
        }
      } catch (error) {
        // 文件不存在，继续检测其他文件
      }
    }

    return techStacks;
  }

  parseConfigFile(filename, content) {
    try {
      const decodedContent = atob(content.content);
      
      switch (filename) {
        case 'package.json':
          return this.parsePackageJson(decodedContent);
        case 'requirements.txt':
          return this.parseRequirementsTxt(decodedContent);
        case 'pom.xml':
          return this.parsePomXml(decodedContent);
        // 添加其他文件类型的解析
        default:
          return null;
      }
    } catch (error) {
      console.error(`Error parsing ${filename}:`, error);
      return null;
    }
  }

  parsePackageJson(content) {
    try {
      const packageData = JSON.parse(content);
      return {
        type: 'Node.js',
        name: packageData.name || 'Unknown',
        version: packageData.version,
        dependencies: packageData.dependencies,
        devDependencies: packageData.devDependencies,
        scripts: packageData.scripts,
        metadata: {
          filePath: 'package.json'
        }
      };
    } catch (error) {
      return null;
    }
  }

  parseRequirementsTxt(content) {
    const lines = content.split('\n').filter(line => line.trim() && !line.startsWith('#'));
    const dependencies = {};
    
    lines.forEach(line => {
      const match = line.match(/^([^=<>!]+)([=<>!]+.+)?$/);
      if (match) {
        dependencies[match[1].trim()] = match[2] ? match[2].trim() : '*';
      }
    });

    return {
      type: 'Python',
      name: 'Python Project',
      dependencies,
      metadata: {
        filePath: 'requirements.txt'
      }
    };
  }

  parsePomXml(content) {
    // 简化的XML解析
    const nameMatch = content.match(/<artifactId>(.*?)<\/artifactId>/);
    const versionMatch = content.match(/<version>(.*?)<\/version>/);
    
    return {
      type: 'Java (Maven)',
      name: nameMatch ? nameMatch[1] : 'Java Project',
      version: versionMatch ? versionMatch[1] : 'Unknown',
      metadata: {
        filePath: 'pom.xml'
      }
    };
  }
}

// 初始化插件
new GitHubAICodeReview();
