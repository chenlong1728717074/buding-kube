<template>
  <el-dialog
    :model-value="modelValue"
    title="管理污点"
    width="920px"
    :close-on-click-modal="false"
    class="config-dialog form-dialog--compact"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @open="load"
  >
    <template #header>
      <div class="dialog-header">
        <h3 class="dialog-title">管理污点 · {{ hostname }}</h3>
      </div>
    </template>
    <div v-loading="loading" class="config-editor">
      <div class="config-content">
        <p class="nt-tip">污点让节点拒绝不容忍它的容器组，常用于专用节点（例如只运行数据库）或控制节点。容器组需要在「容忍」中声明相同的键和效果才能调度上来。</p>
        <div class="nt-effects">
          <div v-for="e in TAINT_EFFECTS" :key="e.value" class="nt-effect">
            <strong>{{ e.label }}</strong>
            <span>{{ e.desc }}</span>
          </div>
        </div>

        <p class="fd-subtitle">污点</p>
        <div class="nt-table">
          <div class="nt-row nt-row--head">
            <span>键</span>
            <span>值</span>
            <span>效果</span>
            <span />
          </div>
          <div v-for="(t, i) in rows" :key="i" class="nt-row">
            <el-input v-model="t.key" placeholder="例如 dedicated" class="mono-input" />
            <el-input v-model="t.value" placeholder="可选，例如 database" class="mono-input" />
            <el-select v-model="t.effect">
              <el-option v-for="e in TAINT_EFFECTS" :key="e.value" :label="e.label" :value="e.value" />
            </el-select>
            <el-button link type="danger" @click="rows.splice(i, 1)"><el-icon><Delete /></el-icon></el-button>
          </div>
          <div v-if="!rows.length" class="nt-empty">没有自定义污点</div>
          <div class="nt-actions">
            <el-button size="small" @click="rows.push({ key: '', value: '', effect: 'NoSchedule' })">
              <el-icon><Plus /></el-icon>
              添加污点
            </el-button>
            <el-button v-if="!hasControlPlaneTaint && isControlPlane" size="small" @click="rows.push({ key: 'node-role.kubernetes.io/control-plane', value: '', effect: 'NoSchedule' })">
              <el-icon><Plus /></el-icon>
              控制节点不运行普通应用
            </el-button>
          </div>
        </div>
        <el-alert v-if="newNoExecute.length" type="warning" show-icon :closable="false" class="nt-alert">
          新增的 NoExecute 污点 {{ newNoExecute.join('、') }} 会立即驱逐节点上不容忍它的容器组。
        </el-alert>
        <el-alert v-if="removedControlPlane" type="info" show-icon :closable="false" class="nt-alert">
          去掉控制节点的 NoSchedule 污点后，普通应用也会调度到控制节点上。
        </el-alert>

        <template v-if="systemTaints.length">
          <p class="fd-subtitle">系统污点<small>由节点控制器维护（例如节点未就绪、停止调度），这里只读</small></p>
          <div class="nt-system">
            <el-tag v-for="t in systemTaints" :key="t.key + t.effect" type="warning" class="mono">{{ taintText(t) }}</el-tag>
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
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import { nodeApi, type NodeTaintVO } from '@/api/node'
import { TAINT_EFFECTS, taintText } from '@/utils/node'

const props = defineProps<{ modelValue: boolean; clusterId: string; hostname: string }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean]; success: [] }>()

interface TaintRow {
  key: string
  value: string
  effect: string
}

const CONTROL_PLANE_KEYS = ['node-role.kubernetes.io/control-plane', 'node-role.kubernetes.io/master']

const loading = ref(false)
const saving = ref(false)
const rows = ref<TaintRow[]>([])
const original = ref<TaintRow[]>([])
const systemTaints = ref<NodeTaintVO[]>([])
const isControlPlane = ref(false)

const id = (t: TaintRow) => `${t.key.trim()}:${t.effect}`

