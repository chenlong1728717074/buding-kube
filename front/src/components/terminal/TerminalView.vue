<template>
  <div class="tv-root">
    <div ref="termRef" class="tv-terminal" :style="{ height }" @click="term?.focus()" />
    <div class="tv-footer">
      <span>{{ tip }}</span>
      <span class="tv-size">{{ size }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, shallowRef } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { getToken } from '@/utils/ws'
import type { TerminalStatus } from '@/utils/terminal'

// 终端视图: xterm + WebSocket 会话, 容器终端和 Web kubectl 共用.
// 协议见后端 api/terminal.go: 二进制帧为终端输出; 文本帧为 JSON 控制消息(info / ready / error / exit)

const props = withDefaults(
  defineProps<{
    // 生成连接地址; 返回空表示参数不完整
    buildUrl: () => string
    height?: string
    tip?: string
  }>(),
  { height: '66vh', tip: '选中文字即可复制（Ctrl+Insert），Shift+Insert 或右键粘贴；超过 30 分钟没有操作会自动断开' }
)

const emit = defineEmits<{ status: [status: TerminalStatus]; ready: [container: string] }>()

const FONT_KEY = 'terminal.fontSize'
const termRef = ref<HTMLElement>()
// 终端实例不能被 Vue 深度代理(会破坏 xterm 内部状态), 使用 shallowRef
const term = shallowRef<Terminal | null>(null)
let fitAddon: FitAddon | null = null
let resizeObserver: ResizeObserver | null = null
let socket: WebSocket | null = null
// 每次连接递增, 旧连接的回调直接忽略
let sessionId = 0
let ended = false
let fitFrame = 0
const fontSize = ref(Math.min(24, Math.max(10, Number(localStorage.getItem(FONT_KEY)) || 14)))
const size = ref('')

const setStatus = (s: TerminalStatus) => emit('status', s)

const write = (text: string, color: 'gray' | 'red' | 'yellow' | 'green' = 'gray') => {
  const code = { gray: '90', red: '31', yellow: '33', green: '32' }[color]
  term.value?.write(`\r\n\x1b[${code}m${text}\x1b[0m\r\n`)
}

const send = (msg: Record<string, unknown>) => {
  if (socket && socket.readyState === WebSocket.OPEN) socket.send(JSON.stringify(msg))
}

const fit = () => {
  cancelAnimationFrame(fitFrame)
  fitFrame = requestAnimationFrame(() => {
    if (!term.value || !fitAddon || !termRef.value?.offsetParent) return
    try {
      fitAddon.fit()
    } catch {
      // 容器不可见时忽略
    }
  })
}

const onContextMenu = async (e: MouseEvent) => {
  if (!navigator.clipboard?.readText || term.value?.hasSelection()) return
  e.preventDefault()
  try {
    const text = await navigator.clipboard.readText()
    if (text) term.value?.paste(text)
  } catch {
    // 浏览器拒绝读取剪贴板时, 仍可用 Shift+Insert 粘贴
  }
}

const createTerminal = () => {
  if (!termRef.value || term.value) return
  const t = new Terminal({
    cursorBlink: true,
    fontSize: fontSize.value,
    fontFamily: "'JetBrains Mono', 'Cascadia Mono', Consolas, Menlo, 'Courier New', monospace",
    lineHeight: 1.15,
    scrollback: 5000,
    allowTransparency: false,
    rightClickSelectsWord: true,
    theme: {
      background: '#0f172a',
      foreground: '#e2e8f0',
      cursor: '#60a5fa',
      cursorAccent: '#0f172a',
      selectionBackground: 'rgba(96, 165, 250, 0.35)',
      black: '#1e293b',
      red: '#f87171',
      green: '#4ade80',
      yellow: '#facc15',
      blue: '#60a5fa',
      magenta: '#c084fc',
      cyan: '#22d3ee',
      white: '#e2e8f0',
      brightBlack: '#64748b',
      brightRed: '#fca5a5',
      brightGreen: '#86efac',
      brightYellow: '#fde047',
      brightBlue: '#93c5fd',
      brightMagenta: '#d8b4fe',
      brightCyan: '#67e8f9',
      brightWhite: '#f8fafc'
    }
  })
  fitAddon = new FitAddon()
  t.loadAddon(fitAddon)
  t.open(termRef.value)
  t.onData(data => send({ type: 'stdin', data }))
  t.onResize(({ cols, rows }) => {
    size.value = `${cols} × ${rows}`
    send({ type: 'resize', cols, rows })
  })
  termRef.value.addEventListener('contextmenu', onContextMenu)
  term.value = t
  resizeObserver = new ResizeObserver(fit)
  resizeObserver.observe(termRef.value)
  fit()
}

