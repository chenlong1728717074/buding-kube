<template>
  <el-dialog
    :model-value="modelValue"
    title="排空节点"
    width="1000px"
    top="6vh"
    :close-on-click-modal="false"
    :close-on-press-escape="!draining"
    :show-close="!draining"
    class="config-dialog form-dialog--compact"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @open="reset"
  >
    <template #header>
      <div class="dialog-header">
        <h3 class="dialog-title">排空节点 · {{ hostname }}</h3>
      </div>
    </template>
    <div class="config-editor">
      <div class="config-content">
        <!-- 结果 -->
        <template v-if="result">
          <el-result
            :icon="result.failed.length ? 'warning' : 'success'"
            :title="result.failed.length ? '部分容器组驱逐失败' : '排空完成'"
            :sub-title="`已驱逐 ${result.evicted.length} 个，失败 ${result.failed.length} 个，跳过 ${result.skipped} 个；节点已停止调度`"
          />
          <el-table v-if="result.failed.length" :data="result.failed" style="width: 100%" :header-cell-class-name="tableHeaderClass">
            <el-table-column label="容器组" min-width="260">
              <template #default="{ row }"><span class="mono">{{ row.namespace }}/{{ row.name }}</span></template>
            </el-table-column>
            <el-table-column label="失败原因" min-width="360" prop="reason" />
          </el-table>
          <p class="nd-tip">驱逐只是发起删除，容器组会按优雅终止时间退出，并由所属的工作负载在其他节点重建。维护完成后记得「启用调度」。</p>
        </template>

        <template v-else>
          <el-alert type="warning" show-icon :closable="false" class="nd-alert">
            排空会先停止节点调度，再通过 Eviction API 驱逐节点上的容器组（遵守 PodDisruptionBudget，与 kubectl drain 相同），常用于节点维护或下线。DaemonSet 和静态容器组会跳过。
          </el-alert>
          <el-form :model="opts" label-width="150px" class="nd-form">
            <el-form-item label="删除临时存储数据">
              <el-switch v-model="opts.deleteEmptyDirData" @change="loadPreview" />
              <span class="fd-inline-hint">驱逐使用 emptyDir 的容器组，emptyDir 中的数据会丢失</span>
            </el-form-item>
            <el-form-item label="驱逐单独创建的容器组">
              <el-switch v-model="opts.force" @change="loadPreview" />
              <span class="fd-inline-hint">没有工作负载管理的容器组驱逐后不会重建</span>
            </el-form-item>
            <el-form-item label="优雅终止时间">
              <el-radio-group v-model="graceMode">
                <el-radio-button value="default">使用容器组的设置</el-radio-button>
                <el-radio-button value="custom">自定义</el-radio-button>
              </el-radio-group>
              <el-input-number v-if="graceMode === 'custom'" v-model="opts.gracePeriodSeconds" :min="0" :max="3600" controls-position="right" style="margin-left: 12px" />
              <span v-if="graceMode === 'custom'" class="fd-inline-hint">秒</span>
            </el-form-item>
            <el-form-item label="超时时间">
              <el-input-number v-model="opts.timeoutSeconds" :min="30" :max="600" :step="30" controls-position="right" />
              <span class="fd-inline-hint">秒，受 PodDisruptionBudget 限制的容器组会在这段时间内重试</span>
            </el-form-item>
          </el-form>

          <div class="nd-summary">
            <span class="nd-chip is-evict">将驱逐 {{ preview?.evict ?? 0 }}</span>
            <span class="nd-chip is-skip">跳过 {{ preview?.skip ?? 0 }}</span>
            <span :class="['nd-chip', 'is-block', { 'is-zero': !preview?.blocked }]">需要处理 {{ preview?.blocked ?? 0 }}</span>
          </div>
          <el-table v-loading="previewLoading" :data="preview?.pods || []" max-height="340" style="width: 100%" :header-cell-class-name="tableHeaderClass">
            <el-table-column label="容器组" min-width="260" show-overflow-tooltip>
              <template #default="{ row }"><span class="mono">{{ row.namespace }}/{{ row.name }}</span></template>
            </el-table-column>
            <el-table-column label="所属工作负载" min-width="180" show-overflow-tooltip>
              <template #default="{ row }">{{ row.ownerName ? `${row.ownerKind} ${row.ownerName}` : '无' }}</template>
            </el-table-column>
            <el-table-column label="处理方式" width="110">
              <template #default="{ row }">
                <el-tag :type="ACTION[row.action as DrainPodVO['action']].tag" size="small">{{ ACTION[row.action as DrainPodVO['action']].text }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="说明" min-width="300">
              <template #default="{ row }"><span :class="{ 'nd-block': row.action === 'block' }">{{ row.reason || '-' }}</span></template>
            </el-table-column>
            <template #empty>
              <span class="nd-muted">节点上没有容器组</span>
            </template>
          </el-table>
        </template>
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <template v-if="result">
          <el-button type="primary" @click="emit('update:modelValue', false)">完成</el-button>
        </template>
        <template v-else>
          <el-button :disabled="draining" @click="emit('update:modelValue', false)">取消</el-button>
          <el-button type="danger" :loading="draining" :disabled="previewLoading || !preview || preview.blocked > 0" @click="submit">
            {{ draining ? '正在驱逐…' : '停止调度并排空' }}
          </el-button>
        </template>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { nodeApi, type DrainPodVO, type DrainPreviewVO, type DrainResultVO } from '@/api/node'
import { tableHeaderClass } from '@/utils/table'

const props = defineProps<{ modelValue: boolean; clusterId: string; hostname: string }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; success: [] }>()

