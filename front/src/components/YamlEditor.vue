<template>
  <Teleport to="body" :disabled="!isFullscreen">
    <div v-if="isFullscreen" class="ye-backdrop" @click="exitFullscreen" />
    <div
      v-bind="$attrs"
      ref="rootRef"
      class="yaml-editor"
      :class="[`is-${theme}`, { 'is-fullscreen': isFullscreen }]"
      :style="rootStyle"
      @keydown="onRootKeydown"
    >
      <!-- 工具栏 -->
      <div class="ye-toolbar">
        <div class="ye-toolbar__left">
          <span class="ye-lang">YAML</span>
          <span class="ye-title" :title="title">{{ title }}</span>
          <span v-if="readonly" class="ye-chip">
            <el-icon><Lock /></el-icon>
            只读
          </span>
        </div>
        <div class="ye-toolbar__right">
          <button
            v-for="action in actions"
            :key="action.key"
            type="button"
            class="ye-btn"
            :class="{ 'is-done': action.key === 'copy' && copied }"
            :data-tip="action.label"
            :aria-label="action.label"
            @click="action.run"
          >
            <el-icon><component :is="action.icon" /></el-icon>
          </button>
        </div>
      </div>

      <!-- 编辑器 -->
      <div class="ye-body">
        <codemirror
          v-model="content"
          :style="{ height: '100%' }"
          :autofocus="false"
          :indent-with-tab="!readonly"
          :tab-size="2"
          :extensions="extensions"
          @ready="handleReady"
          @update="handleUpdate"
        />
      </div>

      <!-- 状态栏 -->
      <div class="ye-status">
        <div class="ye-status__left">
          <span>第 {{ cursor.line }} 行，第 {{ cursor.col }} 列</span>
          <span>共 {{ lineCount }} 行</span>
          <span>{{ sizeText }}</span>
        </div>
        <div class="ye-status__right">
          <span class="ye-syntax" :class="syntaxErrorLine ? 'is-error' : 'is-ok'">
            <i class="ye-syntax__dot" />
            {{ syntaxErrorLine ? `第 ${syntaxErrorLine} 行附近可能有语法错误` : '语法正确' }}
          </span>
          <span>空格: 2</span>
          <span>UTF-8</span>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, shallowRef, watch, type Component } from 'vue'
import { ElMessage } from 'element-plus'
import { Check, Close, DocumentCopy, Download, Expand, Fold, FullScreen, Lock, Search } from '@element-plus/icons-vue'
import { Codemirror } from 'vue-codemirror'
import { EditorState, type Extension } from '@codemirror/state'
import {
  EditorView,
  ViewUpdate,
  crosshairCursor,
  drawSelection,
  dropCursor,
  highlightActiveLine,
  highlightActiveLineGutter,
  highlightSpecialChars,
  keymap,
  lineNumbers,
  rectangularSelection
} from '@codemirror/view'
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import {
  HighlightStyle,
  bracketMatching,
  ensureSyntaxTree,
  foldAll,
  foldGutter,
  foldKeymap,
  indentOnInput,
  syntaxHighlighting,
  syntaxTree,
  unfoldAll
} from '@codemirror/language'
import { highlightSelectionMatches, openSearchPanel, searchKeymap } from '@codemirror/search'
import { yaml } from '@codemirror/lang-yaml'
import { tags as t } from '@lezer/highlight'

// 根节点包在 Teleport 里(全屏时移到 body), 外部传入的 class / style 手动绑定到编辑器根元素
defineOptions({ inheritAttrs: false })

interface Props {
  modelValue: string
  title?: string
  readonly?: boolean
  height?: string
  theme?: 'light' | 'dark'
  filename?: string
}