const disconnect = () => {
  sessionId++
  if (socket) {
    socket.onopen = socket.onmessage = socket.onclose = socket.onerror = null
    if (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING) socket.close(1000)
    socket = null
  }
  setStatus('closed')
}

const closeReason = (e: CloseEvent) => {
  if (e.code === 1006) return '连接已断开（网络中断、后端不可用或登录已过期）'
  if (e.code === 1000) return '连接已关闭'
  return `连接已关闭（${e.code}）`
}

// 建立新连接; message 为连接前输出的提示
const connect = (message = '正在连接...') => {
  disconnect()
  if (!term.value) return
  const token = getToken()
  if (!token) {
    write('未登录或登录已过期，请重新登录', 'red')
    return
  }
  const url = props.buildUrl()
  if (!url) return
  const id = sessionId
  ended = false
  setStatus('connecting')
  write(message)
  // 令牌通过子协议传递, 不出现在地址中
  const ws = new WebSocket(url, ['buding-terminal', `bearer.${token}`])
  ws.binaryType = 'arraybuffer'
  socket = ws

  ws.onopen = () => {
    if (id !== sessionId) return
    fit()
    if (term.value) send({ type: 'resize', cols: term.value.cols, rows: term.value.rows })
  }
  ws.onmessage = e => {
    if (id !== sessionId || !term.value) return
    if (e.data instanceof ArrayBuffer) {
      term.value.write(new Uint8Array(e.data))
      return
    }
    let msg: { type: string; data?: string }
    try {
      msg = JSON.parse(e.data as string)
    } catch {
      return
    }
    if (msg.type === 'info') {
      write(msg.data || '')
    } else if (msg.type === 'ready') {
      setStatus('connected')
      emit('ready', msg.data || '')
      write(`已连接到容器 ${msg.data}`, 'green')
      // 后端开始转发时终端尺寸可能已变化, 重新同步一次
      send({ type: 'resize', cols: term.value.cols, rows: term.value.rows })
      term.value.focus()
    } else if (msg.type === 'error') {
      ended = true
      write(msg.data || '终端连接失败', 'red')
    } else if (msg.type === 'exit') {
      ended = true
      write(`会话已结束（退出码 ${msg.data}），点击「重新连接」可以再次打开`, 'yellow')
    }
  }
  ws.onclose = e => {
    if (id !== sessionId) return
    socket = null
    setStatus('closed')
    if (!ended) write(closeReason(e), 'yellow')
  }
}

const changeFontSize = (delta: number) => {
  fontSize.value = Math.min(24, Math.max(10, fontSize.value + delta))
  localStorage.setItem(FONT_KEY, String(fontSize.value))
  if (term.value) term.value.options.fontSize = fontSize.value
  fit()
}

const dispose = () => {
  disconnect()
  cancelAnimationFrame(fitFrame)
  resizeObserver?.disconnect()
  resizeObserver = null
  termRef.value?.removeEventListener('contextmenu', onContextMenu)
  term.value?.dispose()
  term.value = null
  fitAddon = null
  size.value = ''
}

onMounted(createTerminal)
onBeforeUnmount(dispose)

defineExpose({
  connect,
  disconnect,
  fit,
  changeFontSize,
  clear: () => term.value?.clear(),
  focus: () => term.value?.focus()
})
</script>

<style scoped>
.tv-root {
  display: flex;
  flex-direction: column;
  min-height: 0;
  background: #0f172a;
}

.tv-terminal {
  flex: 1;
  min-height: 0;
  padding: 8px 4px 8px 12px;
  background: #0f172a;
}

.tv-terminal :deep(.xterm) {
  height: 100%;
}

.tv-footer {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 6px 14px;
  background: #111a2e;
  border-top: 1px solid #1e293b;
  color: #64748b;
  font-size: 12px;
}

.tv-size {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}
</style>
