import { ref } from 'vue'

export type Language = 'zh-CN' | 'en-US'

export const currentLanguage = ref<Language>('zh-CN')

type MessageKey = 
  | 'appTitle' | 'login' | 'logout' | 'user'
  | 'documents' | 'newDocumentTitle' | 'create' | 'settings' | 'title' | 'updatedAt' | 'wordCount' | 'actions' | 'open' | 'rename' | 'delete' | 'confirmDelete' | 'confirmDeleteMessage' | 'confirm' | 'cancel' | 'pleaseInputTitle'
  | 'editor' | 'chars' | 'savedAt' | 'saving' | 'startWriting' | 'save' | 'backHome'
  | 'aiProvider' | 'apiKey' | 'inputApiKey' | 'tongyi' | 'wenxin' | 'zhipu'
  | 'username' | 'password' | 'submit' | 'goToRegister' | 'goToLogin'
  | 'registerSuccess' | 'loginSuccess' | 'invalidRequestData' | 'userExists' | 'passwordEncryptFailed' | 'invalidCredentials' | 'tokenGenerationFailed' | 'missingToken' | 'invalidToken' | 'documentNotFound'

export const messages: Record<Language, Record<MessageKey, string>> = {
  'zh-CN': {
    // App
    appTitle: 'AI 写作助手',
    login: '登录',
    logout: '退出',
    user: '用户',
    
    // Home
    documents: '文档',
    newDocumentTitle: '新建文档标题',
    create: '新建',
    settings: '设置',
    title: '标题',
    updatedAt: '更新时间',
    wordCount: '字符数',
    actions: '操作',
    open: '打开',
    rename: '重命名',
    delete: '删除',
    confirmDelete: '确认删除',
    confirmDeleteMessage: '确认删除"{title}"吗？',
    confirm: '确认',
    cancel: '取消',
    pleaseInputTitle: '请输入标题',
    
    // Editor
    editor: '编辑器',
    chars: '字符数：',
    savedAt: '保存于：',
    saving: '保存中...',
    startWriting: '开始写作...',
    save: '保存',
    backHome: '返回首页',
    
    // Settings
    aiProvider: 'AI 提供商',
    apiKey: 'API 密钥',
    inputApiKey: '请输入 API 密钥',
    tongyi: '通义',
    wenxin: '文心',
    zhipu: '智谱',
    
    // Login
    username: '用户名',
    password: '密码',
    submit: '提交',
    goToRegister: '去注册',
    goToLogin: '去登录',
    
    // Messages
    registerSuccess: '注册成功',
    loginSuccess: '登录成功',
    invalidRequestData: '无效的请求数据',
    userExists: '用户已存在',
    passwordEncryptFailed: '密码加密失败',
    invalidCredentials: '用户名或密码错误',
    tokenGenerationFailed: '生成令牌失败',
    missingToken: '缺少认证令牌',
    invalidToken: '无效的认证令牌',
    documentNotFound: '文档不存在'
  },
  'en-US': {
    // App
    appTitle: 'AI Writing Assistant',
    login: 'Login',
    logout: 'Logout',
    user: 'User',
    
    // Home
    documents: 'Documents',
    newDocumentTitle: 'New document title',
    create: 'Create',
    settings: 'Settings',
    title: 'Title',
    updatedAt: 'Updated',
    wordCount: 'Chars',
    actions: 'Actions',
    open: 'Open',
    rename: 'Rename',
    delete: 'Delete',
    confirmDelete: 'Confirm',
    confirmDeleteMessage: 'Delete "{title}"?',
    confirm: 'OK',
    cancel: 'Cancel',
    pleaseInputTitle: 'Please input a title',
    
    // Editor
    editor: 'Editor',
    chars: 'Chars: ',
    savedAt: 'Saved at: ',
    saving: 'Saving...',
    startWriting: 'Start writing...',
    save: 'Save',
    backHome: 'Back Home',
    
    // Settings
    aiProvider: 'AI Provider',
    apiKey: 'API Key',
    inputApiKey: 'Paste your API key',
    tongyi: 'Tongyi',
    wenxin: 'Wenxin',
    zhipu: 'Zhipu',
    
    // Login
    username: 'Username',
    password: 'Password',
    submit: 'Submit',
    goToRegister: 'Go to Register',
    goToLogin: 'Go to Login',
    
    // Messages
    registerSuccess: 'Registration successful',
    loginSuccess: 'Login successful',
    invalidRequestData: 'Invalid request data',
    userExists: 'User already exists',
    passwordEncryptFailed: 'Password encryption failed',
    invalidCredentials: 'Invalid username or password',
    tokenGenerationFailed: 'Token generation failed',
    missingToken: 'Missing authentication token',
    invalidToken: 'Invalid authentication token',
    documentNotFound: 'Document not found'
  }
}

export function t(key: MessageKey): string {
  const lang = currentLanguage.value
  return messages[lang]?.[key] || key
}

export function setLanguage(lang: Language) {
  currentLanguage.value = lang
  localStorage.setItem('language', lang)
}

export function initLanguage() {
  const saved = localStorage.getItem('language') as Language
  if (saved && messages[saved]) {
    currentLanguage.value = saved
  }
} 