const hasControlPlaneTaint = computed(() => rows.value.some(r => CONTROL_PLANE_KEYS.includes(r.key.trim())))
const newNoExecute = computed(() => {
  const old = new Set(original.value.map(id))
  return rows.value.filter(r => r.key.trim() && r.effect === 'NoExecute' && !old.has(id(r))).map(r => r.key.trim())
})
const removedControlPlane = computed(
  () => original.value.some(r => CONTROL_PLANE_KEYS.includes(r.key) && r.effect === 'NoSchedule') && !hasControlPlaneTaint.value
)

const load = async () => {
  rows.value = []
  systemTaints.value = []
  loading.value = true
  try {
    const { data } = await nodeApi.getDetail({ clusterId: props.clusterId, hostname: props.hostname })
    const taints = data?.taints || []
    systemTaints.value = taints.filter(t => t.system)
    rows.value = taints.filter(t => !t.system).map(t => ({ key: t.key, value: t.value || '', effect: t.effect }))
    original.value = rows.value.map(r => ({ ...r }))
    isControlPlane.value = (data?.roles || []).some(r => r === 'control-plane' || r === 'master')
  } catch {
    emit('update:modelValue', false)
  } finally {
    loading.value = false
  }
}

const KEY = /^([a-z0-9]([-a-z0-9.]*[a-z0-9])?\/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$/
const VALUE = /^([A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?)?$/

const save = async () => {
  const seen = new Set<string>()
  const taints: TaintRow[] = []
  for (const r of rows.value) {
    const key = r.key.trim()
    if (!key) continue
    if (!KEY.test(key)) {
      ElMessage.warning(`污点键 ${key} 不合法`)
      return
    }
    if (!VALUE.test(r.value.trim())) {
      ElMessage.warning(`污点 ${key} 的值不合法：最多 63 个字符，只能包含字母、数字、"-"、"_" 和 "."`)
      return
    }
    if (key.startsWith('node.kubernetes.io/')) {
      ElMessage.warning(`${key} 由节点控制器维护，不能手动设置`)
      return
    }
    if (seen.has(id(r))) {
      ElMessage.warning(`污点 ${id(r)} 重复：同一个键和效果只能有一个`)
      return
    }
    seen.add(id(r))
    taints.push({ key, value: r.value.trim(), effect: r.effect })
  }
  if (newNoExecute.value.length) {
    try {
      await ElMessageBox.confirm(`NoExecute 污点会立即驱逐节点上不容忍它的容器组，确定保存吗？`, '确认', {
        type: 'warning',
        confirmButtonText: '保存',
        cancelButtonText: '取消'
      })
    } catch {
      return
    }
  }
  saving.value = true
  try {
    await nodeApi.updateTaints({ clusterId: props.clusterId, hostname: props.hostname, taints })
    ElMessage.success('污点已更新')
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
.nt-tip {
  margin: 0 0 12px;
  color: #64748b;
  font-size: 13px;
  line-height: 1.6;
}

.nt-effects {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  margin-bottom: 18px;
}

.nt-effect {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 12px;
  border: 1px solid #e3eaf4;
  border-radius: 10px;
  background: #fbfcfe;
  font-size: 12px;
  line-height: 1.5;
}

.nt-effect strong {
  color: #1e293b;
  font-size: 13px;
}

.nt-effect span {
  color: #64748b;
}

.nt-table {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}

.nt-row {
  display: grid;
  grid-template-columns: 2fr 1.4fr 1.6fr 32px;
  align-items: center;
  gap: 8px;
}

.nt-row--head span {
  color: #64748b;
  font-size: 12px;
}

.nt-empty {
  padding: 12px;
  border: 1px dashed #e3eaf4;
  border-radius: 8px;
  color: #94a3b8;
  font-size: 13px;
  text-align: center;
}

.nt-actions {
  display: flex;
  gap: 8px;
}

.nt-alert {
  margin-bottom: 12px;
}

.nt-system {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.mono,
.mono-input :deep(input) {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}
</style>
