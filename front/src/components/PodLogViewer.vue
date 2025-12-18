<template>
  <UnifiedDialog
    v-model="visible"
    :title="title"
    :subtitle="subtitle"
    :width="width"
  >
    <div class="log-header">
      <div class="log-controls">
        <el-form inline>
          <el-form-item v-if="containerSelectable" label="容器:">
            <el-select
              v-model="localContainerName"
              placeholder="全部容器"
              style="width: 200px"
              @change="handleContainerChange"
            >
              <el-option v-if="allowAllContainers" label="全部容器" value="" />
              <el-option
                v-for="c in normalizedContainers"
                :key="c"
                :label="c"
                :value="c"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="字体大小:">
            <el-input-number
              v-model="logFontSize"
              :min="10"
              :max="28"
              :step="1"
              style="width: 120px"
            />
          </el-form-item>
          <el-form-item label="查找:">
            <el-input
              v-model="searchQuery"
              placeholder="输入关键字"
              clearable
              style="width: 200px"
              @keyup.enter="handleFindNext"
            />
          </el-form-item>
          <el-form-item>
            <el-button size="small" @click="handleFindPrev" :disabled="searchMatchCount === 0">上一个</el-button>
            <el-button size="small" @click="handleFindNext" :disabled="searchMatchCount === 0">下一个</el-button>
            <span class="log-search-count">{{ searchMatchCount === 0 ? '0/0' : `${searchActiveIndex + 1}/${searchMatchCount}` }}</span>
          </el-form-item>
          <el-form-item label="开始时间:">
            <el-date-picker
              v-model="sinceTime"
              type="datetime"
              placeholder="选择开始时间"
              style="width: 200px"
              format="YYYY年MM月DD日 HH:mm:ss"
              value-format="YYYY-MM-DD HH:mm:ss"
              clearable
              @change="handleTimeChange"
              :teleported="false"
            />
          </el-form-item>
          <el-form-item label="行数:">
            <el-input-number
              v-model="tailLines"
              :min="100"
              :max="10000"
              :step="100"
              style="width: 140px"
            />
          </el-form-item>
          <el-form-item v-if="allowRealTime" label="实时日志:">
            <el-button
              :type="followLogs ? 'danger' : 'success'"
              @click="handleRealTimeToggle"
              :loading="logLoading"
            >
              <el-icon>
                <VideoPlay v-if="!followLogs" />
                <VideoPause v-else />
              </el-icon>
              {{ followLogs ? '停止实时' : '开始实时' }}
            </el-button>
          </el-form-item>
          <el-form-item label="配色:">
            <el-radio-group v-model="logColorMode" size="small">
              <el-radio-button label="custom">自定义</el-radio-button>
              <el-radio-button label="theme">主题</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item v-if="logColorMode === 'theme'" label="主题:">
            <el-select v-model="logTheme" style="width: 160px">
              <el-option v-for="t in logThemeOptions" :key="t.value" :label="t.label" :value="t.value" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="logColorMode === 'custom'" label="字体颜色:">
            <el-color-picker v-model="logTextColor" />
          </el-form-item>
          <el-form-item v-if="logColorMode === 'custom'" label="背景颜色:">
            <el-color-picker v-model="logBackgroundColor" />
          </el-form-item>
          <el-form-item label="ANSI渲染:">
            <el-switch v-model="ansiEnabled" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleRefreshLogs" :loading="logLoading">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
            <el-button @click="handleClearLogs">
              <el-icon><Delete /></el-icon>
              清空
            </el-button>
            <el-button @click="handleResetLogs">
              <el-icon><RefreshLeft /></el-icon>
              重置
            </el-button>
            <el-button @click="handleDownloadLogs">
              <el-icon><Download /></el-icon>
              下载
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>

    <div v-loading="logLoading" class="log-container" :style="{ backgroundColor: effectiveBackgroundColor }">
      <div v-if="followLogs" class="real-time-indicator">
        <el-icon class="blinking"><VideoPause /></el-icon>
        实时日志连接中
      </div>
      <pre
        ref="logContentRef"
        class="log-content"
        :style="{ color: effectiveTextColor, backgroundColor: effectiveBackgroundColor, fontSize: `${logFontSize}px` }"
        v-html="renderedLog.html"
      ></pre>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">关闭</el-button>
      </div>
    </template>
  </UnifiedDialog>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  VideoPlay,
  VideoPause,
  Refresh,
  RefreshLeft,
  Download,
  Delete
} from '@element-plus/icons-vue'
import UnifiedDialog from '@/components/UnifiedDialog.vue'
import { podApi, type PodLogDTO } from '@/api/pod'

