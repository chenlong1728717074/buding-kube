<template>
  <el-dialog
    :model-value="modelValue"
    :fullscreen="fullscreen"
    width="min(1320px, 94vw)"
    top="4vh"
    :show-close="false"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    append-to-body
    class="pod-terminal-dialog"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @opened="onOpened"
    @closed="fullscreen = false"
  >
    <template #header>
      <div class="pt-header">
        <div class="pt-title">
          <el-icon class="pt-title__icon"><Monitor /></el-icon>
          <span class="pt-title__text">终端</span>
          <span class="pt-title__target">{{ namespace }} / {{ podName }}</span>
          <TerminalStatusTag :status="status" />
        </div>
        <div class="pt-actions">
          <el-select v-model="container" size="small" placeholder="默认容器" style="width: 190px" :loading="containerLoading" @change="reconnect">
            <el-option v-for="c in containerOptions" :key="c.name" :label="c.init ? `${c.name}（边车）` : c.name" :value="c.name" />
          </el-select>
          <el-select v-model="shell" size="small" style="width: 120px" @change="reconnect">
            <el-option label="自动" value="auto" />
            <el-option label="bash" value="bash" />
            <el-option label="sh" value="sh" />
          </el-select>
          <TerminalToolbar :view="view" :fullscreen="fullscreen" @reconnect="reconnect" @fullscreen="toggleFullscreen" @close="emit('update:modelValue', false)" />
        </div>
      </div>
    </template>
    <!-- 每次打开重新创建终端, 关闭时断开连接 -->
    <TerminalView
      v-if="modelValue"
      ref="view"
      :build-url="buildUrl"
      :height="fullscreen ? 'calc(100vh - 92px)' : '66vh'"
      @status="(s: TerminalStatus) => (status = s)"
      @ready="(c: string) => (container = c)"
    />
  </el-dialog>
</template>

<script setup lang="ts">
import { nextTick, ref } from 'vue'
import { Monitor } from '@element-plus/icons-vue'
import TerminalView from '@/components/terminal/TerminalView.vue'
import type { TerminalStatus } from '@/utils/terminal'
import TerminalStatusTag from '@/components/terminal/TerminalStatusTag.vue'
import TerminalToolbar from '@/components/terminal/TerminalToolbar.vue'
import { podApi } from '@/api/pod'
import { buildWsUrl } from '@/utils/ws'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    clusterId: string
    namespace: string
    podName: string
    // 默认进入的容器, 不传时由后端选择(默认容器注解或第一个运行中的容器)
    container?: string
  }>(),
  { container: '' }
)

const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()

const view = ref<InstanceType<typeof TerminalView> | null>(null)
const status = ref<TerminalStatus>('closed')
const container = ref('')
const shell = ref<'auto' | 'bash' | 'sh'>('auto')
const containerOptions = ref<{ name: string; init: boolean }[]>([])
const containerLoading = ref(false)
const fullscreen = ref(false)

const buildUrl = () =>
  buildWsUrl('/terminal/pod', {
    clusterId: props.clusterId,
    namespace: props.namespace,
    name: props.podName,
    container: container.value,
    shell: shell.value
  })

const reconnect = () => view.value?.connect(`正在连接 ${container.value || '默认容器'}（${shell.value === 'auto' ? '自动选择 Shell' : shell.value}）...`)

const loadContainers = async () => {
  containerLoading.value = true
  try {
    const { data } = await podApi.getDetail({ clusterId: props.clusterId, namespace: props.namespace, name: props.podName })
    containerOptions.value = (data?.containerStatuses || []).filter(s => s.state === 'Running').map(s => ({ name: s.name, init: s.init }))
  } catch {
    containerOptions.value = []
  } finally {
    containerLoading.value = false
  }
  if (props.container && containerOptions.value.some(c => c.name === props.container)) container.value = props.container
}

const toggleFullscreen = async () => {
  fullscreen.value = !fullscreen.value
  await nextTick()
  view.value?.fit()
  view.value?.focus()
}

const onOpened = async () => {
  container.value = props.container || ''
  shell.value = 'auto'
  await loadContainers()
  reconnect()
}
</script>

<style>
/* 弹窗挂载在 body 下, 样式不能用 scoped */
.pod-terminal-dialog.el-dialog {
  padding: 0;
  overflow: hidden;
  border-radius: 12px;
  background: #0b1220;
}

.pod-terminal-dialog.el-dialog.is-fullscreen {
  border-radius: 0;
}

.pod-terminal-dialog .el-dialog__header {
  margin: 0;
  padding: 10px 14px;
  background: #111a2e;
  border-bottom: 1px solid #1e293b;
}

.pod-terminal-dialog .el-dialog__body {
  padding: 0;
}

.pod-terminal-dialog .pt-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.pod-terminal-dialog .pt-title {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  color: #e2e8f0;
}

.pod-terminal-dialog .pt-title__icon {
  color: #60a5fa;
  font-size: 18px;
}

.pod-terminal-dialog .pt-title__text {
  font-weight: 600;
}

.pod-terminal-dialog .pt-title__target {
  overflow: hidden;
  color: #94a3b8;
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  font-size: 12px;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.pod-terminal-dialog .pt-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.pod-terminal-dialog .pt-actions .el-button + .el-button {
  margin-left: 0;
}

</style>
