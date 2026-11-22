<template>
  <el-dialog
    :model-value="modelValue"
    title="管理标签"
    width="860px"
    :close-on-click-modal="false"
    class="config-dialog form-dialog--compact"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @open="load"
  >
    <template #header>
      <div class="dialog-header">
        <h3 class="dialog-title">管理标签 · {{ hostname }}</h3>
      </div>
    </template>
    <div v-loading="loading" class="config-editor">
      <div class="config-content">
        <p class="nl-tip">标签用于节点选择器和亲和性调度，例如给 GPU 节点加上 <code>gpu=true</code>，再在工作负载的节点选择器中使用。</p>
        <p class="fd-subtitle">常用标签</p>
        <div class="nl-presets">
          <button v-for="p in PRESETS" :key="p.key" type="button" class="nl-preset" :title="`${p.key}=${p.value}`" @click="addPreset(p.key, p.value)">
            <el-icon><Plus /></el-icon>
            {{ p.label }}
          </button>
        </div>
        <KVRowsTable v-model="rows" title="标签" add-text="添加标签" empty-text="暂无自定义标签" />
        <template v-if="systemRows.length">
          <p class="fd-subtitle">系统标签<small>由 kubelet / 云厂商组件维护，修改会被覆盖，这里只读</small></p>
          <div class="nl-system">
            <el-tag v-for="r in systemRows" :key="r.key" type="info" class="mono">{{ r.key }}={{ r.value }}</el-tag>
          </div>
        </template>
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="emit('update:modelValue', false)">取消</el-button>
        <el-button type="primary" :loading="saving" :disabled="loading" @click="save">保存</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import KVRowsTable from '@/components/KVRowsTable.vue'
import { nodeApi } from '@/api/node'
import { isKubeletLabel } from '@/utils/node'
import type { KVRow } from '@/utils/job'

const props = defineProps<{ modelValue: boolean; clusterId: string; hostname: string }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; success: [] }>()

const PRESETS = [
  { label: '工作节点角色', key: 'node-role.kubernetes.io/worker', value: '' },
  { label: '基础设施节点角色', key: 'node-role.kubernetes.io/infra', value: '' },
  { label: '边缘节点角色', key: 'node-role.kubernetes.io/edge', value: '' },
  { label: 'GPU 节点', key: 'gpu', value: 'true' },
  { label: 'SSD 磁盘', key: 'disktype', value: 'ssd' }
]

const loading = ref(false)
const saving = ref(false)
const rows = ref<KVRow[]>([])
const systemRows = ref<KVRow[]>([])

// 打开时读取最新的标签
const load = async () => {
  rows.value = []
  systemRows.value = []
  loading.value = true
  try {
    const { data } = await nodeApi.getDetail({ clusterId: props.clusterId, hostname: props.hostname })
    Object.entries(data?.labels || {})
      .sort(([a], [b]) => a.localeCompare(b))
      .forEach(([key, value]) => (isKubeletLabel(key) ? systemRows.value : rows.value).push({ key, value }))
  } catch {
    emit('update:modelValue', false)
  } finally {
    loading.value = false
  }
}

const addPreset = (key: string, value: string) => {
  if (rows.value.some(r => r.key === key)) {
    ElMessage.info('该标签已存在')
    return
  }
  rows.value.push({ key, value })
}

// 键: [前缀/]名称; 值: 最多 63 个字符
const KEY = /^([a-z0-9]([-a-z0-9.]*[a-z0-9])?\/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$/
const VALUE = /^([A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?)?$/

const save = async () => {
  const labels: Record<string, string> = {}
  for (const r of rows.value) {
    const key = r.key.trim()
    if (!key) continue
    if (!KEY.test(key)) {
      ElMessage.warning(`标签键 ${key} 不合法`)
      return
    }
    if (!VALUE.test(r.value.trim())) {
      ElMessage.warning(`标签 ${key} 的值不合法：最多 63 个字符，只能包含字母、数字、"-"、"_" 和 "."`)
      return
    }
    if (isKubeletLabel(key)) {
      ElMessage.warning(`${key} 由系统维护，不能修改`)
      return
    }
    labels[key] = r.value.trim()
  }
  saving.value = true
  try {
    await nodeApi.updateLabels({ clusterId: props.clusterId, hostname: props.hostname, labels })
    ElMessage.success('标签已更新')
    emit('update:modelValue', false)
    emit('success')
  } catch {
    // 请求拦截器已提示错误
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.nl-tip {
  margin: 0 0 14px;
  color: #64748b;
  font-size: 13px;
  line-height: 1.6;
}

.nl-tip code {
  padding: 1px 6px;
  border-radius: 4px;
  background: #f1f5f9;
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}

.nl-presets {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 16px;
}

.nl-preset {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 30px;
  padding: 0 12px;
  border: 1px dashed #c9d6ea;
  border-radius: 15px;
  background: #fff;
  color: #475569;
  font-size: 13px;
  cursor: pointer;
}

.nl-preset:hover {
  border-color: #3b74f2;
  color: #3b74f2;
}

.nl-system {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.mono {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}
</style>
