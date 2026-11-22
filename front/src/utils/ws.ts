// 由接口地址(VITE_API_BASE_URL)生成 WebSocket 地址: http -> ws, https -> wss; 相对地址使用当前页面的域名
export function buildWsUrl(path: string, params: Record<string, string | undefined> = {}) {
  const base = (import.meta.env.VITE_API_BASE_URL || '/api').replace(/\/+$/, '')
  const origin = /^https?:\/\//i.test(base)
    ? base.replace(/^http/i, 'ws')
    : `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}${base.startsWith('/') ? base : `/${base}`}`
  const query = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value) query.set(key, value)
  })
  const qs = query.toString()
  return `${origin}${path}${qs ? `?${qs}` : ''}`
}

export const getToken = () => sessionStorage.getItem('token') || localStorage.getItem('token') || ''
