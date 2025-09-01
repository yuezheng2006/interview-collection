/**
 * GitHub AI代码审查插件 - 后台脚本
 * 处理API请求和流式响应
 */

// API配置
const API_BASE_URL = 'http://115.190.121.185:3000';

/**
 * 处理流式聊天连接
 */
async function handleStreamChat(port) {
  let abortController = null;
  
  port.onMessage.addListener(async (message) => {
    if (message.type === 'start-stream') {
      try {
        console.log('🌊 开始处理流式请求:', message.data.url);
        
        abortController = new AbortController();
        const { url, method, headers, body } = message.data;
        
        const response = await fetch(url, {
          method,
          headers,
          body,
          signal: abortController.signal
        });
        
        console.log('📥 流式响应状态:', response.status, response.statusText);
        
        if (!response.ok) {
          const errorText = await response.text();
          port.postMessage({
            type: 'error',
            error: `HTTP ${response.status}: ${errorText}`
          });
          return;
        }
        
        const reader = response.body?.getReader();
        if (!reader) {
          port.postMessage({
            type: 'error',
            error: 'No readable stream available'
          });
          return;
        }
        
        const decoder = new TextDecoder();
        let buffer = '';
        
        try {
          while (true) {
            const { done, value } = await reader.read();
            
            if (done) {
              console.log('✅ 流式响应读取完成');
              port.postMessage({ type: 'end' });
              break;
            }
            
            const chunk = decoder.decode(value, { stream: true });
            console.log('📦 收到数据块:', chunk.substring(0, 100) + '...');
            
            buffer += chunk;
            const lines = buffer.split('\n');
            buffer = lines.pop() || '';
            
            for (const line of lines) {
              if (line.startsWith('data: ')) {
                const data = line.substring(6);
                
                if (data === '[DONE]') {
                  console.log('✅ 收到流结束标记');
                  port.postMessage({ type: 'end' });
                  return;
                }
                
                try {
                  const parsedData = JSON.parse(data);
                  console.log('📦 解析数据:', parsedData);
                  
                  if (parsedData.type === 'error') {
                    console.warn('⚠️ 收到错误数据，但继续处理:', parsedData);
                    continue;
                  }
                  
                  if (['text-start', 'text-delta', 'text-end', 'data-history'].includes(parsedData.type)) {
                    port.postMessage({
                      type: 'chunk',
                      chunk: parsedData
                    });
                  }
                } catch (error) {
                  console.warn('⚠️ 解析数据块失败:', line, error);
                }
              }
            }
          }
        } finally {
          reader.releaseLock();
        }
      } catch (error) {
        console.error('❌ 流式请求处理失败:', error);
        port.postMessage({
          type: 'error',
          error: error instanceof Error ? error.message : 'Unknown error'
        });
      }
    } else if (message.type === 'abort') {
      console.log('🛑 中止流式请求');
      if (abortController) {
        abortController.abort();
      }
    }
  });
  
  port.onDisconnect.addListener(() => {
    console.log('🔌 流式聊天连接断开');
    if (abortController) {
      abortController.abort();
    }
  });
}

/**
 * 处理API请求
 */
async function handleApiRequest(requestData) {
  try {
    let response;
    const { url, method, headers, body } = requestData;
    
    console.log('🔗 Background: 发送API请求到', url);
    console.log('📤 请求数据:', {
      method,
      headers,
      bodyLength: body?.length
    });
    console.log('📤 完整请求头:', headers);
    
    if (body) {
      try {
        const parsedBody = JSON.parse(body);
        console.log('📋 请求体内容:', {
          system: parsedBody.system?.substring(0, 150),
          prompt: parsedBody.prompt?.substring(0, 150),
          descriptionLength: parsedBody.description?.length || 0,
          description: parsedBody.description?.substring(0, 300)
        });
        console.log('🔍 字段检查:', {
          hasSystem: !!parsedBody.system,
          hasPrompt: !!parsedBody.prompt,
          hasDescription: !!parsedBody.description,
          systemType: typeof parsedBody.system,
          promptType: typeof parsedBody.prompt,
          descriptionType: typeof parsedBody.description
        });
      } catch (error) {
        console.log('📋 请求体（原始）:', body.substring(0, 500) + '...');
      }
    }
    
    const fetchResponse = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        'Accept': '*/*',
        'Cache-Control': 'no-cache',
        'Pragma': 'no-cache',
        ...headers
      },
      body
    });
    
    console.log('📥 响应状态:', fetchResponse.status, fetchResponse.statusText);
    console.log('📋 响应头:', Object.fromEntries(fetchResponse.headers.entries()));
    
    const contentType = fetchResponse.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      response = await fetchResponse.json();
      console.log('✅ JSON 响应解析成功');
      console.log('📄 响应数据:', {
        issuesCount: response?.issues?.length || 0,
        summary: response?.summary || null,
        error: response?.error || null,
        fullResponse: response
      });
    } else {
      response = await fetchResponse.text();
      console.log('📝 文本响应:', response.substring(0, 500));
    }
    
    const result = {
      ok: fetchResponse.ok,
      status: fetchResponse.status,
      statusText: fetchResponse.statusText,
      data: response
    };
    
    console.log('🎯 Background: 返回结果:', result);
    return result;
    
  } catch (error) {
    console.error('❌ Background: API 请求失败', error);
    
    if (error instanceof TypeError && error.message.includes('fetch')) {
      return {
        ok: false,
        status: 0,
        statusText: 'Network Connection Failed',
        error: '无法连接到后端服务，请确保服务已启动'
      };
    }
    
    return {
      ok: false,
      status: 500,
      statusText: 'Network Error',
      error: error instanceof Error ? error.message : 'Unknown error'
    };
  }
}

/**
 * 获取API URL
 */
function getApiUrl(endpoint) {
  return `${API_BASE_URL}${endpoint}`;
}

/**
 * 获取API基础URL
 */
function getApiBaseUrl() {
  return API_BASE_URL;
}

// 监听消息
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.type === 'API_REQUEST') {
    handleApiRequest(message.data)
      .then(result => sendResponse(result))
      .catch(error => sendResponse({
        ok: false,
        status: 500,
        statusText: 'Internal Error',
        error: error.message
      }));
    
    return true; // 保持消息通道开放
  }
});

// 监听端口连接
chrome.runtime.onConnect.addListener((port) => {
  if (port.name === 'stream-chat') {
    console.log('🔌 建立流式聊天连接');
    handleStreamChat(port);
  }
});

// 插件安装时的初始化
chrome.runtime.onInstalled.addListener(() => {
  console.log('Code Review Extension installed');
});

// 导出配置
export const config = {
  API_BASE_URL,
  getApiUrl,
  getApiBaseUrl
};
