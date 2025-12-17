<template>
  <div class="yaml-editor" :class="{ dark: props.theme === 'dark' }">
    <!-- 工具栏 -->
    <div class="editor-toolbar">
      <div class="toolbar-left">
        <el-text type="info" size="small">{{ title }}</el-text>
      </div>
      <div class="toolbar-right">
        <el-button-group size="small">
          <el-button @click="handleCopy" :icon="DocumentCopy" title="复制">
            复制
          </el-button>
          <el-button @click="handleDownload" :icon="Download" title="下载">
            下载
          </el-button>
          <el-button @click="handleFullscreen" :icon="FullScreen" title="全屏">
            全屏
          </el-button>
        </el-button-group>
      </div>
    </div>
    
    <!-- 编辑器 -->
    <div class="editor-container" :class="{ 'fullscreen': isFullscreen }">
      <codemirror
        v-model="content"
        :style="{ height: editorHeight }"
        :autofocus="false"
        :indent-with-tab="true"
        :tab-size="2"
        :extensions="extensions"
        :readonly="readonly"
        @ready="handleReady"
        @change="handleChange"
      />
    </div>
    
    <!-- 全屏遮罩 -->
    <div v-if="isFullscreen" class="fullscreen-overlay" @click="exitFullscreen">
      <div class="fullscreen-content" @click.stop>
        <div class="fullscreen-header">
          <span>{{ title }}</span>
          <el-button @click="exitFullscreen" :icon="Close" circle size="small" />
        </div>
        <codemirror
          v-model="content"
          style="height: calc(100vh - 120px);"
          :autofocus="false"
          :indent-with-tab="true"
          :tab-size="2"
          :extensions="extensions"
          :readonly="readonly"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { DocumentCopy, Download, FullScreen, Close } from '@element-plus/icons-vue'
import { Codemirror } from 'vue-codemirror'
import { yaml } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'
import { EditorView } from '@codemirror/view'
import { foldGutter, foldAll, unfoldAll } from '@codemirror/language'
import { lineNumbers } from '@codemirror/view'
import { searchKeymap } from '@codemirror/search'
import { keymap } from '@codemirror/view'

// Props
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
  height: '400px',
  theme: 'dark',
  filename: 'config.yaml'
})

// Emits
const emit = defineEmits<{
  'update:modelValue': [value: string]
  'change': [value: string]
}>()

// 响应式数据
const content = ref(props.modelValue)
const isFullscreen = ref(false)
const editorView = ref<EditorView | null>(null)

// 计算属性
const editorHeight = computed(() => props.height)

// CodeMirror扩展配置
const extensions = computed(() => {
  const baseExtensions = [
    yaml(), // YAML语法高亮
    lineNumbers(), // 行号
    foldGutter(), // 代码折叠
    EditorView.lineWrapping, // 自动换行
    keymap.of(searchKeymap), // 搜索快捷键
    EditorView.theme({
      '&': {
        fontSize: '14px',
        fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace'
      },
      '.cm-content': {
        padding: '12px',
        minHeight: '200px'
      },
      '.cm-focused': {
        outline: 'none'
      },
      '.cm-editor': {
        borderRadius: '6px'
      },
      '.cm-scroller': {
        fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace'
      },
      '.cm-foldGutter': {
        width: '16px'
      }
    })
  ]
  
  // 根据主题添加暗色主题
  if (props.theme === 'dark') {
    baseExtensions.push(oneDark)
  }
  
  return baseExtensions
})

// 监听props变化
watch(() => props.modelValue, (newValue) => {
  if (newValue !== content.value) {
    content.value = newValue
  }
})

// 监听内容变化
watch(content, (newValue) => {
  emit('update:modelValue', newValue)
  emit('change', newValue)
})

// 编辑器就绪
const handleReady = ({ view }: { view: EditorView }) => {
  editorView.value = view
}

// 内容变化
const handleChange = (value: string) => {
  content.value = value
}

// 复制内容
const handleCopy = async () => {
  try {
    await navigator.clipboard.writeText(content.value)
    ElMessage.success('已复制到剪贴板')
  } catch (error) {
    console.error('复制失败:', error)
    ElMessage.error('复制失败')
  }
}

// 下载文件
const handleDownload = () => {
  try {
    const blob = new Blob([content.value], { type: 'text/yaml;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = props.filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    ElMessage.success('下载成功')
  } catch (error) {
    console.error('下载失败:', error)
    ElMessage.error('下载失败')
  }
}

// 全屏显示
const handleFullscreen = () => {
  isFullscreen.value = true
}

// 退出全屏
const exitFullscreen = () => {
  isFullscreen.value = false
}

// 折叠所有代码
const foldAllCode = () => {
  if (editorView.value) {
    foldAll(editorView.value)
  }
}

// 展开所有代码
const unfoldAllCode = () => {
  if (editorView.value) {
    unfoldAll(editorView.value)
  }
}

// 暴露方法
defineExpose({
  foldAll: foldAllCode,
  unfoldAll: unfoldAllCode,
  copy: handleCopy,
  download: handleDownload,
  fullscreen: handleFullscreen
})
</script>

<style scoped>
.yaml-editor {
  border: 1px solid #e4e7ed;
  border-radius: var(--radius-md);
  overflow: hidden;
  background: #fff;
}

.editor-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  font-size: 12px;
}

.toolbar-left {
  display: flex;
  align-items: center;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.editor-container {
  position: relative;
}

.editor-container :deep(.cm-editor) {
  border: none;
  border-radius: 0;
}

.editor-container :deep(.cm-focused) {
  outline: none;
}

/* 全屏样式 */
.fullscreen-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
}

.fullscreen-content {
  width: 90vw;
  height: 90vh;
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.fullscreen-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #f8f9fa;
  border-bottom: 1px solid #e4e7ed;
  font-weight: 500;
}

/* 暗色主题适配 */
.yaml-editor.dark {
  background: #1e1e1e;
  border-color: #3c3c3c;
}

.yaml-editor.dark .editor-toolbar {
  background: #2d2d2d;
  border-bottom-color: #3c3c3c;
  color: #fff;
}

.yaml-editor.dark .fullscreen-content {
  background: #1e1e1e;
}

.yaml-editor.dark .fullscreen-header {
  background: #2d2d2d;
  border-bottom-color: #3c3c3c;
  color: #fff;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .fullscreen-content {
    width: 95vw;
    height: 95vh;
  }
  
  .toolbar-right .el-button span {
    display: none;
  }
}
</style>
