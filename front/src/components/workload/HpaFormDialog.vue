<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="820px"
    :close-on-click-modal="false"
    class="config-dialog form-dialog--compact"
    @update:model-value="(v: boolean) => emit('update:modelValue', v)"
    @open="reset"
  >
    <template #header>
      <div class="dialog-header">
        <h3 class="dialog-title">{{ title }}</h3>
      </div>
    </template>
    <div v-loading="loading" class="config-editor">
      <div class="config-content">
        <p class="hf-tip">
          根据 CPU / 内存的平均使用率自动调整副本数。使用率按容器的资源预留（requests）计算，需要集群安装 metrics-server，且容器设置了对应的资源预留。
        </p>
        <el-form :model="form" label-width="120px">
          <el-form-item label="命名空间" required>
            <el-select v-model="form.namespace" filterable :disabled="fixedTarget || isEdit" placeholder="请选择命名空间" style="width: 100%" @change="onNamespaceChange">
              <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
            </el-select>
          </el-form-item>
          <el-form-item label="目标工作负载" required>
            <el-select v-model="form.target" filterable :disabled="fixedTarget || isEdit" placeholder="选择 Deployment / StatefulSet" style="width: 100%" :loading="workloadLoading" @change="onTargetChange">
              <el-option v-for="o in targetOptions" :key="o.value" :label="o.label" :value="o.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="名称" required>
            <el-input v-model="form.name" :disabled="isEdit" placeholder="默认与工作负载同名" maxlength="253" />
          </el-form-item>
          <el-form-item label="副本范围" required>
            <el-input-number v-model="form.minReplicas" :min="1" :max="1000" controls-position="right" />
            <span class="hf-sep">至</span>
            <el-input-number v-model="form.maxReplicas" :min="1" :max="1000" controls-position="right" />
          </el-form-item>
          <el-form-item label="CPU 使用率">
            <el-switch v-model="form.cpuEnabled" />
            <template v-if="form.cpuEnabled">
              <el-input-number v-model="form.cpu" :min="1" :max="500" controls-position="right" class="hf-num" />
              <span class="fd-inline-hint">% ，超过时扩容，低于时缩容</span>
            </template>
          </el-form-item>
          <el-form-item label="内存使用率">
            <el-switch v-model="form.memEnabled" />
            <template v-if="form.memEnabled">
              <el-input-number v-model="form.mem" :min="1" :max="500" controls-position="right" class="hf-num" />
              <span class="fd-inline-hint">%</span>
            </template>
          </el-form-item>
          <el-form-item label="缩容稳定窗口">
            <el-input-number v-model="form.scaleDownWindow" :min="0" :max="3600" :step="60" controls-position="right" />
            <span class="fd-inline-hint">秒，在这段时间内按最高的建议副本数缩容，避免频繁抖动（默认 300）</span>
          </el-form-item>
        </el-form>
        <el-alert v-if="otherMetrics" type="info" :closable="false" show-icon>
          该弹性伸缩还有 {{ otherMetrics }} 个其他指标（自定义 / 外部指标），保存时会保留，如需修改请使用「编辑YAML」。
        </el-alert>
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="emit('update:modelValue', false)">取消</el-button>
        <el-button type="primary" :loading="saving" :disabled="loading" @click="save">{{ isEdit ? '保存' : '创建' }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { editableCopy, resourceApi, saveResourceObject, type KubeObject } from '@/api/resource'
import { workloadApi } from '@/api/workload'
import { namespaceApi } from '@/api/namespace'

const HPA_REF = { kind: 'horizontalpodautoscaler' as const }

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    clusterId: string
    namespace?: string
    // 固定的目标(工作负载详情页)
    target?: { kind: 'Deployment' | 'StatefulSet'; name: string } | null
    // 编辑已有的弹性伸缩
    editName?: string
  }>(),
  { namespace: '', target: null, editName: '' }
)

const emit = defineEmits<{ 'update:modelValue': [value: boolean]; success: [] }>()

const isEdit = computed(() => !!props.editName)
const fixedTarget = computed(() => !!props.target)
const title = computed(() => (isEdit.value ? '编辑弹性伸缩' : '设置弹性伸缩'))

const loading = ref(false)
const saving = ref(false)
const namespaces = ref<string[]>([])
const workloadLoading = ref(false)
const targetOptions = ref<{ value: string; label: string }[]>([])
const original = ref<KubeObject | null>(null)
const otherMetrics = ref(0)

const createForm = () => ({
  namespace: '',
  target: '',
  name: '',
  minReplicas: 1,
  maxReplicas: 5,
  cpuEnabled: true,
  cpu: 70,
  memEnabled: false,
  mem: 80,
  scaleDownWindow: 300
})
const form = reactive(createForm())

const isCpuMem = (m: KubeObject) => m.type === 'Resource' && ['cpu', 'memory'].includes(m.resource?.name)

const loadNamespaces = async () => {
  try {
    const resp = await namespaceApi.getList({ clusterId: props.clusterId, page: 1, pageSize: 10000 })
    namespaces.value = (resp.data?.items || []).map(ns => ns.name)
  } catch {
    namespaces.value = []
  }
}

