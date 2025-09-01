/**
 * GitHub AI代码审查插件 - 弹窗脚本
 * 处理插件弹窗的交互逻辑
 */

class PopupController {
  constructor() {
    this.currentTab = null;
    this.init();
  }

  /**
   * 初始化弹窗
   */
  async init() {
    await this.getCurrentTab();
    this.updatePageStatus();
    this.loadSettings();
    this.bindEvents();
  }

  /**
   * 获取当前标签页
   */
  async getCurrentTab() {
    try {
      const tabs = await chrome.tabs.query({ active: true, currentWindow: true });
      this.currentTab = tabs[0];
    } catch (error) {
      console.error('获取当前标签页失败:', error);
    }
  }

  /**
   * 更新页面状态显示
   */
  updatePageStatus() {
    const currentPageElement = document.getElementById('current-page');
    const pluginStatusElement = document.getElementById('plugin-status');

    if (!this.currentTab) {
      currentPageElement.textContent = '未知页面';
      pluginStatusElement.textContent = '未激活';
      pluginStatusElement.className = 'status-value status-inactive';
      return;
    }

    const url = this.currentTab.url;
    
    if (this.isGitHubPage(url)) {
      if (this.isPullRequestPage(url)) {
        currentPageElement.textContent = 'GitHub Pull Request';
      } else if (this.isCommitPage(url)) {
        currentPageElement.textContent = 'GitHub Commit';
      } else if (this.isRepositoryPage(url)) {
        currentPageElement.textContent = 'GitHub 仓库';
      } else {
        currentPageElement.textContent = 'GitHub 页面';
      }
      
      pluginStatusElement.textContent = '已激活';
      pluginStatusElement.className = 'status-value status-active';
    } else {
      currentPageElement.textContent = '非GitHub页面';
      pluginStatusElement.textContent = '未激活';
      pluginStatusElement.className = 'status-value status-inactive';
    }
  }

  /**
   * 加载设置
   */
  async loadSettings() {
    try {
      const result = await chrome.storage.sync.get(['githubToken', 'aiModel']);
      
      const tokenInput = document.getElementById('github-token');
      const modelSelect = document.getElementById('ai-model');
      
      if (result.githubToken) {
        tokenInput.value = result.githubToken;
      }
      
      if (result.aiModel) {
        modelSelect.value = result.aiModel;
      }
    } catch (error) {
      console.error('加载设置失败:', error);
    }
  }

  /**
   * 保存设置
   */
  async saveSettings() {
    try {
      const tokenInput = document.getElementById('github-token');
      const modelSelect = document.getElementById('ai-model');
      
      await chrome.storage.sync.set({
        githubToken: tokenInput.value,
        aiModel: modelSelect.value
      });
      
      this.showNotification('设置已保存', 'success');
    } catch (error) {
      console.error('保存设置失败:', error);
      this.showNotification('保存失败', 'error');
    }
  }

  /**
   * 绑定事件
   */
  bindEvents() {
    // 保存Token按钮
    const saveTokenBtn = document.getElementById('save-token');
    saveTokenBtn.addEventListener('click', () => {
      this.saveSettings();
    });

    // AI模型选择
    const aiModelSelect = document.getElementById('ai-model');
    aiModelSelect.addEventListener('change', () => {
      this.saveSettings();
    });

    // 刷新页面按钮
    const refreshBtn = document.getElementById('refresh-page');
    refreshBtn.addEventListener('click', () => {
      this.refreshCurrentPage();
    });

    // 高级设置按钮
    const settingsBtn = document.getElementById('open-settings');
    settingsBtn.addEventListener('click', () => {
      this.openSettings();
    });

    // 帮助链接
    const helpLink = document.getElementById('help-link');
    helpLink.addEventListener('click', (e) => {
      e.preventDefault();
      this.openHelp();
    });

    // 反馈链接
    const feedbackLink = document.getElementById('feedback-link');
    feedbackLink.addEventListener('click', (e) => {
      e.preventDefault();
      this.openFeedback();
    });

    // Enter键保存Token
    const tokenInput = document.getElementById('github-token');
    tokenInput.addEventListener('keypress', (e) => {
      if (e.key === 'Enter') {
        this.saveSettings();
      }
    });
  }

  /**
   * 刷新当前页面
   */
  async refreshCurrentPage() {
    try {
      if (this.currentTab) {
        await chrome.tabs.reload(this.currentTab.id);
        window.close();
      }
    } catch (error) {
      console.error('刷新页面失败:', error);
      this.showNotification('刷新失败', 'error');
    }
  }

  /**
   * 打开设置页面
   */
  openSettings() {
    chrome.tabs.create({
      url: chrome.runtime.getURL('options.html')
    });
    window.close();
  }

  /**
   * 打开帮助页面
   */
  openHelp() {
    chrome.tabs.create({
      url: 'https://github.com/lijiarui/ai-code-review-to-github/blob/main/README.md'
    });
    window.close();
  }

  /**
   * 打开反馈页面
   */
  openFeedback() {
    chrome.tabs.create({
      url: 'https://github.com/lijiarui/ai-code-review-to-github/issues'
    });
    window.close();
  }

  /**
   * 显示通知
   */
  showNotification(message, type = 'info') {
    // 创建通知元素
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.textContent = message;

    // 添加到页面
    document.body.appendChild(notification);

    // 显示动画
    setTimeout(() => {
      notification.classList.add('show');
    }, 10);

    // 自动隐藏
    setTimeout(() => {
      notification.classList.remove('show');
      setTimeout(() => {
        if (notification.parentNode) {
          notification.parentNode.removeChild(notification);
        }
      }, 300);
    }, 2000);
  }

  /**
   * 检查是否为GitHub页面
   */
  isGitHubPage(url) {
    return url && url.includes('github.com');
  }

  /**
   * 检查是否为仓库页面
   */
  isRepositoryPage(url) {
    return /github\.com\/[^\/]+\/[^\/]+(?:\/(?:tree|blob|commits?|releases|issues|pull|actions|projects|wiki|security|insights|settings)?(?:\/.*)?)?$/.test(url);
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
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
  new PopupController();
});

// 检查页面可见性变化
document.addEventListener('visibilitychange', () => {
  if (!document.hidden) {
    // 页面重新获得焦点时更新状态
    const controller = new PopupController();
    controller.updatePageStatus();
  }
});