interface Props {
  modelValue: boolean
  title: string
  subtitle?: string
  width?: string
  clusterId: string
  namespace: string
  podName: string
  containers?: Array<string | { name: string }>
  containerSelectable?: boolean
  allowAllContainers?: boolean
  containerName?: string
  allowRealTime?: boolean
  defaultTailLines?: number
}

const props = withDefaults(defineProps<Props>(), {
  width: '90%',
  containers: () => [],
  containerSelectable: true,
  allowAllContainers: true,
  containerName: '',
  allowRealTime: true,
  defaultTailLines: 1000
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'update:containerName': [value: string]
}>()

const visible = computed({
  get: () => props.modelValue,
  set: v => emit('update:modelValue', v)
})

const normalizedContainers = computed(() =>
  (props.containers || [])
    .map(v => (typeof v === 'string' ? v : v?.name))
    .filter((v): v is string => !!v)
)

const localContainerName = computed({
  get: () => props.containerName || '',
  set: v => emit('update:containerName', v)
})

const logContent = ref('')
const logLoading = ref(false)
const followLogs = ref(false)
const sinceTime = ref('')
const tailLines = ref<number>(props.defaultTailLines)
const currentAbortController = ref<AbortController | null>(null)

const logTextColor = ref('#ffffff')
const logBackgroundColor = ref('#000000')
const logFontSize = ref(14)
const searchQuery = ref('')
const searchActiveIndex = ref(0)
const ansiEnabled = ref(true)
const logContentRef = ref<HTMLElement | null>(null)

type LogColorMode = 'custom' | 'theme'
type LogThemeKey = 'dark' | 'light' | 'monokai' | 'solarized-dark' | 'solarized-light'

const logColorMode = ref<LogColorMode>('custom')
const logTheme = ref<LogThemeKey>('dark')

const logThemeOptions: Array<{ value: LogThemeKey; label: string }> = [
  { value: 'dark', label: 'Dark' },
  { value: 'light', label: 'Light' },
  { value: 'monokai', label: 'Monokai' },
  { value: 'solarized-dark', label: 'Solarized Dark' },
  { value: 'solarized-light', label: 'Solarized Light' }
]

const logThemePresets: Record<LogThemeKey, { background: string; text: string }> = {
  dark: { background: '#000000', text: '#ffffff' },
  light: { background: '#ffffff', text: '#111827' },
  monokai: { background: '#272822', text: '#f8f8f2' },
  'solarized-dark': { background: '#002b36', text: '#93a1a1' },
  'solarized-light': { background: '#fdf6e3', text: '#657b83' }
}

const effectiveTextColor = computed(() => {
  if (logColorMode.value === 'theme') return logThemePresets[logTheme.value].text
  return logTextColor.value
})

const effectiveBackgroundColor = computed(() => {
  if (logColorMode.value === 'theme') return logThemePresets[logTheme.value].background
  return logBackgroundColor.value
})

const escapeHtml = (value: string) =>
  value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')

const xterm16: string[] = [
  '#000000',
  '#cd3131',
  '#0dbc79',
  '#e5e510',
  '#2472c8',
  '#bc3fbc',
  '#11a8cd',
  '#e5e5e5',
  '#666666',
  '#f14c4c',
  '#23d18b',
  '#f5f543',
  '#3b8eea',
  '#d670d6',
  '#29b8db',
  '#ffffff'
]

const xterm256ToCss = (n: number) => {
  if (Number.isNaN(n)) return undefined
  if (n < 0) return undefined
  if (n < 16) return xterm16[n]
  if (n >= 16 && n <= 231) {
    const i = n - 16
    const r = Math.floor(i / 36)
    const g = Math.floor((i % 36) / 6)
    const b = i % 6
    const steps = [0, 95, 135, 175, 215, 255]
    return `rgb(${steps[r]},${steps[g]},${steps[b]})`
  }
  if (n >= 232 && n <= 255) {
    const v = 8 + (n - 232) * 10
    return `rgb(${v},${v},${v})`
  }
  return undefined
}

type AnsiState = {
  fg: string | null
  bg: string | null
  bold: boolean
}

const ansiBasicFg = (code: number) => xterm16[code - 30]
const ansiBrightFg = (code: number) => xterm16[8 + (code - 90)]
const ansiBasicBg = (code: number) => xterm16[code - 40]
const ansiBrightBg = (code: number) => xterm16[8 + (code - 100)]

const highlightChunk = (
  raw: string,
  query: string,
  matchIndexOffset: number,
  activeIndex: number
): { html: string; count: number } => {
  if (!query) return { html: escapeHtml(raw), count: 0 }
  const lower = raw.toLowerCase()
  const q = query.toLowerCase()
  let idx = 0
  let out = ''
  let count = 0
  while (true) {
    const found = lower.indexOf(q, idx)
    if (found === -1) break
    out += escapeHtml(raw.slice(idx, found))
    const hitIndex = matchIndexOffset + count
    const active = hitIndex === activeIndex ? '1' : '0'
    out += `<span class="log-search-hit" data-log-hit="${hitIndex}" data-log-active="${active}">${escapeHtml(
      raw.slice(found, found + query.length)
    )}</span>`
    count += 1
    idx = found + query.length
  }
  out += escapeHtml(raw.slice(idx))
  return { html: out, count }
}

const toStyleAttr = (state: AnsiState) => {
  const styles: string[] = []
  if (state.fg) styles.push(`color:${state.fg}`)
  if (state.bg) styles.push(`background-color:${state.bg}`)
  if (state.bold) styles.push('font-weight:600')
  return styles.length ? styles.join(';') : ''
}

const renderPlainHtml = (input: string, query: string, activeIndex: number) => {
  const { html, count } = highlightChunk(input, query, 0, activeIndex)
  return { html, matchCount: count }
}

const renderAnsiHtml = (input: string, query: string, activeIndex: number) => {
  let out = ''
  let i = 0
  let matchCount = 0
  const state: AnsiState = { fg: null, bg: null, bold: false }

  const flushText = (text: string) => {
    if (!text) return
    const highlighted = highlightChunk(text, query, matchCount, activeIndex)
    matchCount += highlighted.count
    const style = toStyleAttr(state)
    if (style) {
      out += `<span style="${style}">${highlighted.html}</span>`
    } else {
      out += highlighted.html
    }
  }

  while (i < input.length) {
    const escIndex = input.indexOf('\u001b', i)
    if (escIndex === -1) {
      flushText(input.slice(i))
      break
    }
    flushText(input.slice(i, escIndex))
    i = escIndex
    if (input[i] !== '\u001b' || input[i + 1] !== '[') {
      i += 1
      continue
    }
    let j = i + 2
    while (j < input.length) {
      const ch = input[j]
      if (ch >= '@' && ch <= '~') break
      j += 1
    }
    if (j >= input.length) break
    const finalChar = input[j]
    const seq = input.slice(i + 2, j)
    i = j + 1
    if (finalChar !== 'm') continue
    const parts = seq.length ? seq.split(';') : ['0']
    const nums = parts.map(p => (p === '' ? 0 : Number.parseInt(p, 10)))
    for (let k = 0; k < nums.length; k++) {
      const code = nums[k]
      if (code === 0) {
        state.fg = null
        state.bg = null
        state.bold = false
      } else if (code === 1) {
        state.bold = true
      } else if (code === 22) {
        state.bold = false
      } else if (code === 39) {
        state.fg = null
      } else if (code === 49) {
        state.bg = null
      } else if (code >= 30 && code <= 37) {
        state.fg = ansiBasicFg(code)
      } else if (code >= 90 && code <= 97) {
        state.fg = ansiBrightFg(code)
      } else if (code >= 40 && code <= 47) {
        state.bg = ansiBasicBg(code)
      } else if (code >= 100 && code <= 107) {
        state.bg = ansiBrightBg(code)
      } else if (code === 38 || code === 48) {
        const isFg = code === 38
        const mode = nums[k + 1]
        if (mode === 5) {
          const n = nums[k + 2]
          const css = xterm256ToCss(n)
          if (css) {
            if (isFg) state.fg = css
            else state.bg = css
          }
          k += 2
        } else if (mode === 2) {
          const r = nums[k + 2]
          const g = nums[k + 3]
          const b = nums[k + 4]
          if (
            Number.isFinite(r) &&
            Number.isFinite(g) &&
            Number.isFinite(b) &&
            r >= 0 &&
            r <= 255 &&
            g >= 0 &&
            g <= 255 &&
            b >= 0 &&
            b <= 255
          ) {
            const css = `rgb(${r},${g},${b})`
            if (isFg) state.fg = css
            else state.bg = css
          }
          k += 4
        }
      }
    }
  }

  return { html: out, matchCount }
}

const renderedLog = computed(() => {
  const query = searchQuery.value.trim()
  if (!logContent.value) return { html: '', matchCount: 0 }
  if (ansiEnabled.value) return renderAnsiHtml(logContent.value, query, searchActiveIndex.value)
  return renderPlainHtml(logContent.value, query, searchActiveIndex.value)
})

const searchMatchCount = computed(() => renderedLog.value.matchCount)

const scrollLogToBottom = () => {
  if (!logContentRef.value) return
  logContentRef.value.scrollTop = logContentRef.value.scrollHeight
}

const scrollToActiveMatch = async () => {
  await nextTick()
  const root = logContentRef.value
  if (!root) return
  const el = root.querySelector(`[data-log-hit="${searchActiveIndex.value}"]`) as HTMLElement | null
  if (el) el.scrollIntoView({ block: 'center' })
}

const handleFindNext = async () => {
  const total = searchMatchCount.value
  if (total === 0) return
  searchActiveIndex.value = (searchActiveIndex.value + 1) % total
  await scrollToActiveMatch()
}

const handleFindPrev = async () => {
  const total = searchMatchCount.value
  if (total === 0) return
  searchActiveIndex.value = (searchActiveIndex.value - 1 + total) % total
  await scrollToActiveMatch()
}

const handleClearLogs = () => {
  logContent.value = ''
  nextTick(() => {
    if (!logContentRef.value) return
    logContentRef.value.scrollTop = 0
  })
}

watch(searchQuery, async () => {
  searchActiveIndex.value = 0
  await scrollToActiveMatch()
})

watch(searchMatchCount, async (count) => {
  if (count === 0) {
    searchActiveIndex.value = 0
    return
  }
  if (searchActiveIndex.value >= count) {
    searchActiveIndex.value = count - 1
    await scrollToActiveMatch()
  }
})

const stopRealTimeLogs = () => {
  followLogs.value = false
  if (currentAbortController.value) {
    currentAbortController.value.abort()
    currentAbortController.value = null
  }
  logLoading.value = false
}

const shouldFetch = () => {
  if (!props.clusterId || !props.namespace || !props.podName) return false
  if (props.containerSelectable && !props.allowAllContainers && normalizedContainers.value.length > 0 && !localContainerName.value) {
    return false
  }
  return true
}

const fetchPodLogs = async () => {
  if (!shouldFetch()) {
    ElMessage.warning('缺少必要参数，无法获取日志')
    return
  }

  if (!followLogs.value) {
    logLoading.value = true
  }

  try {
    if (currentAbortController.value) {
      currentAbortController.value.abort()
      currentAbortController.value = null
    }

    logContent.value = ''

    const params: PodLogDTO = {
      clusterId: props.clusterId,
      namespace: props.namespace,
      name: props.podName,
      containerName: localContainerName.value ? localContainerName.value : undefined,
      follow: props.allowRealTime ? followLogs.value : false,
      sinceTime: sinceTime.value || undefined,
      tailLines: tailLines.value
    }

    currentAbortController.value = new AbortController()

    await podApi.getLogs(
      params,
      (data: string) => {
        const trimmed = data.trim()
        if (
          trimmed &&
          !trimmed.startsWith('[START]') &&
          !trimmed.startsWith('[END]') &&
          !trimmed.startsWith('[ERROR]') &&
          !trimmed.startsWith('[STOP]') &&
          trimmed !== 'ping' &&
          trimmed !== 'Connected successfully'
        ) {
          logContent.value += data
          nextTick(() => {
            scrollLogToBottom()
          })
        }
      },
      (error: Error) => {
        if (error.name !== 'AbortError') {
          ElMessage.error('日志流连接中断')
          followLogs.value = false
        }
      },
      currentAbortController.value?.signal
    )
  } catch (error: any) {
    if (!error?.name || error.name !== 'AbortError') {
      logContent.value = '获取日志失败: ' + (error?.message || '网络错误')
      ElMessage.error('获取日志失败')
      followLogs.value = false
    }
  } finally {
    if (!followLogs.value) {
      logLoading.value = false
    }
  }
}

const handleRefreshLogs = () => {
  fetchPodLogs()
}

const startRealTimeLogs = () => {
  followLogs.value = true
  fetchPodLogs()
}

const handleRealTimeToggle = () => {
  if (!props.allowRealTime) return
  if (followLogs.value) {
    stopRealTimeLogs()
    ElMessage.success('已停止实时日志')
    return
  }
  startRealTimeLogs()
}

const handleTimeChange = () => {
  fetchPodLogs()
}

const handleResetLogs = () => {
  sinceTime.value = ''
  tailLines.value = props.defaultTailLines
  logTextColor.value = '#ffffff'
  logBackgroundColor.value = '#000000'
  logFontSize.value = 14
  searchQuery.value = ''
  searchActiveIndex.value = 0
  ansiEnabled.value = true
  logColorMode.value = 'custom'
  logTheme.value = 'dark'
  fetchPodLogs()
}

const handleContainerChange = () => {
  fetchPodLogs()
}

const handleDownloadLogs = () => {
  if (!logContent.value) {
    ElMessage.warning('暂无日志可下载')
    return
  }
  const blob = new Blob([logContent.value], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  const containerSuffix = localContainerName.value ? localContainerName.value : 'all-containers'
  a.download = `${props.podName}-${containerSuffix}.log`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

const resetOnClose = () => {
  stopRealTimeLogs()
  logContent.value = ''
  sinceTime.value = ''
  tailLines.value = props.defaultTailLines
  logTextColor.value = '#ffffff'
  logBackgroundColor.value = '#000000'
  logFontSize.value = 14
  searchQuery.value = ''
  searchActiveIndex.value = 0
  ansiEnabled.value = true
  logColorMode.value = 'custom'
  logTheme.value = 'dark'
}

const handleClose = () => {
  visible.value = false
}

watch(
  () => visible.value,
  v => {
    if (!v) {
      resetOnClose()
      return
    }
    if (props.containerSelectable && normalizedContainers.value.length > 0 && !localContainerName.value && !props.allowAllContainers) {
      localContainerName.value = normalizedContainers.value[0]
    }
    fetchPodLogs()
  }
)

watch(
  () => normalizedContainers.value,
  containers => {
    if (!visible.value) return
    if (!props.containerSelectable) return
    if (containers.length === 0) return
    if (localContainerName.value) return
    if (!props.allowAllContainers) {
      localContainerName.value = containers[0]
    }
  }
)
</script>

<style scoped>
.log-header {
  margin-bottom: 16px;
  border-bottom: 1px solid #e4e7ed;
  padding-bottom: 16px;
}

.log-container {
  border-radius: 8px;
  overflow: hidden;
  min-height: 420px;
}

.log-content {
  margin: 0;
  padding: 12px;
  white-space: pre-wrap;
  word-break: break-word;
  height: 520px;
  overflow: auto;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

.dialog-footer {
  text-align: right;
}

.real-time-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  font-size: 12px;
  color: #e5e7eb;
  background: rgba(0, 0, 0, 0.25);
}

.log-search-count {
  margin-left: 8px;
  font-size: 12px;
  color: #909399;
}

.log-content :deep(.log-search-hit) {
  padding: 0 2px;
  border-radius: 3px;
  background: rgba(250, 204, 21, 0.2);
}

.log-content :deep(.log-search-hit[data-log-active="1"]) {
  background: rgba(250, 204, 21, 0.55);
}

.blinking {
  animation: blink 1.2s infinite;
}

@keyframes blink {
  0% { opacity: 1; }
  50% { opacity: 0.35; }
  100% { opacity: 1; }
}
</style>