const loadTargets = async () => {
  if (fixedTarget.value || !form.namespace) return
  workloadLoading.value = true
  try {
    const resp = await workloadApi.selectors({ clusterId: props.clusterId, namespace: form.namespace })
    targetOptions.value = (resp.data || [])
      .filter(w => w.kind === 'deployment' || w.kind === 'statefulset')
      .map(w => {
        const kind = w.kind === 'deployment' ? 'Deployment' : 'StatefulSet'
        return { value: `${kind}/${w.name}`, label: `${kind} / ${w.name}` }
      })
  } catch {
    targetOptions.value = []
  } finally {
    workloadLoading.value = false
  }
}

const onNamespaceChange = () => {
  form.target = ''
  loadTargets()
}

const onTargetChange = (value: string) => {
  if (!isEdit.value && !form.name) form.name = value.split('/')[1] || ''
}

const reset = async () => {
  Object.assign(form, createForm())
  original.value = null
  otherMetrics.value = 0
  form.namespace = props.namespace || ''
  if (props.target) {
    form.target = `${props.target.kind}/${props.target.name}`
    targetOptions.value = [{ value: form.target, label: `${props.target.kind} / ${props.target.name}` }]
    form.name = props.target.name
  }
  if (!fixedTarget.value) loadNamespaces()
  if (!isEdit.value) {
    loadTargets()
    return
  }
  loading.value = true
  try {
    const { data } = await resourceApi.object({ ...HPA_REF, clusterId: props.clusterId, namespace: form.namespace, name: props.editName })
    original.value = data
    const spec = data.spec || {}
    const ref = spec.scaleTargetRef || {}
    form.name = data.metadata?.name || props.editName
    form.target = `${ref.kind}/${ref.name}`
    targetOptions.value = [{ value: form.target, label: `${ref.kind} / ${ref.name}` }]
    form.minReplicas = spec.minReplicas ?? 1
    form.maxReplicas = spec.maxReplicas ?? 1
    const metrics: KubeObject[] = spec.metrics || []
    const cpu = metrics.find(m => isCpuMem(m) && m.resource.name === 'cpu' && m.resource.target?.type === 'Utilization')
    const mem = metrics.find(m => isCpuMem(m) && m.resource.name === 'memory' && m.resource.target?.type === 'Utilization')
    form.cpuEnabled = !!cpu
    form.cpu = cpu?.resource.target.averageUtilization ?? 70
    form.memEnabled = !!mem
    form.mem = mem?.resource.target.averageUtilization ?? 80
    otherMetrics.value = metrics.filter(m => m !== cpu && m !== mem).length
    form.scaleDownWindow = spec.behavior?.scaleDown?.stabilizationWindowSeconds ?? 300
  } catch {
    emit('update:modelValue', false)
  } finally {
    loading.value = false
  }
}

const NAME = /^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$/

const save = async () => {
  const [kind, targetName] = form.target.split('/')
  if (!form.namespace || !kind || !targetName) {
    ElMessage.warning('请选择命名空间和目标工作负载')
    return
  }
  const name = form.name.trim() || targetName
  if (!NAME.test(name)) {
    ElMessage.warning('名称只能包含小写字母、数字、"-" 和 "."')
    return
  }
  if (form.maxReplicas < form.minReplicas) {
    ElMessage.warning('最大副本数不能小于最小副本数')
    return
  }
  const base: KubeObject = original.value
    ? editableCopy(original.value)
    : { apiVersion: 'autoscaling/v2', kind: 'HorizontalPodAutoscaler', metadata: { name, namespace: form.namespace }, spec: {} }
  // 保留表单不支持的指标(CPU / 内存的平均值指标也一并保留)
  const kept: KubeObject[] = (base.spec.metrics || []).filter(
    (m: KubeObject) => !(isCpuMem(m) && m.resource.target?.type === 'Utilization')
  )
  const metrics: KubeObject[] = []
  if (form.cpuEnabled) metrics.push({ type: 'Resource', resource: { name: 'cpu', target: { type: 'Utilization', averageUtilization: form.cpu } } })
  if (form.memEnabled) metrics.push({ type: 'Resource', resource: { name: 'memory', target: { type: 'Utilization', averageUtilization: form.mem } } })
  metrics.push(...kept)
  if (!metrics.length) {
    ElMessage.warning('请至少启用一个指标')
    return
  }
  base.spec.scaleTargetRef = { apiVersion: 'apps/v1', kind, name: targetName }
  base.spec.minReplicas = form.minReplicas
  base.spec.maxReplicas = form.maxReplicas
  base.spec.metrics = metrics
  base.spec.behavior = {
    ...(base.spec.behavior || {}),
    scaleDown: { ...(base.spec.behavior?.scaleDown || {}), stabilizationWindowSeconds: form.scaleDownWindow }
  }
  saving.value = true
  try {
    await saveResourceObject(props.clusterId, HPA_REF, base, !isEdit.value)
    ElMessage.success(isEdit.value ? '保存成功' : '弹性伸缩已创建')
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
.hf-tip {
  margin: 0 0 16px;
  color: #64748b;
  font-size: 13px;
  line-height: 1.6;
}

.hf-sep {
  margin: 0 10px;
  color: #64748b;
}

.hf-num {
  margin-left: 12px;
}
</style>