const props = withDefaults(defineProps<Props>(), {
  title: 'YAML配置',
  readonly: true,
  theme: 'dark',
  filename: 'config.yaml'
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  change: [value: string]
}>()

const content = ref(props.modelValue ?? '')
const editorView = shallowRef<EditorView | null>(null)
const isFullscreen = ref(false)
const copied = ref(false)
const cursor = ref({ line: 1, col: 1 })
const syntaxErrorLine = ref(0)
const sizeBytes = ref(0)

// ---------------- 尺寸 ----------------
const lineCount = computed(() => (content.value ? content.value.split('\n').length : 1))

// 未指定高度时按行数自适应(含工具栏和状态栏)
const autoHeight = computed(() => {
  const estimated = lineCount.value * 22 + 24 + 80
  return `${Math.max(280, Math.min(600, estimated))}px`
})

const rootStyle = computed(() => (isFullscreen.value ? undefined : { height: props.height || autoHeight.value }))

const sizeText = computed(() => {
  const size = sizeBytes.value
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(2)} MB`
})

// ---------------- 主题 ----------------
const palettes = {
  dark: {
    bg: '#151a24',
    gutter: '#151a24',
    gutterText: '#4a5367',
    gutterActive: '#9aa7c0',
    activeLine: '#1b2231',
    selection: 'rgba(91, 156, 255, 0.28)',
    cursor: '#7cb3ff',
    text: '#d5dbe6',
    key: '#7aa2f7',
    value: '#9ece6a',
    string: '#a5d6a7',
    comment: '#5c6680',
    punct: '#89ddff',
    label: '#e0af68',
    keyword: '#bb9af7',
    match: 'rgba(224, 175, 104, 0.22)',
    panel: '#1b2130',
    panelBorder: '#262e40',
    input: '#121722'
  },
  light: {
    bg: '#fbfcfe',
    gutter: '#f5f7fb',
    gutterText: '#a3adbd',
    gutterActive: '#334155',
    activeLine: '#eef4ff',
    selection: 'rgba(59, 116, 242, 0.18)',
    cursor: '#3b74f2',
    text: '#1e293b',
    key: '#2563eb',
    value: '#15803d',
    string: '#16a34a',
    comment: '#94a3b8',
    punct: '#0891b2',
    label: '#b45309',
    keyword: '#7c3aed',
    match: 'rgba(245, 158, 11, 0.2)',
    panel: '#f3f6fb',
    panelBorder: '#e2e8f0',
    input: '#ffffff'
  }
}

const buildTheme = (mode: 'light' | 'dark'): Extension => {
  const c = palettes[mode]
  const editorTheme = EditorView.theme(
    {
      '&': { height: '100%', color: c.text, backgroundColor: c.bg, fontSize: '14px' },
      '&.cm-focused': { outline: 'none' },
      '.cm-scroller': {
        fontFamily: "'JetBrains Mono', 'SF Mono', Menlo, Consolas, 'Liberation Mono', monospace",
        lineHeight: '22px'
      },
      '.cm-content': { padding: '12px 0', caretColor: c.cursor },
      '.cm-line': { padding: '0 20px 0 10px' },
      '.cm-cursor, .cm-dropCursor': { borderLeftColor: c.cursor, borderLeftWidth: '2px' },
      '&.cm-focused > .cm-scroller > .cm-selectionLayer .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection': {
        backgroundColor: c.selection
      },
      '.cm-activeLine': { backgroundColor: c.activeLine },
      '.cm-gutters': { backgroundColor: c.gutter, color: c.gutterText, border: 'none', paddingLeft: '6px' },
      '.cm-activeLineGutter': { backgroundColor: 'transparent', color: c.gutterActive },
      '.cm-lineNumbers .cm-gutterElement': { minWidth: '36px', padding: '0 8px 0 0', fontSize: '13px' },
      '.cm-foldGutter .cm-gutterElement': { width: '14px', color: c.gutterText, cursor: 'pointer' },
      '.cm-foldGutter .cm-gutterElement:hover': { color: c.gutterActive },
      '.cm-foldPlaceholder': { backgroundColor: c.activeLine, border: 'none', color: c.punct, padding: '0 6px', borderRadius: '4px' },
      '.cm-selectionMatch': { backgroundColor: c.match },
      '.cm-matchingBracket, &.cm-focused .cm-matchingBracket': { backgroundColor: c.match, outline: 'none' },
      '.cm-searchMatch': { backgroundColor: c.match, outline: `1px solid ${c.label}` },
      '.cm-searchMatch.cm-searchMatch-selected': { backgroundColor: 'rgba(91, 156, 255, 0.35)' },
      // 搜索面板
      '.cm-panels': { backgroundColor: c.panel, color: c.text, borderColor: c.panelBorder },
      '.cm-panels.cm-panels-top': { borderBottom: `1px solid ${c.panelBorder}` },
      '.cm-panels.cm-panels-bottom': { borderTop: `1px solid ${c.panelBorder}` },
      '.cm-search': { padding: '8px 12px', fontSize: '12px' },
      '.cm-search label': { color: c.gutterActive },
      '.cm-textfield': {
        backgroundColor: c.input,
        color: c.text,
        border: `1px solid ${c.panelBorder}`,
        borderRadius: '6px',
        padding: '3px 8px',
        outline: 'none'
      },
      '.cm-textfield:focus': { borderColor: c.cursor },
      '.cm-button': {
        backgroundImage: 'none',
        backgroundColor: c.input,
        color: c.text,
        border: `1px solid ${c.panelBorder}`,
        borderRadius: '6px',
        padding: '3px 10px'
      },
      '.cm-button:hover': { borderColor: c.cursor, color: c.cursor },
      '.cm-panel.cm-search [name=close]': { color: c.gutterActive, fontSize: '16px', right: '8px' },
      '.cm-tooltip': { backgroundColor: c.panel, border: `1px solid ${c.panelBorder}`, color: c.text }
    },
    { dark: mode === 'dark' }
  )
  const highlight = HighlightStyle.define([
    { tag: [t.definition(t.propertyName), t.propertyName], color: c.key, fontWeight: '500' },
    { tag: t.content, color: c.value },
    { tag: [t.string, t.special(t.string)], color: c.string },
    { tag: [t.number, t.bool, t.null, t.atom], color: c.label },
    { tag: [t.lineComment, t.blockComment, t.comment], color: c.comment, fontStyle: 'italic' },
    { tag: [t.separator, t.punctuation, t.squareBracket, t.brace], color: c.punct },
    { tag: [t.labelName, t.typeName, t.meta], color: c.label },
    { tag: [t.keyword, t.processingInstruction], color: c.keyword },
    { tag: t.invalid, color: '#f87171', textDecoration: 'underline wavy' }
  ])
  return [editorTheme, syntaxHighlighting(highlight)]
}

// 搜索面板等内置文案
const zhPhrases = EditorState.phrases.of({
  Find: '查找',
  Replace: '替换',
  next: '下一个',
  previous: '上一个',
  all: '全部',
  'match case': '区分大小写',
  regexp: '正则',
  'by word': '全词匹配',
  replace: '替换',
  'replace all': '全部替换',
  close: '关闭',
  'Go to line': '跳转到行',
  go: '跳转',
  'Fold line': '折叠',
  'Unfold line': '展开',
  'folded code': '已折叠'
})

const extensions = computed<Extension[]>(() => [
  lineNumbers(),
  highlightActiveLineGutter(),
  foldGutter({ openText: '▾', closedText: '▸' }),
  highlightSpecialChars(),
  history(),
  drawSelection(),
  dropCursor(),
  EditorState.allowMultipleSelections.of(true),
  indentOnInput(),
  bracketMatching(),
  rectangularSelection(),
  crosshairCursor(),
  highlightActiveLine(),
  highlightSelectionMatches(),
  EditorView.lineWrapping,
  // 只读时仍可选中、复制、搜索, 只是不能修改
  EditorState.readOnly.of(props.readonly),
  keymap.of([...defaultKeymap, ...searchKeymap, ...historyKeymap, ...foldKeymap]),
  yaml(),
  zhPhrases,
  buildTheme(props.theme)
])

// ---------------- 状态 ----------------
let syntaxTimer: ReturnType<typeof setTimeout> | undefined

const refreshStats = (state: EditorState) => {
  // 尽量解析完整文档; 文档很大时最多等 200ms, 剩余部分按已解析的结果判断
  const tree = ensureSyntaxTree(state, state.doc.length, 200) ?? syntaxTree(state)
  let errorPos = -1
  tree.iterate({
    enter: node => {
      if (errorPos >= 0) return false
      if (node.type.isError) {
        errorPos = node.from
        return false
      }
    }
  })
  syntaxErrorLine.value = errorPos >= 0 && state.doc.length > 0 ? state.doc.lineAt(errorPos).number : 0
  sizeBytes.value = new TextEncoder().encode(state.doc.toString()).length
}

const scheduleStats = (state: EditorState) => {
  clearTimeout(syntaxTimer)
  syntaxTimer = setTimeout(() => refreshStats(state), 250)
}

const updateCursor = (state: EditorState) => {
  const pos = state.selection.main.head
  const line = state.doc.lineAt(pos)
  cursor.value = { line: line.number, col: pos - line.from + 1 }
}

const handleReady = ({ view }: { view: EditorView }) => {
  editorView.value = view
  updateCursor(view.state)
  refreshStats(view.state)
}

const handleUpdate = (update: ViewUpdate) => {
  if (update.selectionSet || update.docChanged) updateCursor(update.state)
  if (update.docChanged) scheduleStats(update.state)
}

watch(
  () => props.modelValue,
  value => {
    if (value !== content.value) content.value = value ?? ''
  }
)

watch(content, value => {
  emit('update:modelValue', value)
  emit('change', value)
})

// ---------------- 工具栏操作 ----------------
let copiedTimer: ReturnType<typeof setTimeout> | undefined

const handleCopy = async () => {
  try {
    await navigator.clipboard.writeText(content.value)
    copied.value = true
    clearTimeout(copiedTimer)
    copiedTimer = setTimeout(() => (copied.value = false), 1500)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败，请手动选择内容复制')
  }
}

const handleDownload = () => {
  const blob = new Blob([content.value], { type: 'text/yaml;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = props.filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

const openSearch = () => {
  if (editorView.value) openSearchPanel(editorView.value)
}

const foldAllCode = () => {
  if (editorView.value) foldAll(editorView.value)
}

const unfoldAllCode = () => {
  if (editorView.value) unfoldAll(editorView.value)
}

const rootRef = ref<HTMLElement>()

// 全屏时按 Esc: 搜索面板打开时先关闭搜索面板(CodeMirror 会 preventDefault), 否则退出全屏.
// 都要阻止事件继续冒泡, 否则外层 el-dialog 也会响应 Esc 被关闭
const onRootKeydown = (e: KeyboardEvent) => {
  if (!isFullscreen.value || e.key !== 'Escape') return
  e.stopPropagation()
  if (!e.defaultPrevented) exitFullscreen()
}

// 焦点不在编辑器内时(例如点了遮罩), 在捕获阶段处理 Esc
const onWindowEscape = (e: KeyboardEvent) => {
  if (e.key !== 'Escape' || rootRef.value?.contains(e.target as Node)) return
  e.stopPropagation()
  exitFullscreen()
}

const handleFullscreen = () => {
  isFullscreen.value = true
}

const exitFullscreen = () => {
  isFullscreen.value = false
}

watch(isFullscreen, value => {
  if (value) window.addEventListener('keydown', onWindowEscape, true)
  else window.removeEventListener('keydown', onWindowEscape, true)
  // DOM 移动后让编辑器重新测量尺寸
  requestAnimationFrame(() => editorView.value?.requestMeasure())
})

const actions = computed<Array<{ key: string; label: string; icon: Component; run: () => void }>>(() => [
  { key: 'search', label: '查找 (Ctrl+F)', icon: Search, run: openSearch },
  { key: 'fold', label: '全部折叠', icon: Fold, run: foldAllCode },
  { key: 'unfold', label: '全部展开', icon: Expand, run: unfoldAllCode },
  { key: 'copy', label: copied.value ? '已复制' : '复制', icon: copied.value ? Check : DocumentCopy, run: handleCopy },
  { key: 'download', label: '下载', icon: Download, run: handleDownload },
  isFullscreen.value
    ? { key: 'fullscreen', label: '退出全屏 (Esc)', icon: Close, run: exitFullscreen }
    : { key: 'fullscreen', label: '全屏', icon: FullScreen, run: handleFullscreen }
])

onBeforeUnmount(() => {
  clearTimeout(syntaxTimer)
  clearTimeout(copiedTimer)
  window.removeEventListener('keydown', onWindowEscape, true)
})

defineExpose({
  foldAll: foldAllCode,
  unfoldAll: unfoldAllCode,
  copy: handleCopy,
  download: handleDownload,
  fullscreen: handleFullscreen,
  search: openSearch
})
</script>

<style scoped>
.yaml-editor {
  --ye-bg: #151a24;
  --ye-bar: #1a2030;
  --ye-border: #252d3f;
  --ye-text: #c9d2e3;
  --ye-muted: #6b7690;
  --ye-hover: #242c3f;
  --ye-accent: #7cb3ff;
  display: flex;
  flex-direction: column;
  min-width: 0;
  border: 1px solid var(--ye-border);
  border-radius: 12px;
  background: var(--ye-bg);
  overflow: hidden;
}

.yaml-editor.is-light {
  --ye-bg: #fbfcfe;
  --ye-bar: #f3f6fb;
  --ye-border: #e2e8f0;
  --ye-text: #334155;
  --ye-muted: #94a3b8;
  --ye-hover: #e8eef8;
  --ye-accent: #3b74f2;
}

/* ---------- 工具栏 ---------- */
.ye-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex: none;
  height: 44px;
  padding: 0 8px 0 14px;
  border-bottom: 1px solid var(--ye-border);
  background: var(--ye-bar);
}

.ye-toolbar__left {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.ye-lang {
  flex: none;
  padding: 1px 7px;
  border-radius: 5px;
  background: linear-gradient(135deg, #5b9cff 0%, #3b74f2 100%);
  color: #ffffff;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.5px;
  line-height: 18px;
}

.ye-title {
  overflow: hidden;
  color: var(--ye-text);
  font-size: 14px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ye-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  flex: none;
  padding: 1px 8px;
  border: 1px solid var(--ye-border);
  border-radius: 999px;
  color: var(--ye-muted);
  font-size: 12px;
  line-height: 18px;
}

.ye-toolbar__right {
  display: flex;
  align-items: center;
  gap: 2px;
}

.ye-btn {
  position: relative;
  display: grid;
  place-items: center;
  width: 32px;
  height: 32px;
  padding: 0;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--ye-muted);
  font-size: 16px;
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease;
}

.ye-btn:hover {
  background: var(--ye-hover);
  color: var(--ye-accent);
}

.ye-btn:focus-visible {
  outline: 2px solid var(--ye-accent);
  outline-offset: 1px;
}

.ye-btn.is-done {
  color: #34d399;
}

/* 按钮提示: 纯 CSS, 全屏时也不会被遮挡 */
.ye-btn[data-tip]::after {
  content: attr(data-tip);
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  z-index: 10;
  padding: 4px 8px;
  border-radius: 6px;
  background: #0b0f17;
  color: #e2e8f0;
  font-size: 12px;
  line-height: 16px;
  white-space: nowrap;
  opacity: 0;
  pointer-events: none;
  transform: translateY(-2px);
  transition: opacity 0.15s ease 0.3s, transform 0.15s ease 0.3s;
}

.ye-btn[data-tip]:hover::after {
  opacity: 1;
  transform: none;
}

/* ---------- 编辑区 ---------- */
.ye-body {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.ye-body :deep(.cm-editor) {
  flex: 1;
  min-height: 0;
}

.ye-body :deep(.cm-scroller) {
  scrollbar-width: thin;
  scrollbar-color: var(--ye-border) transparent;
}

.ye-body :deep(.cm-scroller::-webkit-scrollbar) {
  width: 10px;
  height: 10px;
}

.ye-body :deep(.cm-scroller::-webkit-scrollbar-thumb) {
  border: 3px solid transparent;
  border-radius: 10px;
  background: var(--ye-border) padding-box;
}

/* ---------- 状态栏 ---------- */
.ye-status {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex: none;
  height: 30px;
  padding: 0 16px;
  border-top: 1px solid var(--ye-border);
  background: var(--ye-bar);
  color: var(--ye-muted);
  font-size: 12px;
  white-space: nowrap;
}

.ye-status__left,
.ye-status__right {
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
  overflow: hidden;
}

.ye-syntax {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.ye-syntax__dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.ye-syntax.is-ok .ye-syntax__dot {
  background: #34d399;
  box-shadow: 0 0 0 3px rgba(52, 211, 153, 0.15);
}

.ye-syntax.is-error {
  color: #f87171;
}

.ye-syntax.is-error .ye-syntax__dot {
  background: #f87171;
  box-shadow: 0 0 0 3px rgba(248, 113, 113, 0.18);
}

/* ---------- 全屏 ---------- */
.ye-backdrop {
  position: fixed;
  inset: 0;
  z-index: 2999;
  background: rgba(8, 12, 20, 0.72);
  backdrop-filter: blur(2px);
}

.yaml-editor.is-fullscreen {
  position: fixed;
  inset: 24px;
  z-index: 3000;
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.45);
}

@media (max-width: 768px) {
  .yaml-editor.is-fullscreen {
    inset: 8px;
  }

  .ye-status__left span:nth-child(n + 2),
  .ye-status__right span:nth-child(n + 2) {
    display: none;
  }
}
</style>