const ACTION: Record<DrainPodVO['action'], { text: string; tag: 'primary' | 'info' | 'danger' | 'warning' }> = {
  evict: { text: '驱逐', tag: 'primary' },
  delete: { text: '删除', tag: 'warning' },
  skip: { text: '跳过', tag: 'info' },
  block: { text: '需要处理', tag: 'danger' }
}

const opts = reactive({ deleteEmptyDirData: false, force: false, gracePeriodSeconds: 30, timeoutSeconds: 120 })
const graceMode = ref<'default' | 'custom'>('default')
const preview = ref<DrainPreviewVO | null>(null)
const previewLoading = ref(false)
const draining = ref(false)
const result = ref<DrainResultVO | null>(null)

const params = () => ({
  clusterId: props.clusterId,
  hostname: props.hostname,
  deleteEmptyDirData: opts.deleteEmptyDirData,
  force: opts.force,
  gracePeriodSeconds: graceMode.value === 'custom' ? opts.gracePeriodSeconds : -1,
  timeoutSeconds: opts.timeoutSeconds
})

let seq = 0
const loadPreview = async () => {
  const id = ++seq
  previewLoading.value = true
  try {
    const resp = await nodeApi.drainPreview(params())
    if (id === seq) preview.value = resp.data
  } catch {
    if (id === seq) preview.value = null
  } finally {
    if (id === seq) previewLoading.value = false
  }
}

const reset = () => {
  Object.assign(opts, { deleteEmptyDirData: false, force: false, gracePeriodSeconds: 30, timeoutSeconds: 120 })
  graceMode.value = 'default'
  preview.value = null
  result.value = null
  draining.value = false
  loadPreview()
}

const submit = async () => {
  try {
    await ElMessageBox.confirm(
      `将停止节点 ${props.hostname} 的调度，并驱逐 ${preview.value?.evict ?? 0} 个容器组。确定继续吗？`,
      '确认排空',
      { type: 'warning', confirmButtonText: '排空', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  draining.value = true
  try {
    const resp = await nodeApi.drain(params())
    result.value = resp.data
    emit('success')
  } catch {
    // 请求拦截器已提示错误(例如有容器组需要处理), 刷新预览
    loadPreview()
  } finally {
    draining.value = false
  }
  if (result.value && !result.value.failed.length) ElMessage.success('排空完成')
}
</script>

<style scoped>
.nd-alert {
  margin-bottom: 16px;
}

.nd-form {
  margin-bottom: 8px;
}

.nd-summary {
  display: flex;
  gap: 8px;
  margin-bottom: 10px;
}

.nd-chip {
  padding: 3px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.nd-chip.is-evict {
  background: #eef4ff;
  color: #3b74f2;
}

.nd-chip.is-skip {
  background: #f1f5f9;
  color: #64748b;
}

.nd-chip.is-block {
  background: #fef2f2;
  color: #ef4444;
}

.nd-chip.is-block.is-zero {
  background: #f0fdf4;
  color: #16a34a;
}

.nd-block {
  color: #ef4444;
}

.nd-muted {
  color: #94a3b8;
}

.nd-tip {
  margin: 14px 0 0;
  color: #64748b;
  font-size: 13px;
  line-height: 1.6;
}

.mono {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}
</style>
