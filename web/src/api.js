const API_BASE = '/api/v1'
const HEALTH_CHECK_URL = '/healthz'

async function request(url, options = {}) {
  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.error || 'Request failed')
    }

    return { data, error: null }
  } catch (error) {
    return { data: null, error: error.message }
  }
}

export const api = {
  // Health check
  healthCheck: () => request(HEALTH_CHECK_URL),

  // IP List management
  getIPList: () => request(`${API_BASE}/iplist`),
  
  createIPList: (iplist) => request(`${API_BASE}/iplist`, {
    method: 'POST',
    body: JSON.stringify({ iplist }),
  }),
  
  updateIPList: (iplist) => request(`${API_BASE}/iplist`, {
    method: 'PUT',
    body: JSON.stringify({ iplist }),
  }),
  
  deleteIPList: () => request(`${API_BASE}/iplist`, {
    method: 'DELETE',
  }),

  // NCCL Test
  getNCCLDefaults: () => request(`${API_BASE}/nccl/defaults`),
  
  // 普通 NCCL Test - 等待完成后返回
  runNCCLTest: (params) => request(`${API_BASE}/nccl/run`, {
    method: 'POST',
    body: JSON.stringify(params),
  }),

  // 停止当前运行的 NCCL Test
  stopNCCLTest: () => request(`${API_BASE}/nccl/stop`, {
    method: 'POST',
  }),

  // Precheck - 检查节点 GPU 进程状态
  precheck: (filename) => request(`${API_BASE}/nccl/precheck?filename=${encodeURIComponent(filename)}`),
  
  // Stream NCCL Test - 使用 Fetch API 进行流式传输
  runNCCLTestStream: (params, callbacks) => {
    return new Promise((resolve, reject) => {
      // 使用 fetch 发送 POST 请求并获取流
      fetch(`${API_BASE}/nccl/run-stream`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(params),
      })
      .then(response => {
        if (!response.ok) {
          throw new Error('Request failed')
        }
        
        const reader = response.body.getReader()
        const decoder = new TextDecoder()
        let buffer = ''
        
        function read() {
          reader.read().then(({ done, value }) => {
            if (done) {
              resolve({ status: 'success' })
              return
            }
            
            const text = decoder.decode(value, { stream: true })
            buffer += text
            
            // 按行处理
            const lines = buffer.split('\n')
            buffer = lines.pop() || '' // 保留不完整的行
            
            let currentEvent = ''
            
            for (const line of lines) {
              if (line.startsWith('event:')) {
                currentEvent = line.substring(6).trim()
              } else if (line.startsWith('data:')) {
                const data = line.substring(5)  // 不要 trim，保留所有空格
                
                // 根据事件类型处理
                if (currentEvent === 'command' && callbacks.onCommand) {
                  try {
                    callbacks.onCommand(JSON.parse(data))
                  } catch (e) {
                    callbacks.onCommand(data)
                  }
                } else if (currentEvent === 'output' && callbacks.onOutput) {
                  try {
                    callbacks.onOutput(JSON.parse(data))
                  } catch (e) {
                    callbacks.onOutput(data)
                  }
                } else if (currentEvent === 'done' && callbacks.onComplete) {
                  callbacks.onComplete()
                } else if (currentEvent === 'error' && callbacks.onError) {
                  try {
                    const parsed = JSON.parse(data)
                    callbacks.onError(parsed.message || data)
                  } catch (e) {
                    callbacks.onError(data)
                  }
                }
              }
            }
            
            read()
          }).catch(error => {
            if (callbacks.onError) {
              callbacks.onError(error.message)
            }
            reject(error)
          })
        }
        
        read()
      })
      .catch(error => {
        if (callbacks.onError) {
          callbacks.onError(error.message)
        }
        reject(error)
      })
    })
  },
}
